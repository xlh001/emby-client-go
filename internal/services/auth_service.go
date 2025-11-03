package services

import (
	"emby-client-go/internal/models"
	"errors"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// AuthService 认证服务
type AuthService struct {
	db        *gorm.DB
	jwtSecret string
}

// TokenResponse 令牌响应
type TokenResponse struct {
	Token     string `json:"token"`
	ExpiresAt int64  `json:"expires_at"`
	User      models.User `json:"user"`
}

// NewAuthService 创建认证服务
func NewAuthService(jwtSecret string) *AuthService {
	return &AuthService{
		jwtSecret: jwtSecret,
	}
}

// SetDB 设置数据库连接
func (a *AuthService) SetDB(db *gorm.DB) {
	a.db = db
}

// Register 注册新用户
func (a *AuthService) Register(user *models.User) error {
	// 检查用户名是否已存在
	var existingUser models.User
	err := a.db.Where("username = ?", user.Username).First(&existingUser).Error
	if err == nil {
		return errors.New("用户名已存在")
	}

	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.Password = string(hashedPassword)

	// 设置默认角色
	if user.Role == "" {
		user.Role = "user"
	}

	// 创建用户
	return a.db.Create(user).Error
}

// Login 用户登录
func (a *AuthService) Login(username, password string) (*TokenResponse, error) {
	var user models.User
	err := a.db.Where("username = ? AND is_active = ?", username, true).First(&user).Error
	if err != nil {
		return nil, errors.New("用户名或密码错误")
	}

	// 验证密码
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return nil, errors.New("用户名或密码错误")
	}

	// 生成JWT令牌
	token, expiresAt, err := a.GenerateToken(user)
	if err != nil {
		return nil, err
	}

	// 更新最后登录时间
	now := time.Now()
	a.db.Model(&user).Update("last_login", &now)

	// 清除密码字段
	user.Password = ""

	return &TokenResponse{
		Token:     token,
		ExpiresAt: expiresAt,
		User:      user,
	}, nil
}

// GenerateToken 生成JWT令牌
func (a *AuthService) GenerateToken(user models.User) (string, int64, error) {
	claims := jwt.MapClaims{
		"user_id": user.ID,
		"username": user.Username,
		"role": user.Role,
		"exp": time.Now().Add(time.Hour * 24).Unix(), // 24小时过期
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(a.jwtSecret))
	if err != nil {
		return "", 0, err
	}

	return tokenString, claims["exp"].(int64), nil
}

// ValidateToken 验证JWT令牌
func (a *AuthService) ValidateToken(tokenString string) (*jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("无效的签名方法")
		}
		return []byte(a.jwtSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return &claims, nil
	}

	return nil, errors.New("无效的令牌")
}

// GetUserFromToken 从令牌获取用户信息
func (a *AuthService) GetUserFromToken(tokenString string) (*models.User, error) {
	claims, err := a.ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}

	userID := (*claims)["user_id"].(float64)
	var user models.User
	err = a.db.First(&user, uint(userID)).Error
	if err != nil {
		return nil, errors.New("用户不存在")
	}

	if !user.IsActive {
		return nil, errors.New("用户已被禁用")
	}

	// 清除密码字段
	user.Password = ""
	return &user, nil
}

// RequireAuth middleware中间件
func (a *AuthService) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(401, gin.H{"error": "缺少授权头部"})
			c.Abort()
			return
		}

		// 提取令牌 "Bearer token"
		if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
			c.JSON(401, gin.H{"error": "授权格式错误"})
			c.Abort()
			return
		}

		tokenString := authHeader[7:]
		user, err := a.GetUserFromToken(tokenString)
		if err != nil {
			c.JSON(401, gin.H{"error": err.Error()})
			c.Abort()
			return
		}

		// 将用户信息存入上下文
		c.Set("user", user)
		c.Next()
	}
}

// RequireAuthAdmin 需要管理员权限
func (a *AuthService) RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := c.Get("user")
		if !exists {
			c.JSON(401, gin.H{"error": "未认证"})
			c.Abort()
			return
		}

		if user.(*models.User).Role != "admin" {
			c.JSON(403, gin.H{"error": "需要管理员权限"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// ChangePassword 修改密码
func (a *AuthService) ChangePassword(userID uint, oldPassword, newPassword string) error {
	var user models.User
	err := a.db.First(&user, userID).Error
	if err != nil {
		return errors.New("用户不存在")
	}

	// 验证旧密码
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPassword))
	if err != nil {
		return errors.New("原密码错误")
	}

	// 加密新密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// 更新密码
	return a.db.Model(&user).Update("password", string(hashedPassword)).Error
}

// CreateAdminUser 创建默认管理员用户
func (a *AuthService) CreateAdminUser(db *gorm.DB, username, password string) error {
	// 检查是否已有管理员
	var adminCount int64
	db.Model(&models.User{}).Where("role = ?", "admin").Count(&adminCount)
	if adminCount > 0 {
		return nil // 已有管理员，不创建
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	admin := models.User{
		Username: username,
		Password: string(hashedPassword),
		Role:     "admin",
		IsActive: true,
	}

	return db.Create(&admin).Error
}