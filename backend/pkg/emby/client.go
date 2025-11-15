package emby

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// ConnectionStatus 连接状态
type ConnectionStatus int

const (
	StatusDisconnected ConnectionStatus = iota
	StatusConnecting
	StatusConnected
	StatusError
)

// Client Emby API客户端
type Client struct {
	BaseURL       string
	APIKey        string
	HTTPClient    *http.Client
	mutex         sync.RWMutex
	status        ConnectionStatus
	lastCheck     int64 // Unix纳秒时间戳
	retryCount    int32
	maxRetries    int32
	baseRetryDelay time.Duration

	// 状态监控
	onStatusChange func(status ConnectionStatus, err error)
}

// NewClient 创建新的Emby客户端
func NewClient(baseURL, apiKey string) *Client {
	return &Client{
		BaseURL:        strings.TrimSuffix(baseURL, "/"),
		APIKey:         apiKey,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		status:         StatusDisconnected,
		maxRetries:     3,
		baseRetryDelay: time.Second,
	}
}

// SetStatusChangeCallback 设置状态变化回调
func (c *Client) SetStatusChangeCallback(callback func(ConnectionStatus, error)) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.onStatusChange = callback
}

// updateStatus 更新连接状态
func (c *Client) updateStatus(status ConnectionStatus, err error) {
	c.mutex.Lock()
	oldStatus := c.status
	c.status = status
	atomic.StoreInt64(&c.lastCheck, time.Now().UnixNano())
	callback := c.onStatusChange
	c.mutex.Unlock()

	if oldStatus != status && callback != nil {
		go callback(status, err)
	}
}

// GetStatus 获取当前连接状态
func (c *Client) GetStatus() ConnectionStatus {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.status
}

// buildURL 构建API URL
func (c *Client) buildURL(path string, params map[string]string) string {
	u, err := url.Parse(c.BaseURL + path)
	if err != nil {
		return c.BaseURL + path
	}

	q := u.Query()
	q.Set("api_key", c.APIKey)

	for key, value := range params {
		q.Set(key, value)
	}

	u.RawQuery = q.Encode()
	return u.String()
}

// SystemInfo Emby系统信息
type SystemInfo struct {
	ServerName         string `json:"ServerName"`
	Version            string `json:"Version"`
	OperatingSystem    string `json:"OperatingSystem"`
	ID                 string `json:"Id"`
	LocalAddress       string `json:"LocalAddress"`
	WanAddress         string `json:"WanAddress"`
	HasPendingRestart  bool   `json:"HasPendingRestart"`
	HasUpdateAvailable bool   `json:"HasUpdateAvailable"`
}

// Device Emby设备信息
type DeviceInfo struct {
	Name         string    `json:"Name"`
	ID           string    `json:"Id"`
	LastUserName string    `json:"LastUserName"`
	AppName      string    `json:"AppName"`
	AppVersion   string    `json:"AppVersion"`
	DateLastActivity string `json:"DateLastActivity"`
}

// Library 媒体库信息
type Library struct {
	Name           string `json:"Name"`
	ID             string `json:"Id"`
	CollectionType string `json:"CollectionType"`
	ItemCount      int    `json:"ItemCount"`
}

