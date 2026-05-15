<template>
  <AppLayout wide>
    <div class="grid gap-4 lg:grid-cols-[224px_minmax(0,1fr)]">
      <!-- 左 rail：分组选择（仅 lg+ 显示，与右侧 surface-card 同一设计语言）-->
      <aside class="hidden lg:block">
        <div class="surface-card sticky top-4 overflow-hidden">
          <header class="flex items-center justify-between border-b border-gray-100 px-3.5 py-2.5 dark:border-dark-700/60">
            <h3 class="text-[12px] font-semibold text-gray-700 dark:text-dark-200">
              {{ t('imageGen.groupRailTitle') }}
            </h3>
            <span v-if="imageCapableGroups.length > 0" class="rounded-full bg-gray-100 px-1.5 py-px text-[11px] tabular-nums text-gray-500 dark:bg-dark-800 dark:text-dark-400">
              {{ imageCapableGroups.length }}
            </span>
          </header>

          <!-- 空态 -->
          <div
            v-if="imageCapableGroups.length === 0"
            class="px-3.5 py-5 text-[12px] leading-relaxed text-gray-500 dark:text-dark-400"
          >
            <p class="font-medium text-gray-600 dark:text-dark-300">{{ t('imageGen.groupEmpty') }}</p>
            <p class="mt-1.5 text-[11px] text-gray-400 dark:text-dark-500">{{ t('imageGen.groupEmptyHint') }}</p>
          </div>

          <!-- 分组列表 -->
          <div v-else class="p-1.5">
            <button
              v-for="entry in imageCapableGroups"
              :key="entry.group.id"
              type="button"
              :class="[
                'block w-full rounded-lg px-2.5 py-2 text-left transition',
                entry.group.id === form.groupId
                  ? 'bg-blue-50 dark:bg-blue-500/10'
                  : 'hover:bg-gray-50 dark:hover:bg-dark-800/40',
              ]"
              @click="form.groupId = entry.group.id"
            >
              <div class="flex items-center gap-1.5">
                <span
                  :class="[
                    'min-w-0 flex-1 truncate text-[13px]',
                    entry.group.id === form.groupId
                      ? 'font-semibold text-blue-700 dark:text-blue-300'
                      : 'font-medium text-gray-800 dark:text-dark-200',
                  ]"
                >
                  {{ entry.group.name }}
                </span>
                <span
                  v-if="entry.group.is_exclusive"
                  class="shrink-0 rounded bg-amber-100 px-1 py-px text-[10px] font-medium text-amber-700 dark:bg-amber-500/15 dark:text-amber-300"
                >
                  {{ t('imageGen.groupExclusive') }}
                </span>
              </div>
              <!-- 二级信息：价格/倍率 -->
              <p class="mt-1 truncate text-[11px] tabular-nums text-gray-500 dark:text-dark-500">
                <template v-if="hasImagePrice(entry.group)">{{ priceSummary(entry.group) }}</template>
                <template v-else>{{ t('imageGen.groupRateLabel') }} ×{{ effectiveImageRate(entry.group) }}</template>
              </p>
            </button>
          </div>
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

        <!-- 2. 参考图缩略图（可多张，已上传时显示）-->
        <div v-if="referencePreviews.length" class="mt-2 rounded-lg bg-gray-50 px-3 py-2.5 dark:bg-dark-800/40">
          <div class="flex flex-wrap gap-2">
            <div
              v-for="(ref, idx) in referencePreviews"
              :key="ref.url"
              class="group/ref relative h-14 w-14 overflow-hidden rounded-md ring-1 ring-gray-200 dark:ring-dark-700"
            >
              <img :src="ref.url" :alt="ref.name" class="h-full w-full object-cover" />
              <button
                type="button"
                class="absolute inset-0 flex items-center justify-center bg-black/45 opacity-0 transition-opacity group-hover/ref:opacity-100"
                :title="t('imageGen.referenceRemove')"
                @click="removeReference(idx)"
              >
                <Icon name="x" size="sm" class="text-white" />
              </button>
            </div>
          </div>
          <p class="mt-1.5 text-[11px] text-gray-400 dark:text-dark-500">
            {{ t('imageGen.referencePrivacyHint') }}
          </p>
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
            multiple
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
        <div v-if="keyError || referenceError" class="mt-3 rounded-md bg-rose-50/60 px-3 py-2.5 text-[12px] text-rose-700 dark:bg-rose-500/10 dark:text-rose-300">
          <p class="flex items-start gap-1.5">
            <Icon name="exclamationTriangle" size="xs" class="mt-0.5 shrink-0" />
            <span>{{ keyError || referenceError }}</span>
          </p>
          <!-- 无密钥时页内一键创建，省去跳「API 密钥」页 -->
          <button
            v-if="canCreateKeyInline"
            type="button"
            class="mt-2 inline-flex items-center gap-1.5 rounded-md bg-rose-600 px-3 py-1.5 text-[12px] font-medium text-white transition hover:bg-rose-700 disabled:cursor-not-allowed disabled:opacity-60 dark:bg-rose-600 dark:hover:bg-rose-500"
            :disabled="creatingKey"
            @click="createKeyInline"
          >
            <span v-if="creatingKey" class="h-3 w-3 animate-spin rounded-full border-2 border-white border-t-transparent"></span>
            <Icon v-else name="plus" size="xs" />
            {{ creatingKey ? t('imageGen.creatingKey') : t('imageGen.createKeyHere') }}
          </button>
        </div>
      </section>

      <!-- 4. 结果区 -->
      <section class="surface-card overflow-hidden">
        <header class="flex items-center justify-between border-b border-gray-100 px-5 py-3 dark:border-dark-700/60">
          <h2 class="text-sm font-semibold text-gray-700 dark:text-dark-200">{{ t('imageGen.resultsTitle') }}</h2>
          <span v-if="hasResults" class="text-xs text-gray-400">{{ results.length }} {{ t('imageGen.countLabel') }}</span>
        </header>

        <div class="p-5">
          <!-- 生成中：深色画布 + 居中预览 + 明确"非最终图"标识 + 扫描动效 -->
          <div v-if="submitting" class="generating-canvas">
            <div class="dot-grid"></div>
            <div class="relative z-10 flex w-full max-w-lg flex-col items-center gap-4">
              <div class="relative aspect-square w-full overflow-hidden rounded-2xl ring-1 ring-white/10">
                <template v-if="currentPreview">
                  <!-- 中间帧：降饱和+轻微暗化，读起来明显"还没完成" -->
                  <img
                    :src="currentPreview"
                    alt="rendering preview"
                    class="block h-full w-full object-cover opacity-90 saturate-[0.85]"
                  />
                  <!-- 扫描光带 -->
                  <div class="scan-beam"></div>
                  <!-- 左上角"非最终图"角标 -->
                  <div class="absolute left-2.5 top-2.5 flex items-center gap-1.5 rounded-md bg-amber-500/90 px-2 py-1 text-[11px] font-semibold text-white shadow-sm backdrop-blur">
                    <span class="h-1.5 w-1.5 animate-pulse rounded-full bg-white"></span>
                    {{ t('imageGen.previewBadge') }}
                  </div>
                </template>
                <div v-else class="flex h-full w-full items-center justify-center">
                  <span class="brush-pulse"><Icon name="sparkles" size="xl" class="text-slate-500" /></span>
                  <div class="scan-beam"></div>
                </div>
              </div>

              <!-- 状态行：进度文字 + 已用时 -->
              <div class="flex flex-col items-center gap-1.5 text-center">
                <div class="flex items-center gap-2 text-sm font-medium text-slate-200">
                  <span class="h-2 w-2 animate-pulse rounded-full bg-amber-400"></span>
                  <span>{{ statusText }}</span>
                  <span v-if="elapsedSeconds > 0" class="text-[11px] tabular-nums text-slate-400">
                    · {{ t('imageGen.elapsed', { sec: elapsedSeconds }) }}
                  </span>
                </div>
                <p class="max-w-xs text-[11px] leading-relaxed text-slate-400">
                  {{ t('imageGen.previewHint') }}
                </p>
              </div>
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

          <!-- 结果：单图居中限宽，多图网格；带揭示动画 -->
          <div
            v-else-if="hasResults"
            :class="results.length === 1
              ? 'mx-auto max-w-xl'
              : 'grid gap-3 sm:grid-cols-2 xl:grid-cols-3'"
          >
            <div
              v-for="(img, idx) in results"
              :key="idx"
              class="reveal-img group/img relative overflow-hidden rounded-xl bg-gray-50 ring-1 ring-gray-200/70 dark:bg-dark-800/40 dark:ring-dark-700/60"
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

