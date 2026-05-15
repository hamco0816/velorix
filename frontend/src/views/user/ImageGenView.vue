<template>
  <AppLayout>
    <div class="grid gap-4 lg:grid-cols-[260px_minmax(0,1fr)]">
      <!-- 左 rail：分组选择（仅 lg+ 显示）-->
      <aside class="hidden lg:block">
        <div class="surface-card sticky top-4 p-3">
          <header class="flex items-center justify-between px-1 pb-2">
            <h3 class="text-xs font-semibold uppercase tracking-wider text-gray-500 dark:text-dark-400">
              {{ t('imageGen.groupRailTitle') }}
            </h3>
            <span v-if="imageCapableGroups.length > 0" class="text-[11px] text-gray-400 dark:text-dark-500">
              {{ imageCapableGroups.length }}
            </span>
          </header>
          <p v-if="imageCapableGroups.length === 0" class="px-1 py-3 text-[12px] text-gray-400 dark:text-dark-500">
            {{ t('imageGen.groupEmpty') }}
          </p>
          <div v-else class="space-y-1.5">
            <button
              v-for="entry in imageCapableGroups"
              :key="entry.group.id"
              type="button"
              :class="[
                'block w-full rounded-lg border px-3 py-2.5 text-left transition',
                entry.group.id === form.groupId
                  ? 'border-blue-500 bg-blue-50/50 ring-1 ring-blue-500/20 dark:border-blue-400 dark:bg-blue-500/10'
                  : 'border-gray-200 bg-white hover:border-gray-300 hover:bg-gray-50/70 dark:border-dark-700/60 dark:bg-dark-900/30 dark:hover:bg-dark-800/40',
              ]"
              @click="form.groupId = entry.group.id"
            >
              <div class="flex items-center justify-between gap-2">
                <span class="truncate text-sm font-medium text-gray-900 dark:text-dark-100">
                  {{ entry.group.name }}
                </span>
                <span
                  v-if="entry.group.is_exclusive"
                  class="shrink-0 rounded-sm bg-amber-100 px-1 py-px text-[10px] font-medium text-amber-700 dark:bg-amber-500/15 dark:text-amber-300"
                >
                  {{ t('imageGen.groupExclusive') }}
                </span>
              </div>
              <p class="mt-1 text-[11px] text-gray-500 dark:text-dark-400">
                {{ t('imageGen.groupModelCount', { n: entry.models.length }) }}
              </p>
              <!-- 价格档：admin 配了 image_price_*K 才显示 -->
              <div v-if="hasImagePrice(entry.group)" class="mt-2 flex flex-wrap gap-1">
                <span
                  v-if="entry.group.image_price_1k && entry.group.image_price_1k > 0"
                  class="rounded bg-gray-100 px-1.5 py-0.5 text-[10px] tabular-nums text-gray-600 dark:bg-dark-800/60 dark:text-dark-300"
                >1K ¥{{ entry.group.image_price_1k.toFixed(2) }}</span>
                <span
                  v-if="entry.group.image_price_2k && entry.group.image_price_2k > 0"
                  class="rounded bg-gray-100 px-1.5 py-0.5 text-[10px] tabular-nums text-gray-600 dark:bg-dark-800/60 dark:text-dark-300"
                >2K ¥{{ entry.group.image_price_2k.toFixed(2) }}</span>
                <span
                  v-if="entry.group.image_price_4k && entry.group.image_price_4k > 0"
                  class="rounded bg-gray-100 px-1.5 py-0.5 text-[10px] tabular-nums text-gray-600 dark:bg-dark-800/60 dark:text-dark-300"
                >4K ¥{{ entry.group.image_price_4k.toFixed(2) }}</span>
              </div>
              <!-- 没配单图定价：显示倍率 -->
              <div v-else class="mt-2">
                <span class="rounded bg-gray-100 px-1.5 py-0.5 text-[10px] tabular-nums text-gray-600 dark:bg-dark-800/60 dark:text-dark-300">
                  ×{{ effectiveImageRate(entry.group) }}
                </span>
              </div>
            </button>
          </div>
          <p
            v-if="imageCapableGroups.length > 1"
            class="mt-3 px-1 text-[11px] leading-relaxed text-gray-400 dark:text-dark-500"
          >
            {{ t('imageGen.groupRailHint') }}
          </p>
        </div>
      </aside>

      <!-- 右主区 -->
      <div class="space-y-4 min-w-0">
      <!-- 体验版提示：保持简洁，单行 -->
      <p class="flex items-start gap-1.5 rounded-lg bg-amber-50/60 px-3 py-2 text-[12px] leading-relaxed text-amber-800 dark:bg-amber-500/10 dark:text-amber-300">
        <Icon name="exclamationTriangle" size="xs" class="mt-0.5 shrink-0" />
        <span>{{ t('imageGen.notSavedHint') }}</span>
      </p>

      <!-- 主输入卡片：prompt 在上、工具栏在下、圆形提交在右 -->
      <section class="surface-card p-4 sm:p-5">
        <!-- 1. Prompt 大输入框 -->
        <textarea
          v-model="form.prompt"
          rows="3"
          class="block w-full resize-y rounded-xl border-0 bg-transparent px-2 py-1.5 text-[15px] leading-relaxed text-gray-900 placeholder-gray-400 focus:outline-none focus:ring-0 dark:text-white dark:placeholder-dark-500"
          :placeholder="t('imageGen.promptPlaceholder')"
          :disabled="submitting"
          @keydown="onPromptKeydown"
        />

        <!-- 2. 参考图缩略图（已上传时显示）-->
        <div v-if="referenceImagePreview" class="mt-2 flex items-center gap-3 rounded-lg bg-gray-50 px-3 py-2 dark:bg-dark-800/40">
          <img :src="referenceImagePreview" alt="reference" class="h-12 w-12 rounded-md object-cover ring-1 ring-gray-200 dark:ring-dark-700" />
          <div class="min-w-0 flex-1">
            <p class="truncate text-[13px] font-medium text-gray-800 dark:text-dark-100">{{ referenceImageName }}</p>
            <p class="text-[11px] text-gray-400 dark:text-dark-500">{{ t('imageGen.referencePrivacyHint') }}</p>
          </div>
          <button type="button" class="rounded-md p-1 text-gray-400 hover:bg-gray-200/60 hover:text-gray-700 dark:hover:bg-dark-700" @click="clearReference">
            <Icon name="x" size="sm" />
          </button>
        </div>

        <!-- 3. 工具栏：参考图按钮 / group / model / 比例 / 分辨率 / 张数 / 提交 -->
        <div class="mt-3 flex flex-wrap items-center gap-2 border-t border-gray-100 pt-3 dark:border-dark-700/60">
          <!-- 参考图按钮 -->
          <button
            type="button"
            class="toolbar-btn"
            :title="t('imageGen.referenceUpload')"
            :disabled="submitting"
            @click="triggerFilePicker"
          >
            <Icon name="upload" size="xs" />
            <span class="hidden sm:inline">{{ t('imageGen.referenceShort') }}</span>
          </button>
          <input
            ref="fileInputRef"
            type="file"
            accept="image/png,image/jpeg,image/webp"
            class="hidden"
            @change="onFileSelected"
          />

          <!-- Group 选择：lg+ 由左侧 rail 接管，这里只在中小屏作为兜底显示 -->
          <div v-if="availableGroups.length > 1" class="min-w-[8rem] lg:hidden">
            <Select v-model="form.groupId" :options="groupOptions" :disabled="submitting" />
          </div>

          <!-- Model 选择 -->
          <div class="min-w-[10rem]">
            <Select v-model="form.model" :options="modelOptions" :disabled="submitting || modelOptions.length === 0" />
          </div>

          <!-- 比例 -->
          <div class="min-w-[7rem]">
            <Select v-model="form.aspect" :options="aspectOptions" :disabled="submitting" />
          </div>

          <!-- 分辨率 -->
          <div class="min-w-[8rem]">
            <Select v-model="form.resolution" :options="resolutionOptions" :disabled="submitting" />
          </div>

          <!-- 张数 -->
          <div class="min-w-[5rem]">
            <Select v-model="form.n" :options="countOptions" :disabled="submitting" />
          </div>

          <!-- 圆形提交按钮 -->
          <button
            type="button"
            class="submit-fab ml-auto"
            :class="canSubmit ? 'submit-fab-active' : 'submit-fab-disabled'"
            :disabled="!canSubmit"
            :title="t('imageGen.submitTooltip')"
            @click="generate"
          >
            <span v-if="submitting" class="h-4 w-4 animate-spin rounded-full border-2 border-white border-t-transparent"></span>
            <Icon v-else name="arrowUp" size="sm" :stroke-width="2.5" />
          </button>
        </div>

        <!-- 错误 / 无 key 提示 -->
        <p v-if="apiKeyError" class="mt-3 flex items-start gap-1.5 rounded-md bg-rose-50/60 px-3 py-2 text-[12px] text-rose-700 dark:bg-rose-500/10 dark:text-rose-300">
          <Icon name="exclamationTriangle" size="xs" class="mt-0.5 shrink-0" />
          <span>{{ apiKeyError }}</span>
        </p>
      </section>

      <!-- 4. 结果区 -->
      <section class="surface-card overflow-hidden">
        <header class="flex items-center justify-between border-b border-gray-100 px-5 py-3 dark:border-dark-700/60">
          <h2 class="text-sm font-semibold text-gray-700 dark:text-dark-200">{{ t('imageGen.resultsTitle') }}</h2>
          <span v-if="hasResults" class="text-xs text-gray-400">{{ results.length }} {{ t('imageGen.countLabel') }}</span>
        </header>

        <div class="p-5">
          <!-- 生成中：点阵背景 + 中间预览 + 状态文字 -->
          <div v-if="submitting" class="generating-canvas">
            <div class="dot-grid"></div>
            <div class="relative z-10 flex flex-col items-center gap-3">
              <!-- 中间预览图（partial_image 进来后显示，否则空状态）-->
              <div class="aspect-square w-full max-w-md overflow-hidden rounded-xl bg-white/60 ring-1 ring-gray-200/70 backdrop-blur-sm dark:bg-dark-900/60 dark:ring-dark-700/60">
                <img v-if="currentPreview" :src="currentPreview" alt="generating" class="block h-full w-full object-cover transition-opacity duration-300" />
                <div v-else class="flex h-full w-full items-center justify-center">
                  <Icon name="sparkles" size="xl" class="text-gray-300 dark:text-dark-600 animate-pulse" />
                </div>
              </div>

              <!-- 状态文字 + 进度数字 -->
              <div class="flex items-center gap-2 text-sm font-medium text-gray-700 dark:text-dark-200">
                <span class="h-2 w-2 animate-pulse rounded-full bg-amber-500"></span>
                <span>{{ statusText }}</span>
              </div>
              <p v-if="elapsedSeconds > 0" class="text-[11px] tabular-nums text-gray-400 dark:text-dark-500">
                {{ t('imageGen.elapsed', { sec: elapsedSeconds }) }}
              </p>
            </div>
          </div>

          <!-- 错误态 -->
          <div v-else-if="lastError" class="rounded-lg border border-rose-200/60 bg-rose-50/40 p-4 text-sm dark:border-rose-500/20 dark:bg-rose-500/5">
            <p class="flex items-start gap-2 font-semibold text-rose-800 dark:text-rose-200">
              <Icon name="exclamationTriangle" size="sm" class="mt-0.5 shrink-0" />
              {{ t('imageGen.errorHeader') }}
            </p>
            <p class="ml-6 mt-1 break-words text-rose-700 dark:text-rose-300">{{ lastError }}</p>
            <button class="btn btn-secondary btn-sm ml-6 mt-3" @click="generate">{{ t('imageGen.errorRetry') }}</button>
          </div>

          <!-- 结果网格 -->
          <div v-else-if="hasResults" class="grid gap-3 sm:grid-cols-2 xl:grid-cols-3">
            <div
              v-for="(img, idx) in results"
              :key="idx"
              class="group/img relative overflow-hidden rounded-xl bg-gray-50 ring-1 ring-gray-200/70 dark:bg-dark-800/40 dark:ring-dark-700/60"
            >
              <img :src="img.src" :alt="`generated ${idx + 1}`" class="block w-full" />
              <div class="absolute inset-x-0 bottom-0 flex items-center justify-between gap-2 bg-gradient-to-t from-black/60 to-transparent p-3 opacity-0 transition-opacity group-hover/img:opacity-100">
                <a
                  :href="img.src"
                  target="_blank"
                  rel="noopener"
                  class="inline-flex items-center gap-1 rounded-md bg-white/95 px-2.5 py-1 text-xs font-medium text-gray-800 backdrop-blur hover:bg-white"
                >
                  <Icon name="externalLink" size="xs" />
                  {{ t('imageGen.open') }}
                </a>
                <button
                  type="button"
                  class="inline-flex items-center gap-1 rounded-md bg-white/95 px-2.5 py-1 text-xs font-medium text-gray-800 backdrop-blur hover:bg-white"
                  @click="downloadImage(img, idx)"
                >
                  <Icon name="download" size="xs" />
                  {{ t('imageGen.download') }}
                </button>
              </div>
            </div>
          </div>

          <!-- 空状态 -->
          <div v-else class="flex h-64 items-center justify-center text-sm text-gray-400 dark:text-dark-500">
            <div class="flex flex-col items-center gap-2 text-center">
              <Icon name="sparkles" size="lg" class="text-gray-300 dark:text-dark-600" />
              <span>{{ t('imageGen.emptyHint') }}</span>
            </div>
          </div>
        </div>
      </section>
      </div>
      <!-- /右主区 -->
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import Select, { type SelectOption } from '@/components/common/Select.vue'
import { keysAPI, userChannelsAPI } from '@/api'
import { extractApiErrorMessage } from '@/utils/apiError'
import type { ApiKey } from '@/types'
import type { UserAvailableGroup, UserSupportedModel } from '@/api/channels'

