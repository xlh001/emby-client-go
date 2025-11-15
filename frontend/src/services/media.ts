import request from './request'
import type { MediaLibrary, MediaItem } from '@/types'

export const getMediaLibraries = (serverId?: number) => {
  return request.get<MediaLibrary[]>('/api/media/libraries', {
    params: serverId ? { server_id: serverId } : {}
  })
}

export const getMediaLibrary = (id: number) => {
  return request.get<MediaLibrary>(`/api/media/libraries/${id}`)
}

export const getMediaItems = (params: {
  library_id: number
  type?: string
  limit?: number
  offset?: number
}) => {
  return request.get<{
    items: MediaItem[]
    total: number
    limit: number
    offset: number
  }>('/api/media/items', { params })
}

export const getMediaItem = (id: number) => {
  return request.get<MediaItem>(`/api/media/items/${id}`)
}

export const syncMediaLibraries = (serverId: number) => {
  return request.post(`/api/media/sync/${serverId}`)
}

export const syncAllServers = () => {
  return request.post('/api/media/sync-all')
}

export const getMediaStats = () => {
  return request.get('/api/media/stats')
}
