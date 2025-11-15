package models

import (
	"time"
	"gorm.io/gorm"
)

// User 用户模型
type User struct {
	ID               uint           `json:"id" gorm:"primaryKey"`
	Username         string         `json:"username" gorm:"uniqueIndex;not null"`
	Email            string         `json:"email" gorm:"uniqueIndex;not null"`
	Password         string         `json:"-" gorm:"not null"`
	Nickname         string         `json:"nickname"`
	Role             string         `json:"role" gorm:"default:'user'"` // admin, user
	Status           string         `json:"status" gorm:"default:'active'"` // active, inactive, locked
	LastLogin        *time.Time     `json:"last_login"`
	FailedLoginCount int            `json:"-" gorm:"default:0"` // 登录失败次数
	LockedUntil      *time.Time     `json:"-"` // 账户锁定截止时间
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `json:"-" gorm:"index"`

	// 关联
	EmbyServers []EmbyServer `json:"emby_servers,omitempty" gorm:"many2many:user_emby_servers;"`
}

// EmbyServer Emby服务器模型
type EmbyServer struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"not null"`
	URL         string    `json:"url" gorm:"not null"`
	APIKey      string    `json:"api_key" gorm:"not null"`
	Version     string    `json:"version"`
	OS          string    `json:"os"`
	Status      string    `json:"status" gorm:"default:'offline'"` // online, offline, error
	LastCheck   *time.Time `json:"last_check"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`

	// 关联
	Users []User `json:"users,omitempty" gorm:"many2many:user_emby_servers;"`
	MediaLibraries []MediaLibrary `json:"media_libraries,omitempty"`
}

// MediaLibrary 媒体库模型
type MediaLibrary struct {
	ID             uint      `json:"id" gorm:"primaryKey"`
	EmbyServerID   uint      `json:"emby_server_id" gorm:"not null"`
	Name           string    `json:"name" gorm:"not null"`
	Type           string    `json:"type" gorm:"not null"` // movies, tvshows, music, photos
	Path           string    `json:"path"`
	TotalItems     int       `json:"total_items" gorm:"default:0"`
	TotalSize      int64     `json:"total_size" gorm:"default:0"`
	LastRefresh    *time.Time `json:"last_refresh"`
	CollectionType string    `json:"collection_type"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `json:"-" gorm:"index"`

	// 关联
	EmbyServer EmbyServer `json:"emby_server,omitempty" gorm:"foreignKey:EmbyServerID"`
}

// ConnectionLog 连接日志模型
type ConnectionLog struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	EmbyServerID uint      `json:"emby_server_id" gorm:"not null"`
	UserID       uint      `json:"user_id" gorm:"not null"`
	Action       string    `json:"action" gorm:"not null"` // connect, disconnect, error
	Message      string    `json:"message"`
	ResponseTime int       `json:"response_time"` // 毫秒
	Status       string    `json:"status" gorm:"not null"` // success, failed
	CreatedAt    time.Time `json:"created_at"`

	// 关联
	EmbyServer EmbyServer `json:"emby_server,omitempty" gorm:"foreignKey:EmbyServerID"`
	User       User       `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// Device 设备模型
type Device struct {
	ID             uint           `json:"id" gorm:"primaryKey"`
	EmbyServerID   uint           `json:"emby_server_id" gorm:"not null;index"`
	EmbyDeviceID   string         `json:"emby_device_id" gorm:"not null;index"` // Emby服务器中的设备ID
	Name           string         `json:"name" gorm:"not null"`
	DeviceType     string         `json:"device_type"`                          // Web, Android, iOS, etc.
	AppName        string         `json:"app_name"`
	AppVersion     string         `json:"app_version"`
	LastUserName   string         `json:"last_user_name"`
	LastActivityAt *time.Time     `json:"last_activity_at"`
	IsActive       bool           `json:"is_active" gorm:"default:false"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `json:"-" gorm:"index"`

	// 关联
	EmbyServer     EmbyServer      `json:"emby_server,omitempty" gorm:"foreignKey:EmbyServerID"`
	PlaybackSessions []PlaybackSession `json:"playback_sessions,omitempty"`
}

