<script setup lang="ts">
import { computed, onBeforeUnmount, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'

import { createOrder, getOrderQueueStatus } from '@/api/order'
import { getEventDetail, type ProgramEventDetail, type ProgramTicketTierInfo } from '@/api/program'
import { listAddresses, listTicketBuyers, type AddressPayload, type TicketBuyerPayload } from '@/api/user'
import {
  canPurchaseEvent,
  deriveTicketTierState,
  formatCurrency,
  formatEventDetailTime,
  getEventStatusLabel,
  resolvePosterUrl,
} from '@/utils/program'

type DeliveryMethod = 'eticket' | 'paper'
type SelectableAttendee = TicketBuyerPayload & { selected: boolean }

const route = useRoute()
const router = useRouter()

const pageLoading = ref(false)
const submitting = ref(false)
const queueToken = ref('')
const eventDetail = ref<ProgramEventDetail | null>(null)
const attendees = ref<SelectableAttendee[]>([])
const addresses = ref<AddressPayload[]>([])
const deliveryMethod = ref<DeliveryMethod>('eticket')
const selectedAddressId = ref('')
const pollSerial = ref(0)

function readQueryValue(value: unknown) {
  if (Array.isArray(value)) {
    return typeof value[0] === 'string' ? value[0] : ''
  }

  return typeof value === 'string' ? value : ''
}

function parseCount(value: string) {
  const parsed = Number.parseInt(value, 10)
  return Number.isFinite(parsed) && parsed > 0 ? parsed : 1
}

function maskIdCard(idCard: string) {
  if (!idCard) {
    return '未填写身份证'
  }

  if (idCard.length <= 8) {
    return idCard
  }

  return `${idCard.slice(0, 4)}******${idCard.slice(-4)}`
}

function formatFullAddress(address: AddressPayload) {
  return [address.province, address.city, address.district, address.detail]
    .filter(Boolean)
    .join(' ')
}

function openProfileTab(tab: 'attendees' | 'addresses') {
  router.push({
    path: '/user/profile',
    query: { tab },
  })
}

const activityId = computed(() => readQueryValue(route.query.activity))
const requestedTicketId = computed(() => readQueryValue(route.query.ticket))
const count = computed(() => parseCount(readQueryValue(route.query.count)))

const selectedTicket = computed<ProgramTicketTierInfo | null>(() => {
  const tiers = eventDetail.value?.ticketTiers || []
  if (!tiers.length) {
    return null
  }

  return (
    tiers.find((item) => item.id === requestedTicketId.value) ||
    tiers.find((item) => !deriveTicketTierState(item, eventDetail.value).disabled) ||
    tiers[0] ||
    null
  )
})

const selectedAttendees = computed(() => attendees.value.filter((item) => item.selected))
const selectedAddress = computed(() =>
  addresses.value.find((item) => item.id === selectedAddressId.value) || null,
)
const requiresAddress = computed(() => deliveryMethod.value === 'paper')
const ticketState = computed(() =>
  selectedTicket.value ? deriveTicketTierState(selectedTicket.value, eventDetail.value) : null,
)
const posterUrl = computed(() =>
  resolvePosterUrl(eventDetail.value?.posterUrl, eventDetail.value?.title),
)
const subtotal = computed(() => (selectedTicket.value?.price || 0) * count.value)
const freight = computed(() => (requiresAddress.value ? 18 : 0))
const finalPrice = computed(() => subtotal.value + freight.value)
const remainingAttendeeCount = computed(() =>
  Math.max(count.value - selectedAttendees.value.length, 0),
)
const submitDisabled = computed(() =>
  pageLoading.value || submitting.value || !eventDetail.value || !selectedTicket.value,
)

function applyAttendees(items: TicketBuyerPayload[]) {
  const ordered = [...items].sort((a, b) => Number(b.isDefault) - Number(a.isDefault))
  const presetIds = new Set(ordered.slice(0, count.value).map((item) => item.id))
  attendees.value = ordered.map((item) => ({
    ...item,
    selected: presetIds.has(item.id),
  }))
}

function applyAddresses(items: AddressPayload[]) {
  const ordered = [...items].sort((a, b) => Number(b.isDefault) - Number(a.isDefault))
  addresses.value = ordered
  selectedAddressId.value = ordered.find((item) => item.isDefault)?.id || ordered[0]?.id || ''
}

function syncDeliveryMethod() {
  if (eventDetail.value?.ticketType === 2) {
    deliveryMethod.value = 'paper'
    if (!selectedAddressId.value) {
      selectedAddressId.value = addresses.value[0]?.id || ''
    }
    return
  }

  deliveryMethod.value = 'eticket'
  selectedAddressId.value = ''
}

function handleAttendeeChange(item: SelectableAttendee) {
  if (selectedAttendees.value.length > count.value) {
    ElMessage.warning(`最多只能选择 ${count.value} 位观演人`)
    item.selected = false
  }
}

async function initializePage() {
  pollSerial.value += 1
  queueToken.value = ''
  pageLoading.value = true
  try {
    const tasks: [
      PromiseSettledResult<Awaited<ReturnType<typeof getEventDetail>>>?,
      PromiseSettledResult<Awaited<ReturnType<typeof listTicketBuyers>>>?,
      PromiseSettledResult<Awaited<ReturnType<typeof listAddresses>>>?,
    ] = await Promise.allSettled([
      activityId.value ? getEventDetail(activityId.value) : Promise.resolve({ event: null as never }),
      listTicketBuyers(),
      listAddresses(),
    ]) as [
      PromiseSettledResult<Awaited<ReturnType<typeof getEventDetail>>>,
      PromiseSettledResult<Awaited<ReturnType<typeof listTicketBuyers>>>,
      PromiseSettledResult<Awaited<ReturnType<typeof listAddresses>>>,
    ]

    const [eventResult, attendeeResult, addressResult] = tasks

    if (eventResult?.status === 'fulfilled' && activityId.value) {
      eventDetail.value = eventResult.value.event
    } else {
      eventDetail.value = null
    }

    if (attendeeResult?.status === 'fulfilled') {
      applyAttendees(attendeeResult.value.ticketBuyers || [])
    } else {
      attendees.value = []
    }

    if (addressResult?.status === 'fulfilled') {
      applyAddresses(addressResult.value.addresses || [])
    } else {
      addresses.value = []
      selectedAddressId.value = ''
    }

    syncDeliveryMethod()
  } finally {
    pageLoading.value = false
  }
}

function generateRequestId() {
  if (typeof crypto !== 'undefined' && typeof crypto.randomUUID === 'function') {
    return crypto.randomUUID()
  }

  return `${Date.now()}-${Math.random().toString(36).slice(2, 10)}`
}

function sleep(ms: number) {
  return new Promise((resolve) => window.setTimeout(resolve, ms))
}

async function pollQueue(queueTokenValue: string, fallbackOrderNo: string) {
  const currentSerial = ++pollSerial.value
  queueToken.value = queueTokenValue

  for (let attempt = 0; attempt < 20; attempt += 1) {
    if (pollSerial.value !== currentSerial) {
      return
    }

    const queueResp = await getOrderQueueStatus({ queueToken: queueTokenValue })
    if (queueResp.queueStatus === 3) {
      ElMessage.success('抢票成功，订单已生成')
      await router.push({
        path: '/order/list',
        query: {
          highlight: queueResp.order?.orderNo || fallbackOrderNo,
        },
      })
      return
    }

    if (queueResp.queueStatus === 4) {
      ElMessage.error(queueResp.message || '抢票失败，请重新尝试')
      return
    }

    await sleep(1000)
  }

  ElMessage.warning('当前仍在排队中，已跳转到订单列表继续查看')
  await router.push({
    path: '/order/list',
    query: {
      highlight: fallbackOrderNo,
    },
  })
}

async function submitOrder() {
  if (!eventDetail.value || !selectedTicket.value) {
    ElMessage.error('订单活动信息加载失败')
    return
  }

  if (!canPurchaseEvent(eventDetail.value)) {
    ElMessage.error('当前活动不在可下单时间范围内')
    return
  }

  if (ticketState.value?.disabled) {
    ElMessage.error(ticketState.value.label || '当前票档不可下单')
    return
  }

  if (count.value > eventDetail.value.purchaseLimit) {
    ElMessage.error(`当前活动每单最多购买 ${eventDetail.value.purchaseLimit} 张`)
    return
  }

  if (count.value > selectedTicket.value.remainStock) {
    ElMessage.error('当前票档余票不足，请返回详情页重新选择')
    return
  }

  if (selectedAttendees.value.length !== count.value) {
    ElMessage.error(`请选择 ${count.value} 位观演人`)
    return
  }

  if (requiresAddress.value && !selectedAddress.value) {
    ElMessage.error('请选择收货地址')
    return
  }

  submitting.value = true
  try {
    const resp = await createOrder({
      eventId: eventDetail.value.id,
      ticketTierId: selectedTicket.value.id,
      quantity: count.value,
      ticketBuyerIds: selectedAttendees.value.map((item) => item.id),
      addressId: requiresAddress.value ? selectedAddress.value?.id : undefined,
      payMethod: 1,
      requestId: generateRequestId(),
    })

    ElMessage.info(resp.message || '已进入抢票队列，请稍候')
    await pollQueue(resp.queueToken, resp.orderNo)
  } finally {
    submitting.value = false
    queueToken.value = ''
  }
}

watch(
  () => route.fullPath,
  () => {
    void initializePage()
  },
  { immediate: true },
)

onBeforeUnmount(() => {
  pollSerial.value += 1
})
</script>

<template>
  <div class="order-create-container wrapper" v-loading="pageLoading">
    <h2 class="page-title">确认订单信息</h2>

    <el-empty
      v-if="!pageLoading && (!eventDetail || !selectedTicket)"
      description="未找到当前下单所需的活动或票档信息"
    />

    <template v-else-if="eventDetail && selectedTicket">
      <div class="section-box">
        <h3 class="section-title">购买明细</h3>
        <div class="activity-info-card">
          <img class="cover" :src="posterUrl" :alt="eventDetail.title" />
          <div class="info">
            <div class="title">{{ eventDetail.title }}</div>
            <div class="desc">
              <el-icon><Calendar /></el-icon>
              {{ formatEventDetailTime(eventDetail.eventStartTime, eventDetail.eventEndTime) }}
            </div>
            <div class="desc">
              <el-icon><Location /></el-icon>
              {{ eventDetail.venue.name }}
            </div>
            <div class="desc">
              <el-icon><Ticket /></el-icon>
              {{ selectedTicket.name }} | ¥{{ formatCurrency(selectedTicket.price) }}
            </div>
            <div class="status-row">
              <el-tag size="small" type="success">{{ getEventStatusLabel(eventDetail.status) }}</el-tag>
              <el-tag size="small" :type="ticketState?.tagType || 'info'">{{ ticketState?.label || '状态未知' }}</el-tag>
            </div>
          </div>
          <div class="price-box">
            <div class="price">¥{{ formatCurrency(selectedTicket.price) }} x {{ count }}</div>
            <div class="stock">余票 {{ selectedTicket.remainStock }}</div>
            <div class="total">小计：<span class="highlight">¥{{ formatCurrency(subtotal) }}</span></div>
          </div>
        </div>
      </div>

      <div class="section-box">
        <div class="section-header">
          <h3 class="section-title">选择观演人</h3>
          <span class="subtitle">
            还需要选择 <span class="highlight">{{ remainingAttendeeCount }}</span> 位观演人（请真实有效填写，入场需一致）
          </span>
        </div>

        <div v-if="!attendees.length" class="empty-block">
          <el-empty description="还没有常用观演人" :image-size="90" />
          <el-button type="primary" plain @click="openProfileTab('attendees')">
            <el-icon><Plus /></el-icon>
            <span>前往新增观演人</span>
          </el-button>
        </div>

        <div v-else class="attendees-list">
          <el-checkbox-button
            v-for="item in attendees"
            :key="item.id"
            v-model="item.selected"
            @change="handleAttendeeChange(item)"
            class="attendee-btn"
          >
            <div class="attendee-info">
              <span class="name">{{ item.name }}</span>
              <span class="idcard">{{ maskIdCard(item.idCard) }}</span>
              <el-tag v-if="item.isDefault" size="small" type="success">默认</el-tag>
            </div>
          </el-checkbox-button>
          <el-button type="primary" plain class="add-btn" @click="openProfileTab('attendees')">
            <el-icon><Plus /></el-icon>
            <span>新增观演人</span>
          </el-button>
        </div>
      </div>

      <div class="section-box">
        <h3 class="section-title">选择配送方式</h3>
        <el-radio-group v-model="deliveryMethod" size="large">
          <el-radio-button value="eticket" :disabled="eventDetail.ticketType === 2">电子票</el-radio-button>
          <el-radio-button value="paper" :disabled="eventDetail.ticketType !== 2">顺丰快递 (运费¥18)</el-radio-button>
        </el-radio-group>

        <div class="delivery-tip">
          当前活动票种：{{ eventDetail.ticketType === 2 ? '纸质票配送' : '电子票入场' }}
        </div>

        <div v-if="requiresAddress" class="address-box">
          <el-divider border-style="dashed" />
          <h4>选择收货地址</h4>

          <div v-if="!addresses.length" class="empty-block">
            <el-empty description="还没有可用的收货地址" :image-size="90" />
            <el-button type="primary" plain @click="openProfileTab('addresses')">
              <el-icon><Plus /></el-icon>
              <span>前往新增地址</span>
            </el-button>
          </div>

          <template v-else>
            <el-radio-group v-model="selectedAddressId" class="address-group">
              <el-radio
                v-for="addr in addresses"
                :key="addr.id"
                :label="addr.id"
                border
                class="address-radio-card"
              >
                <div class="address-info">
                  <span class="name">{{ addr.receiverName }}</span>
                  <span class="phone">{{ addr.receiverPhone }}</span>
                  <span class="detail" :title="formatFullAddress(addr)">{{ formatFullAddress(addr) }}</span>
                  <el-tag v-if="addr.isDefault" size="small" type="success">默认</el-tag>
                </div>
              </el-radio>
            </el-radio-group>
            <div class="add-address-wrapper">
              <el-button type="primary" plain class="add-addr-btn" @click="openProfileTab('addresses')">
                <el-icon><Plus /></el-icon>
                <span>新增收货地址</span>
              </el-button>
            </div>
          </template>
        </div>
      </div>

      <div class="summary-box">
        <div class="freight">运费: ¥ {{ formatCurrency(freight) }}</div>
        <div class="final-price">
          实付款：<span class="highlight text-xl">¥ {{ formatCurrency(finalPrice) }}</span>
        </div>
        <el-button
          type="danger"
          size="large"
          class="submit-btn"
          @click="submitOrder"
          :loading="submitting"
          :disabled="submitDisabled"
        >
          {{ queueToken ? '排队中...' : submitting ? '提交中...' : '提交订单' }}
        </el-button>
      </div>
    </template>
  </div>
</template>

<style scoped>
.wrapper {
  max-width: 1000px;
  margin: 0 auto;
  padding: 30px 0;
}

.page-title {
  font-size: 24px;
  margin-top: 0;
  margin-bottom: 24px;
  color: var(--el-text-color-primary);
}

.section-box {
  background: var(--el-bg-color);
  border-radius: 8px;
  padding: 24px;
  margin-bottom: 20px;
  box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.1);
}

