/**
 * Promise 式确认弹窗组合函数：替代浏览器原生 confirm()，统一走项目 ConfirmDialog 视觉体系。
 *
 * 用法：
 *   const { confirmState, requestConfirm, handleConfirmAccept, handleConfirmCancel } = useConfirm()
 *
 *   模板中放置（每个使用方放一个实例）：
 *   <ConfirmDialog
 *     :show="confirmState.show"
 *     :title="confirmState.title"
 *     :message="confirmState.message"
 *     :confirm-text="confirmState.confirmText"
 *     :cancel-text="confirmState.cancelText"
 *     :danger="confirmState.danger"
 *     @confirm="handleConfirmAccept"
 *     @cancel="handleConfirmCancel"
 *   />
 *
 *   代码中等待用户选择：
 *   const ok = await requestConfirm({ title: t('...'), message: t('...'), danger: true })
 *   if (!ok) return
 */
import { reactive } from 'vue'

export interface ConfirmRequestOptions {
  title: string
  message: string
  confirmText?: string
  cancelText?: string
  danger?: boolean
}

export function useConfirm() {
  const confirmState = reactive({
    show: false,
    title: '',
    message: '',
    confirmText: undefined as string | undefined,
    cancelText: undefined as string | undefined,
    danger: false
  })

  let resolveConfirm: ((confirmed: boolean) => void) | null = null

  // 发起确认请求：弹出对话框并返回 Promise，用户选择后 resolve
  const requestConfirm = (options: ConfirmRequestOptions): Promise<boolean> => {
    // 防御：已有未决确认时先按"取消"结算旧请求，避免 Promise 悬挂
    resolveConfirm?.(false)

    confirmState.title = options.title
    confirmState.message = options.message
    confirmState.confirmText = options.confirmText
    confirmState.cancelText = options.cancelText
    confirmState.danger = options.danger ?? false
    confirmState.show = true

    return new Promise<boolean>((resolve) => {
      resolveConfirm = resolve
    })
  }

  // 结算确认请求并关闭对话框
  const settleConfirm = (confirmed: boolean) => {
    confirmState.show = false
    resolveConfirm?.(confirmed)
    resolveConfirm = null
  }

  const handleConfirmAccept = () => settleConfirm(true)
  const handleConfirmCancel = () => settleConfirm(false)

  return { confirmState, requestConfirm, handleConfirmAccept, handleConfirmCancel }
}
