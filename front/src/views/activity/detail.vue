<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'

import {
  getEventDetail,
  type ProgramEventDetail,
  type ProgramTicketTierInfo,
} from '@/api/program'
import {
  buildNoticeList,
  canPurchaseEvent,
  deriveTicketTierState,
  formatCurrency,
  formatDateTimeLabel,
  formatEventDetailTime,
  getEventStatusLabel,
  resolvePosterUrl,
} from '@/utils/program'

const route = useRoute()
const router = useRouter()

const loading = ref(false)
const eventDetail = ref<ProgramEventDetail | null>(null)
const selectedSessionId = ref('')
const selectedTicketId = ref('')
const ticketCount = ref(1)

const eventId = computed(() => (typeof route.params.id === 'string' ? route.params.id : ''))

const posterUrl = computed(() =>
  resolvePosterUrl(eventDetail.value?.posterUrl, eventDetail.value?.title),
)

const sessionOptions = computed(() => {
  if (!eventDetail.value) {
    return []
  }

  return [
    {
      id: eventDetail.value.id,
      label: formatEventDetailTime(
        eventDetail.value.eventStartTime,
        eventDetail.value.eventEndTime,
      ),
      disabled: !canPurchaseEvent(eventDetail.value),
      stateText: getEventStatusLabel(eventDetail.value.status),
    },
  ]
})

const ticketOptions = computed(() => eventDetail.value?.ticketTiers || [])

const selectedTicket = computed(() =>
  ticketOptions.value.find((item) => item.id === selectedTicketId.value) || null,
)

const maxTicketCount = computed(() => {
  const purchaseLimit = eventDetail.value?.purchaseLimit || 1
  const remainStock = selectedTicket.value?.remainStock || purchaseLimit
  return Math.max(1, Math.min(purchaseLimit, remainStock))
})

const totalPrice = computed(() =>
  selectedTicket.value ? selectedTicket.value.price * ticketCount.value : 0,
)

const purchaseNotices = computed(() =>
  eventDetail.value ? buildNoticeList(eventDetail.value) : [],
)

const isPurchaseDisabled = computed(() => {
  if (!eventDetail.value || !selectedTicket.value || !canPurchaseEvent(eventDetail.value)) {
    return true
  }

  return deriveTicketTierState(selectedTicket.value, eventDetail.value).disabled
})

function getTicketState(tier: ProgramTicketTierInfo) {
  return deriveTicketTierState(tier, eventDetail.value)
}

function selectDefaultTicket() {
  selectedSessionId.value = eventDetail.value?.id || ''
  const firstAvailable = ticketOptions.value.find((tier) => !getTicketState(tier).disabled)
  selectedTicketId.value = firstAvailable?.id || ticketOptions.value[0]?.id || ''
  ticketCount.value = 1
}

async function loadEventDetail() {
  if (!eventId.value) {
    eventDetail.value = null
    return
  }

  loading.value = true
  try {
    const response = await getEventDetail(eventId.value)
    eventDetail.value = response.event
    selectDefaultTicket()
  } catch {
    eventDetail.value = null
    selectedSessionId.value = ''
    selectedTicketId.value = ''
  } finally {
    loading.value = false
  }
}

function buyTicket() {
  if (!eventDetail.value || !selectedTicket.value) {
    ElMessage.warning('请先选择票档')
    return
  }

  if (isPurchaseDisabled.value) {
    ElMessage.warning('当前活动暂不可购买')
    return
  }

  router.push({
    path: '/order/create',
    query: {
      activity: eventDetail.value.id,
      session: selectedSessionId.value,
      ticket: selectedTicket.value.id,
      count: String(ticketCount.value),
    },
  })
}

watch(
  eventId,
  () => {
    void loadEventDetail()
  },
  { immediate: true },
)

watch(selectedTicket, () => {
  if (ticketCount.value > maxTicketCount.value) {
    ticketCount.value = maxTicketCount.value
  }
})
</script>

