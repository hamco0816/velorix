<template>
  <AppLayout>
    <div
      data-testid="profile-shell"
      class="space-y-5 pb-10"
    >
      <!-- Hero：sky 渐变标题区，作为个人中心入口 -->
      <header class="page-hero page-hero-sky">
        <div class="relative z-10 max-w-3xl">
          <span class="page-hero-tag page-hero-tag-sky">
            <Icon name="user" size="sm" />
            {{ t('profile.title') }}
          </span>
          <h1 class="mt-3 text-2xl font-semibold tracking-tight text-gray-950 dark:text-white md:text-[28px]">
            {{ t('profile.title') }}
          </h1>
          <p class="mt-2 max-w-2xl text-sm leading-6 text-gray-600 dark:text-dark-200">
            {{ t('profile.description') }}
          </p>
        </div>
      </header>

      <ProfileInfoCard
        :user="user"
        :linuxdo-enabled="linuxdoOAuthEnabled"
        :oidc-enabled="oidcOAuthEnabled"
        :oidc-provider-name="oidcOAuthProviderName"
        :wechat-enabled="wechatOAuthEnabled"
        :wechat-open-enabled="wechatOAuthOpenEnabled"
        :wechat-mp-enabled="wechatOAuthMPEnabled"
      />

      <section data-testid="profile-security-panel" class="space-y-4">
        <div class="flex flex-col gap-3 sm:flex-row sm:items-end sm:justify-between">
          <div>
            <div class="flex items-center gap-3">
              <div class="flex h-10 w-10 items-center justify-center rounded-2xl bg-emerald-500/10 text-emerald-600 dark:bg-emerald-500/10 dark:text-emerald-300">
                <Icon name="shield" size="md" />
              </div>
              <div>
                <h2 class="text-xl font-bold text-gray-900 dark:text-white">
                  {{ t('profile.securityTitle') }}
                </h2>
                <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
                  {{ t('profile.securityDescription') }}
                </p>
              </div>
            </div>
          </div>
          <div class="rounded-full bg-white px-3 py-1 text-xs font-medium text-gray-500 shadow-sm ring-1 ring-gray-200 dark:bg-dark-900 dark:text-gray-300 dark:ring-dark-700">
            Password · 2FA
          </div>
        </div>

        <div class="grid gap-6 xl:grid-cols-[minmax(0,1fr)_360px]">
          <ProfilePasswordForm embedded />
          <ProfileTotpCard />
        </div>
      </section>

      <div class="grid gap-6 xl:grid-cols-[minmax(0,1fr)_320px]">
        <ProfileBalanceNotifyCard
          v-if="user && balanceLowNotifyEnabled"
          :enabled="user.balance_notify_enabled ?? true"
          :threshold="user.balance_notify_threshold"
          :extra-emails="user.balance_notify_extra_emails ?? []"
          :system-default-threshold="systemDefaultThreshold"
          :user-email="user.email"
        />

        <div
          v-if="hasContactMethods"
          class="card overflow-hidden border border-primary-100 bg-white dark:border-primary-900/40 dark:bg-dark-900/70"
        >
          <div class="border-b border-primary-100 bg-primary-50/80 px-5 py-4 dark:border-primary-900/40 dark:bg-primary-950/20">
            <div class="flex items-center gap-3">
              <div class="rounded-xl bg-primary-100 p-2.5 text-primary-600 dark:bg-primary-900/50 dark:text-primary-300">
                <Icon name="chat" size="md" />
              </div>
              <h3 class="font-semibold text-primary-800 dark:text-primary-200">
                {{ t('common.contactSupport') }}
              </h3>
            </div>
          </div>
          <div class="px-5 py-4">
            <ContactMethodsDisplay
              :methods="contactMethods"
              :legacy-info="contactInfo"
            />
          </div>
        </div>
      </div>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { Icon } from '@/components/icons'
import AppLayout from '@/components/layout/AppLayout.vue'
import ProfileBalanceNotifyCard from '@/components/user/profile/ProfileBalanceNotifyCard.vue'
import ProfileInfoCard from '@/components/user/profile/ProfileInfoCard.vue'
import ProfilePasswordForm from '@/components/user/profile/ProfilePasswordForm.vue'
import ProfileTotpCard from '@/components/user/profile/ProfileTotpCard.vue'
import ContactMethodsDisplay from '@/components/common/ContactMethodsDisplay.vue'
import { isWeChatWebOAuthEnabled } from '@/api/auth'
import { useAppStore } from '@/stores/app'
import { useAuthStore } from '@/stores/auth'
import type { ContactMethod } from '@/types'

const { t } = useI18n()
const appStore = useAppStore()
const authStore = useAuthStore()
const user = computed(() => authStore.user)

const contactInfo = ref('')
const contactMethods = ref<ContactMethod[]>([])
const hasContactMethods = computed(() => contactMethods.value.length > 0 || !!contactInfo.value)
const balanceLowNotifyEnabled = ref(false)
const systemDefaultThreshold = ref(0)
const linuxdoOAuthEnabled = ref(false)
const wechatOAuthEnabled = ref(false)
const wechatOAuthOpenEnabled = ref<boolean | undefined>(undefined)
const wechatOAuthMPEnabled = ref<boolean | undefined>(undefined)
const oidcOAuthEnabled = ref(false)
const oidcOAuthProviderName = ref('OIDC')

onMounted(async () => {
  const profileRefresh = authStore.refreshUser().catch((error) => {
    console.error('Failed to refresh profile:', error)
  })

  const settingsLoad = appStore.fetchPublicSettings()
    .then((settings) => {
      if (!settings) {
        return
      }
      contactInfo.value = settings.contact_info || ''
      contactMethods.value = Array.isArray(settings.contact_methods) ? settings.contact_methods : []
      balanceLowNotifyEnabled.value = settings.balance_low_notify_enabled ?? false
      systemDefaultThreshold.value = settings.balance_low_notify_threshold ?? 0
      linuxdoOAuthEnabled.value = settings.linuxdo_oauth_enabled ?? false
      wechatOAuthEnabled.value = isWeChatWebOAuthEnabled(settings)
      wechatOAuthOpenEnabled.value = typeof settings.wechat_oauth_open_enabled === 'boolean'
        ? settings.wechat_oauth_open_enabled
        : undefined
      wechatOAuthMPEnabled.value = typeof settings.wechat_oauth_mp_enabled === 'boolean'
        ? settings.wechat_oauth_mp_enabled
        : undefined
      oidcOAuthEnabled.value = settings.oidc_oauth_enabled ?? false
      oidcOAuthProviderName.value = settings.oidc_oauth_provider_name || 'OIDC'
    })
    .catch((error) => {
      console.error('Failed to load settings:', error)
    })

  await Promise.all([profileRefresh, settingsLoad])
})
</script>
