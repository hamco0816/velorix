<template>
  <!-- Custom Home Content: Full Page Mode -->
  <div v-if="homeContent" class="min-h-screen">
    <iframe
      v-if="isHomeContentUrl"
      :src="homeContent.trim()"
      class="h-screen w-full border-0"
      allowfullscreen
    ></iframe>
    <div v-else v-html="homeContent"></div>
  </div>

  <!-- Default Home Page -->
  <div
    v-else
    class="relative min-h-screen overflow-hidden bg-white text-gray-900 dark:bg-dark-950 dark:text-gray-100"
  >
    <!-- ============== Top Navigation ============== -->
    <header
      class="sticky top-0 z-30 border-b border-gray-200/60 bg-white/85 backdrop-blur-md transition-shadow dark:border-dark-800/60 dark:bg-dark-950/85"
      :class="{ 'shadow-card': scrolled }"
    >
      <nav class="mx-auto flex h-16 max-w-7xl items-center justify-between px-4 sm:px-6 lg:px-8">
        <!-- Brand -->
        <div class="flex items-center gap-2.5">
          <div class="flex h-9 w-9 items-center justify-center overflow-hidden rounded-lg ring-1 ring-gray-200 dark:ring-dark-700">
            <img :src="siteLogo || '/logo.png'" alt="Logo" class="h-full w-full object-contain" />
          </div>
          <span class="text-[15px] font-semibold tracking-tight text-gray-900 dark:text-white">
            {{ siteName }}
          </span>
        </div>

        <!-- Center menu (desktop) -->
        <div class="hidden items-center gap-7 md:flex">
          <a href="#features" class="text-[14px] font-medium text-gray-600 transition-colors hover:text-gray-900 dark:text-dark-300 dark:hover:text-white">
            {{ t('home.nav.product') }}
          </a>
          <a href="#pricing" class="text-[14px] font-medium text-gray-600 transition-colors hover:text-gray-900 dark:text-dark-300 dark:hover:text-white">
            {{ t('home.nav.pricing') }}
          </a>
          <a
            v-if="docUrl"
            :href="docUrl"
            target="_blank"
            rel="noopener noreferrer"
            class="text-[14px] font-medium text-gray-600 transition-colors hover:text-gray-900 dark:text-dark-300 dark:hover:text-white"
          >
            {{ t('home.nav.docs') }}
          </a>
          <a href="#providers" class="text-[14px] font-medium text-gray-600 transition-colors hover:text-gray-900 dark:text-dark-300 dark:hover:text-white">
            {{ t('home.nav.status') }}
          </a>
          <a href="#cta" class="text-[14px] font-medium text-gray-600 transition-colors hover:text-gray-900 dark:text-dark-300 dark:hover:text-white">
            {{ t('home.nav.support') }}
          </a>
        </div>

        <!-- Right actions -->
        <div class="flex items-center gap-1 sm:gap-2">
          <LocaleSwitcher />
          <button
            @click="toggleTheme"
            class="rounded-md p-2 text-gray-500 transition-colors hover:bg-gray-100 hover:text-gray-700 dark:text-dark-400 dark:hover:bg-dark-800 dark:hover:text-white"
            :title="isDark ? t('home.switchToLight') : t('home.switchToDark')"
          >
            <Icon v-if="isDark" name="sun" size="sm" />
            <Icon v-else name="moon" size="sm" />
          </button>
          <router-link
            v-if="isAuthenticated"
            :to="dashboardPath"
            class="btn btn-primary btn-sm ml-1"
          >
            {{ t('home.dashboard') }}
            <Icon name="arrowRight" size="xs" :stroke-width="2" />
          </router-link>
          <template v-else>
            <router-link
              to="/login"
              class="btn btn-ghost btn-sm hidden sm:inline-flex"
            >
              {{ t('home.login') }}
            </router-link>
            <router-link to="/login" class="btn btn-primary btn-sm ml-1">
              {{ t('home.nav.register') }}
            </router-link>
          </template>
        </div>
      </nav>
    </header>

    <!-- ============== Hero ============== -->
    <section class="relative overflow-hidden">
      <!-- 装饰：暖色 wash + 网格 + 右上 brand glow -->
      <div class="pointer-events-none absolute inset-0">
        <div class="absolute inset-x-0 top-0 h-[640px] bg-gradient-to-b from-amber-50/70 via-orange-50/30 to-transparent dark:from-brand-500/[0.04] dark:via-amber-500/[0.02]"></div>
        <div
          class="absolute inset-0 bg-[linear-gradient(rgba(120,113,108,0.05)_1px,transparent_1px),linear-gradient(90deg,rgba(120,113,108,0.05)_1px,transparent_1px)] bg-[size:64px_64px] [mask-image:radial-gradient(ellipse_at_top,black_5%,transparent_70%)]"
        ></div>
        <div class="absolute -right-32 top-32 h-[480px] w-[480px] rounded-full bg-brand-300/25 blur-[100px] dark:bg-brand-500/10"></div>
      </div>

      <div class="relative mx-auto max-w-7xl px-4 pb-20 pt-12 sm:px-6 sm:pb-24 sm:pt-16 lg:px-8 lg:pt-20">
        <div class="grid items-center gap-10 lg:grid-cols-12 lg:gap-12">
          <!-- Left copy -->
          <div class="lg:col-span-6">
            <!-- Eyebrow badge -->
            <div class="inline-flex items-center gap-2 rounded-full border border-brand-200 bg-brand-50/80 px-3 py-1 text-xs font-medium text-brand-800 dark:border-brand-500/30 dark:bg-brand-500/10 dark:text-brand-300">
              <span class="h-1.5 w-1.5 rounded-full bg-brand-500 dark:bg-brand-400"></span>
              {{ t('home.heroEyebrowFull') }}
            </div>

            <!-- Display headline：参考图风格，关键词橙色高亮 -->
            <h1 class="mt-6 text-[40px] font-semibold leading-[1.1] tracking-[-0.025em] text-gray-900 dark:text-white sm:text-5xl lg:text-[52px]">
              {{ t('home.heroDisplayPrimary') }}<br />
              {{ t('home.heroDisplayLine2') }}
              <span class="text-brand-600 dark:text-brand-400">{{ t('home.heroDisplayHighlight') }}</span>
            </h1>

            <!-- Lead -->
            <p class="mt-6 max-w-xl text-[15px] leading-relaxed text-gray-600 dark:text-dark-300 sm:text-base">
              {{ t('home.heroLead') }}
            </p>

            <!-- CTAs -->
            <div class="mt-8 flex flex-wrap items-center gap-3">
              <router-link
                :to="isAuthenticated ? dashboardPath : '/login'"
                class="btn btn-primary btn-lg group"
              >
                {{ isAuthenticated ? t('home.goToDashboard') : t('home.getStarted') }}
                <Icon
                  name="arrowRight"
                  size="sm"
                  class="transition-transform duration-200 group-hover:translate-x-0.5"
                  :stroke-width="2"
                />
              </router-link>
              <a href="#pricing" class="btn btn-secondary btn-lg">
                {{ t('home.heroSecondaryCta') }}
              </a>
            </div>

            <!-- 3 个特性 chip：图标 + 主标 + 副标 -->
            <div class="mt-10 grid grid-cols-1 gap-5 sm:grid-cols-3">
              <div class="flex items-center gap-3">
                <div class="flex h-10 w-10 flex-shrink-0 items-center justify-center rounded-xl bg-emerald-50 ring-1 ring-inset ring-emerald-200/70 dark:bg-emerald-500/15 dark:ring-emerald-500/30">
                  <Icon name="check" size="sm" class="text-emerald-600 dark:text-emerald-300" :stroke-width="2.5" />
                </div>
                <div class="min-w-0">
                  <div class="text-sm font-semibold tracking-tight text-gray-900 dark:text-white">{{ t('home.stats.compatible') }}</div>
                  <div class="text-xs text-gray-500 dark:text-dark-400">{{ t('home.stats.compatibleSub') }}</div>
                </div>
              </div>
              <div class="flex items-center gap-3">
                <div class="flex h-10 w-10 flex-shrink-0 items-center justify-center rounded-xl bg-brand-50 ring-1 ring-inset ring-brand-200/70 dark:bg-brand-500/15 dark:ring-brand-500/30">
                  <Icon name="dollar" size="sm" class="text-brand-600 dark:text-brand-300" :stroke-width="2" />
                </div>
                <div class="min-w-0">
                  <div class="text-sm font-semibold tracking-tight text-gray-900 dark:text-white">{{ t('home.stats.payAsYouGo') }}</div>
                  <div class="text-xs text-gray-500 dark:text-dark-400">{{ t('home.stats.payAsYouGoSub') }}</div>
                </div>
              </div>
              <div class="flex items-center gap-3">
                <div class="flex h-10 w-10 flex-shrink-0 items-center justify-center rounded-xl bg-sky-50 ring-1 ring-inset ring-sky-200/70 dark:bg-sky-500/15 dark:ring-sky-500/30">
                  <Icon name="shield" size="sm" class="text-sky-600 dark:text-sky-300" :stroke-width="2" />
                </div>
                <div class="min-w-0">
                  <div class="text-sm font-semibold tracking-tight text-gray-900 dark:text-white">{{ t('home.stats.stable') }}</div>
                  <div class="text-xs text-gray-500 dark:text-dark-400">{{ t('home.stats.stableSub') }}</div>
                </div>
              </div>
            </div>
          </div>

          <!-- Right: Dashboard Preview Card -->
          <div class="lg:col-span-6">
            <DashboardPreview />
          </div>
        </div>
      </div>
    </section>

    <!-- ============== Pain Points ============== -->
    <section id="pain-points" class="border-y border-gray-200 bg-stone-50/70 py-20 dark:border-dark-800 dark:bg-dark-900/30 sm:py-24">
      <div class="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
        <div class="flex items-center gap-3">
          <span class="h-5 w-1 rounded-full bg-brand-500 dark:bg-brand-400"></span>
          <h2 class="text-xl font-semibold tracking-tight text-gray-900 dark:text-white sm:text-2xl">
            {{ t('home.painPoints.title') }}
          </h2>
        </div>

        <div class="mt-8 grid gap-3 sm:grid-cols-2 lg:grid-cols-4 lg:gap-4">
          <div
            v-for="item in painPointItems"
            :key="item.key"
            class="group rounded-2xl border border-gray-200/70 bg-white p-6 transition-colors hover:border-brand-200/80 dark:border-dark-700/60 dark:bg-dark-800/40 dark:hover:border-brand-500/30"
          >
            <!-- 圆角方块图标筐：与全站 ring inset 调统一 -->
            <div class="flex h-10 w-10 items-center justify-center rounded-xl bg-brand-50 text-brand-600 ring-1 ring-inset ring-brand-200/70 dark:bg-brand-500/15 dark:text-brand-300 dark:ring-brand-500/30">
              <Icon :name="item.icon" size="sm" :stroke-width="2" />
            </div>
            <h3 class="mt-4 text-[15px] font-semibold tracking-tight text-gray-900 dark:text-white">
              {{ item.title }}
            </h3>
            <p class="mt-1.5 text-[13px] leading-relaxed text-gray-600 dark:text-dark-300">
              {{ item.desc }}
            </p>
          </div>
        </div>
      </div>
    </section>

    <!-- ============== Solutions / 3 Steps + Code Demo ============== -->
    <section id="features" class="py-20 sm:py-24">
      <div class="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
        <div class="flex items-center gap-3">
          <span class="h-5 w-1 rounded-full bg-brand-500 dark:bg-brand-400"></span>
          <h2 class="text-xl font-semibold tracking-tight text-gray-900 dark:text-white sm:text-2xl">
            {{ t('home.steps.title') }}
          </h2>
        </div>

        <div class="mt-8 grid gap-6 lg:grid-cols-2 lg:gap-8">
          <!-- Left: 3 Steps -->
          <div class="grid gap-2.5">
            <div
              v-for="step in stepItems"
              :key="step.key"
              class="group flex items-start gap-4 rounded-2xl border border-gray-200/70 bg-white p-5 transition-colors hover:border-brand-200/80 dark:border-dark-700/60 dark:bg-dark-800/40 dark:hover:border-brand-500/30"
            >
              <!-- 步骤编号：font-semibold 克制 -->
              <span class="font-mono text-[22px] font-semibold leading-none tracking-tight text-brand-600 dark:text-brand-400">
                {{ step.num }}
              </span>
              <!-- 图标筐 -->
              <div class="flex h-10 w-10 flex-shrink-0 items-center justify-center rounded-xl bg-gray-50 text-gray-700 ring-1 ring-inset ring-gray-200/70 dark:bg-dark-700/40 dark:text-gray-200 dark:ring-dark-600/60">
                <Icon :name="step.icon" size="sm" :stroke-width="2" />
              </div>
              <div class="min-w-0 flex-1">
                <h3 class="text-[15px] font-semibold tracking-tight text-gray-900 dark:text-white">
                  {{ step.title }}
                </h3>
                <p class="mt-1 text-[13px] leading-relaxed text-gray-600 dark:text-dark-300">
                  {{ step.desc }}
                </p>
              </div>
            </div>
          </div>

          <!-- Right: Code Block -->
          <div class="code-block-card">
            <div class="code-block-header">
              <span class="text-xs font-medium text-gray-300">
                {{ t('home.steps.codeTitle') }}
              </span>
              <span class="font-mono text-[11px] text-gray-500">curl</span>
            </div>
            <pre class="code-block-body"><code><span class="c-cmd">curl</span> <span class="c-flag">-X</span> POST <span class="c-url">https://api.{{ siteName.toLowerCase() }}.com/v1/chat/completions</span> \
  <span class="c-flag">-H</span> <span class="c-str">"Authorization: Bearer sk-xxxxxxxxxxxx"</span> \
  <span class="c-flag">-H</span> <span class="c-str">"Content-Type: application/json"</span> \
  <span class="c-flag">-d</span> <span class="c-str">'{{ '{' }}
    "model": "claude-3-5-sonnet-20240620",
    "messages": [{{ '{' }}"role": "user", "content": "Hello, world!"{{ '}' }}]
  {{ '}' }}'</span></code></pre>
          </div>
        </div>
      </div>
    </section>

    <!-- ============== Comparison ============== -->
    <section class="border-y border-gray-200 bg-stone-50/70 py-20 dark:border-dark-800 dark:bg-dark-900/30 sm:py-24">
      <div class="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
        <div class="flex items-center gap-3">
          <span class="h-5 w-1 rounded-full bg-brand-500 dark:bg-brand-400"></span>
          <h2 class="text-xl font-semibold tracking-tight text-gray-900 dark:text-white sm:text-2xl">
            {{ t('home.comparison.title') }}
          </h2>
        </div>

        <div class="mt-8 grid gap-6 lg:grid-cols-2 lg:gap-8">
          <!-- Comparison Table -->
          <div class="overflow-hidden rounded-2xl border border-gray-200/70 bg-white dark:border-dark-700/60 dark:bg-dark-800/40">
            <table class="w-full text-sm">
              <thead>
                <tr class="bg-gray-50/60 dark:bg-dark-800/40">
                  <th class="px-4 py-3.5 text-left text-[12px] font-medium text-gray-500 dark:text-dark-400 sm:px-5">
                    {{ t('home.comparison.headers.feature') }}
                  </th>
                  <th class="px-4 py-3.5 text-left text-[12px] font-medium text-gray-500 dark:text-dark-400 sm:px-5">
                    {{ t('home.comparison.headers.official') }}
                  </th>
                  <th class="bg-brand-50/60 px-4 py-3.5 text-left text-[12px] font-medium text-brand-700 dark:bg-brand-500/10 dark:text-brand-300 sm:px-5">
                    {{ siteName }} {{ t('home.comparison.headers.us') }}
                  </th>
                </tr>
              </thead>
              <tbody class="divide-y divide-gray-100 dark:divide-dark-700/60">
                <tr v-for="row in comparisonItems" :key="row.key">
                  <td class="px-4 py-3.5 font-medium text-gray-900 dark:text-white sm:px-5">{{ row.feature }}</td>
                  <td class="px-4 py-3.5 text-[13px] text-gray-500 dark:text-dark-400 sm:px-5">
                    <span class="inline-flex items-start gap-1.5">
                      <Icon name="x" size="xs" class="mt-0.5 flex-shrink-0 text-gray-400" :stroke-width="2.5" />
                      <span>{{ row.official }}</span>
                    </span>
                  </td>
                  <td class="bg-brand-50/30 px-4 py-3.5 text-[13px] text-gray-900 dark:bg-brand-500/[0.06] dark:text-gray-100 sm:px-5">
                    <span class="inline-flex items-start gap-1.5">
                      <Icon name="check" size="xs" class="mt-0.5 flex-shrink-0 text-brand-600 dark:text-brand-400" :stroke-width="2.5" />
                      <span class="font-medium">{{ row.us }}</span>
                    </span>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>

          <!-- Provider grid (右侧)：8 个 logo -->
          <div id="providers">
            <div class="flex items-center gap-2">
              <h3 class="text-base font-semibold tracking-tight text-gray-900 dark:text-white">
                {{ t('home.providers.title') }}
              </h3>
            </div>
            <p class="mt-1.5 text-xs text-gray-500 dark:text-dark-400">
              {{ t('home.providers.description') }}
            </p>

            <div class="mt-5 grid grid-cols-2 gap-2.5">
              <div
                v-for="provider in providerCards"
                :key="provider.key"
                class="group flex items-center gap-3 rounded-xl border border-gray-200/70 bg-white p-3.5 transition-colors hover:border-gray-300 dark:border-dark-700/60 dark:bg-dark-800/40"
              >
                <div :class="['flex h-10 w-10 flex-shrink-0 items-center justify-center rounded-xl', provider.bg]">
                  <BrandIcon v-if="provider.brand" :brand="provider.brand" size="22px" />
                  <span v-else :class="['text-sm font-semibold', provider.textClass || 'text-gray-700']">{{ provider.letter }}</span>
                </div>
                <div class="min-w-0 flex-1">
                  <p class="text-[13px] font-semibold tracking-tight text-gray-900 dark:text-white">{{ provider.name }}</p>
                  <p class="mt-0.5 inline-flex items-center gap-1.5 text-[11px] font-medium" :class="provider.disabled ? 'text-gray-500 dark:text-dark-400' : 'text-emerald-600 dark:text-emerald-400'">
                    <span class="h-1.5 w-1.5 rounded-full" :class="provider.disabled ? 'bg-gray-400' : 'bg-emerald-500 animate-pulse'"></span>
                    {{ provider.disabled ? t('home.providers.soon') : t('home.providers.supported') }}
                  </p>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- ============== Pricing：后台配置驱动，无套餐时整段隐藏 ============== -->
    <section v-if="plans.length > 0" id="pricing" class="py-20 sm:py-24">
      <div class="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
        <div class="flex items-center gap-3">
          <span class="h-5 w-1 rounded-full bg-brand-500 dark:bg-brand-400"></span>
          <h2 class="text-xl font-semibold tracking-tight text-gray-900 dark:text-white sm:text-2xl">
            {{ t('home.pricing.title') }}
          </h2>
        </div>
        <p class="mt-2 text-[14px] text-gray-600 dark:text-dark-300">{{ t('home.pricing.subtitle') }}</p>

        <div class="mt-8 grid gap-6 lg:grid-cols-2 xl:grid-cols-[2fr_1fr]">
          <!-- 套餐卡片网格 -->
          <div class="grid gap-4" :class="planGridColsClass">
            <div
              v-for="plan in plans"
              :key="plan.id"
              class="relative flex flex-col rounded-2xl p-6 transition-colors"
              :class="plan.recommended
                ? 'border border-gray-900 bg-gray-950 text-white dark:border-white'
                : 'border border-gray-200/70 bg-white hover:border-gray-300 dark:border-dark-700/60 dark:bg-dark-800/40'"
            >
              <span
                v-if="plan.recommended"
                class="absolute -top-2.5 right-6 inline-flex items-center rounded-full bg-brand-500 px-2.5 py-0.5 text-[11px] font-semibold text-white"
              >
                {{ plan.badgeText }}
              </span>

              <h3 class="text-[15px] font-semibold tracking-tight" :class="plan.recommended ? 'text-white' : 'text-gray-900 dark:text-white'">
                {{ plan.name }}
              </h3>
              <p v-if="plan.description" class="mt-1 text-[12px]" :class="plan.recommended ? 'text-gray-400' : 'text-gray-500 dark:text-dark-400'">
                {{ plan.description }}
              </p>

              <div class="mt-5 flex items-baseline gap-2">
                <span class="text-3xl font-semibold tabular-nums tracking-tight" :class="plan.recommended ? 'text-white' : 'text-gray-900 dark:text-white'">
                  ¥ {{ plan.price }}
                </span>
                <span v-if="plan.originalPrice && plan.originalPrice > plan.price" class="text-sm text-gray-400 line-through tabular-nums">
                  ¥ {{ plan.originalPrice }}
                </span>
              </div>
              <p v-if="plan.validityLabel" class="mt-1 text-[12px]" :class="plan.recommended ? 'text-gray-400' : 'text-gray-500 dark:text-dark-400'">
                {{ plan.validityLabel }}
              </p>

              <ul v-if="plan.features.length > 0" class="mt-6 space-y-2.5 text-[13px]" :class="plan.recommended ? 'text-gray-200' : 'text-gray-700 dark:text-dark-200'">
                <li v-for="feat in plan.features" :key="feat" class="flex items-start gap-2">
                  <Icon name="check" size="xs" class="mt-0.5 flex-shrink-0" :class="plan.recommended ? 'text-brand-400' : 'text-emerald-500'" :stroke-width="2.5" />
                  <span>{{ feat }}</span>
                </li>
              </ul>

              <router-link
                :to="isAuthenticated ? '/purchase' : '/login?redirect=/purchase'"
                class="mt-6 inline-flex w-full items-center justify-center gap-2 rounded-md px-3.5 py-2 text-sm font-semibold transition-colors"
                :class="plan.recommended
                  ? 'bg-white text-gray-900 hover:bg-gray-100'
                  : 'bg-primary-950 text-white hover:bg-primary-800 dark:bg-white dark:text-primary-950 dark:hover:bg-primary-100'"
              >
                {{ isAuthenticated ? t('home.pricing.plans.developer.cta') : t('home.cta.button') }}
                <Icon name="arrowRight" size="xs" :stroke-width="2" />
              </router-link>
            </div>
          </div>

          <!-- Right benefits column -->
          <div class="grid gap-x-5 gap-y-5 sm:grid-cols-2 xl:grid-cols-1 xl:gap-y-6 xl:self-start">
            <div v-for="b in benefitItems" :key="b.key" class="flex items-start gap-3">
              <div class="flex h-10 w-10 flex-shrink-0 items-center justify-center rounded-xl bg-brand-50 text-brand-600 ring-1 ring-inset ring-brand-200/70 dark:bg-brand-500/15 dark:text-brand-300 dark:ring-brand-500/30">
                <Icon :name="b.icon" size="sm" :stroke-width="2" />
              </div>
              <div class="min-w-0">
                <div class="text-[14px] font-semibold tracking-tight text-gray-900 dark:text-white">{{ b.title }}</div>
                <div class="mt-0.5 text-[12px] leading-relaxed text-gray-500 dark:text-dark-400">{{ b.desc }}</div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- ============== Trust FAQ：直面"会不会跑路"等顾虑 ============== -->
    <section id="faq" class="border-y border-gray-200 bg-stone-50/70 py-20 dark:border-dark-800 dark:bg-dark-900/30 sm:py-24">
      <div class="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
        <div class="flex items-center gap-3">
          <span class="h-5 w-1 rounded-full bg-brand-500 dark:bg-brand-400"></span>
          <h2 class="text-xl font-semibold tracking-tight text-gray-900 dark:text-white sm:text-2xl">
            {{ t('home.faq.title') }}
          </h2>
        </div>
        <p class="mt-2 text-[14px] text-gray-600 dark:text-dark-300">{{ t('home.faq.subtitle') }}</p>

        <div class="mt-8 grid gap-3 sm:grid-cols-2 lg:gap-4">
          <div
            v-for="item in faqItems"
            :key="item.key"
            class="rounded-2xl border border-gray-200/70 bg-white p-6 transition-colors hover:border-brand-200/80 dark:border-dark-700/60 dark:bg-dark-800/40 dark:hover:border-brand-500/30"
          >
            <h3 class="flex items-start gap-2 text-[15px] font-semibold tracking-tight text-gray-900 dark:text-white">
              <Icon name="checkCircle" size="sm" class="mt-0.5 shrink-0 text-brand-500 dark:text-brand-400" :stroke-width="2" />
              {{ item.q }}
            </h3>
            <p class="mt-2 pl-7 text-[13px] leading-relaxed text-gray-600 dark:text-dark-300">
              {{ item.a }}
            </p>
          </div>
        </div>
      </div>
    </section>

    <!-- ============== Final CTA ============== -->
    <section id="cta" class="px-4 pb-20 sm:px-6 lg:px-8">
      <div class="relative mx-auto max-w-7xl overflow-hidden rounded-2xl bg-gray-950">
        <div class="pointer-events-none absolute inset-0">
          <div
            class="absolute inset-0 bg-[linear-gradient(rgba(255,255,255,0.025)_1px,transparent_1px),linear-gradient(90deg,rgba(255,255,255,0.025)_1px,transparent_1px)] bg-[size:48px_48px] [mask-image:radial-gradient(ellipse_at_top_right,black_5%,transparent_75%)]"
          ></div>
          <div class="absolute -right-32 -top-32 h-[420px] w-[420px] rounded-full bg-brand-500/25 blur-[100px]"></div>
          <div class="absolute -bottom-40 -left-20 h-[420px] w-[420px] rounded-full bg-amber-400/12 blur-[100px]"></div>
        </div>

        <div class="relative grid items-center gap-8 p-10 sm:grid-cols-12 sm:p-14">
          <div class="sm:col-span-7">
            <h2 class="text-xl font-semibold tracking-tight text-white sm:text-3xl">
              {{ t('home.cta.title') }}
            </h2>
            <p class="mt-3 text-[14px] leading-relaxed text-gray-300/85">
              {{ t('home.cta.description') }}
            </p>
          </div>

          <div class="flex flex-wrap items-center gap-3 sm:col-span-5 sm:justify-end">
            <router-link
              :to="isAuthenticated ? dashboardPath : '/login'"
              class="group inline-flex items-center justify-center gap-2 rounded-md bg-white px-5 py-2.5 text-sm font-semibold text-gray-900 transition-colors hover:bg-gray-100"
            >
              {{ isAuthenticated ? t('home.goToDashboard') : t('home.cta.button') }}
              <Icon name="arrowRight" size="xs" class="transition-transform duration-200 group-hover:translate-x-0.5" :stroke-width="2" />
            </router-link>
            <a
              v-if="docUrl"
              :href="docUrl"
              target="_blank"
              rel="noopener noreferrer"
              class="inline-flex items-center justify-center gap-2 rounded-md bg-white/[0.08] px-5 py-2.5 text-sm font-semibold text-white ring-1 ring-inset ring-white/15 transition-colors hover:bg-white/[0.12]"
            >
              <Icon name="book" size="xs" :stroke-width="2" />
              {{ t('home.viewDocs') }}
            </a>
          </div>
        </div>
      </div>
    </section>

    <!-- ============== Footer ============== -->
    <footer class="border-t border-gray-200 bg-white dark:border-dark-800 dark:bg-dark-950">
      <div class="mx-auto max-w-7xl px-4 py-12 sm:px-6 lg:px-8">
        <div class="grid gap-10 sm:grid-cols-2 lg:grid-cols-3">
          <!-- Brand col -->
          <div class="lg:col-span-1">
            <div class="flex items-center gap-2">
              <div class="flex h-8 w-8 items-center justify-center overflow-hidden rounded-md ring-1 ring-gray-200 dark:ring-dark-700">
                <img :src="siteLogo || '/logo.png'" alt="Logo" class="h-full w-full object-contain" />
              </div>
              <span class="text-sm font-semibold tracking-tight text-gray-900 dark:text-white">{{ siteName }}</span>
            </div>
            <p class="mt-3 text-[12px] leading-relaxed text-gray-500 dark:text-dark-400">
              {{ t('home.heroLead') }}
            </p>
            <p class="mt-4 text-[11px] text-gray-400 dark:text-dark-500">
              &copy; {{ currentYear }} {{ siteName }}. {{ t('home.footer.allRightsReserved') }}
            </p>
          </div>

          <!-- Links: Product -->
          <div>
            <h4 class="text-[13px] font-semibold tracking-tight text-gray-900 dark:text-white">{{ t('home.nav.product') }}</h4>
            <ul class="mt-3 space-y-2.5 text-[13px]">
              <li><a href="#features" class="text-gray-500 transition-colors hover:text-gray-900 dark:text-dark-400 dark:hover:text-white">{{ t('home.steps.title') }}</a></li>
              <li v-if="plans.length > 0"><a href="#pricing" class="text-gray-500 transition-colors hover:text-gray-900 dark:text-dark-400 dark:hover:text-white">{{ t('home.nav.pricing') }}</a></li>
              <li><a v-if="docUrl" :href="docUrl" target="_blank" rel="noopener noreferrer" class="text-gray-500 transition-colors hover:text-gray-900 dark:text-dark-400 dark:hover:text-white">{{ t('home.nav.docs') }}</a></li>
            </ul>
          </div>

          <!-- Links: Support -->
          <div>
            <h4 class="text-[13px] font-semibold tracking-tight text-gray-900 dark:text-white">{{ t('home.nav.support') }}</h4>
            <ul class="mt-3 space-y-2.5 text-[13px]">
              <li><router-link to="/login" class="text-gray-500 transition-colors hover:text-gray-900 dark:text-dark-400 dark:hover:text-white">{{ t('home.login') }}</router-link></li>
              <li><router-link to="/login" class="text-gray-500 transition-colors hover:text-gray-900 dark:text-dark-400 dark:hover:text-white">{{ t('home.nav.register') }}</router-link></li>
              <li><a href="#cta" class="text-gray-500 transition-colors hover:text-gray-900 dark:text-dark-400 dark:hover:text-white">{{ t('home.nav.support') }}</a></li>
            </ul>
          </div>
        </div>
      </div>
    </footer>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAuthStore, useAppStore } from '@/stores'
