<template>
  <div class="search-results">
    <!-- 搜索结果头部信息 -->
    <div v-if="result" class="results-header">
      <div class="results-info">
        <h3 class="results-title">
          搜索结果
          <span v-if="searchQuery" class="search-query">
            "{{ searchQuery }}"
          </span>
        </h3>
        <p class="results-count">
          找到 {{ result.total }} 个结果
          <span v-if="result.total > 0 && result.limit">
            (当前显示第 {{ result.offset + 1 }}-{{ Math.min(result.offset + result.limit, result.total) }} 个)
          </span>
        </p>
      </div>

      <!-- 聚合信息统计 -->
      <div v-if="hasAggregations" class="aggregations">
        <div class="aggregation-section">
          <h4>类型分布</h4>
          <div class="type-stats">
            <span
              v-for="(count, type) in result.aggregations.types"
              :key="type"
              class="type-tag"
              :class="`type-${type.toLowerCase()}`"
            >
              {{ getTypeLabel(type) }}: {{ count }}
            </span>
          </div>
        </div>

        <div v-if="result.aggregations.servers" class="aggregation-section">
          <h4>服务器分布</h4>
          <div class="server-stats">
            <span
              v-for="(count, server) in result.aggregations.servers"
              :key="server"
              class="server-tag"
            >
              {{ server }}: {{ count }}
            </span>
          </div>
        </div>

        <div v-if="result.aggregations.years" class="aggregation-section">
          <h4>年份分布</h4>
          <div class="year-stats">
            <span
              v-for="(count, year) in result.aggregations.years"
              :key="year"
              class="year-tag"
            >
              {{ year }}: {{ count }}
            </span>
          </div>
        </div>
      </div>
    </div>

    <!-- 搜索结果列表 -->
    <div v-if="result && result.items.length > 0" class="results-list">
      <div
        v-for="item in result.items"
        :key="item.id"
        class="result-item"
        @click="handleItemClick(item)"
      >
        <div class="item-poster">
          <img
            v-if="getItemPoster(item)"
            :src="getItemPoster(item)"
            :alt="item.name"
            class="poster-image"
            @error="handleImageError"
          />
          <div v-else class="poster-placeholder">
            <el-icon><video-camera /></el-icon>
            <span>{{ item.type }}</span>
          </div>
        </div>

        <div class="item-info">
          <div class="item-header">
            <h4 class="item-title">{{ item.name }}</h4>
            <div class="item-meta">
              <el-tag :type="getTypeTagType(item.type)" size="small">
                {{ getTypeLabel(item.type) }}
              </el-tag>
              <span v-if="item.year" class="item-year">{{ item.year }}</span>
              <span v-if="getItemDuration(item)" class="item-duration">
                {{ getItemDuration(item) }}
              </span>
            </div>
          </div>

          <div v-if="item.series_name" class="item-series">
            <el-icon><collection /></el-icon>
            <span>系列: {{ item.series_name }}</span>
          </div>

          <p v-if="item.overview" class="item-description">
            {{ truncateText(item.overview, 120) }}
          </p>

          <div class="item-footer">
            <div class="item-library">
              <el-icon><folder /></el-icon>
              <span>{{ item.media_library?.name || '未知媒体库' }}</span>
              <span v-if="item.media_library?.emby_server?.name">
                ({{ item.media_library.emby_server.name }})
              </span>
            </div>
            <div class="item-dates">
              <span v-if="item.premiere_date" class="premiere-date">
                上映: {{ formatDate(item.premiere_date) }}
              </span>
            </div>
          </div>

          <!-- 评分信息 -->
          <div v-if="hasRatings(item)" class="item-ratings">
            <div v-if="item.community_rating" class="rating-item">
              <el-icon><star-filled /></el-icon>
              <span>{{ item.community_rating.toFixed(1) }}</span>
              <small>(社区)</small>
            </div>
            <div v-if="item.critic_rating" class="rating-item">
              <el-icon><star /></el-icon>
              <span>{{ item.critic_rating.toFixed(1) }}</span>
              <small>(专业)</small>
            </div>
          </div>
        </div>

        <!-- 操作按钮 -->
        <div class="item-actions">
          <el-button
            type="primary"
            size="small"
            @click.stop="handlePlay(item)"
          >
            <el-icon><video-play /></el-icon>
            播放
          </el-button>
          <el-button
            size="small"
            @click.stop="handleDetails(item)"
          >
            <el-icon><info-filled /></el-icon>
            详情
          </el-button>
        </div>
      </div>
    </div>

    <!-- 无结果提示 -->
    <div v-else-if="searchQuery && (!result || result.items.length === 0)" class="no-results">
      <el-empty
        description="未找到相关内容"
        :image-size="120"
      >
        <template #image>
          <el-icon size="60" color="#dcdfe6"><search /></el-icon>
        </template>
        <template #description>
          <p>没有找到与 "{{ searchQuery }}" 相关的内容</p>
          <p class="search-tips">
            建议：
          </p>
          <ul class="tips-list">
            <li>尝试使用不同的关键词</li>
            <li>检查拼写是否正确</li>
            <li>使用更通用的词汇</li>
            <li>调整筛选条件</li>
          </ul>
        </template>
        <el-button type="primary" @click="$emit('clear-search')">
          清空搜索
        </el-button>
      </el-empty>
    </div>

    <!-- 加载状态 -->
    <div v-if="loading" class="loading-container">
      <el-skeleton
        v-for="n in 5"
        :key="n"
        animated
        :rows="3"
        class="skeleton-item"
      />
    </div>

    <!-- 分页组件 -->
    <div
      v-if="result && result.total > result.limit"
      class="pagination-container"
    >
      <el-pagination
        v-model:current-page="currentPage"
        v-model:page-size="pageSize"
        :page-sizes="[10, 20, 50, 100]"
        :total="result.total"
        layout="total, sizes, prev, pager, next, jumper"
        @size-change="handleSizeChange"
        @current-change="handleCurrentChange"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import {
  VideoCamera,
  Collection,
  Folder,
  StarFilled,
  Star,
  VideoPlay,
  InfoFilled,
  Search
} from '@element-plus/icons-vue'
import type { SearchResult, MediaItem } from '@/services/search'

