<template>
  <AppLayout>
    <div class="mx-auto max-w-7xl space-y-5">
      <!-- Hero：与兑换码 hero 同款渐变 + 标签 + 大标题 + 描述 -->
      <header class="docs-hero">
        <div class="relative z-10 max-w-3xl">
          <div class="docs-hero-tag">
            <Icon name="book" size="sm" />
            用户接入文档
          </div>
          <h1 class="mt-3 text-3xl font-semibold tracking-normal text-gray-950 dark:text-white md:text-[34px]">
            API 接入教程
          </h1>
          <p class="mt-3 text-sm leading-6 text-gray-600 dark:text-dark-200 md:text-base">
            按这里准备密钥、选择模型和填写 Base URL，即可把本站 API 接入 Claude Code、Codex、OpenAI SDK、Gemini 兼容客户端或普通 HTTP 请求。
          </p>
        </div>
      </header>

      <!-- 客户端拼接路径提示（信息提示条） -->
      <div class="docs-hint">
        <Icon name="infoCircle" size="sm" class="mt-0.5 flex-shrink-0 text-sky-600 dark:text-sky-300" />
        <p class="text-sm leading-6 text-sky-900 dark:text-sky-100">
          如果客户端会自动拼接 <code class="docs-hint-code">/v1/messages</code>，Claude Code 这类客户端通常填写站点根地址；OpenAI 兼容 SDK 和大多数中转客户端通常填写 <code class="docs-hint-code">/v1</code>。
        </p>
      </div>

      <!-- 快速开始三步 -->
      <section class="grid grid-cols-1 gap-3 md:grid-cols-3">
        <article
          v-for="item in quickStart"
          :key="item.title"
          class="docs-step-card"
        >
          <div class="flex items-start gap-3">
            <span class="docs-step-no">{{ item.no }}</span>
            <div class="min-w-0">
              <h2 class="text-sm font-semibold text-gray-950 dark:text-white">{{ item.title }}</h2>
              <p class="mt-1 text-xs leading-5 text-gray-600 dark:text-dark-300">{{ item.desc }}</p>
              <router-link
                v-if="item.to"
                :to="item.to"
                class="mt-2 inline-flex items-center gap-1 text-xs font-medium text-sky-600 transition-colors hover:text-sky-700 dark:text-sky-300 dark:hover:text-sky-200"
              >
                前往配置
                <Icon name="arrowRight" size="xs" />
              </router-link>
            </div>
          </div>
        </article>
      </section>

      <!-- 本站可用地址（violet 头部） -->
      <section class="docs-panel">
        <div class="docs-panel-header docs-panel-header-violet">
          <div class="docs-panel-icon docs-panel-icon-violet">
            <Icon name="link" size="md" />
          </div>
          <div class="min-w-0">
            <h2 class="text-base font-semibold text-gray-950 dark:text-white">本站可用地址</h2>
            <p class="mt-0.5 text-sm text-gray-500 dark:text-dark-400">
              不同客户端使用不同的 Base URL，按下方说明选择
            </p>
          </div>
        </div>
        <div class="p-5 sm:p-6">
          <div class="grid grid-cols-1 gap-3 lg:grid-cols-2">
            <div
              v-for="endpoint in endpoints"
              :key="endpoint.label"
              class="rounded-lg border border-gray-200 bg-gray-50/60 p-3 dark:border-dark-700 dark:bg-dark-800/40"
            >
              <div class="text-sm font-semibold text-gray-950 dark:text-white">{{ endpoint.label }}</div>
              <code class="mt-2 block break-all rounded-md bg-white px-3 py-2 font-mono text-xs text-gray-800 ring-1 ring-gray-200 dark:bg-dark-900 dark:text-gray-100 dark:ring-dark-700">
                {{ endpoint.url }}
              </code>
              <p class="mt-2 text-xs leading-5 text-gray-600 dark:text-dark-300">{{ endpoint.desc }}</p>
            </div>
          </div>

          <div v-if="customEndpoints.length" class="mt-5">
            <div class="text-sm font-semibold text-gray-950 dark:text-white">管理员公开的自定义端点</div>
            <div class="mt-2 grid grid-cols-1 gap-2 lg:grid-cols-2">
              <div
                v-for="endpoint in customEndpoints"
                :key="endpoint.endpoint"
                class="rounded-md border border-gray-200 bg-white p-3 dark:border-dark-700 dark:bg-dark-900"
              >
                <div class="text-xs font-medium text-gray-700 dark:text-dark-200">{{ endpoint.name || '自定义端点' }}</div>
                <code class="mt-1 block break-all font-mono text-xs text-violet-700 dark:text-violet-300">{{ endpoint.endpoint }}</code>
                <p v-if="endpoint.description" class="mt-1 text-xs leading-5 text-gray-500 dark:text-dark-400">
                  {{ endpoint.description }}
                </p>
              </div>
            </div>
          </div>
        </div>
      </section>

      <!-- 接入示例：单个大 panel + 顶部 tab 切换，代码区占满宽度，避免每个 snippet 被压成小卡片 -->
      <section class="space-y-3">
        <div class="flex items-center gap-2 px-1">
          <Icon name="terminal" size="md" class="text-emerald-600 dark:text-emerald-300" />
          <h2 class="text-lg font-semibold text-gray-950 dark:text-white">复制即可改的接入示例</h2>
        </div>

        <article class="docs-panel overflow-hidden">
          <!-- Tab 导航：emerald 浅色背景，下划线指示活动 tab -->
          <div class="snippet-tabs">
            <button
              v-for="(snippet, idx) in snippets"
              :key="snippet.title"
              type="button"
              :class="[
                'snippet-tab',
                activeSnippetIdx === idx && 'snippet-tab-active',
              ]"
              @click="activeSnippetIdx = idx"
            >
              <Icon name="terminal" size="xs" class="opacity-70" />
              {{ snippet.title }}
            </button>
          </div>

          <!-- 当前 snippet 的标题/描述 + 代码 -->
          <template v-if="activeSnippet">
            <div class="px-5 pt-4 pb-3 sm:px-6">
              <div class="flex flex-wrap items-center gap-2">
                <h3 class="text-sm font-semibold text-gray-950 dark:text-white">{{ activeSnippet.title }}</h3>
                <span class="rounded-md bg-emerald-50 px-2 py-0.5 text-[11px] font-medium text-emerald-700 ring-1 ring-emerald-200 dark:bg-emerald-900/20 dark:text-emerald-300 dark:ring-emerald-800/60">
                  {{ activeSnippet.tag }}
                </span>
              </div>
              <p class="mt-1 text-xs leading-5 text-gray-500 dark:text-dark-400">{{ activeSnippet.desc }}</p>
            </div>
            <div class="relative">
              <!-- 一键复制按钮：固定在代码区右上角 -->
              <button
                type="button"
                class="absolute right-3 top-3 z-10 inline-flex items-center gap-1.5 rounded-md bg-slate-800/80 px-2.5 py-1.5 text-xs font-medium text-slate-100 backdrop-blur transition-colors hover:bg-slate-700/90"
                @click="copySnippet(activeSnippet)"
              >
                <Icon
                  :name="copiedSnippetTitle === activeSnippet.title ? 'check' : 'copy'"
                  size="xs"
                />
                {{ copiedSnippetTitle === activeSnippet.title ? '已复制' : '复制' }}
              </button>
              <pre class="overflow-x-auto bg-slate-950 p-5 text-xs leading-6 text-slate-100 sm:p-6"><code>{{ activeSnippet.code }}</code></pre>
            </div>
          </template>
        </article>
      </section>

      <!-- 常见错误 / 使用安全提醒：amber + rose 双色卡片 -->
      <section class="grid grid-cols-1 gap-4 xl:grid-cols-2">
        <article class="docs-panel">
          <div class="docs-panel-header docs-panel-header-amber">
            <div class="docs-panel-icon docs-panel-icon-amber">
              <Icon name="exclamationCircle" size="md" />
            </div>
            <div class="min-w-0">
              <h2 class="text-base font-semibold text-gray-950 dark:text-white">常见错误</h2>
              <p class="mt-0.5 text-sm text-gray-500 dark:text-dark-400">遇到这些状态码时的排查方向</p>
            </div>
          </div>
          <div class="p-5 sm:p-6">
            <div class="divide-y divide-gray-100 dark:divide-dark-700">
              <div
                v-for="error in commonErrors"
                :key="error.code"
                class="py-3 first:pt-0 last:pb-0"
              >
                <div class="flex flex-wrap items-center gap-2">
                  <code class="rounded-md bg-amber-50 px-2 py-0.5 font-mono text-xs text-amber-700 ring-1 ring-amber-200 dark:bg-amber-900/20 dark:text-amber-300 dark:ring-amber-800/60">{{ error.code }}</code>
                  <span class="text-sm font-semibold text-gray-950 dark:text-white">{{ error.title }}</span>
                </div>
                <p class="mt-1 text-xs leading-5 text-gray-600 dark:text-dark-300">{{ error.desc }}</p>
              </div>
            </div>
          </div>
        </article>

        <article class="docs-panel">
          <div class="docs-panel-header docs-panel-header-rose">
            <div class="docs-panel-icon docs-panel-icon-rose">
              <Icon name="shield" size="md" />
            </div>
            <div class="min-w-0">
              <h2 class="text-base font-semibold text-gray-950 dark:text-white">使用安全提醒</h2>
              <p class="mt-0.5 text-sm text-gray-500 dark:text-dark-400">保护密钥，避免触发风控</p>
            </div>
          </div>
          <div class="space-y-3 p-5 text-sm leading-6 text-gray-700 dark:text-dark-200 sm:p-6">
            <p>请不要把 API 密钥发给他人，也不要放到公开仓库、前端网页或截图里。密钥泄露后，别人可以直接消耗你的额度。</p>
            <p>如果页面提示内容可能触发安全限制，请调整问题再发送。系统会保护上游账号稳定，明显违规或重复命中可能会被管理员复核。</p>
            <p>Claude Code、Codex 或 SDK 修改环境变量后，建议重启终端或应用，确认新配置已经生效。</p>
          </div>
        </article>
      </section>

      <!-- 官方参考：indigo 头部 -->
      <section class="docs-panel">
        <div class="docs-panel-header docs-panel-header-indigo">
          <div class="docs-panel-icon docs-panel-icon-indigo">
            <Icon name="externalLink" size="md" />
          </div>
          <div class="min-w-0">
            <h2 class="text-base font-semibold text-gray-950 dark:text-white">官方参考</h2>
            <p class="mt-0.5 text-sm text-gray-500 dark:text-dark-400">遇到细节问题可查阅各家官方文档</p>
          </div>
        </div>
        <div class="p-5 sm:p-6">
          <div class="grid grid-cols-1 gap-2 md:grid-cols-2">
            <a
              v-for="source in sources"
              :key="source.href"
              :href="source.href"
              target="_blank"
              rel="noopener noreferrer"
              class="flex items-center justify-between rounded-md border border-gray-200 px-3 py-2 text-sm text-gray-700 transition-colors hover:border-indigo-300 hover:text-indigo-700 dark:border-dark-700 dark:text-dark-200 dark:hover:border-indigo-700 dark:hover:text-indigo-300"
            >
              <span>{{ source.label }}</span>
              <Icon name="externalLink" size="xs" class="opacity-60" />
            </a>
          </div>
        </div>
      </section>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import { useAppStore } from '@/stores'

