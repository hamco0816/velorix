<template>
  <AppLayout wide>
    <div class="space-y-5">
      <!-- 顶部工具栏：平台筛选 + 时间窗口 + 刷新 -->
      <div class="flex flex-wrap items-center gap-3">
        <div class="inline-flex rounded-lg border border-gray-200/70 bg-white p-0.5 dark:border-dark-700/60 dark:bg-dark-800/40">
          <button
            v-for="opt in platformOptions"
            :key="opt.value"
            type="button"
            class="inline-flex items-center gap-1.5 rounded-md px-3 py-1.5 text-xs font-medium transition-colors"
            :class="platform === opt.value
              ? 'bg-gray-900 text-white dark:bg-white dark:text-gray-900'
              : 'text-gray-600 hover:text-gray-900 dark:text-dark-300 dark:hover:text-white'"
            @click="platform = opt.value"
          >
            <PlatformIcon
              v-if="opt.value"
              :platform="opt.value as GroupPlatform"
              size="xs"
            />
            {{ opt.label }}
          </button>
        </div>
        <div class="inline-flex rounded-lg border border-gray-200/70 bg-white p-0.5 dark:border-dark-700/60 dark:bg-dark-800/40">
          <button
            v-for="opt in daysOptions"
            :key="opt"
            type="button"
            class="rounded-md px-3 py-1.5 text-xs font-medium transition-colors"
            :class="days === opt
              ? 'bg-gray-900 text-white dark:bg-white dark:text-gray-900'
              : 'text-gray-600 hover:text-gray-900 dark:text-dark-300 dark:hover:text-white'"
            @click="days = opt"
          >
            {{ opt }} {{ t('admin.pricingAdvisor.daySuffix') }}
          </button>
        </div>
        <button class="btn btn-secondary btn-sm ml-auto inline-flex items-center gap-1.5" :disabled="loading" @click="loadAll">
          <Icon name="refresh" size="sm" :class="loading ? 'animate-spin' : ''" />
          <span>{{ loading ? t('common.loading') : t('common.refresh') }}</span>
        </button>
      </div>

      <!-- 业务说明：让 admin 看懂这些数字含义 + 重要的语义提醒 -->
      <div class="flex items-start gap-2.5 rounded-xl border border-sky-200/60 bg-sky-50/60 px-4 py-3 dark:border-sky-500/20 dark:bg-sky-500/5">
        <Icon name="infoCircle" size="sm" class="mt-0.5 shrink-0 text-sky-600 dark:text-sky-300" />
        <div class="text-xs leading-5 text-sky-900 dark:text-sky-100">
          <p class="font-semibold">{{ t('admin.pricingAdvisor.helpTitle') }}</p>
          <p class="mt-0.5">{{ t('admin.pricingAdvisor.helpBody') }}</p>
          <p class="mt-1 flex items-start gap-1 text-[11px] opacity-80">
            <Icon name="exclamationTriangle" size="xs" class="mt-0.5 shrink-0" />
            <span>{{ t('admin.pricingAdvisor.helpUsdNote') }}</span>
          </p>
        </div>
      </div>

      <!-- 加载状态 -->
      <div v-if="loading && stats.length === 0" class="flex items-center justify-center py-16">
        <LoadingSpinner size="md" />
      </div>

      <!-- 空状态 -->
      <EmptyState
        v-else-if="!loading && stats.length === 0"
        variant="emerald"
        :title="t('admin.pricingAdvisor.empty.title')"
        :description="t('admin.pricingAdvisor.empty.description')"
      >
        <template #icon>
          <Icon name="chart" class="empty-state-icon" />
        </template>
      </EmptyState>

      <template v-else>
        <!-- 各档位统计表 -->
        <section class="surface-card overflow-hidden">
          <header class="flex items-center gap-3 border-b border-gray-200/60 px-5 py-3.5 dark:border-dark-700/60">
            <Icon name="chartBar" size="sm" class="text-emerald-500" />
            <h2 class="text-[15px] font-semibold text-gray-900 dark:text-white">{{ t('admin.pricingAdvisor.tableTitle') }}</h2>
          </header>
          <!-- 桌面端：表格（扁平 8 列：平台/档位/样本/5h P95/7d P95/日限额/周限额/月限额）
               P50 & Max 不再单独列，hover P95 数字可见详情 -->
          <div class="hidden overflow-x-auto md:block">
            <table class="w-full min-w-[860px] text-[14px]">
              <thead class="bg-gray-50/60 dark:bg-dark-800/60">
                <tr>
                  <th class="px-4 py-3 text-center text-[12px] font-semibold uppercase tracking-wider text-gray-600 dark:text-dark-300">{{ t('admin.pricingAdvisor.col.platform') }}</th>
                  <th class="px-4 py-3 text-center text-[12px] font-semibold uppercase tracking-wider text-gray-600 dark:text-dark-300">{{ t('admin.pricingAdvisor.col.tier') }}</th>
                  <th class="cursor-help px-4 py-3 text-center text-[12px] font-semibold uppercase tracking-wider text-gray-600 dark:text-dark-300" :title="t('admin.pricingAdvisor.tip.samples')">{{ t('admin.pricingAdvisor.col.samples') }}</th>
                  <th class="cursor-help border-l border-gray-200/60 px-4 py-3 text-center text-[12px] font-semibold uppercase tracking-wider text-gray-600 dark:border-dark-700/60 dark:text-dark-300" :title="t('admin.pricingAdvisor.tip.5hP95')">{{ t('admin.pricingAdvisor.col.5hP95') }}</th>
                  <th class="cursor-help px-4 py-3 text-center text-[12px] font-semibold uppercase tracking-wider text-gray-600 dark:text-dark-300" :title="t('admin.pricingAdvisor.tip.7dP95')">{{ t('admin.pricingAdvisor.col.7dP95') }}</th>
                  <th class="cursor-help border-l border-gray-200/60 px-4 py-3 text-center text-[12px] font-semibold uppercase tracking-wider text-emerald-700 dark:border-dark-700/60 dark:text-emerald-300" :title="t('admin.pricingAdvisor.tip.limits')">{{ t('admin.pricingAdvisor.subcol.daily') }}</th>
                  <th class="px-4 py-3 text-center text-[12px] font-semibold uppercase tracking-wider text-emerald-700 dark:text-emerald-300">{{ t('admin.pricingAdvisor.subcol.weekly') }}</th>
                  <th class="px-4 py-3 text-center text-[12px] font-semibold uppercase tracking-wider text-emerald-700 dark:text-emerald-300">{{ t('admin.pricingAdvisor.subcol.monthly') }}</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-gray-100 dark:divide-dark-700/60">
                <tr
                  v-for="row in stats"
                  :key="`${row.platform}|${row.tier}`"
                  class="cursor-pointer transition-colors hover:bg-gray-50/60 dark:hover:bg-dark-800/40"
                  :class="selectedKey === `${row.platform}|${row.tier}` ? 'bg-brand-50/60 dark:bg-brand-500/10' : ''"
                  @click="selectRow(row)"
                >
                  <td class="px-4 py-3.5 text-center font-medium text-gray-900 dark:text-white">
                    <PlatformIcon :platform="row.platform as GroupPlatform" size="sm" class="mr-1.5 inline-block" />
                    <span>{{ row.platform }}</span>
                  </td>
                  <td class="px-4 py-3.5 text-center">
                    <span class="inline-flex items-center rounded-md bg-gray-100 px-2 py-0.5 text-[12px] font-medium text-gray-700 dark:bg-dark-700 dark:text-dark-200">
                      {{ formatTier(row.tier) }}
                    </span>
                  </td>
                  <td class="px-4 py-3.5 text-center tabular-nums">
                    <span :class="row.has_enough_samples ? 'text-gray-700 dark:text-dark-200' : 'text-amber-600 dark:text-amber-400'">
                      {{ row.sample_accounts }}
                    </span>
                    <Icon
                      v-if="!row.has_enough_samples"
                      name="exclamationTriangle"
                      size="xs"
                      class="ml-1 inline-block text-amber-500"
                      :title="t('admin.pricingAdvisor.samplesLow')"
                    />
                  </td>
                  <td class="cursor-help border-l border-gray-100 px-4 py-3.5 text-center tabular-nums text-[15px] text-gray-900 dark:border-dark-700/60 dark:text-white"
                      :title="`P50 $${row.window_5h_p50.toFixed(2)}  •  Max $${row.window_5h_max.toFixed(2)}`">
                    ${{ row.window_5h_p95.toFixed(2) }}
                  </td>
                  <td class="cursor-help px-4 py-3.5 text-center tabular-nums text-[15px] text-gray-900 dark:text-white"
                      :title="`P50 $${row.window_7d_p50.toFixed(2)}  •  Max $${row.window_7d_max.toFixed(2)}`">
                    ${{ row.window_7d_p95.toFixed(2) }}
                  </td>
                  <td class="border-l border-gray-100 px-4 py-3.5 text-center tabular-nums text-[15px] text-emerald-700 dark:border-dark-700/60 dark:text-emerald-300">${{ rowLimits(row).daily.toFixed(2) }}</td>
                  <td class="px-4 py-3.5 text-center tabular-nums text-[15px] text-emerald-700 dark:text-emerald-300">${{ rowLimits(row).weekly.toFixed(2) }}</td>
                  <td class="px-4 py-3.5 text-center tabular-nums text-[15px] font-semibold text-emerald-700 dark:text-emerald-300">${{ rowLimits(row).monthly.toFixed(2) }}</td>
                </tr>
              </tbody>
            </table>
          </div>

          <!-- 移动端：每个 tier 一张卡，纵向堆叠 -->
          <div class="divide-y divide-gray-100 dark:divide-dark-700/60 md:hidden">
            <div
              v-for="row in stats"
              :key="`mob-${row.platform}|${row.tier}`"
              class="cursor-pointer space-y-2 p-4 transition-colors hover:bg-gray-50/60 dark:hover:bg-dark-800/40"
              :class="selectedKey === `${row.platform}|${row.tier}` ? 'bg-brand-50/60 dark:bg-brand-500/10' : ''"
              @click="selectRow(row)"
            >
              <!-- 顶部：平台 + tier + 样本 -->
              <div class="flex flex-wrap items-center gap-2">
                <span class="inline-flex items-center gap-1.5 font-semibold text-gray-900 dark:text-white">
                  <PlatformIcon :platform="row.platform as GroupPlatform" size="sm" />
                  {{ row.platform }}
                </span>
                <span class="inline-flex items-center rounded-md bg-gray-100 px-2 py-0.5 text-xs font-medium text-gray-700 dark:bg-dark-700 dark:text-dark-200">
                  {{ formatTier(row.tier) }}
                </span>
                <span class="ml-auto inline-flex items-center gap-0.5 text-xs" :class="row.has_enough_samples ? 'text-gray-500 dark:text-dark-400' : 'text-amber-600 dark:text-amber-400'">
                  {{ row.sample_accounts }} {{ t('admin.pricingAdvisor.samples') }}
                  <Icon v-if="!row.has_enough_samples" name="exclamationTriangle" size="xs" class="ml-0.5" />
                </span>
              </div>
              <!-- 历史用量分位（参考） -->
              <div class="grid grid-cols-2 gap-2 text-center">
                <div class="rounded-lg bg-gray-50 p-2 dark:bg-dark-800/40">
                  <p class="text-[10px] font-medium uppercase tracking-wider text-gray-500 dark:text-dark-400">{{ t('admin.pricingAdvisor.col.5h') }}</p>
                  <p class="mt-0.5 text-sm font-semibold tabular-nums text-gray-900 dark:text-white">P50 ${{ row.window_5h_p50.toFixed(2) }}</p>
                  <p class="mt-0.5 text-[10px] tabular-nums text-gray-500 dark:text-dark-400">P95 ${{ row.window_5h_p95.toFixed(2) }}</p>
                </div>
                <div class="rounded-lg bg-gray-50 p-2 dark:bg-dark-800/40">
                  <p class="text-[10px] font-medium uppercase tracking-wider text-gray-500 dark:text-dark-400">{{ t('admin.pricingAdvisor.col.7d') }}</p>
                  <p class="mt-0.5 text-sm font-semibold tabular-nums text-gray-900 dark:text-white">P50 ${{ row.window_7d_p50.toFixed(2) }}</p>
                  <p class="mt-0.5 text-[10px] tabular-nums text-gray-500 dark:text-dark-400">P95 ${{ row.window_7d_p95.toFixed(2) }}</p>
                </div>
              </div>
              <!-- 单账号档位限额（核心） -->
              <div class="grid grid-cols-3 gap-2 text-center">
                <div class="rounded-lg bg-emerald-50/60 p-2 dark:bg-emerald-500/5">
                  <p class="text-[10px] font-medium uppercase tracking-wider text-emerald-600 dark:text-emerald-300">{{ t('admin.pricingAdvisor.subcol.daily') }}</p>
                  <p class="mt-0.5 text-sm font-semibold tabular-nums text-emerald-700 dark:text-emerald-300">${{ rowLimits(row).daily.toFixed(2) }}</p>
                </div>
                <div class="rounded-lg bg-emerald-50/60 p-2 dark:bg-emerald-500/5">
                  <p class="text-[10px] font-medium uppercase tracking-wider text-emerald-600 dark:text-emerald-300">{{ t('admin.pricingAdvisor.subcol.weekly') }}</p>
                  <p class="mt-0.5 text-sm font-semibold tabular-nums text-emerald-700 dark:text-emerald-300">${{ rowLimits(row).weekly.toFixed(2) }}</p>
                </div>
                <div class="rounded-lg bg-emerald-50/60 p-2 dark:bg-emerald-500/5">
                  <p class="text-[10px] font-medium uppercase tracking-wider text-emerald-600 dark:text-emerald-300">{{ t('admin.pricingAdvisor.subcol.monthly') }}</p>
                  <p class="mt-0.5 text-sm font-semibold tabular-nums text-emerald-700 dark:text-emerald-300">${{ rowLimits(row).monthly.toFixed(2) }}</p>
                </div>
              </div>
            </div>
          </div>
        </section>

        <!-- ROI / 限额计算器 + 趋势图 -->
        <div class="grid gap-4 lg:grid-cols-[minmax(0,1fr)_minmax(0,1.3fr)]">
          <section class="surface-card p-5">
            <header class="mb-3 flex items-center gap-3">
              <Icon name="calculator" size="sm" class="text-amber-500" />
              <h2 class="text-[15px] font-semibold text-gray-900 dark:text-white">{{ t('admin.pricingAdvisor.calculator.title') }}</h2>
            </header>
            <div class="space-y-3">
              <!-- 主输入：档位 + N 横向并排，节省纵向空间 -->
              <div class="grid grid-cols-[1fr_auto] gap-3">
                <div>
                  <label class="input-label">{{ t('admin.pricingAdvisor.calculator.tier') }}</label>
                  <Select v-model="calcTierKey" :options="calcTierOptions" />
                </div>
                <div class="w-28">
                  <label class="input-label cursor-help" :title="t('admin.pricingAdvisor.calculator.usersPerAccountTip')">
                    {{ t('admin.pricingAdvisor.calculator.usersPerAccount') }}
                    <Icon name="infoCircle" size="xs" class="ml-0.5 inline-block text-gray-400" />
                  </label>
                  <input v-model.number="calcUsersPerAccount" type="number" min="1" step="1" class="input" placeholder="1" />
                </div>
              </div>

              <!-- 主结果：cap + 限额合并为一张卡，cap 在顶部一行展示、限额在下方 4 列大字 -->
              <div v-if="calcResult" class="space-y-2">
                <div class="rounded-xl border border-emerald-200/60 bg-emerald-50/40 p-3 dark:border-emerald-500/20 dark:bg-emerald-500/5">
                  <!-- cap 顶部行：横向展示 5h / 7d 与来源标签 -->
                  <div class="flex items-center justify-between gap-3 border-b border-emerald-200/40 pb-2 text-[12px] dark:border-emerald-500/15">
                    <span class="cursor-help font-medium uppercase tracking-wider text-violet-700/80 dark:text-violet-300/80"
                          :title="t('admin.pricingAdvisor.calculator.tierCapTitle')">
                      {{ t('admin.pricingAdvisor.calculator.tierCapTitle') }}
                    </span>
                    <span class="flex items-center gap-3 tabular-nums">
                      <span class="text-violet-900 dark:text-violet-100">
                        5h <span class="font-semibold">${{ calcResult.cap5hUsd.toFixed(0) }}</span>
                        <span class="ml-1 text-[10px] text-violet-600/70 dark:text-violet-300/60">{{ t('admin.pricingAdvisor.calculator.capSource_' + calcResult.cap5hSource) }}</span>
                      </span>
                      <span class="text-violet-300 dark:text-violet-500">·</span>
                      <span class="text-violet-900 dark:text-violet-100">
                        7d <span class="font-semibold">${{ calcResult.cap7dUsd.toFixed(0) }}</span>
                        <span class="ml-1 text-[10px] text-violet-600/70 dark:text-violet-300/60">{{ t('admin.pricingAdvisor.calculator.capSource_' + calcResult.cap7dSource) }}</span>
                      </span>
                    </span>
                  </div>
                  <!-- 限额结果：4 列大字 -->
                  <div class="mt-2.5 grid grid-cols-4 gap-2 text-center">
                    <div>
                      <p class="text-[11px] font-medium uppercase tracking-wider text-emerald-700/80 dark:text-emerald-300/70">
                        {{ t('admin.pricingAdvisor.calculator.suggested5h') }}
                        <span class="text-emerald-700/50 dark:text-emerald-300/50">{{ t('admin.pricingAdvisor.calculator.suggested5hRefMark') }}</span>
                      </p>
                      <p class="mt-1 text-[18px] font-bold tabular-nums text-emerald-900 dark:text-emerald-100">${{ calcResult.fiveHourLimitUsd.toFixed(2) }}</p>
                    </div>
                    <div>
                      <p class="text-[11px] font-medium uppercase tracking-wider text-emerald-700/80 dark:text-emerald-300/70">{{ t('admin.pricingAdvisor.calculator.suggestedDaily') }}</p>
                      <p class="mt-1 text-[18px] font-bold tabular-nums text-emerald-900 dark:text-emerald-100">${{ calcResult.dailyLimitUsd.toFixed(2) }}</p>
                    </div>
                    <div>
                      <p class="text-[11px] font-medium uppercase tracking-wider text-emerald-700/80 dark:text-emerald-300/70">{{ t('admin.pricingAdvisor.calculator.suggestedWeekly') }}</p>
                      <p class="mt-1 text-[18px] font-bold tabular-nums text-emerald-900 dark:text-emerald-100">${{ calcResult.weeklyLimitUsd.toFixed(2) }}</p>
                    </div>
                    <div>
                      <p class="text-[11px] font-medium uppercase tracking-wider text-emerald-700/80 dark:text-emerald-300/70">{{ t('admin.pricingAdvisor.calculator.suggestedMonthly') }}</p>
                      <p class="mt-1 text-[18px] font-bold tabular-nums text-emerald-900 dark:text-emerald-100">${{ calcResult.monthlyLimitUsd.toFixed(2) }}</p>
                    </div>
                  </div>
                </div>

                <!-- 警告（compact） -->
                <p v-if="calcResult.warning" class="flex items-start gap-1.5 rounded-md bg-amber-50 px-2.5 py-1.5 text-[11px] text-amber-700 dark:bg-amber-500/10 dark:text-amber-300">
                  <Icon name="exclamationTriangle" size="xs" class="mt-0.5 shrink-0" />
                  <span>{{ calcResult.warning }}</span>
                </p>
              </div>
              <p v-else class="rounded-xl bg-gray-50 px-4 py-3 text-xs text-gray-500 dark:bg-dark-800/40 dark:text-dark-400">
                {{ t('admin.pricingAdvisor.calculator.selectFirst') }}
              </p>

              <!-- 高级参数（折叠）：cap 手动覆盖 + 套餐倍率 -->
              <details v-if="calcResult" class="rounded-lg border border-gray-200/60 bg-gray-50/40 px-3 py-2 dark:border-dark-700/60 dark:bg-dark-800/30">
                <summary class="cursor-pointer text-xs font-medium text-gray-600 dark:text-dark-300">
                  {{ t('admin.pricingAdvisor.calculator.advancedTitle') }}
                </summary>
                <p class="mt-2 text-[11px] leading-relaxed text-gray-500 dark:text-dark-400">
                  {{ t('admin.pricingAdvisor.calculator.advancedHint') }}
                </p>
                <div class="mt-2 grid grid-cols-2 gap-3">
                  <div>
                    <label class="input-label cursor-help" :title="t('admin.pricingAdvisor.calculator.capOverrideTip')">{{ t('admin.pricingAdvisor.calculator.capOverride5h') }}</label>
                    <div class="flex items-center gap-1.5">
                      <span class="text-sm text-gray-500 dark:text-dark-400">$</span>
                      <input v-model.number="calcCapOverride5h" type="number" min="0" step="1" class="input flex-1" :placeholder="String(calcResult.cap5hUsd.toFixed(0))" />
                    </div>
                  </div>
                  <div>
                    <label class="input-label cursor-help" :title="t('admin.pricingAdvisor.calculator.capOverrideTip')">{{ t('admin.pricingAdvisor.calculator.capOverride7d') }}</label>
                    <div class="flex items-center gap-1.5">
                      <span class="text-sm text-gray-500 dark:text-dark-400">$</span>
                      <input v-model.number="calcCapOverride7d" type="number" min="0" step="1" class="input flex-1" :placeholder="String(calcResult.cap7dUsd.toFixed(0))" />
                    </div>
                  </div>
                </div>
                <div class="mt-3">
                  <label class="input-label cursor-help" :title="t('admin.pricingAdvisor.calculator.rateMultiplierTip')">
                    {{ t('admin.pricingAdvisor.calculator.rateMultiplier') }}
                    <Icon name="infoCircle" size="xs" class="ml-0.5 inline-block text-gray-400" />
                  </label>
                  <input v-model.number="calcRateMultiplier" type="number" min="0.1" max="10" step="0.1" class="input w-32" placeholder="1" />
                </div>
              </details>

              <!-- 定价（折叠 / 可选）：算建议月费和利润 -->
              <details v-if="calcResult" class="rounded-lg border border-gray-200/60 bg-gray-50/40 px-3 py-2 dark:border-dark-700/60 dark:bg-dark-800/30">
                <summary class="cursor-pointer text-xs font-medium text-gray-600 dark:text-dark-300">
                  {{ t('admin.pricingAdvisor.calculator.pricingTitle') }}
                </summary>
                <p class="mt-2 text-[11px] leading-relaxed text-gray-500 dark:text-dark-400">
                  {{ t('admin.pricingAdvisor.calculator.pricingHint') }}
                </p>
                <div class="mt-2 grid grid-cols-2 gap-3">
                  <div>
                    <label class="input-label">{{ t('admin.pricingAdvisor.calculator.cost') }}</label>
                    <input v-model.number="calcCost" type="number" min="0" step="1" class="input" placeholder="1400" />
                  </div>
                  <div>
                    <label class="input-label cursor-help" :title="t('admin.pricingAdvisor.calculator.markupTip')">
                      {{ t('admin.pricingAdvisor.calculator.markup') }}
                      <Icon name="infoCircle" size="xs" class="ml-0.5 inline-block text-gray-400" />
                    </label>
                    <div class="flex items-center gap-1.5">
                      <input v-model.number="calcMarkup" type="number" min="0" max="500" step="5" class="input flex-1" placeholder="30" />
                      <span class="text-sm text-gray-500 dark:text-dark-400">%</span>
                    </div>
                  </div>
                </div>
                <div class="mt-3 grid grid-cols-2 gap-3">
                  <div class="rounded-lg border border-emerald-200/60 bg-emerald-50/60 p-3 text-center dark:border-emerald-500/20 dark:bg-emerald-500/5">
                    <p class="text-[10px] font-medium uppercase tracking-wider text-emerald-700 dark:text-emerald-300">{{ t('admin.pricingAdvisor.calculator.suggestedPrice') }}</p>
                    <p class="mt-1 flex items-baseline justify-center gap-0.5">
                      <span class="text-base font-semibold text-emerald-700 dark:text-emerald-300">¥</span>
                      <span class="text-2xl font-bold tabular-nums text-emerald-900 dark:text-emerald-100">{{ calcResult.priceCny.toFixed(0) }}</span>
                      <span class="text-[10px] text-emerald-700 dark:text-emerald-300">/{{ t('admin.pricingAdvisor.calculator.perMonth') }}</span>
                    </p>
                  </div>
                  <div class="rounded-lg border p-3 text-center"
                    :class="calcResult.monthlyProfitCny >= 0
                      ? 'border-amber-200/60 bg-amber-50/60 dark:border-amber-500/20 dark:bg-amber-500/5'
                      : 'border-rose-200/60 bg-rose-50/60 dark:border-rose-500/20 dark:bg-rose-500/5'"
                  >
                    <p class="text-[10px] font-medium uppercase tracking-wider"
                      :class="calcResult.monthlyProfitCny >= 0
                        ? 'text-amber-700 dark:text-amber-300'
                        : 'text-rose-700 dark:text-rose-300'"
                    >{{ t('admin.pricingAdvisor.calculator.profit') }}</p>
                    <p class="mt-1 flex items-baseline justify-center gap-0.5"
                      :class="calcResult.monthlyProfitCny >= 0
                        ? 'text-amber-900 dark:text-amber-100'
                        : 'text-rose-900 dark:text-rose-100'"
                    >
                      <span class="text-base font-semibold"
                        :class="calcResult.monthlyProfitCny >= 0
                          ? 'text-amber-700 dark:text-amber-300'
                          : 'text-rose-700 dark:text-rose-300'"
                      >¥</span>
                      <span class="text-2xl font-bold tabular-nums">{{ calcResult.monthlyProfitCny.toFixed(0) }}</span>
                      <span class="text-[10px]"
                        :class="calcResult.monthlyProfitCny >= 0
                          ? 'text-amber-700 dark:text-amber-300'
                          : 'text-rose-700 dark:text-rose-300'"
                      >/{{ t('admin.pricingAdvisor.calculator.perMonth') }}</span>
                    </p>
                  </div>
                </div>
                <p class="mt-2 text-[10px] italic text-gray-500 dark:text-dark-400">
                  {{ t('admin.pricingAdvisor.calculator.pricingFormula', { cost: calcCost, markup: calcMarkup, n: calcResult.n }) }}
                </p>
                <button
                  type="button"
                  class="btn btn-primary mt-3 w-full"
                  @click="applyToPlan"
                >
                  <Icon name="plus" size="sm" class="mr-1.5" />
                  {{ t('admin.pricingAdvisor.calculator.applyToPlan') }}
                </button>
              </details>
            </div>
          </section>

          <!-- 趋势图 -->
          <section class="surface-card overflow-hidden">
            <header class="flex items-center gap-3 border-b border-gray-200/60 px-5 py-3.5 dark:border-dark-700/60">
              <Icon name="trendingUp" size="sm" class="text-violet-500" />
              <h2 class="text-[15px] font-semibold text-gray-900 dark:text-white">{{ t('admin.pricingAdvisor.trendTitle') }}</h2>
            </header>
            <div class="p-4">
              <div v-if="trendChartData" class="h-72">
                <Line :data="trendChartData" :options="trendChartOptions" />
              </div>
              <div v-else class="flex h-72 items-center justify-center text-sm text-gray-400 dark:text-dark-500">
                {{ t('admin.pricingAdvisor.trendEmpty') }}
              </div>
            </div>
          </section>
        </div>
      </template>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import Select from '@/components/common/Select.vue'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import PlatformIcon from '@/components/common/PlatformIcon.vue'
