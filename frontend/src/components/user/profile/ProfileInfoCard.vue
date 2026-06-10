<template>
  <div class="space-y-5">
    <section
      data-testid="profile-overview-hero"
      class="overflow-hidden rounded-2xl border border-gray-200/70 bg-white shadow-card dark:border-dark-700/60 dark:bg-dark-800/40"
    >
      <div class="px-6 py-6 md:px-7">
        <div class="flex flex-col gap-6 lg:flex-row lg:items-start">
          <div
            class="flex h-20 w-20 shrink-0 items-center justify-center overflow-hidden rounded-2xl bg-gradient-to-br from-primary-500 to-primary-600 text-2xl font-bold text-white shadow-[0_8px_24px_-12px_rgba(234,88,12,0.45)]"
          >
            <img
              v-if="avatarUrl"
              :src="avatarUrl"
              :alt="displayName"
              class="h-full w-full object-cover"
            >
            <span v-else>{{ avatarInitial }}</span>
          </div>

          <div class="min-w-0 flex-1 space-y-4">
            <div class="space-y-2.5">
              <div class="flex flex-wrap items-center gap-2">
                <h2 class="truncate text-xl font-semibold tracking-tight text-gray-900 dark:text-white">
                  {{ displayName }}
                </h2>
                <span
                  :class="[
                    'inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-2xs font-medium ring-1 ring-inset',
                    user?.role === 'admin'
                      ? 'bg-primary-50 text-primary-700 ring-primary-200/70 dark:bg-primary-500/15 dark:text-primary-300 dark:ring-primary-500/30'
                      : 'bg-gray-50 text-gray-600 ring-gray-200/70 dark:bg-dark-700/40 dark:text-dark-200 dark:ring-dark-600/60',
                  ]"
                >
                  <Icon v-if="user?.role === 'admin'" name="shield" size="xs" />
                  {{ user?.role === 'admin' ? t('profile.administrator') : t('profile.user') }}
                </span>
                <span
                  :class="[
                    'inline-flex items-center gap-1.5 rounded-full px-2 py-0.5 text-2xs font-medium ring-1 ring-inset',
                    user?.status === 'active'
                      ? 'bg-emerald-50 text-emerald-700 ring-emerald-200/70 dark:bg-emerald-500/15 dark:text-emerald-300 dark:ring-emerald-500/30'
                      : 'bg-rose-50 text-rose-700 ring-rose-200/70 dark:bg-rose-500/15 dark:text-rose-300 dark:ring-rose-500/30',
                  ]"
                >
                  <span
                    class="h-1.5 w-1.5 rounded-full"
                    :class="user?.status === 'active' ? 'bg-emerald-500 animate-pulse' : 'bg-rose-500'"
                  ></span>
                  {{ user?.status === 'active' ? t('common.active') : t('common.disabled') }}
                </span>
              </div>

              <div class="space-y-1.5">
                <p class="truncate text-sm text-gray-600 dark:text-gray-300">
                  {{ primaryEmailDisplay }}
                </p>
                <div
                  v-if="sourceHints.length"
                  class="flex flex-wrap gap-1.5 text-xs text-gray-500 dark:text-gray-400"
                >
                  <span
                    v-for="hint in sourceHints"
                    :key="hint.key"
                    class="inline-flex items-center gap-1 rounded-full bg-gray-50 px-2.5 py-0.5 ring-1 ring-inset ring-gray-200/70 dark:bg-dark-800/60 dark:ring-dark-700/60"
                  >
                    <Icon name="link" size="xs" class="text-gray-400" />
                    {{ hint.text }}
                  </span>
                </div>
              </div>
            </div>

            <div class="grid gap-3 sm:grid-cols-3">
              <div
                data-testid="profile-overview-metric-balance"
                class="rounded-xl border border-gray-200/70 bg-white px-4 py-3 dark:border-dark-700/60 dark:bg-dark-800/40"
              >
                <p class="text-xs font-medium text-gray-500 dark:text-dark-400">
                  {{ t('profile.accountBalance') }}
                </p>
                <p class="mt-1 text-lg font-semibold tabular-nums tracking-tight text-emerald-600 dark:text-emerald-400">
                  {{ formatCurrency(user?.balance || 0) }}
                </p>
              </div>
              <div
                data-testid="profile-overview-metric-concurrency"
                class="rounded-xl border border-gray-200/70 bg-white px-4 py-3 dark:border-dark-700/60 dark:bg-dark-800/40"
              >
                <p class="text-xs font-medium text-gray-500 dark:text-dark-400">
                  {{ t('profile.concurrencyLimit') }}
                </p>
                <p class="mt-1 text-lg font-semibold tabular-nums tracking-tight text-gray-900 dark:text-white">
                  {{ user?.concurrency || 0 }}
                </p>
              </div>
              <div
                data-testid="profile-overview-metric-member-since"
                class="rounded-xl border border-gray-200/70 bg-white px-4 py-3 dark:border-dark-700/60 dark:bg-dark-800/40"
              >
                <p class="text-xs font-medium text-gray-500 dark:text-dark-400">
                  {{ t('profile.memberSince') }}
                </p>
                <p class="mt-1 text-lg font-semibold tabular-nums tracking-tight text-gray-900 dark:text-white">
                  {{ memberSinceLabel }}
                </p>
              </div>
            </div>
          </div>
        </div>
      </div>
    </section>

    <div class="space-y-5">
      <div data-testid="profile-main-column" class="space-y-5">
        <section
          data-testid="profile-basics-panel"
          class="overflow-hidden rounded-2xl border border-gray-200/70 bg-white shadow-card dark:border-dark-700/60 dark:bg-dark-800/40"
        >
          <div class="border-b border-gray-100 bg-gray-50/60 px-5 py-4 dark:border-dark-700/60 dark:bg-dark-800/30">
            <h3 class="text-sm font-semibold tracking-tight text-gray-900 dark:text-white">
              {{ t('profile.basicsTitle') }}
            </h3>
            <p class="mt-0.5 text-xs text-gray-500 dark:text-gray-400">
              {{ t('profile.basicsDescription') }}
            </p>
          </div>
          <div class="grid gap-4 p-5 lg:grid-cols-[280px_minmax(0,1fr)]">
            <div class="rounded-xl border border-gray-200/70 bg-gray-50/40 p-5 dark:border-dark-700/60 dark:bg-dark-800/30">
              <ProfileAvatarCard
                :user="user"
                embedded
              />
            </div>

            <div class="rounded-xl border border-gray-200/70 bg-gray-50/40 p-5 dark:border-dark-700/60 dark:bg-dark-800/30">
              <ProfileEditForm
                :initial-username="user?.username || ''"
                embedded
              />
            </div>
          </div>
        </section>

        <section
          data-testid="profile-auth-bindings-panel"
          class="rounded-2xl border border-gray-200/70 bg-white p-5 shadow-card dark:border-dark-700/60 dark:bg-dark-800/40"
        >
          <ProfileIdentityBindingsSection
            :user="user"
            :linuxdo-enabled="linuxdoEnabled"
            :oidc-enabled="oidcEnabled"
            :oidc-provider-name="oidcProviderName"
            :wechat-enabled="wechatEnabled"
            :wechat-open-enabled="wechatOpenEnabled"
            :wechat-mp-enabled="wechatMpEnabled"
            embedded
            compact
          />
        </section>
      </div>

      <div data-testid="profile-side-column" class="space-y-5">
        <section
          v-if="sourceHints.length"
          class="overflow-hidden rounded-2xl border border-gray-200/70 bg-white shadow-card dark:border-dark-700/60 dark:bg-dark-800/40"
        >
          <div class="border-b border-gray-100 bg-gray-50/60 px-5 py-4 dark:border-dark-700/60 dark:bg-dark-800/30">
            <h3 class="text-sm font-semibold tracking-tight text-gray-900 dark:text-white">
              {{ t('profile.linkedProfileSources') }}
            </h3>
            <p class="mt-0.5 text-xs text-gray-500 dark:text-gray-400">
              {{ t('profile.linkedProfileSourcesDescription') }}
            </p>
          </div>
          <div class="grid gap-2 p-5">
            <div
              v-for="hint in sourceHints"
              :key="hint.key"
              class="flex items-start gap-3 rounded-xl border border-gray-200/70 bg-gray-50/40 px-4 py-3 text-sm text-gray-600 dark:border-dark-700/60 dark:bg-dark-800/30 dark:text-gray-300"
            >
              <Icon name="link" size="sm" class="mt-0.5 text-gray-400 dark:text-gray-500" />
              <span>{{ hint.text }}</span>
            </div>
          </div>
        </section>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import ProfileAvatarCard from '@/components/user/profile/ProfileAvatarCard.vue'
