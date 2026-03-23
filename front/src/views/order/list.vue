<script setup lang="ts">
import { computed, onBeforeUnmount, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'

import {
  applyRefund,
  cancelOrder,
  getOrderDetail,
  listOrders,
  payOrder,
  type OrderDetailPayload,
  type OrderSummaryPayload,
} from '@/api/order'
import { getPaymentStatus } from '@/api/payment'
import { formatCurrency, formatEventDetailTime, resolvePosterUrl } from '@/utils/program'

type OrderTab = 'all' | 'pending' | 'paid' | 'refunded' | 'cancelled' | 'completed'

const route = useRoute()
const router = useRouter()

const activeTab = ref<OrderTab>('all')
const loading = ref(false)
const actionLoading = ref('')
const orders = ref<OrderSummaryPayload[]>([])
const detailVisible = ref(false)
const detailLoading = ref(false)
const detail = ref<OrderDetailPayload | null>(null)
const paymentPollSerial = ref(0)

const tabStatusMap: Record<OrderTab, number | undefined> = {
  all: undefined,
  pending: 1,
  cancelled: 2,
  paid: 3,
  refunded: 4,
  completed: 5,
}

const highlightOrderNo = computed(() => {
  const raw = route.query.highlight
  if (Array.isArray(raw)) {
    return raw[0] || ''
  }
  return typeof raw === 'string' ? raw : ''
})

const highlightPaymentNo = computed(() => {
  const raw = route.query.paymentNo
  if (Array.isArray(raw)) {
    return raw[0] || ''
  }
  return typeof raw === 'string' ? raw : ''
})

const paymentResult = computed(() => {
  const raw = route.query.paymentResult
  if (Array.isArray(raw)) {
    return raw[0] || ''
  }
  return typeof raw === 'string' ? raw : ''
})

function formatOrderTime(order: OrderSummaryPayload) {
  return formatEventDetailTime(order.eventStartTime, order.eventEndTime)
}

function getStatusTag(order: OrderSummaryPayload) {
  switch (order.status) {
    case 1:
      return { type: 'danger', text: '待支付' }
    case 2:
      return { type: 'info', text: '已取消' }
    case 3:
      return { type: 'success', text: '已支付' }
    case 4:
      return { type: 'warning', text: '已退款' }
    case 5:
      return { type: 'info', text: '已完成' }
    default:
      return { type: 'info', text: order.statusText || '未知状态' }
  }
}

function isHighlighted(order: OrderSummaryPayload) {
  return !!highlightOrderNo.value && order.orderNo === highlightOrderNo.value
}

async function loadOrders() {
  loading.value = true
  try {
    const resp = await listOrders({
      status: tabStatusMap[activeTab.value],
      page: 1,
      pageSize: 20,
    })
    orders.value = resp.orders || []
  } finally {
    loading.value = false
  }
}

async function openOrderDetail(orderNo: string) {
  detailVisible.value = true
  detailLoading.value = true
  try {
    const resp = await getOrderDetail({ orderNo })
    detail.value = resp.order
  } finally {
    detailLoading.value = false
  }
}

async function handlePay(order: OrderSummaryPayload) {
  actionLoading.value = order.orderNo
  try {
    const resp = await payOrder({
      orderNo: order.orderNo,
      payMethod: 1,
      channel: 'stripe',
    })
    if (resp.paidAt || resp.orderStatus === 3 || resp.payment?.status === 1) {
      ElMessage.success('鏀粯鎴愬姛')
      await loadOrders()
      await openOrderDetail(order.orderNo)
      return
    }

    if (!resp.checkoutUrl) {
      ElMessage.error('鏈幏鍙栧埌鏀粯閾炬帴')
      return
    }

    window.location.assign(resp.checkoutUrl)
    return
    /*
    ElMessage.success('支付成功')
    await loadOrders()
    await openOrderDetail(order.orderNo)
    */
  } finally {
    actionLoading.value = ''
  }
}

async function handleCancel(order: OrderSummaryPayload) {
  try {
    await ElMessageBox.confirm('确认取消这笔订单吗？', '取消订单', {
      type: 'warning',
    })
  } catch {
    return
  }

  actionLoading.value = order.orderNo
  try {
    await cancelOrder({ orderNo: order.orderNo })
    ElMessage.success('订单已取消')
    await loadOrders()
  } finally {
    actionLoading.value = ''
  }
}

async function handleRefund(order: OrderSummaryPayload) {
  try {
    await ElMessageBox.confirm('确认申请退款吗？退款后将释放库存。', '申请退款', {
      type: 'warning',
    })
  } catch {
    return
  }

  actionLoading.value = order.orderNo
  try {
    await applyRefund({
      orderNo: order.orderNo,
      reason: 'user apply',
    })
    ElMessage.success('退款申请已处理')
    await loadOrders()
    await openOrderDetail(order.orderNo)
  } finally {
    actionLoading.value = ''
  }
}

const buyerRows = computed(() => detail.value?.tickets || [])

function sleep(ms: number) {
  return new Promise((resolve) => window.setTimeout(resolve, ms))
}

async function pollPaymentStatus(orderNo: string, paymentNo: string) {
  const currentSerial = ++paymentPollSerial.value

  for (let attempt = 0; attempt < 20; attempt += 1) {
    if (paymentPollSerial.value !== currentSerial) {
      return
    }

    const resp = await getPaymentStatus({
      orderNo: orderNo || undefined,
      paymentNo: paymentNo || undefined,
      channel: 'stripe',
    })

    if (resp.paid) {
      ElMessage.success('鏀粯鎴愬姛')
      await loadOrders()
      if (orderNo) {
        await openOrderDetail(orderNo)
      }
      return
    }

    await sleep(1500)
  }

  ElMessage.warning('鏀粯缁撴灉纭涓紝璇风◢鍚庡埛鏂扮湅鐪?')
}

watch(
  () => [activeTab.value, highlightOrderNo.value],
  () => {
    void loadOrders()
  },
  { immediate: true },
)

watch(
  () => [paymentResult.value, highlightOrderNo.value, highlightPaymentNo.value],
  ([result, orderNo, paymentNo]) => {
    const safeOrderNo = orderNo || ''
    const safePaymentNo = paymentNo || ''

    if (result === 'success' && (safeOrderNo || safePaymentNo)) {
      void pollPaymentStatus(safeOrderNo, safePaymentNo)
      return
    }

    if (result === 'cancel') {
      ElMessage.info('宸插彇娑堟湰娆℃敮浠?')
    }
  },
  { immediate: true },
)

onBeforeUnmount(() => {
  paymentPollSerial.value += 1
})
</script>

<template>
  <div class="order-list-container wrapper">
    <div class="sidebar">
      <el-menu default-active="orders" class="side-menu" :router="false">
        <el-menu-item index="orders">
          <el-icon><Tickets /></el-icon>
          <span>我的订单</span>
        </el-menu-item>
        <el-menu-item index="profile" @click="router.push('/user/profile')">
          <el-icon><User /></el-icon>
          <span>个人中心</span>
        </el-menu-item>
        <el-menu-item index="attendees" @click="router.push('/user/profile?tab=attendees')">
          <el-icon><Avatar /></el-icon>
          <span>常用观演人</span>
        </el-menu-item>
        <el-menu-item index="addresses" @click="router.push('/user/profile?tab=addresses')">
          <el-icon><Location /></el-icon>
          <span>收货地址</span>
        </el-menu-item>
      </el-menu>
    </div>

    <div class="main-content">
      <el-card shadow="never" class="list-card" v-loading="loading">
        <el-tabs v-model="activeTab" class="order-tabs">
          <el-tab-pane label="全部订单" name="all" />
          <el-tab-pane label="待支付" name="pending" />
          <el-tab-pane label="已支付" name="paid" />
          <el-tab-pane label="已退款" name="refunded" />
          <el-tab-pane label="已取消" name="cancelled" />
          <el-tab-pane label="已完成" name="completed" />
        </el-tabs>

        <div v-if="orders.length === 0" class="empty-status">
          <el-empty description="暂无订单数据" />
        </div>

        <div v-else class="order-items">
          <div
            v-for="order in orders"
            :key="order.orderNo"
            class="order-item"
            :class="{ highlighted: isHighlighted(order) }"
          >
            <div class="order-header">
              <span class="time">{{ order.createdAt }}</span>
              <span class="order-id">订单号：{{ order.orderNo }}</span>
              <div class="flex-spacer"></div>
              <el-tag :type="getStatusTag(order).type" class="status-tag">
                {{ getStatusTag(order).text }}
              </el-tag>
            </div>

            <div class="order-body">
              <img :src="resolvePosterUrl(order.posterUrl, order.eventTitle)" class="cover" :alt="order.eventTitle" />
              <div class="info">
                <div class="title">{{ order.eventTitle }}</div>
                <div class="desc"><el-icon><Calendar /></el-icon> {{ formatOrderTime(order) }}</div>
                <div class="desc"><el-icon><Location /></el-icon> {{ order.venueName }}</div>
                <div class="desc">状态：{{ order.statusText || getStatusTag(order).text }}</div>
              </div>
              <div class="ticket-info">
                <div>{{ order.ticketTierName }}</div>
                <div>x {{ order.quantity }}</div>
              </div>
              <div class="price-info">
                <div class="total-price">¥{{ formatCurrency(order.totalAmount) }}</div>
              </div>
              <div class="actions">
                <el-button
                  size="small"
                  round
                  @click="openOrderDetail(order.orderNo)"
                  :loading="detailLoading && detail?.orderNo === order.orderNo"
                >
                  查看详情
                </el-button>

                <template v-if="order.status === 1">
                  <el-button
                    type="danger"
                    size="small"
                    round
                    :loading="actionLoading === order.orderNo"
                    @click="handlePay(order)"
                  >
                    立即支付
                  </el-button>
                  <el-button
                    size="small"
                    round
                    :loading="actionLoading === order.orderNo"
                    @click="handleCancel(order)"
                  >
                    取消订单
                  </el-button>
                </template>

                <template v-else-if="order.status === 3">
                  <el-button
                    size="small"
                    round
                    :loading="actionLoading === order.orderNo"
                    @click="handleRefund(order)"
                  >
                    申请退款
                  </el-button>
                </template>
              </div>
            </div>
          </div>
        </div>
      </el-card>
    </div>

    <el-drawer v-model="detailVisible" title="订单详情" size="560px" destroy-on-close>
      <div v-loading="detailLoading" class="detail-panel">
        <el-empty v-if="!detailLoading && !detail" description="未获取到订单详情" />

        <template v-else-if="detail">
          <div class="detail-cover-row">
            <img :src="resolvePosterUrl(detail.posterUrl, detail.eventTitle)" class="detail-cover" :alt="detail.eventTitle" />
            <div class="detail-main">
              <div class="detail-title">{{ detail.eventTitle }}</div>
              <div class="detail-line">订单号：{{ detail.orderNo }}</div>
              <div class="detail-line">状态：{{ detail.statusText }}</div>
              <div class="detail-line">场次：{{ formatEventDetailTime(detail.eventStartTime, detail.eventEndTime) }}</div>
              <div class="detail-line">场馆：{{ detail.venueName }}</div>
            </div>
          </div>

          <el-descriptions :column="1" border class="detail-block">
            <el-descriptions-item label="票档">{{ detail.ticketTierName }}</el-descriptions-item>
            <el-descriptions-item label="数量">{{ detail.quantity }}</el-descriptions-item>
            <el-descriptions-item label="订单金额">¥{{ formatCurrency(detail.totalAmount) }}</el-descriptions-item>
            <el-descriptions-item label="下单时间">{{ detail.createdAt || '-' }}</el-descriptions-item>
            <el-descriptions-item label="支付截止">{{ detail.payDeadline || '-' }}</el-descriptions-item>
            <el-descriptions-item label="支付时间">{{ detail.paidAt || '-' }}</el-descriptions-item>
          </el-descriptions>

          <el-descriptions :column="1" border class="detail-block">
            <el-descriptions-item label="支付单号">{{ detail.payment.paymentNo || '-' }}</el-descriptions-item>
            <el-descriptions-item label="支付状态">{{ detail.payment.status || 0 }}</el-descriptions-item>
            <el-descriptions-item label="支付流水">{{ detail.payment.tradeNo || '-' }}</el-descriptions-item>
            <el-descriptions-item label="配送方式">{{ detail.delivery.deliveryMethod || '-' }}</el-descriptions-item>
            <el-descriptions-item label="配送状态">{{ detail.delivery.deliveryStatus || '-' }}</el-descriptions-item>
            <el-descriptions-item label="收货地址">
              {{ [detail.delivery.province, detail.delivery.city, detail.delivery.district, detail.delivery.detail].filter(Boolean).join(' ') || '-' }}
            </el-descriptions-item>
          </el-descriptions>

          <div class="detail-block">
            <div class="detail-subtitle">观演人 / 票码</div>
            <el-table :data="buyerRows" border size="small">
              <el-table-column prop="buyer.name" label="观演人" min-width="100" />
              <el-table-column prop="buyer.phone" label="手机号" min-width="120" />
              <el-table-column prop="ticketCode" label="票码" min-width="180" />
              <el-table-column prop="status" label="票状态" min-width="100" />
            </el-table>
          </div>
        </template>
      </div>
    </el-drawer>
  </div>
</template>

<style scoped>
.wrapper {
  max-width: 1200px;
  margin: 0 auto;
  padding: 30px 0;
  display: flex;
  gap: 20px;
}

.sidebar {
  width: 200px;
  background: white;
  border-radius: 8px;
  box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.05);
  overflow: hidden;
  height: max-content;
}