const { t } = useI18n()

// ── 模型能力映射表（硬编码，新模型加进来才能展现）──
// 后端真支持哪些尺寸由上游决定，前端这个表只是 UX 守门——避免用户选不被支持的组合
// 让后端报错。维护时跟着 OpenAI/Gemini 文档更新即可。
type AspectKey = '1:1' | '4:3' | '3:4' | '3:2' | '2:3' | '16:9' | '9:16' | '21:9' | '9:21' | 'auto'
type ResolutionKey = '1K' | '2K' | '4K'

interface ModelCaps {
  aspects: AspectKey[]
  resolutions: ResolutionKey[]
  /** 是否支持流式 partial_image。dall-e-3 暂不支持 */
  streaming: boolean
  /** 是否支持质量参数 */
  quality: boolean
  /** 是否支持以图生图（/v1/images/edits）*/
  edit: boolean
}

const MODEL_CAPS: Record<string, ModelCaps> = {
  'gpt-image-1': { aspects: ['1:1', '2:3', '3:2', 'auto'], resolutions: ['1K'], streaming: true, quality: true, edit: true },
  'gpt-image-1-mini': { aspects: ['1:1', '2:3', '3:2', 'auto'], resolutions: ['1K'], streaming: true, quality: true, edit: true },
  'gpt-image-1.5': { aspects: ['1:1', '2:3', '3:2', 'auto'], resolutions: ['1K'], streaming: true, quality: true, edit: true },
  'gpt-image-2': {
    aspects: ['1:1', '4:3', '3:4', '3:2', '2:3', '16:9', '9:16', '21:9', '9:21', 'auto'],
    resolutions: ['1K', '2K', '4K'],
    streaming: true, quality: true, edit: true,
  },
  'dall-e-3': { aspects: ['1:1', '16:9', '9:16'], resolutions: ['1K'], streaming: false, quality: true, edit: false },
  'gemini-2.5-flash-image-preview': {
    aspects: ['1:1', '4:3', '3:4', '16:9', '9:16', 'auto'],
    resolutions: ['1K'],
    streaming: false, quality: false, edit: true,
  },
}

