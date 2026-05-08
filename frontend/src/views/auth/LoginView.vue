<template>
  <AuthLayout>
    <!-- 品牌区：logo + 合并标题居中直接展示，不再包渐变容器（避免登录页"卡片化"的视觉冗余） -->
    <template #brand>
      <div class="mb-10 flex flex-col items-center text-center">
        <div class="mb-5 h-16 w-16 overflow-hidden">
          <img
            :src="siteLogo || '/logo.png'"
            alt="Logo"
            class="h-full w-full object-contain"
          />
        </div>
        <h1 class="text-2xl font-bold tracking-tight text-gray-900 dark:text-white">
          {{ siteName }}<span class="auth-brand-dot-sky">·</span>{{ t('auth.signIn') }}
        </h1>
      </div>
    </template>

    <!-- 设置未加载占位 -->
    <div
      v-if="!settingsLoaded"
      class="flex h-64 items-center justify-center"
    >
      <div class="h-6 w-6 animate-spin rounded-full border-2 border-gray-900 border-t-transparent dark:border-white dark:border-t-transparent"></div>
    </div>

    <template v-else>
      <form @submit.prevent="handleLogin" class="space-y-5">
        <!-- 邮箱 -->
        <div>
          <label for="email" class="auth-input-label">{{ t('auth.emailLabel') }}</label>
          <input
            id="email"
            v-model="formData.email"
            type="email"
            required
            autofocus
            autocomplete="email"
            :disabled="isLoading"
            class="auth-input"
            :class="{ 'auth-input-error': errors.email }"
            :placeholder="t('auth.emailPlaceholder')"
          />
        </div>

        <!-- 密码（眼睛图标在输入框内部右侧居中） -->
        <div>
          <label for="password" class="auth-input-label">{{ t('auth.passwordLabel') }}</label>
          <div class="relative">
            <input
              id="password"
              v-model="formData.password"
              :type="showPassword ? 'text' : 'password'"
              required
              autocomplete="current-password"
              :disabled="isLoading"
              class="auth-input pr-10"
              :class="{ 'auth-input-error': errors.password }"
              :placeholder="t('auth.passwordPlaceholder')"
            />
            <button
              type="button"
              tabindex="-1"
              :aria-label="showPassword ? t('common.hide') : t('common.show')"
              @click="showPassword = !showPassword"
              class="absolute inset-y-0 right-0 flex items-center pr-3 text-gray-400 transition-colors hover:text-gray-700 dark:hover:text-dark-200"
            >
              <Icon v-if="showPassword" name="eyeOff" size="sm" />
              <Icon v-else name="eye" size="sm" />
            </button>
          </div>
        </div>

        <!-- 忘记密码（右对齐弱化，紧贴密码框） -->
        <div
          v-if="passwordResetEnabled && !backendModeEnabled"
          class="flex justify-end"
        >
          <router-link
            to="/forgot-password"
            class="text-xs text-gray-500 transition-colors hover:text-gray-900 dark:text-dark-400 dark:hover:text-white"
          >
            {{ t('auth.forgotPassword') }}
          </router-link>
        </div>

        <!-- Turnstile：占位 → ✓/widget → 失败 三段式，避免暴露 Cloudflare 灰色加载矩形 -->
        <div v-if="turnstileEnabled && turnstileSiteKey">
          <!-- 1) 加载中占位（默认 3 秒，遮蔽掉 Cloudflare 自带的灰色加载状态） -->
          <div
            v-if="turnstilePlaceholderVisible"
            class="flex items-center gap-2.5 rounded-md border border-gray-200 bg-gray-50/80 px-3.5 py-3 dark:border-dark-700 dark:bg-dark-800/30"
          >
            <svg class="h-4 w-4 animate-spin text-gray-400" fill="none" viewBox="0 0 24 24">
              <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
              <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
            </svg>
            <span class="text-xs text-gray-600 dark:text-dark-300">{{ t('auth.turnstileLoading') }}</span>
          </div>

          <!-- 2) 已验证：emerald 成功提示 -->
          <div
            v-else-if="turnstileToken"
            class="flex items-center gap-2.5 rounded-md border border-emerald-200 bg-emerald-50/70 px-3.5 py-3 dark:border-emerald-800/60 dark:bg-emerald-900/15"
          >
            <Icon name="checkCircle" size="sm" class="flex-shrink-0 text-emerald-500" />
            <span class="text-xs font-medium text-emerald-700 dark:text-emerald-300">
              {{ t('auth.turnstileVerified') }}
            </span>
          </div>

          <!-- 3) 加载失败：amber 提示 + 重试 -->
          <div
            v-else-if="turnstileLoadFailed"
            class="flex items-start gap-2 rounded-md border border-amber-200 bg-amber-50 px-3 py-2.5 dark:border-amber-800/60 dark:bg-amber-900/15"
          >
            <Icon name="exclamationTriangle" size="sm" class="mt-0.5 flex-shrink-0 text-amber-500" />
            <p class="text-xs leading-relaxed text-amber-800 dark:text-amber-200">
              {{ t('auth.turnstileLoadSlow') }}
              <button
                type="button"
                @click="retryTurnstile"
                class="font-medium underline hover:text-amber-900 dark:hover:text-amber-100"
              >
                {{ t('auth.turnstileRetry') }}
              </button>
            </p>
          </div>

          <!-- 4) Widget 容器：始终挂载让 iframe 后台加载；占位/成功/失败时用 v-show 隐藏 -->
          <TurnstileWidget
            v-show="!turnstilePlaceholderVisible && !turnstileToken && !turnstileLoadFailed"
            ref="turnstileRef"
            :site-key="turnstileSiteKey"
            @verify="onTurnstileVerify"
            @expire="onTurnstileExpire"
            @error="onTurnstileError"
          />
        </div>

        <!-- 主按钮：仅在加载中或 Turnstile 已渲染但未通过时禁用，避免后台开关与 site key 不一致时死锁 -->
        <button
          type="submit"
          :disabled="isLoading || (turnstileEnabled && !!turnstileSiteKey && !turnstileToken)"
          class="auth-primary-btn"
        >
          <svg
            v-if="isLoading"
            class="-ml-1 mr-2 h-4 w-4 animate-spin"
            fill="none"
            viewBox="0 0 24 24"
          >
            <circle
              class="opacity-25"
              cx="12"
              cy="12"
              r="10"
              stroke="currentColor"
              stroke-width="4"
            ></circle>
            <path
              class="opacity-75"
              fill="currentColor"
              d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
            ></path>
          </svg>
          {{ isLoading ? t('auth.signingIn') : t('auth.signIn') }}
        </button>
      </form>

      <!-- 第三方登录：仅在后端启用对应模式时显示，不显示则整段不渲染 -->
      <div
        v-if="!backendModeEnabled && (linuxdoOAuthEnabled || wechatOAuthEnabled || oidcOAuthEnabled)"
        class="mt-3 space-y-3"
      >
        <LinuxDoOAuthSection
          v-if="linuxdoOAuthEnabled"
          :disabled="isLoading"
          :show-divider="false"
          minimal
        />
        <WechatOAuthSection
          v-if="wechatOAuthEnabled"
          :disabled="isLoading"
          :show-divider="false"
          minimal
        />
        <OidcOAuthSection
          v-if="oidcOAuthEnabled"
          :disabled="isLoading"
          :provider-name="oidcOAuthProviderName"
          :show-divider="false"
          minimal
        />
      </div>
    </template>

    <!-- 底部：注册入口（后端模式下整段不渲染） -->
    <template v-if="!backendModeEnabled" #footer>
      <p v-if="!settingsLoaded" class="h-5"></p>
      <p v-else-if="registrationEnabled" class="text-gray-500 dark:text-dark-400">
        {{ t('auth.dontHaveAccount') }}
        <router-link to="/register" class="auth-link ml-1">
          {{ t('auth.signUp') }}
        </router-link>
      </p>
      <p v-else class="text-gray-400 dark:text-dark-500">
        {{ t('auth.registrationDisabled') }}
      </p>
    </template>
  </AuthLayout>

  <!-- 2FA Modal -->
  <TotpLoginModal
    v-if="show2FAModal"
    ref="totpModalRef"
    :temp-token="totpTempToken"
    :user-email-masked="totpUserEmailMasked"
    @verify="handle2FAVerify"
    @cancel="handle2FACancel"
  />
