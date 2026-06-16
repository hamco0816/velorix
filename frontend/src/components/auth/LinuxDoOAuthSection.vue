<template>
  <div class="space-y-4">
    <button
      v-if="minimal"
      type="button"
      :disabled="disabled"
      class="oauth-btn-minimal"
      @click="startLogin"
    >
      <BrandIcon brand="linuxdo" size="18px" />
      {{ t('auth.linuxdo.signIn') }}
    </button>
    <button
      v-else
      type="button"
      :disabled="disabled"
      class="btn w-full border-brand-100 bg-brand-50/80 text-gray-900 shadow-sm shadow-brand-100/70 hover:border-brand-200 hover:bg-brand-100 dark:border-brand-900/40 dark:bg-brand-950/25 dark:text-brand-100 dark:shadow-none"
      @click="startLogin"
    >
      <span class="mr-2 inline-flex h-6 w-6 items-center justify-center rounded-full bg-white shadow-sm ring-1 ring-brand-100 dark:bg-gray-900 dark:ring-brand-900/50">
        <BrandIcon brand="linuxdo" size="18px" />
      </span>
      {{ t('auth.linuxdo.signIn') }}
    </button>

    <div v-if="showDivider" class="flex items-center gap-3">
      <div class="h-px flex-1 bg-gray-200 dark:bg-dark-700"></div>
      <span class="text-xs text-gray-500 dark:text-dark-400">
        {{ t('auth.oauthOrContinue') }}
      </span>
      <div class="h-px flex-1 bg-gray-200 dark:bg-dark-700"></div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import BrandIcon from '@/components/common/BrandIcon.vue'
import { resolveAffiliateReferralCode, storeOAuthAffiliateCode } from '@/utils/oauthAffiliate'

const props = withDefaults(defineProps<{
  disabled?: boolean
  affCode?: string
  showDivider?: boolean
  minimal?: boolean
}>(), {
  showDivider: true,
  minimal: false
})

const route = useRoute()
const { t } = useI18n()

function startLogin(): void {
  const redirectTo = (route.query.redirect as string) || '/dashboard'
  storeOAuthAffiliateCode(resolveAffiliateReferralCode(props.affCode, route.query.aff, route.query.aff_code))
  const apiBase = (import.meta.env.VITE_API_BASE_URL as string | undefined) || '/api/v1'
  const normalized = apiBase.replace(/\/$/, '')
  const startURL = `${normalized}/auth/oauth/linuxdo/start?redirect=${encodeURIComponent(redirectTo)}`
  window.location.href = startURL
}
</script>
