<template>
  <div class="space-y-4">
    <!-- ═══ Terminal States: show result, user clicks to return ═══ -->

    <!-- Success -->
    <template v-if="outcome === 'success'">
      <div class="rounded-2xl border border-gray-200/70 bg-white p-6 shadow-[0_1px_2px_rgba(15,23,42,0.04)] dark:border-dark-700/60 dark:bg-dark-800/40">
        <div class="flex flex-col items-center space-y-4 py-4">
          <div class="flex h-14 w-14 items-center justify-center rounded-2xl bg-emerald-50 text-emerald-600 ring-1 ring-inset ring-emerald-200/70 dark:bg-emerald-500/15 dark:text-emerald-300 dark:ring-emerald-500/30">
            <Icon name="check" size="lg" :stroke-width="2.5" />
          </div>
          <p class="text-base font-semibold tracking-tight text-gray-900 dark:text-white">{{ props.orderType === 'subscription' ? t('payment.result.subscriptionSuccess') : t('payment.result.success') }}</p>
          <div v-if="paidOrder" class="w-full rounded-xl border border-gray-200/70 bg-gray-50/40 p-4 dark:border-dark-700/60 dark:bg-dark-800/30">
            <div class="space-y-2 text-sm">
              <div class="flex justify-between">
                <span class="text-gray-500 dark:text-gray-400">{{ t('payment.orders.orderId') }}</span>
                <span class="font-medium tabular-nums text-gray-900 dark:text-white">#{{ paidOrder.id }}</span>
              </div>
              <div v-if="paidOrder.out_trade_no" class="flex justify-between">
                <span class="text-gray-500 dark:text-gray-400">{{ t('payment.orders.orderNo') }}</span>
                <span class="font-medium tabular-nums text-gray-900 dark:text-white">{{ paidOrder.out_trade_no }}</span>
              </div>
              <div class="flex justify-between">
                <span class="text-gray-500 dark:text-gray-400">{{ t('payment.orders.amount') }}</span>
                <span class="font-medium tabular-nums text-gray-900 dark:text-white">{{ paidOrder.order_type === 'balance' ? '$' : '¥' }}{{ paidOrder.amount.toFixed(2) }}</span>
              </div>
              <div class="flex justify-between">
                <span class="text-gray-500 dark:text-gray-400">{{ t('payment.orders.payAmount') }}</span>
                <span class="font-medium tabular-nums text-emerald-600 dark:text-emerald-400">¥{{ paidOrder.pay_amount.toFixed(2) }}</span>
              </div>
            </div>
          </div>
          <button class="btn btn-primary" @click="handleDone">{{ t('common.confirm') }}</button>
        </div>
      </div>
    </template>

    <!-- Cancelled -->
    <template v-else-if="outcome === 'cancelled'">
      <div class="rounded-2xl border border-gray-200/70 bg-white p-6 shadow-[0_1px_2px_rgba(15,23,42,0.04)] dark:border-dark-700/60 dark:bg-dark-800/40">
        <div class="flex flex-col items-center space-y-4 py-4">
          <div class="flex h-14 w-14 items-center justify-center rounded-2xl bg-gray-50 text-gray-500 ring-1 ring-inset ring-gray-200/70 dark:bg-dark-700/40 dark:text-gray-400 dark:ring-dark-600/60">
            <svg class="h-7 w-7" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2.2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </div>
          <p class="text-base font-semibold tracking-tight text-gray-900 dark:text-white">{{ t('payment.qr.cancelled') }}</p>
          <p class="text-sm text-gray-500 dark:text-gray-400">{{ t('payment.qr.cancelledDesc') }}</p>
          <button class="btn btn-primary" @click="handleDone">{{ t('common.confirm') }}</button>
        </div>
      </div>
    </template>

    <!-- Expired / Failed -->
    <template v-else-if="outcome === 'expired'">
      <div class="rounded-2xl border border-gray-200/70 bg-white p-6 shadow-[0_1px_2px_rgba(15,23,42,0.04)] dark:border-dark-700/60 dark:bg-dark-800/40">
        <div class="flex flex-col items-center space-y-4 py-4">
          <div class="flex h-14 w-14 items-center justify-center rounded-2xl bg-amber-50 text-amber-600 ring-1 ring-inset ring-amber-200/70 dark:bg-amber-500/15 dark:text-amber-300 dark:ring-amber-500/30">
            <Icon name="clock" size="lg" :stroke-width="2.2" />
          </div>
          <p class="text-base font-semibold tracking-tight text-gray-900 dark:text-white">{{ t('payment.qr.expired') }}</p>
          <p class="text-sm text-gray-500 dark:text-gray-400">{{ t('payment.qr.expiredDesc') }}</p>
          <button class="btn btn-primary" @click="handleDone">{{ t('common.confirm') }}</button>
        </div>
      </div>
    </template>

    <!-- ═══ Active States: QR or Popup waiting ═══ -->

    <!-- QR Code Mode -->
    <template v-else-if="qrUrl || qrImageUrl">
      <div class="overflow-hidden rounded-2xl border border-gray-200/70 bg-white shadow-[0_1px_2px_rgba(15,23,42,0.04)] dark:border-dark-700/60 dark:bg-dark-800/40">
        <!-- 品牌头部 -->
        <div :class="['flex items-center gap-3 px-5 py-3', qrHeaderBgClass]">
          <span class="flex h-9 w-9 items-center justify-center rounded-full bg-white shadow-sm ring-1 ring-gray-100 dark:bg-dark-900 dark:ring-dark-700">
            <img :src="isAlipay ? alipayIcon : wxpayIcon" alt="" class="h-6 w-6" />
          </span>
          <div class="flex-1">
            <p class="text-sm font-semibold text-gray-900 dark:text-white">{{ scanTitle }}</p>
            <p v-if="scanHint" class="text-xs text-gray-500 dark:text-gray-400">{{ scanHint }}</p>
          </div>
          <div class="hidden text-right sm:block">
            <p class="text-[11px] uppercase tracking-wide text-gray-400 dark:text-gray-500">{{ t('payment.qr.expiresIn') }}</p>
            <p :class="['text-base font-semibold tabular-nums', countdownColorClass]">{{ countdownDisplay }}</p>
          </div>
        </div>
        <!-- 二维码主区域 -->
        <div class="flex flex-col items-center px-5 py-7">
          <!-- 图片加载失败时改用 canvas/payUrl 备用入口 -->
          <div v-if="qrImageUrl && qrImageError" class="flex flex-col items-center gap-3 rounded-xl border border-dashed border-gray-300 px-6 py-8 text-center dark:border-dark-600">
            <Icon name="infoCircle" size="md" class="text-amber-500" />
            <p class="text-sm font-medium text-gray-700 dark:text-gray-200">{{ t('payment.qr.imageLoadFailed') }}</p>
            <button v-if="payUrl" class="text-sm text-primary-600 underline-offset-2 hover:underline dark:text-primary-400" @click="reopenPopup">
              {{ t('payment.qr.openPayWindow') }}
            </button>
          </div>
          <div v-else :class="['rounded-xl border-2 bg-white p-3', qrBorderClass]">
            <img v-if="qrImageUrl" :src="qrImageUrl" alt="" class="block h-[220px] w-[220px]" @error="onQrImageError" @load="onQrImageLoad" />
            <div v-else class="relative">
              <canvas ref="qrCanvas" class="block"></canvas>
              <span :class="['pointer-events-none absolute left-1/2 top-1/2 -translate-x-1/2 -translate-y-1/2 rounded-full p-2 shadow ring-2 ring-white', qrLogoBgClass]">
                <img :src="isAlipay ? alipayIcon : wxpayIcon" alt="" class="h-5 w-5 brightness-0 invert" />
              </span>
            </div>
          </div>
          <div class="mt-5 flex items-center gap-2 text-xs text-gray-500 dark:text-gray-400">
            <span class="relative flex h-2 w-2">
              <span class="absolute inline-flex h-full w-full animate-ping rounded-full bg-emerald-400 opacity-75"></span>
              <span class="relative inline-flex h-2 w-2 rounded-full bg-emerald-500"></span>
            </span>
            <span>{{ t('payment.qr.waitingPayment') }}</span>
          </div>
          <p class="mt-1 text-[11px] sm:hidden" :class="countdownColorClass">
            {{ t('payment.qr.expiresIn') }} · <span class="tabular-nums">{{ countdownDisplay }}</span>
          </p>
        </div>
        <!-- 底部次级操作 -->
        <div class="flex divide-x divide-gray-100 border-t border-gray-100 text-sm dark:divide-dark-700/60 dark:border-dark-700/60">
          <button v-if="payUrl && (!qrImageUrl || qrImageError)"
            class="flex-1 py-2.5 text-gray-600 transition-colors hover:bg-gray-50 dark:text-gray-300 dark:hover:bg-dark-700/60"
            @click="reopenPopup">
            {{ t('payment.qr.openPayWindow') }}
          </button>
          <button
            class="flex-1 py-2.5 text-gray-500 transition-colors hover:bg-gray-50 disabled:cursor-not-allowed disabled:opacity-50 dark:text-gray-400 dark:hover:bg-dark-700/60"
            :disabled="cancelling" @click="handleCancel">
            {{ cancelling ? t('common.processing') : t('payment.qr.cancelOrder') }}
          </button>
        </div>
      </div>
    </template>

    <!-- Waiting for Popup/Redirect Mode -->
    <template v-else>
      <div class="overflow-hidden rounded-2xl border border-gray-200/70 bg-white shadow-[0_1px_2px_rgba(15,23,42,0.04)] dark:border-dark-700/60 dark:bg-dark-800/40">
        <div class="flex flex-col items-center gap-4 px-5 py-10">
          <div class="h-10 w-10 animate-spin rounded-full border-4 border-primary-500 border-t-transparent"></div>
          <p class="text-sm text-gray-700 dark:text-gray-200">{{ t('payment.qr.payInNewWindowHint') }}</p>
          <button v-if="payUrl"
            class="rounded-lg bg-primary-50 px-4 py-2 text-sm font-medium text-primary-700 ring-1 ring-inset ring-primary-200/70 transition-colors hover:bg-primary-100 dark:bg-primary-500/15 dark:text-primary-300 dark:ring-primary-500/30 dark:hover:bg-primary-500/25"
            @click="reopenPopup">
            <Icon name="arrowRight" size="xs" class="mr-1 inline-block" />
            {{ t('payment.qr.openPayWindow') }}
          </button>
          <div class="flex items-center gap-2 text-xs text-gray-500 dark:text-gray-400">
            <span>{{ t('payment.qr.expiresIn') }}</span>
            <span :class="['font-semibold tabular-nums', countdownColorClass]">{{ countdownDisplay }}</span>
          </div>
          <p class="inline-flex items-center gap-1.5 text-[11px] text-gray-400 dark:text-gray-500">
            <span class="h-1.5 w-1.5 rounded-full bg-sky-500 animate-pulse"></span>
            {{ t('payment.qr.waitingPayment') }}
          </p>
        </div>
        <div class="flex border-t border-gray-100 dark:border-dark-700/60">
          <button
            class="flex-1 py-2.5 text-sm text-gray-500 transition-colors hover:bg-gray-50 disabled:cursor-not-allowed disabled:opacity-50 dark:text-gray-400 dark:hover:bg-dark-700/60"
            :disabled="cancelling" @click="handleCancel">
            {{ cancelling ? t('common.processing') : t('payment.qr.cancelOrder') }}
          </button>
        </div>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onUnmounted, nextTick } from 'vue'
