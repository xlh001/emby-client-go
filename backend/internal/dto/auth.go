package dto

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Password string `json:"password" binding:"required,min=6,max=128"`
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6,max=128"`
	Role     string `json:"role,omitempty"` // 可选，默认为 user
}

// LoginResponse 登录响应
type LoginResponse struct {
	Token      string     `json:"token"`
	ExpiresAt  int64      `json:"expires_at"`
	User       *UserInfo  `json:"user"`
}

// UserInfo 用户信息（不包含敏感信息）
type UserInfo struct {
	ID          uint      `json:"id"`
	Username    string    `json:"username"`
	Email       string    `json:"email"`
	Role        string    `json:"role"`
	Avatar      string    `json:"avatar"`
	Status      string    `json:"status"`
	LastLoginAt *int64    `json:"last_login_at"`
	CreatedAt   int64     `json:"created_at"`
}

// RefreshTokenRequest 刷新令牌请求
type RefreshTokenRequest struct {
	Token string `json:"token" binding:"required"`
}

// UpdatePasswordRequest 更新密码请求
type UpdatePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6,max=128"`
}

// PasswordResetRequest 密码重置请求
type PasswordResetRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// PasswordResetConfirmRequest 确认密码重置请求
type PasswordResetConfirmRequest struct {
	Token    string `json:"token" binding:"required"`
	Password string `json:"password" binding:"required,min=6,max=128"`
}