import LocaleSwitcher from '@/components/common/LocaleSwitcher.vue'
import Icon from '@/components/icons/Icon.vue'
import BrandIcon from '@/components/common/BrandIcon.vue'
import DashboardPreview from '@/components/home/DashboardPreview.vue'
import { paymentAPI, type PublicPlan } from '@/api/payment'

const { t } = useI18n()

const authStore = useAuthStore()
const appStore = useAppStore()

const siteName = computed(() => appStore.cachedPublicSettings?.site_name || appStore.siteName || 'Sub2API')
const siteLogo = computed(() => appStore.cachedPublicSettings?.site_logo || appStore.siteLogo || '')
const docUrl = computed(() => appStore.cachedPublicSettings?.doc_url || appStore.docUrl || '')
const homeContent = computed(() => appStore.cachedPublicSettings?.home_content || '')

const isHomeContentUrl = computed(() => {
  const content = homeContent.value.trim()
  return content.startsWith('http://') || content.startsWith('https://')
})

const isDark = ref(document.documentElement.classList.contains('dark'))
const scrolled = ref(false)

const isAuthenticated = computed(() => authStore.isAuthenticated)
const isAdmin = computed(() => authStore.isAdmin)
const dashboardPath = computed(() => isAdmin.value ? '/admin/dashboard' : '/dashboard')
const currentYear = computed(() => new Date().getFullYear())

