package main

import (
	"emby-client-go/internal/config"
	"emby-client-go/internal/database"
	"emby-client-go/internal/handlers"
	"emby-client-go/internal/services"
	"fmt"
	"log"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// 加载配置
	cfg := config.Load()

	fmt.Printf("🚀 启动Emby管理系统...\n")
	fmt.Printf("📊 数据库类型: %s\n", cfg.Database.Type)

	// 初始化数据库
	db, err := database.Initialize(cfg.Database)
	if err != nil {
		log.Fatalf("❌ 数据库初始化失败: %v", err)
	}

	fmt.Printf("✅ 数据库连接成功\n")

	// 初始化认证服务
	authService := services.NewAuthService(cfg.JWT.Secret)
	authService.SetDB(db)

	// 初始化服务
	deviceService := services.NewDeviceService(db)
	serverService := services.NewServerService(db)

	// 初始化处理器
	r := gin.Default()

	// CORS配置
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// 初始化处理器
	handler := handlers.NewHandler(deviceService, serverService, authService)

	// 注册路由
	handler.RegisterRoutes(r)

	// 添加数据库相关的API
	r.GET("/api/database/info", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"type":     cfg.Database.Type,
			"configured": true,
		})
	})

	r.POST("/api/database/test", func(c *gin.Context) {
		var req struct {
			Type     string `json:"type" binding:"required"`
			Host     string `json:"host"`
			Port     string `json:"port"`
			Database string `json:"database"`
			Username string `json:"username"`
			Password string `json:"password"`
			Path     string `json:"path"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": "参数错误"})
			return
		}

		testConfig := database.DatabaseConfig{
			Type:     req.Type,
			Host:     req.Host,
			Port:     req.Port,
			Database: req.Database,
			Username: req.Username,
			Password: req.Password,
			Path:     req.Path,
		}

		if err := database.TestConnection(testConfig); err != nil {
			c.JSON(400, gin.H{"error": "连接失败: " + err.Error()})
			return
		}

		c.JSON(200, gin.H{"message": "连接测试成功"})
	})

	// 获取可用数据库配置模板
	r.GET("/api/database/configs", func(c *gin.Context) {
		configs := database.GetConfigs()
		c.JSON(200, gin.H{"data": configs})
	})

	// 设置静态文件和服务页面
	r.Static("/static", "./web/static")
	r.LoadHTMLGlob("web/templates/*")
	r.GET("/", func(c *gin.Context) {
		c.HTML(200, "index.html", gin.H{
			"db_type": cfg.Database.Type,
		})
	})

	// 启动服务器
	port := cfg.Server.Port
	if port == "" {
		port = "8080"
	}

	fmt.Printf("🌐 服务地址: http://localhost:%s\n", port)
	fmt.Printf("👤 默认管理员账户: admin / admin123\n")
	fmt.Println("==================================================")

	r.Run(":" + port)
}