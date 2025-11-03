package main

import (
	"log"
	"fmt"

	"emby-client-go/internal/config"
	"emby-client-go/internal/handlers"
	"emby-client-go/internal/services"
	"emby-client-go/internal/models"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// 模拟数据库，避免CGO依赖
var (
	devices  = make(map[uint]*models.Device)
	servers  = make(map[uint]*models.Server)
	users    = make(map[uint]*models.User)
	nextID   = uint(1)
)

func main() {
	// 加载配置
	cfg := config.Load()

	// 初始化模拟数据
	initMockData()

	// 初始化服务
	deviceService := &MockDeviceService{}
	serverService := &MockServerService{}
	authService := services.NewAuthService(cfg.JWT.Secret)

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
		c.HTML(200, "index.html", gin.H{
			"test_mode": true,
		})
	})

	// 启动服务器
	port := cfg.Server.Port
	if port == "" {
		port = "8080"
	}

	fmt.Println("🚀 Emby管理服务启动成功！")
	fmt.Printf("🌐 服务地址: http://localhost:%s\n", port)
	fmt.Println("👤 默认管理员账户: admin / admin123")
	fmt.Println("📊 当前模式: 演示模式（内存存储）")
	fmt.Println("💡 提示: 生产环境请配置真实数据库")
	fmt.Println("================================")

	r.Run(":" + port)
}

// 初始化模拟数据
func initMockData() {
	// 创建默认管理员
	admin := &models.User{
		ID:       1,
		Username: "admin",
		Password: "$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi", // admin123
		Role:     "admin",
		IsActive: true,
	}
	users[1] = admin

	// 创建示例服务器
	server1 := &models.Server{
		ID:          1,
		Name:        "主要Emby服务器",
		URL:         "http://emby1.example.com:8096",
		Description: "主要的媒体服务器",
		IsActive:    true,
		Version:     "4.7.0.0",
	}
	servers[1] = server1

	server2 := &models.Server{
		ID:          2,
		Name:        "备用Emby服务器",
		URL:         "http://emby2.example.com:8096",
		Description: "备用媒体服务器",
		IsActive:    true,
		Version:     "4.6.4.0",
	}
	servers[2] = server2

	// 创建示例设备
	device1 := &models.Device{
		ID:         1,
		Name:       "客厅电视",
		Identifier: "tv-living-room-001",
		Platform:   "Android TV",
		Version:    "2.0.1",
		IPAddress:  "192.168.1.100",
		IsActive:   true,
	}
	devices[1] = device1

	device2 := &models.Device{
		ID:         2,
		Name:       "卧室手机",
		Identifier: "mobile-bedroom-002",
		Platform:   "Android",
		Version:    "1.8.0",
		IPAddress:  "192.168.1.101",
		IsActive:   true,
	}
	devices[2] = device2

	device3 := &models.Device{
		ID:         3,
		Name:       "iPad Pro",
		Identifier: "ipad-pro-003",
		Platform:   "iOS",
		Version:    "3.2.1",
		IPAddress:  "192.168.1.102",
		IsActive:   false, // 示例非活跃设备
	}
	devices[3] = device3

	fmt.Println("✨ 模拟数据初始化完成:")
	fmt.Printf("   📺 已创建 %d 个设备\n", len(devices))
	fmt.Printf("   🖥️  已创建 %d 个服务器\n", len(servers))
	fmt.Printf("   👤 已创建管理员账户\n")
}

// Mock设备服务
type MockDeviceService struct{}

func (s *MockDeviceService) AddDevice(device *models.Device) error {
	device.ID = nextID
	nextID++
	devices[device.ID] = device
	return nil
}

func (s *MockDeviceService) GetDevices() ([]models.Device, error) {
	result := make([]models.Device, 0, len(devices))
	for _, device := range devices {
		result = append(result, *device)
	}
	return result, nil
}

func (s *MockDeviceService) GetDevice(id uint) (*models.Device, error) {
	if device, exists := devices[id]; exists {
		return device, nil
	}
	return nil, fmt.Errorf("设备不存在")
}

func (s *MockDeviceService) UpdateDevice(id uint, updates *models.Device) error {
	if _, exists := devices[id]; !exists {
		return fmt.Errorf("设备不存在")
	}
	updates.ID = id
	devices[id] = updates
	return nil
}

func (s *MockDeviceService) DeleteDevice(id uint) error {
	if _, exists := devices[id]; !exists {
		return fmt.Errorf("设备不存在")
	}
	delete(devices, id)
	return nil
}

func (s *MockDeviceService) GetDeviceServers(deviceID uint) ([]models.Server, error) {
	// 简化实现，返回所有服务器
	result := make([]models.Server, 0, len(servers))
	for _, server := range servers {
		result = append(result, *server)
	}
	return result, nil
}

func (s *MockDeviceService) AddDeviceToServer(deviceID, serverID uint, priority int) error {
	return nil
}

func (s *MockDeviceService) RemoveDeviceFromServer(deviceID, serverID uint) error {
	return nil
}

func (s *MockDeviceService) GetActiveDevices() ([]models.Device, error) {
	result := []models.Device{}
	for _, device := range devices {
		if device.IsActive {
			result = append(result, *device)
		}
	}
	return result, nil
}

func (s *MockDeviceService) GetInactiveDevices() ([]models.Device, error) {
	result := []models.Device{}
	for _, device := range devices {
		if !device.IsActive {
			result = append(result, *device)
		}
	}
	return result, nil
}

// Mock服务器服务
type MockServerService struct{}

func (s *MockServerService) AddServer(server *models.Server) error {
	server.ID = nextID
	nextID++
	servers[server.ID] = server
	return nil
}

func (s *MockServerService) GetServers() ([]models.Server, error) {
	result := make([]models.Server, 0, len(servers))
	for _, server := range servers {
		result = append(result, *server)
	}
	return result, nil
}

func (s *MockServerService) GetServer(id uint) (*models.Server, error) {
	if server, exists := servers[id]; exists {
		return server, nil
	}
	return nil, fmt.Errorf("服务器不存在")
}

func (s *MockServerService) UpdateServer(id uint, updates *models.Server) error {
	if _, exists := servers[id]; !exists {
		return fmt.Errorf("服务器不存在")
	}
	updates.ID = id
	servers[id] = updates
	return nil
}

func (s *MockServerService) DeleteServer(id uint) error {
	if _, exists := servers[id]; !exists {
		return fmt.Errorf("服务器不存在")
	}
	delete(servers, id)
	return nil
}

func (s *MockServerService) TestConnection(id uint) error {
	// 模拟连接测试，总是返回成功
	return nil
}

func (s *MockServerService) GetServerDevices(serverID uint) ([]EmbyDevice, error) {
	return []EmbyDevice{}, nil
}

func (s *MockServerService) SyncDevicesFromServer(serverID uint, deviceService interface{}) error {
	return nil
}