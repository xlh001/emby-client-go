package models

import (
	"time"
	"gorm.io/gorm"
)

// Server Emby服务器模型
type Server struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Name        string         `gorm:"size:100;not null" json:"name"`
	URL         string         `gorm:"size:255;not null" json:"url"`
	APIKey      string         `gorm:"size:255" json:"api_key,omitempty"`
	UserName    string         `gorm:"size:100" json:"user_name"`
	Password    string         `gorm:"size:255" json:"password,omitempty"`
	Description string         `gorm:"size:500" json:"description"`
	IsActive    bool           `gorm:"default:true" json:"is_active"`
	Version     string         `gorm:"size:50" json:"version"`
	LastCheck   *time.Time     `json:"last_check"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	// 关联关系
	Devices []Device `gorm:"many2many:device_servers;" json:"devices,omitempty"`
}

// Device 设备模型
type Device struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	Name         string         `gorm:"size:100;not null" json:"name"`
	Identifier   string         `gorm:"size:255;uniqueIndex" json:"identifier"`
	Platform     string         `gorm:"size:50" json:"platform"` // Windows, Android, iOS, etc.
	Version      string         `gorm:"size:50" json:"version"`
	UserAgent    string         `gorm:"size:500" json:"user_agent"`
	IPAddress    string         `gorm:"size:45" json:"ip_address"` // IPv6支持
	MacAddress   string         `gorm:"size:17" json:"mac_address"`
	LastActive   *time.Time     `json:"last_active"`
	IsActive     bool           `gorm:"default:true" json:"is_active"`
	Description  string         `gorm:"size:500" json:"description"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`

	// 关联关系
	Servers []Server `gorm:"many2many:device_servers;" json:"servers,omitempty"`
}

// DeviceServer 设备-服务器关联模型
type DeviceServer struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	DeviceID  uint      `gorm:"not null" json:"device_id"`
	ServerID  uint      `gorm:"not null" json:"server_id"`
	Priority  int       `gorm:"default:1" json:"priority"`     // 优先级，用于默认服务器选择
	IsEnabled bool      `gorm:"default:true" json:"is_enabled"`
	LastUsed  *time.Time `json:"last_used"`                    // 最后使用时间
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// 关联关系
	Device Device `gorm:"foreignKey:DeviceID" json:"device,omitempty"`
	Server Server `gorm:"foreignKey:ServerID" json:"server,omitempty"`
}

// User 用户模型（用于管理系统）
type User struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Username  string    `gorm:"size:50;uniqueIndex;not null" json:"username"`
	Password  string    `gorm:"size:255;not null" json:"password,omitempty"`
	Email     string    `gorm:"size:100" json:"email"`
	Role      string    `gorm:"size:20;default:'user'" json:"role"` // admin, user
	IsActive  bool      `gorm:"default:true" json:"is_active"`
	LastLogin *time.Time `json:"last_login"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}