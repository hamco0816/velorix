/**
 * 全局明暗主题状态：模块级单例，侧边栏 / 命令面板 / 首页共用同一份 isDark，
 * 任意入口切换主题时所有消费方的图标与文案同步更新。
 */
import { ref } from 'vue'

const THEME_STORAGE_KEY = 'theme'

const isDark = ref(
  typeof document !== 'undefined' && document.documentElement.classList.contains('dark')
)

/** 应用指定主题并持久化到 localStorage */
function applyTheme(dark: boolean): void {
  isDark.value = dark
  document.documentElement.classList.toggle('dark', dark)
  localStorage.setItem(THEME_STORAGE_KEY, dark ? 'dark' : 'light')
}

/** 切换明暗主题 */
function toggleTheme(): void {
  applyTheme(!isDark.value)
}

/** 按「localStorage 记忆 > 系统偏好」初始化主题（应用启动时调用一次即可） */
function initTheme(): void {
  const saved = localStorage.getItem(THEME_STORAGE_KEY)
  const dark =
    saved === 'dark' || (!saved && window.matchMedia('(prefers-color-scheme: dark)').matches)
  isDark.value = dark
  document.documentElement.classList.toggle('dark', dark)
}

export function useTheme() {
  return { isDark, applyTheme, toggleTheme, initTheme }
}
