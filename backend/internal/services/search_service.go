package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/emby-client-go/backend/internal/models"
	"github.com/emby-client-go/backend/internal/database"
	"gorm.io/gorm"
)

// SearchRequest 搜索请求结构
type SearchRequest struct {
	Query          string   `json:"query" form:"query"`                     // 搜索关键词
	Types          []string `json:"types" form:"types"`                     // 媒体类型过滤
	ServerIDs      []uint   `json:"server_ids" form:"server_ids"`           // 服务器ID过滤
	LibraryIDs     []uint   `json:"library_ids" form:"library_ids"`         // 媒体库ID过滤
	Years          []int    `json:"years" form:"years"`                     // 年份过滤
	Genres         []string `json:"genres" form:"genres"`                   // 类型过滤
	SortBy         string   `json:"sort_by" form:"sort_by"`                 // 排序字段
	SortOrder      string   `json:"sort_order" form:"sort_order"`           // 排序方向
	Limit          int      `json:"limit" form:"limit"`                     // 每页数量
	Offset         int      `json:"offset" form:"offset"`                   // 偏移量
	IncludeSeries  bool     `json:"include_series" form:"include_series"`   // 是否包含系列信息
}

// SearchResult 搜索结果结构
type SearchResult struct {
	Total          int                    `json:"total"`           // 总数量
	Limit          int                    `json:"limit"`           // 每页数量
	Offset         int                    `json:"offset"`          // 偏移量
	Items          []models.MediaItem       `json:"items"`           // 媒体项目列表
	Aggregations   map[string]interface{}   `json:"aggregations"`     // 聚合信息（类型统计、服务器统计等）
	Suggestions    []string                `json:"suggestions"`      // 搜索建议
}

// SearchSuggestion 搜索建议结构
type SearchSuggestion struct {
	Text       string `json:"text"`
	Type       string `json:"type"`
	Count      int    `json:"count"`
	Confidence float64 `json:"confidence"`
}

// SearchService 搜索服务
type SearchService struct {
	db *gorm.DB
}

// NewSearchService 创建搜索服务实例
func NewSearchService() *SearchService {
	return &SearchService{
		db: database.DB,
	}
}

// Search 执行搜索
func (s *SearchService) Search(ctx context.Context, req SearchRequest) (*SearchResult, error) {
	// 参数验证和默认值
	if req.Limit <= 0 || req.Limit > 100 {
		req.Limit = 20
	}
	if req.SortBy == "" {
		req.SortBy = "relevance"
	}
	if req.SortOrder == "" {
		req.SortOrder = "desc"
	}

	// 构建查询
	query := s.buildSearchQuery(ctx, req)

	// 执行查询
	var items []models.MediaItem
	var total int64

	if err := query.Model(&models.MediaItem{}).Count(&total).Error; err != nil {
		return nil, fmt.Errorf("统计搜索结果失败: %w", err)
	}

	if err := query.Find(&items).Error; err != nil {
		return nil, fmt.Errorf("执行搜索失败: %w", err)
	}

	// 构建结果
	result := &SearchResult{
		Total:  int(total),
		Limit:  req.Limit,
		Offset: req.Offset,
		Items:  items,
	}

	// 添加聚合信息
	result.Aggregations = s.buildAggregations(ctx, req, items)

	// 添加搜索建议
	suggestions, err := s.GetSuggestions(ctx, req.Query, 5)
	if err == nil {
		result.Suggestions = suggestions
	}

	return result, nil
}

