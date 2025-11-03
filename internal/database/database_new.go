package database

import (
	"emby-client-go/internal/models"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type DatabaseConfig struct {
	Type     string `mapstructure:"type"`
	Path     string `mapstructure:"path"`
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	Database string `mapstructure:"database"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	SSLMode  string `mapstructure:"ssl_mode"`
}

func Initialize(config DatabaseConfig) (*gorm.DB, error) {
	var db *gorm.DB
	var err error

	log.Printf("正在初始化数据库，类型: %s", config.Type)

	switch strings.ToLower(config.Type) {
	case "sqlite", "":
		db, err = initSQLite(config)
	case "mysql":
		db, err = initMySQL(config)
	case "postgres", "postgresql":
		db, err = initPostgreSQL(config)
	default:
		return nil, fmt.Errorf("不支持的数据库类型: %s", config.Type)
	}

	if err != nil {
		return nil, fmt.Errorf("数据库连接失败: %v", err)
	}

	// 自动迁移数据库表
	err = db.AutoMigrate(
		&models.Server{},
		&models.Device{},
		&models.DeviceServer{},
		&models.User{},
	)
	if err != nil {
		return nil, fmt.Errorf("数据库迁移失败: %v", err)
	}

	// 创建默认管理员用户
	if err := createDefaultUser(db); err != nil {
		log.Printf("创建默认用户失败: %v", err)
	}

	log.Printf("数据库初始化成功 (%s)", config.Type)
	return db, nil
}

func initSQLite(config DatabaseConfig) (*gorm.DB, error) {
	if config.Path == "" {
		config.Path = "./data/emby.db"
	}

	// 确保数据目录存在
	dir := filepath.Dir(config.Path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("创建目录失败: %v", err)
	}

	dsn := config.Path
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		return nil, fmt.Errorf("SQLite连接失败: %v", err)
	}

	log.Printf("数据库文件: %s", config.Path)
	return db, nil
}

func initMySQL(config DatabaseConfig) (*gorm.DB, error) {
	host := config.Host
	if host == "" {
		host = "localhost"
	}
	port := config.Port
	if port == "" {
		port = "3306"
	}
	database := config.Database
	if database == "" {
		database = "emby_mgmt"
	}
	username := config.Username
	if username == "" {
		username = "root"
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		username, config.Password, host, port, database)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		return nil, fmt.Errorf("MySQL连接失败: %v", err)
	}

	log.Printf("MySQL服务器: %s:%s, 数据库: %s", host, port, database)
	return db, nil
}

func initPostgreSQL(config DatabaseConfig) (*gorm.DB, error) {
	host := config.Host
	if host == "" {
		host = "localhost"
	}
	port := config.Port
	if port == "" {
		port = "5432"
	}
	database := config.Database
	if database == "" {
		database = "emby_mgmt"
	}
	username := config.Username
	if username == "" {
		username = "postgres"
	}
	sslMode := config.SSLMode
	if sslMode == "" {
		sslMode = "disable"
	}

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=Asia/Shanghai",
		host, username, config.Password, database, port, sslMode)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		return nil, fmt.Errorf("PostgreSQL连接失败: %v", err)
	}

	log.Printf("PostgreSQL服务器: %s:%s, 数据库: %s", host, port, database)
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

	if err := db.Create(&admin).Error; err != nil {
		return err
	}

	log.Println("默认管理员账户已创建: admin/admin123")
	return nil
}

func TestConnection(config DatabaseConfig) error {
	var db *gorm.DB
	var err error

	switch strings.ToLower(config.Type) {
	case "sqlite", "":
		_, err = initSQLite(config)
	case "mysql":
		_, err = initMySQL(config)
	case "postgres", "postgresql":
		_, err = initPostgreSQL(config)
	default:
		return fmt.Errorf("不支持的数据库类型: %s", config.Type)
	}

	return err
}

func GetConfigs() []DatabaseConfig {
	return []DatabaseConfig{
		{
			Type: "sqlite",
			Path: "./data/emby.db",
		},
		{
			Type:     "mysql",
			Host:     "localhost",
			Port:     "3306",
			Database: "emby_mgmt",
			Username: "root",
			Password: "",
		},
		{
			Type:     "postgres",
			Host:     "localhost",
			Port:     "5432",
			Database: "emby_mgmt",
			Username: "postgres",
			Password: "",
			SSLMode:  "disable",
		},
	}
}