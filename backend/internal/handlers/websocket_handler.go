package handlers

import (
	"log"
	"net/http"
	"strconv"

	"github.com/emby-client-go/backend/pkg/websocket"
	"github.com/gin-gonic/gin"
	gorilla "github.com/gorilla/websocket"
)

var upgrader = gorilla.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// 生产环境应该检查Origin
		return true
	},
}

// WebSocketHandler WebSocket处理器
type WebSocketHandler struct {
	hub     *websocket.Hub
	manager *websocket.Manager
}

// NewWebSocketHandler 创建WebSocket处理器
func NewWebSocketHandler(hub *websocket.Hub, manager *websocket.Manager) *WebSocketHandler {
	return &WebSocketHandler{
		hub:     hub,
		manager: manager,
	}
}

// HandleWebSocket 处理WebSocket连接
// @Summary WebSocket连接
// @Description 建立WebSocket连接用于实时通信
// @Tags WebSocket
// @Security BearerAuth
// @Param server_id query string false "服务器ID（可选）"
// @Success 101 {string} string "Switching Protocols"
// @Failure 400 {object} map[string]interface{} "请求错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Failure 500 {object} map[string]interface{} "服务器错误"
// @Router /ws [get]
func (h *WebSocketHandler) HandleWebSocket(c *gin.Context) {
	// 从上下文获取用户ID（由认证中间件设置）
	userIDValue, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	userID, ok := userIDValue.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "用户ID类型错误"})
		return
	}

	// 获取可选的服务器ID
	serverID := c.Query("server_id")

	// 升级HTTP连接为WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket升级失败: %v", err)
		return
	}

	// 创建客户端
	client := &websocket.Client{
		ID:       generateClientID(userID, serverID),
		UserID:   userID,
		ServerID: serverID,
		Conn:     conn,
		Send:     make(chan websocket.Message, 256),
		Manager:  h.hub,
	}

	// 注册客户端
	h.hub.Register <- client

	// 启动读写循环
	go client.WritePump()
	go client.ReadPump()

	log.Printf("WebSocket连接已建立: 用户=%d, 服务器=%s", userID, serverID)
}

// GetConnectionStatus 获取连接状态
// @Summary 获取WebSocket连接状态
// @Description 获取所有Emby服务器的WebSocket连接状态
// @Tags WebSocket
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "连接状态"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Router /api/ws/status [get]
func (h *WebSocketHandler) GetConnectionStatus(c *gin.Context) {
	status := h.manager.GetConnectionStatus()
	clientInfo := h.hub.GetClientInfo()

	c.JSON(http.StatusOK, gin.H{
		"emby_connections": status,
		"client_info":      clientInfo,
	})
}

// GetServerConnection 获取特定服务器的连接信息
// @Summary 获取服务器连接信息
// @Description 获取特定Emby服务器的WebSocket连接详细信息
// @Tags WebSocket
// @Security BearerAuth
// @Param id path string true "服务器ID"
// @Success 200 {object} map[string]interface{} "连接信息"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Failure 404 {object} map[string]interface{} "连接不存在"
// @Router /api/ws/server/:id [get]
func (h *WebSocketHandler) GetServerConnection(c *gin.Context) {
	serverID := c.Param("id")

	conn, exists := h.manager.GetConnection(serverID)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "连接不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"server_id":       conn.ServerID,
		"status":          conn.GetStatus().String(),
		"last_connected":  conn.LastConnected,
		"reconnect_count": conn.ReconnectCount,
		"max_reconnects":  conn.MaxReconnects,
	})
}

// ReconnectServer 重连服务器
// @Summary 重连Emby服务器
// @Description 强制重连指定的Emby服务器WebSocket连接
// @Tags WebSocket
// @Security BearerAuth
// @Param id path string true "服务器ID"
// @Success 200 {object} map[string]interface{} "重连成功"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Failure 404 {object} map[string]interface{} "连接不存在"
// @Router /api/ws/server/:id/reconnect [post]
func (h *WebSocketHandler) ReconnectServer(c *gin.Context) {
	serverID := c.Param("id")

	conn, exists := h.manager.GetConnection(serverID)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "连接不存在"})
		return
	}

	go conn.ForceReconnect()

	c.JSON(http.StatusOK, gin.H{
		"message": "重连请求已发送",
	})
}

// generateClientID 生成客户端ID
func generateClientID(userID uint, serverID string) string {
	if serverID != "" {
		return "user_" + strconv.FormatUint(uint64(userID), 10) + "_server_" + serverID
	}
	return "user_" + strconv.FormatUint(uint64(userID), 10)
}
