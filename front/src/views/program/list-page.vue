<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import {
  getHotRecommend,
  listEvents,
  listProgramCategories,
  searchEvents,
  type ProgramCategoryInfo,
  type ProgramCityInfo,
  type ProgramEventBrief,
} from '@/api/program'
import {
  buildDateRangeByFilter,
  buildEventCardDescription,
  buildKeywordTitle,
  formatCurrency,
  formatEventCardTime,
  getPageTitle,
  resolveCategoryIdForPage,
  resolvePosterUrl,
  type ProgramPageKind,
  type TimeFilter,
} from '@/utils/program'

const props = withDefaults(defineProps<{
  pageKind?: ProgramPageKind
  showHero?: boolean
}>(), {
  pageKind: 'all',
  showHero: false,
})

const route = useRoute()
const router = useRouter()

const loading = ref(false)
const hotLoading = ref(false)
const events = ref<ProgramEventBrief[]>([])
const hotEvents = ref<ProgramEventBrief[]>([])
const categories = ref<ProgramCategoryInfo[]>([])
const cities = ref<ProgramCityInfo[]>([])

const filters = reactive({
  city: '',
  categoryId: 0,
  time: 'all' as TimeFilter,
})

const pagination = reactive({
  page: 1,
  pageSize: 20,
  total: 0,
})

const keyword = computed(() =>
  typeof route.query.keyword === 'string' ? route.query.keyword.trim() : '',
)

const showHeroSection = computed(() => props.showHero && !keyword.value)
const showCategoryFilter = computed(() => props.pageKind === 'all')
const heroEvents = computed(() => hotEvents.value.slice(0, 3))
const pageTitle = computed(() => buildKeywordTitle(keyword.value, props.pageKind))
const effectiveCategoryId = computed(() => {
  if (props.pageKind === 'all') {
    return filters.categoryId
  }

  return resolveCategoryIdForPage(categories.value, props.pageKind)
})

const resultDescription = computed(() => {
  if (keyword.value) {
    return `共找到 ${pagination.total} 场与“${keyword.value}”相关的活动`
  }
  return `共 ${pagination.total} 场活动`
})

const emptyDescription = computed(() => {
  if (keyword.value) {
    return `没有找到与“${keyword.value}”相关的活动`
  }
  return '暂无符合条件的活动'
})

async function loadCategories() {
  try {
    const response = await listProgramCategories()
    categories.value = response.categories || []
    cities.value = response.cities || []
  } catch {
    categories.value = []
    cities.value = []
  }
}

async function loadHotEvents() {
  if (!showHeroSection.value) {
    hotEvents.value = []
    return
  }

  hotLoading.value = true
  try {
    const response = await getHotRecommend({
      limit: 8,
      city: filters.city || undefined,
    })
    hotEvents.value = response.events || []
  } catch {
    hotEvents.value = []
  } finally {
    hotLoading.value = false
  }
}

function buildListParams() {
  const dateRange = buildDateRangeByFilter(filters.time)

  return {
    page: pagination.page,
    pageSize: pagination.pageSize,
    categoryId: effectiveCategoryId.value || undefined,
    city: filters.city || undefined,
    startDate: dateRange.startDate,
    endDate: dateRange.endDate,
  }
}

async function loadEvents() {
  loading.value = true
  try {
    const params = buildListParams()
    const response = keyword.value
      ? await searchEvents({
          ...params,
          keyword: keyword.value,
        })
      : await listEvents({
          ...params,
          sortBy: props.pageKind === 'all' ? 'hot' : 'time',
        })

    events.value = response.events || []
    pagination.total = Number(response.total || 0)
    pagination.page = response.page || pagination.page
    pagination.pageSize = response.pageSize || pagination.pageSize
  } catch {
    events.value = []
    pagination.total = 0
  } finally {
    loading.value = false
  }
}

async function loadPageData() {
  await Promise.all([loadEvents(), loadHotEvents()])
}

function handleFilterChange() {
  pagination.page = 1
  void loadPageData()
}

function handlePageChange(page: number) {
  pagination.page = page
  void loadEvents()
}

function handlePageSizeChange(pageSize: number) {
  pagination.page = 1
  pagination.pageSize = pageSize
  void loadEvents()
}

function goToDetail(id: string) {
  router.push({
    name: 'ActivityDetail',
    params: { id },
  })
}

onMounted(async () => {
  await loadCategories()
  await loadPageData()
})

watch(
  () => route.query.keyword,
  () => {
    pagination.page = 1
    void loadPageData()
  },
)
</script>

