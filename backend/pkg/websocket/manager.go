package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// EmbyConnection 单个Emby服务器的WebSocket连接
type EmbyConnection struct {
	ServerID       string
	ServerURL      string
	APIKey         string
	Conn           *websocket.Conn
	Status         ConnectionStatus
	LastConnected  time.Time
	ReconnectCount int
	MaxReconnects  int
	ReconnectDelay time.Duration

	// 控制通道
	stopChan       chan struct{}
	reconnectChan  chan struct{}

	// 消息处理
	messageHandler func(serverID string, message []byte)

	// 并发安全
	mutex          sync.RWMutex

	// 管理器引用
	manager        *Manager
}

// ConnectionStatus 连接状态
type ConnectionStatus int

const (
	Disconnected ConnectionStatus = iota
	Connecting
	Connected
	Reconnecting
	Failed
)

// String 返回连接状态的字符串表示
func (s ConnectionStatus) String() string {
	switch s {
	case Disconnected:
		return "disconnected"
	case Connecting:
		return "connecting"
	case Connected:
		return "connected"
	case Reconnecting:
		return "reconnecting"
	case Failed:
		return "failed"
	default:
		return "unknown"
	}
}

// Manager WebSocket连接管理器
type Manager struct {
	// 连接池：serverID -> EmbyConnection
	connections map[string]*EmbyConnection

	// Hub引用（用于向前端客户端广播）
	hub *Hub

	// 并发安全
	mutex sync.RWMutex

	// 全局配置
	maxReconnects  int
	reconnectDelay time.Duration

	// 消息处理器
	messageHandler func(serverID string, message []byte)
}

// NewManager 创建新的WebSocket管理器
func NewManager(hub *Hub) *Manager {
	return &Manager{
		connections:    make(map[string]*EmbyConnection),
		hub:            hub,
		maxReconnects:  5,
		reconnectDelay: 5 * time.Second,
	}
}

// SetMessageHandler 设置消息处理器
func (m *Manager) SetMessageHandler(handler func(serverID string, message []byte)) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.messageHandler = handler
}

// AddConnection 添加新的Emby服务器连接
func (m *Manager) AddConnection(serverID, serverURL, apiKey string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 检查是否已存在
	if _, exists := m.connections[serverID]; exists {
		return fmt.Errorf("服务器 %s 的连接已存在", serverID)
	}

	// 创建新连接
	conn := &EmbyConnection{
		ServerID:       serverID,
		ServerURL:      serverURL,
		APIKey:         apiKey,
		Status:         Disconnected,
		MaxReconnects:  m.maxReconnects,
		ReconnectDelay: m.reconnectDelay,
		stopChan:       make(chan struct{}),
		reconnectChan:  make(chan struct{}, 1),
		messageHandler: m.messageHandler,
		manager:        m,
	}

	m.connections[serverID] = conn

	// 启动连接
	go conn.Start()

	log.Printf("已添加Emby服务器连接: %s (%s)", serverID, serverURL)
	return nil
}

// RemoveConnection 移除Emby服务器连接
func (m *Manager) RemoveConnection(serverID string) error {
	m.mutex.Lock()
	conn, exists := m.connections[serverID]
	if !exists {
		m.mutex.Unlock()
		return fmt.Errorf("服务器 %s 的连接不存在", serverID)
	}
	delete(m.connections, serverID)
	m.mutex.Unlock()

	// 停止连接
	conn.Stop()

	log.Printf("已移除Emby服务器连接: %s", serverID)
	return nil
}

// GetConnection 获取指定服务器的连接
func (m *Manager) GetConnection(serverID string) (*EmbyConnection, bool) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	conn, exists := m.connections[serverID]
	return conn, exists
}

// GetAllConnections 获取所有连接
func (m *Manager) GetAllConnections() map[string]*EmbyConnection {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	result := make(map[string]*EmbyConnection, len(m.connections))
	for k, v := range m.connections {
		result[k] = v
	}
	return result
}

// GetConnectionStatus 获取所有连接的状态
func (m *Manager) GetConnectionStatus() map[string]string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	status := make(map[string]string, len(m.connections))
	for serverID, conn := range m.connections {
		conn.mutex.RLock()
		status[serverID] = conn.Status.String()
		conn.mutex.RUnlock()
	}
	return status
}

