<!--
  仪表盘加载骨架屏：用内容形状的脉冲占位替代转圈，感知更快。
  镜像新版布局：账户英雄卡（整行）+ 核心数据（3 卡）+ Token 与性能（4 卡）+ 图表（2 图）+ 最近用量/快捷操作。
-->
<template>
  <div class="space-y-8">
    <!-- 账户英雄卡：整行大卡（左余额 + 右配额面板） -->
    <div class="skel-card">
      <div class="flex flex-col gap-6 p-1 sm:p-2 lg:flex-row lg:items-center lg:gap-10">
        <div class="min-w-0 flex-1">
          <div class="flex items-center gap-3">
            <div class="h-10 w-10 skel-soft rounded-xl"></div>
            <div class="h-3.5 w-24 skel-bar"></div>
          </div>
          <div class="mt-4 h-10 w-48 skel-bar"></div>
          <div class="mt-3 h-3 w-56 skel-soft"></div>
          <div class="mt-5 h-9 w-24 skel-soft rounded-lg"></div>
        </div>
        <div class="w-full rounded-xl border border-gray-100 p-4 dark:border-dark-700/50 lg:w-[340px] lg:flex-shrink-0">
          <div class="h-3 w-16 skel-bar"></div>
          <div class="mt-2 h-3.5 w-28 skel-soft"></div>
          <div class="mt-3 space-y-2.5">
            <div v-for="i in 2" :key="`quota-${i}`" class="flex items-center gap-2.5">
              <div class="h-2.5 w-8 skel-soft"></div>
              <div class="h-1.5 flex-1 skel-soft rounded-full"></div>
              <div class="h-2.5 w-16 skel-soft"></div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 核心数据（3 卡） -->
    <section class="space-y-4">
      <div class="h-4 w-20 skel-bar"></div>
      <div class="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
        <div v-for="i in 3" :key="`core-${i}`" class="skel-card">
          <div class="flex items-center justify-between">
            <div class="h-10 w-10 skel-soft rounded-xl"></div>
            <div class="h-4 w-12 skel-soft"></div>
          </div>
          <div class="mt-4 h-3 w-16 skel-bar"></div>
          <div class="mt-2 h-7 w-28 skel-bar"></div>
          <div class="mt-2.5 h-3 w-20 skel-soft"></div>
        </div>
      </div>
    </section>

    <!-- Token 与性能（4 卡） -->
    <section class="space-y-4">
      <div class="h-4 w-28 skel-bar"></div>
      <div class="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <div v-for="i in 4" :key="`token-${i}`" class="skel-card">
          <div class="flex items-center justify-between">
            <div class="h-10 w-10 skel-soft rounded-xl"></div>
            <div class="h-4 w-12 skel-soft"></div>
          </div>
          <div class="mt-4 h-3 w-16 skel-bar"></div>
          <div class="mt-2 h-7 w-28 skel-bar"></div>
          <div class="mt-2.5 h-3 w-24 skel-soft"></div>
        </div>
      </div>
    </section>

    <!-- 图表区 -->
    <div class="grid grid-cols-1 gap-4 xl:grid-cols-2">
      <div v-for="i in 2" :key="`chart-${i}`" class="skel-card">
        <div class="h-4 w-32 skel-bar"></div>
        <div class="mt-6 h-44 skel-soft rounded-xl"></div>
      </div>
    </div>

    <!-- 最近用量 + 快捷操作 -->
    <div class="grid grid-cols-1 gap-4 lg:grid-cols-3">
      <div class="skel-card lg:col-span-2">
        <div class="h-4 w-28 skel-bar"></div>
        <div class="mt-5 space-y-3">
          <div v-for="i in 4" :key="`row-${i}`" class="flex items-center gap-3">
            <div class="h-9 w-9 skel-soft rounded-xl"></div>
            <div class="h-3 flex-1 skel-soft"></div>
            <div class="h-3 w-16 skel-soft"></div>
          </div>
        </div>
      </div>
      <div class="skel-card lg:col-span-1">
        <div class="h-4 w-24 skel-bar"></div>
        <div class="mt-5 space-y-3">
          <div v-for="i in 3" :key="`qa-${i}`" class="flex items-center gap-3">
            <div class="h-9 w-9 skel-soft rounded-xl"></div>
            <div class="flex-1 space-y-1.5">
              <div class="h-3 w-24 skel-bar"></div>
              <div class="h-2.5 w-32 skel-soft"></div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.skel-card {
  @apply rounded-2xl border border-gray-200/70 bg-white p-5 shadow-card;
  @apply dark:border-dark-700/60 dark:bg-dark-800/40;
}
/* 较实的占位（标题/数字） */
.skel-bar {
  @apply animate-pulse rounded bg-gray-200 dark:bg-dark-700;
}
/* 较浅的占位（图标/副文本） */
.skel-soft {
  @apply animate-pulse rounded bg-gray-100 dark:bg-dark-700/70;
}
</style>
