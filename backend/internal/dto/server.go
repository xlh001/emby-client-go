package dto

// CreateServerRequest 创建服务器请求
type CreateServerRequest struct {
	Name        string `json:"name" binding:"required"`
	URL         string `json:"url" binding:"required,url"`
	APIKey      string `json:"api_key" binding:"required"`
	Description string `json:"description"`
}

// UpdateServerRequest 更新服务器请求
type UpdateServerRequest struct {
	Name        string `json:"name"`
	URL         string `json:"url" binding:"omitempty,url"`
	APIKey      string `json:"api_key"`
	Description string `json:"description"`
}

// ServerResponse 服务器响应
type ServerResponse struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	URL         string `json:"url"`
	Version     string `json:"version"`
	OS          string `json:"os"`
	Status      string `json:"status"`
	LastCheck   string `json:"last_check"`
	Description string `json:"description"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// TestConnectionResponse 测试连接响应
type TestConnectionResponse struct {
	Status       string `json:"status"`
	ResponseTime int64  `json:"response_time"` // 毫秒
	Message      string `json:"message"`
}