const appStore = useAppStore()

// 接入示例当前激活的 tab 与"已复制"反馈状态
const activeSnippetIdx = ref(0)
const copiedSnippetTitle = ref<string | null>(null)
let copyResetTimer: ReturnType<typeof setTimeout> | null = null

// 把当前 snippet 的代码复制到剪贴板，2 秒后还原"已复制"提示
async function copySnippet(snippet: { title: string; code: string }): Promise<void> {
  try {
    await navigator.clipboard.writeText(snippet.code)
    copiedSnippetTitle.value = snippet.title
    if (copyResetTimer) clearTimeout(copyResetTimer)
    copyResetTimer = setTimeout(() => {
      copiedSnippetTitle.value = null
    }, 2000)
  } catch (error) {
    console.error('Failed to copy snippet:', error)
  }
}

onMounted(() => {
  if (!appStore.publicSettingsLoaded) {
    void appStore.fetchPublicSettings()
  }
})

// 移除尾部斜杠，避免拼接 URL 时出现重复 /
function normalizeUrl(url: string): string {
  return url.replace(/\/+$/, '')
}

// 把后台配置的 api_base_url 还原为站点根地址（去掉 /v1、/v1beta、/backend-api/codex 后缀）
function rootFromConfiguredUrl(url: string): string {
  const normalized = normalizeUrl(url)
  if (!normalized) {
    return typeof window === 'undefined' ? 'https://你的域名' : window.location.origin
  }
  if (!/^https?:\/\//i.test(normalized)) {
    return typeof window === 'undefined' ? 'https://你的域名' : window.location.origin
  }
  return normalized
    .replace(/\/backend-api\/codex$/i, '')
    .replace(/\/v1beta$/i, '')
    .replace(/\/v1$/i, '')
}