import ProfileEditForm from '@/components/user/profile/ProfileEditForm.vue'
import ProfileIdentityBindingsSection from '@/components/user/profile/ProfileIdentityBindingsSection.vue'
import type { User, UserAuthBindingStatus, UserAuthProvider, UserProfileSourceContext } from '@/types'

const props = withDefaults(defineProps<{
  user: User | null
  linuxdoEnabled?: boolean
  oidcEnabled?: boolean
  oidcProviderName?: string
  wechatEnabled?: boolean
  wechatOpenEnabled?: boolean
  wechatMpEnabled?: boolean
}>(), {
  linuxdoEnabled: false,
  oidcEnabled: false,
  oidcProviderName: 'OIDC',
  wechatEnabled: false,
  wechatOpenEnabled: undefined,
  wechatMpEnabled: undefined,
})

const { t } = useI18n()

function normalizeBindingStatus(binding: boolean | UserAuthBindingStatus | undefined): boolean | null {
  if (typeof binding === 'boolean') {
    return binding
  }
  if (!binding) {
    return null
  }
  if (typeof binding.bound === 'boolean') {
    return binding.bound
  }
  return Boolean(binding.provider_subject || binding.issuer || binding.provider_key)
}

function isEmailBound(user: User | null | undefined): boolean {
  if (typeof user?.email_bound === 'boolean') {
    return user.email_bound
  }

  const nested = user?.auth_bindings?.email ?? user?.identity_bindings?.email
  const normalized = normalizeBindingStatus(nested)
  return normalized ?? false
}

