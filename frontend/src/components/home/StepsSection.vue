<template>
  <!-- 接入流程区块：左侧三步说明 + 右侧 curl 示例代码 -->
  <section id="features" class="py-20 sm:py-24">
    <div class="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
      <SectionHeading
        data-reveal
        :title="t('home.steps.title')"
        :subtitle="t('home.solutions.subtitle')"
      />

      <div class="mt-10 grid gap-6 lg:grid-cols-2 lg:gap-8">
        <!-- 左：三步流程 -->
        <div class="grid gap-2.5">
          <div
            v-for="(step, index) in stepItems"
            :key="step.key"
            :data-reveal="`${index * 80}ms`"
            class="group flex items-start gap-4 rounded-2xl border border-gray-200/70 bg-white p-5 transition-all duration-200 hover:-translate-y-0.5 hover:border-brand-200/80 hover:shadow-card-hover dark:border-dark-700/60 dark:bg-dark-800/40 dark:hover:border-brand-500/30"
          >
            <span class="font-mono text-[22px] font-semibold leading-none tracking-tight text-brand-600 dark:text-brand-400">
              {{ step.num }}
            </span>
            <div class="flex h-10 w-10 flex-shrink-0 items-center justify-center rounded-xl bg-gray-50 text-gray-700 ring-1 ring-inset ring-gray-200/70 dark:bg-dark-700/40 dark:text-gray-200 dark:ring-dark-600/60">
              <Icon :name="step.icon" size="sm" :stroke-width="2" />
            </div>
            <div class="min-w-0 flex-1">
              <h3 class="text-base font-semibold tracking-tight text-gray-900 dark:text-white">
                {{ step.title }}
              </h3>
              <p class="mt-1 text-sm leading-relaxed text-gray-600 dark:text-dark-300">
                {{ step.desc }}
              </p>
            </div>
          </div>
        </div>

        <!-- 右：示例代码块 -->
        <div class="code-block-card" data-reveal="160ms">
          <div class="code-block-header">
            <span class="text-xs font-medium text-gray-300">
              {{ t('home.steps.codeTitle') }}
            </span>
            <span class="font-mono text-2xs text-gray-500">curl</span>
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
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import SectionHeading from '@/components/home/SectionHeading.vue'

defineProps<{
  siteName: string
}>()

const { t } = useI18n()

type StepIcon = 'user' | 'key' | 'link'

interface StepItem {
  key: string
  num: string
  icon: StepIcon
  title: string
  desc: string
}

const stepItems = computed<StepItem[]>(() => [
  { key: 'register', num: t('home.steps.items.register.num'), icon: 'user', title: t('home.steps.items.register.title'), desc: t('home.steps.items.register.desc') },
  { key: 'getKey', num: t('home.steps.items.getKey.num'), icon: 'key', title: t('home.steps.items.getKey.title'), desc: t('home.steps.items.getKey.desc') },
  { key: 'replace', num: t('home.steps.items.replace.num'), icon: 'link', title: t('home.steps.items.replace.title'), desc: t('home.steps.items.replace.desc') }
])
</script>

<style scoped>
/* 示例代码块：深色终端卡片 */
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
