<template>
  <!-- 分屏布局：左品牌叙事面板（lg+）+ 右表单列，窄屏单列 -->
  <AuthSplitShell
    :title="t('auth.createAccount')"
    :subtitle="siteName ? t('auth.signUpToStart', { siteName }) : ''"
  >
    <!-- 设置未加载占位 -->
    <div
      v-if="!settingsLoaded"
      class="flex h-64 items-center justify-center"
    >
      <div class="h-6 w-6 animate-spin rounded-full border-2 border-gray-900 border-t-transparent dark:border-white dark:border-t-transparent"></div>
    </div>

    <!-- 注册关闭：极简提示，引导回登录 -->
    <div
      v-else-if="!registrationEnabled"
      class="rounded-md border border-gray-200 px-5 py-6 text-center dark:border-dark-700"
    >
      <p class="text-sm text-gray-700 dark:text-dark-200">
        {{ t('auth.registrationDisabled') }}
      </p>
      <router-link
        to="/login"
        class="mt-3 inline-block text-sm font-medium text-gray-900 transition-colors hover:underline dark:text-white"
      >
        {{ t('auth.signIn') }} →
      </router-link>
    </div>

    <template v-else>
      <form @submit.prevent="handleRegister" class="space-y-5">
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

        <!-- 密码 -->
        <div>
          <label for="password" class="auth-input-label">{{ t('auth.passwordLabel') }}</label>
          <div class="relative">
            <input
              id="password"
              v-model="formData.password"
              :type="showPassword ? 'text' : 'password'"
              required
              autocomplete="new-password"
              :disabled="isLoading"
              class="auth-input pr-10"
              :class="{ 'auth-input-error': errors.password }"
              :placeholder="t('auth.createPasswordPlaceholder')"
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
          <p class="auth-input-hint">
            {{ t('auth.passwordHint') }}
          </p>
        </div>

        <!-- 邀请码：异步校验，边框颜色随状态变化 -->
        <div v-if="invitationCodeEnabled">
          <label for="invitation_code" class="auth-input-label">
            {{ t('auth.invitationCodeLabel') }}
          </label>
          <div class="relative">
            <input
              id="invitation_code"
              v-model="formData.invitation_code"
              type="text"
              :disabled="isLoading"
              class="auth-input pr-10"
              :class="{
                'auth-input-valid': invitationValidation.valid,
                'auth-input-error': invitationValidation.invalid || errors.invitation_code
              }"
              :placeholder="t('auth.invitationCodePlaceholder')"
              @input="handleInvitationCodeInput"
            />
            <!-- 校验状态指示：spin / 对号 / 叉号（输入框内部右侧居中） -->
            <div class="pointer-events-none absolute inset-y-0 right-0 flex items-center pr-3">
              <LoadingSpinner v-if="invitationValidating" size="sm" color="current" class="text-gray-400" />
              <Icon v-else-if="invitationValidation.valid" name="checkCircle" size="sm" class="text-emerald-500" />
              <Icon
                v-else-if="invitationValidation.invalid || errors.invitation_code"
                name="exclamationCircle"
                size="sm"
                class="text-red-500"
              />
            </div>
          </div>
        </div>

        <!-- 优惠码（可选）：同邀请码的状态指示模式 -->
        <div v-if="promoCodeEnabled">
          <label for="promo_code" class="auth-input-label">
            {{ t('auth.promoCodeLabel') }}
            <span class="ml-1 text-gray-400 dark:text-dark-500">({{ t('common.optional') }})</span>
          </label>
          <div class="relative">
            <input
              id="promo_code"
              v-model="formData.promo_code"
              type="text"
              :disabled="isLoading"
              class="auth-input pr-10"
              :class="{
                'auth-input-valid': promoValidation.valid,
                'auth-input-error': promoValidation.invalid
              }"
              :placeholder="t('auth.promoCodePlaceholder')"
              @input="handlePromoCodeInput"
            />
            <div class="pointer-events-none absolute inset-y-0 right-0 flex items-center pr-3">
              <LoadingSpinner v-if="promoValidating" size="sm" color="current" class="text-gray-400" />
              <Icon v-else-if="promoValidation.valid" name="checkCircle" size="sm" class="text-emerald-500" />
              <Icon v-else-if="promoValidation.invalid" name="exclamationCircle" size="sm" class="text-red-500" />
            </div>
          </div>
          <!-- 优惠码生效时的轻量提示（金额信息） -->
          <p
            v-if="promoValidation.valid && promoValidation.bonusAmount"
            class="mt-1.5 text-xs text-emerald-600 dark:text-emerald-400"
          >
            {{ t('auth.promoCodeValid', { amount: promoValidation.bonusAmount.toFixed(2) }) }}
          </p>
        </div>

        <!-- Turnstile：占位 → 已校验/widget → 失败 三段式，避免暴露 Cloudflare 灰色加载矩形 -->
        <div v-if="turnstileEnabled && turnstileSiteKey">
          <div
            v-if="turnstilePlaceholderVisible"
            class="flex items-center gap-2.5 rounded-md border border-gray-200 bg-gray-50/80 px-3.5 py-3 dark:border-dark-700 dark:bg-dark-800/30"
          >
            <LoadingSpinner size="sm" color="current" class="text-gray-400" />
            <span class="text-xs text-gray-600 dark:text-dark-300">{{ t('auth.turnstileLoading') }}</span>
          </div>

          <div
            v-else-if="turnstileToken"
            class="flex items-center gap-2.5 rounded-md border border-emerald-200 bg-emerald-50/70 px-3.5 py-3 dark:border-emerald-800/60 dark:bg-emerald-900/15"
          >
            <Icon name="checkCircle" size="sm" class="flex-shrink-0 text-emerald-500" />
            <span class="text-xs font-medium text-emerald-700 dark:text-emerald-300">
              {{ t('auth.turnstileVerified') }}
            </span>
          </div>

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

          <TurnstileWidget
            v-show="!turnstilePlaceholderVisible && !turnstileToken && !turnstileLoadFailed"
            ref="turnstileRef"
            :site-key="turnstileSiteKey"
            @verify="onTurnstileVerify"
            @expire="onTurnstileExpire"
            @error="onTurnstileError"
          />
        </div>

        <!-- 主按钮：根据是否启用邮箱验证切换"继续"/"创建账户"；同登录处理 disabled 死锁 -->
        <button
          type="submit"
          :disabled="isLoading || (turnstileEnabled && !!turnstileSiteKey && !turnstileToken)"
          class="auth-primary-btn"
        >
          <LoadingSpinner v-if="isLoading" size="sm" color="current" class="-ml-1 mr-2" />
          {{
            isLoading
              ? t('auth.processing')
              : emailVerifyEnabled
                ? t('auth.continue')
                : t('auth.createAccount')
          }}
        </button>
      </form>

      <!-- 第三方注册：仅当后端启用对应模式时渲染 -->
      <div
        v-if="linuxdoOAuthEnabled || wechatOAuthEnabled || oidcOAuthEnabled"
        class="mt-3 space-y-3"
      >
        <LinuxDoOAuthSection
          v-if="linuxdoOAuthEnabled"
          :disabled="isLoading"
          :aff-code="formData.aff_code"
          :show-divider="false"
          minimal
        />
        <WechatOAuthSection
          v-if="wechatOAuthEnabled"
          :disabled="isLoading"
          :aff-code="formData.aff_code"
          :show-divider="false"
          minimal
        />
        <OidcOAuthSection
          v-if="oidcOAuthEnabled"
          :disabled="isLoading"
          :provider-name="oidcOAuthProviderName"
          :aff-code="formData.aff_code"
          :show-divider="false"
          minimal
        />
      </div>
    </template>

    <!-- 底部：已有账户跳转 -->
    <template #footer>
      <p class="text-gray-500 dark:text-dark-400">
        {{ t('auth.alreadyHaveAccount') }}
        <router-link to="/login" class="auth-link ml-1">
          {{ t('auth.signIn') }}
        </router-link>
      </p>
    </template>
  </AuthSplitShell>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, onUnmounted, watch, computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import AuthSplitShell from '@/components/auth/AuthSplitShell.vue'
