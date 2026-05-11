import type { ContactMethod, ContactMethodType } from '@/types'

export const CONTACT_METHOD_TYPES: Array<{ value: ContactMethodType; labelZh: string; labelEn: string }> = [
  { value: 'qq', labelZh: 'QQ', labelEn: 'QQ' },
  { value: 'qq_group', labelZh: 'QQ\u7fa4', labelEn: 'QQ Group' },
  { value: 'wechat', labelZh: '\u5fae\u4fe1', labelEn: 'WeChat' },
  { value: 'custom', labelZh: '\u81ea\u5b9a\u4e49', labelEn: 'Custom' },
]

export function normalizeContactMethodType(type: unknown): ContactMethodType {
  const normalized = String(type || '').trim().toLowerCase()
  if (normalized === 'qq') return 'qq'
  if (normalized === 'qq_group' || normalized === 'qqgroup' || normalized === 'qq-group') return 'qq_group'
  if (normalized === 'wechat' || normalized === 'weixin' || normalized === 'wx') return 'wechat'
  return 'custom'
}

export function defaultContactMethodLabel(type: unknown, locale = 'zh'): string {
  switch (normalizeContactMethodType(type)) {
    case 'qq':
      return 'QQ'
    case 'qq_group':
      return locale.startsWith('zh') ? 'QQ\u7fa4' : 'QQ Group'
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

// 仅放行 data:image/* base64 dataURL，避免被注入 javascript:/data:text/html 等
function sanitizeContactImageData(raw: unknown): string {
  const value = String(raw || '').trim()
  if (!value) return ''
  if (!/^data:image\/(png|jpeg|jpg|webp|gif|svg\+xml);base64,/i.test(value)) return ''
  // 单图 64KB 上限（admin 端会压缩，跟后端 maxContactImageDataLen 对齐）
  if (value.length > 65536) return ''
  return value
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
      const imageData = sanitizeContactImageData(raw.image_data)
      return {
        type,
        label,
        value,
        ...(url ? { url } : {}),
        ...(imageData ? { image_data: imageData } : {}),
      }
    })
    // qq_group 通常只贴二维码不留群号，允许 image_data only 的条目存在
    .filter((item) => item.value || item.url || item.image_data)
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
