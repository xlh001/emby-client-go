<template>
  <el-card class="player-control">
    <template #header>
      <div class="header">
        <span>播放控制</span>
        <el-tag v-if="session" :type="stateType">{{ session.play_state }}</el-tag>
      </div>
    </template>

    <div v-if="session" class="control-content">
      <div class="media-info">
        <h3>{{ session.media_item?.name || '未知媒体' }}</h3>
        <p class="device-info">
          <el-icon><Monitor /></el-icon>
          {{ session.device?.name || '未知设备' }}
        </p>
      </div>

      <div class="progress-bar">
        <el-slider v-model="position" :max="100" @change="handleSeek" />
        <div class="time-info">
          <span>{{ formatTime(currentPosition) }}</span>
          <span>{{ formatTime(totalDuration) }}</span>
        </div>
      </div>

      <div class="control-buttons">
        <el-button-group>
          <el-button :icon="VideoPlay" @click="handlePlay" :disabled="session.play_state === 'Playing'">播放</el-button>
          <el-button :icon="VideoPause" @click="handlePause" :disabled="session.play_state === 'Paused'">暂停</el-button>
          <el-button :icon="SwitchButton" @click="handleStop">停止</el-button>
        </el-button-group>
      </div>
    </div>

    <el-empty v-else description="暂无活动播放会话" />
  </el-card>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { VideoPlay, VideoPause, SwitchButton, Monitor } from '@element-plus/icons-vue'
import { sendPlayCommand } from '@/services/playback'

const props = defineProps<{
  session: any
  serverID: number
  deviceID: number
}>()

const position = ref(0)

const currentPosition = computed(() => props.session?.position_ticks || 0)
const totalDuration = computed(() => props.session?.media_item?.run_time_ticks || 1)

const stateType = computed(() => {
  const state = props.session?.play_state
  return state === 'Playing' ? 'success' : state === 'Paused' ? 'warning' : 'info'
})

watch(() => props.session?.position_ticks, (val) => {
  if (val && totalDuration.value) {
    position.value = (val / totalDuration.value) * 100
  }
})

const handlePlay = async () => {
  try {
    await sendPlayCommand(props.serverID, props.deviceID, {
      command: 'Play',
      session_id: props.session.emby_session_id
    })
    ElMessage.success('播放命令已发送')
  } catch (error: any) {
    ElMessage.error(error.message || '发送命令失败')
  }
}

const handlePause = async () => {
  try {
    await sendPlayCommand(props.serverID, props.deviceID, {
      command: 'Pause',
      session_id: props.session.emby_session_id
    })
    ElMessage.success('暂停命令已发送')
  } catch (error: any) {
    ElMessage.error(error.message || '发送命令失败')
  }
}

const handleStop = async () => {
  try {
    await sendPlayCommand(props.serverID, props.deviceID, {
      command: 'Stop',
      session_id: props.session.emby_session_id
    })
    ElMessage.success('停止命令已发送')
  } catch (error: any) {
    ElMessage.error(error.message || '发送命令失败')
  }
}

const handleSeek = async (val: number) => {
  const seekPosition = Math.floor((val / 100) * totalDuration.value)
  try {
    await sendPlayCommand(props.serverID, props.deviceID, {
      command: 'Seek',
      session_id: props.session.emby_session_id,
      position: seekPosition
    })
  } catch (error: any) {
    ElMessage.error(error.message || '跳转失败')
  }
}

const formatTime = (ticks: number) => {
  const seconds = Math.floor(ticks / 10000000)
  const h = Math.floor(seconds / 3600)
  const m = Math.floor((seconds % 3600) / 60)
  const s = seconds % 60
  return `${h.toString().padStart(2, '0')}:${m.toString().padStart(2, '0')}:${s.toString().padStart(2, '0')}`
}
</script>

<style scoped>
.player-control {
  margin-bottom: 20px;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.control-content {
  padding: 10px 0;
}

.media-info h3 {
  margin: 0 0 10px 0;
  font-size: 18px;
}

.device-info {
  display: flex;
  align-items: center;
  gap: 5px;
  color: #909399;
  font-size: 14px;
  margin: 0;
}

.progress-bar {
  margin: 20px 0;
}

.time-info {
  display: flex;
  justify-content: space-between;
  font-size: 12px;
  color: #909399;
  margin-top: 5px;
}

.control-buttons {
  display: flex;
  justify-content: center;
}
</style>
