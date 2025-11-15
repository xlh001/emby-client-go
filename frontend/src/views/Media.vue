<template>
  <div class="media-page">
    <el-card class="header-card">
      <div class="header-content">
        <h2>媒体浏览</h2>
        <el-select v-model="selectedLibrary" placeholder="选择媒体库" @change="handleLibraryChange" style="width: 300px">
          <el-option v-for="lib in libraries" :key="lib.id" :label="`${lib.name} (${lib.emby_server?.name})`" :value="lib.id" />
        </el-select>
      </div>
    </el-card>

    <el-card v-if="selectedLibrary" class="content-card">
      <div class="filter-bar">
        <el-radio-group v-model="mediaType" @change="handleTypeChange">
          <el-radio-button label="">全部</el-radio-button>
          <el-radio-button label="Movie">电影</el-radio-button>
          <el-radio-button label="Episode">剧集</el-radio-button>
          <el-radio-button label="Audio">音乐</el-radio-button>
        </el-radio-group>
        <span class="total-count">共 {{ total }} 项</span>
      </div>

      <MediaGrid :items="items" :loading="loading" @item-click="handleItemClick" />

      <el-pagination
        v-if="total > 0"
        class="pagination"
        :current-page="currentPage"
        :page-size="pageSize"
        :total="total"
        layout="prev, pager, next, jumper"
        @current-change="handlePageChange"
      />
    </el-card>

    <el-empty v-else description="请选择媒体库" />

    <el-dialog v-model="detailVisible" title="媒体详情" width="600px">
      <div v-if="currentItem" class="detail-content">
        <el-descriptions :column="1" border>
          <el-descriptions-item label="名称">{{ currentItem.name }}</el-descriptions-item>
          <el-descriptions-item label="类型">{{ currentItem.type }}</el-descriptions-item>
          <el-descriptions-item label="系列" v-if="currentItem.series_name">{{ currentItem.series_name }}</el-descriptions-item>
          <el-descriptions-item label="年份" v-if="currentItem.year">{{ currentItem.year }}</el-descriptions-item>
          <el-descriptions-item label="分辨率" v-if="currentItem.resolution">{{ currentItem.resolution }}</el-descriptions-item>
          <el-descriptions-item label="编码">{{ currentItem.video_codec }} / {{ currentItem.audio_codec }}</el-descriptions-item>
          <el-descriptions-item label="容器">{{ currentItem.container }}</el-descriptions-item>
          <el-descriptions-item label="大小">{{ formatSize(currentItem.size) }}</el-descriptions-item>
        </el-descriptions>
      </div>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import MediaGrid from '@/components/MediaGrid.vue'
import type { MediaLibrary, MediaItem } from '@/types'
import { getMediaLibraries, getMediaItems, getMediaItem } from '@/services/media'

const libraries = ref<MediaLibrary[]>([])
const selectedLibrary = ref<number>()
const mediaType = ref('')
const items = ref<MediaItem[]>([])
const total = ref(0)
const currentPage = ref(1)
const pageSize = ref(20)
const loading = ref(false)
const detailVisible = ref(false)
const currentItem = ref<MediaItem>()

onMounted(async () => {
  await loadLibraries()
})

const loadLibraries = async () => {
  try {
    const res = await getMediaLibraries()
    libraries.value = res.data
  } catch (error: any) {
    ElMessage.error(error.message || '加载媒体库失败')
  }
}

const loadItems = async () => {
  if (!selectedLibrary.value) return

  loading.value = true
  try {
    const res = await getMediaItems({
      library_id: selectedLibrary.value,
      type: mediaType.value,
      limit: pageSize.value,
      offset: (currentPage.value - 1) * pageSize.value
    })
    items.value = res.data.items
    total.value = res.data.total
  } catch (error: any) {
    ElMessage.error(error.message || '加载媒体项目失败')
  } finally {
    loading.value = false
  }
}

const handleLibraryChange = () => {
  currentPage.value = 1
  loadItems()
}

const handleTypeChange = () => {
  currentPage.value = 1
  loadItems()
}

const handlePageChange = (page: number) => {
  currentPage.value = page
  loadItems()
}

const handleItemClick = async (item: MediaItem) => {
  try {
    const res = await getMediaItem(item.id)
    currentItem.value = res.data
    detailVisible.value = true
  } catch (error: any) {
    ElMessage.error(error.message || '加载详情失败')
  }
}

const formatSize = (bytes: number) => {
  if (!bytes) return '-'
  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  let size = bytes
  let unitIndex = 0
  while (size >= 1024 && unitIndex < units.length - 1) {
    size /= 1024
    unitIndex++
  }
  return `${size.toFixed(2)} ${units[unitIndex]}`
}
</script>

<style scoped>
.media-page {
  padding: 20px;
}

.header-card {
  margin-bottom: 20px;
}

.header-content {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header-content h2 {
  margin: 0;
}

.content-card {
  min-height: 500px;
}

.filter-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.total-count {
  color: #909399;
  font-size: 14px;
}

.pagination {
  margin-top: 20px;
  display: flex;
  justify-content: center;
}

.detail-content {
  padding: 10px 0;
}
</style>
