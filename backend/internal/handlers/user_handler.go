package handlers

import (
	"net/http"
	"strconv"

	"github.com/emby-client-go/backend/internal/dto"
	"github.com/emby-client-go/backend/internal/middleware"
	"github.com/emby-client-go/backend/internal/models"
	"github.com/emby-client-go/backend/internal/services"
	"github.com/gin-gonic/gin"
)

// UserHandler 用户处理器
type UserHandler struct {
	userService *services.UserService
}

func NewUserHandler() *UserHandler {
	return &UserHandler{
		userService: services.NewUserService(),
	}
}

// Register 用户注册
// @Summary 用户注册
// @Description 创建新用户账户
// @Tags 用户认证
// @Accept json
// @Produce json
// @Param request body dto.RegisterRequest true "注册信息"
// @Success 200 {object} dto.ApiResponse{data=dto.UserResponse}
// @Failure 400 {object} dto.ApiResponse
// @Router /auth/register [post]
func (h *UserHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ApiResponse{
			Code:    400,
			Message: "参数错误: " + err.Error(),
		})
		return
	}

	user, err := h.userService.Register(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ApiResponse{
			Code:    400,
			Message: err.Error(),
		})
		return
	}

	userResponse := dto.UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		Nickname:  user.Nickname,
		Role:      user.Role,
		Status:    user.Status,
		CreatedAt: user.CreatedAt.Format("2006-01-02 15:04:05"),
	}

	c.JSON(http.StatusOK, dto.ApiResponse{
		Code:    200,
		Message: "注册成功",
		Data:    userResponse,
	})
}

// Login 用户登录
// @Summary 用户登录
// @Description 用户登录获取访问令牌
// @Tags 用户认证
// @Accept json
// @Produce json
// @Param request body dto.LoginRequest true "登录信息"
// @Success 200 {object} dto.ApiResponse{data=dto.LoginResponse}
// @Failure 400 {object} dto.ApiResponse
// @Router /auth/login [post]
func (h *UserHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ApiResponse{
			Code:    400,
			Message: "参数错误: " + err.Error(),
		})
		return
	}

	user, err := h.userService.Login(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ApiResponse{
			Code:    400,
			Message: err.Error(),
		})
		return
	}

	// 生成JWT token
	token, err := middleware.GenerateToken(*user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ApiResponse{
			Code:    500,
			Message: "生成token失败",
		})
		return
	}

	var lastLoginStr string
	if user.LastLogin != nil {
		lastLoginStr = user.LastLogin.Format("2006-01-02 15:04:05")
	}

	userResponse := dto.UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		Nickname:  user.Nickname,
		Role:      user.Role,
		Status:    user.Status,
		LastLogin: &lastLoginStr,
		CreatedAt: user.CreatedAt.Format("2006-01-02 15:04:05"),
	}

	loginResponse := dto.LoginResponse{
		Token: token,
		User:  userResponse,
	}

	c.JSON(http.StatusOK, dto.ApiResponse{
		Code:    200,
		Message: "登录成功",
		Data:    loginResponse,
	})
}

// RefreshToken 刷新令牌
// @Summary 刷新令牌
// @Description 使用旧令牌刷新获取新令牌
// @Tags 用户认证
// @Accept json
// @Produce json
// @Param request body dto.RefreshTokenRequest true "刷新令牌请求"
// @Success 200 {object} dto.ApiResponse{data=dto.RefreshTokenResponse}
// @Failure 400 {object} dto.ApiResponse
// @Router /auth/refresh [post]
func (h *UserHandler) RefreshToken(c *gin.Context) {
	var req dto.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ApiResponse{
			Code:    400,
			Message: "参数错误: " + err.Error(),
		})
		return
	}

	// 刷新令牌
	newToken, err := middleware.RefreshToken(req.Token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.ApiResponse{
			Code:    401,
			Message: "令牌刷新失败: " + err.Error(),
		})
		return
	}

	response := dto.RefreshTokenResponse{
		Token: newToken,
	}

	c.JSON(http.StatusOK, dto.ApiResponse{
		Code:    200,
		Message: "令牌刷新成功",
		Data:    response,
	})
}

