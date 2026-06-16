<template>
  <div
    class="rounded-lg border border-gray-200 bg-gray-50 p-4 dark:border-dark-700 dark:bg-dark-800/40"
  >
      <div class="flex items-start gap-4">
      <div class="flex h-10 w-10 flex-shrink-0 items-center justify-center rounded-lg bg-brand-500">
        <Icon name="link" size="md" class="text-white" />
      </div>
      <div class="flex-1">
        <h4 class="mb-3 font-semibold text-gray-900 dark:text-white">{{ oauthTitle }}</h4>

        <!-- Auth Method Selection -->
        <div v-if="showMethodSelection" class="mb-4">
          <label class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300">
            {{ methodLabel }}
          </label>
          <div class="flex flex-wrap gap-4">
            <label class="flex cursor-pointer items-center gap-2">
              <input
                v-model="inputMethod"
                type="radio"
                value="manual"
                class="text-brand-600 focus:ring-brand-500"
              />
              <span class="text-sm text-gray-900 dark:text-white">{{
                t('admin.accounts.oauth.manualAuth')
              }}</span>
            </label>
            <label v-if="showCookieOption" class="flex cursor-pointer items-center gap-2">
              <input
                v-model="inputMethod"
                type="radio"
                value="cookie"
                class="text-brand-600 focus:ring-brand-500"
              />
              <span class="text-sm text-gray-900 dark:text-white">{{
                t('admin.accounts.oauth.cookieAutoAuth')
              }}</span>
            </label>
            <label v-if="showRefreshTokenOption" class="flex cursor-pointer items-center gap-2">
              <input
                v-model="inputMethod"
                type="radio"
                value="refresh_token"
                class="text-brand-600 focus:ring-brand-500"
              />
              <span class="text-sm text-gray-900 dark:text-white">{{
                t(getOAuthKey('refreshTokenAuth'))
              }}</span>
            </label>
            <label v-if="showMobileRefreshTokenOption" class="flex cursor-pointer items-center gap-2">
              <input
                v-model="inputMethod"
                type="radio"
                value="mobile_refresh_token"
                class="text-brand-600 focus:ring-brand-500"
              />
              <span class="text-sm text-gray-900 dark:text-white">{{
                t('admin.accounts.oauth.openai.mobileRefreshTokenAuth', '手动输入 Mobile RT')
              }}</span>
            </label>
            <label v-if="showSessionTokenOption" class="flex cursor-pointer items-center gap-2">
              <input
                v-model="inputMethod"
                type="radio"
                value="session_token"
                class="text-brand-600 focus:ring-brand-500"
              />
              <span class="text-sm text-gray-900 dark:text-white">{{
                t(getOAuthKey('sessionTokenAuth'))
              }}</span>
            </label>
            <label v-if="showAccessTokenOption" class="flex cursor-pointer items-center gap-2">
              <input
                v-model="inputMethod"
                type="radio"
                value="access_token"
                class="text-brand-600 focus:ring-brand-500"
              />
              <span class="text-sm text-gray-900 dark:text-white">{{
                t('admin.accounts.oauth.openai.accessTokenAuth', '手动输入 AT')
              }}</span>
            </label>
          </div>
        </div>

        <!-- Refresh Token Input (OpenAI / Antigravity / Mobile RT) -->
        <div v-if="inputMethod === 'refresh_token' || inputMethod === 'mobile_refresh_token'" class="space-y-4">
          <div
            class="rounded-lg border border-gray-200 bg-white p-4 dark:border-dark-700 dark:bg-dark-800/60"
          >
            <p class="mb-3 text-sm text-gray-600 dark:text-gray-300">
              {{ t(getOAuthKey('refreshTokenDesc')) }}
            </p>

            <!-- Refresh Token Input -->
            <div class="mb-4">
              <label
                class="mb-2 flex items-center gap-2 text-sm font-semibold text-gray-700 dark:text-gray-300"
              >
                <Icon name="key" size="sm" class="text-brand-500" />
                Refresh Token
                <span
                  v-if="parsedRefreshTokenCount > 1"
                  class="rounded-full bg-brand-500 px-2 py-0.5 text-xs text-white"
                >
                  {{ t('admin.accounts.oauth.keysCount', { count: parsedRefreshTokenCount }) }}
                </span>
              </label>
              <textarea
                v-model="refreshTokenInput"
                rows="3"
                class="input w-full resize-y font-mono text-sm"
                :placeholder="t(getOAuthKey('refreshTokenPlaceholder'))"
              ></textarea>
              <p
                v-if="parsedRefreshTokenCount > 1"
                class="mt-1 text-xs text-gray-500 dark:text-gray-400"
              >
                {{ t('admin.accounts.oauth.batchCreateAccounts', { count: parsedRefreshTokenCount }) }}
              </p>
            </div>

            <!-- Error Message -->
            <div
              v-if="error"
              class="mb-4 rounded-lg border border-danger/30 bg-danger-soft p-3 dark:border-danger/50 dark:bg-danger/20"
            >
              <p class="whitespace-pre-line text-sm text-danger dark:text-danger-soft">
                {{ error }}
              </p>
            </div>

            <!-- Validate Button -->
            <button
              type="button"
              class="btn btn-primary w-full"
              :disabled="loading || !refreshTokenInput.trim()"
              @click="handleValidateRefreshToken"
            >
              <LoadingSpinner v-if="loading" size="sm" color="current" class="-ml-1 mr-2" />
              <Icon v-else name="sparkles" size="sm" class="mr-2" />
              {{
                loading
                  ? t(getOAuthKey('validating'))
                  : t(getOAuthKey('validateAndCreate'))
              }}
            </button>
          </div>
        </div>

        <!-- Cookie Auto-Auth Form -->
        <div v-if="inputMethod === 'cookie'" class="space-y-4">
          <div
            class="rounded-lg border border-gray-200 bg-white p-4 dark:border-dark-700 dark:bg-dark-800/60"
          >
            <p class="mb-3 text-sm text-gray-600 dark:text-gray-300">
              {{ t('admin.accounts.oauth.cookieAutoAuthDesc') }}
            </p>

            <!-- sessionKey Input -->
            <div class="mb-4">
              <label
                class="mb-2 flex items-center gap-2 text-sm font-semibold text-gray-700 dark:text-gray-300"
              >
                <Icon name="key" size="sm" class="text-brand-500" />
                {{ t('admin.accounts.oauth.sessionKey') }}
                <span
                  v-if="parsedKeyCount > 1 && allowMultiple"
                  class="rounded-full bg-brand-500 px-2 py-0.5 text-xs text-white"
                >
                  {{ t('admin.accounts.oauth.keysCount', { count: parsedKeyCount }) }}
                </span>
                <button
                  v-if="showHelp"
                  type="button"
                  class="text-brand-500 hover:text-brand-600"
                  @click="showHelpDialog = !showHelpDialog"
                >
                  <Icon name="questionCircle" size="sm" />
                </button>
              </label>
              <textarea
                v-model="sessionKeyInput"
                rows="3"
                class="input w-full resize-y font-mono text-sm"
                :placeholder="
                  allowMultiple
                    ? t('admin.accounts.oauth.sessionKeyPlaceholder')
                    : t('admin.accounts.oauth.sessionKeyPlaceholderSingle')
                "
              ></textarea>
              <p
                v-if="parsedKeyCount > 1 && allowMultiple"
                class="mt-1 text-xs text-gray-500 dark:text-gray-400"
              >
                {{ t('admin.accounts.oauth.batchCreateAccounts', { count: parsedKeyCount }) }}
              </p>
            </div>

            <!-- Help Section -->
            <div
              v-if="showHelpDialog && showHelp"
              class="mb-4 rounded-lg border border-warning/30 bg-warning-soft p-3 dark:border-warning/50 dark:bg-warning/20"
            >
              <h5 class="mb-2 font-semibold text-warning-deep dark:text-warning-soft">
                {{ t('admin.accounts.oauth.howToGetSessionKey') }}
              </h5>
              <ol
                class="list-inside list-decimal space-y-1 text-xs text-warning-deep dark:text-warning-soft"
              >
                <li>{{ t('admin.accounts.oauth.step1') }}</li>
                <li>{{ t('admin.accounts.oauth.step2') }}</li>
                <li>{{ t('admin.accounts.oauth.step3') }}</li>
                <li>{{ t('admin.accounts.oauth.step4') }}</li>
                <li>{{ t('admin.accounts.oauth.step5') }}</li>
                <li>{{ t('admin.accounts.oauth.step6') }}</li>
              </ol>
              <p
                class="mt-2 text-xs text-warning dark:text-warning-soft"
                v-text="t('admin.accounts.oauth.sessionKeyFormat')"
              ></p>
            </div>

            <!-- Error Message -->
            <div
              v-if="error"
              class="mb-4 rounded-lg border border-danger/30 bg-danger-soft p-3 dark:border-danger/50 dark:bg-danger/20"
            >
              <p class="whitespace-pre-line text-sm text-danger dark:text-danger-soft">
                {{ error }}
              </p>
            </div>

            <!-- Auth Button -->
            <button
              type="button"
              class="btn btn-primary w-full"
              :disabled="loading || !sessionKeyInput.trim()"
              @click="handleCookieAuth"
            >
              <LoadingSpinner v-if="loading" size="sm" color="current" class="-ml-1 mr-2" />
              <Icon v-else name="sparkles" size="sm" class="mr-2" />
              {{
                loading
                  ? t('admin.accounts.oauth.authorizing')
                  : t('admin.accounts.oauth.startAutoAuth')
              }}
            </button>
          </div>
        </div>

        <!-- Manual Authorization Flow -->
        <div v-if="inputMethod === 'manual'" class="space-y-4">
          <p class="mb-4 text-sm text-gray-700 dark:text-gray-300">
            {{ oauthFollowSteps }}
          </p>

          <!-- Step 1: Generate Auth URL -->
          <div
            class="rounded-lg border border-gray-200 bg-white p-4 dark:border-dark-700 dark:bg-dark-800/60"
          >
            <div class="flex items-start gap-3">
              <div
                class="flex h-6 w-6 flex-shrink-0 items-center justify-center rounded-full bg-brand-500 text-xs font-bold text-white"
              >
                1
              </div>
              <div class="flex-1">
                <p class="mb-2 font-medium text-gray-900 dark:text-white">
                  {{ oauthStep1GenerateUrl }}
                </p>
                <div v-if="showProjectId && platform === 'gemini'" class="mb-3">
                  <label class="input-label flex items-center gap-2">
                    {{ t('admin.accounts.oauth.gemini.projectIdLabel') }}
                    <a
                      href="https://console.cloud.google.com/"
                      target="_blank"
                      rel="noopener noreferrer"
                      class="inline-flex items-center gap-1 text-xs font-normal text-brand-600 hover:text-brand-700 dark:text-brand-400"
                    >
                      <Icon name="questionCircle" size="xs" />
                      {{ t('admin.accounts.oauth.gemini.howToGetProjectId') }}
                    </a>
                  </label>
                  <input
                    v-model="projectId"
                    type="text"
                    class="input w-full font-mono text-sm"
                    :placeholder="t('admin.accounts.oauth.gemini.projectIdPlaceholder')"
                  />
                  <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">
                    {{ t('admin.accounts.oauth.gemini.projectIdHint') }}
                  </p>
                </div>
                <button
                  v-if="!authUrl"
                  type="button"
                  :disabled="loading"
                  class="btn btn-primary text-sm"
                  @click="handleGenerateUrl"
                >
                  <LoadingSpinner v-if="loading" size="sm" color="current" class="-ml-1 mr-2" />
                  <Icon v-else name="link" size="sm" class="mr-2" />
                  {{ loading ? t('admin.accounts.oauth.generating') : oauthGenerateAuthUrl }}
                </button>
                <div v-else class="space-y-3">
                  <div class="flex items-center gap-2">
                    <input
                      :value="authUrl"
                      readonly
                      type="text"
                      class="input flex-1 bg-gray-50 font-mono text-xs dark:bg-gray-700"
                    />
                    <button
                      type="button"
                      class="btn btn-secondary p-2"
                      title="Copy URL"
                      @click="handleCopyUrl"
                    >
                      <Icon v-if="!copied" name="clipboard" size="sm" />
                      <Icon
                        v-else
                        name="check"
                        size="sm"
                        class="text-success"
                        :stroke-width="2"
                      />
                    </button>
                  </div>
                  <button
                    type="button"
                    class="text-xs text-brand-600 hover:text-brand-700 dark:text-brand-400"
                    @click="handleRegenerate"
                  >
                    <Icon name="refresh" size="xs" class="mr-1 inline" />
                    {{ t('admin.accounts.oauth.regenerate') }}
                  </button>
                </div>
              </div>
            </div>
          </div>

          <!-- Step 2: Open URL and authorize -->
          <div
            class="rounded-lg border border-gray-200 bg-white p-4 dark:border-dark-700 dark:bg-dark-800/60"
          >
            <div class="flex items-start gap-3">
              <div
                class="flex h-6 w-6 flex-shrink-0 items-center justify-center rounded-full bg-brand-500 text-xs font-bold text-white"
              >
                2
              </div>
              <div class="flex-1">
                <p class="mb-2 font-medium text-gray-900 dark:text-white">
                  {{ oauthStep2OpenUrl }}
                </p>
                <p class="text-sm text-gray-600 dark:text-gray-300">
                  {{ oauthOpenUrlDesc }}
                </p>
                <!-- OpenAI Important Notice -->
                <div
                  v-if="isOpenAI"
                  class="mt-2 rounded border border-warning/40 bg-warning-soft p-3 dark:border-warning/50 dark:bg-warning/20"
                >
                  <p
                    class="text-xs text-warning-deep dark:text-warning-soft"
                    v-text="oauthImportantNotice"
                  ></p>
                </div>
                <!-- Proxy Warning (for non-OpenAI) -->
                <div
                  v-else-if="showProxyWarning"
                  class="mt-2 rounded border border-warning/40 bg-warning-soft p-3 dark:border-warning/50 dark:bg-warning/20"
                >
                  <p
                    class="text-xs text-warning-deep dark:text-warning-soft"
                    v-text="t('admin.accounts.oauth.proxyWarning')"
                  ></p>
                </div>
              </div>
            </div>
          </div>

          <!-- Step 3: Enter authorization code -->
          <div
            class="rounded-lg border border-gray-200 bg-white p-4 dark:border-dark-700 dark:bg-dark-800/60"
          >
            <div class="flex items-start gap-3">
              <div
                class="flex h-6 w-6 flex-shrink-0 items-center justify-center rounded-full bg-brand-500 text-xs font-bold text-white"
              >
                3
              </div>
              <div class="flex-1">
                <p class="mb-2 font-medium text-gray-900 dark:text-white">
                  {{ oauthStep3EnterCode }}
                </p>
                <p
                  class="mb-3 text-sm text-gray-600 dark:text-gray-300"
                  v-text="oauthAuthCodeDesc"
                ></p>
                <div>
                  <label class="input-label">
                    <Icon name="key" size="sm" class="mr-1 inline text-brand-500" />
                    {{ oauthAuthCode }}
                  </label>
                  <textarea
                    v-model="authCodeInput"
                    rows="3"
                    class="input w-full resize-none font-mono text-sm"
                    :placeholder="oauthAuthCodePlaceholder"
                  ></textarea>
                  <p class="mt-2 text-xs text-gray-500 dark:text-gray-400">
                    <Icon name="infoCircle" size="xs" class="mr-1 inline" />
                    {{ oauthAuthCodeHint }}
                  </p>

                  <!-- Gemini-specific state parameter warning -->
                  <div
                    v-if="platform === 'gemini'"
                    class="mt-3 rounded-lg border-2 border-warning/50 bg-warning-soft p-3 dark:border-warning/60 dark:bg-warning/20"
                  >
                    <div class="flex items-start gap-2">
                      <Icon
                        name="exclamationTriangle"
                        size="md"
                        class="flex-shrink-0 text-warning dark:text-warning-soft"
                        :stroke-width="2"
                      />
                      <div class="text-sm text-warning-deep dark:text-warning-soft">
                        <p class="font-semibold">{{ $t('admin.accounts.oauth.gemini.stateWarningTitle') }}</p>
                        <p class="mt-1">{{ $t('admin.accounts.oauth.gemini.stateWarningDesc') }}</p>
                      </div>
                    </div>
                  </div>
                </div>

                <!-- Error Message -->
                <div
                  v-if="error"
                  class="mt-3 rounded-lg border border-danger/30 bg-danger-soft p-3 dark:border-danger/50 dark:bg-danger/20"
                >
                  <p class="whitespace-pre-line text-sm text-danger dark:text-danger-soft">
                    {{ error }}
                  </p>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useClipboard } from '@/composables/useClipboard'
