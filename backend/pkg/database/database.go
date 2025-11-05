package database

import (
	"emby-manager/internal/config"
	"emby-manager/internal/models"

	"log"

	"gorm.io/driver/sqlite"
	"gorm.io/driver/postgres"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// Init 初始化数据库连接
func Init() error {
	var err error

	// 配置GORM日志
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	}

	// 根据数据库类型创建连接
	switch config.AppConfig.Database.Type {
	case "sqlite":
		log.Println("使用SQLite数据库")
		DB, err = gorm.Open(sqlite.Open(config.AppConfig.Database.GetDSN()), gormConfig)
	case "postgres":
		log.Println("使用PostgreSQL数据库")
		DB, err = gorm.Open(postgres.Open(config.AppConfig.Database.GetDSN()), gormConfig)
	case "mysql":
		log.Println("使用MySQL数据库")
		DB, err = gorm.Open(mysql.Open(config.AppConfig.Database.GetDSN()), gormConfig)
	default:
		log.Fatalf("不支持的数据库类型: %s", config.AppConfig.Database.Type)
	}

	if err != nil {
		return err
	}

	// 获取底层的sql.DB对象进行配置
	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}

	// 设置连接池
	sqlDB.SetMaxIdleConns(config.AppConfig.Database.MaxIdleConns)
	sqlDB.SetMaxOpenConns(config.AppConfig.Database.MaxOpenConns)

	log.Println("数据库连接成功")

	// 自动迁移数据库表
	if err := migrate(); err != nil {
		return err
	}

	return nil
}

// migrate 数据库迁移
func migrate() error {
	return DB.AutoMigrate(
		&models.User{},
		&models.EmbyServer{},
		&models.Device{},
		&models.MediaLibrary{},
		&models.MediaItem{},
		&models.PlaybackRecord{},
	)
}

// Close 关闭数据库连接
func Close() error {
	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}