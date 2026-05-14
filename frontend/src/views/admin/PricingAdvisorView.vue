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
        <button class="btn btn-secondary btn-sm ml-auto" :disabled="loading" :title="t('common.refresh')" @click="loadAll">
          <Icon name="refresh" size="sm" :class="loading ? 'animate-spin' : ''" />
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
          <!-- 桌面端：表格 -->
          <div class="hidden overflow-x-auto md:block">
            <table class="w-full min-w-[860px] text-sm">
              <thead class="bg-gray-50/60 dark:bg-dark-800/60">
                <tr>
                  <th class="px-4 py-3 text-left text-[11px] font-medium uppercase tracking-wider text-gray-500 dark:text-dark-400">{{ t('admin.pricingAdvisor.col.platform') }}</th>
                  <th class="px-4 py-3 text-center text-[11px] font-medium uppercase tracking-wider text-gray-500 dark:text-dark-400">{{ t('admin.pricingAdvisor.col.tier') }}</th>
                  <th class="cursor-help px-4 py-3 text-center text-[11px] font-medium uppercase tracking-wider text-gray-500 dark:text-dark-400" :title="t('admin.pricingAdvisor.tip.samples')">{{ t('admin.pricingAdvisor.col.samples') }}</th>
                  <th class="cursor-help border-l border-gray-200/60 px-4 py-3 text-center text-[11px] font-medium uppercase tracking-wider text-gray-500 dark:border-dark-700/60 dark:text-dark-400" :title="t('admin.pricingAdvisor.tip.5h')" colspan="3">{{ t('admin.pricingAdvisor.col.5h') }}</th>
                  <th class="cursor-help border-l border-gray-200/60 px-4 py-3 text-center text-[11px] font-medium uppercase tracking-wider text-gray-500 dark:border-dark-700/60 dark:text-dark-400" :title="t('admin.pricingAdvisor.tip.7d')" colspan="3">{{ t('admin.pricingAdvisor.col.7d') }}</th>
                  <th class="cursor-help border-l border-gray-200/60 px-4 py-3 text-right text-[11px] font-medium uppercase tracking-wider text-gray-500 dark:border-dark-700/60 dark:text-dark-400" :title="t('admin.pricingAdvisor.tip.dailyAvg')">{{ t('admin.pricingAdvisor.col.dailyAvg') }}</th>
                </tr>
                <tr class="border-t border-gray-100 bg-gray-50/30 dark:border-dark-700/40 dark:bg-dark-800/40">
                  <th class="px-4 py-1.5"></th>
                  <th class="px-4 py-1.5"></th>
                  <th class="px-4 py-1.5"></th>
                  <th class="cursor-help border-l border-gray-200/60 px-4 py-1.5 text-right text-[10px] font-medium text-gray-500 dark:border-dark-700/60 dark:text-dark-400" :title="t('admin.pricingAdvisor.tip.p50')">
                    {{ t('admin.pricingAdvisor.subcol.p50') }}<span class="ml-0.5 text-gray-400">?</span>
                  </th>
                  <th class="cursor-help px-4 py-1.5 text-right text-[10px] font-medium text-gray-500 dark:text-dark-400" :title="t('admin.pricingAdvisor.tip.p95')">
                    {{ t('admin.pricingAdvisor.subcol.p95') }}<span class="ml-0.5 text-gray-400">?</span>
                  </th>
                  <th class="cursor-help px-4 py-1.5 text-right text-[10px] font-medium text-gray-500 dark:text-dark-400" :title="t('admin.pricingAdvisor.tip.max')">
                    {{ t('admin.pricingAdvisor.subcol.max') }}<span class="ml-0.5 text-gray-400">?</span>
                  </th>
                  <th class="cursor-help border-l border-gray-200/60 px-4 py-1.5 text-right text-[10px] font-medium text-gray-500 dark:border-dark-700/60 dark:text-dark-400" :title="t('admin.pricingAdvisor.tip.p50')">
                    {{ t('admin.pricingAdvisor.subcol.p50') }}<span class="ml-0.5 text-gray-400">?</span>
                  </th>
                  <th class="cursor-help px-4 py-1.5 text-right text-[10px] font-medium text-gray-500 dark:text-dark-400" :title="t('admin.pricingAdvisor.tip.p95')">
                    {{ t('admin.pricingAdvisor.subcol.p95') }}<span class="ml-0.5 text-gray-400">?</span>
                  </th>
                  <th class="cursor-help px-4 py-1.5 text-right text-[10px] font-medium text-gray-500 dark:text-dark-400" :title="t('admin.pricingAdvisor.tip.max')">
                    {{ t('admin.pricingAdvisor.subcol.max') }}<span class="ml-0.5 text-gray-400">?</span>
                  </th>
                  <th class="border-l border-gray-200/60 px-4 py-1.5 dark:border-dark-700/60"></th>
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
                  <td class="px-4 py-3 font-medium text-gray-900 dark:text-white">
                    <PlatformIcon :platform="row.platform as GroupPlatform" size="sm" class="mr-1.5 inline-block" />
                    {{ row.platform }}
                  </td>
                  <td class="px-4 py-3 text-center">
                    <span class="inline-flex items-center rounded-md bg-gray-100 px-2 py-0.5 text-xs font-medium text-gray-700 dark:bg-dark-700 dark:text-dark-200">
                      {{ formatTier(row.tier) }}
                    </span>
                  </td>
                  <td class="px-4 py-3 text-center tabular-nums">
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
                  <td class="border-l border-gray-100 px-4 py-3 text-right tabular-nums text-gray-900 dark:border-dark-700/60 dark:text-white">${{ row.window_5h_p50.toFixed(2) }}</td>
                  <td class="px-4 py-3 text-right tabular-nums text-gray-800 dark:text-dark-100">${{ row.window_5h_p95.toFixed(2) }}</td>
                  <td class="px-4 py-3 text-right tabular-nums text-gray-700 dark:text-dark-200">${{ row.window_5h_max.toFixed(2) }}</td>
                  <td class="border-l border-gray-100 px-4 py-3 text-right tabular-nums text-gray-900 dark:border-dark-700/60 dark:text-white">${{ row.window_7d_p50.toFixed(2) }}</td>
                  <td class="px-4 py-3 text-right tabular-nums text-gray-800 dark:text-dark-100">${{ row.window_7d_p95.toFixed(2) }}</td>
                  <td class="px-4 py-3 text-right tabular-nums text-gray-700 dark:text-dark-200">${{ row.window_7d_max.toFixed(2) }}</td>
                  <td class="border-l border-gray-100 px-4 py-3 text-right tabular-nums font-semibold text-emerald-700 dark:border-dark-700/60 dark:text-emerald-300">${{ row.daily_30d_avg.toFixed(2) }}</td>
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
              <!-- 5h / 7d 窗口 + 日均：三列 -->
              <div class="grid grid-cols-3 gap-2 text-center">
                <div class="rounded-lg bg-gray-50 p-2 dark:bg-dark-800/40">
                  <p class="text-[10px] font-medium uppercase tracking-wider text-gray-500 dark:text-dark-400">{{ t('admin.pricingAdvisor.col.5h') }} P50</p>
                  <p class="mt-0.5 text-sm font-semibold tabular-nums text-gray-900 dark:text-white">${{ row.window_5h_p50.toFixed(2) }}</p>
                  <p class="mt-0.5 text-[10px] tabular-nums text-gray-500 dark:text-dark-400">P95 ${{ row.window_5h_p95.toFixed(2) }}</p>
                </div>
                <div class="rounded-lg bg-gray-50 p-2 dark:bg-dark-800/40">
                  <p class="text-[10px] font-medium uppercase tracking-wider text-gray-500 dark:text-dark-400">{{ t('admin.pricingAdvisor.col.7d') }} P50</p>
                  <p class="mt-0.5 text-sm font-semibold tabular-nums text-gray-900 dark:text-white">${{ row.window_7d_p50.toFixed(2) }}</p>
                  <p class="mt-0.5 text-[10px] tabular-nums text-gray-500 dark:text-dark-400">P95 ${{ row.window_7d_p95.toFixed(2) }}</p>
                </div>
                <div class="rounded-lg bg-emerald-50/60 p-2 dark:bg-emerald-500/5">
                  <p class="text-[10px] font-medium uppercase tracking-wider text-emerald-600 dark:text-emerald-300">{{ t('admin.pricingAdvisor.col.dailyAvg') }}</p>
                  <p class="mt-0.5 text-sm font-semibold tabular-nums text-emerald-700 dark:text-emerald-300">${{ row.daily_30d_avg.toFixed(2) }}</p>
                </div>
              </div>
            </div>
          </div>
        </section>

        <!-- ROI / 限额计算器 + 趋势图 -->
        <div class="grid gap-4 lg:grid-cols-[minmax(0,1fr)_minmax(0,1.3fr)]">
          <section class="surface-card p-5">
            <header class="mb-4 flex items-center gap-3">
              <Icon name="calculator" size="sm" class="text-amber-500" />
              <h2 class="text-[15px] font-semibold text-gray-900 dark:text-white">{{ t('admin.pricingAdvisor.calculator.title') }}</h2>
            </header>
            <div class="space-y-4">
              <div>
                <label class="input-label">{{ t('admin.pricingAdvisor.calculator.tier') }}</label>
                <Select v-model="calcTierKey" :options="calcTierOptions" />
              </div>
              <div class="grid grid-cols-2 gap-3">
                <div>
                  <label class="input-label">{{ t('admin.pricingAdvisor.calculator.cost') }}</label>
                  <input v-model.number="calcCost" type="number" min="0" step="1" class="input" placeholder="1400" />
                </div>
                <div>
                  <label class="input-label cursor-help" :title="t('admin.pricingAdvisor.calculator.usersPerAccountTip')">
                    {{ t('admin.pricingAdvisor.calculator.usersPerAccount') }}
                    <Icon name="infoCircle" size="xs" class="ml-0.5 inline-block text-gray-400" />
                  </label>
                  <input v-model.number="calcUsersPerAccount" type="number" min="1" step="1" class="input" placeholder="5" />
                </div>
              </div>
              <div class="grid grid-cols-3 gap-3">
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
                <div>
                  <label class="input-label cursor-help" :title="t('admin.pricingAdvisor.calculator.occupancyTip')">
                    {{ t('admin.pricingAdvisor.calculator.occupancy') }}
                    <Icon name="infoCircle" size="xs" class="ml-0.5 inline-block text-gray-400" />
                  </label>
                  <div class="flex items-center gap-1.5">
                    <input v-model.number="calcOccupancy" type="number" min="10" max="100" step="5" class="input flex-1" placeholder="80" />
                    <span class="text-sm text-gray-500 dark:text-dark-400">%</span>
                  </div>
                </div>
                <div>
                  <label class="input-label cursor-help" :title="t('admin.pricingAdvisor.calculator.safetyTip')">
                    {{ t('admin.pricingAdvisor.calculator.safety') }}
                    <Icon name="infoCircle" size="xs" class="ml-0.5 inline-block text-gray-400" />
                  </label>
                  <input v-model.number="calcSafety" type="number" min="0.3" max="1.5" step="0.1" class="input" placeholder="0.8" />
                </div>
              </div>

              <!-- 结果区：上层主结果（建议月费 + 期望月利润）；下层每用户 5h/日/周/月限额（反推自上游 cap） -->
              <div v-if="calcResult" class="mt-4 space-y-3">
                <!-- 主结果：建议月费 + 预计月利润，两个左右并排大字号居中 -->
                <div class="grid grid-cols-2 gap-3">
                  <div class="rounded-xl border border-emerald-200/60 bg-emerald-50/60 p-4 text-center dark:border-emerald-500/20 dark:bg-emerald-500/5">
                    <p class="text-[11px] font-medium uppercase tracking-wider text-emerald-700 dark:text-emerald-300">{{ t('admin.pricingAdvisor.calculator.suggestedPrice') }}</p>
                    <p class="mt-1 flex items-baseline justify-center gap-1">
                      <span class="text-xl font-semibold text-emerald-700 dark:text-emerald-300">¥</span>
                      <span class="text-3xl font-bold tabular-nums tracking-tight text-emerald-900 dark:text-emerald-100">{{ calcResult.priceCny.toFixed(0) }}</span>
                      <span class="text-xs text-emerald-700 dark:text-emerald-300">/{{ t('admin.pricingAdvisor.calculator.perMonth') }}</span>
                    </p>
                  </div>
                  <div class="rounded-xl border p-4 text-center"
                    :class="calcResult.monthlyProfitExpectedCny >= 0
                      ? 'border-amber-200/60 bg-amber-50/60 dark:border-amber-500/20 dark:bg-amber-500/5'
                      : 'border-rose-200/60 bg-rose-50/60 dark:border-rose-500/20 dark:bg-rose-500/5'"
                  >
                    <p class="cursor-help text-[11px] font-medium uppercase tracking-wider"
                      :class="calcResult.monthlyProfitExpectedCny >= 0
                        ? 'text-amber-700 dark:text-amber-300'
                        : 'text-rose-700 dark:text-rose-300'"
                      :title="t('admin.pricingAdvisor.calculator.profitTip', { full: calcResult.monthlyProfitFullCny.toFixed(0) })"
                    >
                      {{ t('admin.pricingAdvisor.calculator.profit') }}
                      <Icon name="infoCircle" size="xs" class="ml-0.5 inline-block opacity-70" />
                    </p>
                    <p class="mt-1 flex items-baseline justify-center gap-1"
                      :class="calcResult.monthlyProfitExpectedCny >= 0
                        ? 'text-amber-900 dark:text-amber-100'
                        : 'text-rose-900 dark:text-rose-100'"
                    >
                      <span class="text-xl font-semibold"
                        :class="calcResult.monthlyProfitExpectedCny >= 0
                          ? 'text-amber-700 dark:text-amber-300'
                          : 'text-rose-700 dark:text-rose-300'"
                      >¥</span>
                      <span class="text-3xl font-bold tabular-nums tracking-tight">{{ calcResult.monthlyProfitExpectedCny.toFixed(0) }}</span>
                      <span class="text-xs"
                        :class="calcResult.monthlyProfitExpectedCny >= 0
                          ? 'text-amber-700 dark:text-amber-300'
                          : 'text-rose-700 dark:text-rose-300'"
                      >/{{ t('admin.pricingAdvisor.calculator.perMonth') }}</span>
                    </p>
                  </div>
                </div>

                <!-- 限额区：5h / 日 / 周 / 月，反推自上游 cap，按 N 个用户均分 + 安全系数 -->
                <div class="rounded-xl border border-gray-200/70 bg-gray-50/60 p-3 dark:border-dark-700/60 dark:bg-dark-800/40">
                  <p class="mb-2 cursor-help text-center text-[11px] font-medium uppercase tracking-wider text-gray-500 dark:text-dark-400"
                    :title="t('admin.pricingAdvisor.calculator.limitsTip')">
                    {{ t('admin.pricingAdvisor.calculator.limitsTitle') }}
                    <Icon name="infoCircle" size="xs" class="ml-0.5 inline-block text-gray-400" />
                  </p>
                  <div class="grid grid-cols-4 gap-2 text-center">
                    <div>
                      <p class="text-[10px] text-gray-500 dark:text-dark-400">{{ t('admin.pricingAdvisor.calculator.suggested5h') }}</p>
                      <p class="mt-0.5 text-sm font-semibold tabular-nums text-gray-900 dark:text-white">${{ calcResult.fiveHourLimitUsd.toFixed(2) }}</p>
                    </div>
                    <div>
                      <p class="text-[10px] text-gray-500 dark:text-dark-400">{{ t('admin.pricingAdvisor.calculator.suggestedDaily') }}</p>
                      <p class="mt-0.5 text-sm font-semibold tabular-nums text-gray-900 dark:text-white">${{ calcResult.dailyLimitUsd.toFixed(2) }}</p>
                    </div>
                    <div>
                      <p class="text-[10px] text-gray-500 dark:text-dark-400">{{ t('admin.pricingAdvisor.calculator.suggestedWeekly') }}</p>
                      <p class="mt-0.5 text-sm font-semibold tabular-nums text-gray-900 dark:text-white">${{ calcResult.weeklyLimitUsd.toFixed(2) }}</p>
                    </div>
                    <div>
                      <p class="text-[10px] text-gray-500 dark:text-dark-400">{{ t('admin.pricingAdvisor.calculator.suggestedMonthly') }}</p>
                      <p class="mt-0.5 text-sm font-semibold tabular-nums text-gray-900 dark:text-white">${{ calcResult.monthlyLimitUsd.toFixed(2) }}</p>
                    </div>
                  </div>
                  <!-- 数据来源说明：让 admin 看出 cap 是反推还是回退、分别是多少 -->
                  <p class="mt-2 text-center text-[10px] text-gray-500 dark:text-dark-400">
                    {{ calcResult.capSource === 'reversed'
                      ? t('admin.pricingAdvisor.calculator.limitsSourceReversed', {
                          cap5h: calcResult.cap5hUsd.toFixed(2),
                          cap7d: calcResult.cap7dUsd.toFixed(2),
                          samples: calcResult.capSampleCount,
                          n: calcResult.n,
                          safety: calcResult.safety,
                        })
                      : t('admin.pricingAdvisor.calculator.limitsSourceFallback', {
                          cap5h: calcResult.cap5hUsd.toFixed(2),
                          cap7d: calcResult.cap7dUsd.toFixed(2),
                          n: calcResult.n,
                          safety: calcResult.safety,
                        })
                    }}
                  </p>
                  <!-- 4 条限额是独立约束，提醒 admin 实际生效是最严的那条 -->
                  <p class="mt-1 text-center text-[10px] italic text-gray-400 dark:text-dark-500">
                    {{ t('admin.pricingAdvisor.calculator.limitsCapsIndependentNote') }}
                  </p>
                </div>

                <!-- 一键创建套餐按钮 -->
                <button
                  type="button"
                  class="btn btn-primary w-full"
                  @click="applyToPlan"
                >
                  <Icon name="plus" size="sm" class="mr-1.5" />
                  {{ t('admin.pricingAdvisor.calculator.applyToPlan') }}
                </button>

                <p v-if="calcResult.warning" class="flex items-start gap-1.5 rounded-md bg-amber-50 px-2.5 py-1.5 text-[11px] text-amber-700 dark:bg-amber-500/10 dark:text-amber-300">
                  <Icon name="exclamationTriangle" size="xs" class="mt-0.5 shrink-0" />
                  <span>{{ calcResult.warning }}</span>
                </p>
              </div>
              <p v-else class="rounded-xl bg-gray-50 px-4 py-3 text-xs text-gray-500 dark:bg-dark-800/40 dark:text-dark-400">
                {{ t('admin.pricingAdvisor.calculator.selectFirst') }}
              </p>
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
    },
  })
}

