<template>
  <div class="table-page-layout" :class="{ 'mobile-mode': isMobile }">
    <!-- 顶部 Hero 插槽：列表页可在此放彩色标题区，与全站调性统一 -->
    <div v-if="$slots.hero" class="layout-section-fixed">
      <slot name="hero" />
    </div>

    <!-- 固定区域：操作按钮 -->
    <div v-if="$slots.actions" class="layout-section-fixed">
      <slot name="actions" />
    </div>

    <!-- 固定区域：搜索和过滤器，包成白底 panel 让筛选条与表格视觉成组 -->
    <div v-if="$slots.filters" class="layout-section-fixed">
      <div class="filters-panel">
        <slot name="filters" />
      </div>
    </div>

    <!-- 滚动区域：表格 -->
    <div class="layout-section-scrollable">
      <div class="card table-scroll-container">
        <slot name="table" />
      </div>
    </div>

    <!-- 固定区域：分页器 -->
    <div v-if="$slots.pagination" class="layout-section-fixed">
      <slot name="pagination" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'

const isMobile = ref(false)

const checkMobile = () => {
  isMobile.value = window.innerWidth < 1024
}

onMounted(() => {
  checkMobile()
  window.addEventListener('resize', checkMobile)
})

onUnmounted(() => {
  window.removeEventListener('resize', checkMobile)
})
</script>

<style scoped>
/* 桌面端：Flexbox 布局 */
.table-page-layout {
  @apply flex flex-col gap-6;
  height: calc(100vh - 64px - 4rem); /* 减去 header + lg:p-8 的上下padding */
}

.layout-section-fixed {
  @apply flex-shrink-0;
}

.layout-section-scrollable {
  @apply flex-1 min-h-0 flex flex-col;
}

/* 筛选条 panel：与兑换码 / 文档 panel 同款外观 */
.filters-panel {
  @apply rounded-lg border border-gray-200 bg-white p-4 dark:border-dark-700 dark:bg-dark-900 sm:p-5;
  box-shadow: 0 18px 44px -34px rgb(15 23 42 / 0.55);
}

:global(:root.dark) .filters-panel {
  box-shadow: none;
}

/* 表格滚动容器 - 与全站 panel 统一的"远距离淡阴影" */
.table-scroll-container {
  @apply flex flex-col overflow-hidden h-full bg-white dark:bg-dark-800 rounded-lg border border-gray-200 dark:border-dark-700;
  box-shadow: 0 18px 44px -34px rgb(15 23 42 / 0.55);
}

:global(:root.dark) .table-scroll-container {
  box-shadow: none;
}

.table-scroll-container :deep(.table-wrapper) {
  @apply flex-1 overflow-x-auto overflow-y-auto;
  /* 确保横向滚动条显示在最底部 */
  scrollbar-gutter: stable;
}

.table-scroll-container :deep(table) {
  @apply w-full;
  min-width: max-content; /* 关键：确保表格宽度根据内容撑开，从而触发横向滚动 */
  display: table; /* 使用标准 table 布局以支持 sticky 列 */
}

/* 表头：浅 sky 背景 + 弱化分隔线，与全站浅色头部 panel 风格统一 */
.table-scroll-container :deep(thead) {
  @apply bg-sky-50/50 backdrop-blur-sm;
}

:global(:root.dark) .table-scroll-container :deep(thead) {
  background-color: rgb(14 165 233 / 0.08);
}

.table-scroll-container :deep(tbody) {
  /* 保持默认 table-row-group 显示，不使用 block */
}

.table-scroll-container :deep(th) {
  @apply px-5 py-4 text-left text-sm font-medium text-gray-600 dark:text-dark-300 border-b border-gray-200 dark:border-dark-700;
}

.table-scroll-container :deep(td) {
  @apply px-5 py-4 text-sm text-gray-700 dark:text-gray-300 border-b border-gray-100 dark:border-dark-800;
}

/* 移动端：恢复正常滚动 */
.table-page-layout.mobile-mode .table-scroll-container {
  @apply h-auto overflow-visible border-none shadow-none bg-transparent;
  box-shadow: none;
}

.table-page-layout.mobile-mode .filters-panel {
  @apply border-none p-0 bg-transparent;
  box-shadow: none;
}

.table-page-layout.mobile-mode .layout-section-scrollable {
  @apply flex-none min-h-fit;
}

.table-page-layout.mobile-mode .table-scroll-container :deep(.table-wrapper) {
  @apply overflow-visible;
}

.table-page-layout.mobile-mode .table-scroll-container :deep(table) {
  @apply flex-none;
  display: table;
  min-width: 100%;
}
</style>
