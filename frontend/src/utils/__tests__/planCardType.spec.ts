import { describe, expect, it } from 'vitest'
import {
  derivePlanCardType,
  normalizeToDays,
  collectCardTypes,
  CARD_TYPE_ORDER,
  cardTypeBadgeClass,
} from '../planCardType'

describe('normalizeToDays', () => {
  it('converts day/week/month units to total days', () => {
    expect(normalizeToDays(1, 'day')).toBe(1)
    expect(normalizeToDays(7, 'day')).toBe(7)
    expect(normalizeToDays(1, 'week')).toBe(7)
    expect(normalizeToDays(2, 'week')).toBe(14)
    expect(normalizeToDays(1, 'month')).toBe(30)
    expect(normalizeToDays(2, 'month')).toBe(60)
  })

  it('handles plural unit forms', () => {
    expect(normalizeToDays(7, 'days')).toBe(7)
    expect(normalizeToDays(2, 'weeks')).toBe(14)
    expect(normalizeToDays(3, 'months')).toBe(90)
  })

  it('treats unknown / empty unit as days', () => {
    expect(normalizeToDays(5, '')).toBe(5)
    expect(normalizeToDays(5, 'foo')).toBe(5)
    expect(normalizeToDays(5, null)).toBe(5)
    expect(normalizeToDays(5, undefined)).toBe(5)
  })

  it('case-insensitive unit', () => {
    expect(normalizeToDays(1, 'WEEK')).toBe(7)
    expect(normalizeToDays(1, 'Month')).toBe(30)
  })

  it('returns 0 for invalid days', () => {
    expect(normalizeToDays(0, 'day')).toBe(0)
    expect(normalizeToDays(-5, 'day')).toBe(0)
    expect(normalizeToDays(null, 'day')).toBe(0)
    expect(normalizeToDays(undefined, 'day')).toBe(0)
  })
})

describe('derivePlanCardType', () => {
  it('maps standard durations correctly', () => {
    expect(derivePlanCardType(1, 'day')).toBe('daily')
    expect(derivePlanCardType(7, 'day')).toBe('weekly')
    expect(derivePlanCardType(30, 'day')).toBe('monthly')
    expect(derivePlanCardType(90, 'day')).toBe('quarterly')
    expect(derivePlanCardType(365, 'day')).toBe('yearly')
  })

  it('treats equivalent unit/days combos consistently', () => {
    // 7 day == 1 week
    expect(derivePlanCardType(1, 'week')).toBe('weekly')
    // 30 day == 1 month
    expect(derivePlanCardType(1, 'month')).toBe('monthly')
    // 90 day == 3 month
    expect(derivePlanCardType(3, 'month')).toBe('quarterly')
  })

  it('returns custom for non-standard durations', () => {
    expect(derivePlanCardType(15, 'day')).toBe('custom')   // 半月
    expect(derivePlanCardType(2, 'week')).toBe('custom')   // 双周
    expect(derivePlanCardType(2, 'month')).toBe('custom')  // 60 天
    expect(derivePlanCardType(180, 'day')).toBe('custom')  // 半年
    expect(derivePlanCardType(0, 'day')).toBe('custom')    // 无效
    expect(derivePlanCardType(null, null)).toBe('custom')
  })
})

describe('collectCardTypes', () => {
  it('returns present types in canonical order', () => {
    const plans = [
      { validity_days: 30, validity_unit: 'day' },
      { validity_days: 1, validity_unit: 'day' },
      { validity_days: 365, validity_unit: 'day' },
      { validity_days: 15, validity_unit: 'day' }, // custom
    ]
    expect(collectCardTypes(plans)).toEqual(['daily', 'monthly', 'yearly', 'custom'])
  })

  it('deduplicates equivalent representations', () => {
    const plans = [
      { validity_days: 7, validity_unit: 'day' },
      { validity_days: 1, validity_unit: 'week' }, // same as 7 day
    ]
    expect(collectCardTypes(plans)).toEqual(['weekly'])
  })

  it('returns empty for empty input', () => {
    expect(collectCardTypes([])).toEqual([])
  })
})

describe('cardTypeBadgeClass', () => {
  it('returns a non-empty class string for every type', () => {
    for (const t of CARD_TYPE_ORDER) {
      const cls = cardTypeBadgeClass(t)
      expect(cls.length).toBeGreaterThan(0)
      expect(cls).toContain('text-')
    }
  })
})