// buildSearchQuery 构建搜索查询
func (s *SearchService) buildSearchQuery(ctx context.Context, req SearchRequest) *gorm.DB {
	db := s.db.WithContext(ctx)

	// 基础查询，包含关联数据
	query := db.Model(&models.MediaItem{}).
		Preload("MediaLibrary").
		Preload("MediaLibrary.EmbyServer")

	// 关键词搜索
	if req.Query != "" {
		keyword := strings.TrimSpace(req.Query)
		if keyword != "" {
			// 全文搜索：名称、系列名称
			searchCondition := fmt.Sprintf("name LIKE ? OR series_name LIKE ?")
			searchParams := []interface{}{
				"%" + keyword + "%",
				"%" + keyword + "%",
			}

			// 支持模糊匹配和拼音搜索
			if len(keyword) >= 2 {
				// 添加拼音搜索逻辑（可以后续扩展）
				searchCondition += " OR name COLLATE utf8_unicode_ci LIKE ?"
				searchParams = append(searchParams, keyword+"%")
			}

			query = query.Where(searchCondition, searchParams...)
		}
	}

	// 类型过滤
	if len(req.Types) > 0 {
		query = query.Where("type IN ?", req.Types)
	}

	// 服务器过滤
	if len(req.ServerIDs) > 0 {
		query = query.Joins("JOIN media_libraries ON media_items.media_library_id = media_libraries.id").
			Where("media_libraries.emby_server_id IN ?", req.ServerIDs)
	}

	// 媒体库过滤
	if len(req.LibraryIDs) > 0 {
		query = query.Where("media_library_id IN ?", req.LibraryIDs)
	}

	// 年份过滤
	if len(req.Years) > 0 {
		query = query.Where("year IN ?", req.Years)
	}

	// 排序
	switch req.SortBy {
	case "name":
		query = query.Order("name " + req.SortOrder)
	case "created_at":
		query = query.Order("created_at " + req.SortOrder)
	case "updated_at":
		query = query.Order("updated_at " + req.SortOrder)
	case "year":
		query = query.Order("year " + req.SortOrder + ", name " + req.SortOrder)
	case "relevance":
		fallthrough
	default:
		// 相关性排序：名称完全匹配 > 系列名称匹配 > 名称部分匹配
		if req.Query != "" {
			keyword := strings.TrimSpace(req.Query)
			relevanceCase := fmt.Sprintf(`
				CASE
					WHEN name = ? THEN 100
					WHEN name LIKE ? THEN 90
					WHEN series_name = ? THEN 80
					WHEN series_name LIKE ? THEN 70
					ELSE 50
				END
			`)
			query = query.Order(gorm.Expr(relevanceCase + " " + req.SortOrder, keyword, keyword+"%", keyword, keyword+"%"))
		} else {
			query = query.Order("updated_at " + req.SortOrder)
		}
	}

	// 分页
	query = query.Offset(req.Offset).Limit(req.Limit)

	return query
}

// buildAggregations 构建聚合信息
func (s *SearchService) buildAggregations(ctx context.Context, req SearchRequest, items []models.MediaItem) map[string]interface{} {
	aggregations := make(map[string]interface{})

	// 类型统计
	typeStats := make(map[string]int)
	yearStats := make(map[int]int)
	serverStats := make(map[string]int)

	for _, item := range items {
		// 类型统计
		typeStats[item.Type]++

		// 年份统计
		if item.Year != nil {
			yearStats[*item.Year]++
		}

		// 服务器统计
		if item.MediaLibrary.EmbyServer.Name != "" {
			serverStats[item.MediaLibrary.EmbyServer.Name]++
		}
	}

	aggregations["types"] = typeStats
	aggregations["years"] = yearStats
	aggregations["servers"] = serverStats

	return aggregations
}

// GetSuggestions 获取搜索建议
func (s *SearchService) GetSuggestions(ctx context.Context, query string, limit int) ([]string, error) {
	if len(strings.TrimSpace(query)) < 2 {
		return nil, nil
	}

	var suggestions []string

	// 从历史搜索记录或热门标签获取建议（这里简化实现）
	var names []string
	err := s.db.WithContext(ctx).
		Model(&models.MediaItem{}).
		Where("name LIKE ? OR series_name LIKE ?",
			strings.TrimSpace(query)+"%", strings.TrimSpace(query)+"%").
		Limit(limit).
		Pluck("DISTINCT name", &names).Error

	if err != nil {
		return nil, fmt.Errorf("获取搜索建议失败: %w", err)
	}

	for _, name := range names {
		if len(name) > 0 && !contains(suggestions, name) {
			suggestions = append(suggestions, name)
		}
	}

	return suggestions, nil
}

// contains 检查字符串是否在切片中
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// GetPopularKeywords 获取热门搜索关键词
func (s *SearchService) GetPopularKeywords(ctx context.Context, limit int) ([]string, error) {
	// 这里可以从搜索日志或统计表中获取
	// 简化实现，返回一些示例关键词
	keywords := []string{"电影", "电视剧", "音乐", "动作", "科幻", "喜剧"}

	if limit > 0 && len(keywords) > limit {
		keywords = keywords[:limit]
	}

	return keywords, nil
}

// GetSearchHistory 获取用户搜索历史
func (s *SearchService) GetSearchHistory(ctx context.Context, userID uint, limit int) ([]string, error) {
	// 这里需要实现搜索历史表记录用户搜索行为
	// 简化实现，返回空历史
	return []string{}, nil
}