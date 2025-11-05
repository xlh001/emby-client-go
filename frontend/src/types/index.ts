// 通用响应接口
export interface ApiResponse<T = any> {
  code: number
  message: string
  data: T
  timestamp: string
}

// 用户相关接口
export interface LoginRequest {
  username: string
  password: string
}

export interface LoginResponse {
  token: string
  user: User
}

export interface User {
  id: number
  username: string
  email: string
  role: 'admin' | 'user'
  avatar?: string
  status: 'active' | 'inactive' | 'banned'
  created_at: string
  updated_at: string
}

// Emby服务器相关接口
export interface EmbyServer {
  id: number
  name: string
  url: string
  version?: string
  status: 'online' | 'offline' | 'error'
  last_check?: string
  description?: string
  created_by: number
  created_at: string
  updated_at: string
  creator?: User
  device_count?: number
  library_count?: number
}

export interface EmbyServerCreate {
  name: string
  url: string
  api_key: string
  description?: string
}

export interface EmbyServerUpdate {
  name?: string
  url?: string
  api_key?: string
  description?: string
}

export interface ConnectionTestResult {
  success: boolean
  message: string
  version?: string
}

// 媒体库相关接口
export interface MediaLibrary {
  id: number
  server_id: number
  emby_id: string
  name: string
  type: 'movies' | 'tvshows' | 'music' | 'photos'
  path: string
  item_count: number
  synced_at?: string
  created_at: string
  updated_at: string
}

export interface MediaItem {
  id: number
  library_id: number
  emby_id: string
  title: string
  type: 'Movie' | 'Series' | 'Episode' | 'MusicAlbum' | 'MusicTrack' | 'Photo'
  path: string
  size?: number
  duration?: number
  year?: number
  genres?: string[]
  rating?: number
  parental_rating?: string
  thumbnail?: string
  synced_at?: string
  created_at: string
  updated_at: string
  library?: MediaLibrary
}

export interface MediaSearchParams {
  keyword?: string
  type?: string
  library_id?: number
  server_id?: number
  page?: number
  page_size?: number
  sort_by?: 'title' | 'year' | 'rating' | 'created_at'
  sort_order?: 'asc' | 'desc'
}

export interface MediaSearchResponse {
  items: MediaItem[]
  total: number
  page: number
  page_size: number
  total_pages: number
}

// 设备相关接口
export interface Device {
  id: number
  name: string
  device_id: string
  user_id: number
  server_id?: number
  client: string
  last_activity?: string
  ip: string
  user_agent: string
  status: 'active' | 'inactive'
  created_at: string
  updated_at: string
  user?: User
  emby_server?: EmbyServer
  playback_records_count?: number
}

// 播放相关接口
export interface PlaybackRecord {
  id: number
  device_id: number
  server_id: number
  media_item_id: number
  user_id: number
  position: number
  duration: number
  completed: boolean
  played_at: string
  synced: boolean
  created_at: string
  updated_at: string
  device?: Device
  emby_server?: EmbyServer
  media_item?: MediaItem
  user?: User
}

export interface PlaybackControlRequest {
  action: 'play' | 'pause' | 'stop' | 'seek' | 'next' | 'previous'
  device_id: number
  server_id: number
  media_item_id: number
  position?: number
  volume?: number
}

export interface PlaybackProgressRequest {
  device_id: number
  server_id: number
  media_item_id: number
  position: number
  duration: number
  completed: boolean
}

// 路由元信息
export interface RouteMetaInfo {
  title: string
  icon?: string
  requiresAuth?: boolean
  roles?: string[]
  keepAlive?: boolean
}

// 分页信息
export interface Pagination {
  page: number
  page_size: number
  total: number
  total_pages: number
}