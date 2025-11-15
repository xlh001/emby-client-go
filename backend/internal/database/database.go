package database

import (
	"fmt"
	"log"
	"time"

	"github.com/emby-client-go/backend/internal/config"
	"github.com/emby-client-go/backend/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func Init() error {
	var dialector gorm.Dialector

	switch config.AppConfig.Database.Type {
	case "postgres":
		dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			config.AppConfig.Database.Host,
			config.AppConfig.Database.Port,
			config.AppConfig.Database.Username,
			config.AppConfig.Database.Password,
			config.AppConfig.Database.Database,
			config.AppConfig.Database.SSLMode,
		)
		dialector = postgres.Open(dsn)
	case "mysql":
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			config.AppConfig.Database.Username,
			config.AppConfig.Database.Password,
			config.AppConfig.Database.Host,
			config.AppConfig.Database.Port,
			config.AppConfig.Database.Database,
		)
		dialector = mysql.Open(dsn)
	case "sqlite":
		dsn := config.AppConfig.Database.Database
		dialector = sqlite.Open(dsn)
	default:
		return fmt.Errorf("不支持的数据库类型: %s", config.AppConfig.Database.Type)
	}

	// GORM配置
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
		NowFunc: func() time.Time {
			return time.Now().Local()
		},
	}

	var err error
	DB, err = gorm.Open(dialector, gormConfig)
	if err != nil {
		return fmt.Errorf("数据库连接失败: %v", err)
	}

	// 获取底层sql.DB对象进行连接池配置
	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("获取数据库实例失败: %v", err)
	}

	// 设置连接池参数
	sqlDB.SetMaxIdleConns(config.AppConfig.Database.MaxIdleConns)
	sqlDB.SetMaxOpenConns(config.AppConfig.Database.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// 测试连接
	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("数据库连接测试失败: %v", err)
	}

	log.Printf("数据库连接成功 [%s]", config.AppConfig.Database.Type)

	// 自动迁移数据表
	if err := autoMigrate(); err != nil {
		return fmt.Errorf("数据表迁移失败: %v", err)
	}

	return nil
}

func autoMigrate() error {
	return DB.AutoMigrate(
		&models.User{},
		&models.EmbyServer{},
		&models.MediaLibrary{},
		&models.Device{},
		&models.MediaItem{},
		&models.PlaybackSession{},
		&models.PlaybackRecord{},
		&models.ConnectionLog{},
		&models.SystemConfig{},
	)
}

func Close() error {
	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}