import { pricingAdvisorAPI, type TierStats, type TierTrend } from '@/api/admin/pricingAdvisor'
import { useAppStore } from '@/stores/app'
import { extractApiErrorMessage } from '@/utils/apiError'
import { useRouter } from 'vue-router'
import type { GroupPlatform } from '@/types'
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Tooltip,
  Legend,
  Filler,
} from 'chart.js'
import { Line } from 'vue-chartjs'

ChartJS.register(CategoryScale, LinearScale, PointElement, LineElement, Tooltip, Legend, Filler)

const { t } = useI18n()
const appStore = useAppStore()
const router = useRouter()

// ── 控件状态 ──
const platform = ref<string>('') // '' = 全部
const days = ref<number>(30)
const daysOptions = [7, 30, 90]
const platformOptions = computed(() => [
  { value: '', label: t('common.all') },
  { value: 'openai', label: 'OpenAI' },
  { value: 'anthropic', label: 'Anthropic' },
  { value: 'gemini', label: 'Gemini' },
])

const loading = ref(false)
const stats = ref<TierStats[]>([])
const trends = ref<TierTrend[]>([])
const selectedKey = ref<string>('') // 选中行的 platform|tier，用于计算器联动

// ── 数据加载 ──
async function loadAll() {
  loading.value = true
  try {
    const [statsRes, trendRes] = await Promise.all([
      pricingAdvisorAPI.tierStats({ days: days.value, platform: platform.value || undefined }),
      pricingAdvisorAPI.tierTrend({ days: days.value, platform: platform.value || undefined }),
    ])
    stats.value = statsRes.items || []
    trends.value = trendRes.items || []
    // 选中第一个有足够样本的 tier 作为计算器默认值
    if (!selectedKey.value && stats.value.length > 0) {
      const preferred = stats.value.find((r) => r.has_enough_samples) || stats.value[0]
      selectedKey.value = `${preferred.platform}|${preferred.tier}`
    }
  } catch (err) {
    appStore.showError(extractApiErrorMessage(err, t('common.error')))
  } finally {
    loading.value = false
  }
}