<template>
  <div class="program-page">
    <div v-if="showHeroSection" class="hero-section wrapper" v-loading="hotLoading">
      <el-carousel
        v-if="heroEvents.length"
        height="420px"
        trigger="click"
        indicator-position="outside"
      >
        <el-carousel-item v-for="event in heroEvents" :key="event.id">
          <div class="hero-slide" @click="goToDetail(event.id)">
            <img
              :src="resolvePosterUrl(event.posterUrl, event.title)"
              :alt="event.title"
              class="hero-image"
            />
            <div class="hero-overlay">
              <div class="hero-topline">
                <span class="hero-chip">{{ event.category.name || getPageTitle(pageKind) }}</span>
                <span class="hero-time">{{ formatEventCardTime(event.eventStartTime, event.eventEndTime) }}</span>
              </div>
              <h2 class="hero-title">{{ event.title }}</h2>
              <p class="hero-subtitle">{{ buildEventCardDescription(event) }}</p>
              <div class="hero-price">
                <span>¥{{ formatCurrency(event.minPrice) }}</span>
                <span class="hero-price-suffix">起</span>
              </div>
            </div>
          </div>
        </el-carousel-item>
      </el-carousel>
    </div>

    <div class="filters-section wrapper">
      <el-card shadow="never" class="filter-card">
        <div class="filter-row">
          <span class="filter-label">城市</span>
          <el-radio-group v-model="filters.city" size="small" @change="handleFilterChange">
            <el-radio-button value="">全部</el-radio-button>
            <el-radio-button
              v-for="city in cities"
              :key="city.id"
              :value="city.name"
            >
              {{ city.name }}
            </el-radio-button>
          </el-radio-group>
        </div>

        <div v-if="showCategoryFilter" class="filter-row">
          <span class="filter-label">分类</span>
          <el-radio-group v-model="filters.categoryId" size="small" @change="handleFilterChange">
            <el-radio-button :value="0">全部</el-radio-button>
            <el-radio-button
              v-for="category in categories"
              :key="category.id"
              :value="category.id"
            >
              {{ category.name }}
            </el-radio-button>
          </el-radio-group>
        </div>

        <div class="filter-row last-row">
          <span class="filter-label">时间</span>
          <el-radio-group v-model="filters.time" size="small" @change="handleFilterChange">
            <el-radio-button value="all">全部</el-radio-button>
            <el-radio-button value="today">今天</el-radio-button>
            <el-radio-button value="tomorrow">明天</el-radio-button>
            <el-radio-button value="weekend">本周末</el-radio-button>
            <el-radio-button value="next30">未来30天</el-radio-button>
          </el-radio-group>
        </div>
      </el-card>
    </div>

    <div class="content-section wrapper" v-loading="loading">
      <div class="section-header">
        <div>
          <h2 class="section-title">{{ pageTitle }}</h2>
          <p class="section-subtitle">{{ resultDescription }}</p>
        </div>
      </div>

      <el-empty v-if="!events.length && !loading" :description="emptyDescription" />

      <el-row v-else :gutter="24">
        <el-col
          v-for="event in events"
          :key="event.id"
          :xs="24"
          :sm="12"
          :lg="6"
        >
          <el-card
            shadow="hover"
            class="event-card"
            :body-style="{ padding: '0px' }"
            @click="goToDetail(event.id)"
          >
            <div class="poster-wrapper">
              <img
                :src="resolvePosterUrl(event.posterUrl, event.title)"
                :alt="event.title"
                class="poster-image"
              />
              <div class="category-tag">{{ event.category.name || '活动' }}</div>
            </div>

            <div class="event-info">
              <h3 class="event-title" :title="event.title">{{ event.title }}</h3>
              <p class="event-meta">
                <el-icon><Calendar /></el-icon>
                {{ formatEventCardTime(event.eventStartTime, event.eventEndTime) }}
              </p>
              <p class="event-meta">
                <el-icon><Location /></el-icon>
                {{ buildEventCardDescription(event) }}
              </p>
              <div class="event-bottom">
                <span class="event-price">
                  <span class="currency">¥</span>{{ formatCurrency(event.minPrice) }}
                  <span class="price-suffix">起</span>
                </span>
              </div>
            </div>
          </el-card>
        </el-col>
      </el-row>

      <div v-if="pagination.total > pagination.pageSize" class="pagination-wrapper">
        <el-pagination
          background
          layout="total, sizes, prev, pager, next"
          :current-page="pagination.page"
          :page-size="pagination.pageSize"
          :page-sizes="[8, 20, 40]"
          :total="pagination.total"
          @current-change="handlePageChange"
          @size-change="handlePageSizeChange"
        />
      </div>
    </div>
  </div>
