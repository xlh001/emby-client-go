package handlers

import (
	"emby-client-go/internal/models"
	"emby-client-go/internal/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// Handler API处理器
type Handler struct {
	deviceService  *services.DeviceService
	serverService  *services.ServerService
	authService    *services.AuthService
}

// NewHandler 创建处理器
func NewHandler(deviceService *services.DeviceService, serverService *services.ServerService, authService *services.AuthService) *Handler {
	return &Handler{
		deviceService: deviceService,
		serverService: serverService,
		authService:   authService,
	}
}

// RegisterRoutes 注册路由
func (h *Handler) RegisterRoutes(r *gin.Engine) {
	// 健康检查
	r.GET("/health", h.Health)

	// 认证路由
	auth := r.Group("/api/auth")
	{
		auth.POST("/login", h.Login)
		auth.POST("/register", h.Register)
		auth.POST("/change-password", h.RequireAuth(), h.ChangePassword)
	}

	// 设备管理路由
	devices := r.Group("/api/devices")
	devices.Use(h.RequireAuth())
	{
		devices.GET("", h.GetDevices)
		devices.GET("/:id", h.GetDevice)
		devices.POST("", h.RequireAdmin(), h.AddDevice)
		devices.PUT("/:id", h.RequireAdmin(), h.UpdateDevice)
		devices.DELETE("/:id", h.RequireAdmin(), h.DeleteDevice)
		devices.GET("/:id/servers", h.GetDeviceServers)
		devices.POST("/:id/servers/:serverId", h.RequireAdmin(), h.AddDeviceToServer)
		devices.DELETE("/:id/servers/:serverId", h.RequireAdmin(), h.RemoveDeviceFromServer)
		devices.GET("/active", h.GetActiveDevices)
		devices.GET("/inactive", h.GetInactiveDevices)
	}

	// 服务器管理路由
	servers := r.Group("/api/servers")
	servers.Use(h.RequireAuth())
	{
		servers.GET("", h.GetServers)
		servers.GET("/:id", h.GetServer)
		servers.POST("", h.RequireAdmin(), h.AddServer)
		servers.PUT("/:id", h.RequireAdmin(), h.UpdateServer)
		servers.DELETE("/:id", h.RequireAdmin(), h.DeleteServer)
		servers.POST("/:id/test", h.RequireAdmin(), h.TestServerConnection)
		servers.GET("/:id/devices", h.RequireAdmin(), h.GetServerDevices)
		servers.POST("/:id/sync-devices", h.RequireAdmin(), h.SyncDevices)
	}
}

// Health 健康检查
func (h *Handler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"service": "emby-client-go",
	})
}

// Login 登录
func (h *Handler) Login(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	response, err := h.authService.Login(req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// Register 注册
func (h *Handler) Register(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
		Email    string `json:"email"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	user := models.User{
		Username: req.Username,
		Password: req.Password,
		Email:    req.Email,
	}

	err := h.authService.Register(&user)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 清除密码字段
	user.Password = ""
	c.JSON(http.StatusCreated, gin.H{
		"message": "用户注册成功",
		"user":    user,
	})
}

// ChangePassword 修改密码
func (h *Handler) ChangePassword(c *gin.Context) {
	user := c.MustGet("user").(*models.User)

	var req struct {
		OldPassword string `json:"old_password" binding:"required"`
		NewPassword string `json:"new_password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	err := h.authService.ChangePassword(user.ID, req.OldPassword, req.NewPassword)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "密码修改成功"})
}

// GetDevices 获取设备列表
func (h *Handler) GetDevices(c *gin.Context) {
	devices, err := h.deviceService.GetDevices()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取设备列表失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": devices})
}

// GetDevice 获取单个设备
func (h *Handler) GetDevice(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的设备ID"})
		return
	}

	device, err := h.deviceService.GetDevice(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "设备不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": device})
}

// AddDevice 添加设备
func (h *Handler) AddDevice(c *gin.Context) {
	var device models.Device
	if err := c.ShouldBindJSON(&device); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	err := h.deviceService.AddDevice(&device)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "设备添加成功",
		"data":    device,
	})
}

// UpdateDevice 更新设备
func (h *Handler) UpdateDevice(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的设备ID"})
		return
	}

	var updates models.Device
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	err = h.deviceService.UpdateDevice(uint(id), &updates)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新设备失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "设备更新成功"})
}

// DeleteDevice 删除设备
func (h *Handler) DeleteDevice(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的设备ID"})
		return
	}

	err = h.deviceService.DeleteDevice(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除设备失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "设备删除成功"})
}

// GetDeviceServers 获取设备关联的服务器
func (h *Handler) GetDeviceServers(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的设备ID"})
		return
	}

	servers, err := h.deviceService.GetDeviceServers(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取设备服务器失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": servers})
}

// AddDeviceToServer 添加设备到服务器
func (h *Handler) AddDeviceToServer(c *gin.Context) {
	deviceID, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	serverID, _ := strconv.ParseUint(c.Param("serverId"), 10, 32)

	var req struct {
		Priority int `json:"priority"`
	}

	c.ShouldBindJSON(&req)

	err := h.deviceService.AddDeviceToServer(uint(deviceID), uint(serverID), req.Priority)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "设备关联服务器成功"})
}

// RemoveDeviceFromServer 移除设备-服务器关联
func (h *Handler) RemoveDeviceFromServer(c *gin.Context) {
	deviceID, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	serverID, _ := strconv.ParseUint(c.Param("serverId"), 10, 32)

	err := h.deviceService.RemoveDeviceFromServer(uint(deviceID), uint(serverID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "移除关联失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "移除设备关联成功"})
}

// GetActiveDevices 获取活跃设备
func (h *Handler) GetActiveDevices(c *gin.Context) {
	devices, err := h.deviceService.GetActiveDevices()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取活跃设备失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": devices})
}

// GetInactiveDevices 获取非活跃设备
func (h *Handler) GetInactiveDevices(c *gin.Context) {
	devices, err := h.deviceService.GetInactiveDevices()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取非活跃设备失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": devices})
}