import Icon from '@/components/icons/Icon.vue'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import type { AddMethod, AuthInputMethod } from '@/composables/useAccountOAuth'
import type { AccountPlatform } from '@/types'

interface Props {
  addMethod: AddMethod
  authUrl?: string
  sessionId?: string
  loading?: boolean
  error?: string
  showHelp?: boolean
  showProxyWarning?: boolean
  allowMultiple?: boolean
  methodLabel?: string
  showCookieOption?: boolean // Whether to show cookie auto-auth option
  showRefreshTokenOption?: boolean // Whether to show refresh token input option (OpenAI only)
  showMobileRefreshTokenOption?: boolean // Whether to show mobile refresh token option (OpenAI only)
  showSessionTokenOption?: boolean
  showAccessTokenOption?: boolean
  platform?: AccountPlatform // Platform type for different UI/text
  showProjectId?: boolean // New prop to control project ID visibility
}

const props = withDefaults(defineProps<Props>(), {
  authUrl: '',
  sessionId: '',
  loading: false,
  error: '',
  showHelp: true,
  showProxyWarning: true,
  allowMultiple: false,
  methodLabel: 'Authorization Method',
  showCookieOption: true,
  showRefreshTokenOption: false,
  showMobileRefreshTokenOption: false,
  showSessionTokenOption: false,
  showAccessTokenOption: false,
  platform: 'anthropic',
  showProjectId: true
})