import LinuxDoOAuthSection from '@/components/auth/LinuxDoOAuthSection.vue'
import OidcOAuthSection from '@/components/auth/OidcOAuthSection.vue'
import WechatOAuthSection from '@/components/auth/WechatOAuthSection.vue'
import Icon from '@/components/icons/Icon.vue'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import TurnstileWidget from '@/components/TurnstileWidget.vue'
import { useAuthStore, useAppStore } from '@/stores'
import {
  getPublicSettings,
  isWeChatWebOAuthEnabled,
  validatePromoCode,
  validateInvitationCode
} from '@/api/auth'
import { buildAuthErrorMessage } from '@/utils/authError'
import {
  isRegistrationEmailSuffixAllowed,
  normalizeRegistrationEmailSuffixWhitelist
} from '@/utils/registrationEmailPolicy'
import {
  clearAffiliateReferralCode,
  loadAffiliateReferralCode,
  resolveAffiliateReferralCode
} from '@/utils/oauthAffiliate'

const { t, locale } = useI18n()

// ==================== Router & Stores ====================

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()
const appStore = useAppStore()

// ==================== State ====================

const isLoading = ref<boolean>(false)
const settingsLoaded = ref<boolean>(false)
const errorMessage = ref<string>('')
const showPassword = ref<boolean>(false)

