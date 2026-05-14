<template>
  <AppLayout wide>
    <div class="payment-page space-y-5">
      <div v-if="loading" class="flex items-center justify-center py-20">
        <LoadingSpinner size="md" />
      </div>
      <template v-else>
        <!-- Tab Switcher：Notion 风下划线分段，黑色 indicator + 13px 字号 -->
        <div v-if="tabs.length > 1 && paymentPhase === 'select' && !selectedPlan" class="border-b border-gray-200/70 dark:border-dark-700/60">
          <div class="mx-auto flex max-w-xl items-center gap-1">
            <button v-for="tab in tabs" :key="tab.key"
              type="button"
              class="relative flex items-center justify-center gap-1.5 px-5 py-3 text-[13px] font-medium transition-colors -mb-px border-b-2"
              :class="activeTab === tab.key
                ? 'border-gray-900 text-gray-900 dark:border-white dark:text-white'
                : 'border-transparent text-gray-500 hover:text-gray-900 dark:text-dark-400 dark:hover:text-white'"
              @click="activeTab = tab.key">
              <Icon :name="tab.key === 'subscription' ? 'badge' : 'creditCard'" size="sm" :stroke-width="2" />
              {{ tab.label }}
            </button>
          </div>
        </div>
        <!-- Payment in progress (shared by recharge and subscription) -->
        <template v-if="paymentPhase === 'paying'">
          <PaymentStatusPanel
            :order-id="paymentState.orderId"
            :qr-code="paymentState.qrCode"
            :qr-code-image="paymentState.qrCodeImage"
            :expires-at="paymentState.expiresAt"
            :payment-type="paymentState.paymentType"
            :pay-url="paymentState.payUrl"
            :order-type="paymentState.orderType"
            @done="onPaymentDone"
            @success="onPaymentSuccess"
            @settled="onPaymentSettled"
          />
        </template>
        <!-- Tab content (select phase) -->
        <template v-else>
          <!-- Top-up Tab -->
          <template v-if="activeTab === 'recharge'">
            <EmptyState
              v-if="enabledMethods.length === 0"
              variant="amber"
              :description="t('payment.notAvailable')"
            >
              <template #icon>
                <Icon name="creditCard" class="empty-state-icon" />
              </template>
            </EmptyState>
            <template v-else>
            <!-- 充值主卡：白底简洁 + 右侧 amber 结算栏（视觉聚焦） -->
            <section class="overflow-hidden rounded-2xl border border-gray-200/70 bg-white shadow-[0_1px_2px_rgba(15,23,42,0.04),0_8px_24px_-18px_rgba(15,23,42,0.18)] dark:border-dark-700/60 dark:bg-dark-800/40">
              <!-- md 起就用左右两栏（之前 lg 才触发，平板上单列拉伸过宽）-->
              <div class="grid md:grid-cols-[minmax(0,1fr)_300px] lg:grid-cols-[minmax(0,1fr)_340px]">
                <!-- 左：表单区（白底） -->
                <div class="space-y-5 p-5 sm:p-6">
                  <!-- 顶部：充值账户信息 + 钱包图标 -->
                  <div class="flex items-center justify-between gap-4">
                    <div class="flex items-center gap-3">
                      <div class="flex h-11 w-11 items-center justify-center rounded-xl bg-amber-50 text-amber-600 ring-1 ring-inset ring-amber-200/70 dark:bg-amber-500/15 dark:text-amber-300 dark:ring-amber-500/30">
                        <Icon name="creditCard" size="md" />
                      </div>
                      <div>
                        <p class="text-[12px] font-medium text-gray-500 dark:text-dark-400">{{ t('payment.rechargeAccount') }}</p>
                        <p class="text-sm font-semibold text-gray-900 dark:text-white">{{ user?.username || '' }}</p>
                      </div>
                    </div>
                    <!-- 当前余额：右上紧凑 chip -->
                    <div class="rounded-xl bg-emerald-50 px-3 py-2 ring-1 ring-inset ring-emerald-200/70 dark:bg-emerald-500/15 dark:ring-emerald-500/30">
                      <p class="text-[10px] font-medium text-emerald-700/70 dark:text-emerald-300/70">{{ t('payment.currentBalance') }}</p>
                      <p class="text-lg font-semibold tabular-nums leading-tight text-emerald-700 dark:text-emerald-300">${{ user?.balance?.toFixed(2) || '0.00' }}</p>
                    </div>
                  </div>

                  <AmountInput
                    v-model="amount"
                    :amounts="quickAmounts"
                    :min="globalMinAmount"
                    :max="globalMaxAmount"
                  />
                  <p v-if="amountError" class="flex items-start gap-2 rounded-xl border border-rose-200/70 bg-rose-50/60 px-3 py-2 text-xs text-rose-700 dark:border-rose-500/30 dark:bg-rose-500/10 dark:text-rose-300">
                    <Icon name="exclamationTriangle" size="xs" class="mt-0.5 shrink-0" />
                    <span>{{ amountError }}</span>
                  </p>

                  <PaymentMethodSelector
                    :methods="methodOptions"
                    :selected="selectedMethod"
                    @select="selectedMethod = $event"
                  />
                </div>

                <!-- 右：结算栏（白底为主 + 极淡 amber 装饰 + amber 按钮强调） -->
                <aside class="relative border-t border-gray-100 bg-gray-50/40 p-6 dark:border-dark-700/60 dark:bg-dark-800/30 md:border-l md:border-t-0">
                  <!-- 实付金额：左侧 amber 强调条 + 金额放大到 5xl，建立明确视觉锚点 -->
                  <div class="relative pl-4">
                    <span class="absolute left-0 top-1 h-[calc(100%-0.5rem)] w-[3px] rounded-full bg-gradient-to-b from-amber-400 to-amber-500" aria-hidden="true"></span>
                    <p class="text-[11px] font-medium uppercase tracking-wider text-gray-500 dark:text-dark-400">
                      {{ t('payment.actualPay') }}
                    </p>
                    <p class="mt-1.5 flex items-baseline gap-1 text-gray-900 dark:text-white">
                      <span class="text-xl font-semibold tabular-nums leading-none text-gray-500 dark:text-dark-400">¥</span>
                      <span class="text-5xl font-bold tabular-nums tracking-tight leading-none">{{ totalAmount.toFixed(2) }}</span>
                    </p>
                  </div>

                  <!-- 费用明细：极简列表 -->
                  <div class="mt-5 space-y-2 border-t border-gray-100 pt-4 text-sm dark:border-dark-700/60">
                    <div class="flex justify-between">
                      <span class="text-gray-500 dark:text-dark-400">{{ t('payment.paymentAmount') }}</span>
                      <span class="font-medium tabular-nums text-gray-900 dark:text-white">¥{{ validAmount.toFixed(2) }}</span>
                    </div>
                    <div v-if="feeRate > 0" class="flex justify-between">
                      <span class="text-gray-500 dark:text-dark-400">{{ t('payment.fee') }} ({{ feeRate }}%)</span>
                      <span class="font-medium tabular-nums text-gray-900 dark:text-white">¥{{ feeAmount.toFixed(2) }}</span>
                    </div>
                    <div v-if="balanceRechargeMultiplier !== 1" class="flex justify-between">
                      <span class="text-gray-500 dark:text-dark-400">{{ t('payment.creditedBalance') }}</span>
                      <span class="font-medium tabular-nums text-emerald-600 dark:text-emerald-400">${{ creditedAmount.toFixed(2) }}</span>
                    </div>
                  </div>

                  <p v-if="balanceRechargeMultiplier !== 1" class="mt-3 text-[11px] leading-relaxed text-gray-500 dark:text-dark-400">
                    {{ t('payment.rechargeRatePreview', { usd: balanceRechargeMultiplier.toFixed(2) }) }}
                  </p>

                  <!-- 确认支付按钮：gray-900 实色（专业稳重）+ hover amber 提示 -->
                  <button
                    class="mt-6 w-full rounded-xl bg-gray-900 px-4 py-3 text-sm font-semibold text-white shadow-[0_4px_12px_-2px_rgba(15,23,42,0.25)] transition-all hover:bg-gray-800 hover:shadow-[0_8px_20px_-4px_rgba(15,23,42,0.35)] disabled:cursor-not-allowed disabled:bg-gray-200 disabled:text-gray-400 disabled:shadow-none dark:bg-white dark:text-gray-900 dark:hover:bg-gray-100 dark:disabled:bg-dark-700 dark:disabled:text-dark-400"
                    :disabled="!canSubmit || submitting"
                    @click="handleSubmitRecharge"
                  >
                    <span v-if="submitting" class="flex items-center justify-center gap-2">
                      <span class="h-4 w-4 animate-spin rounded-full border-2 border-current border-t-transparent opacity-60"></span>
                      {{ t('common.processing') }}
                    </span>
                    <span v-else class="flex items-center justify-center gap-2">
                      {{ t('payment.createOrder') }}
                      <Icon name="arrowRight" size="sm" :stroke-width="2.5" />
                    </span>
                  </button>
                </aside>
              </div>
            </section>
            </template>
          </template>
          <!-- Subscribe Tab -->
          <template v-else-if="activeTab === 'subscription'">
            <!-- Subscription confirm (inline, replaces plan list) -->
            <template v-if="selectedPlan">
              <!-- 续费提示横幅：从「我的独享号」跳转过来时显示 -->
              <!-- 设计原则：主体中性 + 左侧 4px 强调条；价格对比区强化字号 hierarchy 让新价成为视觉中心 -->
              <div v-if="pendingRenewSeatId > 0"
                class="overflow-hidden rounded-2xl border bg-white shadow-sm shadow-gray-100 dark:bg-dark-900 dark:shadow-none"
                :class="renewBannerOuterClass">
                <!-- 顶部：图标 + 标题 + hint。hint 文字色随趋势走，让用户一眼看出涨/降态 -->
                <div class="flex items-start gap-3.5 px-5 py-4">
                  <div class="flex h-11 w-11 shrink-0 items-center justify-center rounded-xl"
                    :class="renewBannerIconBoxClass">
                    <Icon :name="renewPriceTrend === 'up' ? 'exclamationTriangle' : 'refresh'" size="md" :stroke-width="2.2" />
                  </div>
                  <div class="flex-1 min-w-0">
                    <p class="text-sm font-semibold text-gray-900 dark:text-gray-50">
                      {{ t('payment.renewalBanner.title') }}
                    </p>
                    <!-- min-h-[2lh] 锁定 2 行高度，避免三态 hint 长短不一导致整 banner 顶部参差 -->
                    <p class="mt-1 text-xs leading-relaxed min-h-[2lh]" :class="renewBannerHintClass">
                      <span v-if="renewPriceTrend === 'up'">{{ t('payment.renewalBanner.hintUp') }}</span>
                      <span v-else-if="renewPriceTrend === 'down'">{{ t('payment.renewalBanner.hintDown') }}</span>
                      <span v-else>{{ t('payment.renewalBanner.hint') }}</span>
                    </p>
                  </div>
                </div>

                <!-- 价格对比卡片：3 列同基线对齐；新价右侧紧贴 delta chip（一行内，不再上下分布）-->
                <div v-if="renewLastPaidPrice > 0 && selectedPlan"
                  class="border-t border-gray-100 bg-gradient-to-b from-gray-50/40 to-gray-50/80 px-5 py-4 dark:border-dark-700 dark:from-dark-800/30 dark:to-dark-800/60">
                  <!-- label 行（统一在金额上方） -->
                  <div class="grid grid-cols-[1fr_auto_1fr] gap-4">
                    <div class="text-right text-[10px] font-medium uppercase tracking-wider text-gray-400 dark:text-dark-400">
                      {{ t('payment.renewalBanner.lastPaid') }}
                    </div>
                    <div></div>
                    <div class="text-left text-[10px] font-medium uppercase tracking-wider text-gray-400 dark:text-dark-400">
                      {{ t('payment.renewalBanner.thisTime') }}
                    </div>
                  </div>
                  <!-- 金额行：旧价 → 大圆趋势锚点 → 新价+chip，items-center 让圆和金额完美对齐 -->
                  <div class="mt-1.5 grid grid-cols-[1fr_auto_1fr] items-center gap-4">
                    <!-- 旧价 -->
                    <div class="text-right font-mono text-base font-medium tabular-nums leading-none"
                      :class="renewPriceTrend === 'same'
                        ? 'text-gray-700 dark:text-dark-100'
                        : 'text-gray-400 line-through decoration-2 decoration-gray-300/80 dark:text-dark-400 dark:decoration-dark-600/80'">
                      ¥{{ renewLastPaidPrice.toFixed(2) }}
                    </div>

                    <!-- 中央趋势锚点：实色 + ring 光晕 -->
                    <div class="flex h-11 w-11 items-center justify-center rounded-full"
                      :class="renewTrendDotClass">
                      <svg v-if="renewPriceTrend === 'up'" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor" class="h-5 w-5">
                        <path fill-rule="evenodd" d="M10 17a.75.75 0 01-.75-.75V5.612L5.29 9.77a.75.75 0 11-1.08-1.04l5.25-5.5a.75.75 0 011.08 0l5.25 5.5a.75.75 0 11-1.08 1.04L10.75 5.612V16.25A.75.75 0 0110 17z" clip-rule="evenodd" />
                      </svg>
                      <svg v-else-if="renewPriceTrend === 'down'" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor" class="h-5 w-5">
                        <path fill-rule="evenodd" d="M10 3a.75.75 0 01.75.75v10.638l3.96-4.158a.75.75 0 111.08 1.04l-5.25 5.5a.75.75 0 01-1.08 0l-5.25-5.5a.75.75 0 111.08-1.04l3.96 4.158V3.75A.75.75 0 0110 3z" clip-rule="evenodd" />
                      </svg>
                      <svg v-else xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor" class="h-5 w-5">
                        <path fill-rule="evenodd" d="M3 10a.75.75 0 01.75-.75h12.5a.75.75 0 010 1.5H3.75A.75.75 0 013 10z" clip-rule="evenodd" />
                      </svg>
                    </div>

                    <!-- 新价：大字号 + 紧邻 delta chip 做视觉锚点群 -->
                    <div class="flex items-baseline gap-2">
                      <span class="font-mono text-2xl font-bold tabular-nums leading-none" :class="renewPriceTextClass">
                        ¥{{ selectedPlan.price.toFixed(2) }}
                      </span>
                      <span v-if="renewPriceTrend === 'up'"
                        class="inline-flex items-center rounded-md bg-amber-100/80 px-1.5 py-0.5 text-[10px] font-bold tabular-nums text-amber-700 dark:bg-amber-900/40 dark:text-amber-300">
                        +¥{{ (selectedPlan.price - renewLastPaidPrice).toFixed(2) }}
                      </span>
                      <span v-else-if="renewPriceTrend === 'down'"
                        class="inline-flex items-center rounded-md bg-emerald-100/80 px-1.5 py-0.5 text-[10px] font-bold tabular-nums text-emerald-700 dark:bg-emerald-900/40 dark:text-emerald-300">
                        −¥{{ (renewLastPaidPrice - selectedPlan.price).toFixed(2) }}
                      </span>
                      <span v-else
                        class="inline-flex items-center rounded-md bg-gray-100 px-1.5 py-0.5 text-[10px] font-medium text-gray-500 dark:bg-dark-700 dark:text-dark-300">
                        {{ t('payment.renewalBanner.priceSame') }}
                      </span>
                    </div>
                  </div>
                </div>
              </div>
              <section :class="['overflow-hidden rounded-2xl border bg-white shadow-[0_1px_2px_rgba(15,23,42,0.04)] dark:bg-dark-800/40 dark:shadow-none', planBorderClass]">
                <div :class="['h-1.5', planAccentClass]" />
                <div class="grid gap-0 md:grid-cols-[minmax(0,1fr)_280px] lg:grid-cols-[minmax(0,1fr)_320px]">
                  <div class="p-5 sm:p-6">
                    <div class="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
                      <div class="flex min-w-0 items-start gap-3">
                        <span class="flex h-12 w-12 shrink-0 items-center justify-center rounded-2xl bg-white shadow-sm ring-1 ring-inset ring-gray-200/70 dark:bg-dark-800 dark:ring-dark-700/60">
                          <BrandIcon
                            v-if="selectedPlanPlatformBrand"
                            :brand="selectedPlanPlatformBrand"
                            size="25px"
                          />
                          <span v-else class="text-sm font-bold text-gray-500 dark:text-gray-300">{{ selectedPlanPlatformInitial }}</span>
                        </span>
                        <div class="min-w-0">
                          <div class="flex flex-wrap items-center gap-2">
                            <h3 class="text-lg font-semibold tracking-tight text-gray-900 dark:text-white">{{ selectedPlan.name }}</h3>
                            <span :class="['rounded-full px-2 py-0.5 text-[11px] font-medium', planBadgeLightClass]">
                              {{ platformLabel(selectedPlan.group_platform || '') }}
                            </span>
                          </div>
                          <p v-if="selectedPlan.description" class="mt-1.5 max-w-2xl text-sm leading-relaxed text-gray-500 dark:text-gray-400">
                            {{ selectedPlan.description }}
                          </p>
                        </div>
                      </div>
                    </div>

                    <div class="mt-5 rounded-2xl border border-gray-200/70 bg-white p-3 dark:border-dark-700/60 dark:bg-dark-800/30">
                      <div class="grid gap-2 sm:grid-cols-2 xl:grid-cols-4">
                      <div class="rounded-xl bg-gray-50/70 px-4 py-3 dark:bg-dark-800/40">
                        <span class="block text-[12px] font-medium text-gray-500 dark:text-dark-400">{{ t('payment.planCard.rate') }}</span>
                        <span :class="['mt-1 block text-lg font-semibold tabular-nums tracking-tight', planTextClass]">{{ selectedPlanRateDisplay }}</span>
                      </div>
                      <!-- 限额展示规则（见 utils/planLimits.ts）：
                           - value > 0 → 具体额度
                           - value <= 0 → 绿色"无限制"
                           - 被更紧的限额覆盖（如 weekly >= daily × 7）→ 自动隐藏，不放出来误导用户 -->
                      <div v-if="selectedPlanLimitVisibility.showDaily" class="rounded-xl bg-gray-50/70 px-4 py-3 dark:bg-dark-800/40">
                        <span class="block text-[12px] font-medium text-gray-500 dark:text-dark-400">{{ t('payment.planCard.dailyLimit') }}</span>
                        <span v-if="(selectedPlan.daily_limit_usd ?? 0) > 0" class="mt-1 block text-lg font-semibold tabular-nums tracking-tight text-gray-900 dark:text-white">${{ selectedPlan.daily_limit_usd }}</span>
                        <span v-else class="mt-1 block text-lg font-semibold tracking-tight text-emerald-600 dark:text-emerald-400">{{ t('payment.planCard.unlimited') }}</span>
                      </div>
                      <div v-if="selectedPlanLimitVisibility.showWeekly" class="rounded-xl bg-gray-50/70 px-4 py-3 dark:bg-dark-800/40">
                        <span class="block text-[12px] font-medium text-gray-500 dark:text-dark-400">{{ t('payment.planCard.weeklyLimit') }}</span>
                        <span v-if="(selectedPlan.weekly_limit_usd ?? 0) > 0" class="mt-1 block text-lg font-semibold tabular-nums tracking-tight text-gray-900 dark:text-white">${{ selectedPlan.weekly_limit_usd }}</span>
                        <span v-else class="mt-1 block text-lg font-semibold tracking-tight text-emerald-600 dark:text-emerald-400">{{ t('payment.planCard.unlimited') }}</span>
                      </div>
                      <div v-if="selectedPlanLimitVisibility.showMonthly" class="rounded-xl bg-gray-50/70 px-4 py-3 dark:bg-dark-800/40">
                        <span class="block text-[12px] font-medium text-gray-500 dark:text-dark-400">{{ t('payment.planCard.monthlyLimit') }}</span>
                        <span v-if="(selectedPlan.monthly_limit_usd ?? 0) > 0" class="mt-1 block text-lg font-semibold tabular-nums tracking-tight text-gray-900 dark:text-white">${{ selectedPlan.monthly_limit_usd }}</span>
                        <span v-else class="mt-1 block text-lg font-semibold tracking-tight text-emerald-600 dark:text-emerald-400">{{ t('payment.planCard.unlimited') }}</span>
                      </div>
                      <div v-if="selectedPlan.daily_limit_usd == null && selectedPlan.weekly_limit_usd == null && selectedPlan.monthly_limit_usd == null" class="rounded-xl bg-gray-50/70 px-4 py-3 dark:bg-dark-800/40">
                        <span class="block text-[12px] font-medium text-gray-500 dark:text-dark-400">{{ t('payment.planCard.quota') }}</span>
                        <span class="mt-1 block text-lg font-semibold tracking-tight text-emerald-600 dark:text-emerald-400">{{ t('payment.planCard.unlimited') }}</span>
                      </div>
                      </div>

                      <div v-if="selectedPlanModelScopeItems.length > 0" class="mt-2 rounded-xl bg-gray-50/70 px-4 py-3 dark:bg-dark-800/40">
                        <div class="mb-2 flex items-center justify-between gap-3">
                          <span class="text-[12px] font-medium text-gray-500 dark:text-dark-400">{{ t('payment.planCard.models') }}</span>
                          <span class="h-px flex-1 bg-gray-200/70 dark:bg-dark-700/60" />
                        </div>
                        <div class="flex flex-wrap gap-1.5">
                        <span
                          v-for="scope in selectedPlanModelScopeItems"
                          :key="scope.key"
                          class="inline-flex items-center gap-1.5 rounded-full bg-white px-2.5 py-1 text-xs font-medium text-gray-700 ring-1 ring-inset ring-gray-200/70 dark:bg-dark-700/40 dark:text-gray-200 dark:ring-dark-600/60"
                        >
                          <ModelIcon :model="scope.iconModel" size="14px" />
                          {{ scope.label }}
                        </span>
                        </div>
                      </div>
                    </div>

                    <div v-if="selectedPlan.features.length > 0" class="mt-4 grid gap-2 sm:grid-cols-2">
                      <div v-for="feature in selectedPlan.features" :key="feature" class="flex items-start gap-2 rounded-xl bg-white px-3 py-2 text-sm text-gray-600 ring-1 ring-inset ring-gray-200/70 dark:bg-dark-800/40 dark:text-gray-300 dark:ring-dark-700/60">
                        <Icon name="check" size="sm" :class="planIconClass" />
                        <span>{{ feature }}</span>
                      </div>
                    </div>
                  </div>

                  <aside class="border-t border-gray-100 bg-gray-50/60 p-5 dark:border-dark-700/60 dark:bg-dark-800/30 lg:border-l lg:border-t-0">
                    <div class="rounded-2xl border border-gray-200/70 bg-white px-4 py-5 text-center dark:border-dark-700/60 dark:bg-dark-800/40">
                      <!-- 订阅场景下用"套餐金额"更准确，"充值金额"是余额充值的措辞 -->
                      <p class="text-[12px] font-medium text-gray-500 dark:text-dark-400">{{ t('payment.planAmount') }}</p>
                      <div class="mt-2 flex items-end justify-center gap-1.5 whitespace-nowrap">
                        <span :class="['mb-1.5 text-2xl font-semibold leading-none', planTextClass]">¥</span>
                        <span :class="['text-5xl font-semibold leading-none tracking-tight tabular-nums', planTextClass]">{{ selectedPlan.price }}</span>
                      </div>
                      <div class="mt-3 flex flex-wrap items-center justify-center gap-2">
                        <span class="inline-flex items-center gap-1.5 whitespace-nowrap rounded-full bg-gray-50 px-2.5 py-1 text-xs font-medium text-gray-600 ring-1 ring-inset ring-gray-200/70 dark:bg-dark-800/60 dark:text-dark-200 dark:ring-dark-700/60">
                          <Icon name="calendar" size="xs" :stroke-width="2" />
                          {{ planValiditySuffix }}
                        </span>
                        <span v-if="selectedPlan.original_price" class="inline-flex items-center gap-2">
                          <span class="text-xs text-gray-400 line-through dark:text-gray-500">¥{{ selectedPlan.original_price }}</span>
                          <span v-if="selectedPlanDiscountText" :class="['inline-flex rounded-full px-2 py-0.5 text-[11px] font-semibold', selectedPlanDiscountClass]">
                            {{ selectedPlanDiscountText }}
                          </span>
                        </span>
                      </div>
                    </div>

                    <div class="mt-4 space-y-3 rounded-2xl border border-gray-200/70 bg-white p-4 text-sm dark:border-dark-700/60 dark:bg-dark-800/40">
                      <div class="flex justify-between gap-4">
                        <span class="text-gray-500 dark:text-gray-400">{{ t('payment.planAmount') }}</span>
                        <span class="font-medium tabular-nums text-gray-900 dark:text-white">¥{{ selectedPlan.price.toFixed(2) }}</span>
                      </div>
                      <div v-if="feeRate > 0 && selectedPlan.price > 0" class="flex justify-between gap-4">
                        <span class="text-gray-500 dark:text-gray-400">{{ t('payment.fee') }} ({{ feeRate }}%)</span>
                        <span class="font-medium tabular-nums text-gray-900 dark:text-white">¥{{ subFeeAmount.toFixed(2) }}</span>
                      </div>
                      <div class="flex justify-between gap-4 border-t border-gray-100 pt-3 dark:border-dark-700/60">
                        <span class="font-semibold text-gray-700 dark:text-gray-200">{{ t('payment.actualPay') }}</span>
                        <span :class="['text-lg font-semibold tabular-nums', planTextClass]">¥{{ (feeRate > 0 ? subTotalAmount : selectedPlan.price).toFixed(2) }}</span>
                      </div>
                    </div>

                    <div v-if="enabledMethods.length >= 1" class="mt-4 rounded-2xl border border-gray-200/70 bg-white p-4 dark:border-dark-700/60 dark:bg-dark-800/40">
                      <PaymentMethodSelector
                        :methods="subMethodOptions"
                        :selected="selectedMethod"
                        layout="list"
                        @select="selectedMethod = $event"
                      />
                    </div>

                    <div class="mt-5 space-y-2">
                      <button :class="['btn w-full py-3 text-base font-semibold', paymentButtonClass]" :disabled="!canSubmitSubscription || submitting" @click="confirmSubscribe">
                        <span v-if="submitting" class="flex items-center justify-center gap-2">
                          <span class="h-4 w-4 animate-spin rounded-full border-2 border-white border-t-transparent"></span>
                          {{ t('common.processing') }}
                        </span>
                        <span v-else>{{ pendingRenewSeatId > 0 ? t('payment.confirmRenewal') : t('payment.createOrder') }} ¥{{ (feeRate > 0 ? subTotalAmount : selectedPlan.price).toFixed(2) }}</span>
                      </button>
                      <button
                        class="block w-full py-2 text-center text-sm text-gray-500 transition-colors hover:text-gray-800 dark:text-gray-400 dark:hover:text-gray-200"
                        @click="cancelSubscriptionFlow">
                        {{ t('common.cancel') }}
                      </button>
                    </div>
                  </aside>
                </div>
              </section>
            </template>
            <!-- Plan list -->
            <template v-else>
              <EmptyState
                v-if="checkout.plans.length === 0"
                variant="emerald"
                :description="t('payment.noPlans')"
              >
                <template #icon>
                  <Icon name="gift" class="empty-state-icon" />
                </template>
              </EmptyState>
              <template v-else>
                <!-- 卡类型筛选：当套餐数量 > 1 且存在多种卡类型时才显示 -->
                <div v-if="cardTypeFilters.length > 1" class="-mx-1 mb-1 flex flex-wrap gap-1 px-1">
                  <button v-for="ct in cardTypeFilters" :key="ct"
                    type="button"
                    :class="[
                      'rounded-full border px-3 py-1 text-xs font-medium transition-colors',
                      activeCardType === ct
                        ? 'border-gray-900 bg-gray-900 text-white dark:border-white dark:bg-white dark:text-gray-900'
                        : 'border-gray-200 bg-white text-gray-600 hover:border-gray-300 hover:text-gray-900 dark:border-dark-700 dark:bg-dark-800 dark:text-dark-300 dark:hover:border-dark-600 dark:hover:text-white',
                    ]"
                    @click="activeCardType = ct">
                    {{ ct === 'all' ? t('payment.admin.cardType.all') : t(`payment.admin.cardType.${ct}`) }}
                  </button>
                </div>
                <div :class="planGridClass">
                  <SubscriptionPlanCard v-for="plan in filteredPlans" :key="plan.id" :plan="plan" :active-subscriptions="activeSubscriptions" @select="selectPlan" />
                </div>
              </template>
              <!-- Active subscriptions (compact, below plan list) -->
              <div v-if="activeSubscriptions.length > 0">
                <p class="mb-2 text-[12px] font-medium text-gray-500 dark:text-dark-400">{{ t('payment.activeSubscription') }}</p>
                <div class="space-y-2">
                  <div v-for="sub in activeSubscriptions" :key="sub.id"
                    class="flex items-center gap-3 rounded-xl border border-gray-200/70 bg-white px-3 py-2 dark:border-dark-700/60 dark:bg-dark-800/40">
                    <div :class="['h-6 w-1 shrink-0 rounded-full', platformAccentBarClass(sub.group?.platform || '')]" />
                    <div class="min-w-0 flex-1">
                      <div class="flex items-center gap-1.5">
                        <span class="truncate text-xs font-semibold text-gray-900 dark:text-white">{{ sub.group?.name || t('payment.groupFallback', { id: sub.group_id }) }}</span>
                        <span :class="['shrink-0 rounded-full px-1.5 py-0.5 text-[9px] font-medium', platformBadgeLightClass(sub.group?.platform || '')]">{{ platformLabel(sub.group?.platform || '') }}</span>
                      </div>
                      <div class="flex flex-wrap gap-x-3 text-[11px] text-gray-500 dark:text-dark-400">
                        <span>{{ t('payment.planCard.rate') }}: ×{{ sub.group?.rate_multiplier ?? 1 }}</span>
                        <span v-if="sub.group?.daily_limit_usd == null && sub.group?.weekly_limit_usd == null && sub.group?.monthly_limit_usd == null">{{ t('payment.planCard.quota') }}: {{ t('payment.planCard.unlimited') }}</span>
                        <span v-if="sub.expires_at">{{ t('userSubscriptions.daysRemaining', { days: getDaysRemaining(sub.expires_at) }) }}</span>
                        <span v-else>{{ t('userSubscriptions.noExpiration') }}</span>
                      </div>
                    </div>
                    <span class="inline-flex shrink-0 items-center gap-1 rounded-full bg-emerald-50 px-2 py-0.5 text-[10px] font-medium text-emerald-700 ring-1 ring-inset ring-emerald-200/70 dark:bg-emerald-500/15 dark:text-emerald-300 dark:ring-emerald-500/30">
                      <span class="h-1.5 w-1.5 rounded-full bg-emerald-500 animate-pulse"></span>
                      {{ t('userSubscriptions.status.active') }}
                    </span>
                  </div>
                </div>
              </div>
            </template>
          </template>
        </template>
        <div v-if="(checkout.help_text || checkout.help_image_url) && paymentPhase === 'select' && !selectedPlan" class="rounded-2xl border border-gray-200/70 bg-white p-5 dark:border-dark-700/60 dark:bg-dark-800/40">
          <div class="flex flex-col items-center gap-3">
            <img v-if="checkout.help_image_url" :src="checkout.help_image_url" alt=""
              class="h-40 max-w-full cursor-pointer rounded-xl object-contain transition-opacity hover:opacity-80"
              @click="previewImage = checkout.help_image_url" />
            <p v-if="checkout.help_text" class="text-center text-sm text-gray-500 dark:text-gray-400">{{ checkout.help_text }}</p>
          </div>
        </div>
      </template>
    </div>
    <!-- Renewal Plan Selection Modal -->
    <Teleport to="body">
      <Transition name="modal">
        <div v-if="showRenewalModal" class="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm p-4" @click.self="closeRenewalModal">
          <div class="relative w-full max-w-lg rounded-2xl border border-gray-200/70 bg-white p-6 shadow-2xl dark:border-dark-700/60 dark:bg-dark-800">
            <!-- Close button -->
            <button class="absolute right-4 top-4 rounded-lg p-1 text-gray-400 transition-colors hover:bg-gray-100 hover:text-gray-600 dark:hover:bg-dark-700 dark:hover:text-gray-200" @click="closeRenewalModal">
              <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" /></svg>
            </button>
            <h3 class="mb-4 text-base font-semibold tracking-tight text-gray-900 dark:text-white">{{ t('payment.selectPlan') }}</h3>
            <div class="space-y-3">
              <SubscriptionPlanCard v-for="plan in renewalPlans" :key="plan.id" :plan="plan" :active-subscriptions="activeSubscriptions" @select="selectPlanFromModal" />
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>
    <!-- Image Preview Overlay -->
    <Teleport to="body">
      <Transition name="modal">
        <div v-if="previewImage" class="fixed inset-0 z-[60] flex items-center justify-center bg-black/70 backdrop-blur-sm" @click="previewImage = ''">
          <img :src="previewImage" alt="" class="max-h-[85vh] max-w-[90vw] rounded-xl object-contain shadow-2xl" />
        </div>
      </Transition>
    </Teleport>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { usePaymentStore } from '@/stores/payment'
