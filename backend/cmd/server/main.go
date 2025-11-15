package main

import (
	"fmt"
	"log"

	"github.com/emby-client-go/backend/internal/config"
	"github.com/emby-client-go/backend/internal/database"
	"github.com/emby-client-go/backend/internal/handlers"
	"github.com/emby-client-go/backend/pkg/websocket"
	"github.com/gin-gonic/gin"
)

// @title Emby Manager API
// @version 1.0
// @description 统一的Emby服务器管理平台API
// @host localhost:8080
// @BasePath /api
func main() {
	// 初始化配置
	config.Init()

	// 初始化数据库
	if err := database.Init(); err != nil {
		log.Fatal("数据库初始化失败:", err)
	}
	defer database.Close()

	// 初始化WebSocket Hub
	hub := websocket.NewHub()
	go hub.Run()
	log.Println("WebSocket Hub已启动")

	// 初始化WebSocket Manager
	wsManager := websocket.NewManager(hub)
	log.Println("WebSocket Manager已初始化")

	// 设置Gin模式
	if config.AppConfig.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 创建Gin引擎
	r := gin.Default()

	// 设置路由
	handlers.SetupRoutes(r, hub, wsManager)

	// 启动服务器
	addr := fmt.Sprintf("%s:%d", config.AppConfig.Server.Host, config.AppConfig.Server.Port)
	log.Printf("服务器启动在: %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatal("服务器启动失败:", err)
	}
}