// 后端公开设置（决定渲染哪些区块及业务策略）
const registrationEnabled = ref<boolean>(false)
const emailVerifyEnabled = ref<boolean>(false)
const promoCodeEnabled = ref<boolean>(true)
const invitationCodeEnabled = ref<boolean>(false)
const turnstileEnabled = ref<boolean>(false)
const turnstileSiteKey = ref<string>('')
// settings 未加载完成时为空，标题区只显示动作（避免首屏闪默认名 'Sub2API'）
const siteName = ref<string>('')
const linuxdoOAuthEnabled = ref<boolean>(false)
const wechatOAuthEnabled = ref<boolean>(false)
const oidcOAuthEnabled = ref<boolean>(false)
const oidcOAuthProviderName = ref<string>('OIDC')
const registrationEmailSuffixWhitelist = ref<string[]>([])

// Turnstile：与 LoginView 一致的三段式占位策略，避免暴露 Cloudflare 灰色加载矩形
const turnstileRef = ref<InstanceType<typeof TurnstileWidget> | null>(null)
const turnstileToken = ref<string>('')
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
  // 3 秒后撤掉占位（隐形通过的用户已 verify；剩下用户切到的是已渲染的 widget）
  turnstileSwapTimer = setTimeout(() => {
    if (!turnstileToken.value && !turnstileLoadFailed.value) {
      turnstilePlaceholderVisible.value = false
    }
  }, 3000)
  // 8 秒仍无 token → 视为加载失败
  turnstileLoadTimer = setTimeout(() => {
    if (!turnstileToken.value) {
      turnstileLoadFailed.value = true
      turnstilePlaceholderVisible.value = false
    }
  }, 8000)
}

function retryTurnstile(): void {
  turnstileLoadFailed.value = false
  turnstileToken.value = ''
  turnstilePlaceholderVisible.value = true
  if (turnstileRef.value) {
    turnstileRef.value.reset()
  }
  startTurnstileTimers()
}

// 优惠码异步校验状态
const promoValidating = ref<boolean>(false)
const promoValidation = reactive({
  valid: false,
  invalid: false,
  bonusAmount: null as number | null,
  message: ''
})
let promoValidateTimeout: ReturnType<typeof setTimeout> | null = null

// 邀请码异步校验状态
const invitationValidating = ref<boolean>(false)
const invitationValidation = reactive({
  valid: false,
  invalid: false,
  message: ''
})
let invitationValidateTimeout: ReturnType<typeof setTimeout> | null = null

