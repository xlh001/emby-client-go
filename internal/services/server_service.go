package services

import (
	"emby-client-go/internal/models"
	"errors"
	"fmt"
	"log"
	"time"

	"gorm.io/gorm"
)

// ServerService 服务器管理服务
type ServerService struct {
	db *gorm.DB
}

// NewServerService 创建服务器服务
func NewServerService(db *gorm.DB) *ServerService {
	return &ServerService{db: db}
}

// AddServer 添加服务器
func (s *ServerService) AddServer(server *models.Server) error {
	// 创建Emby客户端测试连接
	client := NewEmbyClient(server.URL, false)

	var serverInfo *EmbyServerInfo
	var err error

	// 测试连接和认证
	if server.UserName != "" && server.Password != "" {
		serverInfo, err = client.TestWithCredentials(server.UserName, server.Password)
	} else {
		serverInfo, err = client.GetServerInfo()
		if err != nil {
			return fmt.Errorf("无法连接到服务器，请检查URL和认证信息: %v", err)
		}
	}

	if err != nil {
		return fmt.Errorf("服务器连接测试失败: %v", err)
	}

	// 更新服务器信息
	server.Version = serverInfo.Version
	now := time.Now()
	server.LastCheck = &now

	// 保存到数据库
	if err := s.db.Create(server).Error; err != nil {
		return fmt.Errorf("保存服务器失败: %v", err)
	}

	log.Printf("成功添加服务器: %s", server.Name)
	return nil
}

// GetServers 获取服务器列表
func (s *ServerService) GetServers() ([]models.Server, error) {
	var servers []models.Server
	err := s.db.Find(&servers).Error
	return servers, err
}

// GetServer 获取单个服务器
func (s *ServerService) GetServer(id uint) (*models.Server, error) {
	var server models.Server
	err := s.db.First(&server, id).Error
	if err != nil {
		return nil, err
	}
	return &server, nil
}

// UpdateServer 更新服务器
func (s *ServerService) UpdateServer(id uint, updates *models.Server) error {
	var server models.Server
	if err := s.db.First(&server, id).Error; err != nil {
		return err
	}

	// 如果URL或认证信息有变化，重新测试连接
	if updates.URL != server.URL || updates.UserName != server.UserName || updates.Password != server.Password {
		if updates.URL != "" {
			client := NewEmbyClient(updates.URL, false)

			if updates.UserName != "" && updates.Password != "" {
				serverInfo, err := client.TestWithCredentials(updates.UserName, updates.Password)
				if err != nil {
					return fmt.Errorf("更新服务器连接测试失败: %v", err)
				}
				updates.Version = serverInfo.Version
			}
		}
	}

	return s.db.Model(&server).Updates(updates).Error
}

// DeleteServer 删除服务器
func (s *ServerService) DeleteServer(id uint) error {
	var server models.Server
	if err := s.db.First(&server, id).Error; err != nil {
		return err
	}

	// 删除关联的设备-服务器关系
	s.db.Where("server_id = ?", id).Delete(&models.DeviceServer{})

	return s.db.Delete(&server).Error
}

// TestConnection 测试服务器连接
func (s *ServerService) TestConnection(id uint) error {
	var server models.Server
	if err := s.db.First(&server, id).Error; err != nil {
		return err
	}

	client := NewEmbyClient(server.URL, false)
	var err error

	if server.UserName != "" && server.Password != "" {
		_, err = client.TestWithCredentials(server.UserName, server.Password)
	} else {
		err = client.TestConnection()
	}

	if err != nil {
		return err
	}

	// 更新最后检查时间
	now := time.Now()
	return s.db.Model(&server).Update("last_check", &now).Error
}

// GetServerDevices 获取服务器的设备列表（从Emby服务器）
func (s *ServerService) GetServerDevices(id uint) ([]EmbyDevice, error) {
	server, err := s.GetServer(id)
	if err != nil {
		return nil, err
	}

	if server.UserName == "" || server.Password == "" {
		return nil, errors.New("需要认证信息才能获取设备列表")
	}

	client := NewEmbyClient(server.URL, false)
	authResp, err := client.Authenticate(server.UserName, server.Password)
	if err != nil {
		return nil, err
	}

	devices, err := client.GetDevices(authResp.AccessToken)
	if err != nil {
		return nil, err
	}

	return devices, nil
}

// SyncDevicesFromServer 从服务器同步设备
func (s *ServerService) SyncDevicesFromServer(serverID uint, deviceService *DeviceService) error {
	embyDevices, err := s.GetServerDevices(serverID)
	if err != nil {
		return err
	}

	server, err := s.GetServer(serverID)
	if err != nil {
		return err
	}

	var syncedCount int
	for _, embyDevice := range embyDevices {
		device := &models.Device{
			Name:       embyDevice.Name,
			Identifier: embyDevice.DeviceId,
			Platform:   embyDevice.AppName,
			Version:    embyDevice.AppVersion,
			UserAgent:  fmt.Sprintf("%s/%s", embyDevice.AppName, embyDevice.AppVersion),
		}

		// 创建或更新设备
		var existingDevice models.Device
		err := s.db.Where("identifier = ?", device.Identifier).First(&existingDevice).Error
		if err == nil {
			// 设备已存在，更新信息
			device.ID = existingDevice.ID
			if err := s.db.Model(&existingDevice).Updates(device).Error; err != nil {
				continue
			}
			device.ID = existingDevice.ID
		} else {
			// 创建新设备
			if err := s.db.Create(device).Error; err != nil {
				continue
			}
		}

		// 关联设备和服务器
		var deviceServer models.DeviceServer
		err = s.db.Where("device_id = ? AND server_id = ?", device.ID, serverID).First(&deviceServer).Error
		if err == nil {
			// 更新使用时间
			now := time.Now()
			s.db.Model(&deviceServer).Update("last_used", &now)
		} else {
			// 创建新的关联
			deviceServer = models.DeviceServer{
				DeviceID:  device.ID,
				ServerID:  serverID,
				Priority:  1,
				IsEnabled: true,
			}
			s.db.Create(&deviceServer)
		}

		syncedCount++
	}

	log.Printf("从服务器 %s 同步了 %d 个设备", server.Name, syncedCount)
	return nil
}