interface Props {
  result?: SearchResult | null
  loading?: boolean
  searchQuery?: string
}

interface Emits {
  (e: 'item-click', item: MediaItem): void
  (e: 'play', item: MediaItem): void
  (e: 'details', item: MediaItem): void
  (e: 'page-change', page: number, pageSize: number): void
  (e: 'clear-search'): void
}

const props = withDefaults(defineProps<Props>(), {
  loading: false,
  searchQuery: ''
})

const emit = defineEmits<Emits>()

// 分页相关
const currentPage = ref(1)
const pageSize = ref(20)

// 计算属性
const hasAggregations = computed(() => {
  return props.result?.aggregations &&
         Object.keys(props.result.aggregations).length > 0 &&
         Object.values(props.result.aggregations).some(val =>
           val && typeof val === 'object' && Object.keys(val).length > 0
         )
})

// 获取类型标签样式
const getTypeTagType = (type: string) => {
  const typeMap: Record<string, string> = {
    'Movie': 'success',
    'Episode': 'primary',
    'Audio': 'warning',
    'Other': 'info'
  }
  return typeMap[type] || 'info'
}

// 获取类型显示文本
const getTypeLabel = (type: string) => {
  const typeMap: Record<string, string> = {
    'Movie': '电影',
    'Episode': '剧集',
    'Audio': '音乐',
    'Other': '其他'
  }
  return typeMap[type] || type
}

// 获取项目海报
const getItemPoster = (item: MediaItem) => {
  // 这里可以根据实际Emby API规则获取海报URL
  // 暂时返回null，使用占位符
  return null
}

// 获取项目时长
const getItemDuration = (item: MediaItem) => {
  if (!item.run_time_ticks) return null

  // Emby的ticks转换为毫秒，再转换为时分秒
  const ticks = item.run_time_ticks
  const milliseconds = Math.floor(ticks / 10000)
  const hours = Math.floor(milliseconds / 3600000)
  const minutes = Math.floor((milliseconds % 3600000) / 60000)

  if (hours > 0) {
    return `${hours}小时${minutes}分钟`
  } else {
    return `${minutes}分钟`
  }
}

// 是否有评分信息
const hasRatings = (item: MediaItem) => {
  return item.community_rating || item.critic_rating
}

// 截断文本
const truncateText = (text: string, maxLength: number) => {
  if (text.length <= maxLength) return text
  return text.substring(0, maxLength) + '...'
}

// 格式化日期
const formatDate = (dateString: string) => {
  try {
    const date = new Date(dateString)
    return date.toLocaleDateString('zh-CN')
  } catch {
    return dateString
  }
}

// 处理图片加载错误
const handleImageError = (event: Event) => {
  const img = event.target as HTMLImageElement
  img.style.display = 'none'
}

// 事件处理函数
const handleItemClick = (item: MediaItem) => {
  emit('item-click', item)
}

const handlePlay = (item: MediaItem) => {
  emit('play', item)
}

const handleDetails = (item: MediaItem) => {
  emit('details', item)
}

// 分页处理
const handleSizeChange = (size: number) => {
  pageSize.value = size
  currentPage.value = 1
  emit('page-change', currentPage.value, pageSize.value)
}