type HomeIconName =
  | 'dollar' | 'cube' | 'bolt' | 'eye' | 'server' | 'users' | 'creditCard'
  | 'check' | 'x' | 'arrowRight' | 'book' | 'sun' | 'moon' | 'plus'
  | 'shield' | 'lock' | 'sparkles' | 'fire' | 'clock' | 'globe' | 'chart'
  | 'key' | 'beaker' | 'sync' | 'edit' | 'database'

interface PainPoint { key: string; icon: HomeIconName; title: string; desc: string }
const painPointItems = computed<PainPoint[]>(() => [
  { key: 'expensive', icon: 'dollar', title: t('home.painPoints.items.expensive.title'), desc: t('home.painPoints.items.expensive.desc') },
  { key: 'complex', icon: 'users', title: t('home.painPoints.items.complex.title'), desc: t('home.painPoints.items.complex.desc') },
  { key: 'unstable', icon: 'chart', title: t('home.painPoints.items.unstable.title'), desc: t('home.painPoints.items.unstable.desc') },
  { key: 'noControl', icon: 'shield', title: t('home.painPoints.items.noControl.title'), desc: t('home.painPoints.items.noControl.desc') }
])

interface FaqItem { key: string; q: string; a: string }
const faqItems = computed<FaqItem[]>(() => [
  { key: 'runaway', q: t('home.faq.items.runaway.q'), a: t('home.faq.items.runaway.a') },
  { key: 'invoice', q: t('home.faq.items.invoice.q'), a: t('home.faq.items.invoice.a') },
  { key: 'refund', q: t('home.faq.items.refund.q'), a: t('home.faq.items.refund.a') },
  { key: 'billing', q: t('home.faq.items.billing.q'), a: t('home.faq.items.billing.a') },
  { key: 'security', q: t('home.faq.items.security.q'), a: t('home.faq.items.security.a') },
  { key: 'support', q: t('home.faq.items.support.q'), a: t('home.faq.items.support.a') }
])

