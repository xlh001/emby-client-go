package main

import (
	"emby-client-go/internal/config"
	"emby-client-go/internal/database"
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "emby-db",
		Short: "Emby管理系统数据库初始化工具",
	}

	var initCmd = &cobra.Command{
		Use:   "init",
		Short: "初始化数据库",
		Run: func(cmd *cobra.Command, args []string) {
			cfg := config.Load()

			fmt.Printf("🗄️  正在初始化数据库 (%s)...\n", cfg.Database.Type)

			db, err := database.Initialize(cfg.Database)
			if err != nil {
				log.Fatalf("❌ 数据库初始化失败: %v", err)
			}

			fmt.Printf("✅ 数据库初始化成功\n")
			fmt.Printf("👤 默认管理员账户: admin / admin123\n")

			// 关闭数据库连接
			sqlDB, _ := db.DB()
			sqlDB.Close()
		},
	}

	var testCmd = &cobra.Command{
		Use:   "test",
		Short: "测试数据库连接",
		Run: func(cmd *cobra.Command, args []string) {
			cfg := config.Load()

			fmt.Printf("🔍 测试数据库连接 (%s)...\n", cfg.Database.Type)

			err := database.TestConnection(cfg.Database)
			if err != nil {
				log.Fatalf("❌ 数据库连接测试失败: %v", err)
			}

			fmt.Printf("✅ 数据库连接测试成功\n")
		},
	}

	var configCmd = &cobra.Command{
		Use:   "config",
		Short: "查看当前数据库配置",
		Run: func(cmd *cobra.Command, args []string) {
			cfg := config.Load()

			fmt.Printf("📋 当前数据库配置:\n")
			fmt.Printf("   类型: %s\n", cfg.Database.Type)
			if cfg.Database.Type == "sqlite" {
				fmt.Printf("   文件路径: %s\n", cfg.Database.Path)
			} else {
				fmt.Printf("   主机: %s\n", cfg.Database.Host)
				fmt.Printf("   端口: %s\n", cfg.Database.Port)
				fmt.Printf("   数据库: %s\n", cfg.Database.Database)
				fmt.Printf("   用户名: %s\n", cfg.Database.Username)
				if cfg.Database.Type == "postgres" {
					fmt.Printf("   SSL模式: %s\n", cfg.Database.SSLMode)
				}
			}
			fmt.Printf("   JWT密钥: %s\n", cfg.JWT.Secret)
		},
	}

	var listCmd = &cobra.Command{
		Use:   "list",
		Short: "列出可用的数据库配置模板",
		Run: func(cmd *cobra.Command, args []string) {
			configs := database.GetConfigs()

			fmt.Printf("📋 可用的数据库配置模板:\n\n")
			for i, config := range configs {
				fmt.Printf("%d. %s\n", i+1, config.Type)
				if config.Type == "sqlite" {
					fmt.Printf("   文件路径: %s\n", config.Path)
				} else {
					fmt.Printf("   主机: %s\n", config.Host)
					fmt.Printf("   端口: %s\n", config.Port)
					fmt.Printf("   数据库: %s\n", config.Database)
					fmt.Printf("   用户名: %s\n", config.Username)
					if config.Type == "postgres" {
						fmt.Printf("   SSL模式: %s\n", config.SSLMode)
					}
				}
				fmt.Println()
			}
		},
	}

	var setupCmd = &cobra.Command{
		Use:   "setup [type]",
		Short: "交互式设置数据库配置",
		Run: func(cmd *cobra.Command, args []string) {
			var dbType string
			if len(args) > 0 {
				dbType = args[0]
			} else {
				fmt.Printf("请选择数据库类型 (sqlite/mysql/postgres): ")
				fmt.Scanln(&dbType)
			}

			cfg := config.Load()
			cfg.Database.Type = dbType

			switch dbType {
			case "sqlite":
				var path string
				fmt.Printf("SQLite文件路径 (默认: ./data/emby.db): ")
				fmt.Scanln(&path)
				if path != "" {
					cfg.Database.Path = path
				}

			case "mysql":
				var host, port, database, username, password string
				fmt.Printf("主机 (默认: localhost): ")
				fmt.Scanln(&host)
				if host != "" {
					cfg.Database.Host = host
				}

				fmt.Printf("端口 (默认: 3306): ")
				fmt.Scanln(&port)
				if port != "" {
					cfg.Database.Port = port
				}

				fmt.Printf("数据库名 (默认: emby_mgmt): ")
				fmt.Scanln(&database)
				if database != "" {
					cfg.Database.Database = database
				}

				fmt.Printf("用户名 (默认: root): ")
				fmt.Scanln(&username)
				if username != "" {
					cfg.Database.Username = username
				}

				fmt.Printf("密码: ")
				fmt.Scanln(&password)
				cfg.Database.Password = password

			case "postgres":
				var host, port, database, username, password, sslmode string
				fmt.Printf("主机 (默认: localhost): ")
				fmt.Scanln(&host)
				if host != "" {
					cfg.Database.Host = host
				}

				fmt.Printf("端口 (默认: 5432): ")
				fmt.Scanln(&port)
				if port != "" {
					cfg.Database.Port = port
				}

				fmt.Printf("数据库名 (默认: emby_mgmt): ")
				fmt.Scanln(&database)
				if database != "" {
					cfg.Database.Database = database
				}

				fmt.Printf("用户名 (默认: postgres): ")
				fmt.Scanln(&username)
				if username != "" {
					cfg.Database.Username = username
				}

				fmt.Printf("密码: ")
				fmt.Scanln(&password)
				cfg.Database.Password = password

				fmt.Printf("SSL模式 (默认: disable): ")
				fmt.Scanln(&sslmode)
				if sslmode != "" {
					cfg.Database.SSLMode = sslmode
				}

			default:
				log.Fatalf("❌ 不支持的数据库类型: %s", dbType)
			}

			// 保存配置
			err := config.Save(cfg)
			if err != nil {
				log.Fatalf("❌ 保存配置失败: %v", err)
			}

			fmt.Printf("✅ 配置已保存到 config.yaml\n")
		},
	}

	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(testCmd)
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(setupCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}