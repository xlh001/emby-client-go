<template>
  <div class="servers-container">
    <div class="page-header">
      <h2 class="page-title">Emby 服务器管理</h2>
      <el-button type="primary" :icon="Plus" @click="showAddDialog">
        添加服务器
      </el-button>
    </div>

    <!-- 服务器列表 -->
    <div v-loading="loading" class="servers-grid">
      <ServerCard
        v-for="server in servers"
        :key="server.id"
        :server="server"
        @edit="handleEdit"
        @delete="handleDelete"
        @test="handleTest"
        @sync="handleSync"
        @connect="handleConnect"
        @manage="handleManage"
      />

      <div v-if="servers.length === 0 && !loading" class="empty-state">
        <el-empty description="暂无服务器">
          <el-button type="primary" @click="showAddDialog">添加第一个服务器</el-button>
        </el-empty>
      </div>
    </div>

    <!-- 添加/编辑服务器对话框 -->
    <el-dialog
      v-model="dialogVisible"
      :title="isEdit ? '编辑服务器' : '添加服务器'"
      width="500px"
      @close="resetForm"
    >
      <el-form
        ref="formRef"
        :model="formData"
        :rules="formRules"
        label-width="100px"
      >
        <el-form-item label="服务器名称" prop="name">
          <el-input
            v-model="formData.name"
            placeholder="例如: 家庭Emby服务器"
            clearable
          />
        </el-form-item>

        <el-form-item label="服务器地址" prop="url">
          <el-input
            v-model="formData.url"
            placeholder="例如: http://192.168.1.100:8096"
            clearable
          >
            <template #prepend>
              <el-select v-model="urlProtocol" style="width: 90px">
                <el-option label="http://" value="http://" />
                <el-option label="https://" value="https://" />
              </el-select>
            </template>
          </el-input>
          <div class="form-tip">请输入完整的服务器地址，包括端口号</div>
        </el-form-item>

        <el-form-item label="API密钥" prop="api_key">
          <el-input
            v-model="formData.api_key"
            type="password"
            show-password
            placeholder="请输入Emby API密钥"
            clearable
          />
          <div class="form-tip">
            在Emby服务器的 设置 > API密钥 中获取
          </div>
        </el-form-item>

        <el-form-item label="描述" prop="description">
          <el-input
            v-model="formData.description"
            type="textarea"
            :rows="3"
            placeholder="可选：添加服务器描述信息"
            maxlength="200"
            show-word-limit
          />
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="handleSubmit">
          {{ isEdit ? '保存' : '添加' }}
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, onUnmounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import ServerCard from '@/components/ServerCard.vue'
import type { EmbyServer } from '@/types/server'
import {
  getServers,
  createServer,
  updateServer,
  deleteServer,
  testConnection,
  syncDevices,
  syncLibraries
} from '@/services/server'
import wsClient, { type WebSocketMessage } from '@/services/websocket'

const loading = ref(false)
const servers = ref<EmbyServer[]>([])

// 对话框相关
const dialogVisible = ref(false)
const isEdit = ref(false)
const submitting = ref(false)
const formRef = ref<FormInstance>()
const urlProtocol = ref('http://')

const formData = reactive({
  id: 0,
  name: '',
  url: '',
  api_key: '',
  description: ''
})

const formRules: FormRules = {
  name: [
    { required: true, message: '请输入服务器名称', trigger: 'blur' },
    { min: 2, max: 50, message: '长度在 2 到 50 个字符', trigger: 'blur' }
  ],
  url: [
    { required: true, message: '请输入服务器地址', trigger: 'blur' },
    {
      pattern: /^(http|https):\/\/.+/,
      message: '请输入有效的URL地址',
      trigger: 'blur'
    }
  ],
  api_key: [
    { required: true, message: '请输入API密钥', trigger: 'blur' },
    { min: 10, message: 'API密钥长度不能少于10位', trigger: 'blur' }
  ]
}

// 加载服务器列表
const loadServers = async () => {
  loading.value = true
  try {
    const res = await getServers()
    servers.value = res.data || []
  } catch (error) {
    console.error('加载服务器列表失败:', error)
    ElMessage.error('加载服务器列表失败')
  } finally {
    loading.value = false
  }
}

// 显示添加对话框
const showAddDialog = () => {
  isEdit.value = false
  dialogVisible.value = true
}

// 处理编辑
const handleEdit = (server: EmbyServer) => {
  isEdit.value = true
  formData.id = server.id
  formData.name = server.name
  formData.url = server.url
  formData.api_key = server.api_key
  formData.description = server.description || ''

  // 提取协议
  if (server.url.startsWith('https://')) {
    urlProtocol.value = 'https://'
    formData.url = server.url.replace('https://', '')
  } else {
    urlProtocol.value = 'http://'
    formData.url = server.url.replace('http://', '')
  }

  dialogVisible.value = true
}

