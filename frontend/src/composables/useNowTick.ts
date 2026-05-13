import { ref, onMounted, onBeforeUnmount, type Ref } from 'vue'

/**
 * 共享 1Hz Date 时钟：返回随时间自动 +1 秒的 ref，组件挂载时启动、卸载时停止。
 *
 * 多个组件实例各调用一次，每个实例独立一个 timer，但实际开销忽略不计（setInterval 1000ms）。
 * 用于需要"每秒重新计算"的 UI（倒计时、相对时间等），避免每个调用方都自己写 setInterval。
 */
export function useNowTick(): Ref<number> {
  const now = ref(Date.now())
  let timer: number | undefined

  onMounted(() => {
    timer = setInterval(() => {
      now.value = Date.now()
    }, 1000) as unknown as number
  })

  onBeforeUnmount(() => {
    if (timer !== undefined) clearInterval(timer)
  })

  return now
}
