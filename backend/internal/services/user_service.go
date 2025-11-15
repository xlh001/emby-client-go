package services

import (
	"fmt"
	"time"

	"github.com/emby-client-go/backend/internal/database"
	"github.com/emby-client-go/backend/internal/dto"
	"github.com/emby-client-go/backend/internal/models"
	"github.com/emby-client-go/backend/internal/utils"
	"gorm.io/gorm"
)

type UserService struct{}

func NewUserService() *UserService {
	return &UserService{}
}

// Register 用户注册
func (s *UserService) Register(req dto.RegisterRequest) (*models.User, error) {
	// 检查用户名是否已存在
	var existingUser models.User
	if err := database.DB.Where("username = ?", req.Username).First(&existingUser).Error; err == nil {
		return nil, fmt.Errorf("用户名已存在")
	}

	// 检查邮箱是否已存在
	if err := database.DB.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		return nil, fmt.Errorf("邮箱已存在")
	}

	// 哈希密码
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("密码加密失败: %v", err)
	}

	// 创建用户
	user := models.User{
		Username: req.Username,
		Email:    req.Email,
		Password: hashedPassword,
		Nickname: req.Nickname,
		Role:     "user",
		Status:   "active",
	}

	if err := database.DB.Create(&user).Error; err != nil {
		return nil, fmt.Errorf("创建用户失败: %v", err)
	}

	return &user, nil
}

// Login 用户登录
func (s *UserService) Login(req dto.LoginRequest) (*models.User, error) {
	var user models.User

	// 查找用户
	if err := database.DB.Where("username = ? OR email = ?", req.Username, req.Username).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("用户不存在")
		}
		return nil, fmt.Errorf("查询用户失败: %v", err)
	}

	// 检查账户是否被锁定
	if user.LockedUntil != nil && time.Now().Before(*user.LockedUntil) {
		remainingTime := time.Until(*user.LockedUntil).Minutes()
		return nil, fmt.Errorf("账户已被锁定，请在 %.0f 分钟后重试", remainingTime)
	}

	// 如果锁定时间已过，解锁账户
	if user.LockedUntil != nil && time.Now().After(*user.LockedUntil) {
		user.LockedUntil = nil
		user.FailedLoginCount = 0
		user.Status = "active"
		database.DB.Save(&user)
	}

	// 检查用户状态
	if user.Status != "active" {
		return nil, fmt.Errorf("用户已被禁用")
	}

	// 验证密码
	if !utils.CheckPasswordHash(req.Password, user.Password) {
		// 增加失败次数
		user.FailedLoginCount++

		// 如果失败次数达到5次，锁定账户15分钟
		if user.FailedLoginCount >= 5 {
			lockUntil := time.Now().Add(15 * time.Minute)
			user.LockedUntil = &lockUntil
			user.Status = "locked"
			database.DB.Save(&user)
			return nil, fmt.Errorf("登录失败次数过多，账户已被锁定15分钟")
		}

		database.DB.Save(&user)
		return nil, fmt.Errorf("密码错误，还剩 %d 次尝试机会", 5-user.FailedLoginCount)
	}

	// 登录成功，重置失败次数
	now := time.Now()
	user.LastLogin = &now
	user.FailedLoginCount = 0
	user.LockedUntil = nil
	database.DB.Save(&user)

	return &user, nil
}

// GetUserByID 根据ID获取用户
func (s *UserService) GetUserByID(id uint) (*models.User, error) {
	var user models.User
	if err := database.DB.First(&user, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("用户不存在")
		}
		return nil, fmt.Errorf("查询用户失败: %v", err)
	}
	return &user, nil
}

// UpdateUser 更新用户信息
func (s *UserService) UpdateUser(id uint, updates map[string]interface{}) error {
	return database.DB.Model(&models.User{}).Where("id = ?", id).Updates(updates).Error
}

// ChangePassword 修改密码
func (s *UserService) ChangePassword(userID uint, req dto.ChangePasswordRequest) error {
	// 获取用户
	user, err := s.GetUserByID(userID)
	if err != nil {
		return err
	}

	// 验证旧密码
	if !utils.CheckPasswordHash(req.OldPassword, user.Password) {
		return fmt.Errorf("原密码错误")
	}

	// 加密新密码
	hashedPassword, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		return fmt.Errorf("密码加密失败: %v", err)
	}

	// 更新密码
	return database.DB.Model(user).Update("password", hashedPassword).Error
}

// GetUsers 获取用户列表（分页）
func (s *UserService) GetUsers(page, pageSize int, search string) ([]models.User, int64, error) {
	var users []models.User
	var total int64

	query := database.DB.Model(&models.User{})

	// 搜索条件
	if search != "" {
		query = query.Where("username LIKE ? OR email LIKE ? OR nickname LIKE ?",
			"%"+search+"%", "%"+search+"%", "%"+search+"%")
	}

	// 统计总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// DeleteUser 删除用户
func (s *UserService) DeleteUser(id uint) error {
	return database.DB.Delete(&models.User{}, id).Error
}