import { useSubscriptionStore } from '@/stores/subscriptions'
import { useAppStore } from '@/stores'
import { paymentAPI } from '@/api/payment'
import { extractApiErrorMessage, extractI18nErrorMessage } from '@/utils/apiError'
import { isMobileDevice } from '@/utils/device'
import type { SubscriptionPlan, CheckoutInfoResponse, CreateOrderResult, OrderType } from '@/types/payment'
import AppLayout from '@/components/layout/AppLayout.vue'
import AmountInput from '@/components/payment/AmountInput.vue'
import PaymentMethodSelector from '@/components/payment/PaymentMethodSelector.vue'
import { METHOD_ORDER, getPaymentPopupFeatures } from '@/components/payment/providerConfig'
import {
  PAYMENT_RECOVERY_STORAGE_KEY,
  buildCreateOrderPayload,
  clearPaymentRecoverySnapshot,
  decidePaymentLaunch,
  getVisibleMethods,
  normalizeVisibleMethod,
  readPaymentRecoverySnapshot,
  type PaymentRecoverySnapshot,
  writePaymentRecoverySnapshot,
} from '@/components/payment/paymentFlow'
import { collectCardTypes, derivePlanCardType, type PlanCardType } from '@/utils/planCardType'
import { getEffectiveLimitVisibility } from '@/utils/planLimits'
import {
  platformAccentBarClass,
  platformBadgeLightClass,
  platformBorderClass,
  platformDiscountClass,
  platformIconClass,
  platformTextClass,
  platformLabel,
} from '@/utils/platformColors'
import SubscriptionPlanCard from '@/components/payment/SubscriptionPlanCard.vue'
import PaymentStatusPanel from '@/components/payment/PaymentStatusPanel.vue'
import Icon from '@/components/icons/Icon.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import BrandIcon from '@/components/common/BrandIcon.vue'
import ModelIcon from '@/components/common/ModelIcon.vue'
import type { PaymentMethodOption } from '@/components/payment/PaymentMethodSelector.vue'
import { buildPaymentErrorToastMessage, describePaymentScenarioError } from './paymentUx'
import { hasWechatResumeQuery, parseWechatResumeRoute, stripWechatResumeQuery } from './paymentWechatResume'

