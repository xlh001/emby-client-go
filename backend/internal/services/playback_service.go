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

// PlaybackService 播放控制服务
type PlaybackService struct {
	db *gorm.DB
}

// NewPlaybackService 创建播放控制服务
func NewPlaybackService() *PlaybackService {
	return &PlaybackService{
		db: database.DB,
	}
}

// PlayCommand 播放控制命令
type PlayCommand struct {
	Command   string `json:"command"` // Play, Pause, Stop, Seek
	SessionID string `json:"session_id"`
	Position  int64  `json:"position,omitempty"`
}

// SendPlayCommand 发送播放控制命令
func (s *PlaybackService) SendPlayCommand(ctx context.Context, serverID uint, deviceID uint, cmd PlayCommand) error {
	var server models.EmbyServer
	if err := s.db.First(&server, serverID).Error; err != nil {
		return fmt.Errorf("服务器不存在: %w", err)
	}

	var device models.Device
	if err := s.db.First(&device, deviceID).Error; err != nil {
		return fmt.Errorf("设备不存在: %w", err)
	}

	client := emby.NewClient(server.URL, server.APIKey)

	switch cmd.Command {
	case "Play":
		return client.SendPlayCommand(ctx, cmd.SessionID)
	case "Pause":
		return client.SendPauseCommand(ctx, cmd.SessionID)
	case "Stop":
		return client.SendStopCommand(ctx, cmd.SessionID)
	case "Seek":
		return client.SendSeekCommand(ctx, cmd.SessionID, cmd.Position)
	default:
		return fmt.Errorf("未知的播放命令: %s", cmd.Command)
	}
}

// GetActiveSessions 获取活动播放会话
func (s *PlaybackService) GetActiveSessions(serverID uint) ([]models.PlaybackSession, error) {
	var sessions []models.PlaybackSession
	err := s.db.Where("emby_server_id = ? AND play_state != ?", serverID, "Stopped").
		Preload("Device").
		Preload("User").
		Preload("MediaItem").
		Order("last_update_at DESC").
		Find(&sessions).Error

	if err != nil {
		return nil, fmt.Errorf("获取活动会话失败: %w", err)
	}

	return sessions, nil
}

// GetSessionByID 获取会话详情
func (s *PlaybackService) GetSessionByID(sessionID uint) (*models.PlaybackSession, error) {
	var session models.PlaybackSession
	err := s.db.Preload("Device").
		Preload("User").
		Preload("MediaItem").
		Preload("EmbyServer").
		First(&session, sessionID).Error

	if err != nil {
		return nil, fmt.Errorf("会话不存在: %w", err)
	}

	return &session, nil
}

// SyncPlaybackSessions 同步播放会话
func (s *PlaybackService) SyncPlaybackSessions(ctx context.Context, serverID uint) error {
	var server models.EmbyServer
	if err := s.db.First(&server, serverID).Error; err != nil {
		return fmt.Errorf("服务器不存在: %w", err)
	}

	client := emby.NewClient(server.URL, server.APIKey)
	sessions, err := client.GetSessions(ctx)
	if err != nil {
		return fmt.Errorf("获取会话列表失败: %w", err)
	}

	for _, sess := range sessions {
		if sess.NowPlayingItem == nil {
			continue
		}

		var session models.PlaybackSession
		err := s.db.Where("emby_session_id = ?", sess.Id).First(&session).Error

		now := time.Now()
		if err == gorm.ErrRecordNotFound {
			// 创建新会话
			session = models.PlaybackSession{
				EmbyServerID:  serverID,
				EmbySessionID: sess.Id,
				PlayState:     sess.PlayState.PlayState,
				PositionTicks: sess.PlayState.PositionTicks,
				IsMuted:       sess.PlayState.IsMuted,
				VolumeLevel:   sess.PlayState.VolumeLevel,
				StartedAt:     now,
				LastUpdateAt:  now,
			}
			s.db.Create(&session)
		} else {
			// 更新会话
			updates := map[string]interface{}{
				"play_state":     sess.PlayState.PlayState,
				"position_ticks": sess.PlayState.PositionTicks,
				"is_muted":       sess.PlayState.IsMuted,
				"volume_level":   sess.PlayState.VolumeLevel,
				"last_update_at": now,
			}
			s.db.Model(&session).Updates(updates)
		}
	}

	return nil
}

// RecordPlayback 记录播放历史
func (s *PlaybackService) RecordPlayback(record *models.PlaybackRecord) error {
	if err := s.db.Create(record).Error; err != nil {
		return fmt.Errorf("记录播放历史失败: %w", err)
	}
	return nil
}

// GetPlaybackHistory 获取播放历史
func (s *PlaybackService) GetPlaybackHistory(userID uint, limit, offset int) ([]models.PlaybackRecord, int64, error) {
	var records []models.PlaybackRecord
	var total int64

	query := s.db.Model(&models.PlaybackRecord{}).Where("user_id = ?", userID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("统计播放历史失败: %w", err)
	}

	err := query.Preload("MediaItem").
		Preload("EmbyServer").
		Order("played_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&records).Error

	if err != nil {
		return nil, 0, fmt.Errorf("获取播放历史失败: %w", err)
	}

	return records, total, nil
}
