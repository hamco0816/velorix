import { describe, expect, it } from 'vitest'
import { getEffectiveLimitVisibility } from '../planLimits'

describe('getEffectiveLimitVisibility', () => {
  it('hides weekly and monthly windows on daily plans even when set to unlimited', () => {
    expect(getEffectiveLimitVisibility({
      daily_limit_usd: 20,
      weekly_limit_usd: 0,
      monthly_limit_usd: 0,
      validity_days: 1,
      validity_unit: 'day',
    })).toEqual({
      showDaily: true,
      showWeekly: false,
      showMonthly: false,
    })
  })

  it('hides monthly windows on weekly plans but keeps configured weekly limits visible', () => {
    expect(getEffectiveLimitVisibility({
      daily_limit_usd: 20,
      weekly_limit_usd: 0,
      monthly_limit_usd: 0,
      validity_days: 1,
      validity_unit: 'week',
    })).toEqual({
      showDaily: true,
      showWeekly: true,
      showMonthly: false,
    })
  })

  it('allows monthly plans to show configured daily, weekly and monthly limits', () => {
    expect(getEffectiveLimitVisibility({
      daily_limit_usd: 20,
      weekly_limit_usd: 90,
      monthly_limit_usd: 250,
      validity_days: 1,
      validity_unit: 'month',
    })).toEqual({
      showDaily: true,
      showWeekly: true,
      showMonthly: true,
    })
  })

  it('still hides redundant weekly limits when a tighter daily limit already applies', () => {
    expect(getEffectiveLimitVisibility({
      daily_limit_usd: 10,
      weekly_limit_usd: 70,
      monthly_limit_usd: null,
      validity_days: 30,
      validity_unit: 'day',
    })).toEqual({
      showDaily: true,
      showWeekly: false,
      showMonthly: false,
    })
  })
})
