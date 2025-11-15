package handlers

import (
	"net/http"
	"strconv"

	"github.com/emby-client-go/backend/internal/dto"
	"github.com/emby-client-go/backend/internal/models"
	"github.com/emby-client-go/backend/internal/services"
	"github.com/gin-gonic/gin"
)

type ServerHandler struct {
	serverService *services.ServerService
}

func NewServerHandler() *ServerHandler {
	return &ServerHandler{
		serverService: services.NewServerService(),
	}
}

// CreateServer 创建服务器
// @Summary 创建Emby服务器
// @Description 添加新的Emby服务器并测试连接
// @Tags 服务器管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body dto.CreateServerRequest true "服务器信息"
// @Success 200 {object} dto.ApiResponse{data=dto.ServerResponse}
// @Failure 400 {object} dto.ApiResponse
// @Router /server/create [post]
func (h *ServerHandler) CreateServer(c *gin.Context) {
	var req dto.CreateServerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ApiResponse{
			Code:    400,
			Message: "参数错误: " + err.Error(),
		})
		return
	}

	userID, _ := c.Get("user_id")

	server := &models.EmbyServer{
		Name:        req.Name,
		URL:         req.URL,
		APIKey:      req.APIKey,
		Description: req.Description,
	}

	if err := h.serverService.CreateServer(server, userID.(uint)); err != nil {
		c.JSON(http.StatusBadRequest, dto.ApiResponse{
			Code:    400,
			Message: err.Error(),
		})
		return
	}

	response := dto.ServerResponse{
		ID:          server.ID,
		Name:        server.Name,
		URL:         server.URL,
		Version:     server.Version,
		OS:          server.OS,
		Status:      server.Status,
		Description: server.Description,
		CreatedAt:   server.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:   server.UpdatedAt.Format("2006-01-02 15:04:05"),
	}

	if server.LastCheck != nil {
		response.LastCheck = server.LastCheck.Format("2006-01-02 15:04:05")
	}

	c.JSON(http.StatusOK, dto.ApiResponse{
		Code:    200,
		Message: "服务器创建成功",
		Data:    response,
	})
}

// GetServers 获取服务器列表
// @Summary 获取服务器列表
// @Description 获取当前用户的Emby服务器列表
// @Tags 服务器管理
// @Produce json
// @Security ApiKeyAuth
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Success 200 {object} dto.ApiResponse{data=dto.PageResponse}
// @Failure 500 {object} dto.ApiResponse
// @Router /server/list [get]
func (h *ServerHandler) GetServers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	userID, _ := c.Get("user_id")

	servers, total, err := h.serverService.GetServers(userID.(uint), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ApiResponse{
			Code:    500,
			Message: "获取服务器列表失败",
		})
		return
	}

	var serverResponses []dto.ServerResponse
	for _, server := range servers {
		response := dto.ServerResponse{
			ID:          server.ID,
			Name:        server.Name,
			URL:         server.URL,
			Version:     server.Version,
			OS:          server.OS,
			Status:      server.Status,
			Description: server.Description,
			CreatedAt:   server.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:   server.UpdatedAt.Format("2006-01-02 15:04:05"),
		}

		if server.LastCheck != nil {
			response.LastCheck = server.LastCheck.Format("2006-01-02 15:04:05")
		}

		serverResponses = append(serverResponses, response)
	}

	pageResponse := dto.PageResponse{
		List:     serverResponses,
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

// GetServer 获取服务器详情
// @Summary 获取服务器详情
// @Description 获取指定服务器的详细信息
// @Tags 服务器管理
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "服务器ID"
// @Success 200 {object} dto.ApiResponse{data=dto.ServerResponse}
// @Failure 404 {object} dto.ApiResponse
// @Router /server/{id} [get]
func (h *ServerHandler) GetServer(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	server, err := h.serverService.GetServer(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ApiResponse{
			Code:    404,
			Message: err.Error(),
		})
		return
	}

	response := dto.ServerResponse{
		ID:          server.ID,
		Name:        server.Name,
		URL:         server.URL,
		Version:     server.Version,
		OS:          server.OS,
		Status:      server.Status,
		Description: server.Description,
		CreatedAt:   server.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:   server.UpdatedAt.Format("2006-01-02 15:04:05"),
	}

	if server.LastCheck != nil {
		response.LastCheck = server.LastCheck.Format("2006-01-02 15:04:05")
	}

	c.JSON(http.StatusOK, dto.ApiResponse{
		Code:    200,
		Message: "获取成功",
		Data:    response,
	})
}

