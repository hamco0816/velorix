<template>
  <AppLayout wide>
    <div class="space-y-5">
      <!-- 顶部 toolbar：概览 chip + 搜索 + 刷新 + 汇率切换 -->
      <div class="flex flex-col gap-3 lg:flex-row lg:items-center lg:justify-between">
        <div class="flex flex-wrap items-center gap-2">
          <span
            class="inline-flex items-center gap-1.5 rounded-full bg-gray-50 px-2.5 py-1 text-xs font-medium text-gray-600 ring-1 ring-inset ring-gray-200/70 dark:bg-dark-800/60 dark:text-dark-200 dark:ring-dark-700/60"
          >
            <Icon name="cube" size="xs" class="text-gray-400" />
            <span class="tabular-nums">{{ filteredModels.length }}</span>
            <span class="text-gray-400 dark:text-dark-400">{{ t('pricing.statModels') }}</span>
          </span>
          <span
            v-if="availableGroups.length > 0"
            class="inline-flex items-center gap-1.5 rounded-full bg-violet-50 px-2.5 py-1 text-xs font-medium text-violet-700 ring-1 ring-inset ring-violet-200/70 dark:bg-violet-500/15 dark:text-violet-300 dark:ring-violet-500/30"
          >
            <Icon name="shield" size="xs" />
            <span class="tabular-nums">{{ availableGroups.length }}</span>
            <span class="opacity-70">{{ t('pricing.statGroups') }}</span>
          </span>
          <span
            v-if="selectedGroup"
            class="inline-flex items-center gap-1.5 rounded-full bg-brand-50 px-2.5 py-1 text-xs font-medium text-brand-700 ring-1 ring-inset ring-brand-200/70 dark:bg-brand-500/15 dark:text-brand-300 dark:ring-brand-500/30"
          >
            <Icon name="bolt" size="xs" />
            <span class="opacity-70">{{ t('pricing.appliedRate') }}</span>
            <span class="tabular-nums">×{{ effectiveRate(selectedGroup).toFixed(2) }}</span>
          </span>
          <!-- 限时活动 chip：当选中分组在 promo 窗口内，显示活动名 + 倒计时 -->
          <span
            v-if="selectedGroup && promoActive(selectedGroup)"
            class="inline-flex items-center gap-1.5 rounded-full bg-rose-50 px-2.5 py-1 text-xs font-medium text-rose-700 ring-1 ring-inset ring-rose-200/70 dark:bg-rose-500/15 dark:text-rose-300 dark:ring-rose-500/30"
          >
            <Icon name="fire" size="xs" :stroke-width="2.2" />
            <span v-if="selectedGroup.promo_label">{{ selectedGroup.promo_label }}</span>
            <span v-else>{{ t('pricing.promoActive') }}</span>
            <span class="tabular-nums opacity-90">{{ countdownLabel(selectedGroup) }}</span>
          </span>
        </div>
        <div class="flex flex-wrap items-center gap-2">
          <!-- 货币切换：USD 标准价 / CNY 换算后价 -->
          <div class="inline-flex rounded-lg border border-gray-200/70 bg-white p-0.5 text-xs dark:border-dark-700/60 dark:bg-dark-800/40">
            <button
              type="button"
              class="rounded-md px-2.5 py-1 font-medium transition-colors"
              :class="
                currency === 'USD'
                  ? 'bg-gray-900 text-white dark:bg-white dark:text-gray-900'
                  : 'text-gray-500 hover:text-gray-900 dark:text-dark-400 dark:hover:text-white'
              "
              @click="currency = 'USD'"
              :title="t('pricing.usdHint', { rate: usdToCny.toFixed(2) })"
            >
              USD
            </button>
            <button
              type="button"
              class="rounded-md px-2.5 py-1 font-medium transition-colors"
              :class="
                currency === 'CNY'
                  ? 'bg-gray-900 text-white dark:bg-white dark:text-gray-900'
                  : 'text-gray-500 hover:text-gray-900 dark:text-dark-400 dark:hover:text-white'
              "
              @click="currency = 'CNY'"
              :title="t('pricing.cnyHint2')"
            >
              CNY
            </button>
          </div>
          <div class="relative w-full sm:w-72">
            <Icon
              name="search"
              size="sm"
              class="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-gray-400 dark:text-gray-500"
            />
            <input
              v-model="searchQuery"
              type="text"
              :placeholder="t('pricing.searchPlaceholder')"
              class="input pl-9 text-sm"
            />
          </div>
          <button
            type="button"
            class="btn btn-secondary btn-sm"
            :disabled="loading"
            :title="t('common.refresh')"
            :aria-label="t('common.refresh')"
            @click="loadAll"
          >
            <Icon name="refresh" size="sm" :class="loading ? 'animate-spin' : ''" />
          </button>
        </div>
      </div>

      <!-- 移动端：筛选面板折叠按钮（lg 以下显示）-->
      <button
        type="button"
        class="flex w-full items-center justify-between rounded-xl border border-gray-200/70 bg-white px-4 py-2.5 text-sm font-medium text-gray-700 transition-colors hover:border-gray-300 dark:border-dark-700/60 dark:bg-dark-800/40 dark:text-dark-200 lg:hidden"
        @click="mobileFiltersOpen = !mobileFiltersOpen"
      >
        <span class="inline-flex items-center gap-2">
          <Icon name="filter" size="sm" class="text-gray-400" />
          {{ t('pricing.filterPanelToggle') }}
          <span
            v-if="activeFilterCount > 0"
            class="inline-flex items-center justify-center rounded-full bg-brand-100 px-1.5 py-0.5 text-[10px] font-semibold tabular-nums text-brand-700 dark:bg-brand-500/20 dark:text-brand-300"
          >
            {{ activeFilterCount }}
          </span>
        </span>
        <Icon name="chevronDown" size="sm" :class="['transition-transform', mobileFiltersOpen ? 'rotate-180' : '']" />
      </button>

      <!-- 主体：左侧筛选 + 右侧网格 -->
      <div class="grid gap-5 lg:grid-cols-[300px_minmax(0,1fr)]">
        <!-- 左侧筛选 panel：sticky 跟随滚动；字号 + 间距增大；移动端默认折叠 -->
        <aside
          class="lg:sticky lg:top-20 lg:self-start"
          :class="mobileFiltersOpen ? '' : 'hidden lg:block'"
        >
          <div class="surface-card card-emerald space-y-5 p-5">
            <!-- 我可用的分组 -->
            <section>
              <div class="mb-2.5 flex items-center justify-between">
                <h3 class="text-sm font-semibold tracking-tight text-gray-900 dark:text-white">
                  {{ t('pricing.filterGroups') }}
                </h3>
                <button
                  v-if="selectedGroupId !== null"
                  type="button"
                  class="text-xs font-medium text-brand-600 hover:text-brand-700 dark:text-brand-400"
                  @click="selectedGroupId = null"
                >
                  {{ t('common.reset') }}
                </button>
              </div>
              <div class="flex flex-col gap-1.5">
                <!-- "全部分组"：选中时显示原价（×1） -->
                <button
                  type="button"
                  class="group flex items-center justify-between rounded-lg border px-3 py-2 text-left text-sm transition-colors"
                  :class="
                    selectedGroupId === null
                      ? 'border-brand-300 bg-brand-50 text-brand-700 dark:border-brand-500/40 dark:bg-brand-500/10 dark:text-brand-300'
                      : 'border-gray-200/70 bg-white text-gray-700 hover:border-gray-300 dark:border-dark-700/60 dark:bg-dark-800/40 dark:text-dark-200 dark:hover:border-dark-500'
                  "
                  @click="selectedGroupId = null"
                >
                  <span class="font-medium">{{ t('pricing.allGroups') }}</span>
                  <span class="text-xs opacity-70 tabular-nums">{{ availableGroups.length }}</span>
                </button>

                <button
                  v-for="group in availableGroups"
                  :key="group.id"
                  type="button"
                  class="group relative flex flex-col gap-1 rounded-xl border px-3 py-2 text-left text-sm transition-colors"
                  :class="
                    selectedGroupId === group.id
                      ? 'border-brand-300 bg-brand-50 text-brand-700 dark:border-brand-500/40 dark:bg-brand-500/10 dark:text-brand-300'
                      : 'border-gray-200/70 bg-white text-gray-700 hover:border-gray-300 dark:border-dark-700/60 dark:bg-dark-800/40 dark:text-dark-200 dark:hover:border-dark-500'
                  "
                  :title="group.name"
                  @click="selectedGroupId = group.id"
                >
                  <!-- 主行：标识图标 + 分组名（可换行）+ 右侧倍率 chip -->
                  <div class="flex items-center justify-between gap-2">
                    <div class="flex min-w-0 flex-1 items-center gap-1.5">
                      <Icon
                        v-if="group.is_exclusive"
                        name="shield"
                        size="xs"
                        class="shrink-0 text-violet-500"
                        :title="t('pricing.exclusiveGroup')"
                      />
                      <PlatformIcon :platform="group.platform as GroupPlatform" size="xs" class="shrink-0" />
                      <span class="break-all font-medium leading-tight">{{ group.name }}</span>
                    </div>
                    <span
                      class="inline-flex shrink-0 items-center rounded-full px-2 py-0.5 text-[11px] font-semibold tabular-nums ring-1 ring-inset"
                      :class="
                        promoActive(group)
                          ? 'bg-rose-50 text-rose-700 ring-rose-200/70 dark:bg-rose-500/15 dark:text-rose-300 dark:ring-rose-500/30'
                          : selectedGroupId === group.id
                            ? 'bg-brand-100 text-brand-700 ring-brand-200/70 dark:bg-brand-500/20 dark:text-brand-200 dark:ring-brand-500/30'
                            : 'bg-gray-100 text-gray-600 ring-gray-200/70 dark:bg-dark-700/40 dark:text-dark-200 dark:ring-dark-600/60'
                      "
                    >
                      ×{{ effectiveRate(group).toFixed(2) }}
                    </span>
                  </div>

                  <!-- 副行：仅在 promo 激活时显示 — fire 图标 + 倒计时 + 原价划线 -->
                  <div
                    v-if="promoActive(group)"
                    class="flex items-center justify-between gap-2 text-[10px] tabular-nums"
                  >
                    <span class="inline-flex items-center gap-1 text-rose-600 dark:text-rose-400">
                      <Icon name="fire" size="xs" class="animate-pulse" />
                      <span class="font-medium">{{ countdownLabel(group) }}</span>
                    </span>
                    <span class="text-gray-400 line-through dark:text-dark-500">
                      ×{{ baseRate(group).toFixed(2) }}
                    </span>
                  </div>
                </button>

                <p
                  v-if="!loading && availableGroups.length === 0"
                  class="text-xs text-gray-500 dark:text-dark-400"
                >
                  {{ t('pricing.noGroups') }}
                </p>
              </div>
            </section>

            <!-- 平台筛选 -->
            <section v-if="availablePlatforms.length > 1">
              <h3 class="mb-2.5 text-sm font-semibold tracking-tight text-gray-900 dark:text-white">
                {{ t('pricing.filterPlatform') }}
              </h3>
              <div class="flex flex-wrap gap-1.5">
                <button
                  type="button"
                  class="rounded-full border px-3 py-1 text-xs font-medium transition-colors"
                  :class="
                    selectedPlatform === null
                      ? 'border-gray-900 bg-gray-900 text-white dark:border-white dark:bg-white dark:text-gray-900'
                      : 'border-gray-200/70 bg-white text-gray-600 hover:border-gray-300 dark:border-dark-700/60 dark:bg-dark-800/40 dark:text-dark-200'
                  "
                  @click="selectedPlatform = null"
                >
                  {{ t('common.all') }}
                </button>
                <button
                  v-for="p in availablePlatforms"
                  :key="p"
                  type="button"
                  class="inline-flex items-center gap-1 rounded-full border px-3 py-1 text-xs font-medium transition-colors"
                  :class="
                    selectedPlatform === p
                      ? 'border-gray-900 bg-gray-900 text-white dark:border-white dark:bg-white dark:text-gray-900'
                      : 'border-gray-200/70 bg-white text-gray-600 hover:border-gray-300 dark:border-dark-700/60 dark:bg-dark-800/40 dark:text-dark-200'
                  "
                  @click="selectedPlatform = p"
                >
                  <PlatformIcon :platform="p as GroupPlatform" size="xs" />
                  {{ p }}
                </button>
              </div>
            </section>

            <!-- 计费类型筛选 -->
            <section v-if="availableBillingModes.length > 1">
              <h3 class="mb-2.5 text-sm font-semibold tracking-tight text-gray-900 dark:text-white">
                {{ t('pricing.filterBillingMode') }}
              </h3>
              <div class="flex flex-wrap gap-1.5">
                <button
                  type="button"
                  class="rounded-full border px-3 py-1 text-xs font-medium transition-colors"
                  :class="
                    selectedBillingMode === null
                      ? 'border-gray-900 bg-gray-900 text-white dark:border-white dark:bg-white dark:text-gray-900'
                      : 'border-gray-200/70 bg-white text-gray-600 hover:border-gray-300 dark:border-dark-700/60 dark:bg-dark-800/40 dark:text-dark-200'
                  "
                  @click="selectedBillingMode = null"
                >
                  {{ t('common.all') }}
                </button>
                <button
                  v-for="m in availableBillingModes"
                  :key="m"
                  type="button"
                  class="rounded-full border px-3 py-1 text-xs font-medium transition-colors"
                  :class="
                    selectedBillingMode === m
                      ? 'border-gray-900 bg-gray-900 text-white dark:border-white dark:bg-white dark:text-gray-900'
                      : 'border-gray-200/70 bg-white text-gray-600 hover:border-gray-300 dark:border-dark-700/60 dark:bg-dark-800/40 dark:text-dark-200'
                  "
                  @click="selectedBillingMode = m"
                >
                  {{ billingModeLabel(m) }}
                </button>
              </div>
            </section>
          </div>
        </aside>

        <!-- 右侧：模型卡片网格 -->
        <main>
          <!-- Loading -->
          <div v-if="loading" class="flex items-center justify-center py-16">
            <LoadingSpinner size="md" />
          </div>

          <!-- Empty -->
          <EmptyState
            v-else-if="filteredModels.length === 0"
            variant="emerald"
            :description="searchQuery || selectedPlatform || selectedBillingMode ? t('pricing.emptyFiltered') : t('pricing.empty')"
          >
            <template #icon>
              <Icon name="search" class="empty-state-icon" />
            </template>
          </EmptyState>

          <!-- Grid -->
          <div v-else class="grid gap-3 sm:grid-cols-2 sm:gap-4 xl:grid-cols-3">
            <div
              v-for="model in filteredModels"
              :key="`${model.platform}-${model.name}`"
              class="flex flex-col rounded-2xl border border-gray-200/70 bg-white p-4 shadow-card transition-all duration-200 hover:-translate-y-0.5 hover:border-gray-300 hover:shadow-card-hover sm:p-5 dark:border-dark-700/60 dark:bg-dark-800/40 dark:hover:border-dark-500"
            >
              <!-- Header：平台 icon + 模型名 + 计费 chip -->
              <div class="flex items-start gap-3">
                <div
                  :class="[
                    'flex h-10 w-10 shrink-0 items-center justify-center rounded-xl ring-1 ring-inset',
                    platformIconBg(model.platform),
                  ]"
                >
                  <PlatformIcon :platform="model.platform as GroupPlatform" size="md" />
                </div>
                <div class="min-w-0 flex-1">
                  <p class="truncate text-[15px] font-semibold tracking-tight text-gray-900 dark:text-white" :title="model.name">
                    {{ model.name }}
                  </p>
                  <div class="mt-1 flex flex-wrap items-center gap-1.5">
                    <span
                      :class="[
                        'inline-flex items-center rounded-full px-2 py-0.5 text-[11px] font-medium ring-1 ring-inset',
                        billingModeChipClass(model.billingMode),
                      ]"
                    >
                      {{ billingModeLabel(model.billingMode) }}
                    </span>
                    <span
                      v-if="!model.fromChannel"
                      class="inline-flex items-center gap-1 rounded-full bg-gray-50 px-2 py-0.5 text-[11px] font-medium text-gray-500 ring-1 ring-inset ring-gray-200/70 dark:bg-dark-800/60 dark:text-dark-200 dark:ring-dark-700/60"
                      :title="t('pricing.standardModelHint')"
                    >
                      <Icon name="database" size="xs" />
                      {{ t('pricing.standardModel') }}
                    </span>
                  </div>
                </div>
              </div>

              <!-- 价格主体：按计费模式分支 -->
              <div class="mt-4 flex-1 space-y-2 text-sm">
                <template v-if="!model.hasPricing">
                  <p class="text-gray-400 dark:text-dark-500">{{ t('pricing.noPricing') }}</p>
                </template>

                <!-- Per-request 计费 -->
                <template v-else-if="model.billingMode === 'per_request'">
                  <div class="flex items-center justify-between">
                    <span class="text-gray-500 dark:text-dark-400">{{ t('pricing.perRequest') }}</span>
                    <span class="text-base font-semibold tabular-nums text-gray-900 dark:text-white">
                      {{ formatPrice(applyRate(model.perRequestPrice)) }}
                    </span>
                  </div>
                </template>

                <!-- Image 计费：用 image_rate_multiplier 而不是 rate_multiplier；如果分组有固定 per-image 价（1K/2K/4K），优先用那些 -->
                <template v-else-if="model.billingMode === 'image'">
                  <!-- 分支 A：分组配了固定每图价 → 展示三档（1K/2K/4K） -->
                  <template v-if="hasGroupImagePrices">
                    <div v-if="selectedGroupImagePrices.k1 != null" class="flex items-center justify-between">
                      <span class="text-gray-500 dark:text-dark-400">{{ t('pricing.imagePer1K') }}</span>
                      <span class="text-base font-semibold tabular-nums text-gray-900 dark:text-white">{{ formatPrice(applyImageRate(selectedGroupImagePrices.k1)) }}</span>
                    </div>
                    <div v-if="selectedGroupImagePrices.k2 != null" class="flex items-center justify-between">
                      <span class="text-gray-500 dark:text-dark-400">{{ t('pricing.imagePer2K') }}</span>
                      <span class="text-base font-semibold tabular-nums text-gray-900 dark:text-white">{{ formatPrice(applyImageRate(selectedGroupImagePrices.k2)) }}</span>
                    </div>
                    <div v-if="selectedGroupImagePrices.k4 != null" class="flex items-center justify-between">
                      <span class="text-gray-500 dark:text-dark-400">{{ t('pricing.imagePer4K') }}</span>
                      <span class="text-base font-semibold tabular-nums text-gray-900 dark:text-white">{{ formatPrice(applyImageRate(selectedGroupImagePrices.k4)) }}</span>
                    </div>
                  </template>
                  <!-- 分支 B：用 LiteLLM 默认价（per-image 或 per-image-token）× image_rate_multiplier -->
                  <template v-else>
                    <div class="flex items-center justify-between">
                      <span class="text-gray-500 dark:text-dark-400">
                        {{ t('pricing.imageOutput') }}
                        <span v-if="model.imageOutputPriceUnit === 'per_image'" class="ml-1 text-xs text-gray-400">{{ t('pricing.imageOutputUnitPerImage') }}</span>
                        <span v-else-if="model.imageOutputPriceUnit === 'per_image_token'" class="ml-1 text-xs text-gray-400">{{ t('pricing.imageOutputUnitPerToken') }}</span>
                      </span>
                      <span class="text-base font-semibold tabular-nums text-gray-900 dark:text-white">
                        <template v-if="model.imageOutputPriceUnit === 'per_image_token'">
                          {{ formatPrice(applyImageRate(model.imageOutputPrice != null ? model.imageOutputPrice * 1_000_000 : null)) }}
                        </template>
                        <template v-else>
                          {{ formatPrice(applyImageRate(model.imageOutputPrice)) }}
                        </template>
                      </span>
                    </div>
                    <p v-if="selectedGroupImageIndependent" class="text-[10px] italic text-gray-400 dark:text-dark-500">
                      {{ t('pricing.imageRateNote', { rate: selectedGroupImageRateLabel }) }}
                    </p>
                  </template>
                </template>

                <!-- Token 计费（默认） -->
                <template v-else>
                  <div class="flex items-center justify-between">
                    <span class="inline-flex items-center gap-1.5 text-gray-500 dark:text-dark-400">
                      <Icon name="arrowDown" size="sm" class="text-emerald-500" />
                      {{ t('pricing.input') }}
                    </span>
                    <span class="inline-flex items-center gap-1.5 tabular-nums">
                      <span v-if="selectedGroupPromoActive" class="text-xs text-gray-400 line-through">
                        {{ formatPricePerM(applyBaseRate(model.inputPrice)) }}
                      </span>
                      <span
                        class="font-semibold"
                        :class="selectedGroupPromoActive ? 'text-rose-600 dark:text-rose-400' : 'text-gray-900 dark:text-white'"
                      >
                        {{ formatPricePerM(applyRate(model.inputPrice)) }}
                      </span>
                    </span>
                  </div>
                  <div class="flex items-center justify-between">
                    <span class="inline-flex items-center gap-1.5 text-gray-500 dark:text-dark-400">
                      <Icon name="arrowUp" size="sm" class="text-violet-500" />
                      {{ t('pricing.output') }}
                    </span>
                    <span class="inline-flex items-center gap-1.5 tabular-nums">
                      <span v-if="selectedGroupPromoActive" class="text-xs text-gray-400 line-through">
                        {{ formatPricePerM(applyBaseRate(model.outputPrice)) }}
                      </span>
                      <span
                        class="font-semibold"
                        :class="selectedGroupPromoActive ? 'text-rose-600 dark:text-rose-400' : 'text-gray-900 dark:text-white'"
                      >
                        {{ formatPricePerM(applyRate(model.outputPrice)) }}
                      </span>
                    </span>
                  </div>
                  <div v-if="model.cacheReadPrice != null && model.cacheReadPrice > 0" class="flex items-center justify-between">
                    <span class="inline-flex items-center gap-1.5 text-gray-500 dark:text-dark-400">
                      <Icon name="inbox" size="sm" class="text-sky-500" />
                      {{ t('pricing.cacheRead') }}
                    </span>
                    <span class="inline-flex items-center gap-1.5 tabular-nums">
                      <span v-if="selectedGroupPromoActive" class="text-xs text-gray-400 line-through">
                        {{ formatPricePerM(applyBaseRate(model.cacheReadPrice)) }}
                      </span>
                      <span class="font-medium" :class="selectedGroupPromoActive ? 'text-rose-600 dark:text-rose-400' : 'text-sky-700 dark:text-sky-300'">
                        {{ formatPricePerM(applyRate(model.cacheReadPrice)) }}
                      </span>
                    </span>
                  </div>
                  <div v-if="model.cacheWritePrice != null && model.cacheWritePrice > 0" class="flex items-center justify-between">
                    <span class="inline-flex items-center gap-1.5 text-gray-500 dark:text-dark-400">
                      <Icon name="edit" size="sm" class="text-amber-500" />
                      {{ t('pricing.cacheWrite') }}
                    </span>
                    <span class="inline-flex items-center gap-1.5 tabular-nums">
                      <span v-if="selectedGroupPromoActive" class="text-xs text-gray-400 line-through">
                        {{ formatPricePerM(applyBaseRate(model.cacheWritePrice)) }}
                      </span>
                      <span class="font-medium" :class="selectedGroupPromoActive ? 'text-rose-600 dark:text-rose-400' : 'text-amber-700 dark:text-amber-300'">
                        {{ formatPricePerM(applyRate(model.cacheWritePrice)) }}
                      </span>
                    </span>
                  </div>
                </template>
              </div>

              <!-- 可用分组：精简后只在 >1 时显示作为快速切换器；=1 时省略避免噪音 -->
              <div
                v-if="model.accessibleGroups.length > 1"
                class="mt-4 flex flex-wrap items-center gap-1.5 border-t border-gray-100 pt-3 dark:border-dark-700/60"
              >
                <button
                  v-for="g in sortGroupsByRate(model.accessibleGroups)"
                  :key="g.id"
                  type="button"
                  class="inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-[11px] font-medium ring-1 ring-inset transition-colors"
                  :class="
                    selectedGroupId === g.id
                      ? 'bg-brand-50 text-brand-700 ring-brand-200/70 dark:bg-brand-500/15 dark:text-brand-300 dark:ring-brand-500/30'
                      : 'bg-gray-50 text-gray-600 ring-gray-200/70 hover:bg-gray-100 dark:bg-dark-800/60 dark:text-dark-200 dark:ring-dark-700/60 dark:hover:bg-dark-700/60'
                  "
                  :title="t('pricing.switchToGroup', { name: g.name })"
                  @click="selectedGroupId = g.id"
                >
                  <Icon v-if="g.is_exclusive" name="shield" size="xs" class="text-violet-500" />
                  <Icon v-if="promoActive(g)" name="fire" size="xs" class="text-rose-500 animate-pulse" />
                  <span>{{ g.name }}</span>
                  <span class="tabular-nums opacity-70">×{{ effectiveRate(g).toFixed(2) }}</span>
                </button>
              </div>
            </div>
          </div>

          <!-- 倍率提示：选中"全部分组"时说明显示原价 -->
          <p
            v-if="!loading && filteredModels.length > 0 && !selectedGroup"
            class="mt-4 inline-flex items-center gap-1.5 rounded-full bg-amber-50 px-3 py-1 text-[12px] font-medium text-amber-700 ring-1 ring-inset ring-amber-200/70 dark:bg-amber-500/15 dark:text-amber-300 dark:ring-amber-500/30"
          >
            <Icon name="infoCircle" size="xs" />
            {{ t('pricing.standardPriceHint') }}
          </p>
        </main>
      </div>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import PlatformIcon from '@/components/common/PlatformIcon.vue'
