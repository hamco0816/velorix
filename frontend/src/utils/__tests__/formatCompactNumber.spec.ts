import { describe, expect, it } from 'vitest'
import { formatCompactNumber } from '../format'

describe('formatCompactNumber', () => {
  it('formats boundary values with K/M/B/T/P', () => {
    expect(formatCompactNumber(0)).toBe('0')
    expect(formatCompactNumber(999)).toBe('999')
    expect(formatCompactNumber(1000)).toBe('1.0K')
    expect(formatCompactNumber(999999)).toBe('1000.0K')
    expect(formatCompactNumber(1_000_000)).toBe('1.0M')
    expect(formatCompactNumber(1_000_000_000)).toBe('1.0B')
    expect(formatCompactNumber(1_000_000_000_000)).toBe('1.0T')
    expect(formatCompactNumber(1_000_000_000_000_000)).toBe('1.0P')
  })

  it('handles values in T range with reasonable precision', () => {
    expect(formatCompactNumber(2_500_000_000_000)).toBe('2.5T')
    expect(formatCompactNumber(12_345_678_901_234)).toBe('12.3T')
  })

  it('supports disabling billion unit (requests style)', () => {
    expect(formatCompactNumber(1_000_000_000, { allowBillions: false })).toBe('1000.0M')
    // disabling B also disables T/P (老接口语义)
    expect(formatCompactNumber(1_000_000_000_000, { allowBillions: false })).toBe('1000000.0M')
  })

  it('supports maxScale option', () => {
    expect(formatCompactNumber(5_000_000_000_000, { maxScale: 'B' })).toBe('5000.0B')
    expect(formatCompactNumber(5_000_000_000_000, { maxScale: 'T' })).toBe('5.0T')
  })

  it('supports decimals option', () => {
    expect(formatCompactNumber(1_234_567, { decimals: 2 })).toBe('1.23M')
    expect(formatCompactNumber(1_234_567, { decimals: 0 })).toBe('1M')
  })

  it('handles negative values', () => {
    expect(formatCompactNumber(-1_500_000)).toBe('-1.5M')
    expect(formatCompactNumber(-2_000_000_000_000)).toBe('-2.0T')
  })

  it('returns 0 for nullish input', () => {
    expect(formatCompactNumber(null)).toBe('0')
    expect(formatCompactNumber(undefined)).toBe('0')
  })
})