// 平台官方图片模型目录。透传渠道的 supported_models 天生为空（模型配在账号上、
// 渠道不限制即支持上游全部模型），此时这个目录就是"用户能调哪些图片模型"的事实来源，
// 不是猜测。维护时跟着 MODEL_CAPS 一起更新即可。
const IMAGE_MODEL_CATALOG: Record<string, string[]> = {
  openai: ['gpt-image-1.5', 'gpt-image-2', 'gpt-image-1', 'gpt-image-1-mini', 'dall-e-3'],
  gemini: ['gemini-2.5-flash-image-preview'],
}

// 比例 × 分辨率 → 上游 size 字符串。
// 关键：这些串必须与后端 normalizeOpenAIImageSizeTier 的判定一致，否则
// 用户选 1K 实际被按 2K 计费。后端只认这些精确串：
//   1K = 1024x1024（仅 1:1）
//   2K = 1536x1024 / 1024x1536 / 1792x1024 / 1024x1792 / 2048x2048 / 2048x1152 / 1152x2048
//   4K = 3840x2160 / 2160x3840（其余按像素阈值归 2K/4K）
// 所以每档只提供能被后端正确归类到该档的比例；始终下发具体 size，绝不发 auto
// （后端把空/auto 默认当 2K，会导致 1K 计费错档）。
const SIZE_MAP: Record<ResolutionKey, Partial<Record<AspectKey, string>>> = {
  '1K': {
    '1:1': '1024x1024',
    'auto': '1024x1024',
  },
  '2K': {
    '1:1': '2048x2048',
    '3:2': '1536x1024', '2:3': '1024x1536',
    '16:9': '2048x1152', '9:16': '1152x2048',
    'auto': '2048x2048',
  },
  '4K': {
    '1:1': '4096x4096',
    '16:9': '3840x2160', '9:16': '2160x3840',
    'auto': '4096x4096',
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
// 参考图校验错误（不阻塞普通生成）
const referenceError = ref('')
// apiKeyError 仅用于硬错误：拉取失败 / 创建密钥失败
const apiKeyError = ref('')
// 用户全部 active 密钥；实际用哪个由选中分组决定（见 activeKey computed）
const activeKeys = ref<ApiKey[]>([])
// 页内创建密钥的进行中状态
const creatingKey = ref(false)
// group → models 映射（从 /channels/available 聚合得到的图片可用 group）
const imageCapableGroups = ref<Array<{ group: UserAvailableGroup; models: string[] }>>([])
// 参考图：支持多张（以图生图最多 MAX_REFERENCE_IMAGES 张）
const MAX_REFERENCE_IMAGES = 4
const referenceFiles = ref<File[]>([])
const referencePreviews = ref<Array<{ name: string; url: string }>>([])
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
// 价格摘要：rail 单行展示，取最便宜的一档（用户最低门槛）
function priceSummary(g: UserAvailableGroup): string {
  const prices = [g.image_price_1k, g.image_price_2k, g.image_price_4k].filter(
    (p): p is number => typeof p === 'number' && p > 0,
  )
  if (prices.length === 0) return ''
  const min = Math.min(...prices)
  return `¥${min.toFixed(2)}/张起`
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
  // 可选比例 = 模型支持 ∩ 当前分辨率档下后端能正确计费的比例（SIZE_MAP 的 key）
  const sizeKeys = SIZE_MAP[form.value.resolution] ?? {}
  const all: AspectKey[] = ['auto', '1:1', '4:3', '3:4', '3:2', '2:3', '16:9', '9:16', '21:9', '9:21']
  return all.map((a) => ({
    value: a,
    label: a === 'auto' ? t('imageGen.aspectAuto') : a,
    disabled: (caps ? !caps.aspects.includes(a) : false) || !(a in sizeKeys),
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

// 实际生效的密钥：images 网关按密钥绑定分组的 platform 硬校验，
// 必须用一个 group_id === 选中分组 的密钥，不能盲取第一个。
const activeKey = computed<ApiKey | null>(() => {
  if (form.value.groupId == null) return null
  return activeKeys.value.find((k) => k.group_id === form.value.groupId) ?? null
})

// 选中分组没有绑定密钥时的指引（区别于"一个密钥都没有"的 apiKeyError）
const keyError = computed(() => {
  if (apiKeyError.value) return apiKeyError.value
  if (form.value.groupId == null || imageCapableGroups.value.length === 0) return ''
  if (!activeKey.value) {
    const g = currentGroupEntry.value?.group
    return t('imageGen.noKeyForGroup', { group: g?.name ?? '' })
  }
  return ''
})

// 可在页内一键创建密钥：已选分组、该分组无可用密钥、且非硬错误（拉取/创建失败）
const canCreateKeyInline = computed(
  () =>
    form.value.groupId != null &&
    imageCapableGroups.value.length > 0 &&
    !activeKey.value &&
    !apiKeyError.value,
)

// 页内创建一个绑定当前分组的真实密钥（等价于「API 密钥」页的创建，省去跳页）
async function createKeyInline() {
  const gid = form.value.groupId
  if (gid == null || creatingKey.value) return
  creatingKey.value = true
  apiKeyError.value = ''
  try {
    const groupName = currentGroupEntry.value?.group.name ?? ''
    const name = `生图工坊 - ${groupName}`.slice(0, 64)
    const created = await keysAPI.create(name, gid)
    // 新建密钥绑定了当前分组，加入列表后 activeKey 立即解析、keyError 自动消除
    activeKeys.value = [created, ...activeKeys.value]
  } catch (err) {
    apiKeyError.value = extractApiErrorMessage(err, t('common.error'))
  } finally {
    creatingKey.value = false
  }
}

const hasResults = computed(() => results.value.length > 0)
const canSubmit = computed(() =>
  !submitting.value &&
  !!form.value.prompt.trim() &&
  !!activeKey.value &&
  !keyError.value &&
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

// ── 选分辨率时，当前比例若该档不支持（会导致计费错档）则重置为 auto ──
watch(() => form.value.resolution, (r) => {
  const sizeKeys = SIZE_MAP[r] ?? {}
  if (!(form.value.aspect in sizeKeys)) {
    form.value.aspect = 'auto'
  }
})


// 判断模型是否属于图片生成类。多路兜底，避免后端 billing_mode 未正确标记时漏过。
function isImageModel(m: UserSupportedModel): boolean {
  // 1) 后端明确标记为图片计费
  if (m.pricing?.billing_mode === 'image') return true
  // 2) 后端配了图片单价（说明 admin/LiteLLM 把它当图片模型）
  if (m.pricing?.image_output_price != null && m.pricing.image_output_price > 0) return true
  // 3) 前端能力表里有的已知图片模型
  if (m.name in MODEL_CAPS) return true
  // 4) 名称模式兜底（避免 admin 起了个变体名，比如 gpt-image-1.5-beta）
  const n = m.name.toLowerCase()
  if (n.startsWith('gpt-image') || n.startsWith('dall-e') || n.startsWith('dalle')) return true
  if (n.includes('flash-image') || n.includes('imagen')) return true
  return false
}

async function loadGroupsAndModels() {
  try {
    const channels = await userChannelsAPI.getAvailable()
    // 判定唯一信号：group.allow_image_generation === true（管理员显式声明本分组提供生图）。
    // 不再依赖 supported_models 筛分组——透传渠道它天生为空（模型配在账号层）。
    // 模型清单来源：渠道若显式列出了图片模型就用它（精确，适用于配了模型限制的渠道）；
    // 否则用该分组平台的官方图片模型目录（透传场景，目录即事实）。
    const groupMap = new Map<number, { group: UserAvailableGroup; models: Set<string> }>()
    for (const ch of channels) {
      for (const sec of ch.platforms) {
        const explicitImageModels = sec.supported_models.filter(isImageModel).map((m) => m.name)
        for (const g of sec.groups) {
          if (g.allow_image_generation !== true) continue
          if (!groupMap.has(g.id)) {
            groupMap.set(g.id, { group: g, models: new Set() })
          }
          const bucket = groupMap.get(g.id)!.models
          const models =
            explicitImageModels.length > 0
              ? explicitImageModels
              : IMAGE_MODEL_CATALOG[g.platform] ?? []
          for (const m of models) bucket.add(m)
        }
      }
    }
    // models 为空的分组丢弃（平台未知且渠道也没列出——genuinely 无可用图片模型）
    imageCapableGroups.value = Array.from(groupMap.values())
      .map((x) => ({ group: x.group, models: Array.from(x.models) }))
      .filter((x) => x.models.length > 0)
    // 默认选第一个 group 和它的第一个 model
    if (imageCapableGroups.value.length > 0) {
      form.value.groupId = imageCapableGroups.value[0].group.id
      form.value.model = imageCapableGroups.value[0].models[0] ?? ''
    }
  } catch {
    // 拉失败不阻塞，分组 rail 会显示空态
  }
}

// 拉全部 active 密钥（拉满 100 个足够覆盖个人用户）；按分组匹配在 activeKey computed 里
async function loadKey() {
  try {
    const res = await keysAPI.list(1, 100, { status: 'active' })
    const keys = res?.items ?? []
    if (keys.length === 0) {
      apiKeyError.value = t('imageGen.noKeyError')
      return
    }
    activeKeys.value = keys
  } catch (err) {
    apiKeyError.value = extractApiErrorMessage(err, t('common.error'))
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
  const files = Array.from(target.files ?? [])
  target.value = '' // 允许重复选同一文件
  if (files.length === 0) return
  referenceError.value = ''
  // 单张大小限制 4 MB
  if (files.some((f) => f.size > 4 * 1024 * 1024)) {
    referenceError.value = t('imageGen.referenceTooLarge')
    return
  }
  // 总数限制
  const remaining = MAX_REFERENCE_IMAGES - referenceFiles.value.length
  if (remaining <= 0) {
    referenceError.value = t('imageGen.referenceTooMany', { n: MAX_REFERENCE_IMAGES })
    return
  }
  const accepted = files.slice(0, remaining)
  if (files.length > remaining) {
    referenceError.value = t('imageGen.referenceTooMany', { n: MAX_REFERENCE_IMAGES })
  }
  for (const file of accepted) {
    referenceFiles.value.push(file)
    const reader = new FileReader()
    reader.onload = () => {
      referencePreviews.value.push({ name: file.name, url: String(reader.result ?? '') })
    }
    reader.readAsDataURL(file)
  }
}

function removeReference(idx: number) {
  referenceFiles.value.splice(idx, 1)
  referencePreviews.value.splice(idx, 1)
  if (referenceFiles.value.length === 0) referenceError.value = ''
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
  // canSubmit 已保证 activeKey 存在；这里仅作 TS 空值收窄
  if (!activeKey.value) return
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
    const useEdit = referenceFiles.value.length > 0 && (caps?.edit ?? false)
    // 始终解析出具体 size：当前比例若该档无映射，回落到该档方形（auto 项必有），
    // 确保下发的 size 一定能被后端正确归类到所选分辨率档，计费不错档
    const sizeStr =
      SIZE_MAP[form.value.resolution]?.[form.value.aspect] ??
      SIZE_MAP[form.value.resolution]?.['auto'] ??
      '1024x1024'
    const partialImages = useStream ? 2 : 0

    let response: Response
    if (useEdit) {
      // 以图生图：multipart/form-data 走 /v1/images/edits
      // 显式 response_format=b64_json 避免拿到的 URL 1 小时后过期（OpenAI 默认行为）
      const fd = new FormData()
      // 多张参考图：OpenAI images/edits 用 image[] 数组字段
      for (const f of referenceFiles.value) fd.append('image[]', f)
      fd.append('prompt', form.value.prompt.trim())
      fd.append('model', form.value.model)
      fd.append('n', String(form.value.n))
      fd.append('response_format', 'b64_json')
      fd.append('size', sizeStr)
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
      payload.size = sizeStr
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
/* 生成画布：统一深色基调（明确"工作中"语境，比浅灰更聚焦）*/
.generating-canvas {
  position: relative;
  display: flex;
  min-height: 22rem;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  border-radius: 0.75rem;
  padding: 2.5rem 1rem;
  background: linear-gradient(160deg, #0f172a 0%, #1e293b 60%, #0f172a 100%);
}

/* 点阵背景：深色画布上的细网格，缓慢呼吸 */
.dot-grid {
  position: absolute;
  inset: 0;
  background-image: radial-gradient(circle, rgba(148, 163, 184, 0.12) 1px, transparent 1px);
  background-size: 20px 20px;
  animation: dot-pulse 4s ease-in-out infinite;
}
@keyframes dot-pulse {
  0%, 100% { opacity: 0.5; }
  50% { opacity: 0.9; }
}

/* 扫描光带：自上而下匀速扫过预览区，传达"正在绘制" */
.scan-beam {
  position: absolute;
  inset: 0;
  pointer-events: none;
  background: linear-gradient(
    180deg,
    transparent 0%,
    rgba(251, 191, 36, 0.06) 42%,
    rgba(251, 191, 36, 0.22) 50%,
    rgba(251, 191, 36, 0.06) 58%,
    transparent 100%
  );
  background-size: 100% 220%;
  animation: scan-sweep 2.2s ease-in-out infinite;
}
@keyframes scan-sweep {
  0% { background-position: 0 -110%; }
  100% { background-position: 0 110%; }
}

/* 无预览时的笔刷脉冲 */
.brush-pulse {
  display: inline-flex;
  animation: brush-breathe 1.8s ease-in-out infinite;
}
@keyframes brush-breathe {
  0%, 100% { transform: scale(0.92); opacity: 0.55; }
  50% { transform: scale(1.08); opacity: 1; }
}

/* 终图揭示：完成瞬间淡入并轻微放大归位 */
.reveal-img {
  animation: reveal-in 0.45s cubic-bezier(0.22, 1, 0.36, 1) both;
}
@keyframes reveal-in {
  from { opacity: 0; transform: scale(0.97); }
  to { opacity: 1; transform: scale(1); }
}

@media (prefers-reduced-motion: reduce) {
  .dot-grid,
  .scan-beam,
  .brush-pulse,
  .reveal-img { animation: none; }
}
</style>