.section-title {
  font-size: 18px;
  margin-top: 0;
  margin-bottom: 20px;
  color: var(--el-text-color-primary);
  border-left: 4px solid #409EFF;
  padding-left: 10px;
}

.section-header {
  display: flex;
  align-items: center;
  margin-bottom: 20px;
}

.section-header .section-title {
  margin-bottom: 0;
}

.subtitle {
  margin-left: 16px;
  font-size: 13px;
  color: var(--el-text-color-secondary);
}

.activity-info-card {
  display: flex;
  background: var(--el-fill-color-light);
  padding: 16px;
  border-radius: 8px;
}

.cover {
  width: 90px;
  height: 120px;
  object-fit: cover;
  border-radius: 4px;
  margin-right: 20px;
}

.info {
  flex: 1;
}

.info .title {
  font-size: 16px;
  font-weight: bold;
  color: var(--el-text-color-primary);
  margin-bottom: 12px;
}

.info .desc {
  font-size: 13px;
  color: #606266;
  margin-bottom: 6px;
  display: flex;
  align-items: center;
  gap: 6px;
}

.status-row {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-top: 10px;
}

.price-box {
  width: 200px;
  text-align: right;
  display: flex;
  flex-direction: column;
  justify-content: center;
  gap: 6px;
}

.price,
.stock {
  font-size: 14px;
  color: #606266;
}

