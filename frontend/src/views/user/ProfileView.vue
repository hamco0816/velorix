<template>
  <AppLayout wide>
    <div
      data-testid="profile-shell"
      class="space-y-5 pb-10"
    >
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
        <!-- Section header：克制的水平 divider 风，与 AppHeader 标题不重复 -->
        <div class="flex items-center gap-3">
          <div class="flex h-8 w-8 items-center justify-center rounded-xl bg-emerald-50 text-emerald-600 ring-1 ring-inset ring-emerald-200/70 dark:bg-emerald-500/15 dark:text-emerald-300 dark:ring-emerald-500/30">
            <Icon name="shield" size="sm" />
          </div>
          <div class="min-w-0 flex-1">
            <h2 class="text-sm font-semibold tracking-tight text-gray-900 dark:text-white">
              {{ t('profile.securityTitle') }}
            </h2>
            <p class="mt-0.5 text-xs text-gray-500 dark:text-gray-400">
              {{ t('profile.securityDescription') }}
            </p>
          </div>
          <div class="hidden sm:inline-flex items-center gap-1.5 rounded-full bg-gray-50 px-2.5 py-1 text-[11px] font-medium text-gray-500 ring-1 ring-inset ring-gray-200/70 dark:bg-dark-800/60 dark:text-dark-200 dark:ring-dark-700/60">
            <Icon name="lock" size="xs" class="text-gray-400" />
            Password · 2FA
          </div>
        </div>

        <div class="grid gap-5 lg:grid-cols-[minmax(0,1fr)_360px]">
          <ProfilePasswordForm embedded />
          <ProfileTotpCard />
        </div>
      </section>

      <div class="grid gap-5 lg:grid-cols-[minmax(0,1fr)_320px]">
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
          class="overflow-hidden rounded-2xl border border-gray-200/70 bg-white shadow-[0_1px_2px_rgba(15,23,42,0.04)] dark:border-dark-700/60 dark:bg-dark-800/40"
        >
          <div class="flex items-center gap-3 border-b border-gray-100 bg-gray-50/60 px-5 py-4 dark:border-dark-700/60 dark:bg-dark-800/30">
            <div class="flex h-8 w-8 items-center justify-center rounded-xl bg-primary-50 text-primary-600 ring-1 ring-inset ring-primary-200/70 dark:bg-primary-500/15 dark:text-primary-300 dark:ring-primary-500/30">
              <Icon name="chat" size="sm" />
            </div>
            <h3 class="text-sm font-semibold text-gray-900 dark:text-white">
              {{ t('common.contactSupport') }}
            </h3>
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