const { t } = useI18n()
const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()
const paymentStore = usePaymentStore()
const subscriptionStore = useSubscriptionStore()
const appStore = useAppStore()

const user = computed(() => authStore.user)
const activeSubscriptions = computed(() => subscriptionStore.activeSubscriptions)

const MODEL_SCOPE_META: Record<string, { label: string; iconModel: string }> = {
  claude: { label: 'Claude', iconModel: 'claude-3-5-sonnet' },
  gemini_text: { label: 'Gemini', iconModel: 'gemini-2.5-pro' },
  gemini_image: { label: 'Imagen', iconModel: 'imagen-3' },
}

function getDaysRemaining(expiresAt: string): number {
  const diff = new Date(expiresAt).getTime() - Date.now()
  return Math.max(0, Math.ceil(diff / (1000 * 60 * 60 * 24)))
}

const loading = ref(true)
const submitting = ref(false)
const errorMessage = ref('')
const errorHintMessage = ref('')
const activeTab = ref<'recharge' | 'subscription'>('recharge')
const amount = ref<number | null>(null)
const selectedMethod = ref('')
const selectedPlan = ref<SubscriptionPlan | null>(null)
const previewImage = ref('')
// 续费独享名额时携带的 seat ID（来自 ?renew_seat=<id>），confirmSubscribe 透传给 CreateOrder
const pendingRenewSeatId = ref<number>(0)
// 续费场景下用户上次实际支付的金额（来自 PreviewRenewal API），用于「上次 → 本次」对比展示
const renewLastPaidPrice = ref<number>(0)
// 卡类型筛选：'all' 显示全部，否则按 derivePlanCardType 推导后过滤
const activeCardType = ref<'all' | PlanCardType>('all')

