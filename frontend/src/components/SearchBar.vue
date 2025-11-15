<template>
  <div class="search-bar">
    <div class="search-input-container">
      <el-input
        v-model="searchQuery"
        :placeholder="placeholder"
        :prefix-icon="Search"
        clearable
        @input="handleInput"
        @clear="handleClear"
        @keyup.enter="handleSearch"
        class="search-input"
        size="large"
      />
      <div v-if="showSuggestions" class="suggestions-dropdown">
        <div v-if="loading" class="suggestion-loading">
          <el-icon class="is-loading"><loading /></el-icon>
          <span>加载中...</span>
        </div>
        <ul v-else-if="suggestions.length > 0" class="suggestions-list">
          <li
            v-for="(suggestion, index) in suggestions"
            :key="index"
            class="suggestion-item"
            @click="selectSuggestion(suggestion)"
          >
            <el-icon><search /></el-icon>
            <span>{{ suggestion }}</span>
          </li>
        </ul>
        <div v-else-if="searchQuery.length >= 2" class="no-suggestions">
          <span>暂无建议</span>
        </div>
      </div>
    </div>

    <div class="search-filters" v-if="showFilters">
      <el-select
        v-model="filters.types"
        placeholder="媒体类型"
        multiple
        clearable
        size="small"
        style="width: 120px"
      >
        <el-option label="电影" value="Movie" />
        <el-option label="剧集" value="Episode" />
        <el-option label="音乐" value="Audio" />
        <el-option label="其他" value="Other" />
      </el-select>

      <el-select
        v-model="filters.sort_by"
        placeholder="排序方式"
        size="small"
        style="width: 120px"
      >
        <el-option label="相关性" value="relevance" />
        <el-option label="名称" value="name" />
        <el-option label="年份" value="year" />
        <el-option label="创建时间" value="created_at" />
        <el-option label="更新时间" value="updated_at" />
      </el-select>

      <el-select
        v-model="filters.sort_order"
        placeholder="排序方向"
        size="small"
        style="width: 100px"
      >
        <el-option label="降序" value="desc" />
        <el-option label="升序" value="asc" />
      </el-select>
    </div>

    <div class="search-actions">
      <el-button @click="toggleFilters" size="small">
        <el-icon><filter /></el-icon>
        {{ showFilters ? '收起筛选' : '展开筛选' }}
      </el-button>
      <el-button type="primary" @click="handleSearch" size="small">
        <el-icon><search /></el-icon>
        搜索
      </el-button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, nextTick } from 'vue'
import { Search, Loading, Filter } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { getSearchSuggestions } from '@/services/search'

interface Props {
  placeholder?: string
  showFilterToggle?: boolean
  autoFocus?: boolean
}

interface Emits {
  (e: 'search', params: any): void
  (e: 'clear'): void
}

const props = withDefaults(defineProps<Props>(), {
  placeholder: '搜索媒体内容...',
  showFilterToggle: true,
  autoFocus: false
})

const emit = defineEmits<Emits>()

// 响应式数据
const searchQuery = ref('')
const suggestions = ref<string[]>([])
const showSuggestions = ref(false)
const loading = ref(false)
const showFilters = ref(false)
const searchTimer = ref<NodeJS.Timeout>()

// 筛选条件
const filters = ref({
  types: [] as string[],
  sort_by: 'relevance',
  sort_order: 'desc'
})

// 计算属性
const hasSearchContent = computed(() => searchQuery.value.trim().length > 0)

// 监听搜索框焦点
const searchInputRef = ref<HTMLInputElement>()

// 处理输入事件
const handleInput = async (value: string) => {
  searchQuery.value = value

  // 防抖处理建议获取
  if (searchTimer.value) {
    clearTimeout(searchTimer.value)
  }

  if (value.trim().length >= 2) {
    searchTimer.value = setTimeout(async () => {
      await fetchSuggestions(value.trim())
    }, 300)
    showSuggestions.value = true
  } else {
    showSuggestions.value = false
    suggestions.value = []
  }
}