const emit = defineEmits<{
  'generate-url': []
  'exchange-code': [code: string]
  'cookie-auth': [sessionKey: string]
  'validate-refresh-token': [refreshToken: string]
  'validate-mobile-refresh-token': [refreshToken: string]
  'validate-session-token': [sessionToken: string]
  'import-access-token': [accessToken: string]
  'update:inputMethod': [method: AuthInputMethod]
}>()

const { t } = useI18n()

const isOpenAI = computed(() => props.platform === 'openai')

// Get translation key based on platform
const getOAuthKey = (key: string) => {
  if (props.platform === 'openai') return `admin.accounts.oauth.openai.${key}`
  if (props.platform === 'gemini') return `admin.accounts.oauth.gemini.${key}`
  if (props.platform === 'antigravity') return `admin.accounts.oauth.antigravity.${key}`
  return `admin.accounts.oauth.${key}`
}

// Computed translations for current platform
const oauthTitle = computed(() => t(getOAuthKey('title')))
const oauthFollowSteps = computed(() => t(getOAuthKey('followSteps')))
const oauthStep1GenerateUrl = computed(() => t(getOAuthKey('step1GenerateUrl')))
const oauthGenerateAuthUrl = computed(() => t(getOAuthKey('generateAuthUrl')))
const oauthStep2OpenUrl = computed(() => t(getOAuthKey('step2OpenUrl')))
const oauthOpenUrlDesc = computed(() => t(getOAuthKey('openUrlDesc')))
const oauthStep3EnterCode = computed(() => t(getOAuthKey('step3EnterCode')))
const oauthAuthCodeDesc = computed(() => t(getOAuthKey('authCodeDesc')))
const oauthAuthCode = computed(() => t(getOAuthKey('authCode')))
const oauthAuthCodePlaceholder = computed(() => t(getOAuthKey('authCodePlaceholder')))
const oauthAuthCodeHint = computed(() => t(getOAuthKey('authCodeHint')))
const oauthImportantNotice = computed(() => {
  if (props.platform === 'openai') return t('admin.accounts.oauth.openai.importantNotice')
  if (props.platform === 'antigravity') return t('admin.accounts.oauth.antigravity.importantNotice')
  return ''
})

