import request from './request'
import type { EmbyServer } from '@/types/server'

// 服务器API服务

/**
 * 获取服务器列表
 */
export function getServers() {
  return request.get<EmbyServer[]>('/server/list')
}

/**
 * 获取单个服务器详情
 */
export function getServer(id: number) {
  return request.get<EmbyServer>(`/server/${id}`)
}

/**
 * 创建服务器
 */
export function createServer(data: {
  name: string
  url: string
  api_key: string
  description?: string
}) {
  return request.post<EmbyServer>('/server/create', data)
}

/**
 * 更新服务器
 */
export function updateServer(id: number, data: {
  name?: string
  url?: string
  api_key?: string
  description?: string
}) {
  return request.put<EmbyServer>(`/server/${id}`, data)
}

/**
 * 删除服务器
 */
export function deleteServer(id: number) {
  return request.delete(`/server/${id}`)
}

/**
 * 测试服务器连接
 */
export function testConnection(id: number) {
  return request.post<{
    status: string
    latency: number
    version?: string
    server_name?: string
  }>(`/server/${id}/test`)
}

/**
 * 同步服务器设备
 */
export function syncDevices(id: number) {
  return request.post<{
    count: number
    devices: any[]
  }>(`/server/${id}/sync-devices`)
}

/**
 * 同步服务器媒体库
 */
export function syncLibraries(id: number) {
  return request.post<{
    count: number
    libraries: any[]
  }>(`/server/${id}/sync-libraries`)
}