function formatTier(tier: string): string {
  if (!tier) return t('admin.accounts.tierUnclassified')
  // 把 underscored key 转成人类友好显示：pro_5x → Pro 5x
  return tier.split('_').map((w) => w.charAt(0).toUpperCase() + w.slice(1)).join(' ')
}

// ── ROI 计算器 ──
// 计算逻辑：从档位反推上游 cap（cap = 当前窗口已耗费 / utilization%），按 N 个用户均分；
// 没有 cap 样本时回退到 7d_P95 估计（peak 接近 cap）。详见 calcResult。
const calcCost = ref<number>(1400)
const calcUsersPerAccount = ref<number>(5) // 单账号承载用户数 N
const calcMarkup = ref<number>(30) // 期望加价率 % (profit/cost)
const calcSafety = ref<number>(0.8)
const calcOccupancy = ref<number>(80) // 上座率 %（卖出去几成座位才算盈亏平衡）
const calcTierKey = ref<string>('') // 跟 selectedKey 联动

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

  const safety = Math.max(0.3, Math.min(1.5, calcSafety.value || 0.8))
  const cost = Math.max(0, calcCost.value || 0)
  const markup = Math.max(0, Math.min(500, calcMarkup.value || 0))
  const n = Math.max(1, Math.floor(calcUsersPerAccount.value || 1))
  const occupancy = Math.max(1, Math.min(100, calcOccupancy.value || 100)) / 100

  // 1) 优先用反推 cap（cap = 已耗费 / utilization）。无可用 util 样本时回退用 7d_P95（接近 cap）。
  //    用 P95 而不是 P50：P50 是中位账号峰值，会低估真实 cap；P95 是 95% 账号能跑到的水平，更接近上游硬上限。
  const useReversedCap = tier.cap_sample_count > 0 && tier.cap_5h_usd > 0 && tier.cap_7d_usd > 0
  const cap5hUsd = useReversedCap ? tier.cap_5h_usd : tier.window_5h_p95
  const cap7dUsd = useReversedCap ? tier.cap_7d_usd : tier.window_7d_p95
  const capSource: 'reversed' | 'fallback' = useReversedCap ? 'reversed' : 'fallback'

  // 2) 限额按 N 均分：
  //    - 5h 限额 = cap_5h × safety / N
  //    - 周限额 = cap_7d × safety / N
  //    - 日限额 = 周限额 / 7（用户希望的均匀分布）
  //    - 月限额 = 周限额 × 4（用户给的简化系数；理论 30/7≈4.286，× 4 偏保守不易超卖）
  const fiveHourLimitUsd = (cap5hUsd * safety) / n
  const weeklyLimitUsd = (cap7dUsd * safety) / n
  const dailyLimitUsd = weeklyLimitUsd / 7
  const monthlyLimitUsd = weeklyLimitUsd * 4

  // 3) 建议月费：基于"加价率"（profit/cost），不是利润率。例 cost=¥1400, markup=30% → revenue=¥1820
  //    考虑上座率：实际只有 N×occupancy 个座位真正卖出，所以 raw = revenue / (N × occupancy)
  const totalRevenueCny = cost * (1 + markup / 100)
  const rawPriceCny = totalRevenueCny / Math.max(1, n * occupancy)
  const priceCny = roundPriceCny(rawPriceCny)

  // 4) 利润分两种口径：
  //    - 满座月利润 = price × N - cost（最理想）
  //    - 期望月利润 = price × (N × occupancy) - cost（按上座率）
  const monthlyProfitFullCny = priceCny * n - cost
  const monthlyProfitExpectedCny = priceCny * Math.floor(n * occupancy) - cost

  // 5) 风险提示
  let warning = ''
  if (!tier.has_enough_samples) {
    warning = t('admin.pricingAdvisor.calculator.warnSamplesLow')
  } else if (capSource === 'fallback') {
    warning = t('admin.pricingAdvisor.calculator.warnNoCapSample')
  } else if (n > 50) {
    warning = t('admin.pricingAdvisor.calculator.warnNTooHigh')
  } else if (monthlyProfitExpectedCny < 0) {
    warning = t('admin.pricingAdvisor.calculator.warnUnprofitable')
  }

  return {
    priceCny,
    monthlyLimitUsd,
    weeklyLimitUsd,
    dailyLimitUsd,
    fiveHourLimitUsd,
    cap5hUsd,
    cap7dUsd,
    capSource,
    capSampleCount: tier.cap_sample_count,
    monthlyProfitFullCny,
    monthlyProfitExpectedCny,
    warning,
    n,
    safety,
    occupancy,
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