import userChannelsAPI, {
  type UserAvailableChannel,
  type UserAvailableGroup,
  type PricingListEntry,
} from '@/api/channels'
import userGroupsAPI from '@/api/groups'
import { paymentAPI } from '@/api/payment'
import { useAppStore } from '@/stores/app'
import { extractApiErrorMessage } from '@/utils/apiError'
import type { GroupPlatform } from '@/types'
import type { BillingMode } from '@/constants/channel'

const { t } = useI18n()
const appStore = useAppStore()

// ============ 原始数据 ============
const channels = ref<UserAvailableChannel[]>([])
const userGroupRates = ref<Record<number, number>>({})
const allPricingEntries = ref<PricingListEntry[]>([])
const loading = ref(false)

// ============ 筛选状态 ============
const searchQuery = ref('')
const selectedGroupId = ref<number | null>(null)
const selectedPlatform = ref<string | null>(null)
const selectedBillingMode = ref<BillingMode | null>(null)
// 移动端筛选面板折叠状态：默认折叠，避免列表被推到屏幕外
const mobileFiltersOpen = ref(false)
// 已选筛选数量（用作折叠按钮上的角标，让用户一眼看到当前有几个生效中的过滤条件）
const activeFilterCount = computed(() => {
  let n = 0
  if (selectedGroupId.value !== null) n++
  if (selectedPlatform.value !== null) n++
  if (selectedBillingMode.value !== null) n++
  return n
})

