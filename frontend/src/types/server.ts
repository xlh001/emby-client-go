// Emby服务器相关类型定义

export interface EmbyServer {
  id: number
  name: string
  url: string
  api_key: string
  version?: string
  os?: string
  status: 'online' | 'offline' | 'error'
  last_check?: string
  description?: string
  created_at: string
  updated_at: string
}

export interface MediaLibrary {
  id: number
  emby_server_id: number
  name: string
  type: 'movies' | 'tvshows' | 'music' | 'photos'
  path?: string
  total_items: number
  total_size: number
  last_refresh?: string
  collection_type?: string
  created_at: string
  updated_at: string
}

export interface Device {
  id: number
  emby_server_id: number
  emby_device_id: string
  name: string
  device_type?: string
  app_name?: string
  app_version?: string
  last_user_name?: string
  last_activity_at?: string
  is_active: boolean
  created_at: string
  updated_at: string
}

export interface ConnectionLog {
  id: number
  emby_server_id: number
  user_id: number
  action: 'connect' | 'disconnect' | 'error'
  message?: string
  response_time?: number
  status: 'success' | 'failed'
  created_at: string
}