const siteRootUrl = computed(() => rootFromConfiguredUrl(appStore.cachedPublicSettings?.api_base_url || appStore.apiBaseUrl || ''))
const openAIBaseUrl = computed(() => `${siteRootUrl.value}/v1`)
const geminiBaseUrl = computed(() => `${siteRootUrl.value}/v1beta`)
const codexDirectBaseUrl = computed(() => `${siteRootUrl.value}/backend-api/codex`)
const customEndpoints = computed(() => appStore.cachedPublicSettings?.custom_endpoints || [])

const quickStart = [
  { no: '1', title: '创建 API 密钥', desc: '进入 API 密钥页面创建密钥，复制以 sk- 开头的 Key，并确认密钥有可访问分组。', to: '/keys' },
  { no: '2', title: '确认可用模型', desc: '进入可用渠道或模型广场，查看你的分组能使用哪些模型，以及模型当前是否可用。', to: '/available-channels' },
  { no: '3', title: '填写 Base URL', desc: '根据客户端类型填写根地址、/v1、/v1beta 或 Codex 专用路径，填错路径通常会 404。', to: '' },
]

const endpoints = computed(() => [
  { label: 'Claude Code / Anthropic 网关根地址', url: siteRootUrl.value, desc: 'Claude Code 会调用 `/v1/messages` 和 `/v1/messages/count_tokens`，通常填根地址。' },
  { label: 'OpenAI 兼容 Base URL', url: openAIBaseUrl.value, desc: '适用于 OpenAI SDK、Chat Completions、Responses、图片生成和大多数中转客户端。' },
  { label: 'Codex Responses 直连地址', url: codexDirectBaseUrl.value, desc: '适用于需要走 Codex `/backend-api/codex/responses` 兼容入口的客户端。' },
  { label: 'Gemini Native Base URL', url: geminiBaseUrl.value, desc: '适用于走 Gemini `/v1beta/models` 兼容入口的客户端。' },
])

