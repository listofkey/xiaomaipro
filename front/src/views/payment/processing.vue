<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'

import { getPaymentStatus } from '@/api/payment'

type ProcessingStage = 'checking' | 'success' | 'timeout' | 'invalid'

const INITIAL_POLL_DELAY = 1800
const POLL_INTERVAL = 2000
const MAX_POLL_ATTEMPTS = 15
const SUCCESS_REDIRECT_DELAY = 1200

const route = useRoute()
const router = useRouter()

const stage = ref<ProcessingStage>('checking')
const attempts = ref(0)
const statusTitle = ref('支付结果确认中...')
const statusDescription = ref('我们正在同步 Stripe 支付结果与系统订单状态，请暂时不要关闭页面。')
const backHintLocked = ref(false)
const pollSerial = ref(0)

const timerIds = new Set<number>()

function getQueryValue(key: string) {
  const raw = route.query[key]
  if (Array.isArray(raw)) {
    return raw[0] || ''
  }
  return typeof raw === 'string' ? raw : ''
}

const orderNo = computed(() => getQueryValue('orderNo'))
const paymentNo = computed(() => getQueryValue('paymentNo'))
const sessionId = computed(() => getQueryValue('session_id') || getQueryValue('sessionId'))

const hintText = computed(() => {
  if (stage.value === 'timeout') {
    return '如果你已完成付款，这不代表支付失败。系统可能仍在处理回调，请稍后前往订单列表刷新查看。'
  }
  if (stage.value === 'success') {
    return '支付已确认，正在为你返回订单列表。'
  }
  if (stage.value === 'invalid') {
    return '当前页面缺少必要的支付参数，建议直接前往订单列表确认订单状态。'
  }
  return '为了避免重复支付，请留在当前页面等待系统完成确认。'
})

const progressWidth = computed(() => {
  if (stage.value === 'success') {
    return '100%'
  }
  if (stage.value === 'timeout' || stage.value === 'invalid') {
    return '100%'
  }
  const value = 18 + (attempts.value / MAX_POLL_ATTEMPTS) * 70
  return `${Math.min(88, Math.max(18, value)).toFixed(0)}%`
})

function trackTimer(timerId: number) {
  timerIds.add(timerId)
  return timerId
}

function clearTrackedTimers() {
  timerIds.forEach((timerId) => window.clearTimeout(timerId))
  timerIds.clear()
}

function wait(ms: number) {
  return new Promise<void>((resolve) => {
    const timerId = trackTimer(window.setTimeout(() => {
      timerIds.delete(timerId)
      resolve()
    }, ms))
  })
}

function buildOrderListQuery() {
  return {
    highlight: orderNo.value || undefined,
    paymentNo: paymentNo.value || undefined,
  }
}

async function goToOrderList() {
  await router.replace({
    path: '/order/list',
    query: buildOrderListQuery(),
  })
}

function lockBackNavigation() {
  window.history.pushState({ paymentProcessing: true }, '', window.location.href)
}

function handlePopState() {
  lockBackNavigation()
  if (backHintLocked.value) {
    return
  }

  backHintLocked.value = true
  ElMessage.info('支付确认中，请稍候由系统自动返回订单列表。')

  trackTimer(window.setTimeout(() => {
    backHintLocked.value = false
  }, 1800))
}

async function startPolling() {
  const currentSerial = ++pollSerial.value
  clearTrackedTimers()
  attempts.value = 0

  if (!orderNo.value && !paymentNo.value) {
    stage.value = 'invalid'
    statusTitle.value = '缺少支付参数'
    statusDescription.value = '页面没有拿到订单号或支付单号，无法继续自动确认支付结果。'
    return
  }

  stage.value = 'checking'
  statusTitle.value = '支付结果确认中...'
  statusDescription.value = 'Stripe 已完成收银页跳转，系统正在等待 webhook 与订单状态同步。'

  await wait(INITIAL_POLL_DELAY)
  if (pollSerial.value !== currentSerial) {
    return
  }

  for (let attempt = 1; attempt <= MAX_POLL_ATTEMPTS; attempt += 1) {
    if (pollSerial.value !== currentSerial) {
      return
    }

    attempts.value = attempt

    try {
      const resp = await getPaymentStatus(
        {
          orderNo: orderNo.value || undefined,
          paymentNo: paymentNo.value || undefined,
          channel: 'stripe',
        },
        {
          silentError: true,
        },
      )

      if (pollSerial.value !== currentSerial) {
        return
      }

      if (resp.paid) {
        stage.value = 'success'
        statusTitle.value = '支付已确认'
        statusDescription.value = '订单状态同步完成，正在返回订单列表。'

        await wait(SUCCESS_REDIRECT_DELAY)
        if (pollSerial.value !== currentSerial) {
          return
        }

        await goToOrderList()
        return
      }
    } catch {
      if (pollSerial.value !== currentSerial) {
        return
      }

      statusDescription.value = '系统正在和支付平台继续同步结果，页面会自动重试，请不要重复付款。'
    }

    if (attempt < MAX_POLL_ATTEMPTS) {
      await wait(POLL_INTERVAL)
    }
  }

  if (pollSerial.value !== currentSerial) {
    return
  }

  stage.value = 'timeout'
  statusTitle.value = '系统处理中，请稍后前往订单列表刷新查看'
  statusDescription.value = '支付平台与订单系统的最终同步可能仍在进行中。如果你已完成付款，请不要重复支付。'
}