const paymentPhase = ref<'select' | 'paying'>('select')

interface CreateOrderOptions {
  openid?: string
  wechatResumeToken?: string
  paymentType?: string
  isResume?: boolean
  mobileQrFallbackAttempted?: boolean
  /** 续费目标 seat ID（>0 时本订单走 RenewSeat 路径而不是 AssignSeat） */
  renewSeatId?: number
}

interface WeixinJSBridgeLike {
  invoke(
    action: string,
    payload: Record<string, unknown>,
    callback: (result: Record<string, unknown>) => void,
  ): void
}

function emptyPaymentState(): PaymentRecoverySnapshot {
  return {
    orderId: 0,
    amount: 0,
    qrCode: '',
    qrCodeImage: '',
    expiresAt: '',
    paymentType: '',
    payUrl: '',
    outTradeNo: '',
    clientSecret: '',
    payAmount: 0,
    orderType: '',
    paymentMode: '',
    resumeToken: '',
    createdAt: 0,
  }
}

function getWeixinJSBridge(): WeixinJSBridgeLike | undefined {
  return (window as Window & { WeixinJSBridge?: WeixinJSBridgeLike }).WeixinJSBridge
}

function waitForWeixinJSBridge(timeoutMs = 4000): Promise<WeixinJSBridgeLike | null> {
  const existing = getWeixinJSBridge()
  if (existing) return Promise.resolve(existing)

  return new Promise((resolve) => {
    let settled = false
    const finish = (bridge: WeixinJSBridgeLike | null) => {
      if (settled) return
      settled = true
      document.removeEventListener('WeixinJSBridgeReady', handleReady)
      document.removeEventListener('onWeixinJSBridgeReady', handleReady)
      window.clearTimeout(timer)
      resolve(bridge)
    }
    const handleReady = () => finish(getWeixinJSBridge() ?? null)
    const timer = window.setTimeout(() => finish(getWeixinJSBridge() ?? null), timeoutMs)
    document.addEventListener('WeixinJSBridgeReady', handleReady, false)
    document.addEventListener('onWeixinJSBridgeReady', handleReady, false)
  })
}