<template>
  <div class="detail-page wrapper" v-loading="loading">
    <el-empty v-if="!loading && !eventDetail" description="活动不存在或已下线" />

    <template v-else-if="eventDetail">
      <el-breadcrumb separator="/" class="breadcrumb">
        <el-breadcrumb-item :to="{ path: '/home' }">首页</el-breadcrumb-item>
        <el-breadcrumb-item>{{ eventDetail.category.name || '活动详情' }}</el-breadcrumb-item>
        <el-breadcrumb-item>{{ eventDetail.title }}</el-breadcrumb-item>
      </el-breadcrumb>

      <div class="purchase-card">
        <img :src="posterUrl" :alt="eventDetail.title" class="poster" />

        <div class="purchase-content">
          <div class="title-row">
            <h1 class="title">{{ eventDetail.title }}</h1>
            <el-tag effect="dark" type="success">{{ getEventStatusLabel(eventDetail.status) }}</el-tag>
          </div>

          <div class="info-grid">
            <div class="info-item">
              <span class="label">时间</span>
              <span class="value">{{ formatEventDetailTime(eventDetail.eventStartTime, eventDetail.eventEndTime) }}</span>
            </div>
            <div class="info-item">
              <span class="label">场馆</span>
              <span class="value">{{ eventDetail.city }} | {{ eventDetail.venue.name }}</span>
            </div>
            <div class="info-item">
              <span class="label">开售</span>
              <span class="value">{{ formatDateTimeLabel(eventDetail.saleStartTime) }}</span>
            </div>
            <div class="info-item">
              <span class="label">票种</span>
              <span class="value">{{ eventDetail.ticketType === 2 ? '纸质票' : '电子票' }}</span>
            </div>
          </div>

          <el-divider />

          <div class="selector-group">
            <span class="selector-label">场次</span>
            <div class="selector-options">
              <el-button
                v-for="session in sessionOptions"
                :key="session.id"
                :disabled="session.disabled"
                :class="{ selected: selectedSessionId === session.id }"
                plain
                @click="selectedSessionId = session.id"
              >
                {{ session.label }}
              </el-button>
            </div>
          </div>

          <div class="selector-group">
            <span class="selector-label">票档</span>
            <div class="selector-options ticket-grid">
              <el-button
                v-for="ticket in ticketOptions"
                :key="ticket.id"
                :disabled="getTicketState(ticket).disabled"
                :class="{ selected: selectedTicketId === ticket.id }"
                plain
                @click="selectedTicketId = ticket.id"
              >
                <div class="ticket-button">
                  <span class="ticket-name">{{ ticket.name }}</span>
                  <span class="ticket-price">¥{{ formatCurrency(ticket.price) }}</span>
                </div>
              </el-button>
            </div>
          </div>

          <div class="selector-group">
            <span class="selector-label">数量</span>
            <div class="selector-options quantity-box">
              <el-input-number
                v-model="ticketCount"
                :min="1"
                :max="maxTicketCount"
                :disabled="isPurchaseDisabled"
              />
              <span class="limit-text">每单限购 {{ eventDetail.purchaseLimit || 1 }} 张</span>
            </div>
          </div>

          <div class="action-bar">
            <div class="price-summary">
              <span>合计</span>
              <strong>¥{{ formatCurrency(totalPrice) }}</strong>
            </div>
            <div class="action-buttons">
              <el-button size="large" plain>收藏</el-button>
              <el-button
                type="danger"
                size="large"
                class="buy-button"
                :disabled="isPurchaseDisabled"
                @click="buyTicket"
              >
                立即购票
              </el-button>
            </div>
          </div>
        </div>
      </div>

      <div class="content-grid">
        <el-card shadow="never" class="content-card">
          <template #header>
            <div class="card-header">活动详情</div>
          </template>
          <p class="content-text">{{ eventDetail.description || '暂无活动描述' }}</p>
        </el-card>

        <el-card shadow="never" class="content-card">
          <template #header>
            <div class="card-header">购票须知</div>
          </template>
          <ul class="notice-list">
            <li v-for="notice in purchaseNotices" :key="notice">{{ notice }}</li>
          </ul>
        </el-card>

        <el-card shadow="never" class="content-card">
          <template #header>
            <div class="card-header">场馆信息</div>
          </template>
          <div class="venue-block">
            <p><strong>{{ eventDetail.venue.name }}</strong></p>
            <p>{{ eventDetail.venue.city }} {{ eventDetail.venue.address }}</p>
            <p>容纳人数：{{ eventDetail.venue.capacity || 0 }}</p>
            <p>{{ eventDetail.venue.description || '暂无场馆描述' }}</p>
          </div>
        </el-card>
      </div>
    </template>
  </div>
