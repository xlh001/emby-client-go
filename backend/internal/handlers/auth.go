package handlers

import (
	"net/http"
	"strconv"

	"emby-manager/internal/dto"
	"emby-manager/internal/services"
	"emby-manager/pkg/database"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// 全局服务实例
var authService *services.AuthService

// InitAuthServices 初始化认证服务
func InitAuthServices() {
	authService = services.NewAuthService(database.DB)
}

// Health 健康检查
// @Summary 健康检查
// @Description 检查API服务健康状态
// @Tags 系统
// @Accept json
// @Produce json
// @Success 200 {object} gin.H
// @Router /health [get]
func Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"message": "Emby Manager API服务正常运行",
	})
}

// Login 用户登录
// @Summary 用户登录
// @Description 用户身份验证
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body dto.LoginRequest true "登录信息"
// @Success 200 {object} dto.LoginResponse
// @Failure 400 {object} gin.H
// @Failure 401 {object} gin.H
// @Router /auth/login [post]
func Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "参数错误",
			"message": err.Error(),
		})
		return
	}

	response, err := authService.Login(&req)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"username": req.Username,
			"error":    err.Error(),
		}).Warn("用户登录失败")

		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "登录失败",
			"message": err.Error(),
		})
		return
	}

	logrus.WithFields(logrus.Fields{
		"user_id":  response.User.ID,
		"username": response.User.Username,
	}).Info("用户登录成功")

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// Register 用户注册
// @Summary 用户注册
// @Description 新用户注册
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body dto.RegisterRequest true "注册信息"
// @Success 201 {object} gin.H
// @Failure 400 {object} gin.H
// @Failure 409 {object} gin.H
// @Router /auth/register [post]
func Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "参数错误",
			"message": err.Error(),
		})
		return
	}

	userInfo, err := authService.Register(&req)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"username": req.Username,
			"email":    req.Email,
			"error":    err.Error(),
		}).Warn("用户注册失败")

		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "注册失败",
			"message": err.Error(),
		})
		return
	}

	logrus.WithFields(logrus.Fields{
		"user_id":  userInfo.ID,
		"username": userInfo.Username,
	}).Info("用户注册成功")

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "注册成功",
		"data":    userInfo,
	})
}

// RefreshToken 刷新令牌
// @Summary 刷新令牌
// @Description 使用现有令牌获取新令牌
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body dto.RefreshTokenRequest true "刷新请求"
// @Success 200 {object} gin.H
// @Failure 401 {object} gin.H
// @Router /auth/refresh [post]
func RefreshToken(c *gin.Context) {
	var req dto.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "参数错误",
			"message": err.Error(),
		})
		return
	}

	response, err := authService.RefreshToken(&req)
	if err != nil {
		logrus.WithField("error", err.Error()).Warn("令牌刷新失败")

		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "刷新失败",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// GetProfile 获取用户资料
// @Summary 获取用户资料
// @Description 获取当前登录用户的资料信息
// @Tags 用户
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} gin.H
// @Failure 401 {object} gin.H
// @Router /auth/profile [get]
func GetProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "未授权",
			"message": "用户信息不存在",
		})
		return
	}

	userInfo, err := authService.GetUserByID(userID.(uint))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "获取失败",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    userInfo,
	})
}

// UpdatePassword 更新密码
// @Summary 更新密码
// @Description 用户更新自己的密码
// @Tags 用户
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body dto.UpdatePasswordRequest true "密码更新请求"
// @Success 200 {object} gin.H
// @Failure 400 {object} gin.H
// @Failure 401 {object} gin.H
// @Router /auth/password [put]
func UpdatePassword(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "未授权",
			"message": "用户信息不存在",
		})
		return
	}

	var req dto.UpdatePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "参数错误",
			"message": err.Error(),
		})
		return
	}

	if err := authService.UpdatePassword(userID.(uint), &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "更新失败",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "密码更新成功",
	})
}

// GetUsers 获取用户列表占位符
func GetUsers(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "获取用户列表功能待实现",
	})
}

func GetUser(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "获取用户详情功能待实现",
	})
}

func UpdateUser(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "更新用户信息功能待实现",
	})
}

func DeleteUser(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "删除用户功能待实现",
	})
}

// Emby服务器相关的占位符函数
func GetEmbyServers(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "获取Emby服务器列表功能待实现",
	})
}

func CreateEmbyServer(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "创建Emby服务器功能待实现",
	})
}

func UpdateEmbyServer(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "更新Emby服务器功能待实现",
	})
}

func DeleteEmbyServer(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "删除Emby服务器功能待实现",
	})
}

func TestEmbyConnection(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "测试Emby连接功能待实现",
	})
}

// 媒体库相关的占位符函数
func GetMediaLibraries(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "获取媒体库列表功能待实现",
	})
}

func SearchMedia(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "媒体搜索功能待实现",
	})
}

func GetMediaItem(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "获取媒体详情功能待实现",
	})
}

// 播放控制相关的占位符函数
func PlaybackControl(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "播放控制功能待实现",
	})
}

func UpdateProgress(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "更新播放进度功能待实现",
	})
}

func GetPlaybackHistory(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "获取播放历史功能待实现",
	})
}

// 设备管理相关的占位符函数
func GetDevices(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "获取设备列表功能待实现",
	})
}

func UpdateDeviceStatus(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "更新设备状态功能待实现",
	})
}