// ============ 货币 ============
// 站点计价模型说明：
//   - 后端 model_pricing 字段（input_cost_per_token 等）单位虽为 USD，但语义上是
//     「美元余额」（账户上扣多少美元额度）。
//   - 站点充值倍率 balance_recharge_multiplier（M）= ¥1 充值能得到的美元余额，
//     本站默认 M=1.0（¥1 = $1 余额，1:1）。
//   - 用户实付 CNY = 后端价 / M
//   - 真实 USD（按外汇汇率算）= CNY / usdToCny
//
// 默认显示 CNY（中国用户更直接）；USD 切回真实美元价（÷ usdToCny）。
const currency = ref<'USD' | 'CNY'>('CNY')

// 站点充值倍率（¥1 = M 美元余额）；通过 paymentAPI.getCheckoutInfo 异步获取
const rechargeMultiplier = ref(1.0)

// USD ↔ CNY 真实外汇汇率：默认 7.2，可从 public_settings.usd_to_cny_rate 覆盖
const usdToCny = computed(() => {
  const fromSettings = appStore.cachedPublicSettings as { usd_to_cny_rate?: number } | null
  const rate = fromSettings?.usd_to_cny_rate
  return typeof rate === 'number' && rate > 0 ? rate : 7.2
})

// ============ 数据加载 ============
async function loadAll() {
  loading.value = true
  try {
    const [list, rates, pricingResp, checkoutResp] = await Promise.all([
      userChannelsAPI.getAvailable(),
      userGroupsAPI.getUserGroupRates().catch(() => ({}) as Record<number, number>),
      userChannelsAPI.listAllPricing().catch(() => ({ models: [], metadata: {} as never })),
      // 拉取 balance_recharge_multiplier — 用于 USD/CNY 换算；失败时回退到默认 1.0
      paymentAPI.getCheckoutInfo().catch(() => null),
    ])
    channels.value = list
    userGroupRates.value = rates
    allPricingEntries.value = pricingResp.models || []
    const m = checkoutResp?.data?.balance_recharge_multiplier
    if (typeof m === 'number' && m > 0) rechargeMultiplier.value = m
    // 默认选中最便宜的分组（按 effectiveRate 升序后第一个）
    if (selectedGroupId.value === null && availableGroups.value.length > 0) {
      selectedGroupId.value = availableGroups.value[0].id
    }
  } catch (err: unknown) {
    appStore.showError(extractApiErrorMessage(err, t('common.error')))
  } finally {
    loading.value = false
  }
}

