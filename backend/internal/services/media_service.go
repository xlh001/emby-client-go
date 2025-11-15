package services

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/emby-client-go/backend/internal/database"
	"github.com/emby-client-go/backend/internal/models"
	"github.com/emby-client-go/backend/pkg/emby"
)

// MediaService 媒体库服务
type MediaService struct {
	cache      sync.Map // 简单的内存缓存
	cacheTTL   time.Duration
	syncMutex  sync.Mutex // 同步锁，防止并发同步同一服务器
}

// NewMediaService 创建媒体库服务
func NewMediaService() *MediaService {
	return &MediaService{
		cacheTTL: 5 * time.Minute, // 缓存5分钟
	}
}

// cacheKey 缓存键结构
type cacheKey struct {
	serverID uint
	key      string
}

// cacheValue 缓存值结构
type cacheValue struct {
	data      interface{}
	expiresAt time.Time
}

// setCache 设置缓存
func (s *MediaService) setCache(serverID uint, key string, value interface{}) {
	ck := cacheKey{serverID: serverID, key: key}
	cv := cacheValue{
		data:      value,
		expiresAt: time.Now().Add(s.cacheTTL),
	}
	s.cache.Store(ck, cv)
}

// getCache 获取缓存
func (s *MediaService) getCache(serverID uint, key string) (interface{}, bool) {
	ck := cacheKey{serverID: serverID, key: key}
	val, ok := s.cache.Load(ck)
	if !ok {
		return nil, false
	}

	cv := val.(cacheValue)
	if time.Now().After(cv.expiresAt) {
		s.cache.Delete(ck)
		return nil, false
	}

	return cv.data, true
}

// clearCache 清除特定服务器的缓存
func (s *MediaService) clearCache(serverID uint) {
	s.cache.Range(func(key, value interface{}) bool {
		ck := key.(cacheKey)
		if ck.serverID == serverID {
			s.cache.Delete(key)
		}
		return true
	})
}

// SyncMediaLibraries 同步服务器的媒体库
func (s *MediaService) SyncMediaLibraries(ctx context.Context, serverID uint) (int, error) {
	// 获取服务器信息
	var server models.EmbyServer
	if err := database.DB.First(&server, serverID).Error; err != nil {
		return 0, fmt.Errorf("服务器不存在: %w", err)
	}

	// 防止并发同步
	s.syncMutex.Lock()
	defer s.syncMutex.Unlock()

	// 创建Emby客户端
	client := emby.NewClient(server.URL, server.APIKey)

	// 获取媒体库列表
	libraries, err := client.GetLibraries(ctx)
	if err != nil {
		return 0, fmt.Errorf("获取媒体库列表失败: %w", err)
	}

	log.Printf("从服务器 %s 获取到 %d 个媒体库", server.Name, len(libraries))

	// 开始事务
	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	syncCount := 0

	for _, lib := range libraries {
		// 检查媒体库是否已存在
		var existingLib models.MediaLibrary
		err := tx.Where("emby_server_id = ? AND name = ?", serverID, lib.Name).
			First(&existingLib).Error

		now := time.Now()

		if err != nil {
			// 不存在，创建新记录
			newLib := models.MediaLibrary{
				EmbyServerID:   serverID,
				Name:           lib.Name,
				Type:           lib.CollectionType,
				CollectionType: lib.CollectionType,
				TotalItems:     lib.ItemCount,
				LastRefresh:    &now,
			}

			if err := tx.Create(&newLib).Error; err != nil {
				tx.Rollback()
				return 0, fmt.Errorf("创建媒体库失败: %w", err)
			}

			log.Printf("创建新媒体库: %s (类型: %s, 项目数: %d)", lib.Name, lib.CollectionType, lib.ItemCount)
			syncCount++
		} else {
			// 已存在，更新记录
			updates := map[string]interface{}{
				"type":            lib.CollectionType,
				"collection_type": lib.CollectionType,
				"total_items":     lib.ItemCount,
				"last_refresh":    &now,
			}

			if err := tx.Model(&existingLib).Updates(updates).Error; err != nil {
				tx.Rollback()
				return 0, fmt.Errorf("更新媒体库失败: %w", err)
			}

			log.Printf("更新媒体库: %s (项目数: %d)", lib.Name, lib.ItemCount)
			syncCount++
		}
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return 0, fmt.Errorf("提交事务失败: %w", err)
	}

	// 清除缓存
	s.clearCache(serverID)

	log.Printf("服务器 %s 媒体库同步完成，共同步 %d 个媒体库", server.Name, syncCount)

	return syncCount, nil
}

// GetMediaLibraries 获取服务器的媒体库列表（带缓存）
func (s *MediaService) GetMediaLibraries(serverID uint) ([]models.MediaLibrary, error) {
	// 尝试从缓存获取
	if cached, ok := s.getCache(serverID, "libraries"); ok {
		return cached.([]models.MediaLibrary), nil
	}

	// 从数据库查询
	var libraries []models.MediaLibrary
	if err := database.DB.Where("emby_server_id = ?", serverID).
		Order("created_at DESC").
		Find(&libraries).Error; err != nil {
		return nil, fmt.Errorf("查询媒体库失败: %w", err)
	}

	// 存入缓存
	s.setCache(serverID, "libraries", libraries)

	return libraries, nil
}

