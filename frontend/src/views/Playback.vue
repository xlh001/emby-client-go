<template>
  <div class="playback-page">
    <el-card class="header-card">
      <h2>播放控制</h2>
    </el-card>

    <el-row :gutter="20">
      <el-col :span="16">
        <el-card>
          <template #header>
            <div class="section-header">
              <span>活动会话</span>
              <el-select v-model="selectedServer" placeholder="选择服务器" @change="loadSessions" style="width: 200px">
                <el-option v-for="srv in servers" :key="srv.id" :label="srv.name" :value="srv.id" />
              </el-select>
            </div>
          </template>

          <el-skeleton v-if="loading" :rows="3" animated />
          <div v-else-if="sessions.length > 0">
            <PlayerControl
              v-for="session in sessions"
              :key="session.id"
              :session="session"
              :server-i-d="selectedServer"
              :device-i-d="session.device_id"
            />
          </div>
          <el-empty v-else description="暂无活动播放会话" />
        </el-card>
      </el-col>

      <el-col :span="8">
        <el-card>
          <template #header>播放历史</template>
          <el-skeleton v-if="historyLoading" :rows="5" animated />
          <el-timeline v-else-if="history.length > 0">
            <el-timeline-item v-for="record in history" :key="record.id" :timestamp="formatDate(record.played_at)">
              <div class="history-item">
                <div class="history-title">{{ record.media_item?.name }}</div>
                <div class="history-meta">
                  <el-tag size="small">{{ record.device_name }}</el-tag>
                  <span v-if="record.is_completed" class="completed">已完成</span>
                </div>
              </div>
            </el-timeline-item>
          </el-timeline>
          <el-empty v-else description="暂无播放历史" />
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import PlayerControl from '@/components/PlayerControl.vue'
import { getActiveSessions, getPlaybackHistory } from '@/services/playback'
import { getServers } from '@/services/server'

const servers = ref<any[]>([])
const selectedServer = ref<number>()
const sessions = ref<any[]>([])
const history = ref<any[]>([])
const loading = ref(false)
const historyLoading = ref(false)

onMounted(async () => {
  await loadServers()
  await loadHistory()
})

const loadServers = async () => {
  try {
    const res = await getServers()
    servers.value = res.data
    if (servers.value.length > 0) {
      selectedServer.value = servers.value[0].id
      await loadSessions()
    }
  } catch (error: any) {
    ElMessage.error(error.message || '加载服务器列表失败')
  }
}

const loadSessions = async () => {
  if (!selectedServer.value) return
  loading.value = true
  try {
    const res = await getActiveSessions(selectedServer.value)
    sessions.value = res.data
  } catch (error: any) {
    ElMessage.error(error.message || '加载会话失败')
  } finally {
    loading.value = false
  }
}

const loadHistory = async () => {
  historyLoading.value = true
  try {
    const res = await getPlaybackHistory({ limit: 10, offset: 0 })
    history.value = res.data.records
  } catch (error: any) {
    ElMessage.error(error.message || '加载历史失败')
  } finally {
    historyLoading.value = false
  }
}

const formatDate = (date: string) => {
  return new Date(date).toLocaleString('zh-CN')
}
</script>

<style scoped>
.playback-page {
  padding: 20px;
}

.header-card {
  margin-bottom: 20px;
}

.header-card h2 {
  margin: 0;
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.history-item {
  margin-bottom: 10px;
}

.history-title {
  font-weight: 500;
  margin-bottom: 5px;
}

.history-meta {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 12px;
}

.completed {
  color: #67c23a;
}
</style>
