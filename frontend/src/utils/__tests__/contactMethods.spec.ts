import { describe, expect, it } from 'vitest'
import {
  defaultContactMethodLabel,
  normalizeContactMethods,
  summarizeContactMethods,
} from '@/utils/contactMethods'

describe('contactMethods', () => {
  it('normalizes known contact types and filters unsafe URLs', () => {
    const methods = normalizeContactMethods([
      { type: 'qq', value: ' 123456 ', url: 'javascript:alert(1)' },
      { type: 'wx', value: 'velorix', url: 'https://example.com/wechat' },
      { type: 'custom', label: 'Telegram', value: '', url: 'ftp://example.com' },
    ], '', 'zh-CN')

    expect(methods).toEqual([
      { type: 'qq', label: 'QQ', value: '123456' },
      {
        type: 'wechat',
        label: defaultContactMethodLabel('wechat', 'zh-CN'),
        value: 'velorix',
        url: 'https://example.com/wechat',
      },
    ])
  })

  it('keeps legacy contact info readable', () => {
    const methods = normalizeContactMethods([], 'support@example.com', 'en')

    expect(methods).toEqual([
      { type: 'custom', label: 'Support', value: 'support@example.com' },
    ])
    expect(summarizeContactMethods(methods)).toBe('Support: support@example.com')
  })
})
