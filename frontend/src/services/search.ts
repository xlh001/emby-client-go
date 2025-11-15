import request from '@/utils/request'
import type { AxiosResponse } from 'axios'

// 搜索请求参数接口
export interface SearchRequest {
  query?: string // 搜索关键词
  types?: string[] // 媒体类型过滤，多个用逗号分隔
  server_ids?: string[] // 服务器ID过滤，多个用逗号分隔
  library_ids?: string[] // 媒体库ID过滤，多个用逗号分隔
  years?: number[] // 年份过滤，多个用逗号分隔
  sort_by?: string // 排序字段：name,year,created_at,updated_at,relevance
  sort_order?: string // 排序方向：asc,desc
  limit?: number // 每页数量，默认20，最大100
  offset?: number // 偏移量，默认0
  include_series?: boolean // 是否包含系列信息
}

// 媒体项目接口
export interface MediaItem {
  id: number
  media_library_id: number
  emby_item_id: string
  name: string
  type: string // Movie, Episode, Audio, etc.
  series_name?: string
  year?: number
  premiere_date?: string
  community_rating?: number
  critic_rating?: number
  overview?: string
  run_time_ticks?: number
  size?: number
  media_streams?: any[]
  people?: any[]
  studios?: any[]
  genres?: any[]
  parent_id?: number
  path?: string
  created_at: string
  updated_at: string
  media_library?: {
    id: number
    name: string
    type: string
    emby_server_id: number
    emby_server?: {
      id: number
      name: string
      url: string
      user_id: string
    }
  }
}

// 搜索结果接口
export interface SearchResult {
  total: number // 总数量
  limit: number // 每页数量
  offset: number // 偏移量
  items: MediaItem[] // 媒体项目列表
  aggregations: Record<string, any> // 聚合信息（类型统计、服务器统计等）
  suggestions: string[] // 搜索建议
}

// 搜索建议接口
export interface SearchSuggestion {
  text: string
  type: string
  count: number
  confidence: number
}

// API响应包装接口
export interface ApiResponse<T = any> {
  success: boolean
  data?: T
  message?: string
  code?: number
}

// 搜索媒体内容
export const searchMedia = async (params: SearchRequest): Promise<AxiosResponse<ApiResponse<SearchResult>>> => {
  return request.get('/api/search', { params })
}

// 获取搜索建议
export const getSearchSuggestions = async (query: string, limit = 5): Promise<AxiosResponse<ApiResponse<{ query: string; suggestions: string[] }>>> => {
  return request.get('/api/search/suggestions', {
    params: { query, limit }
  })
}

// 获取热门搜索关键词
export const getPopularKeywords = async (limit = 10): Promise<AxiosResponse<ApiResponse<{ keywords: string[] }>>> => {
  return request.get('/api/search/popular', {
    params: { limit }
  })
}

// 获取用户搜索历史
export const getSearchHistory = async (limit = 20): Promise<AxiosResponse<ApiResponse<{ history: string[] }>>> => {
  return request.get('/api/search/history', {
    params: { limit }
  })
}

// 获取搜索统计信息
export const getSearchStats = async (): Promise<AxiosResponse<ApiResponse<Record<string, any>>>> => {
  return request.get('/api/search/stats')
}