onMounted(() => {
  lockBackNavigation()
  window.addEventListener('popstate', handlePopState)
  void startPolling()
})

onBeforeUnmount(() => {
  pollSerial.value += 1
  clearTrackedTimers()
  window.removeEventListener('popstate', handlePopState)
})
</script>

<template>
  <div class="processing-page">
    <div class="ambient ambient-left"></div>
    <div class="ambient ambient-right"></div>

    <section class="processing-card">
      <div class="status-chip">
        <span class="chip-dot" :class="stage"></span>
        <span>{{ stage === 'success' ? '已确认' : stage === 'timeout' ? '处理中' : '确认中' }}</span>
      </div>

      <div class="visual-shell">
        <div class="pulse-ring ring-one"></div>
        <div class="pulse-ring ring-two"></div>
        <div class="pulse-core" :class="stage">
          <span class="core-dot dot-one"></span>
          <span class="core-dot dot-two"></span>
          <span class="core-dot dot-three"></span>
        </div>
      </div>

      <h1 class="title">{{ statusTitle }}</h1>
      <p class="description">{{ statusDescription }}</p>

      <div class="progress-panel">
        <div class="progress-label">
          <span>订单状态同步进度</span>
          <span v-if="stage === 'checking'">第 {{ attempts }}/{{ MAX_POLL_ATTEMPTS }} 次检查</span>
          <span v-else-if="stage === 'success'">同步完成</span>
          <span v-else>等待人工确认</span>
        </div>
        <div class="progress-track">
          <div class="progress-fill" :class="stage" :style="{ width: progressWidth }"></div>
        </div>
      </div>

      <div class="info-grid">
        <div class="info-item">
          <span class="label">订单号</span>
          <span class="value">{{ orderNo || '-' }}</span>
        </div>
        <div class="info-item">
          <span class="label">支付单号</span>
          <span class="value">{{ paymentNo || '-' }}</span>
        </div>
        <div class="info-item">
          <span class="label">Stripe Session</span>
          <span class="value">{{ sessionId || '-' }}</span>
        </div>
      </div>

      <div class="notice-box">
        {{ hintText }}
      </div>

      <div class="actions">
        <el-button
          v-if="stage === 'timeout'"
          type="primary"
          round
          size="large"
          @click="startPolling"
        >
          继续检查
        </el-button>
        <el-button
          v-if="stage !== 'success'"
          plain
          round
          size="large"
          @click="goToOrderList"
        >
          前往订单列表
        </el-button>
      </div>
    </section>
  </div>
</template>