// Local state
const inputMethod = ref<AuthInputMethod>(props.showCookieOption ? 'manual' : 'manual')
const authCodeInput = ref('')
const sessionKeyInput = ref('')
const refreshTokenInput = ref('')
const sessionTokenInput = ref('')
const showHelpDialog = ref(false)
const oauthState = ref('')
const projectId = ref('')

// Computed: show method selection when either cookie or refresh token option is enabled
const showMethodSelection = computed(() => props.showCookieOption || props.showRefreshTokenOption || props.showMobileRefreshTokenOption || props.showSessionTokenOption || props.showAccessTokenOption)

// Clipboard
const { copied, copyToClipboard } = useClipboard()

// Computed
const parsedKeyCount = computed(() => {
  return sessionKeyInput.value
    .split('\n')
    .map((k) => k.trim())
    .filter((k) => k).length
})

// Computed: count of refresh tokens entered
const parsedRefreshTokenCount = computed(() => {
  return refreshTokenInput.value
    .split('\n')
    .map((rt) => rt.trim())
    .filter((rt) => rt).length
})

// Watchers
watch(inputMethod, (newVal) => {
  emit('update:inputMethod', newVal)
})

// Auto-extract code from callback URL (OpenAI/Gemini/Antigravity)
// e.g., http://localhost:8085/callback?code=xxx...&state=...
watch(authCodeInput, (newVal) => {
  if (props.platform !== 'openai' && props.platform !== 'gemini' && props.platform !== 'antigravity') return

  const trimmed = newVal.trim()
  // Check if it looks like a URL with code parameter
  if (trimmed.includes('?') && trimmed.includes('code=')) {
    try {
      // Try to parse as URL
      const url = new URL(trimmed)
      const code = url.searchParams.get('code')
      const stateParam = url.searchParams.get('state')
      if ((props.platform === 'openai' || props.platform === 'gemini' || props.platform === 'antigravity') && stateParam) {
        oauthState.value = stateParam
      }
      if (code && code !== trimmed) {
        // Replace the input with just the code
        authCodeInput.value = code
      }
    } catch {
      // If URL parsing fails, try regex extraction
      const match = trimmed.match(/[?&]code=([^&]+)/)
      const stateMatch = trimmed.match(/[?&]state=([^&]+)/)
      if ((props.platform === 'openai' || props.platform === 'gemini' || props.platform === 'antigravity') && stateMatch && stateMatch[1]) {
        oauthState.value = stateMatch[1]
      }
      if (match && match[1] && match[1] !== trimmed) {
        authCodeInput.value = match[1]
      }
    }
  }
})

