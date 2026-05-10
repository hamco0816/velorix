// 套餐卡类型推导：纯前端从 validity_days + validity_unit 算出"日/周/月/季/年/自定义"
// 标签，不需要后端字段，跟独享/共享 kind 完全正交。
//
// 归一化规则（与后端 psComputeValidityDays 对齐）：
//   day   → days × 1
//   week  → days × 7
//   month → days × 30   ← 简化为 30，不是自然月（schema 限制）
//
// 标签匹配（必须用归一化后的总天数判定，不能只看 unit）：
//   1   day  → daily
//   7   day  → weekly
//   30  day  → monthly
//   90  day  → quarterly
//   365 day  → yearly
//   其它      → custom

export type PlanCardType = 'daily' | 'weekly' | 'monthly' | 'quarterly' | 'yearly' | 'custom'

const DAY = 1
const WEEK = 7
const MONTH = 30
const QUARTER = 90
const YEAR = 365

/**
 * 推导套餐的卡类型。
 *
 * @param validityDays 套餐 validity_days 字段
 * @param validityUnit 套餐 validity_unit 字段（'day' / 'week' / 'month'，宽容大小写）
 * @returns 卡类型，无法归类时返回 'custom'
 */
export function derivePlanCardType(
  validityDays: number | null | undefined,
  validityUnit: string | null | undefined,
): PlanCardType {
  const total = normalizeToDays(validityDays, validityUnit)
  if (total <= 0) return 'custom'
  switch (total) {
    case DAY: return 'daily'
    case WEEK: return 'weekly'
    case MONTH: return 'monthly'
    case QUARTER: return 'quarterly'
    case YEAR: return 'yearly'
    default: return 'custom'
  }
}

/** 归一化为总天数；与后端 psComputeValidityDays 保持一致。 */
export function normalizeToDays(
  validityDays: number | null | undefined,
  validityUnit: string | null | undefined,
): number {
  const days = typeof validityDays === 'number' ? validityDays : 0
  if (days <= 0) return 0
  const unit = (validityUnit || 'day').trim().toLowerCase()
  // 兼容单复数：days/day、weeks/week、months/month
  switch (unit) {
    case 'week':
    case 'weeks':
      return days * 7
    case 'month':
    case 'months':
      return days * 30
    case 'day':
    case 'days':
    default:
      return days
  }
}

// 排序权重：custom 排到最后，方便 tab 显示顺序
export const CARD_TYPE_ORDER: PlanCardType[] = ['daily', 'weekly', 'monthly', 'quarterly', 'yearly', 'custom']

/**
 * 给定一组套餐（含 validity_days/unit），返回当前实际存在的卡类型集合，
 * 按 CARD_TYPE_ORDER 排序。用于动态生成筛选 tab，不显示空类型。
 */
export function collectCardTypes(
  plans: Array<{ validity_days: number; validity_unit: string }>,
): PlanCardType[] {
  const seen = new Set<PlanCardType>()
  for (const p of plans) {
    seen.add(derivePlanCardType(p.validity_days, p.validity_unit))
  }
  return CARD_TYPE_ORDER.filter((t) => seen.has(t))
}

/** 卡类型 → 徽章颜色样式（套餐卡片用）。和「独享」徽章不同色系，避免视觉冲突。 */
export function cardTypeBadgeClass(type: PlanCardType): string {
  switch (type) {
    case 'daily':
      return 'bg-sky-50 text-sky-700 ring-1 ring-sky-200 dark:bg-sky-900/30 dark:text-sky-300 dark:ring-sky-900/50'
    case 'weekly':
      return 'bg-cyan-50 text-cyan-700 ring-1 ring-cyan-200 dark:bg-cyan-900/30 dark:text-cyan-300 dark:ring-cyan-900/50'
    case 'monthly':
      return 'bg-emerald-50 text-emerald-700 ring-1 ring-emerald-200 dark:bg-emerald-900/30 dark:text-emerald-300 dark:ring-emerald-900/50'
    case 'quarterly':
      return 'bg-amber-50 text-amber-700 ring-1 ring-amber-200 dark:bg-amber-900/30 dark:text-amber-300 dark:ring-amber-900/50'
    case 'yearly':
      return 'bg-rose-50 text-rose-700 ring-1 ring-rose-200 dark:bg-rose-900/30 dark:text-rose-300 dark:ring-rose-900/50'
    case 'custom':
    default:
      return 'bg-gray-50 text-gray-600 ring-1 ring-gray-200 dark:bg-dark-800 dark:text-gray-400 dark:ring-dark-600'
  }
}
