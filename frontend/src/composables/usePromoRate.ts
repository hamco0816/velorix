import { computed, type ComputedRef } from 'vue'
import { useI18n } from 'vue-i18n'
import { useNowTick } from './useNowTick'

/**
 * 限时倍率（promo rate）展示辅助：根据分组的 promo 配置判定当前是否在限时窗口内，
 * 并生成随秒刷新的倒计时文案。供分组徽章 / 选项等需要展示限时折扣的组件复用，
 * 与价格页（PricingView）保持同一套窗口判定与倒计时格式。
 *
 * billing 计费由后端按相同窗口自动处理，这里仅负责前端展示。
 */
export interface PromoRateInput {
  /** 限时倍率，null/undefined = 未配置限时活动 */
  promoRateMultiplier?: number | null
  /** 限时开始时间（ISO 字符串），空 = 立即生效 */
  promoStartsAt?: string | null
  /** 限时结束时间（ISO 字符串），空 = 永久（不显示倒计时） */
  promoEndsAt?: string | null
}

export interface UsePromoRateResult {
  /** 当前是否处于限时窗口内 */
  promoActive: ComputedRef<boolean>
  /** 倒计时文案（"2天 12:34:56" / "12:34:56"）；无结束时间时为空串 */
  promoCountdown: ComputedRef<string>
}

const MILLIS_PER_DAY = 86_400_000
const MILLIS_PER_HOUR = 3_600_000
const MILLIS_PER_MINUTE = 60_000
const MILLIS_PER_SECOND = 1_000

export function usePromoRate(input: () => PromoRateInput): UsePromoRateResult {
  const now = useNowTick()
  const { t } = useI18n()

  // 是否在 promo 窗口内：依赖 now 让 UI 到点自动切换
  const promoActive = computed<boolean>(() => {
    const { promoRateMultiplier, promoStartsAt, promoEndsAt } = input()
    if (promoRateMultiplier == null) return false
    const current = now.value
    if (promoStartsAt && current < Date.parse(promoStartsAt)) return false
    if (promoEndsAt && current >= Date.parse(promoEndsAt)) return false
    return true
  })

  // 倒计时格式化：返回 "2天 12:34:56" 或 "23:45:12"
  const promoCountdown = computed<string>(() => {
    const { promoEndsAt } = input()
    if (!promoEndsAt) return ''
    const remaining = Math.max(0, Date.parse(promoEndsAt) - now.value)
    const days = Math.floor(remaining / MILLIS_PER_DAY)
    const hours = Math.floor((remaining % MILLIS_PER_DAY) / MILLIS_PER_HOUR)
    const minutes = Math.floor((remaining % MILLIS_PER_HOUR) / MILLIS_PER_MINUTE)
    const seconds = Math.floor((remaining % MILLIS_PER_MINUTE) / MILLIS_PER_SECOND)
    const pad = (n: number) => String(n).padStart(2, '0')
    if (days > 0) {
      return t('pricing.promoCountdownDays', {
        d: days,
        h: pad(hours),
        m: pad(minutes),
        s: pad(seconds)
      })
    }
    return `${pad(hours)}:${pad(minutes)}:${pad(seconds)}`
  })

  return { promoActive, promoCountdown }
}
