<template>
  <AppLayout>
    <div
      data-testid="profile-shell"
      class="mx-auto max-w-[1180px] space-y-6"
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

      <div class="grid gap-6 xl:grid-cols-[minmax(0,1fr)_360px]">
        <div class="space-y-6">
          <section
            data-testid="profile-security-panel"
            class="card border border-gray-100 bg-white/95 dark:border-dark-700 dark:bg-dark-900/60"
          >
            <div class="border-b border-gray-100 px-6 py-4 dark:border-dark-700">
              <h2 class="text-lg font-semibold text-gray-900 dark:text-white">
                {{ t('profile.securityTitle') }}
              </h2>
              <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
                {{ t('profile.securityDescription') }}
              </p>
            </div>
            <div class="px-6 py-6">
              <ProfilePasswordForm embedded />
            </div>
          </section>

          <ProfileBalanceNotifyCard
            v-if="user && balanceLowNotifyEnabled"
            :enabled="user.balance_notify_enabled ?? true"
            :threshold="user.balance_notify_threshold"
            :extra-emails="user.balance_notify_extra_emails ?? []"
            :system-default-threshold="systemDefaultThreshold"
            :user-email="user.email"
          />
        </div>

        <aside class="space-y-6 xl:sticky xl:top-24 xl:self-start">
          <div
            v-if="contactInfo"
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
              <p class="break-words text-sm font-medium text-gray-700 dark:text-gray-200">
                {{ contactInfo }}
              </p>
            </div>
          </div>

          <ProfileTotpCard />
        </aside>
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
import { isWeChatWebOAuthEnabled } from '@/api/auth'
import { useAppStore } from '@/stores/app'
import { useAuthStore } from '@/stores/auth'

const { t } = useI18n()
const appStore = useAppStore()
const authStore = useAuthStore()
const user = computed(() => authStore.user)

const contactInfo = ref('')
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
