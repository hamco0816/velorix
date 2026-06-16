/**
 * Centralized platform color definitions.
 *
 * All components that need platform-specific styling should import from here
 * instead of defining their own color mappings.
 */

export type Platform = 'anthropic' | 'openai' | 'antigravity' | 'gemini'

// ── Badge (bg + text + border, for inline badges with border) ───────
const BADGE: Record<Platform, string> = {
  anthropic: 'bg-orange-500/10 text-orange-600 border-orange-500/30 dark:text-orange-400',
  openai: 'bg-green-500/10 text-green-600 border-green-500/30 dark:text-green-400',
  antigravity: 'bg-tea-500/10 text-tea-600 border-tea-500/30 dark:text-tea-400',
  gemini: 'bg-info/10 text-info-deep border-info/30 dark:text-info',
}
const BADGE_DEFAULT = 'bg-gray-500/10 text-gray-600 border-gray-500/30 dark:text-gray-400'

// ── Light badge (softer bg, no border) ──────────────────────────────
const BADGE_LIGHT: Record<Platform, string> = {
  anthropic: 'bg-orange-500/10 text-orange-600 dark:bg-orange-500/10 dark:text-orange-300',
  openai: 'bg-green-500/10 text-green-600 dark:bg-green-500/10 dark:text-green-300',
  antigravity: 'bg-tea-500/10 text-tea-600 dark:bg-tea-500/10 dark:text-tea-300',
  gemini: 'bg-info/10 text-info-deep dark:bg-info/10 dark:text-info',
}

// ── Border ──────────────────────────────────────────────────────────
const BORDER: Record<Platform, string> = {
  anthropic: 'border-orange-500/20 dark:border-orange-500/20',
  openai: 'border-green-500/20 dark:border-green-500/20',
  antigravity: 'border-tea-500/20 dark:border-tea-500/20',
  gemini: 'border-info/20 dark:border-info/20',
}
const BORDER_DEFAULT = 'border-gray-200 dark:border-dark-700'

// ── Accent bar (gradient) ───────────────────────────────────────────
const ACCENT_BAR: Record<Platform, string> = {
  anthropic: 'bg-gradient-to-r from-orange-400 to-orange-500',
  openai: 'bg-gradient-to-r from-emerald-400 to-emerald-500',
  antigravity: 'bg-gradient-to-r from-tea-400 to-tea-500',
  gemini: 'bg-gradient-to-r from-info to-info-deep',
}
const ACCENT_BAR_DEFAULT = 'bg-gradient-to-r from-primary-400 to-primary-500'

// ── Text (price, icon) ─────────────────────────────────────────────
const TEXT: Record<Platform, string> = {
  anthropic: 'text-orange-600 dark:text-orange-400',
  openai: 'text-emerald-600 dark:text-emerald-400',
  antigravity: 'text-tea-600 dark:text-tea-400',
  gemini: 'text-info-deep dark:text-info',
}
const TEXT_DEFAULT = 'text-primary-600 dark:text-primary-400'

// ── Icon (check mark etc.) ──────────────────────────────────────────
const ICON: Record<Platform, string> = {
  anthropic: 'text-orange-500 dark:text-orange-400',
  openai: 'text-emerald-500 dark:text-emerald-400',
  antigravity: 'text-tea-500 dark:text-tea-400',
  gemini: 'text-info dark:text-info',
}
const ICON_DEFAULT = 'text-primary-500 dark:text-primary-400'

// ── Button (solid bg) ───────────────────────────────────────────────
const BUTTON: Record<Platform, string> = {
  anthropic: 'bg-orange-500 text-white hover:bg-orange-600 active:bg-orange-700 dark:bg-orange-500/80 dark:hover:bg-orange-500',
  openai: 'bg-green-600 text-white hover:bg-green-700 active:bg-green-800 dark:bg-green-600/80 dark:hover:bg-green-600',
  antigravity: 'bg-tea-500 text-white hover:bg-tea-600 active:bg-tea-700 dark:bg-tea-500/80 dark:hover:bg-tea-500',
  gemini: 'bg-info text-white hover:bg-info-deep active:bg-info-deep dark:bg-info/80 dark:hover:bg-info',
}
const BUTTON_DEFAULT = 'bg-primary-500 text-white hover:bg-primary-600 dark:bg-primary-600 dark:hover:bg-primary-500'

// ── Discount badge ──────────────────────────────────────────────────
const DISCOUNT: Record<Platform, string> = {
  anthropic: 'bg-orange-100 text-orange-700 dark:bg-orange-900/40 dark:text-orange-300',
  openai: 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/40 dark:text-emerald-300',
  antigravity: 'bg-tea-100 text-tea-700 dark:bg-tea-900/40 dark:text-tea-300',
  gemini: 'bg-info-soft text-info-deep dark:bg-info/20 dark:text-info',
}
const DISCOUNT_DEFAULT = 'bg-danger-soft text-danger dark:bg-danger/20 dark:text-danger'

