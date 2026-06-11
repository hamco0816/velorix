<template>
  <Teleport to="body">
    <Transition name="cmdk">
      <div
        v-if="isOpen"
        class="cmdk-overlay"
        role="presentation"
        @mousedown.self="close"
      >
        <div class="cmdk-panel" role="dialog" aria-modal="true" :aria-label="t('commandPalette.trigger')">
          <!-- 搜索输入行 -->
          <div class="cmdk-input-row">
            <Icon name="search" size="sm" class="shrink-0 text-gray-400 dark:text-dark-400" />
            <input
              ref="inputRef"
              v-model="query"
              type="text"
              class="cmdk-input"
              :placeholder="t('commandPalette.placeholder')"
              role="combobox"
              aria-expanded="true"
              aria-controls="cmdk-list"
              :aria-activedescendant="activeItemId ? `cmdk-item-${activeItemId}` : undefined"
              @keydown="onKeydown"
            />
            <kbd class="kbd">Esc</kbd>
          </div>

          <!-- 结果列表 -->
          <div id="cmdk-list" ref="listRef" class="cmdk-list" role="listbox">
            <template v-for="group in visibleGroups" :key="group.key">
              <p class="cmdk-group-title">{{ group.label }}</p>
              <button
                v-for="item in group.items"
                :id="`cmdk-item-${item.id}`"
                :key="`${group.key}-${item.id}`"
                type="button"
                class="cmdk-item"
                :class="{ 'cmdk-item-active': item.id === activeItemId }"
                role="option"
                :aria-selected="item.id === activeItemId"
                @mousemove="activeItemId = item.id"
                @click="run(item)"
              >
                <Icon :name="item.icon" size="sm" class="cmdk-item-icon" />
                <span class="min-w-0 flex-1 truncate text-left">{{ item.label }}</span>
                <Icon
                  v-if="item.id === activeItemId"
                  name="arrowRight"
                  size="xs"
                  class="shrink-0 opacity-60"
                />
              </button>
            </template>

            <!-- 空结果 -->
            <div v-if="flatItems.length === 0" class="cmdk-empty">
              <Icon name="search" size="md" class="text-gray-300 dark:text-dark-500" />
              <p>{{ t('commandPalette.empty', { query: query.trim() }) }}</p>
            </div>
          </div>

          <!-- 底部键位提示 -->
          <div class="cmdk-footer">
            <span class="inline-flex items-center gap-1.5">
              <kbd class="kbd">↑</kbd><kbd class="kbd">↓</kbd>
              {{ t('commandPalette.hints.navigate') }}
            </span>
            <span class="inline-flex items-center gap-1.5">
              <kbd class="kbd">↵</kbd>
              {{ t('commandPalette.hints.open') }}
            </span>
            <span class="inline-flex items-center gap-1.5">
              <kbd class="kbd">Esc</kbd>
              {{ t('commandPalette.hints.close') }}
            </span>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
// 全局命令面板：Ctrl/Cmd+K 唤起，模糊搜索页面导航与快捷操作，纯键盘可完成全部交互
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import { useCommandPalette } from '@/composables/useCommandPalette'
import {
  useCommandPaletteItems,
  type PaletteGroup,
  type PaletteItem
} from '@/composables/useCommandPaletteItems'

const RECENT_STORAGE_KEY = 'sub2api_cmdk_recent'
const RECENT_MAX = 5

const { t } = useI18n()
const { isOpen, close, toggle } = useCommandPalette()
const { groups } = useCommandPaletteItems()

const query = ref('')
const activeItemId = ref<string | null>(null)
const inputRef = ref<HTMLInputElement | null>(null)
const listRef = ref<HTMLElement | null>(null)
const recentIds = ref<string[]>([])

// ============ 过滤：空查询时展示「最近使用 + 全部分组」，有查询时按词条匹配 ============
const visibleGroups = computed<PaletteGroup[]>(() => {
  const q = query.value.trim().toLowerCase()
  if (!q) {
    const byId = new Map(groups.value.flatMap((g) => g.items).map((i) => [i.id, i]))
    const recentItems = recentIds.value
      .map((id) => byId.get(id))
      .filter((i): i is PaletteItem => !!i)
    const base = groups.value.filter((g) => g.items.length > 0)
    if (recentItems.length === 0) return base
    return [
      { key: 'navigation' as const, label: t('commandPalette.groups.recent'), items: recentItems },
      ...base
    ]
  }
  // 多词查询：每个词都要命中 label 或 keywords
  const terms = q.split(/\s+/).filter(Boolean)
  return groups.value
    .map((g) => ({
      ...g,
      items: g.items.filter((item) => {
        const haystack = `${item.label} ${item.keywords}`.toLowerCase()
        return terms.every((term) => haystack.includes(term))
      })
    }))
    .filter((g) => g.items.length > 0)
})

const flatItems = computed<PaletteItem[]>(() => visibleGroups.value.flatMap((g) => g.items))

// 查询或列表变化时，把活动项重置到第一条
watch([query, flatItems], () => {
  activeItemId.value = flatItems.value[0]?.id ?? null
})

// ============ 打开/关闭副作用：聚焦、清空、锁滚动 ============
watch(isOpen, async (open) => {
  document.body.classList.toggle('modal-open', open)
  if (open) {
    query.value = ''
    loadRecent()
    activeItemId.value = flatItems.value[0]?.id ?? null
    await nextTick()
    inputRef.value?.focus()
  }
})

function loadRecent(): void {
  try {
    const raw = localStorage.getItem(RECENT_STORAGE_KEY)
    const parsed = raw ? JSON.parse(raw) : []
    recentIds.value = Array.isArray(parsed) ? parsed.filter((v) => typeof v === 'string') : []
  } catch {
    recentIds.value = []
  }
}