async function invokeWechatJsapiPayment(payload: Record<string, unknown>): Promise<Record<string, unknown>> {
  const bridge = await waitForWeixinJSBridge()
  if (!bridge) {
    throw new Error('WECHAT_JSAPI_UNAVAILABLE')
  }
  return new Promise((resolve) => {
    bridge.invoke('getBrandWCPayRequest', payload, (result) => resolve(result || {}))
  })
}

const paymentState = ref<PaymentRecoverySnapshot>(emptyPaymentState())

function persistRecoverySnapshot(snapshot: PaymentRecoverySnapshot) {
  if (typeof window === 'undefined' || !snapshot.orderId) return
  writePaymentRecoverySnapshot(window.localStorage, snapshot, PAYMENT_RECOVERY_STORAGE_KEY)
}

function removeRecoverySnapshot() {
  if (typeof window === 'undefined') return
  clearPaymentRecoverySnapshot(window.localStorage, PAYMENT_RECOVERY_STORAGE_KEY)
}

function resetPayment() {
  paymentPhase.value = 'select'
  paymentState.value = emptyPaymentState()
  removeRecoverySnapshot()
}

async function redirectToPaymentResult(state: PaymentRecoverySnapshot): Promise<void> {
  const query: Record<string, string | undefined> = {}
  if (state.orderId > 0) {
    query.order_id = String(state.orderId)
  }
  if (state.outTradeNo) {
    query.out_trade_no = state.outTradeNo
  }
  if (state.resumeToken) {
    query.resume_token = state.resumeToken
  }
  await router.push({
    path: '/payment/result',
    query,
  })
}

function buildWechatOAuthAuthorizeUrl(
  authorizeUrl: string,
  context: { paymentType: string; orderType: OrderType; planId?: number; orderAmount: number; renewSeatId?: number },
): string {
  const normalizedUrl = authorizeUrl.trim()
  if (!normalizedUrl || typeof window === 'undefined') {
    return normalizedUrl
  }

  try {
    const targetUrl = new URL(normalizedUrl, window.location.origin)
    const redirectPath = targetUrl.searchParams.get('redirect') || '/purchase'
    const redirectUrl = new URL(redirectPath, window.location.origin)
    const paymentType = normalizeVisibleMethod(context.paymentType) || context.paymentType.trim() || 'wxpay'

    redirectUrl.searchParams.set('payment_type', paymentType)
    redirectUrl.searchParams.set('order_type', context.orderType)

    if (context.planId) {
      redirectUrl.searchParams.set('plan_id', String(context.planId))
    } else {
      redirectUrl.searchParams.delete('plan_id')
    }

    if (context.orderAmount > 0) {
      redirectUrl.searchParams.set('amount', String(context.orderAmount))
    } else {
      redirectUrl.searchParams.delete('amount')
    }

    // 续费独享名额时把 seat_id 也带进 redirect，OAuth 回跳后前端会一并恢复并发给后端，
    // 避免续费上下文丢失导致重新分配账号 / 库存不足触发自动退款
    if (context.renewSeatId && context.renewSeatId > 0) {
      redirectUrl.searchParams.set('renew_seat', String(context.renewSeatId))
    } else {
      redirectUrl.searchParams.delete('renew_seat')
    }

    targetUrl.searchParams.set('redirect', `${redirectUrl.pathname}${redirectUrl.search}`)
    return targetUrl.toString()
  } catch {
    return normalizedUrl
  }
}

function onPaymentDone() {
  const wasSubscription = paymentState.value.orderType === 'subscription'
  resetPayment()
  selectedPlan.value = null
  if (wasSubscription) {
    subscriptionStore.fetchActiveSubscriptions(true).catch(() => {})
  }
}

function onPaymentSuccess() {
  removeRecoverySnapshot()
  authStore.refreshUser()
  if (paymentState.value.orderType === 'subscription') {
    subscriptionStore.fetchActiveSubscriptions(true).catch(() => {})
  }
}

function onPaymentSettled() {
  removeRecoverySnapshot()
}

// All checkout data from single API call
const defaultQuickAmounts = [10, 20, 50, 100, 200, 500, 1000, 2000, 5000]
const checkout = ref<CheckoutInfoResponse>({
  methods: {}, global_min: 0, global_max: 0,
  plans: [], balance_disabled: false, balance_recharge_multiplier: 1, recharge_fee_rate: 0, quick_amounts: [], help_text: '', help_image_url: '', stripe_publishable_key: '',
})

const tabs = computed(() => {
  const result: { key: 'recharge' | 'subscription'; label: string }[] = []
  if (!checkout.value.balance_disabled) result.push({ key: 'recharge', label: t('payment.tabTopUp') })
  result.push({ key: 'subscription', label: t('payment.tabSubscribe') })
  return result
})

const visibleMethods = computed(() => getVisibleMethods(checkout.value.methods))
const enabledMethods = computed(() => Object.keys(visibleMethods.value))
const validAmount = computed(() => amount.value ?? 0)
const quickAmounts = computed(() => {
  const values = Array.isArray(checkout.value.quick_amounts)
    ? checkout.value.quick_amounts
        .map((item) => Number(item))
        .filter((item) => Number.isFinite(item) && item > 0)
    : []
  return values.length > 0 ? values : defaultQuickAmounts
})
const balanceRechargeMultiplier = computed(() => {
  const multiplier = checkout.value.balance_recharge_multiplier
  return multiplier > 0 ? multiplier : 1
})
const creditedAmount = computed(() => Math.round((validAmount.value * balanceRechargeMultiplier.value) * 100) / 100)

// Adaptive grid: center single card, 2-col for 2 plans, 3-col for 3+
const planGridClass = computed(() => {
  const n = filteredPlans.value.length
  if (n === 1) return 'grid grid-cols-1 gap-5'
  if (n <= 2) return 'grid grid-cols-1 gap-5 sm:grid-cols-2'
  return 'grid grid-cols-1 gap-5 sm:grid-cols-2 lg:grid-cols-3'
})

// 卡类型筛选 tab：实际存在的类型 + 「全部」
const cardTypeFilters = computed<('all' | PlanCardType)[]>(() => {
  const present = collectCardTypes(checkout.value.plans)
  if (present.length <= 1) return []
  return ['all', ...present]
})

// 按卡类型过滤后的套餐列表
const filteredPlans = computed(() => {
  if (activeCardType.value === 'all') return checkout.value.plans
  return checkout.value.plans.filter(
    (p) => derivePlanCardType(p.validity_days, p.validity_unit) === activeCardType.value,
  )
})

// Check if an amount fits a method's [min, max]. 0 = no limit.
function amountFitsMethod(amt: number, methodType: string): boolean {
  if (amt <= 0) return true
  const ml = visibleMethods.value[methodType]
  if (!ml) return false
  if (ml.single_min > 0 && amt < ml.single_min) return false
  if (ml.single_max > 0 && amt > ml.single_max) return false
  return true
}

// Visible methods decide the amount range shown to users.
const globalMinAmount = computed(() => {
  const limits = Object.values(visibleMethods.value)
  if (limits.length === 0) return 0
  if (limits.some(limit => limit.single_min <= 0)) return 0
  return Math.min(...limits.map(limit => limit.single_min))
})
const globalMaxAmount = computed(() => {
  const limits = Object.values(visibleMethods.value)
  if (limits.length === 0) return 0
  if (limits.some(limit => limit.single_max <= 0)) return 0
  return Math.max(...limits.map(limit => limit.single_max))
})

// Selected method's limits (for validation and error messages)
const selectedLimit = computed(() => visibleMethods.value[selectedMethod.value])

