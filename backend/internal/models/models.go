package models

import (
	"time"

	"gorm.io/gorm"
)

// User 用户模型
type User struct {
	ID             uint           `gorm:"primaryKey" json:"id"`
	Username       string         `gorm:"unique;not null" json:"username"`
	Email          string         `gorm:"unique" json:"email"`
	Password       string         `gorm:"not null" json:"-"`
	Role           string         `gorm:"default:user" json:"role"` // admin, user
	Avatar         string         `json:"avatar"`
	Status         string         `gorm:"default:active" json:"status"` // active, inactive, banned
	FailedAttempts int            `gorm:"default:0" json:"-"`
	LockedUntil    *time.Time     `json:"locked_until"`
	LastLoginAt    *time.Time     `json:"last_login_at"`
	Preferences    string         `json:"preferences"` // JSON配置
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`

	// 关联关系
	EmbyServers []EmbyServer `gorm:"foreignKey:CreatedBy" json:"emby_servers,omitempty"`
	Devices      []Device     `gorm:"foreignKey:UserID" json:"devices,omitempty"`
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}

// EmbyServer Emby服务器模型
type EmbyServer struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Name        string         `gorm:"not null" json:"name"`
	URL         string         `gorm:"not null" json:"url"`
	APIKey      string         `gorm:"not null" json:"-"`
	Version     string         `json:"version"`
	Status      string         `gorm:"default:offline" json:"status"` // online, offline, error
	LastCheck   *time.Time     `json:"last_check"`
	Description string         `json:"description"`
	CreatedBy   uint           `json:"created_by"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	// 关联关系
	Creator       User           `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
	MediaLibraries []MediaLibrary `gorm:"foreignKey:ServerID" json:"media_libraries,omitempty"`
	Devices       []Device       `gorm:"foreignKey:ServerID" json:"devices,omitempty"`
}

// TableName 指定表名
func (EmbyServer) TableName() string {
	return "emby_servers"
}

// Device 设备模型
type Device struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	Name         string         `gorm:"not null" json:"name"`
	DeviceID     string         `gorm:"unique;not null" json:"device_id"`
	UserID       uint           `json:"user_id"`
	ServerID     *uint          `json:"server_id"`
	Client       string         `json:"client"`      // Android, iOS, Web, etc.
	LastActivity *time.Time     `json:"last_activity"`
	IP           string         `json:"ip"`
	UserAgent    string         `json:"user_agent"`
	Status       string         `gorm:"default:active" json:"status"` // active, inactive
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`

	// 关联关系
	User         User         `gorm:"foreignKey:UserID" json:"user,omitempty"`
	EmbyServer   *EmbyServer  `gorm:"foreignKey:ServerID" json:"emby_server,omitempty"`
	PlaybackRecords []PlaybackRecord `gorm:"foreignKey:DeviceID" json:"playback_records,omitempty"`
}

// TableName 指定表名
func (Device) TableName() string {
	return "devices"
}

// MediaLibrary 媒体库模型
type MediaLibrary struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	ServerID    uint           `json:"server_id"`
	EmbyID      string         `gorm:"not null" json:"emby_id"`
	Name        string         `gorm:"not null" json:"name"`
	Type        string         `json:"type"`        // movies, tvshows, music, photos
	Path        string         `json:"path"`
	ItemCount   int            `json:"item_count"`
	SyncedAt    *time.Time     `json:"synced_at"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	// 关联关系
	EmbyServer  EmbyServer  `gorm:"foreignKey:ServerID" json:"emby_server,omitempty"`
	MediaItems  []MediaItem `gorm:"foreignKey:LibraryID" json:"media_items,omitempty"`
}

// TableName 指定表名
func (MediaLibrary) TableName() string {
	return "media_libraries"
}

// MediaItem 媒体项目模型
type MediaItem struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	LibraryID   uint           `json:"library_id"`
	EmbyID      string         `gorm:"unique;not null" json:"emby_id"`
	Title       string         `gorm:"not null" json:"title"`
	Type        string         `json:"type"`        // Movie, Series, Episode, MusicAlbum, etc.
	Path        string         `json:"path"`
	Size        int64          `json:"size"`
	Duration    int            `json:"duration"`    // 秒
	Year        *int           `json:"year"`
	Genres      string         `json:"genres"`      // JSON数组
	Rating      *float64       `json:"rating"`
	ParentalRating string      `json:"parental_rating"`
	Thumbnail   string         `json:"thumbnail"`
	SyncedAt    *time.Time     `json:"synced_at"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	// 关联关系
	MediaLibrary     MediaLibrary      `gorm:"foreignKey:LibraryID" json:"media_library,omitempty"`
	PlaybackRecords  []PlaybackRecord  `gorm:"foreignKey:MediaItemID" json:"playback_records,omitempty"`
}

// TableName 指定表名
func (MediaItem) TableName() string {
	return "media_items"
}

// PlaybackRecord 播放记录模型
type PlaybackRecord struct {
	ID         uint           `gorm:"primaryKey" json:"id"`
	DeviceID   uint           `json:"device_id"`
	ServerID   uint           `json:"server_id"`
	MediaItemID uint          `json:"media_item_id"`
	UserID     uint           `json:"user_id"`
	Position   int            `json:"position"`     // 播放位置（秒）
	Duration   int            `json:"duration"`     // 总时长（秒）
	Completed  bool           `gorm:"default:false" json:"completed"`   // 是否播放完成
	PlayedAt   time.Time      `json:"played_at"`    // 播放时间
	Synced     bool           `gorm:"default:false" json:"synced"`      // 是否已同步到Emby服务器
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`

	// 关联关系
	Device    Device    `gorm:"foreignKey:DeviceID" json:"device,omitempty"`
	EmbyServer EmbyServer `gorm:"foreignKey:ServerID" json:"emby_server,omitempty"`
	MediaItem MediaItem `gorm:"foreignKey:MediaItemID" json:"media_item,omitempty"`
	User      User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// TableName 指定表名
func (PlaybackRecord) TableName() string {
	return "playback_records"
}