// BroadcastToServer 向特定服务器发送消息
func (m *Manager) BroadcastToServer(serverID string, message interface{}) error {
	conn, exists := m.GetConnection(serverID)
	if !exists {
		return fmt.Errorf("服务器 %s 的连接不存在", serverID)
	}

	return conn.SendMessage(message)
}

// Start 启动Emby服务器连接
func (ec *EmbyConnection) Start() {
	ec.mutex.Lock()
	if ec.Status != Disconnected {
		ec.mutex.Unlock()
		return
	}
	ec.Status = Connecting
	ec.mutex.Unlock()

	// 首次连接
	if err := ec.connect(); err != nil {
		log.Printf("服务器 %s 初始连接失败: %v", ec.ServerID, err)
		ec.scheduleReconnect()
	}

	// 启动重连监听
	go ec.reconnectLoop()
}

// Stop 停止连接
func (ec *EmbyConnection) Stop() {
	ec.mutex.Lock()
	defer ec.mutex.Unlock()

	close(ec.stopChan)

	if ec.Conn != nil {
		ec.Conn.Close()
		ec.Conn = nil
	}

	ec.Status = Disconnected
	log.Printf("服务器 %s 连接已停止", ec.ServerID)
}

// connect 建立WebSocket连接
func (ec *EmbyConnection) connect() error {
	// 构建WebSocket URL
	wsURL, err := ec.buildWebSocketURL()
	if err != nil {
		return fmt.Errorf("构建WebSocket URL失败: %w", err)
	}

	// 建立连接
	dialer := websocket.Dialer{
		HandshakeTimeout: 10 * time.Second,
	}

	conn, _, err := dialer.Dial(wsURL, nil)
	if err != nil {
		ec.updateStatus(Failed)
		return fmt.Errorf("连接失败: %w", err)
	}

	ec.mutex.Lock()
	ec.Conn = conn
	ec.Status = Connected
	ec.LastConnected = time.Now()
	ec.ReconnectCount = 0
	ec.mutex.Unlock()

	log.Printf("服务器 %s WebSocket连接已建立", ec.ServerID)

	// 通知前端客户端
	ec.notifyStatusChange()

	// 启动读写循环
	go ec.readLoop()
	go ec.writeLoop()

	return nil
}

// buildWebSocketURL 构建WebSocket URL
func (ec *EmbyConnection) buildWebSocketURL() (string, error) {
	u, err := url.Parse(ec.ServerURL)
	if err != nil {
		return "", err
	}

	// 转换为WebSocket协议
	if u.Scheme == "https" {
		u.Scheme = "wss"
	} else {
		u.Scheme = "ws"
	}

	// 添加WebSocket路径和API密钥
	u.Path = "/embywebsocket"
	q := u.Query()
	q.Set("api_key", ec.APIKey)
	q.Set("deviceId", "EmbyManager")
	u.RawQuery = q.Encode()

	return u.String(), nil
}

// readLoop 读取消息循环
func (ec *EmbyConnection) readLoop() {
	defer func() {
		ec.handleDisconnect()
	}()

	ec.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	ec.Conn.SetPongHandler(func(string) error {
		ec.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		select {
		case <-ec.stopChan:
			return
		default:
			_, message, err := ec.Conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("服务器 %s WebSocket异常关闭: %v", ec.ServerID, err)
				}
				return
			}

			// 处理消息
			ec.handleMessage(message)
		}
	}
}

// writeLoop 写入消息循环
func (ec *EmbyConnection) writeLoop() {
	ticker := time.NewTicker(54 * time.Second) // Ping周期
	defer func() {
		ticker.Stop()
	}()

	for {
		select {
		case <-ec.stopChan:
			return
		case <-ticker.C:
			ec.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := ec.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Printf("服务器 %s 发送Ping失败: %v", ec.ServerID, err)
				return
			}
		}
	}
}