// MediaItem 媒体项目模型
type MediaItem struct {
	ID             uint           `json:"id" gorm:"primaryKey"`
	MediaLibraryID uint           `json:"media_library_id" gorm:"not null;index"`
	EmbyItemID     string         `json:"emby_item_id" gorm:"not null;index"` // Emby中的项目ID
	Name           string         `json:"name" gorm:"not null;index"`
	Type           string         `json:"type" gorm:"not null;index"` // Movie, Episode, Audio, etc.
	Path           string         `json:"path"`
	ParentID       string         `json:"parent_id" gorm:"index"` // 父项目ID（如剧集的剧集ID）
	SeriesName     string         `json:"series_name" gorm:"index"`
	SeasonNumber   *int           `json:"season_number"`
	EpisodeNumber  *int           `json:"episode_number"`
	Year           *int           `json:"year"`
	RunTimeTicks   int64          `json:"run_time_ticks"` // 播放时长（100纳秒为单位）
	Size           int64          `json:"size"`           // 文件大小（字节）
	Container      string         `json:"container"`      // 容器格式
	VideoCodec     string         `json:"video_codec"`
	AudioCodec     string         `json:"audio_codec"`
	Resolution     string         `json:"resolution"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `json:"-" gorm:"index"`

	// 关联
	MediaLibrary   MediaLibrary    `json:"media_library,omitempty" gorm:"foreignKey:MediaLibraryID"`
	PlaybackRecords []PlaybackRecord `json:"playback_records,omitempty"`
}

// PlaybackSession 播放会话模型（实时播放状态）
type PlaybackSession struct {
	ID             uint       `json:"id" gorm:"primaryKey"`
	EmbyServerID   uint       `json:"emby_server_id" gorm:"not null;index"`
	DeviceID       uint       `json:"device_id" gorm:"not null;index"`
	UserID         uint       `json:"user_id" gorm:"not null;index"`
	MediaItemID    uint       `json:"media_item_id" gorm:"index"`
	EmbySessionID  string     `json:"emby_session_id" gorm:"not null;index"` // Emby会话ID
	PlayState      string     `json:"play_state" gorm:"not null"`            // Playing, Paused, Stopped
	PositionTicks  int64      `json:"position_ticks"`                        // 当前播放位置
	IsMuted        bool       `json:"is_muted"`
	VolumeLevel    int        `json:"volume_level"`
	AudioStreamIndex int      `json:"audio_stream_index"`
	SubtitleStreamIndex int   `json:"subtitle_stream_index"`
	PlayMethod     string     `json:"play_method"` // DirectPlay, Transcode, DirectStream
	StartedAt      time.Time  `json:"started_at"`
	LastUpdateAt   time.Time  `json:"last_update_at"`
	EndedAt        *time.Time `json:"ended_at"`

	// 关联
	EmbyServer EmbyServer `json:"emby_server,omitempty" gorm:"foreignKey:EmbyServerID"`
	Device     Device     `json:"device,omitempty" gorm:"foreignKey:DeviceID"`
	User       User       `json:"user,omitempty" gorm:"foreignKey:UserID"`
	MediaItem  *MediaItem `json:"media_item,omitempty" gorm:"foreignKey:MediaItemID"`
}

// PlaybackRecord 播放记录模型（历史记录）
type PlaybackRecord struct {
	ID             uint       `json:"id" gorm:"primaryKey"`
	EmbyServerID   uint       `json:"emby_server_id" gorm:"not null;index"`
	UserID         uint       `json:"user_id" gorm:"not null;index"`
	MediaItemID    uint       `json:"media_item_id" gorm:"not null;index"`
	DeviceID       uint       `json:"device_id" gorm:"index"`
	PlayedAt       time.Time  `json:"played_at" gorm:"not null;index"`
	DurationTicks  int64      `json:"duration_ticks"`  // 播放时长
	PositionTicks  int64      `json:"position_ticks"`  // 停止位置
	IsCompleted    bool       `json:"is_completed"`    // 是否播放完成
	PlayMethod     string     `json:"play_method"`
	ClientName     string     `json:"client_name"`
	DeviceName     string     `json:"device_name"`
	CreatedAt      time.Time  `json:"created_at"`

	// 关联
	EmbyServer EmbyServer `json:"emby_server,omitempty" gorm:"foreignKey:EmbyServerID"`
	User       User       `json:"user,omitempty" gorm:"foreignKey:UserID"`
	MediaItem  MediaItem  `json:"media_item,omitempty" gorm:"foreignKey:MediaItemID"`
	Device     *Device    `json:"device,omitempty" gorm:"foreignKey:DeviceID"`
}

// SystemConfig 系统配置模型
type SystemConfig struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Key       string    `json:"key" gorm:"uniqueIndex;not null"`
	Value     string    `json:"value" gorm:"type:text"`
	Category  string    `json:"category" gorm:"not null"`
	Description string `json:"description"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}