import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { ElMessage } from 'element-plus'
import * as authApi from '@/services/auth'
import type { UserInfo, LoginRequest, RegisterRequest } from '@/services/auth'
import router from '@/router'

export const useUserStore = defineStore('user', () => {
  // 状态
  const token = ref<string>(localStorage.getItem('token') || '')
  const userInfo = ref<UserInfo | null>(
    localStorage.getItem('user') ? JSON.parse(localStorage.getItem('user')!) : null
  )

  // 计算属性
  const isLoggedIn = computed(() => !!token.value)
  const isAdmin = computed(() => userInfo.value?.role === 'admin')

  /**
   * 登录
   */
  async function login(loginData: LoginRequest) {
    try {
      const res = await authApi.login(loginData)

      // 保存token和用户信息
      token.value = res.data!.token
      userInfo.value = res.data!.user

      localStorage.setItem('token', res.data!.token)
      localStorage.setItem('user', JSON.stringify(res.data!.user))

      ElMessage.success('登录成功')
      return true
    } catch (error) {
      console.error('登录失败:', error)
      return false
    }
  }

  /**
   * 注册
   */
  async function register(registerData: RegisterRequest) {
    try {
      await authApi.register(registerData)
      ElMessage.success('注册成功，请登录')
      return true
    } catch (error) {
      console.error('注册失败:', error)
      return false
    }
  }

  /**
   * 登出
   */
  async function logout() {
    try {
      // 调用登出API
      await authApi.logout()
    } catch (error) {
      console.error('登出API调用失败:', error)
    } finally {
      // 无论API调用是否成功，都清除本地数据
      token.value = ''
      userInfo.value = null
      localStorage.removeItem('token')
      localStorage.removeItem('user')

      ElMessage.success('已退出登录')
      router.push('/login')
    }
  }

  /**
   * 刷新令牌
   */
  async function refreshToken() {
    if (!token.value) return false

    try {
      const res = await authApi.refreshToken({ token: token.value })

      // 更新token
      token.value = res.data!.token
      localStorage.setItem('token', res.data!.token)

      return true
    } catch (error) {
      console.error('刷新令牌失败:', error)
      // 刷新失败，清除登录状态
      await logout()
      return false
    }
  }

  /**
   * 获取用户信息
   */
  async function getUserInfo() {
    try {
      const res = await authApi.getUserProfile()
      userInfo.value = res.data!
      localStorage.setItem('user', JSON.stringify(res.data!))
      return true
    } catch (error) {
      console.error('获取用户信息失败:', error)
      return false
    }
  }

  /**
   * 修改密码
   */
  async function changePassword(oldPassword: string, newPassword: string) {
    try {
      await authApi.changePassword({
        old_password: oldPassword,
        new_password: newPassword
      })
      ElMessage.success('密码修改成功')
      return true
    } catch (error) {
      console.error('修改密码失败:', error)
      return false
    }
  }

  return {
    // 状态
    token,
    userInfo,
    // 计算属性
    isLoggedIn,
    isAdmin,
    // 方法
    login,
    register,
    logout,
    refreshToken,
    getUserInfo,
    changePassword
  }
})