interface StepItem { key: string; num: string; icon: HomeIconName; title: string; desc: string }
const stepItems = computed<StepItem[]>(() => [
  { key: 'register', num: t('home.steps.items.register.num'), icon: 'user' as unknown as HomeIconName, title: t('home.steps.items.register.title'), desc: t('home.steps.items.register.desc') },
  { key: 'getKey', num: t('home.steps.items.getKey.num'), icon: 'key', title: t('home.steps.items.getKey.title'), desc: t('home.steps.items.getKey.desc') },
  { key: 'replace', num: t('home.steps.items.replace.num'), icon: 'link' as unknown as HomeIconName, title: t('home.steps.items.replace.title'), desc: t('home.steps.items.replace.desc') }
])

interface ComparisonRow { key: string; feature: string; official: string; us: string }
const comparisonItems = computed<ComparisonRow[]>(() => [
  { key: 'pricing', feature: t('home.comparison.items.pricing.feature'), official: t('home.comparison.items.pricing.official'), us: t('home.comparison.items.pricing.us') },
  { key: 'ban', feature: t('home.comparison.items.ban.feature'), official: t('home.comparison.items.ban.official'), us: t('home.comparison.items.ban.us') },
  { key: 'models', feature: t('home.comparison.items.models.feature'), official: t('home.comparison.items.models.official'), us: t('home.comparison.items.models.us') },
  { key: 'management', feature: t('home.comparison.items.management.feature'), official: t('home.comparison.items.management.official'), us: t('home.comparison.items.management.us') },
  { key: 'stability', feature: t('home.comparison.items.stability.feature'), official: t('home.comparison.items.stability.official'), us: t('home.comparison.items.stability.us') },
  { key: 'control', feature: t('home.comparison.items.control.feature'), official: t('home.comparison.items.control.official'), us: t('home.comparison.items.control.us') }
])

