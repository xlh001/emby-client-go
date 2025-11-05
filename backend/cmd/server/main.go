package main

import (
	"fmt"
	"log"

	"emby-manager/internal/config"
	"emby-manager/internal/handlers"
	"emby-manager/internal/middleware"
	"emby-manager/pkg/database"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title Emby Manager API
// @version 1.0
// @description Emby服务器管理平台API文档
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /api/v1

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
func main() {
	// 加载配置
	config.Init()

	// 初始化数据库
	if err := database.Init(); err != nil {
		log.Fatalf("数据库初始化失败: %v", err)
	}

	// 初始化服务和处理器
	handlers.InitAuthServices()

	// 设置Gin模式
	gin.SetMode(config.AppConfig.Server.Mode)

	// 创建路由
	r := gin.New()

	// 全局中间件
	r.Use(middleware.Logger())
	r.Use(middleware.Recovery())
	r.Use(middleware.CORS())

	// 健康检查
	r.GET("/health", handlers.Health)

	// API路由组
	api := r.Group("/api/v1")
	{
		// 无需认证的路由
		public := api.Group("/")
		{
			public.POST("/auth/login", handlers.Login)
			public.POST("/auth/register", handlers.Register)
			public.POST("/auth/refresh", handlers.RefreshToken)
		}

		// 需要认证的路由
		protected := api.Group("/")
		protected.Use(middleware.JWTAuth())
		{
			// 用户管理
			protected.GET("/users", handlers.GetUsers)
			protected.GET("/users/:id", handlers.GetUser)
			protected.PUT("/users/:id", handlers.UpdateUser)
			protected.DELETE("/users/:id", handlers.DeleteUser)

			// Emby服务器管理
			protected.GET("/emby/servers", handlers.GetEmbyServers)
			protected.POST("/emby/servers", handlers.CreateEmbyServer)
			protected.PUT("/emby/servers/:id", handlers.UpdateEmbyServer)
			protected.DELETE("/emby/servers/:id", handlers.DeleteEmbyServer)
			protected.POST("/emby/servers/:id/test", handlers.TestEmbyConnection)

			// 媒体库管理
			protected.GET("/media/libraries", handlers.GetMediaLibraries)
			protected.GET("/media/search", handlers.SearchMedia)
			protected.GET("/media/:id", handlers.GetMediaItem)

			// 播放控制
			protected.POST("/playback/control", handlers.PlaybackControl)
			protected.POST("/playback/progress", handlers.UpdateProgress)
			protected.GET("/playback/history", handlers.GetPlaybackHistory)

			// 设备管理
			protected.GET("/devices", handlers.GetDevices)
			protected.POST("/devices/:id/status", handlers.UpdateDeviceStatus)
		}
	}

	// Swagger文档
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// 启动服务器
	host := config.AppConfig.Server.Host
	port := config.AppConfig.Server.Port
	log.Printf("服务器启动在 http://%s:%d", host, port)
	log.Printf("API文档地址: http://%s:%d/swagger/index.html", host, port)

	if err := r.Run(fmt.Sprintf("%s:%d", host, port)); err != nil {
		log.Fatalf("服务器启动失败: %v", err)
	}
}