<style scoped>
.processing-page {
  position: relative;
  min-height: 100vh;
  overflow: hidden;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 32px 16px;
  background:
    radial-gradient(circle at top left, rgba(255, 214, 165, 0.72), transparent 42%),
    radial-gradient(circle at bottom right, rgba(247, 133, 90, 0.32), transparent 38%),
    linear-gradient(135deg, #fff7ef 0%, #fffaf6 46%, #fff2e8 100%);
}

.ambient {
  position: absolute;
  border-radius: 999px;
  filter: blur(24px);
  opacity: 0.45;
  pointer-events: none;
}

.ambient-left {
  top: 10%;
  left: -80px;
  width: 220px;
  height: 220px;
  background: rgba(246, 171, 103, 0.55);
  animation: floatAmbient 10s ease-in-out infinite;
}

.ambient-right {
  right: -70px;
  bottom: 12%;
  width: 260px;
  height: 260px;
  background: rgba(225, 114, 68, 0.3);
  animation: floatAmbient 13s ease-in-out infinite reverse;
}

.processing-card {
  position: relative;
  width: min(720px, 100%);
  padding: 32px 28px;
  border-radius: 28px;
  background: rgba(255, 252, 248, 0.86);
  border: 1px solid rgba(225, 153, 82, 0.18);
  box-shadow: 0 24px 70px rgba(150, 83, 32, 0.12);
  backdrop-filter: blur(14px);
}

.status-chip {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  padding: 8px 14px;
  border-radius: 999px;
  background: rgba(255, 245, 233, 0.96);
  color: #9a4c18;
  font-size: 13px;
  font-weight: 600;
}

.chip-dot {
  width: 9px;
  height: 9px;
  border-radius: 50%;
  background: #ef7b45;
  box-shadow: 0 0 0 6px rgba(239, 123, 69, 0.12);
}

.chip-dot.checking {
  animation: pulseDot 1.4s ease-in-out infinite;
}

.chip-dot.success {
  background: #34b368;
  box-shadow: 0 0 0 6px rgba(52, 179, 104, 0.14);
}

.chip-dot.timeout,
.chip-dot.invalid {
  background: #f2a93b;
  box-shadow: 0 0 0 6px rgba(242, 169, 59, 0.14);
}

.visual-shell {
  position: relative;
  width: 172px;
  height: 172px;
  margin: 28px auto 18px;
  display: grid;
  place-items: center;
}

.pulse-ring {
  position: absolute;
  inset: 0;
  border-radius: 50%;
  border: 1px solid rgba(229, 145, 67, 0.22);
}

.ring-one {
  animation: pulseRing 2.8s ease-in-out infinite;
}

.ring-two {
  inset: 16px;
  animation: pulseRing 2.8s ease-in-out infinite 0.3s;
}

.pulse-core {
  position: relative;
  width: 92px;
  height: 92px;
  border-radius: 32px;
  background: linear-gradient(135deg, #f7a35a 0%, #ec6e4d 100%);
  box-shadow: 0 18px 34px rgba(236, 110, 77, 0.28);
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  transform: rotate(12deg);
}

.pulse-core.success {
  background: linear-gradient(135deg, #5ed08c 0%, #25a660 100%);
  box-shadow: 0 18px 34px rgba(37, 166, 96, 0.28);
}

.pulse-core.timeout,
.pulse-core.invalid {
  background: linear-gradient(135deg, #f7c562 0%, #e6953a 100%);
  box-shadow: 0 18px 34px rgba(230, 149, 58, 0.26);
}

.core-dot {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.95);
}

.dot-one {
  animation: bounceDot 1s ease-in-out infinite;
}

.dot-two {
  animation: bounceDot 1s ease-in-out infinite 0.15s;
}

.dot-three {
  animation: bounceDot 1s ease-in-out infinite 0.3s;
}

.title {
  margin: 0;
  text-align: center;
  color: #2b2016;
  font-size: clamp(28px, 4vw, 36px);
  font-weight: 700;
  letter-spacing: -0.02em;
}

.description {
  max-width: 560px;
  margin: 14px auto 0;
  text-align: center;
  color: #6a5445;
  font-size: 15px;
  line-height: 1.75;
}

.progress-panel {
  margin-top: 28px;
  padding: 18px 18px 16px;
  border-radius: 18px;
  background: rgba(255, 255, 255, 0.75);
  border: 1px solid rgba(223, 196, 173, 0.62);
}

.progress-label {
  display: flex;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 12px;
  color: #7a614f;
  font-size: 13px;
}

.progress-track {
  width: 100%;
  height: 10px;
  border-radius: 999px;
  overflow: hidden;
  background: rgba(238, 220, 205, 0.8);
}

.progress-fill {
  height: 100%;
  border-radius: inherit;
  background: linear-gradient(90deg, #f8ae63 0%, #eb6a49 100%);
  transition: width 0.35s ease;
}

.progress-fill.success {
  background: linear-gradient(90deg, #5ed08c 0%, #2aa763 100%);
}

.progress-fill.timeout,
.progress-fill.invalid {
  background: linear-gradient(90deg, #f5c165 0%, #eb9f42 100%);
}

.info-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 14px;
  margin-top: 18px;
}

.info-item {
  min-width: 0;
  padding: 16px;
  border-radius: 18px;
  background: rgba(255, 255, 255, 0.72);
  border: 1px solid rgba(223, 196, 173, 0.56);
}

.label {
  display: block;
  margin-bottom: 8px;
  color: #a08067;
  font-size: 12px;
  text-transform: uppercase;
  letter-spacing: 0.08em;
}

.value {
  display: block;
  color: #2f241a;
  font-size: 14px;
  line-height: 1.6;
  word-break: break-all;
}

.notice-box {
  margin-top: 18px;
  padding: 16px 18px;
  border-radius: 18px;
  background: linear-gradient(135deg, rgba(255, 243, 228, 0.96), rgba(255, 249, 243, 0.98));
  color: #7b5e49;
  line-height: 1.7;
}

.actions {
  display: flex;
  justify-content: center;
  gap: 14px;
  margin-top: 24px;
}

@keyframes pulseRing {
  0% {
    opacity: 0.4;
    transform: scale(0.92);
  }
  50% {
    opacity: 0.78;
    transform: scale(1.02);
  }
  100% {
    opacity: 0.38;
    transform: scale(0.92);
  }
}

@keyframes bounceDot {
  0%,
  80%,
  100% {
    opacity: 0.45;
    transform: translateY(0);
  }
  40% {
    opacity: 1;
    transform: translateY(-7px);
  }
}

@keyframes pulseDot {
  0%,
  100% {
    transform: scale(0.92);
  }
  50% {
    transform: scale(1.08);
  }
}

@keyframes floatAmbient {
  0%,
  100% {
    transform: translate3d(0, 0, 0);
  }
  50% {
    transform: translate3d(0, -18px, 0);
  }
}

@media (max-width: 768px) {
  .processing-card {
    padding: 24px 18px;
    border-radius: 24px;
  }

  .info-grid {
    grid-template-columns: 1fr;
  }

  .progress-label {
    flex-direction: column;
    gap: 6px;
  }

  .actions {
    flex-direction: column;
  }

  .actions :deep(.el-button) {
    width: 100%;
  }
}
</style>