const formData = reactive({
  email: '',
  password: '',
  promo_code: '',
  invitation_code: '',
  aff_code: ''
})

const errors = reactive({
  email: '',
  password: '',
  turnstile: '',
  invitation_code: ''
})

// 校验失败统一通过 toast 弹出，避免在表单内挤额外文字
const validationToastMessage = computed(() =>
  errors.email ||
  errors.password ||
  (invitationValidation.invalid ? invitationValidation.message : '') ||
  errors.invitation_code ||
  (promoValidation.invalid ? promoValidation.message : '') ||
  errors.turnstile ||
  ''
)

watch(validationToastMessage, (value, previousValue) => {
  if (value && value !== previousValue) {
    appStore.showError(value)
  }
})

// 从 URL 同步 affiliate 推荐码到表单（之后会在注册请求中带上）
function syncAffiliateReferralCode(): string {
  const code = resolveAffiliateReferralCode(route.query.aff, route.query.aff_code)
  if (code) {
    formData.aff_code = code
  }
  return code
}

// ==================== Lifecycle ====================

onMounted(async () => {
  syncAffiliateReferralCode()

  try {
    const settings = await getPublicSettings()
    registrationEnabled.value = settings.registration_enabled
    emailVerifyEnabled.value = settings.email_verify_enabled
    promoCodeEnabled.value = settings.promo_code_enabled
    invitationCodeEnabled.value = settings.invitation_code_enabled
    turnstileEnabled.value = settings.turnstile_enabled
    turnstileSiteKey.value = settings.turnstile_site_key || ''
    siteName.value = settings.site_name || 'Sub2API'
    linuxdoOAuthEnabled.value = settings.linuxdo_oauth_enabled
    wechatOAuthEnabled.value = isWeChatWebOAuthEnabled(settings)
    oidcOAuthEnabled.value = settings.oidc_oauth_enabled
    oidcOAuthProviderName.value = settings.oidc_oauth_provider_name || 'OIDC'
    registrationEmailSuffixWhitelist.value = normalizeRegistrationEmailSuffixWhitelist(
      settings.registration_email_suffix_whitelist || []
    )

    // 仅当优惠码功能开启时才从 URL 读取并预填
    if (promoCodeEnabled.value) {
      const promoParam = route.query.promo as string
      if (promoParam) {
        formData.promo_code = promoParam
        await validatePromoCodeDebounced(promoParam)
      }
    }
    syncAffiliateReferralCode()
  } catch (error) {
    console.error('Failed to load public settings:', error)
  } finally {
    settingsLoaded.value = true
  }

  // settings 加载完成后，若启用了 Turnstile 则启动占位/失败的状态机计时器
  if (turnstileEnabled.value && turnstileSiteKey.value && !turnstileToken.value) {
    startTurnstileTimers()
  }
})

watch(
  () => [route.query.aff, route.query.aff_code],
  () => {
    syncAffiliateReferralCode()
  }
)

onUnmounted(() => {
  if (promoValidateTimeout) {
    clearTimeout(promoValidateTimeout)
  }
  if (invitationValidateTimeout) {
    clearTimeout(invitationValidateTimeout)
  }
  clearTurnstileTimers()
})

// ==================== 优惠码校验 ====================

// 输入时去抖发起异步校验，结果改写 promoValidation 状态以驱动下划线颜色
function handlePromoCodeInput(): void {
  const code = formData.promo_code.trim()

  promoValidation.valid = false
  promoValidation.invalid = false
  promoValidation.bonusAmount = null
  promoValidation.message = ''

  if (!code) {
    promoValidating.value = false
    return
  }

  if (promoValidateTimeout) {
    clearTimeout(promoValidateTimeout)
  }

  promoValidateTimeout = setTimeout(() => {
    validatePromoCodeDebounced(code)
  }, 500)
}