const avatarUrl = computed(() => props.user?.avatar_url?.trim() || '')
const displayName = computed(() => props.user?.username?.trim() || props.user?.email?.trim() || t('profile.user'))
const primaryEmailDisplay = computed(() => {
  const email = props.user?.email?.trim() || ''
  if (!email) {
    return ''
  }
  if (email.endsWith('.invalid') && !isEmailBound(props.user)) {
    return ''
  }
  return email
})
const avatarInitial = computed(() => displayName.value.charAt(0).toUpperCase() || 'U')
const memberSinceLabel = computed(() => {
  const raw = props.user?.created_at?.trim()
  if (!raw) {
    return '-'
  }

  const date = new Date(raw)
  if (Number.isNaN(date.getTime())) {
    return '-'
  }

  return new Intl.DateTimeFormat(undefined, {
    year: 'numeric',
    month: 'short',
  }).format(date)
})

const providerLabels = computed<Record<UserAuthProvider, string>>(() => ({
  email: t('profile.authBindings.providers.email'),
  linuxdo: t('profile.authBindings.providers.linuxdo'),
  oidc: t('profile.authBindings.providers.oidc', { providerName: props.oidcProviderName }),
  wechat: t('profile.authBindings.providers.wechat')
}))

function formatCurrency(value: number): string {
  return `$${value.toFixed(2)}`
}

function normalizeProvider(value: string): UserAuthProvider | null {
  const normalized = value.trim().toLowerCase()
  if (normalized === 'email' || normalized === 'linuxdo' || normalized === 'wechat') {
    return normalized
  }
  if (normalized === 'oidc' || normalized.startsWith('oidc:') || normalized.startsWith('oidc/')) {
    return 'oidc'
  }
  return null
}

function readObjectString(source: Record<string, unknown>, ...keys: string[]): string {
  for (const key of keys) {
    const value = source[key]
    if (typeof value === 'string' && value.trim()) {
      return value.trim()
    }
  }
  return ''
}

function resolveThirdPartySource(
  rawSource: string | UserProfileSourceContext | null | undefined
): { provider: UserAuthProvider; label: string } | null {
  if (!rawSource) {
    return null
  }

  if (typeof rawSource === 'string') {
    const provider = normalizeProvider(rawSource)
    if (!provider || provider === 'email') {
      return null
    }
    return {
      provider,
      label: providerLabels.value[provider]
    }
  }

  const sourceRecord = rawSource as Record<string, unknown>
  const provider = normalizeProvider(
    readObjectString(sourceRecord, 'provider', 'source', 'provider_type', 'auth_provider')
  )
  if (!provider || provider === 'email') {
    return null
  }

  const explicitLabel = readObjectString(
    sourceRecord,
    'provider_label',
    'label',
    'provider_name',
    'providerName'
  )

  return {
    provider,
    label: explicitLabel || providerLabels.value[provider]
  }
}

const sourceHints = computed(() => {
  const currentUser = props.user
  if (!currentUser) {
    return []
  }

  const hints: Array<{ key: string; text: string }> = []
  const avatarSource = resolveThirdPartySource(
    currentUser.profile_sources?.avatar ?? currentUser.avatar_source
  )
  const usernameSource = resolveThirdPartySource(
    currentUser.profile_sources?.username ??
      currentUser.profile_sources?.display_name ??
      currentUser.profile_sources?.nickname ??
      currentUser.display_name_source ??
      currentUser.username_source ??
      currentUser.nickname_source
  )

  if (avatarSource) {
    hints.push({
      key: 'avatar',
      text: t('profile.authBindings.source.avatar', { providerName: avatarSource.label })
    })
  }

  if (usernameSource) {
    hints.push({
      key: 'username',
      text: t('profile.authBindings.source.username', { providerName: usernameSource.label })
    })
  }

  return hints
})
</script>
