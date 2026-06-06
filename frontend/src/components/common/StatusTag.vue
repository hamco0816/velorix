<!--
  StatusTag 状态徽章原子：把"状态字符串"统一映射到语义色（成功绿 / 待处理琥珀 /
  失败红 / 信息蓝 / 停用灰），收口全站各页自写的状态色，保证同语义同色。

  用法：
    <StatusTag :status="row.status" :label="t(`invoice.status.${row.status}`)" />
    <StatusTag variant="success" label="已完成" />        显式指定语义
-->
<template>
  <span :class="['badge', variantClass]">
    <slot>{{ label }}</slot>
  </span>
</template>

<script setup lang="ts">
import { computed } from 'vue'

type StatusVariant = 'success' | 'warning' | 'danger' | 'info' | 'neutral'

const props = defineProps<{
  // 显式语义（优先级高于 status 映射）
  variant?: StatusVariant
  // 状态字符串，按下方映射表自动着色
  status?: string
  // 文案（也可用默认插槽）
  label?: string
}>()

// 全站统一的"状态 → 语义色"映射。新增状态在此处补充即可，所有页面自动一致。
const STATUS_VARIANT: Record<string, StatusVariant> = {
  active: 'success', completed: 'success', success: 'success', issued: 'success',
  paid: 'success', normal: 'success', enabled: 'success', online: 'success', valid: 'success',
  pending: 'warning', processing: 'warning', recharging: 'warning', warning: 'warning',
  expiring: 'warning', refunding: 'warning', partially_refunded: 'warning',
  failed: 'danger', rejected: 'danger', error: 'danger', refunded: 'danger',
  expired: 'danger', danger: 'danger', banned: 'danger',
  submitted: 'info', info: 'info', recharged: 'info',
  cancelled: 'neutral', canceled: 'neutral', inactive: 'neutral', disabled: 'neutral',
  closed: 'neutral', offline: 'neutral',
}

const VARIANT_CLASS: Record<StatusVariant, string> = {
  success: 'badge-success',
  warning: 'badge-warning',
  danger: 'badge-danger',
  info: 'badge-info',
  neutral: 'badge-gray',
}

const resolvedVariant = computed<StatusVariant>(
  () => props.variant ?? STATUS_VARIANT[(props.status ?? '').toLowerCase().trim()] ?? 'neutral',
)
const variantClass = computed(() => VARIANT_CLASS[resolvedVariant.value])
</script>
