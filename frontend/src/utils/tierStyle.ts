// 套餐档位样式：升级阶梯预设，越高越豪华。key 与后端 service.allowedPlanTierStyles 一致。
// 每档 = 档位名配色 + 专属图标 + 高档列的点缀 + 高档 CTA 配色。

export type TierStyle = 'basic' | 'standard' | 'advanced' | 'flagship' | 'luxury' | 'supreme'

export const DEFAULT_TIER_STYLE: TierStyle = 'basic'

// 选择器/展示顺序（由低到高）
export const TIER_STYLE_KEYS: TierStyle[] = ['basic', 'standard', 'advanced', 'flagship', 'luxury', 'supreme']

// 档位图标限定在这几个（均已在 Icon.vue 注册），便于与 Icon 的 name 类型对齐
type TierIcon = 'bolt' | 'cube' | 'shield' | 'sparkles' | 'crown' | 'diamond'

interface TierTheme {
  /** 档位名前的图标；null = 不显示图标 */
  icon: TierIcon | null
  /** 档位名文字配色 */
  nameClass: string
  /** 图标配色 */
  iconClass: string
  /** 整列点缀底色（仅高档非空，做视觉突出） */
  columnClass: string
  /** 该档 CTA 按钮配色（空 = 用组件默认黑色按钮） */
  ctaClass: string
  /** 选择器小色块 */
  swatchClass: string
}

const THEMES: Record<TierStyle, TierTheme> = {
  basic: {
    icon: 'bolt',
    nameClass: 'text-gray-600 dark:text-dark-200',
    iconClass: 'text-gray-400 dark:text-dark-400',
    columnClass: '',
    ctaClass: '',
    swatchClass: 'bg-gray-300',
  },
  standard: {
    icon: 'cube',
    nameClass: 'text-slate-600 dark:text-slate-300',
    iconClass: 'text-slate-400 dark:text-slate-400',
    columnClass: '',
    ctaClass: '',
    swatchClass: 'bg-slate-400',
  },
  advanced: {
    icon: 'shield',
    nameClass: 'text-emerald-700 dark:text-emerald-300',
    iconClass: 'text-emerald-500 dark:text-emerald-400',
    columnClass: '',
    ctaClass: '',
    swatchClass: 'bg-emerald-500',
  },
  flagship: {
    icon: 'sparkles',
    nameClass: 'text-blue-700 dark:text-blue-300',
    iconClass: 'text-blue-500 dark:text-blue-400',
    columnClass: '',
    ctaClass: '',
    swatchClass: 'bg-blue-500',
  },
  // 豪华：鎏金 + 皇冠 + 淡金列底 + 金色 CTA
  luxury: {
    icon: 'crown',
    nameClass: 'text-amber-700 dark:text-amber-300',
    iconClass: 'text-amber-500 dark:text-amber-400',
    columnClass: 'bg-amber-50/50 dark:bg-amber-500/[0.06]',
    ctaClass: 'bg-amber-500 text-white shadow-sm shadow-amber-500/25 hover:bg-amber-600',
    swatchClass: 'bg-gradient-to-br from-amber-300 to-amber-500',
  },
  // 至尊：黑金 + 钻石 + 黑底金字 CTA
  supreme: {
    icon: 'diamond',
    nameClass: 'text-gray-900 dark:text-white',
    iconClass: 'text-amber-500 dark:text-amber-400',
    columnClass: 'bg-gray-50 dark:bg-white/[0.04]',
    ctaClass: 'bg-gray-900 text-amber-300 shadow-sm hover:bg-gray-800 dark:bg-black dark:text-amber-300',
    swatchClass: 'bg-gradient-to-br from-gray-700 to-gray-950 ring-1 ring-inset ring-amber-400/50',
  },
}

function normalize(style: string | null | undefined): TierStyle {
  const v = (style || '').trim().toLowerCase()
  return (TIER_STYLE_KEYS as string[]).includes(v) ? (v as TierStyle) : DEFAULT_TIER_STYLE
}

export function tierTheme(style: string | null | undefined): TierTheme {
  return THEMES[normalize(style)]
}

export function tierSwatchClass(style: string | null | undefined): string {
  return THEMES[normalize(style)].swatchClass
}