</template>

<style scoped>
.wrapper {
  max-width: 1200px;
  margin: 0 auto;
  padding: 24px 0 40px;
}

.breadcrumb {
  margin-bottom: 18px;
}

.purchase-card {
  display: grid;
  grid-template-columns: 320px 1fr;
  gap: 28px;
  padding: 28px;
  border-radius: 20px;
  background: var(--el-bg-color);
  box-shadow: 0 12px 28px rgba(15, 23, 42, 0.08);
}

.poster {
  width: 100%;
  height: 430px;
  object-fit: cover;
  border-radius: 16px;
}

.purchase-content {
  min-width: 0;
}

.title-row {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 18px;
  flex-wrap: wrap;
}

.title {
  margin: 0;
  font-size: 30px;
  color: var(--el-text-color-primary);
}

.info-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px 20px;
}

.info-item {
  display: flex;
  gap: 12px;
  font-size: 14px;
}

.label {
  width: 44px;
  color: var(--el-text-color-secondary);
  flex-shrink: 0;
}

.value {
  color: var(--el-text-color-primary);
}

.selector-group {
  display: flex;
  gap: 16px;
  margin-bottom: 20px;
}

.selector-label {
  width: 44px;
  padding-top: 8px;
  color: var(--el-text-color-secondary);
  flex-shrink: 0;
}

.selector-options {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  flex: 1;
}

.selector-options .el-button.selected {
  border-color: #ef4444;
  color: #ef4444;
  background: #fef2f2;
}

.ticket-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.ticket-button {
  display: flex;
  align-items: flex-start;
  gap: 4px;

}
.el-button{
  margin-left: 0 !important;
}

.ticket-name {
  font-weight: 600;
}

.ticket-price {
  color: #ef4444;
}

.ticket-stock {
  font-size: 12px;
  color: var(--el-text-color-secondary);
}

.quantity-box {
  align-items: center;
}

.limit-text {
  color: var(--el-text-color-secondary);
  font-size: 13px;
}

.action-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 20px;
  margin-top: 28px;
  padding: 18px 20px;
  border-radius: 16px;
  background: var(--el-fill-color-light);
}

.price-summary {
  display: flex;
  align-items: baseline;
  gap: 10px;
  font-size: 16px;
}

.price-summary strong {
  font-size: 30px;
  color: #ef4444;
}

.action-buttons {
  display: flex;
  gap: 12px;
}

.buy-button {
  min-width: 160px;
}

.content-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 20px;
  margin-top: 24px;
}

.content-card {
  min-height: 240px;
  border-radius: 18px;
}

.card-header {
  font-weight: 600;
  color: var(--el-text-color-primary);
}

.content-text {
  margin: 0;
  line-height: 1.8;
  color: var(--el-text-color-regular);
}

.notice-list {
  margin: 0;
  padding-left: 18px;
  line-height: 1.9;
  color: var(--el-text-color-regular);
}

.venue-block {
  line-height: 1.9;
  color: var(--el-text-color-regular);
}

@media (max-width: 900px) {
  .purchase-card {
    grid-template-columns: 1fr;
  }

  .poster {
    height: 360px;
  }

  .content-grid {
    grid-template-columns: 1fr;
  }

  .ticket-grid {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 640px) {
  .wrapper {
    padding: 16px 12px 32px;
  }

  .purchase-card {
    padding: 18px;
  }

  .info-grid {
    grid-template-columns: 1fr;
  }

  .selector-group {
    flex-direction: column;
    gap: 10px;
  }

  .selector-label {
    width: auto;
    padding-top: 0;
  }

  .action-bar {
    flex-direction: column;
    align-items: stretch;
  }

  .action-buttons {
    width: 100%;
  }

  .action-buttons .el-button {
    flex: 1;
  }
}
</style>