watch([platform, days], loadAll)
onMounted(loadAll)

function selectRow(row: TierStats) {
  selectedKey.value = `${row.platform}|${row.tier}`
}

// 一键创建套餐：把计算器结果作为 URL 参数带过去，套餐创建页能自动预填
// 不走"创建后绑定"链路，admin 可在套餐创建页继续微调再保存
function applyToPlan() {
  if (!calcResult.value) return
  const tier = stats.value.find((s) => `${s.platform}|${s.tier}` === calcTierKey.value)
  if (!tier) return
  const r = calcResult.value
  // 平台名首字母大写，让生成的套餐名好看：openai → OpenAI
  const platformName = tier.platform.charAt(0).toUpperCase() + tier.platform.slice(1)
  const tierLabel = formatTier(tier.tier) || tier.platform
  router.push({
    path: '/admin/orders/plans',
    query: {
      prefill: '1',
      name: `${platformName} ${tierLabel}`,
      price: r.priceCny.toFixed(0),
      daily_limit_usd: r.dailyLimitUsd.toFixed(2),
      weekly_limit_usd: r.weeklyLimitUsd.toFixed(2),
      monthly_limit_usd: r.monthlyLimitUsd.toFixed(2),
      // 套餐倍率：admin 在计算器里设了非 1 倍率时把它带过去；plan 默认值不写，避免覆盖 group
      ...(r.rateMul !== 1 ? { rate_multiplier: r.rateMul.toFixed(2) } : {}),
    },
  })
}

