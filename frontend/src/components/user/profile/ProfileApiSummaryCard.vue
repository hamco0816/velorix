<!-- 个人资料页「API 概要」卡片：余额与密钥数量入口，点击跳转到充值 / 密钥管理页 -->
<template>
  <section
    data-testid="profile-api-summary"
    class="overflow-hidden rounded-2xl border border-gray-200/70 bg-white shadow-card dark:border-dark-700/60 dark:bg-dark-800/40"
  >
    <div class="flex items-center gap-3 border-b border-gray-100 bg-gray-50/60 px-5 py-4 dark:border-dark-700/60 dark:bg-dark-800/30">
      <div class="flex h-8 w-8 shrink-0 items-center justify-center rounded-xl bg-gray-50 text-gray-600 ring-1 ring-inset ring-gray-200/70 dark:bg-dark-800/60 dark:text-dark-200 dark:ring-dark-700/60">
        <Icon name="key" size="sm" />
      </div>
      <div class="min-w-0 flex-1">
        <h3 class="text-sm font-semibold text-gray-900 dark:text-white">
          {{ t('profile.apiSummary.title') }}
        </h3>
        <p class="mt-0.5 text-xs text-gray-500 dark:text-gray-400">
          {{ t('profile.apiSummary.description') }}
        </p>
      </div>
      <!-- 密钥数量加载失败时提供重试入口 -->
      <button
        v-if="keyCountFailed"
        type="button"
        class="btn btn-ghost btn-sm shrink-0"
        @click="loadKeyCount"
      >
        <Icon name="refresh" size="sm" class="mr-1.5" />
        {{ t('common.retry') }}
      </button>
    </div>

    <div class="grid gap-3 p-5 sm:grid-cols-2">
      <!-- 余额入口：点击前往充值页 -->
      <button
        type="button"
        data-testid="profile-api-summary-balance"
        class="group flex flex-col rounded-xl border border-gray-200/70 bg-gray-50/40 px-4 py-3.5 text-left transition-colors duration-150 hover:border-gray-300 hover:bg-gray-50 motion-reduce:transition-none dark:border-dark-700/60 dark:bg-dark-800/30 dark:hover:border-dark-600"
        @click="router.push('/purchase')"
      >
        <span class="text-xs font-medium text-gray-500 dark:text-dark-400">
          {{ t('profile.accountBalance') }}
        </span>
        <span class="mt-1 text-xl font-semibold tabular-nums tracking-tight text-gray-900 dark:text-white">
          {{ balanceDisplay }}
        </span>
        <span class="mt-2 inline-flex items-center gap-1 text-xs font-medium text-brand-600 dark:text-brand-400">
          {{ t('profile.apiSummary.topUpAction') }}
          <Icon name="arrowRight" size="xs" class="transition-transform duration-150 group-hover:translate-x-0.5 motion-reduce:transition-none" />
        </span>
      </button>

      <!-- 密钥数量入口：点击前往 API 密钥页 -->
      <button
        type="button"
        data-testid="profile-api-summary-keys"
        class="group flex flex-col rounded-xl border border-gray-200/70 bg-gray-50/40 px-4 py-3.5 text-left transition-colors duration-150 hover:border-gray-300 hover:bg-gray-50 motion-reduce:transition-none dark:border-dark-700/60 dark:bg-dark-800/30 dark:hover:border-dark-600"
        @click="router.push('/keys')"
      >
        <span class="text-xs font-medium text-gray-500 dark:text-dark-400">
          {{ t('keys.title') }}
        </span>
        <span class="mt-1 flex items-baseline text-xl font-semibold tabular-nums tracking-tight text-gray-900 dark:text-white">
          <span v-if="keyCountLoading" class="skeleton inline-block h-6 w-12" aria-hidden="true"></span>
          <span v-else-if="keyCountFailed">—</span>
          <span v-else>{{ keyCount }}</span>
        </span>
        <span class="mt-2 inline-flex items-center gap-1 text-xs font-medium text-brand-600 dark:text-brand-400">
          {{ t('profile.apiSummary.manageKeysAction') }}
          <Icon name="arrowRight" size="xs" class="transition-transform duration-150 group-hover:translate-x-0.5 motion-reduce:transition-none" />
        </span>
      </button>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import Icon from '@/components/icons/Icon.vue'
import { keysAPI } from '@/api'
import type { User } from '@/types'

const props = defineProps<{
  user: User | null
}>()

const { t } = useI18n()
const router = useRouter()

const balanceDisplay = computed(() => `$${(props.user?.balance ?? 0).toFixed(2)}`)

// 密钥总数：只取分页 total，不拉取完整列表
const keyCount = ref(0)
const keyCountLoading = ref(false)
const keyCountFailed = ref(false)

// 查询当前用户的 API 密钥总数
async function loadKeyCount() {
  keyCountLoading.value = true
  keyCountFailed.value = false
  try {
    const res = await keysAPI.list(1, 1)
    keyCount.value = res.total ?? 0
  } catch (error) {
    keyCountFailed.value = true
    console.error('Failed to load API key count:', error)
  } finally {
    keyCountLoading.value = false
  }
}

onMounted(loadKeyCount)
</script>