// 比例 × 分辨率 → 上游 size 字符串
// 不在表里的组合返回 'auto'，让上游自动判定
const SIZE_MAP: Record<ResolutionKey, Partial<Record<AspectKey, string>>> = {
  '1K': {
    '1:1': '1024x1024', '4:3': '1024x768', '3:4': '768x1024',
    '3:2': '1536x1024', '2:3': '1024x1536',
    '16:9': '1792x1024', '9:16': '1024x1792',
    '21:9': '1792x768', '9:21': '768x1792',
    'auto': 'auto',
  },
  '2K': {
    '1:1': '2048x2048', '3:2': '3072x2048', '2:3': '2048x3072',
    '16:9': '3584x2048', '9:16': '2048x3584',
    '21:9': '3584x1536', '9:21': '1536x3584',
    'auto': 'auto',
  },
  '4K': {
    '1:1': '4096x4096', '3:2': '6144x4096', '2:3': '4096x6144',
    '16:9': '7168x4096', '9:16': '4096x7168',
    'auto': 'auto',
  },
}

// ── 状态 ──
interface ResultImage {
  src: string
  isDataUrl: boolean
}
const form = ref({
  prompt: '',
  groupId: null as number | null,
  model: '',
  aspect: 'auto' as AspectKey,
  resolution: '1K' as ResolutionKey,
  n: 1,
})
const apiKeyError = ref('')
const activeKey = ref<ApiKey | null>(null)
// group → models 映射（从 /channels/available 聚合得到的图片可用 group）
const imageCapableGroups = ref<Array<{ group: UserAvailableGroup; models: string[] }>>([])
const referenceFile = ref<File | null>(null)
const referenceImagePreview = ref<string | null>(null)
const referenceImageName = ref<string>('')
const fileInputRef = ref<HTMLInputElement | null>(null)