function formatTier(tier: string): string {
  if (!tier) return t('admin.accounts.tierUnclassified')
  // 把 underscored key 转成人类友好显示：pro_5x → Pro 5x
  return tier.split('_').map((w) => w.charAt(0).toUpperCase() + w.slice(1)).join(' ')
}

// ── 限额计算器 ──
// 核心问题：1 个账号分给 N 个用户，每人日 / 周 / 月限额各应该多少。
// 计算路径：从档位反推上游 cap → cap × 套餐倍率 / N。cap 优先级：手动 > utilization 反推 > 7d_P95 兜底。
// 周/日/月限额按 cap_7d 摊算：周 = cap_7d，日 = cap_7d / 7，月 = cap_7d × 30/7（按 30 天计）。
// 定价是次要功能，独立计算：建议月费 = cost × (1+加价率) / N，月利润 = 月费 × N − cost。
const calcUsersPerAccount = ref<number>(1) // 单账号承载用户数 N
const calcRateMultiplier = ref<number>(1) // 套餐倍率：用户实际扣费 = 上游成本 × 倍率
const calcCapOverride5h = ref<number | null>(null) // 手动覆盖 5h cap（USD），留空走自动反推/回退
const calcCapOverride7d = ref<number | null>(null) // 手动覆盖 7d cap（USD）
const calcCost = ref<number>(1400) // 账号成本（仅定价用）
const calcMarkup = ref<number>(30) // 加价率 %（仅定价用）
const calcTierKey = ref<string>('') // 跟 selectedKey 联动