interface ProviderCard {
  key: string
  name: string
  brand?: 'claude' | 'openai' | 'gemini' | 'deepseek'
  letter?: string
  bg: string
  textClass?: string
  disabled?: boolean
}
const providerCards = computed<ProviderCard[]>(() => [
  { key: 'claude', name: t('home.providers.claude'), brand: 'claude', bg: 'bg-orange-50 ring-1 ring-inset ring-orange-100 dark:bg-orange-500/10 dark:ring-orange-500/20' },
  { key: 'gpt', name: 'ChatGPT', brand: 'openai', bg: 'bg-gray-100 ring-1 ring-inset ring-gray-200 dark:bg-white/10 dark:ring-white/15' },
  { key: 'gemini', name: t('home.providers.gemini'), brand: 'gemini', bg: 'bg-blue-50 ring-1 ring-inset ring-blue-100 dark:bg-blue-500/10 dark:ring-blue-500/20' },
  { key: 'deepseek', name: t('home.providersExt.deepseek'), brand: 'deepseek', bg: 'bg-blue-50 ring-1 ring-inset ring-blue-100 dark:bg-blue-500/10 dark:ring-blue-500/20' }
])

interface BenefitItem { key: string; icon: HomeIconName; title: string; desc: string }
const benefitItems = computed<BenefitItem[]>(() => [
  { key: 'payAsYouGo', icon: 'dollar', title: t('home.pricing.benefits.payAsYouGo.title'), desc: t('home.pricing.benefits.payAsYouGo.desc') },
  { key: 'tieredDiscount', icon: 'chart', title: t('home.pricing.benefits.tieredDiscount.title'), desc: t('home.pricing.benefits.tieredDiscount.desc') },
  { key: 'secure', icon: 'shield', title: t('home.pricing.benefits.secure.title'), desc: t('home.pricing.benefits.secure.desc') },
  { key: 'support247', icon: 'clock', title: t('home.pricing.benefits.support247.title'), desc: t('home.pricing.benefits.support247.desc') }
])