import { useI18n } from 'vue-i18n'
import { usePaymentStore } from '@/stores/payment'
import { useAppStore } from '@/stores'
import { paymentAPI } from '@/api/payment'
import { extractI18nErrorMessage } from '@/utils/apiError'
import { getPaymentPopupFeatures } from '@/components/payment/providerConfig'
import type { PaymentOrder } from '@/types/payment'
import Icon from '@/components/icons/Icon.vue'
import QRCode from 'qrcode'
import alipayIcon from '@/assets/icons/alipay.svg'
import wxpayIcon from '@/assets/icons/wxpay.svg'

const props = defineProps<{
  orderId: number
  qrCode: string
  qrCodeImage?: string
  expiresAt: string
  paymentType: string
  payUrl?: string
  orderType?: string
}>()

type PaymentOutcome = 'success' | 'cancelled' | 'expired'

const emit = defineEmits<{ done: []; success: []; settled: [outcome: PaymentOutcome] }>()

const { t } = useI18n()
const paymentStore = usePaymentStore()
const appStore = useAppStore()

const qrCanvas = ref<HTMLCanvasElement | null>(null)
const qrUrl = ref('')
const qrImageUrl = computed(() => props.qrCodeImage || '')
const qrImageError = ref(false)

function onQrImageError() { qrImageError.value = true }
function onQrImageLoad() { qrImageError.value = false }
const remainingSeconds = ref(0)
const cancelling = ref(false)
const paidOrder = ref<PaymentOrder | null>(null)