const methodOptions = computed<PaymentMethodOption[]>(() =>
  enabledMethods.value.map((type) => {
    const ml = visibleMethods.value[type]
    return {
      type,
      fee_rate: ml?.fee_rate ?? 0,
      available: ml?.available !== false && amountFitsMethod(validAmount.value, type),
    }
  })
)

const feeRate = computed(() => checkout.value?.recharge_fee_rate ?? 0)
const feeAmount = computed(() =>
  feeRate.value > 0 && validAmount.value > 0
    ? Math.ceil(((validAmount.value * feeRate.value) / 100) * 100) / 100
    : 0
)
const totalAmount = computed(() =>
  feeRate.value > 0 && validAmount.value > 0
    ? Math.round((validAmount.value + feeAmount.value) * 100) / 100
    : validAmount.value
)

const amountError = computed(() => {
  if (validAmount.value <= 0) return ''
  // No method can handle this amount
  if (!enabledMethods.value.some((m) => amountFitsMethod(validAmount.value, m))) {
    return t('payment.amountNoMethod')
  }
  // Selected method can't handle this amount (but others can)
  const ml = selectedLimit.value
  if (ml) {
    if (ml.single_min > 0 && validAmount.value < ml.single_min) return t('payment.amountTooLow', { min: ml.single_min })
    if (ml.single_max > 0 && validAmount.value > ml.single_max) return t('payment.amountTooHigh', { max: ml.single_max })
  }
  return ''
})

const canSubmit = computed(() =>
  validAmount.value > 0
    && amountFitsMethod(validAmount.value, selectedMethod.value)
    && selectedLimit.value?.available !== false
)

// Subscription-specific: method options based on plan price
const subMethodOptions = computed<PaymentMethodOption[]>(() => {
  const planPrice = selectedPlan.value?.price ?? 0
  return enabledMethods.value.map((type) => {
    const ml = visibleMethods.value[type]
    return {
      type,
      fee_rate: ml?.fee_rate ?? 0,
      available: ml?.available !== false && amountFitsMethod(planPrice, type),
    }
  })
})

const subFeeAmount = computed(() => {
  const price = selectedPlan.value?.price ?? 0
  if (feeRate.value <= 0 || price <= 0) return 0
  return Math.ceil(((price * feeRate.value) / 100) * 100) / 100
})

const subTotalAmount = computed(() => {
  const price = selectedPlan.value?.price ?? 0
  if (feeRate.value <= 0 || price <= 0) return price
  return Math.round((price + subFeeAmount.value) * 100) / 100
})

// 续费场景的价格变动方向（与上次实付价对比，差额 < 0.005 视为不变）
const renewPriceTrend = computed<'up' | 'down' | 'same'>(() => {
  if (!selectedPlan.value || renewLastPaidPrice.value <= 0) return 'same'
  const delta = selectedPlan.value.price - renewLastPaidPrice.value
  if (delta > 0.005) return 'up'
  if (delta < -0.005) return 'down'
  return 'same'
})

// 本次价金额的文本颜色：涨价 amber、降价 emerald、不变保持深灰
const renewPriceTextClass = computed(() => {
  switch (renewPriceTrend.value) {
    case 'up':
      return 'text-amber-600 dark:text-amber-300'
    case 'down':
      return 'text-emerald-600 dark:text-emerald-300'
    default:
      return 'text-gray-900 dark:text-gray-50'
  }
})

// banner 外框：涨价时左侧 4px amber 强调条 + amber 边框（仍保持白底，不全染）
const renewBannerOuterClass = computed(() => {
  switch (renewPriceTrend.value) {
    case 'up':
      return 'border-amber-200 border-l-4 border-l-amber-500 dark:border-amber-900/50 dark:border-l-amber-500'
    case 'down':
      return 'border-emerald-200 border-l-4 border-l-emerald-500 dark:border-emerald-900/50 dark:border-l-emerald-500'
    default:
      return 'border-gray-200 dark:border-dark-700'
  }
})

// 标题旁圆角图标盒
const renewBannerIconBoxClass = computed(() => {
  switch (renewPriceTrend.value) {
    case 'up':
      return 'bg-amber-50 text-amber-600 dark:bg-amber-900/30 dark:text-amber-300'
    case 'down':
      return 'bg-emerald-50 text-emerald-600 dark:bg-emerald-900/30 dark:text-emerald-300'
    default:
      return 'bg-emerald-50 text-emerald-600 dark:bg-emerald-900/30 dark:text-emerald-300'
  }
})

// 中央趋势 dot 容器：双层 ring（外层淡色光晕 + 内层 ring offset）做地标感
const renewTrendDotClass = computed(() => {
  switch (renewPriceTrend.value) {
    case 'up':
      return 'bg-amber-500 text-white ring-4 ring-amber-100 dark:bg-amber-500 dark:text-white dark:ring-amber-900/40'
    case 'down':
      return 'bg-emerald-500 text-white ring-4 ring-emerald-100 dark:bg-emerald-500 dark:text-white dark:ring-emerald-900/40'
    default:
      return 'bg-gray-300 text-white ring-4 ring-gray-100 dark:bg-dark-500 dark:text-dark-200 dark:ring-dark-700'
  }
})

// 顶部 hint 文字色：跟趋势同色（涨 amber-700、降 emerald-700、不变 gray），强化趋势感知
const renewBannerHintClass = computed(() => {
  switch (renewPriceTrend.value) {
    case 'up':
      return 'text-amber-700 dark:text-amber-300/90'
    case 'down':
      return 'text-emerald-700 dark:text-emerald-300/90'
    default:
      return 'text-gray-500 dark:text-dark-300'
  }
})

const canSubmitSubscription = computed(() =>
  selectedPlan.value !== null
    && amountFitsMethod(selectedPlan.value.price, selectedMethod.value)
    && selectedLimit.value?.available !== false
)

// Auto-switch to first available method when current selection can't handle the amount
watch(() => [validAmount.value, selectedMethod.value] as const, ([amt, method]) => {
  if (amt <= 0 || amountFitsMethod(amt, method)) return
  const available = enabledMethods.value.find((m) => amountFitsMethod(amt, m))
  if (available) selectedMethod.value = available
})

// 卡类型筛选自愈：当前 activeCardType 在可用 tab 里消失（管理员下架所有该类套餐）→ 重置到 all
watch(cardTypeFilters, (filters) => {
  if (filters.length > 0 && !filters.includes(activeCardType.value)) {
    activeCardType.value = 'all'
  }
})

// Payment button class: follows selected payment method color
const paymentButtonClass = computed(() => {
  const m = selectedMethod.value
  if (!m) return 'btn-primary'
  if (m.includes('alipay')) return 'btn-alipay'
  if (m.includes('wxpay')) return 'btn-wxpay'
  if (m === 'stripe') return 'btn-stripe'
  return 'btn-primary'
})

// Subscription confirm: platform accent colors (clean card, no gradient)
const planAccentClass = computed(() => platformAccentBarClass(selectedPlan.value?.group_platform || ''))
const planBadgeLightClass = computed(() => platformBadgeLightClass(selectedPlan.value?.group_platform || ''))
const planBorderClass = computed(() => platformBorderClass(selectedPlan.value?.group_platform || ''))
const planIconClass = computed(() => platformIconClass(selectedPlan.value?.group_platform || ''))
const planTextClass = computed(() => platformTextClass(selectedPlan.value?.group_platform || ''))
const selectedPlanDiscountClass = computed(() => platformDiscountClass(selectedPlan.value?.group_platform || ''))
const selectedPlanPlatformBrand = computed<'claude' | 'openai' | 'gemini' | null>(() => {
  const platform = selectedPlan.value?.group_platform || ''
  if (platform === 'anthropic') return 'claude'
  if (platform === 'openai') return 'openai'
  if (platform === 'gemini') return 'gemini'
  return null
})
const selectedPlanPlatformInitial = computed(() => platformLabel(selectedPlan.value?.group_platform || '').charAt(0).toUpperCase())
const selectedPlanRateDisplay = computed(() => {
  const rate = selectedPlan.value?.rate_multiplier ?? 1
  const base = `×${Number(rate.toPrecision(10))}`
  // 倍率 < 1 = 折扣，把"几折"显式标出来；中国用户更熟悉"6 折"这种说法
  if (rate > 0 && rate < 1) {
    // 0.6 → 6折，0.85 → 8.5折，0.5 → 5折
    const discount = Number((rate * 10).toFixed(1))
    return `${base} · ${discount}折`
  }
  return base
})
const selectedPlanDiscountText = computed(() => {
  const plan = selectedPlan.value
  if (!plan?.original_price || plan.original_price <= 0) return ''
  const pct = Math.round((1 - plan.price / plan.original_price) * 100)
  return pct > 0 ? `-${pct}%` : ''
})
// 限额可视性：废限额自动隐藏。规则见 utils/planLimits.ts
const selectedPlanLimitVisibility = computed(() => {
  if (!selectedPlan.value) return { showDaily: false, showWeekly: false, showMonthly: false }
  return getEffectiveLimitVisibility({
    daily_limit_usd: selectedPlan.value.daily_limit_usd,
    weekly_limit_usd: selectedPlan.value.weekly_limit_usd,
    monthly_limit_usd: selectedPlan.value.monthly_limit_usd,
  })
})

