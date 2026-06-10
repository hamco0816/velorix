<template>
  <Teleport to="body">
    <Transition name="modal">
      <div
        v-if="show"
        class="modal-overlay"
        :style="zIndexStyle"
        :aria-labelledby="dialogId"
        role="dialog"
        aria-modal="true"
        @click.self="handleClose"
      >
        <!-- Modal panel -->
        <div ref="dialogRef" :class="['modal-content', widthClasses]" @click.stop>
          <!-- Header -->
          <div class="modal-header">
            <h3 :id="dialogId" class="modal-title">
              {{ title }}
            </h3>
            <button
              @click="emit('close')"
              class="-mr-1.5 rounded-lg p-1.5 text-gray-400 transition-colors hover:bg-gray-100 hover:text-gray-700 dark:text-dark-500 dark:hover:bg-dark-700 dark:hover:text-dark-200"
              aria-label="Close modal"
            >
              <Icon name="x" size="sm" />
            </button>
          </div>

          <!-- Body -->
          <div class="modal-body">
            <slot></slot>
          </div>

          <!-- Footer -->
          <div v-if="$slots.footer" class="modal-footer">
            <slot name="footer"></slot>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script lang="ts">
// 模块级状态：计数器保证每个弹窗实例的标题 ID 全局唯一；
// 弹窗栈记录当前打开的弹窗顺序，多层叠加时只有最顶层响应 Esc 关闭与 Tab 焦点圈定
let dialogIdCounter = 0
const openDialogStack: symbol[] = []
</script>

<script setup lang="ts">
import { computed, watch, onMounted, onUnmounted, ref, nextTick } from 'vue'
import Icon from '@/components/icons/Icon.vue'

const dialogId = `modal-title-${++dialogIdCounter}`
const dialogStackKey = Symbol('dialog')

// 焦点管理
const dialogRef = ref<HTMLElement | null>(null)
let previousActiveElement: HTMLElement | null = null

type DialogWidth = 'narrow' | 'normal' | 'wide' | 'extra-wide' | 'full'

interface Props {
  show: boolean
  title: string
  width?: DialogWidth
  closeOnEscape?: boolean
  closeOnClickOutside?: boolean
  zIndex?: number
}

interface Emits {
  (e: 'close'): void
}

const props = withDefaults(defineProps<Props>(), {
  width: 'normal',
  closeOnEscape: true,
  closeOnClickOutside: false,
  zIndex: 50
})

const emit = defineEmits<Emits>()

// Custom z-index style (overrides the default z-50 from CSS)
const zIndexStyle = computed(() => {
  return props.zIndex !== 50 ? { zIndex: props.zIndex } : undefined
})

const widthClasses = computed(() => {
  // Width guidance: narrow=confirm/short prompts, normal=standard forms,
  // wide=multi-section forms or rich content, extra-wide=analytics/tables,
  // full=full-screen or very dense layouts.
  const widths: Record<DialogWidth, string> = {
    narrow: 'max-w-md',
    normal: 'max-w-lg',
    wide: 'w-full sm:max-w-2xl md:max-w-3xl lg:max-w-4xl',
    'extra-wide': 'w-full sm:max-w-3xl md:max-w-4xl lg:max-w-5xl xl:max-w-6xl',
    full: 'w-full sm:max-w-4xl md:max-w-5xl lg:max-w-6xl xl:max-w-7xl'
  }
  return widths[props.width]
})

const handleClose = () => {
  if (props.closeOnClickOutside) {
    emit('close')
  }
}

// 判断当前实例是否是弹窗栈顶（多层弹窗叠加时只有顶层响应键盘交互）
const isTopmostDialog = () => openDialogStack[openDialogStack.length - 1] === dialogStackKey

const FOCUSABLE_SELECTOR =
  'button:not([disabled]), [href], input:not([disabled]), select:not([disabled]), textarea:not([disabled]), [tabindex]:not([tabindex="-1"])'

// 取弹窗内当前可见的可聚焦元素（过滤掉 display:none 等不可见节点）
const getFocusableElements = (): HTMLElement[] => {
  if (!dialogRef.value) return []
  return Array.from(dialogRef.value.querySelectorAll<HTMLElement>(FOCUSABLE_SELECTOR)).filter(
    (el) => el.getClientRects().length > 0
  )
}

const handleEscape = (event: KeyboardEvent) => {
  if (props.show && props.closeOnEscape && event.key === 'Escape' && isTopmostDialog()) {
    emit('close')
  }
}

// Tab 焦点圈定：焦点循环保持在弹窗内，不逃逸到背景页面
const handleFocusTrap = (event: KeyboardEvent) => {
  if (!props.show || event.key !== 'Tab' || !isTopmostDialog() || !dialogRef.value) return

  const focusables = getFocusableElements()
  if (focusables.length === 0) {
    event.preventDefault()
    return
  }

  const first = focusables[0]
  const last = focusables[focusables.length - 1]
  const active = document.activeElement as HTMLElement | null
  const activeInside = active ? dialogRef.value.contains(active) : false

  if (event.shiftKey) {
    if (!activeInside || active === first) {
      event.preventDefault()
      last.focus()
    }
  } else if (!activeInside || active === last) {
    event.preventDefault()
    first.focus()
  }
}

const handleKeydown = (event: KeyboardEvent) => {
  handleEscape(event)
  handleFocusTrap(event)
}

// 出栈：从弹窗栈中移除本实例（关闭或卸载时调用）
const removeFromDialogStack = () => {
  const index = openDialogStack.indexOf(dialogStackKey)
  if (index !== -1) {
    openDialogStack.splice(index, 1)
  }
}

// Prevent body scroll when modal is open and manage focus
watch(
  () => props.show,
  async (isOpen) => {
    if (isOpen) {
      // 入栈：标记本弹窗为当前最顶层
      openDialogStack.push(dialogStackKey)
      // 保存当前焦点元素
      previousActiveElement = document.activeElement as HTMLElement
      // 使用CSS类而不是直接操作style,更易于管理多个对话框
      document.body.classList.add('modal-open')

      // 等待DOM更新后设置焦点到对话框
      await nextTick()
      if (dialogRef.value) {
        const firstFocusable = dialogRef.value.querySelector<HTMLElement>(
          'button, [href], input, select, textarea, [tabindex]:not([tabindex="-1"])'
        )
        firstFocusable?.focus()
      }
    } else {
      removeFromDialogStack()
      // 仍有其他弹窗打开时保留滚动锁定
      if (openDialogStack.length === 0) {
        document.body.classList.remove('modal-open')
      }
      // 恢复之前的焦点
      if (previousActiveElement && typeof previousActiveElement.focus === 'function') {
        previousActiveElement.focus()
      }
      previousActiveElement = null
    }
  },
  { immediate: true }
)

onMounted(() => {
  document.addEventListener('keydown', handleKeydown)
})

onUnmounted(() => {
  document.removeEventListener('keydown', handleKeydown)
  removeFromDialogStack()
  // 确保组件卸载时移除滚动锁定
  if (openDialogStack.length === 0) {
    document.body.classList.remove('modal-open')
  }
})
</script>