// GetAllMediaLibraries 获取所有服务器的媒体库
func (s *MediaService) GetAllMediaLibraries() ([]models.MediaLibrary, error) {
	var libraries []models.MediaLibrary
	if err := database.DB.Preload("EmbyServer").
		Order("emby_server_id, created_at DESC").
		Find(&libraries).Error; err != nil {
		return nil, fmt.Errorf("查询所有媒体库失败: %w", err)
	}

	return libraries, nil
}

// GetMediaLibrary 获取单个媒体库详情
func (s *MediaService) GetMediaLibrary(id uint) (*models.MediaLibrary, error) {
	var library models.MediaLibrary
	if err := database.DB.Preload("EmbyServer").First(&library, id).Error; err != nil {
		return nil, fmt.Errorf("媒体库不存在: %w", err)
	}

	return &library, nil
}

// SyncAllServers 同步所有在线服务器的媒体库
func (s *MediaService) SyncAllServers(ctx context.Context) (map[uint]int, error) {
	// 获取所有在线服务器
	var servers []models.EmbyServer
	if err := database.DB.Where("status = ?", "online").Find(&servers).Error; err != nil {
		return nil, fmt.Errorf("查询在线服务器失败: %w", err)
	}

	results := make(map[uint]int)
	var wg sync.WaitGroup
	var mu sync.Mutex

	// 并发同步所有服务器
	for _, server := range servers {
		wg.Add(1)
		go func(srv models.EmbyServer) {
			defer wg.Done()

			count, err := s.SyncMediaLibraries(ctx, srv.ID)
			mu.Lock()
			if err != nil {
				log.Printf("同步服务器 %s 失败: %v", srv.Name, err)
				results[srv.ID] = -1
			} else {
				results[srv.ID] = count
			}
			mu.Unlock()
		}(server)
	}

	wg.Wait()

	return results, nil
}

// RefreshMediaLibrary 刷新单个媒体库的统计信息
func (s *MediaService) RefreshMediaLibrary(ctx context.Context, libraryID uint) error {
	// 获取媒体库信息
	var library models.MediaLibrary
	if err := database.DB.Preload("EmbyServer").First(&library, libraryID).Error; err != nil {
		return fmt.Errorf("媒体库不存在: %w", err)
	}

	// 创建Emby客户端
	client := emby.NewClient(library.EmbyServer.URL, library.EmbyServer.APIKey)

	// 获取媒体库项目统计（分页获取第一页即可获取总数）
	_, totalCount, err := client.GetLibraryItems(ctx, library.Name, 0, 1)
	if err != nil {
		return fmt.Errorf("获取媒体库统计失败: %w", err)
	}

	// 更新数据库
	now := time.Now()
	updates := map[string]interface{}{
		"total_items":  totalCount,
		"last_refresh": &now,
	}

	if err := database.DB.Model(&library).Updates(updates).Error; err != nil {
		return fmt.Errorf("更新媒体库统计失败: %w", err)
	}

	// 清除缓存
	s.clearCache(library.EmbyServerID)

	log.Printf("媒体库 %s 统计信息已刷新，总项目数: %d", library.Name, totalCount)

	return nil
}

// GetMediaLibraryStats 获取媒体库统计信息
func (s *MediaService) GetMediaLibraryStats() (map[string]interface{}, error) {
	var stats struct {
		TotalLibraries int
		TotalItems     int64
		TotalSize      int64
	}

	// 统计媒体库数量
	var count int64
	if err := database.DB.Model(&models.MediaLibrary{}).Count(&count).Error; err != nil {
		return nil, fmt.Errorf("统计媒体库数量失败: %w", err)
	}
	stats.TotalLibraries = int(count)

	// 统计总项目数
	if err := database.DB.Model(&models.MediaLibrary{}).
		Select("COALESCE(SUM(total_items), 0)").
		Scan(&stats.TotalItems).Error; err != nil {
		return nil, fmt.Errorf("统计总项目数失败: %w", err)
	}

	// 统计总大小
	if err := database.DB.Model(&models.MediaLibrary{}).
		Select("COALESCE(SUM(total_size), 0)").
		Scan(&stats.TotalSize).Error; err != nil {
		return nil, fmt.Errorf("统计总大小失败: %w", err)
	}

	return map[string]interface{}{
		"total_libraries": stats.TotalLibraries,
		"total_items":     stats.TotalItems,
		"total_size":      stats.TotalSize,
	}, nil
}

// GetMediaItems 获取媒体项目列表
func (s *MediaService) GetMediaItems(libraryID uint, mediaType string, limit, offset int) ([]models.MediaItem, int64, error) {
	query := database.DB.Model(&models.MediaItem{}).
		Where("media_library_id = ?", libraryID).
		Preload("MediaLibrary").
		Preload("MediaLibrary.EmbyServer")

	if mediaType != "" {
		query = query.Where("type = ?", mediaType)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("统计媒体项目失败: %w", err)
	}

	var items []models.MediaItem
	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&items).Error; err != nil {
		return nil, 0, fmt.Errorf("获取媒体项目失败: %w", err)
	}

	return items, total, nil
}

// GetMediaItem 获取媒体项目详情
func (s *MediaService) GetMediaItem(id uint) (*models.MediaItem, error) {
	var item models.MediaItem
	if err := database.DB.Preload("MediaLibrary").
		Preload("MediaLibrary.EmbyServer").
		First(&item, id).Error; err != nil {
		return nil, fmt.Errorf("媒体项目不存在: %w", err)
	}

	return &item, nil
}