</template>

<script setup lang="ts">
import { computed, ref, reactive, onMounted, onUnmounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { AuthLayout } from '@/components/layout'
import LinuxDoOAuthSection from '@/components/auth/LinuxDoOAuthSection.vue'
import OidcOAuthSection from '@/components/auth/OidcOAuthSection.vue'
import WechatOAuthSection from '@/components/auth/WechatOAuthSection.vue'
import TotpLoginModal from '@/components/auth/TotpLoginModal.vue'
import Icon from '@/components/icons/Icon.vue'
import TurnstileWidget from '@/components/TurnstileWidget.vue'
import { useAuthStore, useAppStore } from '@/stores'
import { isTotp2FARequired, isWeChatWebOAuthEnabled } from '@/api/auth'
import type { PublicSettings, TotpLoginResponse } from '@/types'
import { clearAllAffiliateReferralCodes } from '@/utils/oauthAffiliate'
import { sanitizeUrl } from '@/utils/url'

const { t } = useI18n()

// ==================== Router & Stores ====================

const router = useRouter()
const authStore = useAuthStore()
const appStore = useAppStore()

// 站点品牌信息（用于品牌区标题与 logo），随后端配置自动同步
const siteName = computed(() => appStore.siteName || 'Sub2API')
const siteLogo = computed(() => sanitizeUrl(appStore.siteLogo || '', { allowRelative: true, allowDataUrl: true }))

// ==================== State ====================

const isLoading = ref<boolean>(false)
const errorMessage = ref<string>('')
const showPassword = ref<boolean>(false)
const settingsLoaded = ref<boolean>(false)

// 后端公开设置（决定渲染哪些 OAuth / 注册 / 验证码 / 找回密码 入口）
const registrationEnabled = ref<boolean>(false)
const turnstileEnabled = ref<boolean>(false)
const turnstileSiteKey = ref<string>('')
const linuxdoOAuthEnabled = ref<boolean>(false)
const wechatOAuthEnabled = ref<boolean>(false)
const backendModeEnabled = ref<boolean>(false)
const oidcOAuthEnabled = ref<boolean>(false)
const oidcOAuthProviderName = ref<string>('OIDC')
const passwordResetEnabled = ref<boolean>(false)

// Turnstile
const turnstileRef = ref<InstanceType<typeof TurnstileWidget> | null>(null)
const turnstileToken = ref<string>('')
// 三段式占位策略：
//   1. turnstilePlaceholderVisible: 默认 true，显示我们自己的 spinner 占位（替代 Cloudflare 的灰色矩形）
//   2. 3 秒后切换：若已 verify 则显示 ✓；否则显示真实 widget（这时已完成渲染，不会是灰色加载态）
//   3. 8 秒仍无 token 则判定加载失败，给重试入口
const turnstilePlaceholderVisible = ref<boolean>(true)
const turnstileLoadFailed = ref<boolean>(false)
let turnstileSwapTimer: ReturnType<typeof setTimeout> | null = null
let turnstileLoadTimer: ReturnType<typeof setTimeout> | null = null

function clearTurnstileTimers(): void {
  if (turnstileSwapTimer) {
    clearTimeout(turnstileSwapTimer)
    turnstileSwapTimer = null
  }
  if (turnstileLoadTimer) {
    clearTimeout(turnstileLoadTimer)
    turnstileLoadTimer = null
  }
}

function startTurnstileTimers(): void {
  clearTurnstileTimers()
  // 3 秒后撤掉占位（隐形通过的用户已经 verify 了；剩下的用户切换看到的是已渲染的 widget）
  turnstileSwapTimer = setTimeout(() => {
    if (!turnstileToken.value && !turnstileLoadFailed.value) {
      turnstilePlaceholderVisible.value = false
    }
  }, 3000)
  // 8 秒仍未 verify 视为加载失败
  turnstileLoadTimer = setTimeout(() => {
    if (!turnstileToken.value) {
      turnstileLoadFailed.value = true
      turnstilePlaceholderVisible.value = false
    }
  }, 8000)
}

// 用户主动重试：清掉失败标记 + reset widget + 重新走一遍占位 → swap → 失败的状态机
function retryTurnstile(): void {
  turnstileLoadFailed.value = false
  turnstileToken.value = ''
  turnstilePlaceholderVisible.value = true
  if (turnstileRef.value) {
    turnstileRef.value.reset()
  }
  startTurnstileTimers()
}

// 2FA state
const show2FAModal = ref<boolean>(false)
const totpTempToken = ref<string>('')
const totpUserEmailMasked = ref<string>('')
const totpModalRef = ref<InstanceType<typeof TotpLoginModal> | null>(null)

const formData = reactive({
  email: '',
  password: ''
})

const errors = reactive({
  email: '',
  password: '',
  turnstile: ''
})

// 校验失败统一通过 toast 提示，避免在表单里挤额外文字
const validationToastMessage = computed(
  () => errors.email || errors.password || errors.turnstile || ''
)

watch(validationToastMessage, (value, previousValue) => {
  if (value && value !== previousValue) {
    appStore.showError(value)
  }
})

// ==================== Lifecycle ====================

// 把后端返回的公开设置映射到本地 ref，登录页据此决定渲染哪些区块
function applyPublicSettings(settings: PublicSettings): void {
  turnstileEnabled.value = settings.turnstile_enabled
  turnstileSiteKey.value = settings.turnstile_site_key || ''
  linuxdoOAuthEnabled.value = settings.linuxdo_oauth_enabled
  wechatOAuthEnabled.value = isWeChatWebOAuthEnabled(settings)
  backendModeEnabled.value = settings.backend_mode_enabled
  oidcOAuthEnabled.value = settings.oidc_oauth_enabled
  oidcOAuthProviderName.value = settings.oidc_oauth_provider_name || 'OIDC'
  passwordResetEnabled.value = settings.password_reset_enabled
  registrationEnabled.value = settings.registration_enabled
}

if (appStore.cachedPublicSettings) {
  applyPublicSettings(appStore.cachedPublicSettings)
  settingsLoaded.value = true
}

// settings 加载完成后，若启用了 Turnstile 则启动占位/失败的状态机计时器
watch(settingsLoaded, (loaded) => {
  if (loaded && turnstileEnabled.value && turnstileSiteKey.value && !turnstileToken.value) {
    startTurnstileTimers()
  }
}, { immediate: true })

onMounted(async () => {
  // 会话过期跳转回登录时给一次性提示
  const expiredFlag = sessionStorage.getItem('auth_expired')
  if (expiredFlag) {
    sessionStorage.removeItem('auth_expired')
    const message = t('auth.reloginRequired')
    errorMessage.value = message
    appStore.showWarning(message)
  }

  try {
    const settings = await appStore.fetchPublicSettings()
    if (settings) {
      applyPublicSettings(settings)
    }
  } catch (error) {
    console.error('Failed to load public settings:', error)
  } finally {
    settingsLoaded.value = true
  }
})

onUnmounted(() => {
  clearTurnstileTimers()
})

// ==================== Turnstile Handlers ====================

function onTurnstileVerify(token: string): void {
  clearTurnstileTimers()
  turnstilePlaceholderVisible.value = false
  turnstileLoadFailed.value = false
  turnstileToken.value = token
  errors.turnstile = ''
}

function onTurnstileExpire(): void {
  turnstileToken.value = ''
  errors.turnstile = t('auth.turnstileExpired')
}

function onTurnstileError(): void {
  clearTurnstileTimers()
  turnstilePlaceholderVisible.value = false
  turnstileToken.value = ''
  turnstileLoadFailed.value = true
  errors.turnstile = t('auth.turnstileFailed')
}

// ==================== Validation ====================

// 校验邮箱、密码、Turnstile token，命中错误的字段会被打上红色下边框
function validateForm(): boolean {
  errors.email = ''
  errors.password = ''
  errors.turnstile = ''

  let isValid = true

  if (!formData.email.trim()) {
    errors.email = t('auth.emailRequired')
    isValid = false
  } else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(formData.email)) {
    errors.email = t('auth.invalidEmail')
    isValid = false
  }

  if (!formData.password) {
    errors.password = t('auth.passwordRequired')
    isValid = false
  } else if (formData.password.length < 6) {
    errors.password = t('auth.passwordMinLength')
    isValid = false
  }

  if (turnstileEnabled.value && !turnstileToken.value) {
    errors.turnstile = t('auth.completeVerification')
    isValid = false
  }

  return isValid
}

