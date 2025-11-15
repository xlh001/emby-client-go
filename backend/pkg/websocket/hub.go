package websocket

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Message WebSocket消息
type Message struct {
	Type      string      `json:"type"`      // 消息类型：system, server-status, device-update, library-update
	ServerID  string      `json:"server_id"` // 服务器ID（可选）
	Data      interface{} `json:"data"`      // 消息数据
	Timestamp time.Time   `json:"timestamp"` // 时间戳
}

// Client WebSocket客户端
type Client struct {
	ID       string              // 客户端ID
	UserID   uint                // 用户ID
	ServerID string              // 服务器ID（可选，用于服务器特定连接）
	Conn     *websocket.Conn     // WebSocket连接
	Send     chan Message         // 发送消息通道
	Manager  *Manager            // 管理器引用
	LastPing time.Time           // 最后心跳时间
	mutex    sync.RWMutex        // 读写锁
}

// Hub WebSocket连接中心
type Hub struct {
	// 注册新客户端
	Register chan *Client

	// 注销客户端
	Unregister chan *Client

	// 广播消息给所有客户端
	Broadcast chan Message

	// 广播消息给特定用户
	UserBroadcast chan UserMessage

	// 广播消息给特定服务器
	ServerBroadcast chan ServerMessage

	// 客户端列表
	clients map[string]*Client

	// 用户到客户端映射
	userClients map[uint]map[string]*Client

	// 服务器到客户端映射
	serverClients map[string]map[string]*Client

	// 读写锁
	mutex sync.RWMutex
}

// UserMessage 用户特定消息
type UserMessage struct {
	UserID  uint
	Message Message
}

// ServerMessage 服务器特定消息
type ServerMessage struct {
	ServerID string
	Message  Message
}

// NewHub 创建新的Hub
func NewHub() *Hub {
	return &Hub{
		Register:       make(chan *Client),
		Unregister:     make(chan *Client),
		Broadcast:      make(chan Message, 256),
		UserBroadcast:  make(chan UserMessage, 256),
		ServerBroadcast: make(chan ServerMessage, 256),
		clients:        make(map[string]*Client),
		userClients:    make(map[uint]map[string]*Client),
		serverClients:  make(map[string]map[string]*Client),
	}
}

// Run 启动Hub主循环
func (h *Hub) Run() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	go h.startPingChecker(ticker.C)

	for {
		select {
		case client := <-h.Register:
			h.registerClient(client)

		case client := <-h.Unregister:
			h.unregisterClient(client)

		case message := <-h.Broadcast:
			h.broadcastMessage(message)

		case userMsg := <-h.UserBroadcast:
			h.sendToUser(userMsg.UserID, userMsg.Message)

		case serverMsg := <-h.ServerBroadcast:
			h.sendToServer(serverMsg.ServerID, serverMsg.Message)
		}
	}
}

// registerClient 注册新客户端
func (h *Hub) registerClient(client *Client) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	h.clients[client.ID] = client

	// 添加到用户映射
	if _, exists := h.userClients[client.UserID]; !exists {
		h.userClients[client.UserID] = make(map[string]*Client)
	}
	h.userClients[client.UserID][client.ID] = client

	// 如果有服务器ID，添加到服务器映射
	if client.ServerID != "" {
		if _, exists := h.serverClients[client.ServerID]; !exists {
			h.serverClients[client.ServerID] = make(map[string]*Client)
		}
		h.serverClients[client.ServerID][client.ID] = client
	}

	log.Printf("客户端已注册: %s (用户: %d, 服务器: %s)", client.ID, client.UserID, client.ServerID)

	// 发送连接成功消息
	h.sendToClient(client, Message{
		Type:      "system",
		Data:      map[string]string{"status": "connected", "client_id": client.ID},
		Timestamp: time.Now(),
	})
}

// unregisterClient 注销客户端
func (h *Hub) unregisterClient(client *Client) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	if _, exists := h.clients[client.ID]; exists {
		delete(h.clients, client.ID)

		// 从用户映射中移除
		if userMap, exists := h.userClients[client.UserID]; exists {
			delete(userMap, client.ID)
			if len(userMap) == 0 {
				delete(h.userClients, client.UserID)
			}
		}

		// 从服务器映射中移除
		if client.ServerID != "" {
			if serverMap, exists := h.serverClients[client.ServerID]; exists {
				delete(serverMap, client.ID)
				if len(serverMap) == 0 {
					delete(h.serverClients, client.ServerID)
				}
			}
		}

		close(client.Send)
		log.Printf("客户端已注销: %s (用户: %d, 服务器: %s)", client.ID, client.UserID, client.ServerID)
	}
}

// broadcastMessage 广播消息给所有客户端
func (h *Hub) broadcastMessage(message Message) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	for _, client := range h.clients {
		select {
		case client.Send <- message:
		default:
			// 发送失败，关闭连接
			close(client.Send)
			delete(h.clients, client.ID)
		}
	}
}

// sendToUser 发送消息给特定用户的所有客户端
func (h *Hub) sendToUser(userID uint, message Message) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	if userMap, exists := h.userClients[userID]; exists {
		for _, client := range userMap {
			select {
			case client.Send <- message:
			default:
				// 发送失败，关闭连接
				close(client.Send)
				delete(h.clients, client.ID)
			}
		}
	}
}