// 计算单账号档位的天/周/月限额：基于 cap_7d 摊算，cap 缺失回退 7d_P95
function rowLimits(row: TierStats) {
  const has7dCap = row.cap_sample_count > 0 && row.cap_7d_usd > 0
  const cap7d = has7dCap ? row.cap_7d_usd : row.window_7d_p95
  return {
    daily: cap7d / 7,
    weekly: cap7d,
    monthly: (cap7d * 30) / 7,
  }
}

watch(selectedKey, (val) => {
  if (val) calcTierKey.value = val
})

const calcTierOptions = computed(() => {
  return stats.value.map((s) => ({
    value: `${s.platform}|${s.tier}`,
    label: `${s.platform} · ${formatTier(s.tier)} (${s.sample_accounts} ${t('admin.pricingAdvisor.samples')})`,
  }))
})

// USD ↔ CNY 换算：站点 1¥ = 1$，直接 1:1
const USD_TO_CNY = 1

// 价格阶梯取整：低价取 ¥1，中价取 ¥5，高价取 ¥10，避免低价档位被强行抬高 30%+
function roundPriceCny(raw: number): number {
  if (raw < 50) return Math.ceil(raw)
  if (raw < 200) return Math.ceil(raw / 5) * 5
  return Math.ceil(raw / 10) * 10
}