const selectedPlanModelScopeItems = computed(() => {
  // supported_model_scopes 仅对 antigravity 平台有意义（选 claude / gemini_text / gemini_image 子能力）。
  // 其它单平台套餐（openai / anthropic / gemini）的 scope 字段是 createForm 默认值的残留，展示出来会误导用户
  // （例如 GPT 周卡显示 "Claude · Gemini · Imagen"）。这里加平台守门。
  if (selectedPlan.value?.group_platform !== 'antigravity') return []
  const scopes = selectedPlan.value?.supported_model_scopes
  if (!scopes || scopes.length === 0) return []
  return scopes.map(s => ({
    key: s,
    label: MODEL_SCOPE_META[s]?.label || s,
    iconModel: MODEL_SCOPE_META[s]?.iconModel || s,
  }))
})

// Renewal modal state
const showRenewalModal = ref(false)
const renewGroupId = ref<number | null>(null)
const renewalPlans = computed(() => {
  if (renewGroupId.value == null) return []
  return checkout.value.plans.filter(p => p.group_id === renewGroupId.value)
})

const planValiditySuffix = computed(() => {
  if (!selectedPlan.value) return ''
  const u = selectedPlan.value.validity_unit || 'day'
  if (u === 'month') return t('payment.perMonth')
  if (u === 'year') return t('payment.perYear')
  return `${selectedPlan.value.validity_days}${t('payment.days')}`
})

function selectPlan(plan: SubscriptionPlan) {
  // 切换到不同套餐时清掉续费上下文，避免后端 RENEWAL_PLAN_MISMATCH
  if (pendingRenewSeatId.value > 0 && selectedPlan.value?.id !== plan.id) {
    pendingRenewSeatId.value = 0
    renewLastPaidPrice.value = 0
  }
  selectedPlan.value = plan
  errorMessage.value = ''
}

function selectPlanFromModal(plan: SubscriptionPlan) {
  showRenewalModal.value = false
  renewGroupId.value = null
  // 续费弹窗里选套餐属于全新购买流程，清掉残留的续费上下文
  pendingRenewSeatId.value = 0
  renewLastPaidPrice.value = 0
  selectedPlan.value = plan
  errorMessage.value = ''
}

function closeRenewalModal() {
  showRenewalModal.value = false
  renewGroupId.value = null
}

async function handleSubmitRecharge() {
  if (!canSubmit.value || submitting.value) return
  await createOrder(validAmount.value, 'balance')
}

async function confirmSubscribe() {
  if (!selectedPlan.value || submitting.value) return
  await createOrder(selectedPlan.value.price, 'subscription', selectedPlan.value.id, { renewSeatId: pendingRenewSeatId.value })
}

// 取消选中套餐：同时清掉续费上下文，避免再点其他 plan 仍带着 renewal 标记
function cancelSubscriptionFlow() {
  selectedPlan.value = null
  pendingRenewSeatId.value = 0
  renewLastPaidPrice.value = 0
}

async function createOrder(orderAmount: number, orderType: OrderType, planId?: number, options: CreateOrderOptions = {}) {
  submitting.value = true
  errorMessage.value = ''
  errorHintMessage.value = ''
  const requestType = normalizeVisibleMethod(options.paymentType || selectedMethod.value) || options.paymentType || selectedMethod.value
  try {
    const payload = buildCreateOrderPayload({
      amount: orderAmount,
      paymentType: requestType,
      orderType,
      planId,
      origin: typeof window !== 'undefined' ? window.location.origin : '',
      isMobile: isMobileDevice(),
      isWechatBrowser: typeof window !== 'undefined' && /MicroMessenger/i.test(window.navigator.userAgent),
      renewSeatId: options.renewSeatId,
    })
    if (options.openid) {
      payload.openid = options.openid
    }
    if (options.wechatResumeToken) {
      payload.wechat_resume_token = options.wechatResumeToken
    }

    const result = await paymentStore.createOrder(payload) as CreateOrderResult & { resume_token?: string }
    const openWindow = (url: string) => {
      const win = window.open(url, 'paymentPopup', getPaymentPopupFeatures())
      if (!win || win.closed) {
        window.location.href = url
      }
    }
    const visibleMethod = normalizeVisibleMethod(requestType) || requestType
    // When user clicks the dedicated Stripe button, leave method blank so the
    // landing page renders Stripe's full Payment Element (card/link/alipay/wxpay).
    const stripeMethod = visibleMethod === 'stripe'
      ? ''
      : visibleMethod === 'wxpay' ? 'wechat_pay' : 'alipay'
    const stripeRouteUrl = result.client_secret
      ? router.resolve({
        path: '/payment/stripe',
        query: {
          order_id: String(result.order_id),
          client_secret: result.client_secret,
          method: stripeMethod || undefined,
          resume_token: result.resume_token || undefined,
        },
      }).href
      : ''
    const decision = decidePaymentLaunch(result, {
      visibleMethod,
      orderType,
      isMobile: isMobileDevice(),
      isWechatBrowser: typeof window !== 'undefined' && /MicroMessenger/i.test(window.navigator.userAgent),
      stripePopupUrl: stripeRouteUrl,
      stripeRouteUrl,
    })

    if (decision.kind === 'wechat_oauth' && decision.oauth?.authorize_url) {
      window.location.href = buildWechatOAuthAuthorizeUrl(decision.oauth.authorize_url, {
        paymentType: visibleMethod,
        orderType,
        planId,
        orderAmount,
        renewSeatId: options.renewSeatId,
      })
      return
    }

    if (decision.kind === 'unhandled') {
      applyScenarioError({ reason: 'UNHANDLED_PAYMENT_SCENARIO' }, visibleMethod)
      return
    }

    paymentState.value = decision.paymentState
    paymentPhase.value = 'paying'
    persistRecoverySnapshot(decision.recovery)

    if (decision.kind === 'stripe_popup') {
      openWindow(decision.paymentState.payUrl)
      return
    }
    if (decision.kind === 'stripe_route') {
      window.location.href = decision.paymentState.payUrl
      return
    }
    if (decision.kind === 'wechat_jsapi' && decision.jsapi) {
      try {
        const jsapiResult = await invokeWechatJsapiPayment(decision.jsapi as Record<string, unknown>)
        const errMsg = String(jsapiResult.err_msg || '').toLowerCase()
        if (errMsg.includes('cancel')) {
          appStore.showInfo(t('payment.qr.cancelled'))
          resetPayment()
        } else if (errMsg && !errMsg.includes('ok')) {
          resetPayment()
          const fallbackApplied = await attemptMobileQrFallback(
            { reason: 'WECHAT_JSAPI_FAILED', message: errMsg },
            {
              orderAmount,
              orderType,
              planId,
              paymentType: visibleMethod,
              attempted: options.mobileQrFallbackAttempted === true,
              renewSeatId: options.renewSeatId,
            },
          )
          if (!fallbackApplied) {
            applyScenarioError({ reason: 'WECHAT_JSAPI_FAILED', message: errMsg }, visibleMethod)
          }
        } else {
          const resultState = { ...decision.paymentState }
          resetPayment()
          await redirectToPaymentResult(resultState)
        }
      } catch (err: unknown) {
        resetPayment()
        const fallbackApplied = await attemptMobileQrFallback(err, {
          orderAmount,
          orderType,
          planId,
          paymentType: visibleMethod,
          attempted: options.mobileQrFallbackAttempted === true,
          renewSeatId: options.renewSeatId,
        })
        if (!fallbackApplied) {
          throw err
        }
      }
      return
    }
    if (decision.kind === 'redirect_waiting' && decision.paymentState.payUrl) {
      if (isMobileDevice()) {
        window.location.href = decision.paymentState.payUrl
        return
      }
      openWindow(decision.paymentState.payUrl)
    }
  } catch (err: unknown) {
    const apiErr = err as Record<string, unknown>
    if (apiErr.reason === 'TOO_MANY_PENDING') {
      const metadata = apiErr.metadata as Record<string, unknown> | undefined
      errorMessage.value = t('payment.errors.tooManyPending', { max: metadata?.max || '' })
      errorHintMessage.value = ''
    } else if (apiErr.reason === 'CANCEL_RATE_LIMITED') {
      errorMessage.value = t('payment.errors.cancelRateLimited')
      errorHintMessage.value = ''
    } else if (await attemptMobileQrFallback(err, {
      orderAmount,
      orderType,
      planId,
      paymentType: requestType,
      attempted: options.mobileQrFallbackAttempted === true,
      renewSeatId: options.renewSeatId,
    })) {
      return
    } else {
      const handled = applyScenarioError(
        err,
        normalizeVisibleMethod(options.paymentType || selectedMethod.value) || selectedMethod.value,
      )
      if (!handled) {
        errorMessage.value = extractI18nErrorMessage(err, t, 'payment.errors', extractApiErrorMessage(err, t('payment.result.failed')))
        errorHintMessage.value = ''
      }
      if (handled) {
        return
      }
    }
    appStore.showError(buildPaymentErrorToastMessage(errorMessage.value, errorHintMessage.value))
  } finally {
    submitting.value = false
  }
}

interface MobileQrFallbackContext {
  orderAmount: number
  orderType: OrderType
  planId?: number
  paymentType: string
  attempted: boolean
  // 续费场景必须把 seat_id 透传到回退建单，否则后端拿不到 seat 上下文会按"新购"处理，
  // 消耗新库存或 fulfillment 时因独享池无可用账号触发自动退款（GPT round 18 #2）
  renewSeatId?: number
}