async function validatePromoCodeDebounced(code: string): Promise<void> {
  if (!code.trim()) return

  promoValidating.value = true

  try {
    const result = await validatePromoCode(code)

    if (result.valid) {
      promoValidation.valid = true
      promoValidation.invalid = false
      promoValidation.bonusAmount = result.bonus_amount || 0
      promoValidation.message = ''
    } else {
      promoValidation.valid = false
      promoValidation.invalid = true
      promoValidation.bonusAmount = null
      promoValidation.message = getPromoErrorMessage(result.error_code)
    }
  } catch (error) {
    console.error('Failed to validate promo code:', error)
    promoValidation.valid = false
    promoValidation.invalid = true
    promoValidation.message = t('auth.promoCodeInvalid')
  } finally {
    promoValidating.value = false
  }
}

function getPromoErrorMessage(errorCode?: string): string {
  switch (errorCode) {
    case 'PROMO_CODE_NOT_FOUND':
      return t('auth.promoCodeNotFound')
    case 'PROMO_CODE_EXPIRED':
      return t('auth.promoCodeExpired')
    case 'PROMO_CODE_DISABLED':
      return t('auth.promoCodeDisabled')
    case 'PROMO_CODE_MAX_USED':
      return t('auth.promoCodeMaxUsed')
    case 'PROMO_CODE_ALREADY_USED':
      return t('auth.promoCodeAlreadyUsed')
    default:
      return t('auth.promoCodeInvalid')
  }
}

// ==================== 邀请码校验 ====================

function handleInvitationCodeInput(): void {
  const code = formData.invitation_code.trim()

  invitationValidation.valid = false
  invitationValidation.invalid = false
  invitationValidation.message = ''
  errors.invitation_code = ''

  if (!code) {
    return
  }

  if (invitationValidateTimeout) {
    clearTimeout(invitationValidateTimeout)
  }

  invitationValidateTimeout = setTimeout(() => {
    validateInvitationCodeDebounced(code)
  }, 500)
}

async function validateInvitationCodeDebounced(code: string): Promise<void> {
  invitationValidating.value = true

  try {
    const result = await validateInvitationCode(code)

    if (result.valid) {
      invitationValidation.valid = true
      invitationValidation.invalid = false
      invitationValidation.message = ''
    } else {
      invitationValidation.valid = false
      invitationValidation.invalid = true
      invitationValidation.message = getInvitationErrorMessage(result.error_code)
    }
  } catch {
    invitationValidation.valid = false
    invitationValidation.invalid = true
    invitationValidation.message = t('auth.invitationCodeInvalid')
  } finally {
    invitationValidating.value = false
  }
}

function getInvitationErrorMessage(errorCode?: string): string {
  switch (errorCode) {
    case 'INVITATION_CODE_NOT_FOUND':
      return t('auth.invitationCodeInvalid')
    case 'INVITATION_CODE_INVALID':
      return t('auth.invitationCodeInvalid')
    case 'INVITATION_CODE_USED':
      return t('auth.invitationCodeInvalid')
    case 'INVITATION_CODE_DISABLED':
      return t('auth.invitationCodeInvalid')
    default:
      return t('auth.invitationCodeInvalid')
  }
}

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

// ==================== 表单校验 ====================

function validateEmail(email: string): boolean {
  const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
  return emailRegex.test(email)
}

// 后端可配置邮箱后缀白名单，命中失败时给用户清晰提示
function buildEmailSuffixNotAllowedMessage(): string {
  const normalizedWhitelist = normalizeRegistrationEmailSuffixWhitelist(
    registrationEmailSuffixWhitelist.value
  )
  if (normalizedWhitelist.length === 0) {
    return t('auth.emailSuffixNotAllowed')
  }
  const separator = String(locale.value || '').toLowerCase().startsWith('zh') ? '、' : ', '
  return t('auth.emailSuffixNotAllowedWithAllowed', {
    suffixes: normalizedWhitelist.join(separator)
  })
}