// ── Header gradient (subscription confirm) ─────────────────────────
const GRADIENT: Record<Platform, string> = {
  anthropic: 'from-orange-500 to-orange-600',
  openai: 'from-emerald-500 to-emerald-600',
  antigravity: 'from-tea-500 to-tea-600',
  gemini: 'from-info to-info-deep',
}
const GRADIENT_DEFAULT = 'from-primary-500 to-primary-600'

// ── Header text (light text on gradient bg) ────────────────────────
const GRADIENT_TEXT: Record<Platform, string> = {
  anthropic: 'text-orange-100',
  openai: 'text-emerald-100',
  antigravity: 'text-tea-100',
  gemini: 'text-info-soft',
}
const GRADIENT_TEXT_DEFAULT = 'text-primary-100'

const GRADIENT_SUBTEXT: Record<Platform, string> = {
  anthropic: 'text-orange-200',
  openai: 'text-emerald-200',
  antigravity: 'text-tea-200',
  gemini: 'text-info-soft',
}
const GRADIENT_SUBTEXT_DEFAULT = 'text-primary-200'

// ── Public API ──────────────────────────────────────────────────────

function isPlatform(p: string): p is Platform {
  return p === 'anthropic' || p === 'openai' || p === 'antigravity' || p === 'gemini'
}

export function platformBadgeClass(p: string): string {
  return isPlatform(p) ? BADGE[p] : BADGE_DEFAULT
}

export function platformBadgeLightClass(p: string): string {
  return isPlatform(p) ? BADGE_LIGHT[p] : BADGE_DEFAULT
}

export function platformBorderClass(p: string): string {
  return isPlatform(p) ? BORDER[p] : BORDER_DEFAULT
}

export function platformAccentBarClass(p: string): string {
  return isPlatform(p) ? ACCENT_BAR[p] : ACCENT_BAR_DEFAULT
}

export function platformTextClass(p: string): string {
  return isPlatform(p) ? TEXT[p] : TEXT_DEFAULT
}

export function platformIconClass(p: string): string {
  return isPlatform(p) ? ICON[p] : ICON_DEFAULT
}

export function platformButtonClass(p: string): string {
  return isPlatform(p) ? BUTTON[p] : BUTTON_DEFAULT
}

export function platformDiscountClass(p: string): string {
  return isPlatform(p) ? DISCOUNT[p] : DISCOUNT_DEFAULT
}

export function platformGradientClass(p: string): string {
  return isPlatform(p) ? GRADIENT[p] : GRADIENT_DEFAULT
}

export function platformGradientTextClass(p: string): string {
  return isPlatform(p) ? GRADIENT_TEXT[p] : GRADIENT_TEXT_DEFAULT
}

export function platformGradientSubtextClass(p: string): string {
  return isPlatform(p) ? GRADIENT_SUBTEXT[p] : GRADIENT_SUBTEXT_DEFAULT
}

export function platformLabel(p: string): string {
  switch (p) {
    case 'anthropic': return 'Anthropic'
    case 'openai': return 'OpenAI'
    case 'antigravity': return 'Antigravity'
    case 'gemini': return 'Gemini'
    default: return p || 'API'
  }
}

// 面向用户的产品品牌名（订阅页分区/筛选用）：用大众熟悉的 GPT / Claude 而非厂商名，
// 与平台分区"GPT 放一块、Claude 放一块"的心智一致。
export function platformProductLabel(p: string): string {
  switch (p) {
    case 'openai': return 'GPT'
    case 'anthropic': return 'Claude'
    case 'gemini': return 'Gemini'
    case 'antigravity': return 'Antigravity'
    default: return p || 'API'
  }
}

// BrandIcon 用的品牌 key；非主流平台返回 null（前端回退到首字母）。
export function platformBrandKey(p: string): 'claude' | 'openai' | 'gemini' | null {
  if (p === 'anthropic') return 'claude'
  if (p === 'openai') return 'openai'
  if (p === 'gemini') return 'gemini'
  return null
}

// 订阅页平台分区/筛选的展示顺序；未列出的平台排在最后（按字母）。
export const PLATFORM_DISPLAY_ORDER: string[] = ['openai', 'anthropic', 'gemini', 'antigravity']

/** 按 PLATFORM_DISPLAY_ORDER 给平台 key 排序的比较函数（未知平台靠后、字母序）。 */
export function comparePlatformOrder(a: string, b: string): number {
  const ia = PLATFORM_DISPLAY_ORDER.indexOf(a)
  const ib = PLATFORM_DISPLAY_ORDER.indexOf(b)
  const ra = ia === -1 ? PLATFORM_DISPLAY_ORDER.length : ia
  const rb = ib === -1 ? PLATFORM_DISPLAY_ORDER.length : ib
  if (ra !== rb) return ra - rb
  return a.localeCompare(b)
}
