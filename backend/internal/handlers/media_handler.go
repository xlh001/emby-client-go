package handlers

import (
	"net/http"
	"strconv"

	"github.com/emby-client-go/backend/internal/services"
	"github.com/gin-gonic/gin"
)

// MediaHandler 媒体库处理器
type MediaHandler struct {
	mediaService *services.MediaService
}

// NewMediaHandler 创建媒体库处理器
func NewMediaHandler() *MediaHandler {
	return &MediaHandler{
		mediaService: services.NewMediaService(),
	}
}

// GetMediaLibraries 获取媒体库列表
// @Summary 获取媒体库列表
// @Description 获取指定服务器或所有服务器的媒体库列表
// @Tags Media
// @Security BearerAuth
// @Param server_id query int false "服务器ID（可选）"
// @Success 200 {object} map[string]interface{} "媒体库列表"
// @Failure 400 {object} map[string]interface{} "请求错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Failure 500 {object} map[string]interface{} "服务器错误"
// @Router /api/media/libraries [get]
func (h *MediaHandler) GetMediaLibraries(c *gin.Context) {
	serverIDStr := c.Query("server_id")

	if serverIDStr != "" {
		// 获取特定服务器的媒体库
		serverID, err := strconv.ParseUint(serverIDStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的服务器ID"})
			return
		}

		libraries, err := h.mediaService.GetMediaLibraries(uint(serverID))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"code":    200,
			"message": "获取成功",
			"data":    libraries,
		})
	} else {
		// 获取所有媒体库
		libraries, err := h.mediaService.GetAllMediaLibraries()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"code":    200,
			"message": "获取成功",
			"data":    libraries,
		})
	}
}

// GetMediaLibrary 获取单个媒体库详情
// @Summary 获取媒体库详情
// @Description 获取指定ID的媒体库详细信息
// @Tags Media
// @Security BearerAuth
// @Param id path int true "媒体库ID"
// @Success 200 {object} map[string]interface{} "媒体库详情"
// @Failure 400 {object} map[string]interface{} "请求错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Failure 404 {object} map[string]interface{} "媒体库不存在"
// @Router /api/media/libraries/:id [get]
func (h *MediaHandler) GetMediaLibrary(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的媒体库ID"})
		return
	}

	library, err := h.mediaService.GetMediaLibrary(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "媒体库不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data":    library,
	})
}

// SyncMediaLibraries 同步服务器媒体库
// @Summary 同步媒体库
// @Description 同步指定服务器的媒体库信息
// @Tags Media
// @Security BearerAuth
// @Param id path int true "服务器ID"
// @Success 200 {object} map[string]interface{} "同步结果"
// @Failure 400 {object} map[string]interface{} "请求错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Failure 500 {object} map[string]interface{} "同步失败"
// @Router /api/media/sync/:id [post]
func (h *MediaHandler) SyncMediaLibraries(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的服务器ID"})
		return
	}

	count, err := h.mediaService.SyncMediaLibraries(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "同步成功",
		"data": gin.H{
			"count": count,
		},
	})
}

// SyncAllServers 同步所有服务器媒体库
// @Summary 同步所有服务器
// @Description 同步所有在线服务器的媒体库信息
// @Tags Media
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "同步结果"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Failure 500 {object} map[string]interface{} "同步失败"
// @Router /api/media/sync-all [post]
func (h *MediaHandler) SyncAllServers(c *gin.Context) {
	results, err := h.mediaService.SyncAllServers(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "同步完成",
		"data":    results,
	})
}

// RefreshMediaLibrary 刷新媒体库统计
// @Summary 刷新媒体库统计
// @Description 刷新指定媒体库的统计信息
// @Tags Media
// @Security BearerAuth
// @Param id path int true "媒体库ID"
// @Success 200 {object} map[string]interface{} "刷新成功"
// @Failure 400 {object} map[string]interface{} "请求错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Failure 500 {object} map[string]interface{} "刷新失败"
// @Router /api/media/libraries/:id/refresh [post]
func (h *MediaHandler) RefreshMediaLibrary(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的媒体库ID"})
		return
	}

	if err := h.mediaService.RefreshMediaLibrary(c.Request.Context(), uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "刷新成功",
	})
}

// GetMediaLibraryStats 获取媒体库统计
// @Summary 获取媒体库统计
// @Description 获取所有媒体库的统计信息
// @Tags Media
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "统计信息"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Failure 500 {object} map[string]interface{} "获取失败"
// @Router /api/media/stats [get]
func (h *MediaHandler) GetMediaLibraryStats(c *gin.Context) {
	stats, err := h.mediaService.GetMediaLibraryStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data":    stats,
	})
}

// GetMediaItems 获取媒体项目列表
// @Summary 获取媒体项目列表
// @Description 分页获取指定媒体库的媒体项目
// @Tags Media
// @Security BearerAuth
// @Param library_id query int true "媒体库ID"
// @Param type query string false "媒体类型过滤"
// @Param limit query int false "每页数量，默认20"
// @Param offset query int false "偏移量，默认0"
// @Success 200 {object} map[string]interface{} "媒体项目列表"
// @Router /api/media/items [get]
func (h *MediaHandler) GetMediaItems(c *gin.Context) {
	libraryIDStr := c.Query("library_id")
	if libraryIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少媒体库ID"})
		return
	}

	libraryID, err := strconv.ParseUint(libraryIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的媒体库ID"})
		return
	}

	mediaType := c.Query("type")
	limit := 20
	offset := 0

	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	items, total, err := h.mediaService.GetMediaItems(uint(libraryID), mediaType, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data": gin.H{
			"items":  items,
			"total":  total,
			"limit":  limit,
			"offset": offset,
		},
	})
}

// GetMediaItem 获取媒体项目详情
// @Summary 获取媒体项目详情
// @Description 获取指定ID的媒体项目详细信息
// @Tags Media
// @Security BearerAuth
// @Param id path int true "媒体项目ID"
// @Success 200 {object} map[string]interface{} "媒体项目详情"
// @Router /api/media/items/:id [get]
func (h *MediaHandler) GetMediaItem(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的媒体项目ID"})
		return
	}

	item, err := h.mediaService.GetMediaItem(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "媒体项目不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data":    item,
	})
}