const results = ref<ResultImage[]>([])
const submitting = ref(false)
const lastError = ref('')
const currentPreview = ref('')
const statusText = ref('')
const startTime = ref(0)
const elapsedSeconds = ref(0)
let elapsedTimer: number | null = null
let statusFallbackTimer: number | null = null

// ── 分组卡片辅助 ──
// 是否配置了按张计费的单图价格（任一档位 > 0）
function hasImagePrice(g: UserAvailableGroup): boolean {
  return Boolean(
    (g.image_price_1k && g.image_price_1k > 0) ||
    (g.image_price_2k && g.image_price_2k > 0) ||
    (g.image_price_4k && g.image_price_4k > 0),
  )
}
// 图片实际折算倍率：独立倍率开关打开且有值时用 image_rate_multiplier，否则回落到 rate_multiplier
function effectiveImageRate(g: UserAvailableGroup): string {
  const rate =
    g.image_rate_independent && typeof g.image_rate_multiplier === 'number'
      ? g.image_rate_multiplier
      : g.rate_multiplier
  return rate.toFixed(2)
}

// ── 计算属性：可选项 ──
const availableGroups = computed(() => imageCapableGroups.value.map((x) => x.group))

const currentGroupEntry = computed(() => {
  if (form.value.groupId == null) return null
  return imageCapableGroups.value.find((x) => x.group.id === form.value.groupId) ?? null
})