// UpdateServer 更新服务器
// @Summary 更新服务器信息
// @Description 更新指定服务器的信息
// @Tags 服务器管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "服务器ID"
// @Param request body dto.UpdateServerRequest true "服务器信息"
// @Success 200 {object} dto.ApiResponse
// @Failure 400 {object} dto.ApiResponse
// @Router /server/{id} [put]
func (h *ServerHandler) UpdateServer(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	var req dto.UpdateServerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ApiResponse{
			Code:    400,
			Message: "参数错误: " + err.Error(),
		})
		return
	}

	updates := make(map[string]interface{})
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.URL != "" {
		updates["url"] = req.URL
	}
	if req.APIKey != "" {
		updates["api_key"] = req.APIKey
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}

	if err := h.serverService.UpdateServer(uint(id), updates); err != nil {
		c.JSON(http.StatusBadRequest, dto.ApiResponse{
			Code:    400,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.ApiResponse{
		Code:    200,
		Message: "更新成功",
	})
}

// DeleteServer 删除服务器
// @Summary 删除服务器
// @Description 删除指定的服务器
// @Tags 服务器管理
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "服务器ID"
// @Success 200 {object} dto.ApiResponse
// @Failure 400 {object} dto.ApiResponse
// @Router /server/{id} [delete]
func (h *ServerHandler) DeleteServer(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	if err := h.serverService.DeleteServer(uint(id)); err != nil {
		c.JSON(http.StatusBadRequest, dto.ApiResponse{
			Code:    400,
			Message: "删除失败",
		})
		return
	}

	c.JSON(http.StatusOK, dto.ApiResponse{
		Code:    200,
		Message: "删除成功",
	})
}

// TestConnection 测试服务器连接
// @Summary 测试服务器连接
// @Description 测试指定服务器的连接状态
// @Tags 服务器管理
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "服务器ID"
// @Success 200 {object} dto.ApiResponse{data=dto.TestConnectionResponse}
// @Failure 400 {object} dto.ApiResponse
// @Router /server/{id}/test [post]
func (h *ServerHandler) TestConnection(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	userID, _ := c.Get("user_id")

	duration, err := h.serverService.TestConnection(uint(id), userID.(uint))

	response := dto.TestConnectionResponse{}

	if err != nil {
		response.Status = "failed"
		response.Message = err.Error()
		response.ResponseTime = 0

		c.JSON(http.StatusOK, dto.ApiResponse{
			Code:    200,
			Message: "连接测试失败",
			Data:    response,
		})
		return
	}

	response.Status = "success"
	response.Message = "连接成功"
	response.ResponseTime = duration.Milliseconds()

	c.JSON(http.StatusOK, dto.ApiResponse{
		Code:    200,
		Message: "连接测试成功",
		Data:    response,
	})
}

// SyncDevices 同步设备列表
// @Summary 同步设备列表
// @Description 从Emby服务器同步设备列表
// @Tags 服务器管理
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "服务器ID"
// @Success 200 {object} dto.ApiResponse
// @Failure 400 {object} dto.ApiResponse
// @Router /server/{id}/sync-devices [post]
func (h *ServerHandler) SyncDevices(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	if err := h.serverService.SyncDevices(uint(id)); err != nil {
		c.JSON(http.StatusBadRequest, dto.ApiResponse{
			Code:    400,
			Message: "同步设备失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.ApiResponse{
		Code:    200,
		Message: "设备同步成功",
	})
}

// SyncLibraries 同步媒体库列表
// @Summary 同步媒体库列表
// @Description 从Emby服务器同步媒体库列表
// @Tags 服务器管理
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "服务器ID"
// @Success 200 {object} dto.ApiResponse
// @Failure 400 {object} dto.ApiResponse
// @Router /server/{id}/sync-libraries [post]
func (h *ServerHandler) SyncLibraries(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	if err := h.serverService.SyncLibraries(uint(id)); err != nil {
		c.JSON(http.StatusBadRequest, dto.ApiResponse{
			Code:    400,
			Message: "同步媒体库失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.ApiResponse{
		Code:    200,
		Message: "媒体库同步成功",
	})
}
