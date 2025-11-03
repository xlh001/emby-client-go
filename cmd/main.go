package main

import (
	"log"
	"os"
	"path/filepath"

	"emby-client-go/internal/config"
	"emby-client-go/internal/database"
	"emby-client-go/internal/handlers"
	"emby-client-go/internal/services"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// 加载配置
	cfg := config.Load()

	// 确保数据目录存在
	if err := os.MkdirAll(filepath.Dir(cfg.Database.Path), 0755); err != nil {
		log.Fatal("创建数据目录失败:", err)
	}

	// 初始化数据库
	db, err := database.Initialize(cfg.Database.Path)
	if err != nil {
		log.Fatal("数据库初始化失败:", err)
	}

	// 初始化认证服务
	authService := services.NewAuthService(cfg.JWT.Secret)
	authService.SetDB(db)

	// 创建默认管理员用户（如果不存在）
	err = authService.CreateAdminUser(db, "admin", "admin123")
	if err != nil {
		log.Printf("创建默认管理员失败: %v", err)
	} else {
		log.Println("默认管理员账户已创建: admin/admin123")
	}

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

	// 设置静态文件和服务页面
	r.Static("/static", "./web/static")
	r.LoadHTMLGlob("web/templates/*")
	r.GET("/", func(c *gin.Context) {
		c.HTML(200, "index.html", gin.H{})
	})

	// 启动服务器
	log.Printf("Emby管理服务启动，端口: %s", cfg.Server.Port)
	log.Println("默认管理员账户: admin/admin123")
	log.Println("请访问 http://localhost:" + cfg.Server.Port)
	r.Run(":" + cfg.Server.Port)
}