// ============ 限时倍率（promo rate） ============
// 每秒 +1 触发倒计时 / promo 状态判定的响应式 tick；mount 时启动 setInterval
const nowTick = ref<number>(Date.now())
let nowTimer: ReturnType<typeof setInterval> | null = null

// 是否在 promo 窗口内（依赖 nowTick 让 UI 自动响应分秒）
function promoActive(g: UserAvailableGroup): boolean {
  if (g.promo_rate_multiplier == null) return false
  const t = nowTick.value
  if (g.promo_starts_at && t < Date.parse(g.promo_starts_at)) return false
  if (g.promo_ends_at && t >= Date.parse(g.promo_ends_at)) return false
  return true
}

// 分组原价倍率（不应用 promo / 用户专属，仅纯展示用）
function baseRate(group: UserAvailableGroup): number {
  return group.rate_multiplier
}

// 实际倍率优先级：用户专属 > 限时 promo > 分组默认
function effectiveRate(group: UserAvailableGroup): number {
  const custom = userGroupRates.value[group.id]
  if (typeof custom === 'number') return custom
  if (promoActive(group)) return group.promo_rate_multiplier as number
  return group.rate_multiplier
}

// 倒计时格式化：返回 "2天 12:34:56" 或 "23:45:12"
function countdownLabel(g: UserAvailableGroup): string {
  if (!g.promo_ends_at) return ''
  const remaining = Math.max(0, Date.parse(g.promo_ends_at) - nowTick.value)
  const days = Math.floor(remaining / 86_400_000)
  const hours = Math.floor((remaining % 86_400_000) / 3_600_000)
  const mins = Math.floor((remaining % 3_600_000) / 60_000)
  const secs = Math.floor((remaining % 60_000) / 1_000)
  const pad = (n: number) => String(n).padStart(2, '0')
  if (days > 0) return t('pricing.promoCountdownDays', { d: days, h: pad(hours), m: pad(mins), s: pad(secs) })
  return `${pad(hours)}:${pad(mins)}:${pad(secs)}`
}

