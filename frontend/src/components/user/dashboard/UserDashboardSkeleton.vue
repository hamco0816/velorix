<!--
  仪表盘加载骨架屏：用内容形状的脉冲占位替代转圈，感知更快、更高级。
  镜像 UserDashboardStats（4+4 卡）+ UserDashboardCharts（2 图）+ 最近用量/快捷操作。
-->
<template>
  <div class="space-y-8">
    <!-- 核心数据 -->
    <section class="space-y-4">
      <div class="h-4 w-20 skel-bar"></div>
      <div class="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <div v-for="i in 4" :key="`core-${i}`" class="skel-card">
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

    <!-- Token 与性能 -->
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
            <div class="h-8 w-8 skel-soft rounded-lg"></div>
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
