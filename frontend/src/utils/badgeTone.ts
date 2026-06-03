// 套餐自定义角标（热门 / 巨量 等）的尊贵预设色板。
// key 与后端 service.allowedPlanBadgeColors 保持一致；新增/删除色调两端同步。
//
// 设计取向：克制实色 + 金属感，营造"尊贵"质感，刻意不用紫色渐变（项目禁止 AI 味渐变）。
// 角标在卡片上是小尺寸 pill，色调仅决定 bg/text/ring/shadow，结构样式由组件自己控制。

export type BadgeTone = 'gold' | 'obsidian' | 'purple' | 'emerald' | 'sapphire' | 'rose'

export const DEFAULT_BADGE_TONE: BadgeTone = 'gold'

// 选择器与预览的展示顺序
export const BADGE_TONE_KEYS: BadgeTone[] = ['gold', 'obsidian', 'purple', 'emerald', 'sapphire', 'rose']

// pill 色调类：bg + 文字 + ring + 浅阴影（含深色模式）。结构类（圆角/字号/ring-1）留在组件里。
const TONE: Record<BadgeTone, string> = {
  gold: 'bg-amber-50 text-amber-800 ring-amber-300/70 shadow-amber-100/60 dark:bg-amber-500/15 dark:text-amber-200 dark:ring-amber-400/30',
  // 黑金至尊：白卡上黑底鎏金字，最强尊贵感
  obsidian: 'bg-gray-900 text-amber-300 ring-amber-400/40 shadow-gray-900/25 dark:bg-black dark:text-amber-300 dark:ring-amber-400/40',
  purple: 'bg-violet-50 text-violet-800 ring-violet-300/70 shadow-violet-100/60 dark:bg-violet-500/15 dark:text-violet-200 dark:ring-violet-400/30',
  emerald: 'bg-emerald-50 text-emerald-800 ring-emerald-300/70 shadow-emerald-100/60 dark:bg-emerald-500/15 dark:text-emerald-200 dark:ring-emerald-400/30',
  sapphire: 'bg-blue-50 text-blue-800 ring-blue-300/70 shadow-blue-100/60 dark:bg-blue-500/15 dark:text-blue-200 dark:ring-blue-400/30',
  rose: 'bg-rose-50 text-rose-800 ring-rose-300/70 shadow-rose-100/60 dark:bg-rose-500/15 dark:text-rose-200 dark:ring-rose-400/30',
}

// 后台色调选择器用的色块（实心小圆点，金属色用极小渐变点缀，不影响整体克制风格）
const SWATCH: Record<BadgeTone, string> = {
  gold: 'bg-gradient-to-br from-amber-300 to-amber-500',
  obsidian: 'bg-gradient-to-br from-gray-700 to-gray-950 ring-1 ring-inset ring-amber-400/50',
  purple: 'bg-violet-500',
  emerald: 'bg-emerald-500',
  sapphire: 'bg-blue-500',
  rose: 'bg-rose-400',
}

function normalizeTone(key: string | null | undefined): BadgeTone {
  const v = (key || '').trim().toLowerCase()
  return (BADGE_TONE_KEYS as string[]).includes(v) ? (v as BadgeTone) : DEFAULT_BADGE_TONE
}

/** 角标 pill 的色调类（传入 plan.badge_color，非法/空回落到默认鎏金）。 */
export function badgeToneClass(key: string | null | undefined): string {
  return TONE[normalizeTone(key)]
}

/** 后台选择器的色块类。 */
export function badgeToneSwatchClass(key: string | null | undefined): string {
  return SWATCH[normalizeTone(key)]
}