// 按 effectiveRate 升序排列分组（限时窗口内 promo 价更低就排前）
function sortGroupsByRate(groups: UserAvailableGroup[]): UserAvailableGroup[] {
  return [...groups].sort((a, b) => effectiveRate(a) - effectiveRate(b))
}

// ============ 派生：所有可访问分组 ============
const availableGroups = computed<UserAvailableGroup[]>(() => {
  const seen = new Set<number>()
  const out: UserAvailableGroup[] = []
  for (const ch of channels.value) {
    for (const sec of ch.platforms) {
      for (const g of sec.groups) {
        if (seen.has(g.id)) continue
        seen.add(g.id)
        out.push(g)
      }
    }
  }
  return out.sort((a, b) => {
    if (a.is_exclusive !== b.is_exclusive) return a.is_exclusive ? -1 : 1
    return effectiveRate(a) - effectiveRate(b)
  })
})

const selectedGroup = computed<UserAvailableGroup | null>(() => {
  if (selectedGroupId.value === null) return null
  return availableGroups.value.find(g => g.id === selectedGroupId.value) ?? null
})

// ============ 平台 mapping ============
// LiteLLM 的 provider 字段（如 "anthropic" / "openai" / "vertex_ai-anthropic_models"）
// 与系统平台（anthropic / openai / gemini / antigravity）的映射
function providerToPlatform(provider: string, modelName: string): string | null {
  const p = (provider || '').toLowerCase()
  const m = (modelName || '').toLowerCase()
  if (p.includes('anthropic') || m.startsWith('claude')) return 'anthropic'
  if (p === 'openai' || p === 'azure' || p.startsWith('text-completion-openai') || m.startsWith('gpt') || /^o[1-9]/.test(m)) return 'openai'
  if (p.includes('gemini') || p.includes('vertex_ai-language-models') || p === 'google' || m.startsWith('gemini')) return 'gemini'
  return null
}