// ========= 后端驱动的订阅套餐 =========
interface DisplayPlan {
  id: number
  name: string
  description: string
  price: number
  originalPrice?: number
  validityLabel: string  // 如 "30 天" / "永久"
  features: string[]
  productName: string
  recommended: boolean
  badgeText: string
}

const plans = ref<DisplayPlan[]>([])
const plansLoading = ref(false)

function parsePlanFeatures(raw: string): string[] {
  if (!raw) return []
  try {
    const parsed = JSON.parse(raw)
    if (Array.isArray(parsed)) return parsed.filter(Boolean).map(String)
  } catch {
    // 后台可能存的是逗号/换行分隔字符串
  }
  return raw.split(/[\n,，]/).map(s => s.trim()).filter(Boolean)
}

function formatValidity(days: number, unit: string): string {
  if (!days || days <= 0) return ''
  const u = (unit || 'day').toLowerCase()
  if (u.startsWith('day')) return `${days} 天`
  if (u.startsWith('month')) return `${days} 个月`
  if (u.startsWith('year')) return `${days} 年`
  return `${days} ${unit}`
}

// 套餐卡片网格列数：根据后台配置的套餐数自适应（1/2/3 列）
const planGridColsClass = computed(() => {
  const n = plans.value.length
  if (n <= 1) return 'grid-cols-1'
  if (n === 2) return 'grid-cols-1 sm:grid-cols-2'
  return 'grid-cols-1 sm:grid-cols-2 lg:grid-cols-3'
})