const snippets = computed(() => [
  {
    title: 'Claude Code',
    tag: 'Anthropic',
    desc: '推荐在终端环境变量中配置。若客户端要求完整 API 地址，再改用 `/v1`。',
    code: [
      `ANTHROPIC_BASE_URL=${siteRootUrl.value}`,
      'ANTHROPIC_AUTH_TOKEN=sk-你的密钥',
      '',
      '# Windows PowerShell 示例',
      `$env:ANTHROPIC_BASE_URL="${siteRootUrl.value}"`,
      '$env:ANTHROPIC_AUTH_TOKEN="sk-你的密钥"',
    ].join('\n'),
  },
  {
    title: 'Codex CLI',
    tag: 'OpenAI 兼容',
    desc: '不同 Codex 版本配置方式可能不同。环境变量不生效时，使用 `~/.codex/config.toml` 配置 provider。',
    code: [
      'OPENAI_API_KEY=sk-你的密钥',
      `OPENAI_BASE_URL=${openAIBaseUrl.value}`,
      '',
      '# ~/.codex/config.toml 示例',
      '[model_providers.sub2api]',
      'name = "Sub2API"',
      `base_url = "${openAIBaseUrl.value}"`,
      'env_key = "OPENAI_API_KEY"',
      'wire_api = "responses"',
      '',
      'model_provider = "sub2api"',
    ].join('\n'),
  },
  {
    title: 'OpenAI SDK',
    tag: 'Node / Python',
    desc: 'SDK 里把 baseURL/base_url 指向本站 `/v1`，密钥使用本站生成的 API Key。',
    code: [
      '// Node.js',
      'import OpenAI from "openai"',
      'const client = new OpenAI({',
      '  apiKey: "sk-你的密钥",',
      `  baseURL: "${openAIBaseUrl.value}",`,
      '})',
      '',
      '# Python',
      'from openai import OpenAI',
      `client = OpenAI(api_key="sk-你的密钥", base_url="${openAIBaseUrl.value}")`,
    ].join('\n'),
  },
  {
    title: 'HTTP 快速测试',
    tag: 'curl',
    desc: '先用 curl 确认密钥、模型名和 Base URL 都正确，再接入客户端。',
    code: [
      `curl ${openAIBaseUrl.value}/chat/completions \\`,
      '  -H "Authorization: Bearer sk-你的密钥" \\',
      '  -H "Content-Type: application/json" \\',
      '  -d "{\\"model\\":\\"gpt-4o-mini\\",\\"messages\\":[{\\"role\\":\\"user\\",\\"content\\":\\"hello\\"}]}"',
      '',
      `curl ${openAIBaseUrl.value}/messages \\`,
      '  -H "x-api-key: sk-你的密钥" \\',
      '  -H "anthropic-version: 2023-06-01" \\',
      '  -H "Content-Type: application/json" \\',
      '  -d "{\\"model\\":\\"claude-sonnet-4-5\\",\\"max_tokens\\":256,\\"messages\\":[{\\"role\\":\\"user\\",\\"content\\":\\"hello\\"}]}"',
    ].join('\n'),
  },
])