// Methods
const handleGenerateUrl = () => {
  emit('generate-url')
}

const handleCopyUrl = () => {
  if (props.authUrl) {
    copyToClipboard(props.authUrl, 'URL copied to clipboard')
  }
}

const handleRegenerate = () => {
  authCodeInput.value = ''
  emit('generate-url')
}

const handleCookieAuth = () => {
  if (sessionKeyInput.value.trim()) {
    emit('cookie-auth', sessionKeyInput.value)
  }
}

const handleValidateRefreshToken = () => {
  if (refreshTokenInput.value.trim()) {
    if (inputMethod.value === 'mobile_refresh_token') {
      emit('validate-mobile-refresh-token', refreshTokenInput.value.trim())
    } else {
      emit('validate-refresh-token', refreshTokenInput.value.trim())
    }
  }
}

// Expose methods and state
defineExpose({
  authCode: authCodeInput,
  oauthState,
  projectId,
  sessionKey: sessionKeyInput,
  refreshToken: refreshTokenInput,
  sessionToken: sessionTokenInput,
  inputMethod,
  reset: () => {
    authCodeInput.value = ''
    oauthState.value = ''
    projectId.value = ''
    sessionKeyInput.value = ''
    refreshTokenInput.value = ''
    sessionTokenInput.value = ''
    inputMethod.value = 'manual'
    showHelpDialog.value = false
  }
})
</script>