const availablePlatforms = computed<string[]>(() => {
  const set = new Set<string>()
  // 从用户可访问分组得到
  for (const g of availableGroups.value) {
    if (g.platform) set.add(g.platform)
  }
  return Array.from(set).sort()
})

// 可访问平台集合（用户能用的）
const accessiblePlatformSet = computed<Set<string>>(() => new Set(availablePlatforms.value))

// platform → 可访问分组列表（用于反向关联模型 → 分组）
const platformToGroups = computed<Map<string, UserAvailableGroup[]>>(() => {
  const map = new Map<string, UserAvailableGroup[]>()
  for (const g of availableGroups.value) {
    if (!g.platform) continue
    const arr = map.get(g.platform) ?? []
    arr.push(g)
    map.set(g.platform, arr)
  }
  return map
})

// ============ 派生：模型聚合 ============
// 合并两个来源：
//   1) admin 渠道里手动加的支持模型（fromChannel=true，可能有自定义定价 / intervals）
//   2) LiteLLM 全量定价（fromChannel=false，使用标准定价；这是用户期望的"自动列全"）
//
// 同名同平台优先使用 1（admin 自定义）。
interface FlatModel {
  name: string
  platform: string
  fromChannel: boolean
  hasPricing: boolean
  billingMode: BillingMode
  inputPrice: number | null
  outputPrice: number | null
  cacheReadPrice: number | null
  cacheWritePrice: number | null
  perRequestPrice: number | null
  imageOutputPrice: number | null
  // 图片输出价的单位：per_image=每张、per_image_token=每个图片 token；channel 自定义来源默认按渠道侧的 per_token 含义。
  imageOutputPriceUnit: 'per_image' | 'per_image_token' | null
  accessibleGroups: UserAvailableGroup[]
}