function shouldFallbackToDesktopQr(err: unknown, paymentMethod: string, attempted: boolean): boolean {
  if (attempted || !isMobileDevice()) {
    return false
  }

  const normalizedMethod = normalizeVisibleMethod(paymentMethod) || paymentMethod
  const reason = typeof err === 'object' && err && 'reason' in err && typeof err.reason === 'string'
    ? err.reason
    : ''
  const message = err instanceof Error
    ? err.message
    : (typeof err === 'object' && err && 'message' in err && typeof err.message === 'string'
      ? err.message
      : '')
  const normalizedMessage = message.toLowerCase()

  if (normalizedMethod === 'wxpay') {
    return reason === 'WECHAT_H5_NOT_AUTHORIZED'
      || reason === 'WECHAT_PAYMENT_MP_NOT_CONFIGURED'
      || reason === 'WECHAT_JSAPI_FAILED'
      || reason === 'PAYMENT_GATEWAY_ERROR'
      || reason === 'UNHANDLED_PAYMENT_SCENARIO'
      || normalizedMessage.includes('weixinjsbridge is unavailable')
      || normalizedMessage.includes('wechat_jsapi_unavailable')
  }

  if (normalizedMethod === 'alipay') {
    return reason === 'PAYMENT_GATEWAY_ERROR' || reason === 'UNHANDLED_PAYMENT_SCENARIO'
  }

  return false
}

async function attemptMobileQrFallback(err: unknown, context: MobileQrFallbackContext): Promise<boolean> {
  if (!shouldFallbackToDesktopQr(err, context.paymentType, context.attempted)) {
    return false
  }

  try {
    const visibleMethod = normalizeVisibleMethod(context.paymentType) || context.paymentType
    const payload = buildCreateOrderPayload({
      amount: context.orderAmount,
      paymentType: visibleMethod,
      orderType: context.orderType,
      planId: context.planId,
      origin: typeof window !== 'undefined' ? window.location.origin : '',
      isMobile: false,
      isWechatBrowser: false,
      renewSeatId: context.renewSeatId,
    })
    const result = await paymentStore.createOrder(payload) as CreateOrderResult & { resume_token?: string }
    const stripeMethod = visibleMethod === 'wxpay' ? 'wechat_pay' : 'alipay'
    const stripeRouteUrl = result.client_secret
      ? router.resolve({
        path: '/payment/stripe',
        query: {
          order_id: String(result.order_id),
          client_secret: result.client_secret,
          method: stripeMethod,
          resume_token: result.resume_token || undefined,
        },
      }).href
      : ''
    const decision = decidePaymentLaunch(result, {
      visibleMethod,
      orderType: context.orderType,
      isMobile: false,
      isWechatBrowser: false,
      stripePopupUrl: stripeRouteUrl,
      stripeRouteUrl,
    })

    if (decision.kind !== 'qr_waiting' || (!decision.paymentState.qrCode && !decision.paymentState.qrCodeImage)) {
      return false
    }

    errorMessage.value = ''
    errorHintMessage.value = ''
    paymentState.value = decision.paymentState
    paymentPhase.value = 'paying'
    persistRecoverySnapshot(decision.recovery)
    appStore.showWarning(t('payment.errors.mobilePaymentFallbackToQr'))
    return true
  } catch {
    return false
  }
}

function applyScenarioError(err: unknown, paymentMethod: string): boolean {
  const descriptor = describePaymentScenarioError(err, {
    paymentMethod,
    isMobile: isMobileDevice(),
    isWechatBrowser: typeof window !== 'undefined' && /MicroMessenger/i.test(window.navigator.userAgent),
  })
  if (!descriptor) {
    errorMessage.value = ''
    errorHintMessage.value = ''
    return false
  }
  errorMessage.value = t(descriptor.messageKey)
  errorHintMessage.value = descriptor.hintKey ? t(descriptor.hintKey) : ''
  appStore.showError(buildPaymentErrorToastMessage(errorMessage.value, errorHintMessage.value))
  return true
}

async function resumeWechatPaymentFromQuery() {
  const resume = parseWechatResumeRoute(route.query, checkout.value.plans, validAmount.value)
  if (!resume) {
    return
  }

  selectedMethod.value = resume.paymentType
  if (resume.orderType === 'balance' && resume.orderAmount > 0) {
    amount.value = resume.orderAmount
  }
  if (resume.orderType === 'subscription' && resume.planId) {
    selectedPlan.value = checkout.value.plans.find(plan => plan.id === resume.planId) ?? null
  }

  await router.replace({ path: route.path, query: stripWechatResumeQuery(route.query) })

  if (resume.wechatResumeToken) {
    await createOrder(0, resume.orderType, resume.planId, {
      wechatResumeToken: resume.wechatResumeToken,
      paymentType: resume.paymentType,
      isResume: true,
      renewSeatId: resume.renewSeatId,
    })
    return
  }

  if (resume.orderAmount > 0 && resume.openid) {
    await createOrder(resume.orderAmount, resume.orderType, resume.planId, {
      openid: resume.openid,
      paymentType: resume.paymentType,
      isResume: true,
      renewSeatId: resume.renewSeatId,
    })
  }
}

onMounted(async () => {
  try {
    const res = await paymentAPI.getCheckoutInfo()
    checkout.value = res.data
    if (enabledMethods.value.length) {
      const order: readonly string[] = METHOD_ORDER
      const sorted = [...enabledMethods.value].sort((a, b) => {
        const ai = order.indexOf(a)
        const bi = order.indexOf(b)
        return (ai === -1 ? 999 : ai) - (bi === -1 ? 999 : bi)
      })
      selectedMethod.value = sorted[0]
    }
    if (typeof window !== 'undefined') {
      if (hasWechatResumeQuery(route.query)) {
        removeRecoverySnapshot()
      }
      const routeResumeToken = typeof route.query.resume_token === 'string'
        ? route.query.resume_token
        : typeof route.query.wechat_resume_token === 'string'
          ? route.query.wechat_resume_token
          : undefined
      const restored = readPaymentRecoverySnapshot(
        window.localStorage.getItem(PAYMENT_RECOVERY_STORAGE_KEY),
        { resumeToken: routeResumeToken },
      )
      if (restored) {
        paymentState.value = restored
        paymentPhase.value = 'paying'
        const restoredMethod = normalizeVisibleMethod(restored.paymentType)
        if (restoredMethod) {
          selectedMethod.value = restoredMethod
        }
      } else {
        removeRecoverySnapshot()
      }
    }
    await resumeWechatPaymentFromQuery()
    if (checkout.value.balance_disabled) {
      activeTab.value = 'subscription'
    }
    // Handle renewal navigation: ?tab=subscription&group=123
    if (route.query.tab === 'recharge' && !checkout.value.balance_disabled) {
      activeTab.value = 'recharge'
      selectedPlan.value = null
    }
    if (route.query.tab === 'subscription') {
      activeTab.value = 'subscription'
      if (route.query.group) {
        const groupId = Number(route.query.group)
        const groupPlans = checkout.value.plans.filter(p => p.group_id === groupId)
        if (groupPlans.length === 1) {
          selectedPlan.value = groupPlans[0]
        } else if (groupPlans.length > 1) {
          renewGroupId.value = groupId
          showRenewalModal.value = true
        }
      }
      // 续费独享名额：?plan_id=&renew_seat= 时自动选中对应套餐 + 标记为续费订单
      if (route.query.plan_id && route.query.renew_seat) {
        const planId = Number(route.query.plan_id)
        const seatId = Number(route.query.renew_seat)
        const inListPlan = checkout.value.plans.find(p => p.id === planId)
        if (seatId > 0) {
          // 续费允许已下架（for_sale=false）的 plan，但 checkout 列表只含在售套餐 →
          // find 失败时用 previewRenewal 返回的完整 plan 兜底（GPT round 28 #1）。
          // 这样管理员下架老套餐后老用户仍能续费。
          if (inListPlan) {
            selectedPlan.value = inListPlan
            pendingRenewSeatId.value = seatId
          }
          paymentAPI.previewRenewal(seatId).then(res => {
            renewLastPaidPrice.value = res.data.last_paid_price ?? 0
            if (!inListPlan && res.data.plan) {
              // 仅在 checkout 列表没有时才用 preview 返回的 plan，避免覆盖正在售卖的最新数据
              selectedPlan.value = res.data.plan
              pendingRenewSeatId.value = seatId
            }
          }).catch(() => {
            // 失败不影响主流程，banner 退化为通用提示
            renewLastPaidPrice.value = 0
          })
        }
      }
    }
  } catch (err: unknown) { appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error'))) }
  finally { loading.value = false }
  // Fetch active subscriptions (uses cache, non-blocking)
  subscriptionStore.fetchActiveSubscriptions().catch(() => {})
})
</script>

<style scoped>
/* 充值结算栏：amber 渐变背景 + 顶部白色高光 — 视觉锚点，让"实付金额"立刻吸引视线 */
.recharge-summary {
  background:
    radial-gradient(circle at 50% 0%, rgb(255 255 255 / 0.6), transparent 50%),
    radial-gradient(circle at 100% 100%, rgb(249 115 22 / 0.18), transparent 50%),
    linear-gradient(135deg, rgb(254 243 199), rgb(253 230 138) 60%, rgb(252 211 77));
  position: relative;
}

:root.dark .recharge-summary {
  background:
    radial-gradient(circle at 50% 0%, rgb(245 158 11 / 0.10), transparent 50%),
    radial-gradient(circle at 100% 100%, rgb(249 115 22 / 0.14), transparent 50%),
    linear-gradient(135deg, rgb(15 23 42 / 0.85), rgb(17 24 39 / 0.9));
}
</style>