const currentModelCaps = computed<ModelCaps | null>(() => {
  return MODEL_CAPS[form.value.model] ?? null
})

const groupOptions = computed<SelectOption[]>(() =>
  availableGroups.value.map((g) => ({ value: g.id, label: g.name }))
)

const modelOptions = computed<SelectOption[]>(() => {
  const entry = currentGroupEntry.value
  if (!entry) return []
  return entry.models.map((m) => ({ value: m, label: m }))
})

const aspectOptions = computed<SelectOption[]>(() => {
  const caps = currentModelCaps.value
  const all: AspectKey[] = ['auto', '1:1', '4:3', '3:4', '3:2', '2:3', '16:9', '9:16', '21:9', '9:21']
  return all.map((a) => ({
    value: a,
    label: a === 'auto' ? t('imageGen.aspectAuto') : a,
    disabled: caps ? !caps.aspects.includes(a) : false,
  }))
})

const resolutionOptions = computed<SelectOption[]>(() => {
  const caps = currentModelCaps.value
  // 价格按 group 的 image_price_*K 字段算（如果配了）；没配就只显示档位
  const g = currentGroupEntry.value?.group
  const price1k = g?.image_price_1k ?? null
  const price2k = g?.image_price_2k ?? null
  const price4k = g?.image_price_4k ?? null
  const fmtPrice = (p: number | null | undefined): string => {
    if (p == null || p <= 0) return ''
    return ` · ¥${p.toFixed(2)}/张`
  }
  return [
    { value: '1K', label: '1K' + fmtPrice(price1k), disabled: caps ? !caps.resolutions.includes('1K') : false },
    { value: '2K', label: '2K' + fmtPrice(price2k), disabled: caps ? !caps.resolutions.includes('2K') : false },
    { value: '4K', label: '4K' + fmtPrice(price4k), disabled: caps ? !caps.resolutions.includes('4K') : false },
  ]
})

// 张数：数字后带单位（避免单独 "1" 这种没上下文的标签困惑用户）
const countOptions = computed<SelectOption[]>(() => [
  { value: 1, label: `1 ${t('imageGen.countLabel')}` },
  { value: 2, label: `2 ${t('imageGen.countLabel')}` },
  { value: 3, label: `3 ${t('imageGen.countLabel')}` },
  { value: 4, label: `4 ${t('imageGen.countLabel')}` },
])

const hasResults = computed(() => results.value.length > 0)
const canSubmit = computed(() =>
  !submitting.value &&
  !!form.value.prompt.trim() &&
  !!activeKey.value &&
  !apiKeyError.value &&
  !!form.value.model
)

// ── 选 group 时联动选合理 model ──
watch(() => form.value.groupId, (gid) => {
  if (gid == null) return
  const entry = imageCapableGroups.value.find((x) => x.group.id === gid)
  if (!entry) return
  // 如果当前 model 不在新 group 的可选列表里，重置为该 group 第一个
  if (!entry.models.includes(form.value.model)) {
    form.value.model = entry.models[0] ?? ''
  }
})

// ── 选 model 时联动重置不兼容的 aspect / resolution ──
watch(() => form.value.model, (m) => {
  const caps = MODEL_CAPS[m]
  if (!caps) return
  if (!caps.aspects.includes(form.value.aspect)) {
    form.value.aspect = caps.aspects[0]
  }
  if (!caps.resolutions.includes(form.value.resolution)) {
    form.value.resolution = caps.resolutions[0]
  }
})

