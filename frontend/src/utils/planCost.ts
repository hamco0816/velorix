import { normalizeToDays } from './planCardType'

type PlanCostPlan = {
  price?: number | null
  validity_days?: number | null
  validity_unit?: string | null
  rate_multiplier?: number | null
  daily_limit_usd?: number | null
  weekly_limit_usd?: number | null
  monthly_limit_usd?: number | null
}

type PlanCostGroup = {
  rate_multiplier?: number | null
  daily_limit_usd?: number | null
  weekly_limit_usd?: number | null
  monthly_limit_usd?: number | null
}

export type PlanCostBasis = 'daily' | 'weekly' | 'monthly' | 'prorated' | 'none'

export interface PlanCostEstimate {
  periodLimitUSD: number | null
  priceQuotaMultiplier: number | null
  rateMultiplier: number
  effectiveCostMultiplier: number | null
  basis: PlanCostBasis
}

function positive(value: number | null | undefined): number | null {
  return typeof value === 'number' && Number.isFinite(value) && value > 0 ? value : null
}

function effectiveValue(planValue: number | null | undefined, groupValue: number | null | undefined): number | null {
  return positive(planValue) ?? positive(groupValue)
}

function resolvePeriodLimit(plan: PlanCostPlan, group?: PlanCostGroup | null): { limit: number | null; basis: PlanCostBasis } {
  const totalDays = normalizeToDays(plan.validity_days, plan.validity_unit)
  if (totalDays <= 0) return { limit: null, basis: 'none' }

  const daily = effectiveValue(plan.daily_limit_usd, group?.daily_limit_usd)
  const weekly = effectiveValue(plan.weekly_limit_usd, group?.weekly_limit_usd)
  const monthly = effectiveValue(plan.monthly_limit_usd, group?.monthly_limit_usd)

  if (totalDays === 1 && daily) return { limit: daily, basis: 'daily' }
  if (totalDays === 7 && weekly) return { limit: weekly, basis: 'weekly' }
  if (totalDays === 30 && monthly) return { limit: monthly, basis: 'monthly' }
  if (totalDays % 30 === 0 && monthly) return { limit: monthly * (totalDays / 30), basis: 'monthly' }
  if (totalDays % 7 === 0 && weekly) return { limit: weekly * (totalDays / 7), basis: 'weekly' }
  if (daily) return { limit: daily * totalDays, basis: 'daily' }
  if (monthly) return { limit: monthly * (totalDays / 30), basis: 'prorated' }
  if (weekly) return { limit: weekly * (totalDays / 7), basis: 'prorated' }
  return { limit: null, basis: 'none' }
}

export function calculatePlanCostEstimate(plan: PlanCostPlan, group?: PlanCostGroup | null): PlanCostEstimate {
  const { limit, basis } = resolvePeriodLimit(plan, group)
  const price = positive(plan.price)
  const rateMultiplier = effectiveValue(plan.rate_multiplier, group?.rate_multiplier) ?? 1
  const priceQuotaMultiplier = price && limit ? price / limit : null
  return {
    periodLimitUSD: limit,
    priceQuotaMultiplier,
    rateMultiplier,
    effectiveCostMultiplier: priceQuotaMultiplier === null ? null : priceQuotaMultiplier * rateMultiplier,
    basis,
  }
}

export function formatCostMultiplier(value: number | null | undefined, digits = 4): string {
  if (typeof value !== 'number' || !Number.isFinite(value)) return '-'
  const fixed = value >= 1 ? value.toFixed(2) : value.toFixed(digits)
  return `${Number(fixed)}x`
}