const calcResult = computed(() => {
  if (!calcTierKey.value) return null
  const tier = stats.value.find((s) => `${s.platform}|${s.tier}` === calcTierKey.value)
  if (!tier) return null

  const cost = Math.max(0, calcCost.value || 0)
  const markup = Math.max(0, Math.min(500, calcMarkup.value || 0))
  const n = Math.max(1, Math.floor(calcUsersPerAccount.value || 1))
  const rateMul = Math.max(0.1, Math.min(10, calcRateMultiplier.value || 1))

  // cap 选取，5h 与 7d 各自独立判断（不能因为 5h 没采到就把 7d 也降级）：
  // 优先级：手动 override > utilization 反推（util ≥ 5%）> 7d_P95 兜底
  const has5hReversed = tier.cap_sample_count > 0 && tier.cap_5h_usd > 0
  const has7dReversed = tier.cap_sample_count > 0 && tier.cap_7d_usd > 0
  const override5h = calcCapOverride5h.value && calcCapOverride5h.value > 0 ? calcCapOverride5h.value : null
  const override7d = calcCapOverride7d.value && calcCapOverride7d.value > 0 ? calcCapOverride7d.value : null
  const cap5hUsd = override5h ?? (has5hReversed ? tier.cap_5h_usd : tier.window_5h_p95)
  const cap7dUsd = override7d ?? (has7dReversed ? tier.cap_7d_usd : tier.window_7d_p95)
  const cap5hSource: 'override' | 'reversed' | 'fallback' = override5h ? 'override' : has5hReversed ? 'reversed' : 'fallback'
  const cap7dSource: 'override' | 'reversed' | 'fallback' = override7d ? 'override' : has7dReversed ? 'reversed' : 'fallback'

  // 限额按 N 均分，再按套餐倍率换算成"用户侧 ActualCost"单位：
  // - 上游 cap 是原始美元成本；用户的 limit 字段实际扣的是 ActualCost = 上游成本 × rate_multiplier
  // - 所以 user_limit = cap × rate_multiplier / N（倍率=1 时无影响）
  // - 月限额按 30 天计算：weekly × 30/7 ≈ weekly × 4.2857；不再用 ×4（少算 2~3 天）
  const fiveHourLimitUsd = (cap5hUsd * rateMul) / n
  const weeklyLimitUsd = (cap7dUsd * rateMul) / n
  const dailyLimitUsd = weeklyLimitUsd / 7
  const monthlyLimitUsd = (weeklyLimitUsd * 30) / 7

  // 定价（独立可选区域，不再做 occupancy 加权）：
  // - 建议月费/人 = cost × (1+加价率) / N
  // - 月利润 = 月费 × N − cost（直观：N 个座位都按这个价格卖出，扣掉成本）
  const totalRevenueCny = cost * (1 + markup / 100)
  const rawPriceCny = totalRevenueCny / n
  const priceCny = roundPriceCny(rawPriceCny)
  const monthlyProfitCny = priceCny * n - cost

  // 风险提示
  let warning = ''
  if (!tier.has_enough_samples) {
    warning = t('admin.pricingAdvisor.calculator.warnSamplesLow')
  } else if (cap5hSource === 'fallback' || cap7dSource === 'fallback') {
    warning = t('admin.pricingAdvisor.calculator.warnNoCapSample')
  } else if (n > 50) {
    warning = t('admin.pricingAdvisor.calculator.warnNTooHigh')
  }

  return {
    priceCny,
    monthlyLimitUsd,
    weeklyLimitUsd,
    dailyLimitUsd,
    fiveHourLimitUsd,
    cap5hUsd,
    cap7dUsd,
    cap5hSource,
    cap7dSource,
    capSampleCount: tier.cap_sample_count,
    monthlyProfitCny,
    warning,
    n,
    rateMul,
    usdToCny: USD_TO_CNY,
  }
})