// 处理删除
const handleDelete = async (server: EmbyServer) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除服务器 "${server.name}" 吗？此操作不可恢复。`,
      '确认删除',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )

    await deleteServer(server.id)
    ElMessage.success('服务器已删除')
    loadServers()
  } catch (error: any) {
    if (error !== 'cancel') {
      console.error('删除服务器失败:', error)
      ElMessage.error('删除服务器失败')
    }
  }
}

// 处理测试连接
const handleTest = async (server: EmbyServer) => {
  const loadingMsg = ElMessage.info({
    message: '正在测试连接...',
    duration: 0
  })

  try {
    const res = await testConnection(server.id)
    loadingMsg.close()

    if (res.data) {
      ElMessage.success({
        message: `连接成功！延迟: ${res.data.latency}ms`,
        duration: 3000
      })
      loadServers()
    }
  } catch (error) {
    loadingMsg.close()
    console.error('测试连接失败:', error)
    ElMessage.error('连接测试失败，请检查服务器配置')
  }
}

// 处理同步
const handleSync = async (server: EmbyServer) => {
  try {
    await ElMessageBox.confirm(
      '确定要同步服务器数据吗？这将同步设备和媒体库信息。',
      '确认同步',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'info'
      }
    )

    const loadingMsg = ElMessage.info({
      message: '正在同步数据...',
      duration: 0
    })

    try {
      await Promise.all([
        syncDevices(server.id),
        syncLibraries(server.id)
      ])

      loadingMsg.close()
      ElMessage.success('数据同步成功')
      loadServers()
    } catch (error) {
      loadingMsg.close()
      throw error
    }
  } catch (error: any) {
    if (error !== 'cancel') {
      console.error('同步数据失败:', error)
      ElMessage.error('同步数据失败')
    }
  }
}

// 处理连接
const handleConnect = (server: EmbyServer) => {
  ElMessage.info(`连接到服务器: ${server.name}`)
  // TODO: 实现连接逻辑
}

// 处理管理
const handleManage = (server: EmbyServer) => {
  ElMessage.info(`管理服务器: ${server.name}`)
  // TODO: 跳转到服务器详情页
}

// 提交表单
const handleSubmit = async () => {
  if (!formRef.value) return

  try {
    await formRef.value.validate()

    submitting.value = true

    // 组合完整URL
    const fullUrl = urlProtocol.value + formData.url

    const data = {
      name: formData.name,
      url: fullUrl,
      api_key: formData.api_key,
      description: formData.description || undefined
    }

    if (isEdit.value) {
      await updateServer(formData.id, data)
      ElMessage.success('服务器已更新')
    } else {
      await createServer(data)
      ElMessage.success('服务器已添加')
    }

    dialogVisible.value = false
    loadServers()
  } catch (error: any) {
    if (error !== 'cancel') {
      console.error('保存服务器失败:', error)
      ElMessage.error(isEdit.value ? '更新服务器失败' : '添加服务器失败')
    }
  } finally {
    submitting.value = false
  }
}

// 重置表单
const resetForm = () => {
  formRef.value?.resetFields()
  formData.id = 0
  formData.name = ''
  formData.url = ''
  formData.api_key = ''
  formData.description = ''
  urlProtocol.value = 'http://'
}

// WebSocket消息处理
const handleWebSocketMessage = (message: WebSocketMessage) => {
  if (message.type === 'server-status') {
    // 更新服务器状态
    const serverIndex = servers.value.findIndex(
      s => s.id.toString() === message.server_id
    )
    if (serverIndex !== -1) {
      servers.value[serverIndex].status = message.data.status
      servers.value[serverIndex].last_check = new Date().toISOString()
    }
  }
}

// 初始化WebSocket
const initWebSocket = () => {
  const token = localStorage.getItem('token')
  if (token) {
    wsClient.connect(token)
    wsClient.onMessage(handleWebSocketMessage)
  }
}

onMounted(() => {
  loadServers()
  initWebSocket()
})

onUnmounted(() => {
  wsClient.disconnect()
})
</script>

<style scoped>
.servers-container {
  padding: 20px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
}

.page-title {
  margin: 0;
  font-size: 24px;
  font-weight: 600;
  color: #2c3e50;
}

.servers-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(350px, 1fr));
  gap: 20px;
  min-height: 400px;
}

.empty-state {
  grid-column: 1 / -1;
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 400px;
}

.form-tip {
  font-size: 12px;
  color: #8492a6;
  margin-top: 4px;
  line-height: 1.5;
}
</style>
