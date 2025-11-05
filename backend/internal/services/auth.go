package services

import (
	"errors"
	"fmt"
	"time"

	"emby-manager/internal/config"
	"emby-manager/internal/dto"
	"emby-manager/internal/models"
	"emby-manager/internal/utils"
	"emby-manager/pkg/auth"

	"gorm.io/gorm"
)

// AuthService 认证服务
type AuthService struct {
	db *gorm.DB
}

// NewAuthService 创建认证服务实例
func NewAuthService(db *gorm.DB) *AuthService {
	return &AuthService{db: db}
}

// Login 用户登录
func (s *AuthService) Login(req *dto.LoginRequest) (*dto.LoginResponse, error) {
	// 查找用户
	var user models.User
	if err := s.db.Where("username = ?", req.Username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("用户名或密码错误")
		}
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}

	// 检查账户状态
	if user.Status != "active" {
		return nil, errors.New("账户已被禁用")
	}

	// 检查账户锁定状态
	if user.LockedUntil != nil && time.Now().Before(*user.LockedUntil) {
		remainingTime := user.LockedUntil.Sub(time.Now())
		return nil, fmt.Errorf("账户已被锁定，请 %d 分钟后再试", int(remainingTime.Minutes()))
	}

	// 验证密码
	if !utils.CheckPassword(req.Password, user.Password) {
		// 增加失败次数
		if err := s.updateFailedAttempts(&user, true); err != nil {
			return nil, fmt.Errorf("更新失败次数失败: %w", err)
		}
		return nil, errors.New("用户名或密码错误")
	}

	// 登录成功，重置失败次数
	if err := s.updateFailedAttempts(&user, false); err != nil {
		return nil, fmt.Errorf("重置失败次数失败: %w", err)
	}

	// 更新最后登录时间
	now := time.Now()
	user.LastLoginAt = &now
	if err := s.db.Save(&user).Error; err != nil {
		return nil, fmt.Errorf("更新登录时间失败: %w", err)
	}

	// 生成JWT令牌
	token, err := auth.GenerateToken(user.ID, user.Username, user.Role)
	if err != nil {
		return nil, fmt.Errorf("生成令牌失败: %w", err)
	}

	// 构建响应
	userInfo := &dto.UserInfo{
		ID:          user.ID,
		Username:    user.Username,
		Email:       user.Email,
		Role:        user.Role,
		Avatar:      user.Avatar,
		Status:      user.Status,
		LastLoginAt: utils.TimeToUnixPtr(user.LastLoginAt),
		CreatedAt:   user.CreatedAt.Unix(),
	}

	response := &dto.LoginResponse{
		Token:     token,
		ExpiresAt: time.Now().Add(time.Duration(config.AppConfig.JWT.ExpireTime) * time.Second).Unix(),
		User:      userInfo,
	}

	return response, nil
}

// Register 用户注册
func (s *AuthService) Register(req *dto.RegisterRequest) (*dto.UserInfo, error) {
	// 检查用户名是否已存在
	var existingUser models.User
	if err := s.db.Where("username = ?", req.Username).First(&existingUser).Error; err == nil {
		return nil, errors.New("用户名已存在")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("检查用户名失败: %w", err)
	}

	// 检查邮箱是否已存在
	if req.Email != "" {
		if err := s.db.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
			return nil, errors.New("邮箱已被使用")
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("检查邮箱失败: %w", err)
		}
	}

	// 验证密码强度
	if err := utils.ValidatePasswordStrength(req.Password); err != nil {
		return nil, err
	}

	// 哈希密码
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("密码加密失败: %w", err)
	}

	// 设置默认角色
	role := req.Role
	if role == "" {
		role = "user"
	}

	// 创建用户
	user := models.User{
		Username: req.Username,
		Email:    req.Email,
		Password: hashedPassword,
		Role:     role,
		Status:   "active",
	}

	if err := s.db.Create(&user).Error; err != nil {
		return nil, fmt.Errorf("创建用户失败: %w", err)
	}

	// 转换为用户信息
	userInfo := &dto.UserInfo{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		Role:      user.Role,
		Status:    user.Status,
		CreatedAt: user.CreatedAt.Unix(),
	}

	return userInfo, nil
}