// 处理清除事件
const handleClear = () => {
  searchQuery.value = ''
  showSuggestions.value = false
  suggestions.value = []
  emit('clear')
}

// 处理搜索事件
const handleSearch = () => {
  const query = searchQuery.value.trim()
  if (query) {
    showSuggestions.value = false
    const searchParams = {
      query,
      types: filters.value.types.length > 0 ? filters.value.types : undefined,
      sort_by: filters.value.sort_by || 'relevance',
      sort_order: filters.value.sort_order || 'desc',
      limit: 20,
      offset: 0
    }
    emit('search', searchParams)
  }
}

// 获取搜索建议
const fetchSuggestions = async (query: string) => {
  try {
    loading.value = true
    const response = await getSearchSuggestions(query, 5)
    suggestions.value = response.data.data?.suggestions || []
  } catch (error) {
    console.error('获取搜索建议失败:', error)
    suggestions.value = []
    ElMessage.error('获取搜索建议失败')
  } finally {
    loading.value = false
  }
}

// 选择建议项
const selectSuggestion = (suggestion: string) => {
  searchQuery.value = suggestion
  showSuggestions.value = false
  suggestions.value = []
  handleSearch()
}

// 切换筛选器显示状态
const toggleFilters = () => {
  showFilters.value = !showFilters.value
}

// 点击外部关闭建议下拉框
const handleClickOutside = (event: MouseEvent) => {
  if (searchInputRef.value && !searchInputRef.value.contains(event.target as Node)) {
    showSuggestions.value = false
  }
}

// 监听搜索查询变化，自动执行搜索
watch(searchQuery, (newValue) => {
  if (newValue.trim() && !showSuggestions.value) {
    // 可以在这里添加自动搜索逻辑
  }
})

// 组件挂载后自动聚焦
nextTick(() => {
  if (props.autoFocus && searchInputRef.value) {
    searchInputRef.value.focus()
  }
})

// 生命周期
onMounted(() => {
  document.addEventListener('click', handleClickOutside)
})

onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside)
  if (searchTimer.value) {
    clearTimeout(searchTimer.value)
  }
})
</script>

<style scoped>
.search-bar {
  width: 100%;
  max-width: 800px;
  margin: 0 auto;
}

.search-input-container {
  position: relative;
  margin-bottom: 12px;
}

.search-input {
  width: 100%;
}

.search-input :deep(.el-input__wrapper) {
  border-radius: 24px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  transition: all 0.3s ease;
}

.search-input:hover :deep(.el-input__wrapper) {
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
}

.suggestions-dropdown {
  position: absolute;
  top: 100%;
  left: 0;
  right: 0;
  background: white;
  border: 1px solid #e4e7ed;
  border-radius: 8px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
  z-index: 1000;
  max-height: 300px;
  overflow-y: auto;
}

.suggestion-loading {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 16px;
  color: #909399;
  gap: 8px;
}

.suggestions-list {
  list-style: none;
  margin: 0;
  padding: 8px 0;
}

.suggestion-item {
  display: flex;
  align-items: center;
  padding: 8px 16px;
  cursor: pointer;
  gap: 8px;
  color: #303133;
  transition: background-color 0.2s ease;
}

.suggestion-item:hover {
  background-color: #f5f7fa;
}

.suggestion-item .el-icon {
  color: #909399;
  font-size: 14px;
}

.no-suggestions {
  padding: 16px;
  text-align: center;
  color: #909399;
  font-size: 14px;
}

.search-filters {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 12px;
  padding: 12px;
  background: #f8f9fa;
  border-radius: 8px;
  flex-wrap: wrap;
}

.search-actions {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .search-filters {
    flex-direction: column;
    align-items: stretch;
    gap: 8px;
  }

  .search-filters .el-select {
    width: 100% !important;
  }

  .search-actions {
    flex-direction: column;
    align-items: stretch;
  }

  .search-actions .el-button {
    width: 100%;
  }
}
</style>