package handlers

import (
	"net/http"
	"strconv"

	"github.com/emby-client-go/backend/internal/services"
	"github.com/gin-gonic/gin"
)

// PlaybackHandler 播放控制处理器
type PlaybackHandler struct {
	playbackService *services.PlaybackService
}

// NewPlaybackHandler 创建播放控制处理器
func NewPlaybackHandler() *PlaybackHandler {
	return &PlaybackHandler{
		playbackService: services.NewPlaybackService(),
	}
}

// SendPlayCommand 发送播放控制命令
// @Summary 发送播放控制命令
// @Tags Playback
// @Security BearerAuth
// @Param server_id path int true "服务器ID"
// @Param device_id path int true "设备ID"
// @Param command body services.PlayCommand true "播放命令"
// @Success 200 {object} map[string]interface{}
// @Router /api/playback/:server_id/:device_id/command [post]
func (h *PlaybackHandler) SendPlayCommand(c *gin.Context) {
	serverID, _ := strconv.ParseUint(c.Param("server_id"), 10, 32)
	deviceID, _ := strconv.ParseUint(c.Param("device_id"), 10, 32)

	var cmd services.PlayCommand
	if err := c.ShouldBindJSON(&cmd); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的命令格式"})
		return
	}

	if err := h.playbackService.SendPlayCommand(c.Request.Context(), uint(serverID), uint(deviceID), cmd); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "命令发送成功"})
}

// GetActiveSessions 获取活动播放会话
// @Summary 获取活动播放会话
// @Tags Playback
// @Security BearerAuth
// @Param server_id query int true "服务器ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/playback/sessions [get]
func (h *PlaybackHandler) GetActiveSessions(c *gin.Context) {
	serverID, _ := strconv.ParseUint(c.Query("server_id"), 10, 32)

	sessions, err := h.playbackService.GetActiveSessions(uint(serverID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": sessions})
}

// GetPlaybackHistory 获取播放历史
// @Summary 获取播放历史
// @Tags Playback
// @Security BearerAuth
// @Param limit query int false "每页数量"
// @Param offset query int false "偏移量"
// @Success 200 {object} map[string]interface{}
// @Router /api/playback/history [get]
func (h *PlaybackHandler) GetPlaybackHistory(c *gin.Context) {
	userID := c.GetUint("user_id")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	records, total, err := h.playbackService.GetPlaybackHistory(userID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"records": records,
			"total":   total,
			"limit":   limit,
			"offset":  offset,
		},
	})
}