.side-menu {
  border-right: none;
}

.main-content {
  flex: 1;
}

.list-card {
  border-radius: 8px;
  min-height: 500px;
}

.order-tabs {
  margin-bottom: 20px;
}

.order-items {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.order-item {
  border: 1px solid #ebeef5;
  border-radius: 8px;
  overflow: hidden;
  transition: box-shadow 0.3s, border-color 0.3s;
}

.order-item:hover {
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.08);
}

.order-item.highlighted {
  border-color: #f56c6c;
  box-shadow: 0 6px 18px rgba(245, 108, 108, 0.15);
}

.order-header {
  background: var(--el-fill-color-light);
  padding: 12px 20px;
  display: flex;
  align-items: center;
  border-bottom: 1px solid var(--el-border-color-lighter);
  font-size: 13px;
  color: var(--el-text-color-regular);
}

.order-header .time {
  margin-right: 20px;
}

.flex-spacer {
  flex: 1;
}

.order-body {
  padding: 20px;
  display: flex;
  background: var(--el-bg-color);
}

.cover {
  width: 75px;
  height: 100px;
  object-fit: cover;
  border-radius: 4px;
  margin-right: 20px;
}

.info {
  flex: 2;
  display: flex;
  flex-direction: column;
  justify-content: center;
}

