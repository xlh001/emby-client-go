import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { request } from '@/services/request'
import type { User, LoginRequest, LoginResponse } from '@/types'

export const userStore = defineStore('user', () => {
  // 状态
  const user = ref<User | null>(null)
  const token = ref<string | null>(null)
  const isLoading = ref(false)

  // 计算属性
  const isLoggedIn = computed(() => !!token.value && !!user.value)
  const isAdmin = computed(() => user.value?.role === 'admin')

  // Actions
  const login = async (credentials: LoginRequest) => {
    try {
      isLoading.value = true
      const response = await request.post<LoginResponse>('/auth/login', credentials)

      token.value = response.token
      user.value = response.user

      // 保存token到localStorage
      localStorage.setItem('token', response.token)
      localStorage.setItem('user', JSON.stringify(response.user))

      return response
    } catch (error) {
      throw error
    } finally {
      isLoading.value = false
    }
  }

  const logout = () => {
    user.value = null
    token.value = null

    // 清除localStorage
    localStorage.removeItem('token')
    localStorage.removeItem('user')
  }

  const refreshToken = async () => {
    try {
      const response = await request.post<{ token: string }>('/auth/refresh')
      token.value = response.token

      // 更新localStorage
      localStorage.setItem('token', response.token)

      return response.token
    } catch (error) {
      // 刷新token失败，执行登出
      logout()
      throw error
    }
  }

  const updateProfile = async (userData: Partial<User>) => {
    try {
      isLoading.value = true
      const response = await request.put<User>(`/users/${user.value?.id}`, userData)
      user.value = response

      // 更新localStorage
      localStorage.setItem('user', JSON.stringify(response))

      return response
    } catch (error) {
      throw error
    } finally {
      isLoading.value = false
    }
  }

  // 初始化：从localStorage恢复用户信息
  const initFromStorage = () => {
    const savedToken = localStorage.getItem('token')
    const savedUser = localStorage.getItem('user')

    if (savedToken && savedUser) {
      try {
        token.value = savedToken
        user.value = JSON.parse(savedUser)
      } catch (error) {
        // localStorage数据损坏，清除
        logout()
      }
    }
  }

  // 在store创建时初始化
  initFromStorage()

  return {
    // 状态
    user: readonly(user),
    token: readonly(token),
    isLoading: readonly(isLoading),

    // 计算属性
    isLoggedIn,
    isAdmin,

    // Actions
    login,
    logout,
    refreshToken,
    updateProfile,
    initFromStorage,
  }
})