// 当前激活的 snippet（响应 activeSnippetIdx 切换）
const activeSnippet = computed(() => snippets.value[activeSnippetIdx.value])

const commonErrors = [
  { code: '401', title: '密钥无效', desc: '检查 API Key 是否复制完整，Header 是否使用 Bearer 或 x-api-key，密钥是否被禁用。' },
  { code: '403', title: '没有权限或触发风控', desc: '检查密钥分组、订阅权益、模型范围；如果是安全提醒，请调整问题后重试。' },
  { code: '404', title: 'Base URL 填错', desc: 'Claude Code 通常填根地址，OpenAI SDK 通常填 `/v1`，Gemini 填 `/v1beta`。' },
  { code: 'model_not_found', title: '模型不可用', desc: '检查模型名是否和可用渠道展示一致，以及你的分组是否有这个模型。' },
  { code: 'insufficient_quota', title: '额度不足', desc: '检查余额、订阅是否到期、兑换码权益是否已生效。' },
  { code: 'still official api', title: '仍在请求官方接口', desc: '修改环境变量后重启终端或客户端，确认配置文件路径被当前客户端读取。' },
]

const sources = [
  { label: 'Claude Code LLM Gateway', href: 'https://code.claude.com/docs/zh-CN/llm-gateway' },
  { label: 'Claude Code settings', href: 'https://docs.claude.com/en/docs/claude-code/settings' },
  { label: 'OpenAI Codex configuration', href: 'https://developers.openai.com/codex/config' },
  { label: 'OpenAI API reference', href: 'https://platform.openai.com/docs/api-reference' },
]
</script>

<style scoped>
/* ============================================================================
 * 文档页彩色风格：与兑换码页保持视觉一致（hero 渐变 + 浅色头部 panel）
 * 设计原则：每个 section 用一个色调标识信息维度，呼应兑换码的彩色 stat 卡
 * ========================================================================= */

/* Hero：与 redeem-hero 同款渐变 */
.docs-hero {
  position: relative;
  overflow: hidden;
  border-radius: 0.5rem;
  border: 1px solid rgb(219 234 254);
  background:
    radial-gradient(circle at 76% 22%, rgb(56 189 248 / 0.26), transparent 28%),
    radial-gradient(circle at 94% 12%, rgb(168 85 247 / 0.18), transparent 24%),
    linear-gradient(135deg, rgb(240 249 255), rgb(255 255 255) 48%, rgb(245 243 255));
  padding: 2rem;
  box-shadow: 0 18px 44px -34px rgb(15 23 42 / 0.55);
}

