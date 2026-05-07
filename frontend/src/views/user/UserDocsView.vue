<template>
  <AppLayout>
    <div class="mx-auto max-w-6xl space-y-6">
      <header class="space-y-4">
        <div>
          <div class="inline-flex items-center gap-2 rounded-md border border-primary-200 bg-primary-50 px-2.5 py-1 text-xs font-medium text-primary-700 dark:border-primary-800 dark:bg-primary-950/40 dark:text-primary-300">
            <Icon name="book" size="sm" />
            用户接入文档
          </div>
          <h1 class="mt-3 text-2xl font-semibold text-gray-950 dark:text-white">API 接入教程</h1>
          <p class="mt-2 max-w-3xl text-sm leading-6 text-gray-600 dark:text-gray-300">
            按这里准备密钥、选择模型和填写 Base URL，即可把本站 API 接入 Claude Code、Codex、OpenAI SDK、Gemini 兼容客户端或普通 HTTP 请求。
          </p>
        </div>

        <div class="rounded-lg border border-blue-200 bg-blue-50 px-4 py-3 text-sm leading-6 text-blue-900 dark:border-blue-800/70 dark:bg-blue-950/30 dark:text-blue-100">
          如果客户端会自动拼接 `/v1/messages`，Claude Code 这类客户端通常填写站点根地址；OpenAI 兼容 SDK 和大多数中转客户端通常填写 `/v1`。
        </div>
      </header>

      <section class="grid grid-cols-1 gap-3 md:grid-cols-3">
        <article v-for="item in quickStart" :key="item.title" class="rounded-lg border border-gray-200 bg-white p-4 dark:border-dark-700 dark:bg-dark-900">
          <div class="flex items-start gap-3">
            <span class="flex h-8 w-8 flex-shrink-0 items-center justify-center rounded-md bg-primary-600 text-sm font-semibold text-white">
              {{ item.no }}
            </span>
            <div>
              <h2 class="text-sm font-semibold text-gray-950 dark:text-white">{{ item.title }}</h2>
              <p class="mt-1 text-xs leading-5 text-gray-600 dark:text-gray-300">{{ item.desc }}</p>
              <router-link v-if="item.to" :to="item.to" class="mt-2 inline-flex items-center gap-1 text-xs font-medium text-primary-600 hover:text-primary-700 dark:text-primary-400">
                前往配置
                <Icon name="arrowRight" size="xs" />
              </router-link>
            </div>
          </div>
        </article>
      </section>

      <section class="card p-5">
        <div class="flex items-center gap-2">
          <Icon name="link" size="md" class="text-primary-600 dark:text-primary-400" />
          <h2 class="text-lg font-semibold text-gray-950 dark:text-white">本站可用地址</h2>
        </div>
        <div class="mt-4 grid grid-cols-1 gap-3 lg:grid-cols-2">
          <div v-for="endpoint in endpoints" :key="endpoint.label" class="rounded-lg border border-gray-200 bg-gray-50 p-3 dark:border-dark-700 dark:bg-dark-800/70">
            <div class="text-sm font-semibold text-gray-950 dark:text-white">{{ endpoint.label }}</div>
            <code class="mt-2 block break-all rounded-md bg-white px-3 py-2 font-mono text-xs text-gray-800 ring-1 ring-gray-200 dark:bg-dark-900 dark:text-gray-100 dark:ring-dark-700">
              {{ endpoint.url }}
            </code>
            <p class="mt-2 text-xs leading-5 text-gray-600 dark:text-gray-300">{{ endpoint.desc }}</p>
          </div>
        </div>

        <div v-if="customEndpoints.length" class="mt-4">
          <div class="text-sm font-semibold text-gray-950 dark:text-white">管理员公开的自定义端点</div>
          <div class="mt-2 grid grid-cols-1 gap-2 lg:grid-cols-2">
            <div v-for="endpoint in customEndpoints" :key="endpoint.endpoint" class="rounded-md border border-gray-200 bg-white p-3 dark:border-dark-700 dark:bg-dark-900">
              <div class="text-xs font-medium text-gray-700 dark:text-gray-200">{{ endpoint.name || '自定义端点' }}</div>
              <code class="mt-1 block break-all font-mono text-xs text-primary-700 dark:text-primary-300">{{ endpoint.endpoint }}</code>
              <p v-if="endpoint.description" class="mt-1 text-xs leading-5 text-gray-500 dark:text-gray-400">{{ endpoint.description }}</p>
            </div>
          </div>
        </div>
      </section>

      <section class="space-y-3">
        <div class="flex items-center gap-2">
          <Icon name="terminal" size="md" class="text-emerald-600 dark:text-emerald-400" />
          <h2 class="text-lg font-semibold text-gray-950 dark:text-white">复制即可改的接入示例</h2>
        </div>
        <div class="grid grid-cols-1 gap-4 xl:grid-cols-2">
          <article v-for="snippet in snippets" :key="snippet.title" class="card overflow-hidden">
            <div class="border-b border-gray-100 px-5 py-4 dark:border-dark-700">
              <div class="flex flex-wrap items-center gap-2">
                <h3 class="text-sm font-semibold text-gray-950 dark:text-white">{{ snippet.title }}</h3>
                <span class="rounded-md bg-gray-100 px-2 py-0.5 text-[11px] text-gray-600 dark:bg-dark-700 dark:text-gray-300">
                  {{ snippet.tag }}
                </span>
              </div>
              <p class="mt-1 text-xs leading-5 text-gray-500 dark:text-gray-400">{{ snippet.desc }}</p>
            </div>
            <pre class="overflow-x-auto bg-slate-950 p-4 text-xs leading-5 text-slate-100"><code>{{ snippet.code }}</code></pre>
          </article>
        </div>
      </section>

      <section class="grid grid-cols-1 gap-4 xl:grid-cols-2">
        <article class="card p-5">
          <div class="flex items-center gap-2">
            <Icon name="exclamationCircle" size="md" class="text-amber-600 dark:text-amber-400" />
            <h2 class="text-lg font-semibold text-gray-950 dark:text-white">常见错误</h2>
          </div>
          <div class="mt-4 divide-y divide-gray-100 dark:divide-dark-700">
            <div v-for="error in commonErrors" :key="error.code" class="py-3 first:pt-0 last:pb-0">
              <div class="flex flex-wrap items-center gap-2">
                <code class="rounded-md bg-gray-100 px-2 py-0.5 font-mono text-xs text-gray-700 dark:bg-dark-700 dark:text-gray-200">{{ error.code }}</code>
                <span class="text-sm font-semibold text-gray-950 dark:text-white">{{ error.title }}</span>
              </div>
              <p class="mt-1 text-xs leading-5 text-gray-600 dark:text-gray-300">{{ error.desc }}</p>
            </div>
          </div>
        </article>

        <article class="card p-5">
          <div class="flex items-center gap-2">
            <Icon name="shield" size="md" class="text-red-600 dark:text-red-400" />
            <h2 class="text-lg font-semibold text-gray-950 dark:text-white">使用安全提醒</h2>
          </div>
          <div class="mt-4 space-y-3 text-sm leading-6 text-gray-700 dark:text-gray-300">
            <p>请不要把 API 密钥发给他人，也不要放到公开仓库、前端网页或截图里。密钥泄露后，别人可以直接消耗你的额度。</p>
            <p>如果页面提示内容可能触发安全限制，请调整问题再发送。系统会保护上游账号稳定，明显违规或重复命中可能会被管理员复核。</p>
            <p>Claude Code、Codex 或 SDK 修改环境变量后，建议重启终端或应用，确认新配置已经生效。</p>
          </div>
        </article>
      </section>

      <section class="card p-5">
        <div class="flex items-center gap-2">
          <Icon name="externalLink" size="md" class="text-gray-600 dark:text-gray-300" />
          <h2 class="text-lg font-semibold text-gray-950 dark:text-white">官方参考</h2>
        </div>
        <div class="mt-3 grid grid-cols-1 gap-2 md:grid-cols-2">
          <a v-for="source in sources" :key="source.href" :href="source.href" target="_blank" rel="noopener noreferrer" class="rounded-md border border-gray-200 px-3 py-2 text-sm text-gray-700 transition-colors hover:border-primary-300 hover:text-primary-700 dark:border-dark-700 dark:text-gray-300 dark:hover:border-primary-700 dark:hover:text-primary-300">
            {{ source.label }}
          </a>
        </div>
      </section>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import { useAppStore } from '@/stores'

const appStore = useAppStore()

onMounted(() => {
  if (!appStore.publicSettingsLoaded) {
    void appStore.fetchPublicSettings()
  }
})

function normalizeUrl(url: string): string {
  return url.replace(/\/+$/, '')
}

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
