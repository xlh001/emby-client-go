<template>
  <div class="media-grid">
    <el-skeleton v-if="loading" :rows="6" animated />
    <div v-else-if="items.length > 0" class="grid-container">
      <div v-for="item in items" :key="item.id" class="media-card" @click="$emit('item-click', item)">
        <div class="media-poster">
          <el-icon class="placeholder-icon"><VideoPlay /></el-icon>
        </div>
        <div class="media-info">
          <div class="media-name" :title="item.name">{{ item.name }}</div>
          <div class="media-meta">
            <el-tag size="small" type="info">{{ item.type }}</el-tag>
            <span v-if="item.year" class="year">{{ item.year }}</span>
          </div>
        </div>
      </div>
    </div>
    <el-empty v-else description="暂无媒体项目" />
  </div>
</template>

<script setup lang="ts">
import { VideoPlay } from '@element-plus/icons-vue'
import type { MediaItem } from '@/types'

defineProps<{
  items: MediaItem[]
  loading: boolean
}>()

defineEmits<{
  'item-click': [item: MediaItem]
}>()
</script>

<style scoped>
.media-grid {
  min-height: 400px;
}

.grid-container {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(180px, 1fr));
  gap: 20px;
}

.media-card {
  cursor: pointer;
  transition: transform 0.2s;
  border-radius: 8px;
  overflow: hidden;
  background: #f5f7fa;
}

.media-card:hover {
  transform: translateY(-4px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
}

.media-poster {
  width: 100%;
  height: 240px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  display: flex;
  align-items: center;
  justify-content: center;
}

.placeholder-icon {
  font-size: 64px;
  color: rgba(255, 255, 255, 0.8);
}

.media-info {
  padding: 12px;
}

.media-name {
  font-size: 14px;
  font-weight: 500;
  margin-bottom: 8px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.media-meta {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 12px;
  color: #909399;
}

.year {
  font-size: 12px;
}
</style>
