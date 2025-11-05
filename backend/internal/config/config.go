package config

import (
	"fmt"
	"log"

	"github.com/spf13/viper"
)

// Config 应用配置结构
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Redis    RedisConfig    `mapstructure:"redis"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	Emby     EmbyConfig     `mapstructure:"emby"`
	Log      LogConfig      `mapstructure:"log"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	Mode         string `mapstructure:"mode"`
	ReadTimeout  int    `mapstructure:"read_timeout"`
	WriteTimeout int    `mapstructure:"write_timeout"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Type         string `mapstructure:"type"`
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	Username     string `mapstructure:"username"`
	Password     string `mapstructure:"password"`
	Database     string `mapstructure:"database"`
	SSLMode      string `mapstructure:"ssl_mode"`
	MaxIdleConns int    `mapstructure:"max_idle_conns"`
	MaxOpenConns int    `mapstructure:"max_open_conns"`
}

// RedisConfig Redis配置
type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

// JWTConfig JWT配置
type JWTConfig struct {
	Secret     string `mapstructure:"secret"`
	ExpireTime int    `mapstructure:"expire_time"`
	Issuer     string `mapstructure:"issuer"`
}

// EmbyConfig Emby相关配置
type EmbyConfig struct {
	DefaultTimeout int  `mapstructure:"default_timeout"`
	MaxRetryTimes  int  `mapstructure:"max_retry_times"`
	EnableCache    bool `mapstructure:"enable_cache"`
	CacheTTL       int  `mapstructure:"cache_ttl"`
}

// LogConfig 日志配置
type LogConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
	Output string `mapstructure:"output"`
}

var AppConfig *Config

// Init 初始化配置
func Init() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath(".")

	// 设置默认值
	setDefaults()

	// 读取环境变量
	viper.AutomaticEnv()

	// 读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Println("配置文件未找到，使用默认配置")
		} else {
			log.Fatalf("读取配置文件失败: %v", err)
		}
	}

	// 解析到结构体
	if err := viper.Unmarshal(&AppConfig); err != nil {
		log.Fatalf("解析配置失败: %v", err)
	}

	// 验证配置
	validateConfig()

	log.Printf("配置加载完成，服务器地址: %s:%d", AppConfig.Server.Host, AppConfig.Server.Port)
}

// setDefaults 设置默认配置值
func setDefaults() {
	// 服务器配置
	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.mode", "debug")
	viper.SetDefault("server.read_timeout", 30)
	viper.SetDefault("server.write_timeout", 30)

	// 数据库配置
	viper.SetDefault("database.type", "sqlite")
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.username", "")
	viper.SetDefault("database.password", "")
	viper.SetDefault("database.database", "./emby_manager.db")
	viper.SetDefault("database.ssh_mode", "disable")
	viper.SetDefault("database.max_idle_conns", 10)
	viper.SetDefault("database.max_open_conns", 100)

	// Redis配置
	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", 6379)
	viper.SetDefault("redis.password", "")
	viper.SetDefault("redis.db", 0)

	// JWT配置
	viper.SetDefault("jwt.secret", "emby_manager_secret_key")
	viper.SetDefault("jwt.expire_time", 86400) // 24小时
	viper.SetDefault("jwt.issuer", "emby-manager")

	// Emby配置
	viper.SetDefault("emby.default_timeout", 30)
	viper.SetDefault("emby.max_retry_times", 3)
	viper.SetDefault("emby.enable_cache", true)
	viper.SetDefault("emby.cache_ttl", 300) // 5分钟

	// 日志配置
	viper.SetDefault("log.level", "info")
	viper.SetDefault("log.format", "json")
	viper.SetDefault("log.output", "stdout")
}

// validateConfig 验证配置的有效性
func validateConfig() {
	if AppConfig.JWT.Secret == "" || AppConfig.JWT.Secret == "emby_manager_secret_key" {
		log.Println("警告：建议在生产环境中设置自定义的JWT密钥")
	}

	if AppConfig.Database.Type == "postgres" && AppConfig.Database.Host == "" {
		log.Fatal("PostgreSQL数据库配置不完整：缺少主机地址")
	}

	if AppConfig.Database.Type == "mysql" && AppConfig.Database.Host == "" {
		log.Fatal("MySQL数据库配置不完整：缺少主机地址")
	}
}

// GetDSN 获取数据库连接字符串
func (c DatabaseConfig) GetDSN() string {
	switch c.Type {
	case "sqlite":
		return c.Database
	case "postgres":
		return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			c.Host, c.Port, c.Username, c.Password, c.Database, c.SSLMode)
	case "mysql":
		return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			c.Username, c.Password, c.Host, c.Port, c.Database)
	default:
		log.Fatalf("不支持的数据库类型: %s", c.Type)
		return ""
	}
}

// GetRedisAddr 获取Redis连接地址
func (c RedisConfig) GetRedisAddr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}