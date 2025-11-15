// 用户相关类型定义

export interface User {
  id: number
  username: string
  email: string
  nickname: string
  role: 'admin' | 'user'
  status: 'active' | 'inactive' | 'locked'
  last_login?: string
  created_at: string
}

export interface LoginForm {
  username: string
  password: string
}

export interface RegisterForm {
  username: string
  email: string
  password: string
  confirmPassword: string
  nickname?: string
}

export interface ChangePasswordForm {
  oldPassword: string
  newPassword: string
  confirmPassword: string
}