// handleMessage 处理接收到的消息
func (ec *EmbyConnection) handleMessage(message []byte) {
	// 解析消息
	var msg map[string]interface{}
	if err := json.Unmarshal(message, &msg); err != nil {
		log.Printf("服务器 %s 解析消息失败: %v", ec.ServerID, err)
		return
	}

	// 调用消息处理器
	if ec.messageHandler != nil {
		ec.messageHandler(ec.ServerID, message)
	}

	// 广播到前端客户端
	if ec.manager != nil && ec.manager.hub != nil {
		msgType, _ := msg["MessageType"].(string)
		ec.manager.hub.SendMessage(msgType, ec.ServerID, 0, msg)
	}
}

// handleDisconnect 处理断开连接
func (ec *EmbyConnection) handleDisconnect() {
	ec.mutex.Lock()
	if ec.Conn != nil {
		ec.Conn.Close()
		ec.Conn = nil
	}
	ec.mutex.Unlock()

	ec.updateStatus(Disconnected)
	ec.notifyStatusChange()

	// 触发重连
	ec.scheduleReconnect()
}

// scheduleReconnect 安排重连
func (ec *EmbyConnection) scheduleReconnect() {
	select {
	case ec.reconnectChan <- struct{}{}:
	default:
		// 重连已在队列中
	}
}

// reconnectLoop 重连循环
func (ec *EmbyConnection) reconnectLoop() {
	for {
		select {
		case <-ec.stopChan:
			return
		case <-ec.reconnectChan:
			ec.attemptReconnect()
		}
	}
}

// attemptReconnect 尝试重连
func (ec *EmbyConnection) attemptReconnect() {
	ec.mutex.Lock()
	if ec.ReconnectCount >= ec.MaxReconnects {
		ec.Status = Failed
		ec.mutex.Unlock()
		log.Printf("服务器 %s 重连次数已达上限 (%d)", ec.ServerID, ec.MaxReconnects)
		ec.notifyStatusChange()
		return
	}
	ec.ReconnectCount++
	ec.Status = Reconnecting
	ec.mutex.Unlock()

	log.Printf("服务器 %s 尝试重连 (第 %d 次)", ec.ServerID, ec.ReconnectCount)
	ec.notifyStatusChange()

	// 等待一段时间后重连
	time.Sleep(ec.ReconnectDelay)

	if err := ec.connect(); err != nil {
		log.Printf("服务器 %s 重连失败: %v", ec.ServerID, err)
		// 继续尝试重连
		ec.scheduleReconnect()
	}
}

// SendMessage 发送消息到Emby服务器
func (ec *EmbyConnection) SendMessage(message interface{}) error {
	ec.mutex.RLock()
	conn := ec.Conn
	status := ec.Status
	ec.mutex.RUnlock()

	if status != Connected || conn == nil {
		return fmt.Errorf("服务器 %s 未连接", ec.ServerID)
	}

	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("序列化消息失败: %w", err)
	}

	conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
		return fmt.Errorf("发送消息失败: %w", err)
	}

	return nil
}

// GetStatus 获取连接状态
func (ec *EmbyConnection) GetStatus() ConnectionStatus {
	ec.mutex.RLock()
	defer ec.mutex.RUnlock()
	return ec.Status
}

// updateStatus 更新连接状态
func (ec *EmbyConnection) updateStatus(status ConnectionStatus) {
	ec.mutex.Lock()
	ec.Status = status
	ec.mutex.Unlock()
}

// notifyStatusChange 通知状态变化
func (ec *EmbyConnection) notifyStatusChange() {
	if ec.manager != nil && ec.manager.hub != nil {
		ec.manager.hub.SendServerStatus(ec.ServerID, map[string]interface{}{
			"status":          ec.Status.String(),
			"last_connected":  ec.LastConnected,
			"reconnect_count": ec.ReconnectCount,
		})
	}
}

// ResetReconnectCount 重置重连计数
func (ec *EmbyConnection) ResetReconnectCount() {
	ec.mutex.Lock()
	defer ec.mutex.Unlock()
	ec.ReconnectCount = 0
}

// ForceReconnect 强制重连
func (ec *EmbyConnection) ForceReconnect() {
	ec.Stop()
	time.Sleep(time.Second)
	ec.Start()
}
