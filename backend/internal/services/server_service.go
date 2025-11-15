package services

import (
	"context"
	"fmt"
	"time"

	"github.com/emby-client-go/backend/internal/database"
	"github.com/emby-client-go/backend/internal/models"
	"github.com/emby-client-go/backend/pkg/emby"
	"gorm.io/gorm"
)

type ServerService struct{}

func NewServerService() *ServerService {
	return &ServerService{}
}

// CreateServer 创建服务器
func (s *ServerService) CreateServer(server *models.EmbyServer, userID uint) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 测试连接
	client := emby.NewClient(server.URL, server.APIKey)
	info, duration, err := client.GetServerStatus(ctx)
	if err != nil {
		return fmt.Errorf("连接测试失败: %w", err)
	}

	// 更新服务器信息
	server.Version = info.Version
	server.OS = info.OperatingSystem
	server.Status = "online"
	now := time.Now()
	server.LastCheck = &now

	// 创建服务器记录
	if err := database.DB.Create(server).Error; err != nil {
		return fmt.Errorf("创建服务器失败: %w", err)
	}

	// 关联用户
	if err := database.DB.Model(server).Association("Users").Append(&models.User{ID: userID}); err != nil {
		return fmt.Errorf("关联用户失败: %w", err)
	}

	// 记录连接日志
	log := models.ConnectionLog{
		EmbyServerID: server.ID,
		UserID:       userID,
		Action:       "connect",
		Message:      fmt.Sprintf("服务器连接成功，版本: %s", info.Version),
		ResponseTime: int(duration.Milliseconds()),
		Status:       "success",
	}
	database.DB.Create(&log)

	return nil
}

// UpdateServer 更新服务器
func (s *ServerService) UpdateServer(id uint, updates map[string]interface{}) error {
	// 如果更新了URL或APIKey，需要重新测试连接
	if url, hasURL := updates["url"].(string); hasURL {
		if apiKey, hasKey := updates["api_key"].(string); hasKey {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			client := emby.NewClient(url, apiKey)
			info, _, err := client.GetServerStatus(ctx)
			if err != nil {
				return fmt.Errorf("连接测试失败: %w", err)
			}

			// 更新版本和系统信息
			updates["version"] = info.Version
			updates["os"] = info.OperatingSystem
			updates["status"] = "online"
			now := time.Now()
			updates["last_check"] = &now
		}
	}

	return database.DB.Model(&models.EmbyServer{}).Where("id = ?", id).Updates(updates).Error
}

// DeleteServer 删除服务器
func (s *ServerService) DeleteServer(id uint) error {
	return database.DB.Delete(&models.EmbyServer{}, id).Error
}

// GetServer 获取服务器详情
func (s *ServerService) GetServer(id uint) (*models.EmbyServer, error) {
	var server models.EmbyServer
	if err := database.DB.Preload("Users").Preload("MediaLibraries").First(&server, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("服务器不存在")
		}
		return nil, fmt.Errorf("查询服务器失败: %w", err)
	}
	return &server, nil
}

// GetServers 获取服务器列表
func (s *ServerService) GetServers(userID uint, page, pageSize int) ([]models.EmbyServer, int64, error) {
	var servers []models.EmbyServer
	var total int64

	query := database.DB.Model(&models.EmbyServer{}).
		Joins("JOIN user_emby_servers ON user_emby_servers.emby_server_id = emby_servers.id").
		Where("user_emby_servers.user_id = ?", userID)

	// 统计总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Find(&servers).Error; err != nil {
		return nil, 0, err
	}

	return servers, total, nil
}

// TestConnection 测试服务器连接
func (s *ServerService) TestConnection(id uint, userID uint) (time.Duration, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	server, err := s.GetServer(id)
	if err != nil {
		return 0, err
	}

	client := emby.NewClient(server.URL, server.APIKey)
	duration, err := client.TestConnection(ctx)

	// 记录连接日志
	log := models.ConnectionLog{
		EmbyServerID: server.ID,
		UserID:       userID,
		Action:       "connect",
		ResponseTime: int(duration.Milliseconds()),
	}

	if err != nil {
		log.Status = "failed"
		log.Message = err.Error()
		database.DB.Create(&log)
		return 0, err
	}

	log.Status = "success"
	log.Message = "连接测试成功"
	database.DB.Create(&log)

	// 更新服务器状态
	now := time.Now()
	database.DB.Model(server).Updates(map[string]interface{}{
		"status":     "online",
		"last_check": &now,
	})

	return duration, nil
}

// SyncDevices 同步设备列表
func (s *ServerService) SyncDevices(id uint) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	server, err := s.GetServer(id)
	if err != nil {
		return err
	}

	client := emby.NewClient(server.URL, server.APIKey)
	devices, err := client.GetDevices(ctx)
	if err != nil {
		return fmt.Errorf("获取设备列表失败: %w", err)
	}

	// 同步设备到数据库
	for _, dev := range devices {
		var device models.Device
		result := database.DB.Where("emby_server_id = ? AND emby_device_id = ?", server.ID, dev.ID).First(&device)

		if result.Error == gorm.ErrRecordNotFound {
			// 创建新设备
			device = models.Device{
				EmbyServerID: server.ID,
				EmbyDeviceID: dev.ID,
				Name:         dev.Name,
				DeviceType:   dev.AppName,
				AppName:      dev.AppName,
				AppVersion:   dev.AppVersion,
				LastUserName: dev.LastUserName,
				IsActive:     true,
			}
			database.DB.Create(&device)
		} else {
			// 更新设备信息
			now := time.Now()
			database.DB.Model(&device).Updates(map[string]interface{}{
				"name":          dev.Name,
				"app_name":      dev.AppName,
				"app_version":   dev.AppVersion,
				"last_user_name": dev.LastUserName,
				"last_activity_at": &now,
				"is_active":     true,
			})
		}
	}

	return nil
}

// SyncLibraries 同步媒体库列表
func (s *ServerService) SyncLibraries(id uint) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	server, err := s.GetServer(id)
	if err != nil {
		return err
	}

	client := emby.NewClient(server.URL, server.APIKey)
	libraries, err := client.GetLibraries(ctx)
	if err != nil {
		return fmt.Errorf("获取媒体库列表失败: %w", err)
	}

	// 同步媒体库到数据库
	for _, lib := range libraries {
		var library models.MediaLibrary
		result := database.DB.Where("emby_server_id = ? AND name = ?", server.ID, lib.Name).First(&library)

		if result.Error == gorm.ErrRecordNotFound {
			// 创建新媒体库
			library = models.MediaLibrary{
				EmbyServerID:   server.ID,
				Name:           lib.Name,
				Type:           lib.CollectionType,
				CollectionType: lib.CollectionType,
				TotalItems:     lib.ItemCount,
			}
			database.DB.Create(&library)
		} else {
			// 更新媒体库信息
			now := time.Now()
			database.DB.Model(&library).Updates(map[string]interface{}{
				"total_items":  lib.ItemCount,
				"last_refresh": &now,
			})
		}
	}

	return nil
}

// GetConnectionLogs 获取连接日志
func (s *ServerService) GetConnectionLogs(serverID uint, page, pageSize int) ([]models.ConnectionLog, int64, error) {
	var logs []models.ConnectionLog
	var total int64

	query := database.DB.Model(&models.ConnectionLog{}).Where("emby_server_id = ?", serverID)

	// 统计总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).
		Preload("User").Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}
