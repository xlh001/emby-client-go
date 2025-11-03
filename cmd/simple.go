package main

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// 简化版数据结构
type Device struct {
	ID         uint   `json:"id"`
	Name       string `json:"name"`
	Identifier string `json:"identifier"`
	Platform   string `json:"platform"`
	IP         string `json:"ip_address"`
	IsActive   bool   `json:"is_active"`
}

type Server struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	URL         string `json:"url"`
	Version     string `json:"version"`
	IsActive    bool   `json:"is_active"`
}

type User struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Password string `json:"password,omitempty"`
	Role     string `json:"role"`
}

// 模拟数据存储
var mockDevices = []Device{
	{ID: 1, Name: "客厅电视", Identifier: "tv-living-001", Platform: "Android TV", IP: "192.168.1.100", IsActive: true},
	{ID: 2, Name: "卧室手机", Identifier: "mobile-bed-002", Platform: "Android", IP: "192.168.1.101", IsActive: true},
	{ID: 3, Name: "iPad Pro", Identifier: "ipad-pro-003", Platform: "iOS", IP: "192.168.1.102", IsActive: false},
}

var mockServers = []Server{
	{ID: 1, Name: "主要Emby服务器", URL: "http://emby1.example.com:8096", Version: "4.7.0.0", IsActive: true},
	{ID: 2, Name: "备用Emby服务器", URL: "http://emby2.example.com:8096", Version: "4.6.4.0", IsActive: true},
}

func main() {
	r := gin.Default()

	// CORS配置
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": "emby-client-go",
			"mode":    "demo",
		})
	})

	// API路由
	api := r.Group("/api")

	// 认证接口
	auth := api.Group("/auth")
	{
		auth.POST("/login", func(c *gin.Context) {
			var req struct {
				Username string `json:"username"`
				Password string `json:"password"`
			}

			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
				return
			}

			if req.Username == "admin" && req.Password == "admin123" {
				c.JSON(http.StatusOK, gin.H{
					"token": "demo-jwt-token",
					"expires_at": 1640995200,
					"user": User{
						ID: 1,
						Username: "admin",
						Role: "admin",
					},
				})
			} else {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "用户名或密码错误"})
			}
		})
	}

	// 设备接口
	devices := api.Group("/devices")
	{
		devices.GET("", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"data": mockDevices})
		})

		devices.GET("/:id", func(c *gin.Context) {
			id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
			for _, device := range mockDevices {
				if device.ID == uint(id) {
					c.JSON(http.StatusOK, gin.H{"data": device})
					return
				}
			}
			c.JSON(http.StatusNotFound, gin.H{"error": "设备不存在"})
		})

		devices.POST("", func(c *gin.Context) {
			var device Device
			if err := c.ShouldBindJSON(&device); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
				return
			}

			device.ID = uint(len(mockDevices) + 1)
			mockDevices = append(mockDevices, device)

			c.JSON(http.StatusCreated, gin.H{
				"message": "设备添加成功",
				"data":    device,
			})
		})

		devices.DELETE("/:id", func(c *gin.Context) {
			id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
			for i, device := range mockDevices {
				if device.ID == uint(id) {
					mockDevices = append(mockDevices[:i], mockDevices[i+1:]...)
					c.JSON(http.StatusOK, gin.H{"message": "设备删除成功"})
					return
				}
			}
			c.JSON(http.StatusNotFound, gin.H{"error": "设备不存在"})
		})
	}

	// 服务器接口
	servers := api.Group("/servers")
	{
		servers.GET("", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"data": mockServers})
		})

		servers.GET("/:id", func(c *gin.Context) {
			id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
			for _, server := range mockServers {
				if server.ID == uint(id) {
					c.JSON(http.StatusOK, gin.H{"data": server})
					return
				}
			}
			c.JSON(http.StatusNotFound, gin.H{"error": "服务器不存在"})
		})

		servers.POST("", func(c *gin.Context) {
			var server Server
			if err := c.ShouldBindJSON(&server); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
				return
			}

			server.ID = uint(len(mockServers) + 1)
			mockServers = append(mockServers, server)

			c.JSON(http.StatusCreated, gin.H{
				"message": "服务器添加成功",
				"data":    server,
			})
		})

		servers.POST("/:id/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "连接测试成功"})
		})
	}

	// 静态文件
	r.Static("/static", "./web/static")

	// 主页面
	r.LoadHTMLGlob("web/templates/*")
	r.GET("/", func(c *gin.Context) {
		c.HTML(200, "index.html", gin.H{
			"test_mode": true,
			"server_ip": "localhost:8080",
		})
	})

	// 启动信息
	fmt.Println("🚀 Emby管理服务启动成功！")
	fmt.Println("🌐 服务地址: http://localhost:8080")
	fmt.Println("👤 默认管理员账户: admin / admin123")
	fmt.Println("📊 当前模式: 演示模式（内存存储）")
	fmt.Println("💡 API端点: http://localhost:8080/api")
	fmt.Println("==================================================")

	r.Run(":8080")
}