// Terminal outcome: null = still active, 'success' | 'cancelled' | 'expired'
const outcome = ref<PaymentOutcome | null>(null)

let pollTimer: ReturnType<typeof setInterval> | null = null
let countdownTimer: ReturnType<typeof setInterval> | null = null

const isAlipay = computed(() => props.paymentType.includes('alipay'))
const isWxpay = computed(() => props.paymentType.includes('wxpay'))

const qrBorderClass = computed(() => {
  if (isAlipay.value) return 'border-[#00AEEF]/40'
  if (isWxpay.value) return 'border-[#2BB741]/40'
  return 'border-gray-200'
})

const qrHeaderBgClass = computed(() => {
  if (isAlipay.value) return 'bg-gradient-to-r from-[#E6F6FE] to-transparent dark:from-[#00AEEF]/10'
  if (isWxpay.value) return 'bg-gradient-to-r from-[#E8F6EB] to-transparent dark:from-[#2BB741]/10'
  return 'bg-gray-50 dark:bg-dark-700'
})

const qrLogoBgClass = computed(() => {
  if (isAlipay.value) return 'bg-[#00AEEF]'
  if (isWxpay.value) return 'bg-[#2BB741]'
  return 'bg-gray-400'
})

const scanTitle = computed(() => {
  if (isAlipay.value) return t('payment.qr.scanAlipay')
  if (isWxpay.value) return t('payment.qr.scanWxpay')
  return t('payment.qr.scanToPay')
})