// ── 初始化：拉 API key + 可用 group/model ──
async function loadKey() {
  try {
    const res = await keysAPI.list(1, 10, { status: 'active' })
    const keys = res?.items ?? []
    if (keys.length === 0) {
      apiKeyError.value = t('imageGen.noKeyError')
      return
    }
    activeKey.value = keys[0]
  } catch (err) {
    apiKeyError.value = extractApiErrorMessage(err, t('common.error'))
  }
}

async function loadGroupsAndModels() {
  try {
    const channels = await userChannelsAPI.getAvailable()
    // 聚合：按 group.id 去重；从每个 platform section 提取 image_generation 模式的 model 名
    const groupMap = new Map<number, { group: UserAvailableGroup; models: Set<string> }>()
    for (const ch of channels) {
      for (const sec of ch.platforms) {
        // 命中条件：后端 billing_mode=image，或模型名在前端已知图片模型表内（兜底，
        // 避免后端 LiteLLM 回落定价时把 mode 标记成 token 导致下拉空）
        const imageModels = sec.supported_models
          .filter(
            (m: UserSupportedModel) =>
              m.pricing?.billing_mode === 'image' || m.name in MODEL_CAPS,
          )
          .map((m: UserSupportedModel) => m.name)
        if (imageModels.length === 0) continue
        for (const g of sec.groups) {
          // group 显式关掉「允许生图」才剔除；未设置/未返回该字段时默认放行
          if (g.allow_image_generation === false) continue
          if (!groupMap.has(g.id)) {
            groupMap.set(g.id, { group: g, models: new Set() })
          }
          for (const m of imageModels) groupMap.get(g.id)!.models.add(m)
        }
      }
    }
    imageCapableGroups.value = Array.from(groupMap.values()).map((x) => ({
      group: x.group,
      models: Array.from(x.models),
    }))
    // 默认选第一个 group 和它的第一个 model
    if (imageCapableGroups.value.length > 0) {
      form.value.groupId = imageCapableGroups.value[0].group.id
      form.value.model = imageCapableGroups.value[0].models[0] ?? ''
    }
  } catch {
    // 拉失败也不阻塞，用户可以手动选——但实际上 model dropdown 会是空
  }
}

onMounted(() => {
  loadKey()
  loadGroupsAndModels()
})

onUnmounted(() => {
  if (elapsedTimer) clearInterval(elapsedTimer)
  if (statusFallbackTimer) clearInterval(statusFallbackTimer)
})

// ── 参考图上传 ──
function triggerFilePicker() {
  fileInputRef.value?.click()
}

function onFileSelected(e: Event) {
  const target = e.target as HTMLInputElement
  const file = target.files?.[0]
  if (!file) return
  // 大小限制 4 MB，避免上传失败
  if (file.size > 4 * 1024 * 1024) {
    apiKeyError.value = t('imageGen.referenceTooLarge')
    target.value = ''
    return
  }
  referenceFile.value = file
  referenceImageName.value = file.name
  const reader = new FileReader()
  reader.onload = () => {
    referenceImagePreview.value = String(reader.result ?? '')
  }
  reader.readAsDataURL(file)
  target.value = '' // 允许重复选同一文件
}

function clearReference() {
  referenceFile.value = null
  referenceImagePreview.value = null
  referenceImageName.value = ''
}

// ── 状态文字逻辑 ──
// 真实事件优先，事件之间 5 秒兜底切换
const STATUS_FLOW = ['imageGen.statusPrep', 'imageGen.statusStart', 'imageGen.statusRefine', 'imageGen.statusFinal']
let statusIndex = 0
function setStatusByIndex(i: number) {
  statusIndex = Math.min(i, STATUS_FLOW.length - 1)
  statusText.value = t(STATUS_FLOW[statusIndex])
}
function bumpStatusByEvent(eventType: string, partialIndex?: number, totalPartials?: number) {
  if (eventType === 'in_progress') {
    setStatusByIndex(1)
  } else if (eventType === 'partial_image') {
    const idx = (partialIndex ?? 0) + 1
    statusText.value = t('imageGen.statusRefineWithCount', { idx, total: totalPartials ?? form.value.n })
    statusIndex = 2
  } else if (eventType === 'completed') {
    statusText.value = t('imageGen.statusDone')
    statusIndex = STATUS_FLOW.length - 1
  }
}

