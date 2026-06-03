// 套餐自定义角标（热门 / 巨量 等）的尊贵预设色板。
// key 与后端 service.allowedPlanBadgeColors 保持一致；新增/删除色调两端同步。
//
// 设计取向：克制实色 + 金属感，营造"尊贵"质感，刻意不用紫色渐变（项目禁止 AI 味渐变）。
// 角标在卡片上是小尺寸 pill，色调仅决定 bg/text/ring/shadow，结构样式由组件自己控制。

export type BadgeTone = 'gold' | 'obsidian' | 'purple' | 'emerald' | 'sapphire' | 'rose'

export const DEFAULT_BADGE_TONE: BadgeTone = 'gold'

// 选择器与预览的展示顺序
export const BADGE_TONE_KEYS: BadgeTone[] = ['gold', 'obsidian', 'purple', 'emerald', 'sapphire', 'rose']

// pill 色调类：渐变填充 + 文字色 + 内描边（含深浅模式自适应，渐变在两种模式都够鲜明，不再分浅色态）。
// 结构类（圆角/字号/内边距/阴影）留在组件里，组件不再加 ring-1，描边由这里统一负责。
const TONE: Record<BadgeTone, string> = {
  // 香槟金：亮金渐变 + 深棕字，金属"刻字"质感
  gold: 'bg-gradient-to-r from-amber-300 via-yellow-400 to-amber-500 text-amber-950 ring-1 ring-inset ring-white/30',
  // 黑金至尊：黑色渐变 + 鎏金字，最强尊贵感
  obsidian: 'bg-gradient-to-r from-gray-800 to-gray-950 text-amber-300 ring-1 ring-inset ring-amber-400/30',
  // 帝王紫：克制深紫渐变（不碰霓虹紫）
  purple: 'bg-gradient-to-r from-violet-500 to-purple-600 text-white ring-1 ring-inset ring-white/20',
  emerald: 'bg-gradient-to-r from-emerald-400 to-teal-500 text-white ring-1 ring-inset ring-white/20',
  sapphire: 'bg-gradient-to-r from-blue-500 to-indigo-600 text-white ring-1 ring-inset ring-white/20',
  // 玫瑰金：玫粉渐变 + 深玫字
  rose: 'bg-gradient-to-r from-rose-300 to-pink-400 text-rose-950 ring-1 ring-inset ring-white/30',
}

// 后台色调选择器用的色块（与角标同款渐变的小圆点）
const SWATCH: Record<BadgeTone, string> = {
  gold: 'bg-gradient-to-br from-amber-300 to-amber-500',
  obsidian: 'bg-gradient-to-br from-gray-700 to-gray-950 ring-1 ring-inset ring-amber-400/50',
  purple: 'bg-gradient-to-br from-violet-500 to-purple-600',
  emerald: 'bg-gradient-to-br from-emerald-400 to-teal-500',
  sapphire: 'bg-gradient-to-br from-blue-500 to-indigo-600',
  rose: 'bg-gradient-to-br from-rose-300 to-pink-400',
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