// ── 趋势图 ──
const TREND_COLORS = [
  '#10b981', '#f97316', '#3b82f6', '#a855f7', '#ef4444',
  '#06b6d4', '#eab308', '#ec4899', '#84cc16', '#0ea5e9',
]

const trendChartData = computed(() => {
  if (trends.value.length === 0) return null
  // 收集所有日期作为统一 x 轴
  const allDates = new Set<string>()
  trends.value.forEach((t) => t.points.forEach((p) => allDates.add(p.date)))
  const labels = Array.from(allDates).sort()
  // 每个 (platform, tier) 一条线
  const datasets = trends.value.map((tt, idx) => {
    const pointMap = new Map(tt.points.map((p) => [p.date, p.avg_per_acc]))
    const color = TREND_COLORS[idx % TREND_COLORS.length]
    return {
      label: `${tt.platform} · ${formatTier(tt.tier)}`,
      data: labels.map((d) => pointMap.get(d) ?? null),
      borderColor: color,
      backgroundColor: `${color}20`,
      tension: 0.3,
      pointRadius: 2,
      pointHoverRadius: 4,
      spanGaps: true,
    }
  })
  return { labels, datasets }
})

const trendChartOptions = {
  responsive: true,
  maintainAspectRatio: false,
  interaction: { mode: 'index' as const, intersect: false },
  scales: {
    y: {
      type: 'linear' as const,
      title: { display: true, text: 'USD / 账号 / 天' },
      ticks: {
        callback: (v: number | string) => `$${Number(v).toFixed(0)}`,
      },
    },
  },
  plugins: {
    legend: { position: 'bottom' as const, labels: { boxWidth: 12, font: { size: 11 } } },
    tooltip: {
      callbacks: {
        label: (ctx: { dataset: { label?: string }; parsed: { y: number | null } }) => {
          const y = ctx.parsed.y ?? 0
          return `${ctx.dataset.label || ''}: $${y.toFixed(2)}`
        },
      },
    },
  },
}
</script>
