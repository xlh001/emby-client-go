package services

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

// EmbyClient Emby API客户端
type EmbyClient struct {
	BaseURL    string
	HTTPClient *http.Client
}

// EmbyServerInfo Emby服务器信息
type EmbyServerInfo struct {
	LocalAddress string   `json:"LocalAddress"`
	Name         string   `json:"ServerName"`
	Id           string   `json:"Id"`
	Version      string   `json:"Version"`
	ProductName  string   `json:"ProductName"`
	HttpsPort    int      `json:"HttpsPort"`
	HttpPort     int      `json:"HttpPort"`
}

// EmbyAuthResponse 认证响应
type EmbyAuthResponse struct {
	AccessToken string `json:"AccessToken"`
	User        struct {
		ID          string `json:"Id"`
		Name        string `json:"Name"`
		ServerId    string `json:"ServerId"`
		PrimaryImageTag string `json:"PrimaryImageTag"`
	} `json:"User"`
}

// EmbyDevice 设备信息
type EmbyDevice struct {
	Name         string `json:"Name"`
	AppName      string `json:"AppName"`
	AppVersion   string `json:"AppVersion"`
	DeviceId     string `json:"DeviceId"`
	IconUrl      string `json:"IconUrl"`
	UserName     string `json:"UserName"`
	LastWonTime  string `json:"LastWonTime"`
	DateLastMode string `json:"DateLastMode"`
	CustomName   string `json:"CustomName"`
}

// NewEmbyClient 创建新的Emby客户端
func NewEmbyClient(baseURL string, useHTTPS bool) *EmbyClient {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: !useHTTPS},
	}

	client := &EmbyClient{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout:   30 * time.Second,
			Transport: transport,
		},
	}

	return client
}

// TestConnection 测试服务器连接
func (c *EmbyClient) TestConnection() error {
	url := fmt.Sprintf("%s/emby/System/Ping", c.BaseURL)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "MediaBrowser Client=\"Emby Manager\", Device=\"Server\", DeviceId=\"emby-manager\", Version=\"1.0.0\"")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("服务器响应错误: %d", resp.StatusCode)
	}

	return nil
}

// GetServerInfo 获取服务器信息
func (c *EmbyClient) GetServerInfo() (*EmbyServerInfo, error) {
	url := fmt.Sprintf("%s/emby/System/Info", c.BaseURL)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "MediaBrowser Client=\"Emby Manager\", Device=\"Server\", DeviceId=\"emby-manager\", Version=\"1.0.0\"")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("获取服务器信息失败: %d", resp.StatusCode)
	}

	var serverInfo EmbyServerInfo
	err = json.NewDecoder(resp.Body).Decode(&serverInfo)
	if err != nil {
		return nil, err
	}

	return &serverInfo, nil
}

// Authenticate 用户认证
func (c *EmbyClient) Authenticate(username, password string) (*EmbyAuthResponse, error) {
	url := fmt.Sprintf("%s/emby/Users/AuthenticateByName", c.BaseURL)

	authData := map[string]interface{}{
		"Username": username,
		"Pw":       password,
	}

	jsonData, err := json.Marshal(authData)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "MediaBrowser Client=\"Emby Manager\", Device=\"Server\", DeviceId=\"emby-manager\", Version=\"1.0.0\"")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("认证失败: %d", resp.StatusCode)
	}

	var authResp EmbyAuthResponse
	err = json.NewDecoder(resp.Body).Decode(&authResp)
	if err != nil {
		return nil, err
	}

	return &authResp, nil
}

// GetDevices 获取设备列表
func (c *EmbyClient) GetDevices(accessToken string) ([]EmbyDevice, error) {
	url := fmt.Sprintf("%s/emby/Sessions", c.BaseURL)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	req.Header.Set("X-MediaBrowser-Token", accessToken)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("获取设备列表失败: %d", resp.StatusCode)
	}

	var devices []EmbyDevice
	err = json.NewDecoder(resp.Body).Decode(&devices)
	if err != nil {
		return nil, err
	}

	return devices, nil
}

// TestWithCredentials 测试服务器连接和认证
func (c *EmbyClient) TestWithCredentials(username, password string) (*EmbyServerInfo, error) {
	// 首先测试基本连接
	if err := c.TestConnection(); err != nil {
		return nil, fmt.Errorf("连接测试失败: %v", err)
	}

	// 如果提供了凭据，测试认证
	if username != "" && password != "" {
		_, err := c.Authenticate(username, password)
		if err != nil {
			return nil, fmt.Errorf("认证测试失败: %v", err)
		}
	}

	// 获取服务器信息
	serverInfo, err := c.GetServerInfo()
	if err != nil {
		return nil, fmt.Errorf("获取服务器信息失败: %v", err)
	}

	log.Printf("成功连接到Emby服务器: %s 版本: %s", serverInfo.Name, serverInfo.Version)
	return serverInfo, nil
}