const handleCurrentChange = (page: number) => {
  currentPage.value = page
  emit('page-change', currentPage.value, pageSize.value)
}
</script>

<style scoped>
.search-results {
  width: 100%;
  max-width: 1200px;
  margin: 0 auto;
  padding: 20px;
}

.results-header {
  margin-bottom: 24px;
}

.results-info {
  margin-bottom: 16px;
}

.results-title {
  margin: 0 0 8px 0;
  font-size: 24px;
  font-weight: 600;
  color: #303133;
}

.search-query {
  color: #409eff;
  font-weight: 500;
}

.results-count {
  margin: 0;
  color: #606266;
  font-size: 14px;
}

.aggregations {
  display: flex;
  flex-wrap: wrap;
  gap: 24px;
  padding: 16px;
  background: #f8f9fa;
  border-radius: 8px;
  margin-bottom: 16px;
}

.aggregation-section h4 {
  margin: 0 0 12px 0;
  font-size: 16px;
  font-weight: 500;
  color: #303133;
}

.type-stats,
.server-stats,
.year-stats {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.type-tag,
.server-tag,
.year-tag {
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 500;
  color: white;
}

.type-movie { background: #67c23a; }
.type-episode { background: #409eff; }
.type-audio { background: #e6a23c; }
.type-other { background: #909399; }

.server-tag,
.year-tag {
  background: #909399;
}

.results-list {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(350px, 1fr));
  gap: 20px;
  margin-bottom: 32px;
}

.result-item {
  display: flex;
  background: white;
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  overflow: hidden;
  transition: all 0.3s ease;
  cursor: pointer;
}

.result-item:hover {
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.15);
  transform: translateY(-2px);
}

.item-poster {
  width: 120px;
  flex-shrink: 0;
  position: relative;
  background: #f5f7fa;
}

.poster-image {
  width: 100%;
  height: 180px;
  object-fit: cover;
}

.poster-placeholder {
  width: 100%;
  height: 180px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  color: #909399;
  gap: 8px;
}

.poster-placeholder .el-icon {
  font-size: 24px;
}

.item-info {
  flex: 1;
  padding: 16px;
  display: flex;
  flex-direction: column;
}

.item-header {
  margin-bottom: 8px;
}

.item-title {
  margin: 0 0 6px 0;
  font-size: 16px;
  font-weight: 600;
  color: #303133;
  line-height: 1.4;
}

.item-meta {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
  flex-wrap: wrap;
}

.item-year,
.item-duration {
  color: #909399;
  font-size: 12px;
}

.item-series {
  display: flex;
  align-items: center;
  gap: 4px;
  color: #606266;
  font-size: 12px;
  margin-bottom: 8px;
}

.item-description {
  margin: 0 0 12px 0;
  color: #606266;
  font-size: 14px;
  line-height: 1.5;
  flex: 1;
}

.item-footer {
  margin-top: auto;
  padding-top: 12px;
  border-top: 1px solid #f0f0f0;
}

.item-library,
.item-dates {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 12px;
  color: #909399;
  margin-bottom: 4px;
}

.item-ratings {
  display: flex;
  gap: 12px;
  margin-top: 8px;
}

.rating-item {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 14px;
  font-weight: 500;
}

.rating-item small {
  color: #909399;
  font-size: 12px;
}

.item-actions {
  display: flex;
  gap: 8px;
  padding: 16px;
  background: #fafbfc;
  border-top: 1px solid #f0f0f0;
}

.no-results {
  text-align: center;
  padding: 60px 20px;
}

.search-tips {
  color: #606266;
  margin: 16px 0 8px 0;
  font-weight: 500;
}

.tips-list {
  text-align: left;
  color: #909399;
  font-size: 14px;
  margin: 0;
  padding-left: 20px;
}

.tips-list li {
  margin-bottom: 4px;
}

.loading-container {
  margin-bottom: 24px;
}

.skeleton-item {
  margin-bottom: 16px;
}

.pagination-container {
  display: flex;
  justify-content: center;
  margin-top: 32px;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .search-results {
    padding: 16px;
  }

  .aggregations {
    flex-direction: column;
    gap: 16px;
  }

  .results-list {
    grid-template-columns: 1fr;
    gap: 16px;
  }

  .result-item {
    flex-direction: column;
  }

  .item-poster {
    width: 100%;
    height: 200px;
  }

  .poster-image,
  .poster-placeholder {
    height: 100%;
  }

  .item-actions {
    justify-content: center;
  }
}

@media (max-width: 480px) {
  .results-title {
    font-size: 20px;
  }

  .item-title {
    font-size: 14px;
  }

  .item-actions {
    flex-direction: column;
    gap: 8px;
  }

  .item-actions .el-button {
    width: 100%;
  }
}
</style>