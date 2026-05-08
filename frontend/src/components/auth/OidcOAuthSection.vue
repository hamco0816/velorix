<template>
  <div class="space-y-4">
    <button
      v-if="minimal"
      type="button"
      :disabled="disabled"
      class="oauth-btn-minimal"
      @click="startLogin"
    >
      <BrandIcon brand="oidc" size="18px" />
      {{ t('auth.oidc.signIn', { providerName: normalizedProviderName }) }}
    </button>
    <button
      v-else
      type="button"
      :disabled="disabled"
      class="btn w-full border-orange-100 bg-orange-50/80 text-orange-700 shadow-sm shadow-orange-100/70 hover:border-orange-200 hover:bg-orange-100 dark:border-orange-900/40 dark:bg-orange-950/25 dark:text-orange-300 dark:shadow-none"
      @click="startLogin"
    >
      <span class="mr-2 inline-flex h-6 w-6 items-center justify-center rounded-full bg-white shadow-sm ring-1 ring-orange-100 dark:bg-slate-900 dark:ring-orange-900/50">
        <BrandIcon brand="oidc" size="18px" />
      </span>
      {{ t('auth.oidc.signIn', { providerName: normalizedProviderName }) }}
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
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import BrandIcon from '@/components/common/BrandIcon.vue'
import { resolveAffiliateReferralCode, storeOAuthAffiliateCode } from '@/utils/oauthAffiliate'

const props = withDefaults(defineProps<{
  disabled?: boolean
  affCode?: string
  providerName?: string
  showDivider?: boolean
  minimal?: boolean
}>(), {
  providerName: 'OIDC',
  showDivider: true,
  minimal: false
})

const route = useRoute()
const { t } = useI18n()

const normalizedProviderName = computed(() => {
  const name = props.providerName?.trim()
  return name || 'OIDC'
})

function startLogin(): void {
  const redirectTo = (route.query.redirect as string) || '/dashboard'
  storeOAuthAffiliateCode(resolveAffiliateReferralCode(props.affCode, route.query.aff, route.query.aff_code))
  const apiBase = (import.meta.env.VITE_API_BASE_URL as string | undefined) || '/api/v1'
  const normalized = apiBase.replace(/\/$/, '')
  const startURL = `${normalized}/auth/oauth/oidc/start?redirect=${encodeURIComponent(redirectTo)}`
  window.location.href = startURL
}
</script>
