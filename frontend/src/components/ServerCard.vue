<template>
  <el-card class="server-card" :class="{ 'is-online': server.status === 'online' }">
    <div class="server-header">
      <div class="server-status-indicator" :class="server.status"></div>
      <div class="server-info">
        <h3 class="server-name">{{ server.name }}</h3>
        <p class="server-url">{{ server.url }}</p>
      </div>
      <el-dropdown @command="handleCommand" trigger="click">
        <el-button :icon="MoreFilled" circle />
        <template #dropdown>
          <el-dropdown-menu>
            <el-dropdown-item command="edit" :icon="Edit">编辑</el-dropdown-item>
            <el-dropdown-item command="test" :icon="Connection">测试连接</el-dropdown-item>
            <el-dropdown-item command="sync" :icon="Refresh">同步数据</el-dropdown-item>
            <el-dropdown-item command="delete" :icon="Delete" divided>删除</el-dropdown-item>
          </el-dropdown-menu>
        </template>
      </el-dropdown>
    </div>

    <div class="server-details">
      <div class="detail-item">
        <span class="detail-label">版本:</span>
        <span class="detail-value">{{ server.version || '未知' }}</span>
      </div>
      <div class="detail-item">
        <span class="detail-label">系统:</span>
        <span class="detail-value">{{ server.os || '未知' }}</span>
      </div>
      <div class="detail-item">
        <span class="detail-label">最后检查:</span>
        <span class="detail-value">{{ formatTime(server.last_check) }}</span>
      </div>
    </div>

    <div v-if="server.description" class="server-description">
      {{ server.description }}
    </div>

    <div class="server-footer">
      <el-tag :type="getStatusType(server.status)" size="small">
        {{ getStatusText(server.status) }}
      </el-tag>
      <div class="server-actions">
        <el-button size="small" @click="handleConnect">
          连接
        </el-button>
        <el-button size="small" type="primary" @click="handleManage">
          管理
        </el-button>
      </div>
    </div>
  </el-card>
</template>

<script setup lang="ts">
import { MoreFilled, Edit, Delete, Connection, Refresh } from '@element-plus/icons-vue'
import type { EmbyServer } from '@/types/server'

interface Props {
  server: EmbyServer
}

interface Emits {
  (e: 'edit', server: EmbyServer): void
  (e: 'delete', server: EmbyServer): void
  (e: 'test', server: EmbyServer): void
  (e: 'sync', server: EmbyServer): void
  (e: 'connect', server: EmbyServer): void
  (e: 'manage', server: EmbyServer): void
}

const props = defineProps<Props>()
const emit = defineEmits<Emits>()

const handleCommand = (command: string) => {
  switch (command) {
    case 'edit':
      emit('edit', props.server)
      break
    case 'delete':
      emit('delete', props.server)
      break
    case 'test':
      emit('test', props.server)
      break
    case 'sync':
      emit('sync', props.server)
      break
  }
}

const handleConnect = () => {
  emit('connect', props.server)
}

const handleManage = () => {
  emit('manage', props.server)
}

const getStatusType = (status: string) => {
  switch (status) {
    case 'online':
      return 'success'
    case 'offline':
      return 'info'
    case 'error':
      return 'danger'
    default:
      return 'info'
  }
}

const getStatusText = (status: string) => {
  switch (status) {
    case 'online':
      return '在线'
    case 'offline':
      return '离线'
    case 'error':
      return '错误'
    default:
      return '未知'
  }
}

const formatTime = (time?: string) => {
  if (!time) return '从未'

  const date = new Date(time)
  const now = new Date()
  const diff = now.getTime() - date.getTime()
  const minutes = Math.floor(diff / (1000 * 60))

  if (minutes < 1) return '刚刚'
  if (minutes < 60) return `${minutes}分钟前`

  const hours = Math.floor(minutes / 60)
  if (hours < 24) return `${hours}小时前`

  const days = Math.floor(hours / 24)
  if (days < 7) return `${days}天前`

  return date.toLocaleDateString()
}
</script>

<style scoped>
.server-card {
  transition: all 0.3s;
  cursor: pointer;
}

.server-card:hover {
  transform: translateY(-4px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
}

.server-card.is-online {
  border-left: 4px solid #67c23a;
}

.server-header {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  margin-bottom: 16px;
}

.server-status-indicator {
  width: 12px;
  height: 12px;
  border-radius: 50%;
  margin-top: 4px;
  flex-shrink: 0;
}

.server-status-indicator.online {
  background: #67c23a;
  box-shadow: 0 0 8px rgba(103, 194, 58, 0.6);
}

.server-status-indicator.offline {
  background: #909399;
}

.server-status-indicator.error {
  background: #f56c6c;
  box-shadow: 0 0 8px rgba(245, 108, 108, 0.6);
}

.server-info {
  flex: 1;
  min-width: 0;
}

.server-name {
  margin: 0 0 4px 0;
  font-size: 18px;
  font-weight: 600;
  color: #2c3e50;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.server-url {
  margin: 0;
  font-size: 13px;
  color: #8492a6;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.server-details {
  display: flex;
  flex-direction: column;
  gap: 8px;
  margin-bottom: 12px;
  padding: 12px;
  background: #f5f7fa;
  border-radius: 4px;
}

.detail-item {
  display: flex;
  justify-content: space-between;
  font-size: 13px;
}

.detail-label {
  color: #8492a6;
}

.detail-value {
  color: #2c3e50;
  font-weight: 500;
}

.server-description {
  margin-bottom: 12px;
  padding: 8px;
  font-size: 13px;
  color: #606266;
  background: #f9fafc;
  border-radius: 4px;
  line-height: 1.5;
}

.server-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding-top: 12px;
  border-top: 1px solid #ebeef5;
}

.server-actions {
  display: flex;
  gap: 8px;
}
</style>
