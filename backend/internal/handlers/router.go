package handlers

import (
	"github.com/emby-client-go/backend/internal/middleware"
	"github.com/emby-client-go/backend/pkg/websocket"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// SetupRoutes 设置路由
func SetupRoutes(r *gin.Engine, hub *websocket.Hub, wsManager *websocket.Manager) {
	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Swagger文档
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// 创建处理器
	userHandler := NewUserHandler()
	serverHandler := NewServerHandler()
	wsHandler := NewWebSocketHandler(hub, wsManager)
	mediaHandler := NewMediaHandler()
	searchHandler := NewSearchHandler()

	// API路由组
	api := r.Group("/api")
	{
		// 认证路由（无需认证）
		auth := api.Group("/auth")
		{
			auth.POST("/register", userHandler.Register)
			auth.POST("/login", userHandler.Login)
			auth.POST("/refresh", userHandler.RefreshToken)
			// 登出需要认证
			auth.POST("/logout", middleware.AuthMiddleware(), userHandler.Logout)
		}

		// 用户路由（需要认证）
		user := api.Group("/user")
		user.Use(middleware.AuthMiddleware())
		{
			user.GET("/profile", userHandler.GetProfile)
			user.POST("/change-password", userHandler.ChangePassword)

			// 管理员权限路由
			admin := user.Group("")
			admin.Use(middleware.AdminMiddleware())
			{
				admin.GET("/list", userHandler.GetUsers)
			}
		}

		// 服务器管理路由（需要认证）
		server := api.Group("/server")
		server.Use(middleware.AuthMiddleware())
		{
			server.POST("/create", serverHandler.CreateServer)
			server.GET("/list", serverHandler.GetServers)
			server.GET("/:id", serverHandler.GetServer)
			server.PUT("/:id", serverHandler.UpdateServer)
			server.DELETE("/:id", serverHandler.DeleteServer)
			server.POST("/:id/test", serverHandler.TestConnection)
			server.POST("/:id/sync-devices", serverHandler.SyncDevices)
			server.POST("/:id/sync-libraries", serverHandler.SyncLibraries)
		}

		// WebSocket路由（需要认证）
		ws := api.Group("/ws")
		ws.Use(middleware.AuthMiddleware())
		{
			ws.GET("/status", wsHandler.GetConnectionStatus)
			ws.GET("/server/:id", wsHandler.GetServerConnection)
			ws.POST("/server/:id/reconnect", wsHandler.ReconnectServer)
		}

		// 媒体库路由（需要认证）
		media := api.Group("/media")
		media.Use(middleware.AuthMiddleware())
		{
			media.GET("/libraries", mediaHandler.GetMediaLibraries)
			media.GET("/libraries/:id", mediaHandler.GetMediaLibrary)
			media.POST("/sync/:id", mediaHandler.SyncMediaLibraries)
			media.POST("/sync-all", mediaHandler.SyncAllServers)
			media.POST("/libraries/:id/refresh", mediaHandler.RefreshMediaLibrary)
			media.GET("/stats", mediaHandler.GetMediaLibraryStats)
			media.GET("/items", mediaHandler.GetMediaItems)
			media.GET("/items/:id", mediaHandler.GetMediaItem)
		}

		// 搜索路由（需要认证）
		search := api.Group("/search")
		search.Use(middleware.AuthMiddleware())
		{
			search.GET("", searchHandler.SearchMedia)                    // /api/search
			search.GET("/suggestions", searchHandler.GetSearchSuggestions)  // /api/search/suggestions
			search.GET("/popular", searchHandler.GetPopularKeywords)       // /api/search/popular
			search.GET("/history", searchHandler.GetSearchHistory)        // /api/search/history
			search.GET("/stats", searchHandler.GetSearchStats)          // /api/search/stats
		}

		// 播放控制路由（需要认证）
		playback := api.Group("/playback")
		playback.Use(middleware.AuthMiddleware())
		{
			playbackHandler := NewPlaybackHandler()
			playback.POST("/:server_id/:device_id/command", playbackHandler.SendPlayCommand)
			playback.GET("/sessions", playbackHandler.GetActiveSessions)
			playback.GET("/history", playbackHandler.GetPlaybackHistory)
		}
	}

	// WebSocket连接端点（需要认证）
	r.GET("/ws", middleware.AuthMiddleware(), wsHandler.HandleWebSocket)
}