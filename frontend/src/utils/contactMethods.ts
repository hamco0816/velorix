import type { ContactMethod, ContactMethodType } from '@/types'

export const CONTACT_METHOD_TYPES: Array<{ value: ContactMethodType; labelZh: string; labelEn: string }> = [
  { value: 'qq', labelZh: 'QQ', labelEn: 'QQ' },
  { value: 'wechat', labelZh: '\u5fae\u4fe1', labelEn: 'WeChat' },
  { value: 'custom', labelZh: '\u81ea\u5b9a\u4e49', labelEn: 'Custom' },
]

export function normalizeContactMethodType(type: unknown): ContactMethodType {
  const normalized = String(type || '').trim().toLowerCase()
  if (normalized === 'qq') return 'qq'
  if (normalized === 'wechat' || normalized === 'weixin' || normalized === 'wx') return 'wechat'
  return 'custom'
}

export function defaultContactMethodLabel(type: unknown, locale = 'zh'): string {
  switch (normalizeContactMethodType(type)) {
    case 'qq':
      return 'QQ'
    case 'wechat':
      return locale.startsWith('zh') ? '\u5fae\u4fe1' : 'WeChat'
    default:
      return locale.startsWith('zh') ? '\u5ba2\u670d' : 'Support'
  }
}

function normalizeContactMethodUrl(raw: unknown): string {
  const value = String(raw || '').trim()
  if (!value) return ''
  try {
    const url = new URL(value)
    return url.protocol === 'http:' || url.protocol === 'https:' ? value : ''
  } catch {
    return ''
  }
}

export function normalizeContactMethods(
  methods: unknown,
  legacyContactInfo = '',
  locale = 'zh',
): ContactMethod[] {
  const source = Array.isArray(methods) ? methods : []
  const normalized = source
    .map((item) => {
      const raw = (item || {}) as Partial<ContactMethod>
      const type = normalizeContactMethodType(raw.type)
      const label = String(raw.label || '').trim() || defaultContactMethodLabel(type, locale)
      const value = String(raw.value || '').trim()
      const url = normalizeContactMethodUrl(raw.url)
      return { type, label, value, ...(url ? { url } : {}) }
    })
    .filter((item) => item.value || item.url)
    .slice(0, 8)

  const legacy = legacyContactInfo.trim()
  if (normalized.length === 0 && legacy) {
    return [{
      type: 'custom',
      label: defaultContactMethodLabel('custom', locale),
      value: legacy,
    }]
  }
  return normalized
}

export function summarizeContactMethods(methods: ContactMethod[], fallback = ''): string {
  if (!methods.length) return fallback.trim()
  return methods
    .map((method) => {
      const value = method.value || method.url || ''
      if (!value) return ''
      return `${method.label || defaultContactMethodLabel(method.type)}: ${value}`
    })
    .filter(Boolean)
    .join(' / ')
}