.total {
  font-size: 16px;
  color: var(--el-text-color-primary);
}

.highlight {
  color: #f56c6c;
  font-weight: bold;
}

.text-xl {
  font-size: 28px;
}

.attendees-list {
  display: flex;
  flex-wrap: wrap;
  gap: 16px;
}

.attendee-btn :deep(.el-checkbox-button__inner) {
  height: 56px;
  display: flex;
  align-items: center;
  padding: 0 24px;
  border-left: 1px solid #dcdfe6;
  border-radius: 4px !important;
}

.attendee-info {
  display: flex;
  align-items: center;
  gap: 12px;
}

.attendee-info .name {
  font-size: 15px;
  font-weight: bold;
  min-width: 50px;
}

.attendee-info .idcard {
  font-size: 13px;
  color: var(--el-text-color-secondary);
}

.add-btn {
  height: 56px;
  padding: 0 24px;
}

.delivery-tip {
  margin-top: 12px;
  font-size: 13px;
  color: var(--el-text-color-secondary);
}

.address-group {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.address-radio-card {
  width: 100%;
  min-height: 72px;
  margin-left: 0 !important;
  margin-right: 0 !important;
  display: flex;
  align-items: center;
  padding: 0 20px !important;
  border-radius: 4px;
}

.address-radio-card :deep(.el-radio__label) {
  flex: 1;
}

.address-info {
  display: flex;
  align-items: center;
  gap: 20px;
}

.address-info .name {
  font-weight: bold;
  font-size: 15px;
  min-width: 60px;
}

.address-info .phone {
  font-size: 14px;
  color: #606266;
  min-width: 100px;
}

.address-info .detail {
  font-size: 13px;
  color: var(--el-text-color-secondary);
  flex: 1;
}

.empty-block {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
  padding: 12px 0;
}

.add-addr-btn {
  width: 100%;
  height: 64px;
  margin-top: 12px;
}

.summary-box {
  background: var(--el-bg-color);
  padding: 24px;
  border-radius: 8px;
  box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.1);
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 24px;
}

.freight {
  font-size: 14px;
  color: #606266;
}

.final-price {
  font-size: 16px;
  color: var(--el-text-color-primary);
  display: flex;
  align-items: center;
}

.submit-btn {
  width: 160px;
}

@media (max-width: 900px) {
  .activity-info-card {
    flex-direction: column;
  }

  .cover {
    margin-right: 0;
    margin-bottom: 16px;
  }

  .price-box {
    width: 100%;
    text-align: left;
    margin-top: 16px;
  }

  .summary-box {
    flex-direction: column;
    align-items: stretch;
  }

  .submit-btn {
    width: 100%;
  }
}
</style>
