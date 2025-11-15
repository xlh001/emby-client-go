import request, { type ApiResponse } from './request'

// 登录请求参数
export interface LoginRequest {
  username: string
  password: string
}

// 注册请求参数
export interface RegisterRequest {
  username: string
  email: string
  password: string
  nickname?: string
}

// 修改密码请求参数
export interface ChangePasswordRequest {
  old_password: string
  new_password: string
}

// 刷新令牌请求参数
export interface RefreshTokenRequest {
  token: string
}

// 用户信息
export interface UserInfo {
  id: number
  username: string
  email: string
  nickname: string
  role: string
  status: string
  last_login?: string
  created_at: string
}

// 登录响应
export interface LoginResponse {
  token: string
  user: UserInfo
}

// 刷新令牌响应
export interface RefreshTokenResponse {
  token: string
}

/**
 * 用户登录
 */
export function login(data: LoginRequest): Promise<ApiResponse<LoginResponse>> {
  return request.post('/auth/login', data)
}

/**
 * 用户注册
 */
export function register(data: RegisterRequest): Promise<ApiResponse<UserInfo>> {
  return request.post('/auth/register', data)
}

/**
 * 用户登出
 */
export function logout(): Promise<ApiResponse> {
  return request.post('/auth/logout')
}

/**
 * 刷新令牌
 */
export function refreshToken(data: RefreshTokenRequest): Promise<ApiResponse<RefreshTokenResponse>> {
  return request.post('/auth/refresh', data)
}

/**
 * 获取用户信息
 */
export function getUserProfile(): Promise<ApiResponse<UserInfo>> {
  return request.get('/user/profile')
}

/**
 * 修改密码
 */
export function changePassword(data: ChangePasswordRequest): Promise<ApiResponse> {
  return request.post('/user/change-password', data)
}

/**
 * 获取用户列表（管理员）
 */
export interface GetUsersParams {
  page?: number
  page_size?: number
  search?: string
}

export interface PageResponse<T> {
  list: T[]
  total: number
  page: number
  page_size: number
}

export function getUsers(params: GetUsersParams): Promise<ApiResponse<PageResponse<UserInfo>>> {
  return request.get('/user/list', { params })
}