const allModels = computed<FlatModel[]>(() => {
  const map = new Map<string, FlatModel>()

  // ── Step 1: 先放 admin 渠道自定义的模型（优先级最高）
  for (const ch of channels.value) {
    for (const sec of ch.platforms) {
      for (const m of sec.supported_models) {
        const key = `${m.platform}|${m.name}`
        let entry = map.get(key)
        if (!entry) {
          const p = m.pricing
          entry = {
            name: m.name,
            platform: m.platform,
            fromChannel: true,
            hasPricing: !!p,
            billingMode: (p?.billing_mode as BillingMode) || 'token',
            inputPrice: p?.input_price ?? null,
            outputPrice: p?.output_price ?? null,
            cacheReadPrice: p?.cache_read_price ?? null,
            cacheWritePrice: p?.cache_write_price ?? null,
            perRequestPrice: p?.per_request_price ?? null,
            imageOutputPrice: p?.image_output_price ?? null,
            // 渠道自定义的 image_output_price 语义就是 per-token（与渠道编辑器的"图片输出 $/MTok"一致）
            imageOutputPriceUnit: p?.image_output_price != null ? 'per_image_token' : null,
            accessibleGroups: [],
          }
          map.set(key, entry)
        }
        for (const g of sec.groups) {
          if (!entry.accessibleGroups.some(eg => eg.id === g.id)) {
            entry.accessibleGroups.push(g)
          }
        }
      }
    }
  }

  // ── Step 2: 再加 LiteLLM 全量模型（仅限用户可访问平台，不覆盖 Step 1 已存在的 key）
  for (const entry of allPricingEntries.value) {
    const platform = providerToPlatform(entry.provider, entry.model)
    if (!platform) continue
    if (!accessiblePlatformSet.value.has(platform)) continue
    const key = `${platform}|${entry.model}`
    if (map.has(key)) continue

    // 推断 billing mode：mode === 'image_generation' → image，其它都按 token
    const mode: BillingMode = entry.mode === 'image_generation' ? 'image' : 'token'

    // 图片输出价：优先取 per-image 单价（litellm 部分模型有，如 Gemini Image），
    // 没有就 fallback 到 per-image-token 单价（GPT 系列图片模型，如 gpt-image-1.5）。
    // 二者单位不同，但前端展示用同一个字段，详细单位由 imageOutputPriceUnit 区分。
    const imagePerImage = entry.output_cost_per_image || 0
    const imagePerImageToken = entry.output_cost_per_image_token || 0
    const imagePrice = imagePerImage > 0 ? imagePerImage : imagePerImageToken
    const imageUnit: 'per_image' | 'per_image_token' | null =
      imagePerImage > 0 ? 'per_image' : imagePerImageToken > 0 ? 'per_image_token' : null

    map.set(key, {
      name: entry.model,
      platform,
      fromChannel: false,
      hasPricing: entry.input_cost_per_token > 0 || entry.output_cost_per_token > 0 || imagePrice > 0,
      billingMode: mode,
      inputPrice: entry.input_cost_per_token || null,
      outputPrice: entry.output_cost_per_token || null,
      cacheReadPrice: entry.cache_read_input_token_cost || null,
      cacheWritePrice: entry.cache_creation_input_token_cost || null,
      perRequestPrice: null,
      imageOutputPrice: imagePrice || null,
      imageOutputPriceUnit: imageUnit,
      // 全量模型自动关联该平台的所有可访问分组
      accessibleGroups: platformToGroups.value.get(platform) ?? [],
    })
  }

  return Array.from(map.values()).sort((a, b) => {
    if (a.platform !== b.platform) return a.platform.localeCompare(b.platform)
    return a.name.localeCompare(b.name)
  })
})

const availableBillingModes = computed<BillingMode[]>(() => {
  const set = new Set<BillingMode>()
  for (const m of allModels.value) {
    if (m.billingMode) set.add(m.billingMode)
  }
  return Array.from(set).sort()
})

const filteredModels = computed<FlatModel[]>(() => {
  const q = searchQuery.value.trim().toLowerCase()
  return allModels.value.filter(m => {
    if (selectedGroupId.value !== null) {
      if (!m.accessibleGroups.some(g => g.id === selectedGroupId.value)) return false
    }
    if (selectedPlatform.value !== null && m.platform !== selectedPlatform.value) return false
    if (selectedBillingMode.value !== null && m.billingMode !== selectedBillingMode.value) return false
    if (q && !m.name.toLowerCase().includes(q) && !m.platform.toLowerCase().includes(q)) return false
    return true
  })
})