:global(:root.dark) .docs-hero {
  border-color: rgb(55 65 81);
  background:
    radial-gradient(circle at 76% 22%, rgb(14 165 233 / 0.16), transparent 30%),
    radial-gradient(circle at 94% 12%, rgb(139 92 246 / 0.14), transparent 24%),
    linear-gradient(135deg, rgb(15 23 42), rgb(17 24 39));
  box-shadow: none;
}

@media (min-width: 640px) {
  .docs-hero {
    padding: 2.25rem 2.5rem;
  }
}

.docs-hero-tag {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  border-radius: 0.5rem;
  background: rgb(255 255 255 / 0.7);
  padding: 0.25rem 0.625rem;
  font-size: 0.75rem;
  font-weight: 500;
  color: rgb(2 132 199);
  backdrop-filter: blur(8px);
}

:global(:root.dark) .docs-hero-tag {
  background: rgb(31 41 55 / 0.7);
  color: rgb(125 211 252);
}

/* 信息提示条：sky 色 */
.docs-hint {
  display: flex;
  align-items: flex-start;
  gap: 0.625rem;
  border-radius: 0.5rem;
  border: 1px solid rgb(186 230 253);
  background: rgb(240 249 255);
  padding: 0.875rem 1rem;
}

:global(:root.dark) .docs-hint {
  border-color: rgb(7 89 133 / 0.5);
  background: rgb(8 47 73 / 0.3);
}

.docs-hint-code {
  display: inline-block;
  background: rgb(255 255 255 / 0.7);
  border-radius: 0.25rem;
  padding: 0.05rem 0.35rem;
  font-family: ui-monospace, monospace;
  font-size: 0.75rem;
  color: rgb(2 132 199);
}

:global(:root.dark) .docs-hint-code {
  background: rgb(31 41 55 / 0.7);
  color: rgb(125 211 252);
}

/* 步骤卡片：白底 + 微阴影 + 数字色块 */
.docs-step-card {
  position: relative;
  border-radius: 0.5rem;
  border: 1px solid rgb(229 231 235);
  background: rgb(255 255 255);
  padding: 1rem 1.125rem;
  box-shadow: 0 1px 2px rgb(0 0 0 / 0.04);
  transition: border-color 0.15s, box-shadow 0.15s;
}

.docs-step-card:hover {
  border-color: rgb(186 230 253);
  box-shadow: 0 4px 16px -8px rgb(15 23 42 / 0.08);
}

:global(:root.dark) .docs-step-card {
  border-color: rgb(55 65 81);
  background: rgb(31 41 55);
  box-shadow: none;
}

.docs-step-no {
  display: inline-flex;
  flex-shrink: 0;
  height: 2rem;
  width: 2rem;
  align-items: center;
  justify-content: center;
  border-radius: 0.5rem;
  background: linear-gradient(135deg, rgb(56 189 248), rgb(99 102 241));
  font-size: 0.875rem;
  font-weight: 600;
  color: rgb(255 255 255);
  box-shadow: 0 4px 12px -4px rgb(56 189 248 / 0.5);
}

/* 通用 panel：与兑换码 redeem-panel 一致 */
.docs-panel {
  overflow: hidden;
  border-radius: 0.5rem;
  border: 1px solid rgb(229 231 235);
  background: rgb(255 255 255);
  box-shadow: 0 18px 44px -34px rgb(15 23 42 / 0.55);
}

:global(:root.dark) .docs-panel {
  border-color: rgb(55 65 81);
  background: rgb(31 41 55);
  box-shadow: none;
}

/* Panel 头部：浅色背景 + 圆角 icon 盒 + 标题描述 */
.docs-panel-header {
  display: flex;
  align-items: flex-start;
  gap: 0.75rem;
  padding: 1.125rem 1.25rem;
  border-bottom: 1px solid;
}

@media (min-width: 640px) {
  .docs-panel-header {
    padding: 1.125rem 1.5rem;
  }
}