// ==================== Form Handlers ====================

// 提交登录：调用 store 完成认证，必要时弹出 2FA 弹窗，最后跳回 redirect 或 dashboard
async function handleLogin(): Promise<void> {
  errorMessage.value = ''

  if (!validateForm()) {
    return
  }

  isLoading.value = true

  try {
    const response = await authStore.login({
      email: formData.email,
      password: formData.password,
      turnstile_token: turnstileEnabled.value ? turnstileToken.value : undefined
    })

    if (isTotp2FARequired(response)) {
      const totpResponse = response as TotpLoginResponse
      totpTempToken.value = totpResponse.temp_token || ''
      totpUserEmailMasked.value = totpResponse.user_email_masked || ''
      show2FAModal.value = true
      isLoading.value = false
      return
    }

    clearAllAffiliateReferralCodes()
    appStore.showSuccess(t('auth.loginSuccess'))

    const redirectTo = (router.currentRoute.value.query.redirect as string) || '/dashboard'
    await router.push(redirectTo)
  } catch (error: unknown) {
    if (turnstileRef.value) {
      turnstileRef.value.reset()
      turnstileToken.value = ''
    }

    const err = error as { message?: string; response?: { data?: { detail?: string } } }

    if (err.response?.data?.detail) {
      errorMessage.value = err.response.data.detail
    } else if (err.message) {
      errorMessage.value = err.message
    } else {
      errorMessage.value = t('auth.loginFailed')
    }

    appStore.showError(errorMessage.value)
  } finally {
    isLoading.value = false
  }
}

// ==================== 2FA Handlers ====================

// 提交 TOTP 二次验证码：成功后清掉临时 token 并跳转
async function handle2FAVerify(code: string): Promise<void> {
  if (totpModalRef.value) {
    totpModalRef.value.setVerifying(true)
  }

  try {
    await authStore.login2FA(totpTempToken.value, code)

    show2FAModal.value = false
    clearAllAffiliateReferralCodes()
    appStore.showSuccess(t('auth.loginSuccess'))

    const redirectTo = (router.currentRoute.value.query.redirect as string) || '/dashboard'
    await router.push(redirectTo)
  } catch (error: unknown) {
    const err = error as { message?: string; response?: { data?: { message?: string } } }
    const message = err.response?.data?.message || err.message || t('profile.totp.loginFailed')

    if (totpModalRef.value) {
      totpModalRef.value.setError(message)
      totpModalRef.value.setVerifying(false)
    }
  }
}

function handle2FACancel(): void {
  show2FAModal.value = false
  totpTempToken.value = ''
  totpUserEmailMasked.value = ''
}
</script>