// ============ 价格换算 ============
// 当前生效的倍率（无选中分组时为 1，表示展示原价）
const activeRate = computed(() => (selectedGroup.value ? effectiveRate(selectedGroup.value) : 1))

// 选中分组的"原"倍率（即未应用 promo / 用户专属的分组默认倍率），用于划线对比展示
const activeBaseRate = computed(() => (selectedGroup.value ? baseRate(selectedGroup.value) : 1))

function applyRate(price: number | null | undefined): number | null {
  if (price == null) return null
  return price * activeRate.value
}

// ============ 图片模型专用 ============
// 分组开了"图片倍率独立"时，image 模型应当走 image_rate_multiplier 而不是 rate_multiplier。
// 没开 / 没选分组 → 跟 activeRate 一致。
const selectedGroupImageIndependent = computed(() => selectedGroup.value?.image_rate_independent === true)
const activeImageRate = computed(() => {
  const g = selectedGroup.value
  if (!g) return 1
  if (g.image_rate_independent && typeof g.image_rate_multiplier === 'number') {
    return g.image_rate_multiplier
  }
  return effectiveRate(g)
})
const selectedGroupImageRateLabel = computed(() => `×${Number(activeImageRate.value.toPrecision(10))}`)
function applyImageRate(price: number | null | undefined): number | null {
  if (price == null) return null
  return price * activeImageRate.value
}
// 分组配置的"每图固定价"（1K/2K/4K），有任一就走分支 A 渲染
const selectedGroupImagePrices = computed(() => ({
  k1: selectedGroup.value?.image_price_1k ?? null,
  k2: selectedGroup.value?.image_price_2k ?? null,
  k4: selectedGroup.value?.image_price_4k ?? null,
}))
const hasGroupImagePrices = computed(() => {
  const p = selectedGroupImagePrices.value
  return p.k1 != null || p.k2 != null || p.k4 != null
})

// 用"原倍率"换算价格，用于划线对比；仅当 promoActive 时才展示
function applyBaseRate(price: number | null | undefined): number | null {
  if (price == null) return null
  return price * activeBaseRate.value
}

// 选中分组当前是否处于 promo 窗口（展示划线 / 倒计时的触发条件）
const selectedGroupPromoActive = computed(() => (selectedGroup.value ? promoActive(selectedGroup.value) : false))

// 单位说明：API 返回的是 USD per token；CNY 模式下再 × usdToCny 汇率
const PER_MILLION = 1_000_000

function smartFixed(n: number): string {
  if (n === 0) return '0'
  if (n >= 100) return n.toFixed(2)
  if (n >= 1) return n.toFixed(3)
  return n.toFixed(4)
}

// applyCurrency 把后端"美元余额"价格换算为用户可读的实付价。
//   - CNY：后端价 / rechargeMultiplier（本站 multiplier=1 时数值不变，只换符号）
//   - USD（真实美元）：CNY / usdToCny
function applyCurrency(usdPrice: number): { value: number; symbol: string } {
  const cnyPaid = rechargeMultiplier.value > 0 ? usdPrice / rechargeMultiplier.value : usdPrice
  if (currency.value === 'CNY') {
    return { value: cnyPaid, symbol: '¥' }
  }
  return { value: cnyPaid / usdToCny.value, symbol: '$' }
}

// 单次价格（per_request / image）
function formatPrice(price: number | null | undefined): string {
  if (price == null) return '-'
  const { value, symbol } = applyCurrency(price)
  return `${symbol}${smartFixed(value)}`
}

// 每百万 token 价格
function formatPricePerM(price: number | null | undefined): string {
  if (price == null) return '-'
  const { value, symbol } = applyCurrency(price * PER_MILLION)
  return `${symbol}${smartFixed(value)}/M`
}

// ============ 计费类型 chip 配色 ============
function billingModeLabel(mode: BillingMode | string | undefined): string {
  if (!mode) return t('pricing.billingToken')
  if (mode === 'per_request') return t('pricing.billingPerRequest')
  if (mode === 'image') return t('pricing.billingImage')
  return t('pricing.billingToken')
}

function billingModeChipClass(mode: BillingMode | string | undefined): string {
  if (mode === 'per_request')
    return 'bg-violet-50 text-violet-700 ring-violet-200/70 dark:bg-violet-500/15 dark:text-violet-300 dark:ring-violet-500/30'
  if (mode === 'image')
    return 'bg-pink-50 text-pink-700 ring-pink-200/70 dark:bg-pink-500/15 dark:text-pink-300 dark:ring-pink-500/30'
  return 'bg-sky-50 text-sky-700 ring-sky-200/70 dark:bg-sky-500/15 dark:text-sky-300 dark:ring-sky-500/30'
}

// ============ 平台图标筐配色 ============
function platformIconBg(platform: string): string {
  switch (platform) {
    case 'anthropic':
      return 'bg-orange-50 ring-orange-200/70 dark:bg-orange-500/15 dark:ring-orange-500/30'
    case 'openai':
      return 'bg-emerald-50 ring-emerald-200/70 dark:bg-emerald-500/15 dark:ring-emerald-500/30'
    case 'gemini':
      return 'bg-sky-50 ring-sky-200/70 dark:bg-sky-500/15 dark:ring-sky-500/30'
    case 'antigravity':
      return 'bg-violet-50 ring-violet-200/70 dark:bg-violet-500/15 dark:ring-violet-500/30'
    default:
      return 'bg-gray-50 ring-gray-200/70 dark:bg-dark-700/40 dark:ring-dark-600/60'
  }
}

onMounted(() => {
  void loadAll()
  // 每秒 +1 触发倒计时 / promoActive 重算
  nowTimer = setInterval(() => {
    nowTick.value = Date.now()
  }, 1000)
})

onBeforeUnmount(() => {
  if (nowTimer != null) {
    clearInterval(nowTimer)
    nowTimer = null
  }
})
</script>
