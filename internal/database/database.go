package database

import (
	"emby-client-go/internal/models"
	"log"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func Initialize(dbPath string) (*gorm.DB, error) {
	// 使用内存数据库进行演示
	// 在生产环境中，建议使用PostgreSQL或MySQL
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Printf("SQLite连接失败: %v，尝试使用内存数据库", err)
		// 如果SQLite失败，返回一个简单的内存数据库
		return gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Silent),
		})
	}

	// 自动迁移数据库表
	err = db.AutoMigrate(
		&models.Server{},
		&models.Device{},
		&models.DeviceServer{},
		&models.User{},
	)
	if err != nil {
		return nil, err
	}

	// 创建默认管理员用户
	if err := createDefaultUser(db); err != nil {
		log.Printf("创建默认用户失败: %v", err)
	}

	log.Println("数据库初始化完成（内存模式）")
	log.Println("注意：当前使用内存数据库，重启后数据会丢失")
	log.Println("生产环境建议配置PostgreSQL或MySQL数据库")
	return db, nil
}

func createDefaultUser(db *gorm.DB) error {
	// 检查是否已有管理员用户
	var count int64
	db.Model(&models.User{}).Where("role = ?", "admin").Count(&count)
	if count > 0 {
		return nil
	}

	// 创建默认管理员
	admin := models.User{
		Username: "admin",
		Password: "$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi", // admin123
		Role:     "admin",
		IsActive: true,
	}

	return db.Create(&admin).Error
}