.info .title {
  font-size: 15px;
  font-weight: bold;
  color: var(--el-text-color-primary);
  margin-bottom: 12px;
}

.info .desc {
  font-size: 13px;
  color: var(--el-text-color-secondary);
  margin-bottom: 6px;
  display: flex;
  align-items: center;
  gap: 6px;
}

.ticket-info {
  flex: 1;
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  font-size: 14px;
  color: var(--el-text-color-regular);
  border-left: 1px solid var(--el-border-color-lighter);
  border-right: 1px solid var(--el-border-color-lighter);
  padding: 0 20px;
}

.price-info {
  width: 140px;
  display: flex;
  justify-content: center;
  align-items: center;
  border-right: 1px solid var(--el-border-color-lighter);
}

.total-price {
  font-size: 18px;
  font-weight: bold;
  color: var(--el-text-color-primary);
}

.actions {
  width: 140px;
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  gap: 10px;
  padding-left: 20px;
}

.actions .el-button {
  width: 92px;
  margin-left: 0 !important;
}

.detail-panel {
  min-height: 240px;
}

.detail-cover-row {
  display: flex;
  gap: 16px;
  margin-bottom: 20px;
}

.detail-cover {
  width: 96px;
  height: 128px;
  object-fit: cover;
  border-radius: 6px;
}

.detail-main {
  flex: 1;
}

.detail-title {
  font-size: 18px;
  font-weight: 700;
  margin-bottom: 12px;
}

.detail-line {
  color: var(--el-text-color-secondary);
  margin-bottom: 8px;
  line-height: 1.5;
}

.detail-block {
  margin-top: 20px;
}

.detail-subtitle {
  font-size: 15px;
  font-weight: 600;
  margin-bottom: 12px;
}

@media (max-width: 900px) {
  .wrapper {
    flex-direction: column;
  }

  .sidebar {
    width: 100%;
  }

  .order-body {
    flex-direction: column;
    gap: 16px;
  }

  .ticket-info,
  .price-info {
    width: 100%;
    border: none;
    align-items: flex-start;
    padding: 0;
  }

  .actions {
    width: 100%;
    padding-left: 0;
    align-items: stretch;
  }

  .actions .el-button {
    width: 100%;
  }
}
</style>