// Logout 用户登出
// @Summary 用户登出
// @Description 用户登出（客户端需清除本地令牌）
// @Tags 用户认证
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} dto.ApiResponse
// @Failure 401 {object} dto.ApiResponse
// @Router /auth/logout [post]
func (h *UserHandler) Logout(c *gin.Context) {
	// JWT是无状态的，登出主要由客户端处理（删除本地存储的token）
	// 这里可以记录登出日志或执行其他清理操作
	userID, exists := c.Get("user_id")
	if exists {
		// 可以在这里记录登出日志
		// 例如：记录到数据库或日志文件
		_ = userID
	}

	c.JSON(http.StatusOK, dto.ApiResponse{
		Code:    200,
		Message: "登出成功",
	})
}

// GetProfile 获取用户信息
// @Summary 获取用户信息
// @Description 获取当前登录用户的详细信息
// @Tags 用户管理
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} dto.ApiResponse{data=dto.UserResponse}
// @Failure 401 {object} dto.ApiResponse
// @Router /user/profile [get]
func (h *UserHandler) GetProfile(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ApiResponse{
			Code:    401,
			Message: "未找到用户信息",
		})
		return
	}

	userModel := user.(models.User)
	var lastLoginStr string
	if userModel.LastLogin != nil {
		lastLoginStr = userModel.LastLogin.Format("2006-01-02 15:04:05")
	}

	userResponse := dto.UserResponse{
		ID:        userModel.ID,
		Username:  userModel.Username,
		Email:     userModel.Email,
		Nickname:  userModel.Nickname,
		Role:      userModel.Role,
		Status:    userModel.Status,
		LastLogin: &lastLoginStr,
		CreatedAt: userModel.CreatedAt.Format("2006-01-02 15:04:05"),
	}

	c.JSON(http.StatusOK, dto.ApiResponse{
		Code:    200,
		Message: "获取成功",
		Data:    userResponse,
	})
}

// ChangePassword 修改密码
// @Summary 修改密码
// @Description 修改当前用户密码
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body dto.ChangePasswordRequest true "密码信息"
// @Success 200 {object} dto.ApiResponse
// @Failure 400 {object} dto.ApiResponse
// @Router /user/change-password [post]
func (h *UserHandler) ChangePassword(c *gin.Context) {
	var req dto.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ApiResponse{
			Code:    400,
			Message: "参数错误: " + err.Error(),
		})
		return
	}

	userID, _ := c.Get("user_id")
	err := h.userService.ChangePassword(userID.(uint), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ApiResponse{
			Code:    400,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.ApiResponse{
		Code:    200,
		Message: "密码修改成功",
	})
}

// GetUsers 获取用户列表
// @Summary 获取用户列表
// @Description 获取用户列表（管理员权限）
// @Tags 用户管理
// @Produce json
// @Security ApiKeyAuth
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Param search query string false "搜索关键词"
// @Success 200 {object} dto.ApiResponse{data=dto.PageResponse}
// @Failure 403 {object} dto.ApiResponse
// @Router /user/list [get]
func (h *UserHandler) GetUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	search := c.Query("search")

	users, total, err := h.userService.GetUsers(page, pageSize, search)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ApiResponse{
			Code:    500,
			Message: "获取用户列表失败",
		})
		return
	}

	var userResponses []dto.UserResponse
	for _, user := range users {
		var lastLoginStr string
		if user.LastLogin != nil {
			lastLoginStr = user.LastLogin.Format("2006-01-02 15:04:05")
		}

		userResponses = append(userResponses, dto.UserResponse{
			ID:        user.ID,
			Username:  user.Username,
			Email:     user.Email,
			Nickname:  user.Nickname,
			Role:      user.Role,
			Status:    user.Status,
			LastLogin: &lastLoginStr,
			CreatedAt: user.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	pageResponse := dto.PageResponse{
		List:     userResponses,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}

	c.JSON(http.StatusOK, dto.ApiResponse{
		Code:    200,
		Message: "获取成功",
		Data:    pageResponse,
	})
}