// doRequest 执行HTTP请求（带重试机制）
func (c *Client) doRequest(ctx context.Context, method, path string, params map[string]string) ([]byte, error) {
	maxRetries := int(atomic.LoadInt32(&c.maxRetries))
	var lastErr error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			// 指数退避
			backoffDuration := c.baseRetryDelay * time.Duration(1<<uint(attempt-1))
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoffDuration):
			}
		}

		// 更新重试计数
		atomic.StoreInt32(&c.retryCount, int32(attempt))

		// 构建请求
		url := c.buildURL(path, params)
		req, err := http.NewRequestWithContext(ctx, method, url, nil)
		if err != nil {
			lastErr = fmt.Errorf("创建请求失败: %w", err)
			continue
		}

		// 设置请求头
		req.Header.Set("Accept", "application/json")
		req.Header.Set("User-Agent", "EmbyManager/1.0")

		// 执行请求
		resp, err := c.HTTPClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("请求失败: %w", err)
			continue
		}

		// 读取响应
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()

		if err != nil {
			lastErr = fmt.Errorf("读取响应失败: %w", err)
			continue
		}

		// 检查状态码
		if resp.StatusCode == http.StatusOK {
			c.updateStatus(StatusConnected, nil)
			return body, nil
		}

		// 对于5xx错误重试，4xx错误直接返回
		if resp.StatusCode >= 500 {
			lastErr = fmt.Errorf("服务器错误，状态码: %d, 响应: %s", resp.StatusCode, string(body))
			continue
		} else {
			c.updateStatus(StatusError, fmt.Errorf("客户端错误，状态码: %d", resp.StatusCode))
			return nil, fmt.Errorf("请求失败，状态码: %d, 响应: %s", resp.StatusCode, string(body))
		}
	}

	c.updateStatus(StatusError, lastErr)
	atomic.StoreInt32(&c.retryCount, 0)
	return nil, fmt.Errorf("请求失败，已重试%d次: %w", maxRetries, lastErr)
}

// GetSystemInfo 获取系统信息
func (c *Client) GetSystemInfo(ctx context.Context) (*SystemInfo, error) {
	body, err := c.doRequest(ctx, "GET", "/System/Info", nil)
	if err != nil {
		return nil, err
	}

	var info SystemInfo
	if err := json.Unmarshal(body, &info); err != nil {
		return nil, fmt.Errorf("解析系统信息失败: %w", err)
	}

	return &info, nil
}

// GetDevices 获取设备列表
func (c *Client) GetDevices(ctx context.Context) ([]DeviceInfo, error) {
	body, err := c.doRequest(ctx, "GET", "/Devices", nil)
	if err != nil {
		return nil, err
	}

	var response struct {
		Items []DeviceInfo `json:"Items"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("解析设备列表失败: %w", err)
	}

	return response.Items, nil
}

// GetLibraries 获取媒体库列表
func (c *Client) GetLibraries(ctx context.Context) ([]Library, error) {
	body, err := c.doRequest(ctx, "GET", "/Library/VirtualFolders", nil)
	if err != nil {
		return nil, err
	}

	var libraries []Library
	if err := json.Unmarshal(body, &libraries); err != nil {
		return nil, fmt.Errorf("解析媒体库列表失败: %w", err)
	}

	return libraries, nil
}

// Ping 检查服务器连接
func (c *Client) Ping(ctx context.Context) error {
	_, err := c.doRequest(ctx, "GET", "/System/Ping", nil)
	return err
}

// TestConnection 测试连接并返回延迟
func (c *Client) TestConnection(ctx context.Context) (time.Duration, error) {
	start := time.Now()
	c.updateStatus(StatusConnecting, nil)

	err := c.Ping(ctx)
	duration := time.Since(start)

	if err != nil {
		c.updateStatus(StatusError, err)
		return 0, err
	}

	c.updateStatus(StatusConnected, nil)
	return duration, nil
}

// GetServerStatus 获取服务器状态（包含系统信息和连接测试）
func (c *Client) GetServerStatus(ctx context.Context) (*SystemInfo, time.Duration, error) {
	// 测试连接
	duration, err := c.TestConnection(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("连接测试失败: %w", err)
	}

	// 获取系统信息
	info, err := c.GetSystemInfo(ctx)
	if err != nil {
		return nil, duration, fmt.Errorf("获取系统信息失败: %w", err)
	}

	return info, duration, nil
}

// GetRetryCount 获取当前重试次数
func (c *Client) GetRetryCount() int {
	return int(atomic.LoadInt32(&c.retryCount))
}

// SetMaxRetries 设置最大重试次数
func (c *Client) SetMaxRetries(max int) {
	atomic.StoreInt32(&c.maxRetries, int32(max))
}

// SetRetryDelay 设置基础重试延迟
func (c *Client) SetRetryDelay(delay time.Duration) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.baseRetryDelay = delay
}

// HealthCheck 健康检查
func (c *Client) HealthCheck(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	_, err := c.TestConnection(ctx)
	return err
}

// MediaItem 媒体项目信息
type MediaItem struct {
	ID             string `json:"Id"`
	Name           string `json:"Name"`
	Type           string `json:"Type"`
	CollectionType string `json:"CollectionType"`
	Path           string `json:"Path"`
	ItemCount      int    `json:"ChildCount"`
}

// GetMediaItems 获取媒体库项目列表
func (c *Client) GetMediaItems(ctx context.Context, parentID string) ([]MediaItem, error) {
	path := "/Items"
	params := make(map[string]string)

	if parentID != "" {
		params["ParentId"] = parentID
	}
	params["Recursive"] = "false"
	params["Fields"] = "Path,ChildCount"

	body, err := c.doRequest(ctx, "GET", path, params)
	if err != nil {
		return nil, err
	}

	var response struct {
		Items []MediaItem `json:"Items"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("解析媒体项目列表失败: %w", err)
	}

	return response.Items, nil
}

