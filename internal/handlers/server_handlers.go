package handlers

import (
	"emby-client-go/internal/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetServers 获取服务器列表
func (h *Handler) GetServers(c *gin.Context) {
	servers, err := h.serverService.GetServers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取服务器列表失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": servers})
}

// GetServer 获取单个服务器
func (h *Handler) GetServer(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的服务器ID"})
		return
	}

	server, err := h.serverService.GetServer(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "服务器不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": server})
}

// AddServer 添加服务器
func (h *Handler) AddServer(c *gin.Context) {
	var server models.Server
	if err := c.ShouldBindJSON(&server); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	err := h.serverService.AddServer(&server)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "服务器添加成功",
		"data":    server,
	})
}

// UpdateServer 更新服务器
func (h *Handler) UpdateServer(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的服务器ID"})
		return
	}

	var updates models.Server
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	err = h.serverService.UpdateServer(uint(id), &updates)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新服务器失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "服务器更新成功"})
}

// DeleteServer 删除服务器
func (h *Handler) DeleteServer(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的服务器ID"})
		return
	}

	err = h.serverService.DeleteServer(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除服务器失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "服务器删除成功"})
}

// TestServerConnection 测试服务器连接
func (h *Handler) TestServerConnection(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的服务器ID"})
		return
	}

	err = h.serverService.TestConnection(uint(id))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "连接测试失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "连接测试成功"})
}

// GetServerDevices 获取服务器设备列表
func (h *Handler) GetServerDevices(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的服务器ID"})
		return
	}

	devices, err := h.serverService.GetServerDevices(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取服务器设备失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": devices})
}

// SyncDevices 同步设备
func (h *Handler) SyncDevices(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的服务器ID"})
		return
	}

	err = h.serverService.SyncDevicesFromServer(uint(id), h.deviceService)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "同步设备失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "设备同步完成"})
}

// RequireAuth 认证中间件
func (h *Handler) RequireAuth() gin.HandlerFunc {
	return h.authService.RequireAuth()
}

// RequireAdmin 管理员权限中间件
func (h *Handler) RequireAdmin() gin.HandlerFunc {
	return h.authService.RequireAdmin()
}