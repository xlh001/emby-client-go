export interface MediaLibrary {
  id: number
  emby_server_id: number
  name: string
  type: string
  path?: string
  total_items: number
  total_size: number
  last_refresh?: string
  collection_type: string
  created_at: string
  updated_at: string
  emby_server?: {
    id: number
    name: string
    url: string
    status: string
  }
}

export interface MediaItem {
  id: number
  media_library_id: number
  emby_item_id: string
  name: string
  type: string
  path?: string
  parent_id?: string
  series_name?: string
  season_number?: number
  episode_number?: number
  year?: number
  run_time_ticks: number
  size: number
  container?: string
  video_codec?: string
  audio_codec?: string
  resolution?: string
  created_at: string
  updated_at: string
  media_library?: MediaLibrary
}
