<template>
  <div class="dashboard-container">
    <!-- 顶部导航 -->
    <el-header class="dashboard-header">
      <div class="header-left">
        <h1 class="logo">Emby 管理平台</h1>
      </div>
      <div class="header-right">
        <el-dropdown @command="handleUserMenu">
          <span class="user-info">
            <el-avatar :size="32">{{ userInfo.nickname || userInfo.username?.charAt(0) }}</el-avatar>
            <span class="username">{{ userInfo.nickname || userInfo.username }}</span>
            <el-icon class="el-icon--right"><arrow-down /></el-icon>
          </span>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item command="profile">个人资料</el-dropdown-item>
              <el-dropdown-item command="changePassword">修改密码</el-dropdown-item>
              <el-dropdown-item divided command="logout">退出登录</el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </div>
    </el-header>

    <el-container class="main-container">
      <!-- 侧边栏 -->
      <el-aside width="250px" class="sidebar">
        <el-menu
          :default-active="activeMenu"
          class="sidebar-menu"
          @select="handleMenuSelect"
        >
          <el-menu-item index="overview">
            <el-icon><monitor /></el-icon>
            <span>总览</span>
          </el-menu-item>
          <el-menu-item index="servers">
            <el-icon><server /></el-icon>
            <span>Emby 服务器</span>
          </el-menu-item>
          <el-menu-item index="libraries">
            <el-icon><folder /></el-icon>
            <span>媒体库</span>
          </el-menu-item>
          <el-menu-item index="users">
            <el-icon><user /></el-icon>
            <span>用户管理</span>
          </el-menu-item>
          <el-menu-item index="logs">
            <el-icon><document /></el-icon>
            <span>连接日志</span>
          </el-menu-item>
          <el-menu-item index="settings">
            <el-icon><setting /></el-icon>
            <span>系统设置</span>
          </el-menu-item>
        </el-menu>
      </el-aside>

      <!-- 主要内容区 -->
      <el-main class="main-content">
        <!-- 总览页面 -->
        <div v-if="activeMenu === 'overview'" class="overview">
          <h2 class="page-title">系统总览</h2>

          <el-row :gutter="20" class="stats-row">
            <el-col :span="6">
              <el-card class="stat-card">
                <div class="stat-content">
                  <div class="stat-icon servers">
                    <el-icon><server /></el-icon>
                  </div>
                  <div class="stat-info">
                    <div class="stat-number">{{ stats.servers }}</div>
                    <div class="stat-label">Emby 服务器</div>
                  </div>
                </div>
              </el-card>
            </el-col>
            <el-col :span="6">
              <el-card class="stat-card">
                <div class="stat-content">
                  <div class="stat-icon libraries">
                    <el-icon><folder /></el-icon>
                  </div>
                  <div class="stat-info">
                    <div class="stat-number">{{ stats.libraries }}</div>
                    <div class="stat-label">媒体库</div>
                  </div>
                </div>
              </el-card>
            </el-col>
            <el-col :span="6">
              <el-card class="stat-card">
                <div class="stat-content">
                  <div class="stat-icon users">
                    <el-icon><user /></el-icon>
                  </div>
                  <div class="stat-info">
                    <div class="stat-number">{{ stats.users }}</div>
                    <div class="stat-label">用户数量</div>
                  </div>
                </div>
              </el-card>
            </el-col>
            <el-col :span="6">
              <el-card class="stat-card">
                <div class="stat-content">
                  <div class="stat-icon online">
                    <el-icon><connection /></el-icon>
                  </div>
                  <div class="stat-info">
                    <div class="stat-number">{{ stats.onlineServers }}</div>
                    <div class="stat-label">在线服务器</div>
                  </div>
                </div>
              </el-card>
            </el-col>
          </el-row>

          <el-row :gutter="20" class="content-row">
            <el-col :span="12">
              <el-card class="content-card">
                <template #header>
                  <div class="card-header">
                    <span>最近活动</span>
                    <el-button type="text" @click="refreshActivities">刷新</el-button>
                  </div>
                </template>
                <div class="activity-list">
                  <div v-for="activity in activities" :key="activity.id" class="activity-item">
                    <div class="activity-icon" :class="activity.type">
                      <el-icon>
                        <component :is="getActivityIcon(activity.type)" />
                      </el-icon>
                    </div>
                    <div class="activity-content">
                      <div class="activity-title">{{ activity.title }}</div>
                      <div class="activity-time">{{ formatTime(activity.time) }}</div>
                    </div>
                  </div>
                  <div v-if="activities.length === 0" class="empty-state">
                    暂无活动记录
                  </div>
                </div>
              </el-card>
            </el-col>
            <el-col :span="12">
              <el-card class="content-card">
                <template #header>
                  <div class="card-header">
                    <span>服务器状态</span>
                    <el-button type="text" @click="refreshServers">刷新</el-button>
                  </div>
                </template>
                <div class="server-list">
                  <div v-for="server in recentServers" :key="server.id" class="server-item">
                    <div class="server-status" :class="server.status"></div>
                    <div class="server-info">
                      <div class="server-name">{{ server.name }}</div>
                      <div class="server-url">{{ server.url }}</div>
                    </div>
                    <div class="server-actions">
                      <el-button size="small" type="primary" @click="connectToServer(server)">
                        连接
                      </el-button>
                    </div>
                  </div>
                  <div v-if="recentServers.length === 0" class="empty-state">
                    暂无服务器配置
                  </div>
                </div>
              </el-card>
            </el-col>
          </el-row>
        </div>

        <!-- 其他页面内容 -->
        <div v-else class="page-placeholder">
          <el-empty description="功能开发中...">
            <el-button type="primary" @click="activeMenu = 'overview'">返回总览</el-button>
          </el-empty>
        </div>
      </el-main>
    </el-container>

    <!-- 修改密码对话框 -->
    <el-dialog v-model="showChangePassword" title="修改密码" width="400px">
      <el-form
        ref="passwordForm"
        :model="passwordData"
        :rules="passwordRules"
        label-width="80px"
      >
        <el-form-item label="原密码" prop="oldPassword">
          <el-input
            v-model="passwordData.oldPassword"
            type="password"
            show-password
          />
        </el-form-item>
        <el-form-item label="新密码" prop="newPassword">
          <el-input
            v-model="passwordData.newPassword"
            type="password"
            show-password
          />
        </el-form-item>
        <el-form-item label="确认密码" prop="confirmPassword">
          <el-input
            v-model="passwordData.confirmPassword"
            type="password"
            show-password
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showChangePassword = false">取消</el-button>
        <el-button type="primary" @click="handleChangePassword">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import {
  ArrowDown,
  Monitor,
  Server,
  Folder,
  User,
  Document,
  Setting,
  Connection,
  Success,
  Warning,
  Info
} from '@element-plus/icons-vue'

