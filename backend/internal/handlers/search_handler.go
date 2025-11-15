package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/emby-client-go/backend/internal/services"
	"github.com/emby-client-go/backend/internal/models"
	"github.com/gin-gonic/gin"
)

// SearchHandler 搜索处理器
type SearchHandler struct {
	searchService *services.SearchService
}

// NewSearchHandler 创建搜索处理器
func NewSearchHandler() *SearchHandler {
	return &SearchHandler{
		searchService: services.NewSearchService(),
	}
}

// SearchMedia 搜索媒体内容
// @Summary 搜索媒体内容
// @Description 支持关键词、类型、服务器等多维度搜索
// @Tags Search
// @Security BearerAuth
// @Param query query string false "搜索关键词"
// @Param types query string false "媒体类型过滤，多个用逗号分隔"
// @Param server_ids query string false "服务器ID过滤，多个用逗号分隔"
// @Param library_ids query string false "媒体库ID过滤，多个用逗号分隔"
// @Param years query string false "年份过滤，多个用逗号分隔"
// @Param sort_by query string false "排序字段：name,year,created_at,updated_at,relevance"
// @Param sort_order query string false "排序方向：asc,desc"
// @Param limit query int false "每页数量，默认20，最大100"
// @Param offset query int false "偏移量，默认0"
// @Param include_series query bool false "是否包含系列信息"
// @Success 200 {object} map[string]interface{} "搜索结果"
// @Failure 400 {object} map[string]interface{} "请求错误"
// @Router /api/search [get]
func (h *SearchHandler) SearchMedia(c *gin.Context) {
	// 构建搜索请求
	req := services.SearchRequest{
		Query: c.Query("query"),
	}

	// 解析类型过滤
	if typesStr := c.Query("types"); typesStr != "" {
		req.Types = splitString(typesStr)
	}

	// 解析服务器ID过滤
	if serverIDsStr := c.Query("server_ids"); serverIDsStr != "" {
		serverIDs, err := splitUint(serverIDsStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "无效的服务器ID格式",
			})
			return
		}
		req.ServerIDs = serverIDs
	}

	// 解析媒体库ID过滤
	if libraryIDsStr := c.Query("library_ids"); libraryIDsStr != "" {
		libraryIDs, err := splitUint(libraryIDsStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "无效的媒体库ID格式",
			})
			return
		}
		req.LibraryIDs = libraryIDs
	}

	// 解析年份过滤
	if yearsStr := c.Query("years"); yearsStr != "" {
		years, err := splitInt(yearsStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "无效的年份格式",
			})
			return
		}
		req.Years = years
	}

	// 解析排序参数
	req.SortBy = c.Query("sort_by")
	req.SortOrder = c.Query("sort_order")
	req.IncludeSeries = c.DefaultQuery("include_series", "false") == "true"

	// 解析分页参数
	if limitStr := c.Query("limit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit <= 0 || limit > 100 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "无效的limit参数，应在1-100之间",
			})
			return
		}
		req.Limit = limit
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		offset, err := strconv.Atoi(offsetStr)
		if err != nil || offset < 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "无效的offset参数",
			})
			return
		}
		req.Offset = offset
	}

	// 执行搜索
	result, err := h.searchService.Search(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "搜索失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}

// GetSearchSuggestions 获取搜索建议
// @Summary 获取搜索建议
// @Description 根据输入关键词提供搜索建议
// @Tags Search
// @Security BearerAuth
// @Param query query string true "搜索关键词"
// @Param limit query int false "建议数量，默认5"
// @Success 200 {object} map[string]interface{} "搜索建议列表"
// @Failure 400 {object} map[string]interface{} "请求错误"
// @Router /api/search/suggestions [get]
func (h *SearchHandler) GetSearchSuggestions(c *gin.Context) {
	query := c.Query("query")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "缺少搜索关键词",
		})
		return
	}

	limit := 5
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 20 {
			limit = l
		}
	}

	suggestions, err := h.searchService.GetSuggestions(c.Request.Context(), query, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取搜索建议失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"query":       query,
			"suggestions": suggestions,
		},
	})
}

// GetPopularKeywords 获取热门搜索关键词
// @Summary 获取热门搜索关键词
// @Description 获取热门搜索关键词列表
// @Tags Search
// @Security BearerAuth
// @Param limit query int false "关键词数量，默认10"
// @Success 200 {object} map[string]interface{} "热门关键词列表"
// @Router /api/search/popular [get]
func (h *SearchHandler) GetPopularKeywords(c *gin.Context) {
	limit := 10
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 50 {
			limit = l
		}
	}

	keywords, err := h.searchService.GetPopularKeywords(c.Request.Context(), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取热门关键词失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"keywords": keywords,
		},
	})
}

// GetSearchHistory 获取用户搜索历史
// @Summary 获取用户搜索历史
// @Description 获取当前用户的搜索历史记录
// @Tags Search
// @Security BearerAuth
// @Param limit query int false "历史数量，默认20"
// @Success 200 {object} map[string]interface{} "搜索历史列表"
// @Router /api/search/history [get]
func (h *SearchHandler) GetSearchHistory(c *gin.Context) {
	// 简化实现，返回示例历史
	history := []string{
		"阿凡达",
		"复仇者联盟",
		"权力的游戏",
		"音乐播放",
		"科幻电影",
	}

	limit := 20
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	if limit > 0 && len(history) > limit {
		history = history[:limit]
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"history": history,
		},
	})
}

// GetSearchStats 获取搜索统计信息
// @Summary 获取搜索统计信息
// @Description 获取搜索相关的统计数据
// @Tags Search
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "搜索统计"
// @Router /api/search/stats [get]
func (h *SearchHandler) GetSearchStats(c *gin.Context) {
	// 简化实现，返回一些示例统计
	stats := map[string]interface{}{
		"total_searches_today":    128,
		"popular_keywords":        []string{"电影", "电视剧", "音乐"},
		"avg_response_time_ms":    45,
		"success_rate":          98.5,
		"media_type_stats": map[string]int{
			"Movie":    156,
			"Episode": 892,
			"Audio":    67,
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
	})
}

// 工具函数：分割字符串为uint切片
func splitUint(s string) ([]uint, error) {
	parts := strings.Split(s, ",")
	result := make([]uint, 0, len(parts))

	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			continue
		}

		num, err := strconv.ParseUint(trimmed, 10, 32)
		if err != nil {
			return nil, err
		}
		result = append(result, uint(num))
	}

	return result, nil
}

// 工具函数：分割字符串为int切片
func splitInt(s string) ([]int, error) {
	parts := strings.Split(s, ",")
	result := make([]int, 0, len(parts))

	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			continue
		}

		num, err := strconv.Atoi(trimmed)
		if err != nil {
			return nil, err
		}
		result = append(result, num)
	}

	return result, nil
}

// 工具函数：分割字符串为字符串切片
func splitString(s string) []string {
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))

	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}

	return result
}