function saveRecent(id: string): void {
  const next = [id, ...recentIds.value.filter((v) => v !== id)].slice(0, RECENT_MAX)
  recentIds.value = next
  try {
    localStorage.setItem(RECENT_STORAGE_KEY, JSON.stringify(next))
  } catch {
    // localStorage 不可用（隐私模式等）时静默跳过，不影响功能
  }
}

function run(item: PaletteItem): void {
  saveRecent(item.id)
  close()
  item.perform()
}

// ============ 键盘交互 ============
function moveActive(delta: number): void {
  const items = flatItems.value
  if (items.length === 0) return
  const index = items.findIndex((i) => i.id === activeItemId.value)
  const next = (index + delta + items.length) % items.length
  activeItemId.value = items[next].id
  nextTick(() => {
    listRef.value
      ?.querySelector(`#cmdk-item-${CSS.escape(items[next].id)}`)
      ?.scrollIntoView({ block: 'nearest' })
  })
}

function onKeydown(e: KeyboardEvent): void {
  if (e.key === 'ArrowDown') {
    e.preventDefault()
    moveActive(1)
  } else if (e.key === 'ArrowUp') {
    e.preventDefault()
    moveActive(-1)
  } else if (e.key === 'Enter') {
    e.preventDefault()
    const item = flatItems.value.find((i) => i.id === activeItemId.value)
    if (item) run(item)
  } else if (e.key === 'Tab') {
    // 面板内唯一可聚焦控件是输入框，Tab 不外逃
    e.preventDefault()
  }
}

// 全局快捷键：Ctrl/Cmd+K 开关，Esc 关闭
function onGlobalKeydown(e: KeyboardEvent): void {
  if ((e.ctrlKey || e.metaKey) && e.key.toLowerCase() === 'k') {
    e.preventDefault()
    toggle()
    return
  }
  if (e.key === 'Escape' && isOpen.value) {
    e.preventDefault()
    close()
  }
}

onMounted(() => {
  window.addEventListener('keydown', onGlobalKeydown)
})

onBeforeUnmount(() => {
  window.removeEventListener('keydown', onGlobalKeydown)
  document.body.classList.remove('modal-open')
})
</script>

<style scoped>
/* 浮层：顶部留 14vh，让面板出现在视线黄金位 */
.cmdk-overlay {
  @apply fixed inset-0 z-[70] flex items-start justify-center px-4 pb-8 pt-[14vh];
  background: rgb(9 9 11 / 0.4);
  backdrop-filter: blur(2px);
}

.cmdk-panel {
  @apply flex w-full max-w-xl flex-col overflow-hidden rounded-2xl border border-gray-200 bg-white;
  @apply dark:border-dark-700 dark:bg-dark-900;
  box-shadow:
    0 1px 2px rgb(15 23 42 / 0.04),
    0 14px 36px -10px rgb(15 23 42 / 0.2),
    0 24px 70px -24px rgb(15 23 42 / 0.35);
}

.cmdk-input-row {
  @apply flex items-center gap-3 border-b border-gray-100 px-4 dark:border-dark-700;
}

.cmdk-input {
  @apply h-14 min-w-0 flex-1 bg-transparent text-sm text-gray-900 outline-none;
  @apply placeholder:text-gray-500 dark:text-white dark:placeholder:text-dark-400;
}

.cmdk-list {
  @apply max-h-[min(420px,55vh)] overflow-y-auto p-2;
}

.cmdk-group-title {
  @apply px-3 pb-1 pt-3 text-2xs font-medium text-gray-500 first:pt-1.5 dark:text-dark-400;
}

.cmdk-item {
  @apply flex w-full items-center gap-3 rounded-lg px-3 py-2.5 text-sm text-gray-700 dark:text-gray-200;
}

.cmdk-item-icon {
  @apply shrink-0 text-gray-400 dark:text-dark-400;
}

/* 活动态：黑底反白，与侧边栏激活态同一视觉语言 */
.cmdk-item-active {
  @apply bg-gray-900 text-white dark:bg-white dark:text-gray-900;
}

.cmdk-item-active .cmdk-item-icon {
  @apply text-white/80 dark:text-gray-900/70;
}

.cmdk-empty {
  @apply flex flex-col items-center gap-2 px-4 py-10 text-center text-sm text-gray-500 dark:text-dark-400;
}

.cmdk-footer {
  @apply flex items-center gap-4 border-t border-gray-100 px-4 py-2.5 text-2xs text-gray-500;
  @apply dark:border-dark-700 dark:text-dark-400;
}

/* 进出场：背景淡入 + 面板轻微缩放上浮 */
.cmdk-enter-active {
  transition: opacity 0.15s ease-out;
}
.cmdk-leave-active {
  transition: opacity 0.12s ease-in;
}
.cmdk-enter-from,
.cmdk-leave-to {
  opacity: 0;
}
.cmdk-enter-active .cmdk-panel {
  transition:
    transform 0.18s cubic-bezier(0.22, 1, 0.36, 1),
    opacity 0.18s ease-out;
}
.cmdk-enter-from .cmdk-panel {
  transform: scale(0.98) translateY(-6px);
  opacity: 0;
}

@media (prefers-reduced-motion: reduce) {
  .cmdk-enter-active,
  .cmdk-leave-active,
  .cmdk-enter-active .cmdk-panel {
    transition-duration: 1ms;
  }
  .cmdk-enter-from .cmdk-panel {
    transform: none;
  }
}
</style>