const router = useRouter()
const passwordForm = ref<FormInstance>()

// 用户信息
const userInfo = ref(JSON.parse(localStorage.getItem('user') || '{}'))

// 当前菜单
const activeMenu = ref('overview')

// 统计数据
const stats = reactive({
  servers: 0,
  libraries: 0,
  users: 0,
  onlineServers: 0
})

// 活动记录
const activities = ref([
  {
    id: 1,
    type: 'success',
    title: '用户 admin 登录系统',
    time: new Date(Date.now() - 1000 * 60 * 5)
  },
  {
    id: 2,
    type: 'info',
    title: '服务器 Emby-01 连接成功',
    time: new Date(Date.now() - 1000 * 60 * 30)
  }
])

// 服务器列表
const recentServers = ref([
  {
    id: 1,
    name: 'Emby-01',
    url: 'http://192.168.1.100:8096',
    status: 'online'
  },
  {
    id: 2,
    name: 'Emby-02',
    url: 'http://192.168.1.101:8096',
    status: 'offline'
  }
])

// 修改密码
const showChangePassword = ref(false)
const passwordData = reactive({
  oldPassword: '',
  newPassword: '',
  confirmPassword: ''
})

const passwordRules: FormRules = {
  oldPassword: [{ required: true, message: '请输入原密码', trigger: 'blur' }],
  newPassword: [
    { required: true, message: '请输入新密码', trigger: 'blur' },
    { min: 6, message: '密码长度不能少于6位', trigger: 'blur' }
  ],
  confirmPassword: [
    { required: true, message: '请确认新密码', trigger: 'blur' },
    {
      validator: (rule, value, callback) => {
        if (value !== passwordData.newPassword) {
          callback(new Error('两次输入密码不一致'))
        } else {
          callback()
        }
      },
      trigger: 'blur'
    }
  ]
}

// 菜单选择
const handleMenuSelect = (key: string) => {
  if (key === 'servers') {
    router.push('/servers')
  } else {
    activeMenu.value = key
  }
}

// 用户菜单
const handleUserMenu = async (command: string) => {
  switch (command) {
    case 'profile':
      ElMessage.info('个人资料功能开发中')
      break
    case 'changePassword':
      showChangePassword.value = true
      break
    case 'logout':
      try {
        confirmLogout()
      } catch {
        // 用户取消
      }
      break
  }
}

// 确认退出
const confirmLogout = () => {
  ElMessageBox.confirm('确定要退出登录吗？', '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(() => {
    logout()
  })
}

// 退出登录
const logout = () => {
  localStorage.removeItem('token')
  localStorage.removeItem('user')
  ElMessage.success('已退出登录')
  router.push('/login')
}