// ── 生成 ──
async function generate() {
  if (!canSubmit.value) return
  if (!activeKey.value) {
    apiKeyError.value = t('imageGen.noKeyError')
    return
  }
  submitting.value = true
  lastError.value = ''
  results.value = []
  currentPreview.value = ''
  setStatusByIndex(0)
  startTime.value = Date.now()
  elapsedSeconds.value = 0
  elapsedTimer = window.setInterval(() => {
    elapsedSeconds.value = Math.floor((Date.now() - startTime.value) / 1000)
  }, 1000)
  // 兜底切换：每 5 秒没新事件就往下走一步
  statusFallbackTimer = window.setInterval(() => {
    if (statusIndex < STATUS_FLOW.length - 1) {
      setStatusByIndex(statusIndex + 1)
    }
  }, 5000)

  try {
    const caps = currentModelCaps.value
    const useStream = caps?.streaming ?? false
    const useEdit = !!referenceFile.value && (caps?.edit ?? false)
    const sizeStr = SIZE_MAP[form.value.resolution]?.[form.value.aspect] ?? 'auto'
    const partialImages = useStream ? 2 : 0

    let response: Response
    if (useEdit) {
      // 以图生图：multipart/form-data 走 /v1/images/edits
      // 显式 response_format=b64_json 避免拿到的 URL 1 小时后过期（OpenAI 默认行为）
      const fd = new FormData()
      fd.append('image', referenceFile.value!)
      fd.append('prompt', form.value.prompt.trim())
      fd.append('model', form.value.model)
      fd.append('n', String(form.value.n))
      fd.append('response_format', 'b64_json')
      if (sizeStr !== 'auto') fd.append('size', sizeStr)
      if (useStream) {
        fd.append('stream', 'true')
        fd.append('partial_images', String(partialImages))
      }
      response = await fetch('/v1/images/edits', {
        method: 'POST',
        headers: { Authorization: `Bearer ${activeKey.value.key}` },
        body: fd,
      })
    } else {
      const payload: Record<string, unknown> = {
        prompt: form.value.prompt.trim(),
        model: form.value.model,
        n: form.value.n,
        response_format: 'b64_json',
      }
      if (sizeStr !== 'auto') payload.size = sizeStr
      if (useStream) {
        payload.stream = true
        payload.partial_images = partialImages
      }
      response = await fetch('/v1/images/generations', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${activeKey.value.key}`,
        },
        body: JSON.stringify(payload),
      })
    }

    if (!response.ok) {
      const errText = await response.text()
      let msg = errText
      try {
        const parsed = JSON.parse(errText)
        msg = parsed.error?.message ?? parsed.message ?? errText
      } catch {
        /* keep raw */
      }
      throw new Error(msg || `HTTP ${response.status}`)
    }

    const contentType = response.headers.get('content-type') ?? ''
    if (useStream && contentType.includes('text/event-stream')) {
      await consumeSSE(response)
    } else {
      // 非流式：直接 JSON
      const data = await response.json()
      const list = data.data ?? []
      results.value = list
        .map((item: { b64_json?: string; url?: string }) => {
          if (item.b64_json) return { src: `data:image/png;base64,${item.b64_json}`, isDataUrl: true }
          return { src: item.url ?? '', isDataUrl: false }
        })
        .filter((r: ResultImage) => r.src)
      bumpStatusByEvent('completed')
    }
    if (results.value.length === 0 && currentPreview.value) {
      // partial_image 进了但没收到 completed/最终图，把最后一帧作为结果
      results.value = [{ src: currentPreview.value, isDataUrl: true }]
    }
  } catch (err) {
    lastError.value = err instanceof Error ? err.message : String(err)
  } finally {
    submitting.value = false
    if (elapsedTimer) clearInterval(elapsedTimer)
    if (statusFallbackTimer) clearInterval(statusFallbackTimer)
    elapsedTimer = null
    statusFallbackTimer = null
  }
}

async function consumeSSE(response: Response) {
  const reader = response.body?.getReader()
  if (!reader) throw new Error('No response body')
  const decoder = new TextDecoder()
  let buffer = ''
  while (true) {
    const { done, value } = await reader.read()
    if (done) break
    buffer += decoder.decode(value, { stream: true })
    // SSE 协议：事件以 \n\n 分隔
    const events = buffer.split('\n\n')
    buffer = events.pop() ?? ''
    for (const evt of events) {
      processSSEChunk(evt)
    }
  }
  if (buffer.trim()) processSSEChunk(buffer)
}

function processSSEChunk(chunk: string) {
  const lines = chunk.split('\n')
  let dataStr = ''
  for (const line of lines) {
    if (line.startsWith('data:')) {
      dataStr += line.slice(5).trim()
    }
  }
  if (!dataStr || dataStr === '[DONE]') return
  try {
    const parsed = JSON.parse(dataStr)
    const type: string = parsed.type ?? ''
    if (type.endsWith('partial_image')) {
      const b64 = parsed.partial_image_b64 ?? parsed.b64_json
      if (b64) {
        currentPreview.value = `data:image/png;base64,${b64}`
        bumpStatusByEvent('partial_image', parsed.partial_image_index, 2)
      }
    } else if (type.endsWith('in_progress')) {
      bumpStatusByEvent('in_progress')
    } else if (type.endsWith('completed') || type.endsWith('done')) {
      // 完成事件可能带最终图
      const finalB64 = parsed.image ?? parsed.b64_json ?? parsed.result?.b64_json
      if (finalB64) {
        results.value.push({ src: `data:image/png;base64,${finalB64}`, isDataUrl: true })
      }
      bumpStatusByEvent('completed')
    } else if (parsed.data && Array.isArray(parsed.data)) {
      // 兜底：兼容非流式格式被错误标成 SSE 的情况
      for (const item of parsed.data) {
        if (item.b64_json) results.value.push({ src: `data:image/png;base64,${item.b64_json}`, isDataUrl: true })
        else if (item.url) results.value.push({ src: item.url, isDataUrl: false })
      }
    }
  } catch {
    // 解析失败的 chunk 直接跳过，不阻断流
  }
}

// ── 下载 ──
function downloadImage(img: ResultImage, idx: number) {
  const link = document.createElement('a')
  link.href = img.src
  link.download = `image-${Date.now()}-${idx + 1}.png`
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
}

// ── 快捷键：Cmd/Ctrl + Enter 提交 ──
function onPromptKeydown(e: KeyboardEvent) {
  if (e.key === 'Enter' && (e.metaKey || e.ctrlKey)) {
    e.preventDefault()
    generate()
  }
}
</script>

<style scoped>
/* 工具栏按钮：跟 .input / Select 同高（min-height: 2.25rem = 36px），灰底悬停 */
.toolbar-btn {
  display: inline-flex;
  align-items: center;
  gap: 0.375rem;
  min-height: 2.25rem;
  border-radius: 0.5rem;
  border: 1px solid rgba(226, 232, 240, 0.7);
  background: rgba(255, 255, 255, 1);
  padding: 0 0.75rem;
  font-size: 12px;
  font-weight: 500;
  color: rgb(75 85 99);
  transition: background 0.12s ease, border-color 0.12s ease;
}
.toolbar-btn:hover:not(:disabled) {
  background: rgb(249 250 251);
  border-color: rgba(203, 213, 225, 1);
}
.toolbar-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}
:root.dark .toolbar-btn {
  border-color: rgba(51, 65, 85, 0.6);
  background: rgba(30, 41, 59, 0.4);
  color: rgb(203 213 225);
}
:root.dark .toolbar-btn:hover:not(:disabled) {
  background: rgba(30, 41, 59, 0.7);
}

/* 圆形提交按钮：FAB 风格 */
.submit-fab {
  display: inline-flex;
  height: 36px;
  width: 36px;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
  transition: background 0.12s ease, transform 0.08s ease;
}
.submit-fab-active {
  background: rgb(17 24 39);
  color: white;
}
.submit-fab-active:hover {
  background: rgb(31 41 55);
}
.submit-fab-active:active {
  transform: scale(0.96);
}
.submit-fab-disabled {
  background: rgb(229 231 235);
  color: rgb(156 163 175);
  cursor: not-allowed;
}
:root.dark .submit-fab-disabled {
  background: rgb(51 65 85);
  color: rgb(100 116 139);
}

/* 生成中的画布：点阵背景 + 中央预览 */
.generating-canvas {
  position: relative;
  display: flex;
  min-height: 18rem;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  border-radius: 0.75rem;
  padding: 2rem 1rem;
  background: rgba(249, 250, 251, 0.6);
}
:root.dark .generating-canvas {
  background: rgba(15, 23, 42, 0.4);
}

/* 点阵背景：径向渐变 + 平铺 + 缓慢呼吸（不抢戏的微动画）*/
.dot-grid {
  position: absolute;
  inset: 0;
  background-image: radial-gradient(circle, rgba(15, 23, 42, 0.08) 1px, transparent 1px);
  background-size: 18px 18px;
  animation: dot-pulse 4s ease-in-out infinite;
}
:root.dark .dot-grid {
  background-image: radial-gradient(circle, rgba(226, 232, 240, 0.08) 1px, transparent 1px);
}
@keyframes dot-pulse {
  0%, 100% { opacity: 0.55; }
  50% { opacity: 1; }
}
@media (prefers-reduced-motion: reduce) {
  .dot-grid { animation: none; }
}
</style>