</template>

<style scoped>
.wrapper {
  max-width: 1200px;
  margin: 0 auto;
}

.hero-section {
  padding-top: 24px;
}

.hero-slide {
  position: relative;
  height: 420px;
  border-radius: 20px;
  overflow: hidden;
  cursor: pointer;
}

.hero-image {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.hero-overlay {
  position: absolute;
  inset: 0;
  display: flex;
  flex-direction: column;
  justify-content: flex-end;
  padding: 36px;
  color: #fff;
  background:
    linear-gradient(180deg, rgba(0, 0, 0, 0.08) 0%, rgba(0, 0, 0, 0.72) 100%);
}

.hero-topline {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 14px;
  flex-wrap: wrap;
}

.hero-chip {
  padding: 6px 12px;
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.18);
  backdrop-filter: blur(8px);
  font-size: 13px;
}

.hero-time {
  font-size: 14px;
  opacity: 0.92;
}

.hero-title {
  margin: 0 0 10px;
  font-size: 34px;
  line-height: 1.25;
}

.hero-subtitle {
  margin: 0 0 18px;
  font-size: 15px;
  opacity: 0.92;
}

.hero-price {
  display: inline-flex;
  align-items: baseline;
  gap: 4px;
  font-size: 30px;
  font-weight: 700;
}

.hero-price-suffix {
  font-size: 14px;
  font-weight: 400;
}

.filters-section {
  margin-top: 28px;
  margin-bottom: 28px;
}

.filter-card {
  border-radius: 16px;
  border: 1px solid var(--el-border-color-light);
}

.filter-row {
  display: flex;
  align-items: flex-start;
  gap: 18px;
  padding-bottom: 16px;
  margin-bottom: 16px;
  border-bottom: 1px dashed var(--el-border-color-lighter);
}

.last-row {
  margin-bottom: 0;
  padding-bottom: 0;
  border-bottom: none;
}

.filter-label {
  width: 48px;
  color: var(--el-text-color-secondary);
  line-height: 32px;
  flex-shrink: 0;
}

.content-section {
  padding-bottom: 40px;
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: end;
  margin-bottom: 22px;
}

.section-title {
  margin: 0;
  font-size: 28px;
  color: var(--el-text-color-primary);
}

.section-subtitle {
  margin: 8px 0 0;
  color: var(--el-text-color-secondary);
  font-size: 14px;
}

.event-card {
  margin-bottom: 24px;
  border-radius: 16px;
  overflow: hidden;
  cursor: pointer;
  transition:
    transform 0.2s ease,
    box-shadow 0.2s ease;
}

.event-card:hover {
  transform: translateY(-4px);
  box-shadow: 0 12px 28px rgba(15, 23, 42, 0.12);
}

.poster-wrapper {
  position: relative;
  height: 240px;
  overflow: hidden;
}

.poster-image {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.category-tag {
  position: absolute;
  top: 12px;
  right: 12px;
  padding: 5px 10px;
  border-radius: 999px;
  color: #fff;
  background: rgba(15, 23, 42, 0.68);
  font-size: 12px;
}

.event-info {
  padding: 18px;
}

.event-title {
  margin: 0 0 12px;
  font-size: 16px;
  line-height: 1.5;
  color: var(--el-text-color-primary);
  display: -webkit-box;
  overflow: hidden;
  text-overflow: ellipsis;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  min-height: 48px;
}

.event-meta {
  display: flex;
  align-items: center;
  gap: 6px;
  margin: 0 0 8px;
  color: var(--el-text-color-secondary);
  font-size: 13px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.event-bottom {
  display: flex;
  justify-content: flex-end;
  margin-top: 16px;
}

.event-price {
  color: #ef4444;
  font-size: 24px;
  font-weight: 700;
}

.currency {
  font-size: 14px;
  margin-right: 2px;
}

.price-suffix {
  font-size: 12px;
  font-weight: 400;
  margin-left: 2px;
  color: var(--el-text-color-secondary);
}

.pagination-wrapper {
  display: flex;
  justify-content: center;
  margin-top: 12px;
}

@media (max-width: 768px) {
  .hero-slide {
    height: 320px;
    border-radius: 16px;
  }

  .hero-overlay {
    padding: 24px;
  }

  .hero-title {
    font-size: 24px;
  }

  .filter-row {
    flex-direction: column;
    gap: 12px;
  }

  .filter-label {
    width: auto;
    line-height: 1;
  }
}
</style>