// 修改密码
const handleChangePassword = async () => {
  if (!passwordForm.value) return

  try {
    await passwordForm.value.validate()

    // TODO: 调用修改密码API
    const token = localStorage.getItem('token')
    const response = await fetch('/api/user/change-password', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`
      },
      body: JSON.stringify({
        old_password: passwordData.oldPassword,
        new_password: passwordData.newPassword
      })
    })

    const result = await response.json()
    if (result.code === 200) {
      ElMessage.success('密码修改成功')
      showChangePassword.value = false
      passwordForm.value.resetFields()
    } else {
      ElMessage.error(result.message || '密码修改失败')
    }
  } catch (error) {
    console.error('修改密码错误:', error)
    ElMessage.error('密码修改失败，请重试')
  }
}

// 刷新活动
const refreshActivities = () => {
  // TODO: 获取活动记录
  ElMessage.success('活动记录已刷新')
}

// 刷新服务器
const refreshServers = () => {
  // TODO: 获取服务器状态
  ElMessage.success('服务器状态已刷新')
}

// 连接到服务器
const connectToServer = (server: any) => {
  ElMessage.info(`连接到服务器: ${server.name}`)
}

// 获取活动图标
const getActivityIcon = (type: string) => {
  switch (type) {
    case 'success': return Success
    case 'warning': return Warning
    default: return Info
  }
}

// 格式化时间
const formatTime = (time: Date) => {
  const now = new Date()
  const diff = now.getTime() - time.getTime()
  const minutes = Math.floor(diff / (1000 * 60))

  if (minutes < 1) return '刚刚'
  if (minutes < 60) return `${minutes}分钟前`

  const hours = Math.floor(minutes / 60)
  if (hours < 24) return `${hours}小时前`

  const days = Math.floor(hours / 24)
  return `${days}天前`
}

// 加载统计数据
const loadStats = async () => {
  // TODO: 从API获取统计数据
  stats.servers = 2
  stats.libraries = 5
  stats.users = 3
  stats.onlineServers = 1
}

onMounted(() => {
  loadStats()
})
</script>

<style scoped>
.dashboard-container {
  height: 100vh;
  display: flex;
  flex-direction: column;
}

.dashboard-header {
  background: white;
  border-bottom: 1px solid #e4e7ed;
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0 20px;
}

.header-left .logo {
  margin: 0;
  font-size: 20px;
  font-weight: 600;
  color: #409eff;
}

.user-info {
  display: flex;
  align-items: center;
  cursor: pointer;
  gap: 8px;
}

.username {
  font-size: 14px;
  color: #606266;
}

.main-container {
  flex: 1;
  overflow: hidden;
}

.sidebar {
  background: #f5f7fa;
  border-right: 1px solid #e4e7ed;
}

.sidebar-menu {
  border-right: none;
  height: 100%;
}

.main-content {
  padding: 20px;
  background: #f0f2f5;
  overflow-y: auto;
}

.page-title {
  margin: 0 0 20px 0;
  font-size: 24px;
  font-weight: 600;
  color: #2c3e50;
}

.stats-row {
  margin-bottom: 20px;
}

.stat-card {
  height: 100px;
}

.stat-content {
  display: flex;
  align-items: center;
  height: 60px;
  gap: 16px;
}

.stat-icon {
  width: 60px;
  height: 60px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 24px;
  color: white;
}

.stat-icon.servers { background: linear-gradient(135deg, #667eea, #764ba2); }
.stat-icon.libraries { background: linear-gradient(135deg, #f093fb, #f5576c); }
.stat-icon.users { background: linear-gradient(135deg, #4facfe, #00f2fe); }
.stat-icon.online { background: linear-gradient(135deg, #43e97b, #38f9d7); }

.stat-number {
  font-size: 28px;
  font-weight: 600;
  color: #2c3e50;
  margin-bottom: 4px;
}

.stat-label {
  font-size: 14px;
  color: #8492a6;
}

.content-row {
  margin-top: 20px;
}

.content-card {
  height: 400px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.activity-list,
.server-list {
  height: 320px;
  overflow-y: auto;
}

.activity-item,
.server-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 0;
  border-bottom: 1px solid #f0f0f0;
}

.activity-item:last-child,
.server-item:last-child {
  border-bottom: none;
}

.activity-icon {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
}

.activity-icon.success { background: #67c23a; }
.activity-icon.warning { background: #e6a23c; }
.activity-icon.info { background: #909399; }

.activity-content {
  flex: 1;
}

.activity-title {
  font-size: 14px;
  color: #2c3e50;
  margin-bottom: 4px;
}

.activity-time {
  font-size: 12px;
  color: #8492a6;
}

.server-status {
  width: 8px;
  height: 8px;
  border-radius: 50%;
}

.server-status.online { background: #67c23a; }
.server-status.offline { background: #f56c6c; }

.server-info {
  flex: 1;
}

.server-name {
  font-size: 14px;
  color: #2c3e50;
  margin-bottom: 4px;
}

.server-url {
  font-size: 12px;
  color: #8492a6;
}

.server-actions {
  /* 样式根据需要添加 */
}

.empty-state {
  text-align: center;
  color: #8492a6;
  padding: 40px 0;
}

.page-placeholder {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 100%;
}
</style>