.docs-panel-header-violet {
  border-color: rgb(237 233 254 / 0.7);
  background: rgb(245 243 255 / 0.6);
}
.docs-panel-header-emerald {
  border-color: rgb(209 250 229 / 0.7);
  background: rgb(236 253 245 / 0.6);
}
.docs-panel-header-amber {
  border-color: rgb(254 243 199 / 0.7);
  background: rgb(255 251 235 / 0.6);
}
.docs-panel-header-rose {
  border-color: rgb(255 228 230 / 0.7);
  background: rgb(255 241 242 / 0.6);
}
.docs-panel-header-indigo {
  border-color: rgb(224 231 255 / 0.7);
  background: rgb(238 242 255 / 0.6);
}

:global(:root.dark) .docs-panel-header-violet {
  border-color: rgb(55 65 81);
  background: rgb(139 92 246 / 0.08);
}
:global(:root.dark) .docs-panel-header-emerald {
  border-color: rgb(55 65 81);
  background: rgb(16 185 129 / 0.08);
}
:global(:root.dark) .docs-panel-header-amber {
  border-color: rgb(55 65 81);
  background: rgb(245 158 11 / 0.08);
}
:global(:root.dark) .docs-panel-header-rose {
  border-color: rgb(55 65 81);
  background: rgb(244 63 94 / 0.08);
}
:global(:root.dark) .docs-panel-header-indigo {
  border-color: rgb(55 65 81);
  background: rgb(99 102 241 / 0.08);
}

/* 圆角 icon 盒：白底 + 色图标，参考兑换码 panel header 内 icon */
.docs-panel-icon {
  display: flex;
  flex-shrink: 0;
  height: 2.5rem;
  width: 2.5rem;
  align-items: center;
  justify-content: center;
  border-radius: 0.5rem;
  background: rgb(255 255 255);
  box-shadow: 0 1px 2px rgb(0 0 0 / 0.04);
}

:global(:root.dark) .docs-panel-icon {
  background: rgb(55 65 81);
}

.docs-panel-icon-violet { color: rgb(124 58 237); }
.docs-panel-icon-emerald { color: rgb(5 150 105); }
.docs-panel-icon-amber { color: rgb(217 119 6); }
.docs-panel-icon-rose { color: rgb(225 29 72); }
.docs-panel-icon-indigo { color: rgb(79 70 229); }

:global(:root.dark) .docs-panel-icon-violet { color: rgb(196 181 253); }
:global(:root.dark) .docs-panel-icon-emerald { color: rgb(110 231 183); }
:global(:root.dark) .docs-panel-icon-amber { color: rgb(252 211 77); }
:global(:root.dark) .docs-panel-icon-rose { color: rgb(253 164 175); }
:global(:root.dark) .docs-panel-icon-indigo { color: rgb(165 180 252); }

/* ============ 接入示例 Tab 切换 ============ */
.snippet-tabs {
  display: flex;
  gap: 0;
  overflow-x: auto;
  border-bottom: 1px solid rgb(209 250 229 / 0.7);
  background: rgb(236 253 245 / 0.6);
  padding: 0 0.75rem;
  scrollbar-width: none;
}

.snippet-tabs::-webkit-scrollbar {
  display: none;
}

:global(:root.dark) .snippet-tabs {
  border-bottom-color: rgb(55 65 81);
  background: rgb(16 185 129 / 0.08);
}

.snippet-tab {
  position: relative;
  display: inline-flex;
  align-items: center;
  gap: 0.375rem;
  white-space: nowrap;
  padding: 0.875rem 1rem;
  font-size: 0.8125rem;
  font-weight: 500;
  color: rgb(107 114 128);
  border-bottom: 2px solid transparent;
  margin-bottom: -1px;
  transition: color 0.15s, border-color 0.15s;
}

.snippet-tab:hover:not(.snippet-tab-active) {
  color: rgb(17 24 39);
}

.snippet-tab-active {
  color: rgb(4 120 87);
  border-bottom-color: rgb(16 185 129);
}

:global(:root.dark) .snippet-tab {
  color: rgb(156 163 175);
}

:global(:root.dark) .snippet-tab:hover:not(.snippet-tab-active) {
  color: rgb(255 255 255);
}

:global(:root.dark) .snippet-tab-active {
  color: rgb(167 243 208);
  border-bottom-color: rgb(52 211 153);
}
</style>
