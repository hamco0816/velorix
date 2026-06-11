/**
 * 命令面板开闭状态：模块级单例，AppHeader 的触发按钮与 CommandPalette 组件共享。
 */
import { ref } from 'vue'

const isOpen = ref(false)

export function useCommandPalette() {
  const open = (): void => {
    isOpen.value = true
  }
  const close = (): void => {
    isOpen.value = false
  }
  const toggle = (): void => {
    isOpen.value = !isOpen.value
  }
  return { isOpen, open, close, toggle }
}