// GetLibraryItems 获取特定媒体库的所有项目（分页）
func (c *Client) GetLibraryItems(ctx context.Context, libraryID string, startIndex, limit int) ([]MediaItem, int, error) {
	path := "/Items"
	params := map[string]string{
		"ParentId":   libraryID,
		"Recursive":  "true",
		"StartIndex": fmt.Sprintf("%d", startIndex),
		"Limit":      fmt.Sprintf("%d", limit),
		"Fields":     "Path",
	}

	body, err := c.doRequest(ctx, "GET", path, params)
	if err != nil {
		return nil, 0, err
	}

	var response struct {
		Items      []MediaItem `json:"Items"`
		TotalCount int         `json:"TotalRecordCount"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, 0, fmt.Errorf("解析媒体库项目失败: %w", err)
	}

	return response.Items, response.TotalCount, nil
}

// SessionInfo 会话信息
type SessionInfo struct {
	Id            string         `json:"Id"`
	DeviceId      string         `json:"DeviceId"`
	DeviceName    string         `json:"DeviceName"`
	Client        string         `json:"Client"`
	UserName      string         `json:"UserName"`
	NowPlayingItem *MediaItem    `json:"NowPlayingItem"`
	PlayState     PlayStateInfo  `json:"PlayState"`
}

// PlayStateInfo 播放状态信息
type PlayStateInfo struct {
	PlayState      string `json:"PlayState"`
	PositionTicks  int64  `json:"PositionTicks"`
	IsMuted        bool   `json:"IsMuted"`
	VolumeLevel    int    `json:"VolumeLevel"`
}

// GetSessions 获取活动会话列表
func (c *Client) GetSessions(ctx context.Context) ([]SessionInfo, error) {
	body, err := c.doRequest(ctx, "GET", "/Sessions", nil)
	if err != nil {
		return nil, err
	}

	var sessions []SessionInfo
	if err := json.Unmarshal(body, &sessions); err != nil {
		return nil, fmt.Errorf("解析会话列表失败: %w", err)
	}

	return sessions, nil
}

// SendPlayCommand 发送播放命令
func (c *Client) SendPlayCommand(ctx context.Context, sessionID string) error {
	path := fmt.Sprintf("/Sessions/%s/Playing/Unpause", sessionID)
	_, err := c.doRequest(ctx, "POST", path, nil)
	return err
}

// SendPauseCommand 发送暂停命令
func (c *Client) SendPauseCommand(ctx context.Context, sessionID string) error {
	path := fmt.Sprintf("/Sessions/%s/Playing/Pause", sessionID)
	_, err := c.doRequest(ctx, "POST", path, nil)
	return err
}

// SendStopCommand 发送停止命令
func (c *Client) SendStopCommand(ctx context.Context, sessionID string) error {
	path := fmt.Sprintf("/Sessions/%s/Playing/Stop", sessionID)
	_, err := c.doRequest(ctx, "POST", path, nil)
	return err
}

// SendSeekCommand 发送跳转命令
func (c *Client) SendSeekCommand(ctx context.Context, sessionID string, positionTicks int64) error {
	path := fmt.Sprintf("/Sessions/%s/Playing/Seek", sessionID)
	params := map[string]string{
		"SeekPositionTicks": fmt.Sprintf("%d", positionTicks),
	}
	_, err := c.doRequest(ctx, "POST", path, params)
	return err
}