// sendToServer 发送消息给特定服务器的所有客户端
func (h *Hub) sendToServer(serverID string, message Message) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	if serverMap, exists := h.serverClients[serverID]; exists {
		for _, client := range serverMap {
			select {
			case client.Send <- message:
			default:
				// 发送失败，关闭连接
				close(client.Send)
				delete(h.clients, client.ID)
			}
		}
	}
}

// sendToClient 发送消息给特定客户端
func (h *Hub) sendToClient(client *Client, message Message) {
	select {
	case client.Send <- message:
	default:
		// 发送失败，关闭连接
		h.mutex.Lock()
		if _, exists := h.clients[client.ID]; exists {
			delete(h.clients, client.ID)
		}
		h.mutex.Unlock()
		close(client.Send)
	}
}

// GetClientCount 获取客户端总数
func (h *Hub) GetClientCount() int {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	return len(h.clients)
}

// GetUserClientCount 获取特定用户的客户端数量
func (h *Hub) GetUserClientCount(userID uint) int {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	if userMap, exists := h.userClients[userID]; exists {
		return len(userMap)
	}
	return 0
}

// GetServerClientCount 获取特定服务器的客户端数量
func (h *Hub) GetServerClientCount(serverID string) int {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	if serverMap, exists := h.serverClients[serverID]; exists {
		return len(serverMap)
	}
	return 0
}

// GetClientInfo 获取客户端信息统计
func (h *Hub) GetClientInfo() map[string]interface{} {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	info := map[string]interface{}{
		"total_clients":     len(h.clients),
		"total_users":       len(h.userClients),
		"total_servers":     len(h.serverClients),
		"users":            make(map[string]int),
		"servers":          make(map[string]int),
	}

	for userID, clients := range h.userClients {
		info["users"].(map[string]int)[string(rune(userID)+'0')] = len(clients)
	}

	for serverID, clients := range h.serverClients {
		info["servers"].(map[string]int)[serverID] = len(clients)
	}

	return info
}

// startPingChecker 启动心跳检查器
func (h *Hub) startPingChecker(ticker <-chan time.Time) {
	for range ticker {
		h.mutex.RLock()
		deadClients := make([]*Client, 0)

		now := time.Now()
		for _, client := range h.clients {
			client.mutex.RLock()
			if now.Sub(client.LastPing) > 60*time.Second {
				deadClients = append(deadClients, client)
			}
			client.mutex.RUnlock()
		}
		h.mutex.RUnlock()

		// 清理死连接
		for _, client := range deadClients {
			client.Conn.Close()
			h.unregisterClient(client)
		}

		if len(deadClients) > 0 {
			log.Printf("清理了 %d 个死连接", len(deadClients))
		}
	}
}

// SendMessage 发送消息（支持不同类型）
func (h *Hub) SendMessage(msgType string, serverID string, userID uint, data interface{}) {
	message := Message{
		Type:      msgType,
		ServerID:  serverID,
		Data:      data,
		Timestamp: time.Now(),
	}

	if userID != 0 {
		h.UserBroadcast <- UserMessage{UserID: userID, Message: message}
	} else if serverID != "" {
		h.ServerBroadcast <- ServerMessage{ServerID: serverID, Message: message}
	} else {
		h.Broadcast <- message
	}
}

// SendServerStatus 发送服务器状态更新
func (h *Hub) SendServerStatus(serverID string, status map[string]interface{}) {
	h.SendMessage("server-status", serverID, 0, status)
}

// SendDeviceUpdate 发送设备更新通知
func (h *Hub) SendDeviceUpdate(serverID string, devices interface{}) {
	h.SendMessage("device-update", serverID, 0, devices)
}

// SendLibraryUpdate 发送媒体库更新通知
func (h *Hub) SendLibraryUpdate(serverID string, libraries interface{}) {
	h.SendMessage("library-update", serverID, 0, libraries)
}

const (
	// 写入等待时间
	writeWait = 10 * time.Second

	// 读取超时时间
	pongWait = 60 * time.Second

	// Ping周期（必须小于pongWait）
	pingPeriod = (pongWait * 9) / 10

	// 最大消息大小
	maxMessageSize = 512 * 1024 // 512KB
)

// ReadPump 读取客户端消息
func (c *Client) ReadPump() {
	defer func() {
		c.Manager.unregisterClient(c)
		c.Conn.Close()
	}()

	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})
	c.Conn.SetReadLimit(maxMessageSize)

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket错误: %v", err)
			}
			break
		}

		// 更新最后心跳时间
		c.mutex.Lock()
		c.LastPing = time.Now()
		c.mutex.Unlock()

		// 处理接收到的消息
		var msg Message
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Printf("解析消息失败: %v", err)
			continue
		}

		// 根据消息类型处理
		switch msg.Type {
		case "ping":
			// 响应ping消息
			c.Manager.sendToClient(c, Message{
				Type:      "pong",
				Data:      map[string]interface{}{"timestamp": time.Now()},
				Timestamp: time.Now(),
			})
		default:
			log.Printf("收到消息: %s from 用户 %d", msg.Type, c.UserID)
		}
	}
}

// WritePump 向客户端发送消息
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// Hub关闭了通道
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			// 序列化消息
			data, err := json.Marshal(message)
			if err != nil {
				log.Printf("序列化消息失败: %v", err)
				continue
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(data)

			// 将队列中的其他消息一起发送
			n := len(c.Send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				nextMsg := <-c.Send
				nextData, _ := json.Marshal(nextMsg)
				w.Write(nextData)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