async function loadPlans() {
  plansLoading.value = true
  try {
    const list = await paymentAPI.getPlansPublic()
    if (Array.isArray(list) && list.length > 0) {
      const sorted = [...list].sort((a, b) => (a.sort_order ?? 0) - (b.sort_order ?? 0))
      plans.value = sorted.map((p: PublicPlan): DisplayPlan => {
        const badgeText = (p.badge_text || '').trim() || (p.is_popular ? t('home.pricing.recommended') : '')
        return {
          id: p.id,
          name: p.name,
          description: p.description || '',
          price: p.price,
          originalPrice: p.original_price,
          validityLabel: formatValidity(p.validity_days, p.validity_unit),
          features: parsePlanFeatures(p.features),
          productName: p.product_name,
          recommended: badgeText !== '',
          badgeText,
        }
      })
    } else {
      plans.value = []
    }
  } catch {
    plans.value = []
  } finally {
    plansLoading.value = false
  }
}

function toggleTheme() {
  isDark.value = !isDark.value
  document.documentElement.classList.toggle('dark', isDark.value)
  localStorage.setItem('theme', isDark.value ? 'dark' : 'light')
}

function initTheme() {
  const savedTheme = localStorage.getItem('theme')
  if (
    savedTheme === 'dark' ||
    (!savedTheme && window.matchMedia('(prefers-color-scheme: dark)').matches)
  ) {
    isDark.value = true
    document.documentElement.classList.add('dark')
  }
}

