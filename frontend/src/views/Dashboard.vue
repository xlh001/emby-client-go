<template>
  <div class="dashboard-container">
    <!-- 顶部导航栏 -->
    <el-header class="dashboard-header">
      <div class="header-left">
        <h1>Emby Manager</h1>
      </div>
      <div class="header-right">
        <el-dropdown @command="handleCommand">
          <span class="user-dropdown">
            <el-avatar :src="user?.avatar" />
            <span>{{ user?.username }}</span>
          </span>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item command="profile">个人资料</el-dropdown-item>
              <el-dropdown-item command="settings" v-if="isAdmin">系统设置</el-dropdown-item>
              <el-dropdown-item divided command="logout">退出登录</el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </div>
    </el-header>

    <el-container class="dashboard-content">
      <!-- 侧边栏 -->
      <el-aside width="240px" class="dashboard-aside">
        <el-menu
          :default-active="currentRoute"
          :router="true"
          class="dashboard-menu"
        >
          <el-menu-item index="/dashboard">
            <el-icon><Dashboard /></el-icon>
            <span>仪表盘</span>
          </el-menu-item>

          <el-menu-item index="/servers">
            <el-icon><Monitor /></el-icon>
            <span>服务器管理</span>
          </el-menu-item>

          <el-sub-menu index="media">
            <template #title>
              <el-icon><Folder /></el-icon>
              <span>媒体管理</span>
            </template>
            <el-menu-item index="/media/libraries">媒体库</el-menu-item>
            <el-menu-item index="/media/search">媒体搜索</el-menu-item>
          </el-sub-menu>

          <el-menu-item index="/users" v-if="isAdmin">
            <el-icon><User /></el-icon>
            <span>用户管理</span>
          </el-menu-item>

          <el-menu-item index="/settings" v-if="isAdmin">
            <el-icon><Setting /></el-icon>
            <span>系统设置</span>
          </el-menu-item>
        </el-menu>
      </el-aside>

      <!-- 主内容区 -->
      <el-main class="dashboard-main">
        <router-view />
      </el-main>
    </el-container>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { userStore } from '@/stores/user'

const router = useRouter()
const route = useRoute()
const store = userStore()

const user = computed(() => store.user)
const isAdmin = computed(() => store.isAdmin)
const currentRoute = computed(() => route.path)

const handleCommand = async (command: string) => {
  switch (command) {
    case 'profile':
      router.push('/profile')
      break
    case 'settings':
      router.push('/settings')
      break
    case 'logout':
      try {
        await ElMessageBox.confirm('确定要退出登录吗？', '提示', {
          confirmButtonText: '确定',
          cancelButtonText: '取消',
          type: 'warning',
        })
        store.logout()
        ElMessage.success('已退出登录')
        router.push('/login')
      } catch {
        // 用户取消
      }
      break
  }
}
</script>

<style lang="scss" scoped>
.dashboard-container {
  height: 100vh;
  display: flex;
  flex-direction: column;
}

.dashboard-header {
  background: #fff;
  border-bottom: 1px solid #e4e7ed;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 20px;

  .header-left {
    h1 {
      margin: 0;
      font-size: 20px;
      font-weight: 600;
      color: var(--el-color-primary);
    }
  }

  .header-right {
    .user-dropdown {
      display: flex;
      align-items: center;
      gap: 8px;
      cursor: pointer;

      &:hover {
        color: var(--el-color-primary);
      }
    }
  }
}

.dashboard-content {
  flex: 1;
  overflow: hidden;
}

.dashboard-aside {
  background: #fff;
  border-right: 1px solid #e4e7ed;

  .dashboard-menu {
    border-right: none;
    height: 100%;
  }
}

.dashboard-main {
  background: #f5f5f5;
  padding: 20px;
  overflow-y: auto;
}
</style>