// RefreshToken 刷新令牌
func (s *AuthService) RefreshToken(req *dto.RefreshTokenRequest) (*dto.LoginResponse, error) {
	// 验证旧令牌
	claims, err := auth.ValidateToken(req.Token)
	if err != nil {
		return nil, fmt.Errorf("令牌验证失败: %w", err)
	}

	// 获取用户信息
	var user models.User
	if err := s.db.First(&user, claims.UserID).Error; err != nil {
		return nil, fmt.Errorf("获取用户信息失败: %w", err)
	}

	// 检查用户状态
	if user.Status != "active" {
		return nil, errors.New("账户已被禁用")
	}

	// 生成新令牌
	newToken, err := auth.GenerateToken(user.ID, user.Username, user.Role)
	if err != nil {
		return nil, fmt.Errorf("生成令牌失败: %w", err)
	}

	// 构建响应
	userInfo := &dto.UserInfo{
		ID:          user.ID,
		Username:    user.Username,
		Email:       user.Email,
		Role:        user.Role,
		Avatar:      user.Avatar,
		Status:      user.Status,
		LastLoginAt: utils.TimeToUnixPtr(user.LastLoginAt),
		CreatedAt:   user.CreatedAt.Unix(),
	}

	response := &dto.LoginResponse{
		Token:     newToken,
		ExpiresAt: time.Now().Add(time.Duration(config.AppConfig.JWT.ExpireTime) * time.Second).Unix(),
		User:      userInfo,
	}

	return response, nil
}

// GetUserByID 根据ID获取用户信息
func (s *AuthService) GetUserByID(userID uint) (*dto.UserInfo, error) {
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("用户不存在")
		}
		return nil, fmt.Errorf("获取用户失败: %w", err)
	}

	userInfo := &dto.UserInfo{
		ID:          user.ID,
		Username:    user.Username,
		Email:       user.Email,
		Role:        user.Role,
		Avatar:      user.Avatar,
		Status:      user.Status,
		LastLoginAt: utils.TimeToUnixPtr(user.LastLoginAt),
		CreatedAt:   user.CreatedAt.Unix(),
	}

	return userInfo, nil
}

// UpdatePassword 更新密码
func (s *AuthService) UpdatePassword(userID uint, req *dto.UpdatePasswordRequest) error {
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return fmt.Errorf("获取用户失败: %w", err)
	}

	// 验证旧密码
	if !utils.CheckPassword(req.OldPassword, user.Password) {
		return errors.New("原密码错误")
	}

	// 验证新密码强度
	if err := utils.ValidatePasswordStrength(req.NewPassword); err != nil {
		return err
	}

	// 哈希新密码
	hashedPassword, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		return fmt.Errorf("密码加密失败: %w", err)
	}

	// 更新密码
	user.Password = hashedPassword
	if err := s.db.Save(&user).Error; err != nil {
		return fmt.Errorf("更新密码失败: %w", err)
	}

	return nil
}

// updateFailedAttempts 更新或重置失败尝试次数
func (s *AuthService) updateFailedAttempts(user *models.User, increment bool) error {
	if increment {
		user.FailedAttempts++
		// 如果失败次数达到5次，锁定账户30分钟
		if user.FailedAttempts >= 5 {
			lockedUntil := time.Now().Add(30 * time.Minute)
			user.LockedUntil = &lockedUntil
			user.Status = "banned"
		}
	} else {
		user.FailedAttempts = 0
		if user.LockedUntil != nil {
			user.LockedUntil = nil
		}
		// 如果之前被锁定，恢复为活跃状态
		if user.Status == "banned" {
			user.Status = "active"
		}
	}

	return s.db.Save(user).Error
}

// ValidatePermissions 验证用户权限
func (s *AuthService) ValidatePermissions(userID uint, resource, action string) bool {
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return false
	}

	// 管理员拥有所有权限
	if user.Role == "admin" {
		return true
	}

	// 普通用户的权限检查
	switch {
	case resource == "profile" && action == "read":
		return true
	case resource == "profile" && action == "update":
		return true
	case resource == "media" && action == "read":
		return true
	case resource == "playback" && action == "control":
		return true
	default:
		return false
	}
}