function validateForm(): boolean {
  errors.email = ''
  errors.password = ''
  errors.turnstile = ''
  errors.invitation_code = ''

  let isValid = true

  if (!formData.email.trim()) {
    errors.email = t('auth.emailRequired')
    isValid = false
  } else if (!validateEmail(formData.email)) {
    errors.email = t('auth.invalidEmail')
    isValid = false
  } else if (
    !isRegistrationEmailSuffixAllowed(formData.email, registrationEmailSuffixWhitelist.value)
  ) {
    errors.email = buildEmailSuffixNotAllowedMessage()
    isValid = false
  }

  if (!formData.password) {
    errors.password = t('auth.passwordRequired')
    isValid = false
  } else if (formData.password.length < 6) {
    errors.password = t('auth.passwordMinLength')
    isValid = false
  }

  if (invitationCodeEnabled.value) {
    if (!formData.invitation_code.trim()) {
      errors.invitation_code = t('auth.invitationCodeRequired')
      isValid = false
    }
  }

  if (turnstileEnabled.value && !turnstileToken.value) {
    errors.turnstile = t('auth.completeVerification')
    isValid = false
  }

  return isValid
}

// ==================== 注册提交 ====================

// 提交注册：先做本地校验、再确认邀请码/优惠码异步校验已通过，最后调 store 完成注册
async function handleRegister(): Promise<void> {
  if (!settingsLoaded.value || !registrationEnabled.value) {
    return
  }

  errorMessage.value = ''

  if (!validateForm()) {
    return
  }

  // 优惠码：校验中需等待，已判定无效则阻断提交
  if (formData.promo_code.trim()) {
    if (promoValidating.value) {
      errorMessage.value = t('auth.promoCodeValidating')
      return
    }
    if (promoValidation.invalid) {
      errorMessage.value = t('auth.promoCodeInvalidCannotRegister')
      return
    }
  }

  // 邀请码：开启时校验中需等待，已判定无效则阻断；尚未确认有效会再触发一次校验
  if (invitationCodeEnabled.value) {
    if (invitationValidating.value) {
      errorMessage.value = t('auth.invitationCodeValidating')
      return
    }
    if (invitationValidation.invalid) {
      errorMessage.value = t('auth.invitationCodeInvalidCannotRegister')
      return
    }
    if (formData.invitation_code.trim() && !invitationValidation.valid) {
      errorMessage.value = t('auth.invitationCodeValidating')
      await validateInvitationCodeDebounced(formData.invitation_code.trim())
      if (!invitationValidation.valid) {
        errorMessage.value = t('auth.invitationCodeInvalidCannotRegister')
        return
      }
    }
  }

  isLoading.value = true

  try {
    const affCode = formData.aff_code.trim() || loadAffiliateReferralCode()
    if (affCode) {
      formData.aff_code = affCode
    }

    // 启用邮箱验证：先把表单数据存进 sessionStorage，跳到验证页继续完成注册
    if (emailVerifyEnabled.value) {
      sessionStorage.setItem(
        'register_data',
        JSON.stringify({
          email: formData.email,
          password: formData.password,
          turnstile_token: turnstileToken.value,
          promo_code: formData.promo_code || undefined,
          invitation_code: formData.invitation_code || undefined,
          ...(affCode ? { aff_code: affCode } : {})
        })
      )

      await router.push('/email-verify')
      return
    }

    // 否则直接注册并跳转到 dashboard
    await authStore.register({
      email: formData.email,
      password: formData.password,
      turnstile_token: turnstileEnabled.value ? turnstileToken.value : undefined,
      promo_code: formData.promo_code || undefined,
      invitation_code: formData.invitation_code || undefined,
      ...(affCode ? { aff_code: affCode } : {})
    })
    clearAffiliateReferralCode()

    appStore.showSuccess(t('auth.accountCreatedSuccess', { siteName: siteName.value }))

    await router.push('/dashboard')
  } catch (error: unknown) {
    if (turnstileRef.value) {
      turnstileRef.value.reset()
      turnstileToken.value = ''
    }

    errorMessage.value = buildAuthErrorMessage(error, {
      fallback: t('auth.registrationFailed')
    })

    appStore.showError(errorMessage.value)
  } finally {
    isLoading.value = false
  }
}
</script>