function onScroll() {
  scrolled.value = window.scrollY > 4
}

onMounted(() => {
  initTheme()
  authStore.checkAuth()
  if (!appStore.publicSettingsLoaded) {
    appStore.fetchPublicSettings()
  }
  loadPlans()
  window.addEventListener('scroll', onScroll, { passive: true })
})

onBeforeUnmount(() => {
  window.removeEventListener('scroll', onScroll)
})
</script>

<style scoped>
/* Code Block Card */
.code-block-card {
  background: linear-gradient(160deg, #1e293b 0%, #0f172a 100%);
  border-radius: 12px;
  box-shadow:
    0 30px 60px -20px rgba(15, 23, 42, 0.4),
    0 0 0 1px rgba(255, 255, 255, 0.06),
    inset 0 1px 0 rgba(255, 255, 255, 0.06);
  overflow: hidden;
}

.code-block-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 18px;
  background: rgba(15, 23, 42, 0.6);
  border-bottom: 1px solid rgba(255, 255, 255, 0.05);
}

.code-block-body {
  padding: 18px 22px;
  font-family: ui-monospace, 'Fira Code', monospace;
  font-size: 12.5px;
  line-height: 1.85;
  color: #cbd5e1;
  overflow-x: auto;
  white-space: pre;
}

.code-block-body .c-cmd {
  color: #38bdf8;
  font-weight: 600;
}
.code-block-body .c-flag {
  color: #c4b5fd;
}
.code-block-body .c-url {
  color: #2dd4bf;
}
.code-block-body .c-str {
  color: #fbbf24;
}
</style>
