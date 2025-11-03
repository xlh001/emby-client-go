package main

import (
	"log"
	"os"

	"emby-client-go/internal/config"
	"emby-client-go/internal/services"
	"emby-client-go/internal/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	// 加载配置
	cfg := config.Load()

	// 确保数据目录存在
	if err := os.MkdirAll("./data", 0755); err != nil {
		log.Fatal("创建数据目录失败:", err)
	}

	// 初始化数据库
	db, err := gorm.Open(sqlite.Open(cfg.Database.Path), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatal("数据库连接失败:", err)
	}

	// 自动迁移
	err = db.AutoMigrate(
		&models.Server{},
		&models.Device{},
		&models.DeviceServer{},
		&models.User{},
	)
	if err != nil {
		log.Fatal("数据库迁移失败:", err)
	}

	// 初始化认证服务
	authService := services.NewAuthService(cfg.JWT.Secret)
	authService.SetDB(db)

	// 创建默认管理员用户（如果不存在）
	err = authService.CreateAdminUser(db, "admin", "admin123")
	if err != nil {
		log.Printf("创建默认管理员失败: %v", err)
	} else {
		log.Println("默认管理员账户已创建: admin/admin123")
	}

	log.Println("Emby管理系统数据库初始化完成!")
	log.Println("现在可以运行 'go run cmd/main.go' 启动Web服务")
}