package config

import (
	"emby-client-go/internal/database"
	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig     `mapstructure:"server"`
	Database database.DatabaseConfig `mapstructure:"database"`
	JWT      JWTConfig        `mapstructure:"jwt"`
}

type ServerConfig struct {
	Port string `mapstructure:"port"`
}

type JWTConfig struct {
	Secret     string `mapstructure:"secret"`
	ExpireTime int    `mapstructure:"expire_time"`
}

func Load() *Config {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")

	// 设置默认值
	viper.SetDefault("server.port", "8080")
	viper.SetDefault("database.type", "sqlite")
	viper.SetDefault("database.path", "./data/emby.db")
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", "3306")
	viper.SetDefault("database.database", "emby_mgmt")
	viper.SetDefault("database.username", "")
	viper.SetDefault("database.password", "")
	viper.SetDefault("database.ssl_mode", "disable")
	viper.SetDefault("jwt.secret", "your-secret-key-change-in-production")
	viper.SetDefault("jwt.expire_time", 24)

	// 读取配置文件
	err := viper.ReadInConfig()
	if err != nil {
		// 如果配置文件不存在，会使用默认值
	}

	var config Config
	viper.Unmarshal(&config)
	return &config
}

func Save(config *Config) error {
	viper.Set("server", config.Server)
	viper.Set("database", config.Database)
	viper.Set("jwt", config.JWT)
	return viper.WriteConfig()
}