const scanHint = computed(() => {
  if (isAlipay.value) return t('payment.qr.scanAlipayHint')
  if (isWxpay.value) return t('payment.qr.scanWxpayHint')
  return ''
})

const countdownDisplay = computed(() => {
  const m = Math.floor(remainingSeconds.value / 60)
  const s = remainingSeconds.value % 60
  return m.toString().padStart(2, '0') + ':' + s.toString().padStart(2, '0')
})

const countdownColorClass = computed(() => {
  if (remainingSeconds.value <= 60) return 'text-red-500 dark:text-red-400'
  if (remainingSeconds.value <= 300) return 'text-amber-500 dark:text-amber-400'
  return 'text-gray-900 dark:text-white'
})

function isSuccessStatus(status: string | null | undefined): boolean {
  return status === 'COMPLETED' || status === 'PAID' || status === 'RECHARGING'
}

function reopenPopup() {
  if (props.payUrl) {
    const win = window.open(props.payUrl, 'paymentPopup', getPaymentPopupFeatures())
    if (!win || win.closed) {
      window.location.href = props.payUrl
    }
  }
}

function setOutcome(next: PaymentOutcome) {
  if (outcome.value === next) return
  outcome.value = next
  emit('settled', next)
}

async function renderQR() {
  await nextTick()
  if (!qrCanvas.value || !qrUrl.value) return
  await QRCode.toCanvas(qrCanvas.value, qrUrl.value, {
    width: 220, margin: 2,
    errorCorrectionLevel: 'M',
  })
}

async function pollStatus() {
  if (!props.orderId || outcome.value) return
  const order = await paymentStore.pollOrderStatus(props.orderId)
  if (!order) return
  if (isSuccessStatus(order.status)) {
    cleanup()
    paidOrder.value = order
    setOutcome('success')
    emit('success')
  } else if (order.status === 'CANCELLED') {
    cleanup()
    setOutcome('cancelled')
  } else if (order.status === 'EXPIRED' || order.status === 'FAILED') {
    cleanup()
    setOutcome('expired')
  }
}

function startCountdown(seconds: number) {
  remainingSeconds.value = Math.max(0, seconds)
  if (remainingSeconds.value <= 0) { setOutcome('expired'); return }
  countdownTimer = setInterval(() => {
    remainingSeconds.value--
    if (remainingSeconds.value <= 0) { setOutcome('expired'); cleanup() }
  }, 1000)
}

async function handleCancel() {
  if (!props.orderId || cancelling.value) return
  cancelling.value = true
  try {
    await paymentAPI.cancelOrder(props.orderId)
    cleanup()
    setOutcome('cancelled')
  } catch (err: unknown) {
    appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error')))
  } finally {
    cancelling.value = false
  }
}

function handleDone() { cleanup(); emit('done') }

function cleanup() {
  if (pollTimer) { clearInterval(pollTimer); pollTimer = null }
  if (countdownTimer) { clearInterval(countdownTimer); countdownTimer = null }
}

// Initialize on mount
qrUrl.value = props.qrCode
let seconds = 30 * 60
if (props.expiresAt) {
  seconds = Math.floor((new Date(props.expiresAt).getTime() - Date.now()) / 1000)
}
startCountdown(seconds)
pollTimer = setInterval(pollStatus, 3000)
renderQR()

watch(() => qrUrl.value, () => renderQR())
onUnmounted(() => cleanup())
</script>
