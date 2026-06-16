<template>
  <AppLayout wide>
    <div class="space-y-6">
      <!-- 锚点目录：sticky 在顶部，方便长文档快速跳转 -->
      <nav class="sticky top-2 z-10 -mx-1 flex flex-wrap items-center gap-1.5 rounded-2xl border border-gray-200/70 bg-white/80 px-3 py-2 shadow-sm backdrop-blur-md dark:border-dark-700/60 dark:bg-dark-800/70">
        <span class="px-1 text-2xs font-medium uppercase tracking-wide text-gray-400 dark:text-dark-500">目录</span>
        <a
          v-for="item in tocItems"
          :key="item.id"
          :href="`#${item.id}`"
          class="rounded-md px-2 py-1 text-xs font-medium text-gray-600 transition-colors hover:bg-gray-100 hover:text-gray-900 dark:text-dark-300 dark:hover:bg-dark-700 dark:hover:text-white"
        >
          {{ item.label }}
        </a>
      </nav>

      <!-- Base URL 速查表：最高优先级，让用户一眼对到自己客户端该填什么 -->
      <section id="base-url" class="surface-card overflow-hidden scroll-mt-20">
        <header class="flex items-start gap-3 border-b border-gray-200/60 bg-warning-soft/40 px-6 py-4 dark:border-dark-700/60 dark:bg-warning/5">
          <span class="flex h-9 w-9 flex-shrink-0 items-center justify-center rounded-lg bg-warning-soft text-warning dark:bg-warning/15 dark:text-warning-soft">
            <Icon name="exclamationCircle" size="sm" :stroke-width="1.75" />
          </span>
          <div class="min-w-0">
            <h2 class="text-base font-semibold text-gray-900 dark:text-white">Base URL 速查（必看）</h2>
            <p class="mt-0.5 text-xs text-gray-500 dark:text-dark-400">不同客户端走不同协议，Base URL 后缀差一个 /v1 就会 404 或 401</p>
          </div>
        </header>
        <!-- 桌面：表格视图（min-w 保证列宽，窄屏下父容器横向滚动） -->
        <div class="hidden overflow-x-auto sm:block">
          <table class="min-w-[720px] divide-y divide-gray-200/60 dark:divide-dark-700/60">
            <thead class="bg-gray-50/60 dark:bg-dark-800/60">
              <tr>
                <th class="px-4 py-3 text-left text-2xs font-medium text-gray-500 dark:text-dark-400">客户端</th>
                <th class="px-4 py-3 text-left text-2xs font-medium text-gray-500 dark:text-dark-400">协议</th>
                <th class="px-4 py-3 text-left text-2xs font-medium text-gray-500 dark:text-dark-400">Base URL 填法</th>
                <th class="px-4 py-3 text-left text-2xs font-medium text-gray-500 dark:text-dark-400">说明</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-gray-100 dark:divide-dark-700/60 text-sm">
              <tr v-for="row in baseUrlMatrix" :key="row.client" class="transition-colors hover:bg-gray-50/60 dark:hover:bg-dark-800/40">
                <td class="whitespace-nowrap px-4 py-3 font-semibold text-gray-900 dark:text-white">
                  <span class="inline-flex items-center gap-1.5 rounded-md px-2 py-0.5 text-xs font-medium" :class="row.tagClass">
                    {{ row.client }}
                  </span>
                </td>
                <td class="whitespace-nowrap px-4 py-3 text-xs text-gray-600 dark:text-dark-300">{{ row.protocol }}</td>
                <td class="px-4 py-3">
                  <code class="block whitespace-nowrap font-mono text-xs" :class="row.urlClass">{{ row.url }}</code>
                </td>
                <td class="px-4 py-3 text-xs text-gray-500 dark:text-dark-400">{{ row.note }}</td>
              </tr>
            </tbody>
          </table>
        </div>
        <!-- 移动端：堆叠卡片视图，每条占一整行，避免表格挤成竖字 -->
        <div class="divide-y divide-gray-100 dark:divide-dark-700/60 sm:hidden">
          <div v-for="row in baseUrlMatrix" :key="row.client" class="space-y-2 p-4">
            <div class="flex items-center justify-between gap-2">
              <span class="inline-flex items-center gap-1.5 rounded-md px-2 py-0.5 text-xs font-medium" :class="row.tagClass">
                {{ row.client }}
              </span>
              <span class="text-2xs text-gray-500 dark:text-dark-400">{{ row.protocol }}</span>
            </div>
            <code class="block overflow-x-auto rounded-md bg-gray-50 px-2 py-1.5 font-mono text-xs dark:bg-dark-800" :class="row.urlClass">{{ row.url }}</code>
            <p class="text-xs leading-5 text-gray-500 dark:text-dark-400">{{ row.note }}</p>
          </div>
        </div>
      </section>

      <!-- amber 警示条：终端配置避坑（合并去重） -->
      <div class="flex items-start gap-2.5 rounded-xl border border-warning/30 bg-warning-soft/60 px-4 py-3 dark:border-warning/20 dark:bg-warning/5">
        <Icon name="exclamationCircle" size="sm" class="mt-0.5 flex-shrink-0 text-warning dark:text-warning-soft" />
        <p class="text-sm leading-6 text-warning-deep dark:text-warning-soft">
          <span class="font-semibold">终端配置避坑：</span>
          直接复制 <code class="docs-hint-code">VAR=value</code> 这种写法到终端可能报"命令不存在"。
          macOS/Linux 必须用 <code class="docs-hint-code">export VAR="value"</code>；
          Windows PowerShell 用 <code class="docs-hint-code">$env:VAR = "value"</code>；
          Windows CMD 用 <code class="docs-hint-code">set VAR=value</code>。
          下方每个客户端代码示例都按系统分块给出，按你的系统复制对应那段。
        </p>
      </div>

      <!-- 快速开始 3 步 -->
      <section id="quick-start" class="space-y-4 scroll-mt-20">
        <div class="flex items-center gap-2.5">
          <span class="flex h-7 w-7 items-center justify-center rounded-lg bg-success-soft text-success dark:bg-success/15 dark:text-success-soft">
            <Icon name="checkCircle" size="sm" :stroke-width="1.75" />
          </span>
          <h2 class="text-base font-semibold text-gray-900 dark:text-white">快速开始</h2>
        </div>
        <div class="grid grid-cols-1 gap-3 md:grid-cols-3">
          <article
            v-for="item in quickStart"
            :key="item.title"
            class="surface-card group p-5 transition-all hover:-translate-y-0.5 hover:shadow-card-hover"
          >
            <div class="flex items-start gap-3">
              <span class="flex h-8 w-8 flex-shrink-0 items-center justify-center rounded-lg bg-gray-900 text-sm font-semibold text-white dark:bg-white dark:text-gray-900">{{ item.no }}</span>
              <div class="min-w-0">
                <h3 class="text-sm font-semibold text-gray-900 dark:text-white">{{ item.title }}</h3>
                <p class="mt-1.5 text-xs leading-5 text-gray-600 dark:text-dark-300">{{ item.desc }}</p>
                <router-link
                  v-if="item.to"
                  :to="item.to"
                  class="mt-2 inline-flex items-center gap-1 text-xs font-medium text-brand-600 transition-colors hover:text-brand-700 dark:text-brand-400 dark:hover:text-brand-300"
                >
                  前往配置
                  <Icon name="arrowRight" size="xs" />
                </router-link>
              </div>
            </div>
          </article>
        </div>
      </section>

      <!-- 管理员自定义端点（仅当后台公开了额外端点时显示；标准 URL 已在速查表里） -->
      <section v-if="customEndpoints.length" class="surface-card overflow-hidden">
        <header class="flex items-start gap-3 border-b border-gray-200/60 px-6 py-4 dark:border-dark-700/60">
          <span class="flex h-9 w-9 flex-shrink-0 items-center justify-center rounded-lg bg-info-soft text-info dark:bg-info/15 dark:text-info-soft">
            <Icon name="link" size="sm" :stroke-width="1.75" />
          </span>
          <div class="min-w-0">
            <h2 class="text-base font-semibold text-gray-900 dark:text-white">管理员自定义端点</h2>
            <p class="mt-0.5 text-xs text-gray-500 dark:text-dark-400">管理员额外公开的访问入口</p>
          </div>
        </header>
        <div class="p-6">
          <div class="grid grid-cols-1 gap-2 lg:grid-cols-2">
            <div
              v-for="endpoint in customEndpoints"
              :key="endpoint.endpoint"
              class="group rounded-xl border border-gray-200/70 bg-white p-3 transition-colors hover:border-gray-300 dark:border-dark-700/60 dark:bg-dark-900 dark:hover:border-dark-600"
            >
              <div class="flex items-center justify-between gap-2">
                <div class="text-xs font-medium text-gray-700 dark:text-dark-200">{{ endpoint.name || '自定义端点' }}</div>
                <button
                  type="button"
                  class="rounded-md p-1 text-gray-400 opacity-0 transition-all hover:bg-gray-100 hover:text-gray-700 group-hover:opacity-100 dark:hover:bg-dark-700 dark:hover:text-gray-200"
                  :title="copiedEndpoint === endpoint.endpoint ? '已复制' : '复制地址'"
                  @click="copyEndpoint(endpoint.endpoint)"
                >
                  <Icon :name="copiedEndpoint === endpoint.endpoint ? 'check' : 'copy'" size="xs" />
                </button>
              </div>
              <code class="mt-1 block break-all font-mono text-xs text-info dark:text-info-soft">{{ endpoint.endpoint }}</code>
              <p v-if="endpoint.description" class="mt-1 text-xs leading-5 text-gray-500 dark:text-dark-400">{{ endpoint.description }}</p>
            </div>
          </div>
        </div>
      </section>

      <!-- 接入示例：tab 切换 + 代码块 + 复制 -->
      <section id="snippets" class="space-y-4 scroll-mt-20">
        <div class="flex items-center gap-2.5">
          <span class="flex h-7 w-7 items-center justify-center rounded-lg bg-success-soft text-success dark:bg-success/15 dark:text-success-soft">
            <Icon name="terminal" size="sm" :stroke-width="1.75" />
          </span>
          <h2 class="text-base font-semibold text-gray-900 dark:text-white">客户端接入示例</h2>
          <span class="rounded-full bg-gray-100 px-2 py-0.5 text-2xs font-medium text-gray-600 dark:bg-dark-700 dark:text-dark-300">
            {{ snippets.length }} 种客户端
          </span>
        </div>

        <article class="surface-card overflow-hidden">
          <!-- Tab 导航：横向滚动以容纳更多客户端 -->
          <div class="snippet-tabs">
            <button
              v-for="(snippet, idx) in snippets"
              :key="snippet.title"
              type="button"
              :class="['snippet-tab', activeSnippetIdx === idx && 'snippet-tab-active']"
              @click="activeSnippetIdx = idx"
            >
              {{ snippet.title }}
            </button>
          </div>

          <template v-if="activeSnippet">
            <div class="px-6 pb-3 pt-4">
              <div class="flex flex-wrap items-center gap-2">
                <h3 class="text-sm font-semibold text-gray-900 dark:text-white">{{ activeSnippet.title }}</h3>
                <span class="rounded-md bg-success-soft px-2 py-0.5 text-2xs font-medium text-success dark:bg-success/15 dark:text-success-soft">
                  {{ activeSnippet.tag }}
                </span>
              </div>
              <p class="mt-1.5 text-xs leading-5 text-gray-500 dark:text-dark-400">{{ activeSnippet.desc }}</p>
            </div>
            <div class="relative">
              <!-- 顶部工具栏：左侧 language 标签 + 右侧复制按钮 -->
              <div class="absolute left-3 top-3 z-10 inline-flex items-center gap-1 rounded-md bg-dark-800/80 px-2 py-1 text-2xs font-medium uppercase tracking-wider text-gray-300 backdrop-blur">
                <span class="h-1.5 w-1.5 rounded-full" :class="langDotClass(activeSnippet.title)"></span>
                {{ snippetLang(activeSnippet.title) }}
              </div>
              <button
                type="button"
                class="absolute right-3 top-3 z-10 inline-flex items-center gap-1.5 rounded-md bg-dark-800/80 px-2.5 py-1.5 text-xs font-medium text-gray-100 backdrop-blur transition-colors hover:bg-dark-700/90"
                @click="copySnippet(activeSnippet)"
              >
                <Icon :name="copiedSnippetTitle === activeSnippet.title ? 'check' : 'copy'" size="xs" />
                {{ copiedSnippetTitle === activeSnippet.title ? '已复制' : '复制' }}
              </button>
              <pre class="snippet-pre overflow-x-auto bg-dark-950 p-5 pt-12 text-xs leading-6 text-gray-100 sm:p-6 sm:pt-12"><code v-html="highlightCode(activeSnippet.code)"></code></pre>
            </div>
          </template>
        </article>
      </section>

      <!-- 模型选择指南：场景 → 推荐模型映射 -->
      <section id="models" class="surface-card overflow-hidden scroll-mt-20">
        <header class="flex items-start gap-3 border-b border-gray-200/60 px-6 py-4 dark:border-dark-700/60">
          <span class="flex h-9 w-9 flex-shrink-0 items-center justify-center rounded-lg bg-brand-50 text-brand-600 dark:bg-brand-500/15 dark:text-brand-300">
            <Icon name="sparkles" size="sm" :stroke-width="1.75" />
          </span>
          <div class="min-w-0">
            <h2 class="text-base font-semibold text-gray-900 dark:text-white">模型选择指南</h2>
            <p class="mt-0.5 text-xs text-gray-500 dark:text-dark-400">按使用场景选模型；具体模型名以后台分组配置为准</p>
          </div>
        </header>
        <div class="p-6">
          <div class="grid grid-cols-1 gap-3 md:grid-cols-2">
            <div
              v-for="guide in modelGuide"
              :key="guide.scene"
              class="rounded-xl border border-gray-200/70 bg-white p-4 dark:border-dark-700/60 dark:bg-dark-900"
            >
              <div class="flex items-start gap-2.5">
                <span class="flex h-8 w-8 flex-shrink-0 items-center justify-center rounded-lg" :class="guide.iconClass">
                  <Icon :name="(guide.icon as any)" size="sm" :stroke-width="1.75" />
                </span>
                <div class="min-w-0">
                  <div class="text-sm font-semibold text-gray-900 dark:text-white">{{ guide.scene }}</div>
                  <p class="mt-0.5 text-xs leading-5 text-gray-500 dark:text-dark-400">{{ guide.desc }}</p>
                </div>
              </div>
              <div class="mt-3 flex flex-wrap gap-1.5">
                <code
                  v-for="m in guide.models"
                  :key="m"
                  class="rounded-md bg-gray-100 px-2 py-0.5 font-mono text-2xs text-gray-700 dark:bg-dark-800 dark:text-gray-300"
                >{{ m }}</code>
              </div>
            </div>
          </div>
        </div>
      </section>

      <!-- 高级用法：流式、图像、function calling、缓存 -->
      <section id="advanced" class="surface-card overflow-hidden scroll-mt-20">
        <header class="flex items-start gap-3 border-b border-gray-200/60 px-6 py-4 dark:border-dark-700/60">
          <span class="flex h-9 w-9 flex-shrink-0 items-center justify-center rounded-lg bg-info-soft text-info dark:bg-info/15 dark:text-info-soft">
            <Icon name="bolt" size="sm" :stroke-width="1.75" />
          </span>
          <div class="min-w-0">
            <h2 class="text-base font-semibold text-gray-900 dark:text-white">高级用法</h2>
            <p class="mt-0.5 text-xs text-gray-500 dark:text-dark-400">流式、图像、工具调用、提示词缓存</p>
          </div>
        </header>
        <div class="divide-y divide-gray-100 dark:divide-dark-700/60">
          <div v-for="feature in advancedFeatures" :key="feature.title" class="px-6 py-4">
            <div class="flex items-start gap-2.5">
              <Icon :name="(feature.icon as any)" size="sm" class="mt-0.5 flex-shrink-0 text-gray-500 dark:text-dark-400" />
              <div class="min-w-0 flex-1">
                <h3 class="text-sm font-semibold text-gray-900 dark:text-white">{{ feature.title }}</h3>
                <p class="mt-1 text-xs leading-5 text-gray-600 dark:text-dark-300">{{ feature.desc }}</p>
                <details v-if="feature.example" class="mt-2 group">
                  <summary class="cursor-pointer text-xs font-medium text-brand-600 hover:text-brand-700 dark:text-brand-400 dark:hover:text-brand-300 list-none [&::-webkit-details-marker]:hidden inline-flex items-center gap-1">
                    <span>查看示例</span>
                    <Icon name="chevronDown" size="xs" class="transition-transform group-open:rotate-180" />
                  </summary>
                  <pre class="snippet-pre mt-2 overflow-x-auto rounded-lg bg-dark-950 p-4 text-2xs leading-5 text-gray-100"><code v-html="highlightCode(feature.example)"></code></pre>
                </details>
              </div>
            </div>
          </div>
        </div>
      </section>

      <!-- 下游对接：把 API 提供给第三方时该交付什么 -->
      <section id="downstream" class="surface-card overflow-hidden scroll-mt-20">
        <header class="flex items-start gap-3 border-b border-gray-200/60 bg-info-soft/40 px-6 py-4 dark:border-dark-700/60 dark:bg-info/5">
          <span class="flex h-9 w-9 flex-shrink-0 items-center justify-center rounded-lg bg-info-soft text-info dark:bg-info/15 dark:text-info-soft">
            <Icon name="link" size="sm" :stroke-width="1.75" />
          </span>
          <div class="min-w-0 flex-1">
            <h2 class="text-base font-semibold text-gray-900 dark:text-white">下游对接 API</h2>
            <p class="mt-0.5 text-xs text-gray-500 dark:text-dark-400">把 API 提供给客户、代理商或第三方系统时，参考此节交付配置</p>
          </div>
          <div class="flex shrink-0 gap-2">
            <a
              :href="downstreamDocUrl"
              target="_blank"
              rel="noopener noreferrer"
              class="inline-flex items-center gap-1.5 rounded-lg border border-info/40 bg-white px-3 py-1.5 text-xs font-semibold text-info transition-colors hover:bg-info-soft dark:border-info/40 dark:bg-dark-800 dark:text-info-soft dark:hover:bg-info/10"
            >
              <Icon name="externalLink" size="xs" />
              <span>完整文档</span>
            </a>
            <button
              type="button"
              class="inline-flex items-center gap-1.5 rounded-lg border border-gray-200/70 bg-white px-3 py-1.5 text-xs font-semibold text-gray-700 transition-colors hover:bg-gray-50 dark:border-dark-700/60 dark:bg-dark-800 dark:text-dark-200 dark:hover:bg-dark-700/60"
              @click="copyDownstreamConfig"
            >
              <Icon name="copy" size="xs" />
              <span>复制最小配置</span>
            </button>
          </div>
        </header>

        <div class="space-y-5 p-6">
          <!-- 最小配置块：交付给下游的标准卡片 -->
          <div class="rounded-xl border border-info/20 bg-info-soft/40 p-4 dark:border-info/20 dark:bg-info/5">
            <div class="mb-2 flex items-center gap-2">
              <Icon name="cube" size="sm" class="text-info dark:text-info-soft" />
              <span class="text-sm font-semibold text-gray-900 dark:text-white">最小交付配置（推荐 OpenAI 兼容路径）</span>
            </div>
            <pre class="snippet-pre overflow-x-auto rounded-lg bg-dark-950 p-3 text-2xs leading-5 text-gray-100"><code>{{ downstreamMinimalConfig }}</code></pre>
          </div>

          <!-- 关键提醒：四点高频踩坑 -->
          <div class="grid grid-cols-1 gap-3 md:grid-cols-2">
            <article
              v-for="tip in downstreamTips"
              :key="tip.title"
              class="rounded-xl border border-gray-200/70 bg-white p-3.5 dark:border-dark-700/60 dark:bg-dark-800/40"
            >
              <div class="flex items-start gap-2.5">
                <span :class="['flex h-7 w-7 flex-shrink-0 items-center justify-center rounded-lg', tip.iconClass]">
                  <Icon :name="(tip.icon as any)" size="xs" :stroke-width="1.75" />
                </span>
                <div class="min-w-0">
                  <h3 class="text-sm font-semibold text-gray-900 dark:text-white">{{ tip.title }}</h3>
                  <p class="mt-1 text-xs leading-5 text-gray-600 dark:text-dark-300">{{ tip.desc }}</p>
                </div>
              </div>
            </article>
          </div>

          <!-- 提示：完整文档里包含的内容 -->
          <div class="rounded-xl border border-info/20 bg-info-soft/40 p-3.5 dark:border-info/20 dark:bg-info/5">
            <div class="flex items-start gap-2.5">
              <Icon name="infoCircle" size="sm" class="mt-0.5 flex-shrink-0 text-info dark:text-info-soft" />
              <p class="text-xs leading-5 text-info-deep dark:text-info-soft">
                <span class="font-semibold">完整文档包含：</span>
                Base URL 速查、鉴权方式、所有支持的接口路径、模型查询、计费与限流详解、<code class="docs-hint-code">/v1/usage</code> 返回结构、错误码表、CC Switch/Codex/Gemini/Claude 客户端对接示例、上线检查清单。把链接整体发给下游即可。
              </p>
            </div>
          </div>
        </div>
      </section>

      <!-- 常见错误 + 使用提醒（lg 起并排，原来 xl 太晚） -->
      <section id="troubleshoot" class="grid grid-cols-1 gap-4 lg:grid-cols-2 scroll-mt-20">
        <article class="surface-card overflow-hidden">
          <header class="flex items-start gap-3 border-b border-gray-200/60 px-6 py-4 dark:border-dark-700/60">
            <span class="flex h-9 w-9 flex-shrink-0 items-center justify-center rounded-lg bg-warning-soft text-warning dark:bg-warning/15 dark:text-warning-soft">
              <Icon name="exclamationCircle" size="sm" :stroke-width="1.75" />
            </span>
            <div class="min-w-0">
              <h2 class="text-base font-semibold text-gray-900 dark:text-white">常见错误</h2>
              <p class="mt-0.5 text-xs text-gray-500 dark:text-dark-400">遇到这些状态码时的排查方向</p>
            </div>
          </header>
          <div class="divide-y divide-gray-100 dark:divide-dark-700/60">
            <div v-for="error in commonErrors" :key="error.code" class="px-6 py-3">
              <div class="flex flex-wrap items-center gap-2">
                <code class="rounded-md bg-warning-soft px-2 py-0.5 font-mono text-xs font-medium text-warning dark:bg-warning/15 dark:text-warning-soft">{{ error.code }}</code>
                <span class="text-sm font-semibold text-gray-900 dark:text-white">{{ error.title }}</span>
              </div>
              <p class="mt-1 text-xs leading-5 text-gray-600 dark:text-dark-300">{{ error.desc }}</p>
            </div>
          </div>
        </article>

        <article class="surface-card overflow-hidden">
          <header class="flex items-start gap-3 border-b border-gray-200/60 px-6 py-4 dark:border-dark-700/60">
            <span class="flex h-9 w-9 flex-shrink-0 items-center justify-center rounded-lg bg-danger-soft text-danger dark:bg-danger/15 dark:text-danger-soft">
              <Icon name="shield" size="sm" :stroke-width="1.75" />
            </span>
            <div class="min-w-0">
              <h2 class="text-base font-semibold text-gray-900 dark:text-white">使用提醒</h2>
              <p class="mt-0.5 text-xs text-gray-500 dark:text-dark-400">保护密钥，避免触发风控</p>
            </div>
          </header>
          <ul class="space-y-3 px-6 py-4 text-sm leading-6 text-gray-700 dark:text-dark-200">
            <li v-for="(tip, i) in securityTips" :key="i" class="flex items-start gap-2.5">
              <Icon name="checkCircle" size="xs" class="mt-1.5 flex-shrink-0 text-success dark:text-success-soft" />
              <span>{{ tip }}</span>
            </li>
          </ul>
        </article>
      </section>

      <!-- 官方参考 -->
      <section id="sources" class="surface-card overflow-hidden scroll-mt-20">
        <header class="flex items-start gap-3 border-b border-gray-200/60 px-6 py-4 dark:border-dark-700/60">
          <span class="flex h-9 w-9 flex-shrink-0 items-center justify-center rounded-lg bg-info-soft text-info dark:bg-info/15 dark:text-info-soft">
            <Icon name="externalLink" size="sm" :stroke-width="1.75" />
          </span>
          <div class="min-w-0">
            <h2 class="text-base font-semibold text-gray-900 dark:text-white">官方参考</h2>
            <p class="mt-0.5 text-xs text-gray-500 dark:text-dark-400">遇到细节问题可查阅各家官方文档</p>
          </div>
        </header>
        <div class="p-6">
          <div class="grid grid-cols-1 gap-2 md:grid-cols-2">
            <a
              v-for="source in sources"
              :key="source.href"
              :href="source.href"
              target="_blank"
              rel="noopener noreferrer"
              class="flex items-center justify-between rounded-xl border border-gray-200/70 px-3 py-2 text-sm text-gray-700 transition-colors hover:border-brand-300 hover:text-brand-700 dark:border-dark-700/60 dark:text-dark-200 dark:hover:border-brand-500/50 dark:hover:text-brand-300"
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
import { computed, onMounted, ref, watch } from 'vue'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import { useAppStore } from '@/stores'

const appStore = useAppStore()

// 锚点目录：用于顶部 sticky 导航跳转
const tocItems = [
  { id: 'base-url', label: 'Base URL' },
  { id: 'quick-start', label: '快速开始' },
  { id: 'snippets', label: '接入示例' },
  { id: 'models', label: '模型选择' },
  { id: 'advanced', label: '高级用法' },
  { id: 'downstream', label: '下游对接' },
  { id: 'troubleshoot', label: '排障' },
  { id: 'sources', label: '官方参考' }
]

// 客户端 → 代码块语言标签映射，让用户复制后知道该粘到哪里
const SNIPPET_LANG_MAP: Record<string, string> = {
  'Claude Code': 'bash + json',
  'CC Switch / Codex': 'GUI + json + toml',
  'Codex CLI': 'bash + toml',
  'Cursor': 'GUI 配置',
  'Cline': 'GUI 配置',
  'Cherry Studio': 'GUI 配置',
  'Continue.dev': 'json',
  'JetBrains AI Assistant': 'GUI 配置',
  'LobeChat': 'GUI / env',
  'OpenAI SDK': 'js + python',
  'HTTP / curl': 'shell',
}
function snippetLang(title: string): string {
  return SNIPPET_LANG_MAP[title] || 'code'
}
function langDotClass(title: string): string {
  const lang = snippetLang(title)
  if (lang.includes('json')) return 'bg-warning'
  if (lang.includes('bash') || lang.includes('shell')) return 'bg-success'
  if (lang.includes('toml')) return 'bg-brand-400'
  if (lang.includes('python')) return 'bg-info'
  if (lang.includes('GUI')) return 'bg-danger'
  return 'bg-gray-400'
}

// 接入示例当前激活的 tab 与"已复制"反馈状态（持久化到 localStorage 避免刷新丢失）
const ACTIVE_SNIPPET_KEY = 'user-docs-active-snippet'
const activeSnippetIdx = ref(Number(localStorage.getItem(ACTIVE_SNIPPET_KEY) ?? '0') || 0)
watch(activeSnippetIdx, (v) => {
  try { localStorage.setItem(ACTIVE_SNIPPET_KEY, String(v)) } catch { /* ignore quota */ }
})
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
// codexDirectBaseUrl 已不再单独展示（合并进 baseUrlMatrix）；如未来需要 Codex 直连入口，可参考 backend-api/codex 路径
const customEndpoints = computed(() => appStore.cachedPublicSettings?.custom_endpoints || [])

// 下游对接：完整文档（位于 public/docs/）+ 最小配置 + 高频提醒
const downstreamDocUrl = '/docs/downstream-api-integration.md'

const downstreamMinimalConfig = computed(() => `Base URL: ${openAIBaseUrl.value}
API Key: sk-xxxxxxxx
鉴权方式: Authorization: Bearer sk-xxxxxxxx
推荐接口: /v1/responses 或 /v1/chat/completions
模型名查询: GET ${openAIBaseUrl.value}/models
用量查询: GET ${openAIBaseUrl.value}/usage
流式响应: 请求体加 "stream": true`)

const downstreamTips = [
  {
    title: '鉴权放请求头',
    icon: 'shield',
    iconClass: 'bg-danger-soft text-danger dark:bg-danger/15 dark:text-danger-soft',
    desc: '用 Authorization: Bearer ＿。把 Key 放到 URL 查询参数（?key= 或 ?api_key=）会被网关直接 400 拒绝。',
  },
  {
    title: '模型名先查再填',
    icon: 'search',
    iconClass: 'bg-success-soft text-success dark:bg-success/15 dark:text-success-soft',
    desc: '不要凭印象写模型名。调 GET /v1/models 返回的 data[].id 才是分组实际开放的模型。',
  },
  {
    title: '429 不一定是限速',
    icon: 'exclamationCircle',
    iconClass: 'bg-warning-soft text-warning dark:bg-warning/15 dark:text-warning-soft',
    desc: '可能是余额不足、套餐额度耗尽、Key 配额耗尽、并发超限、RPM 超限五种之一。先看 /v1/usage 自检。',
  },
  {
    title: '流式响应调大超时',
    icon: 'clock',
    iconClass: 'bg-info-soft text-info dark:bg-info/15 dark:text-info-soft',
    desc: '代码生成或 Agent 场景建议 HTTP 读取超时 10–60 分钟，并确认 Nginx/Cloudflare 等代理层没有更短的超时设置。',
  },
]

async function copyDownstreamConfig(): Promise<void> {
  try {
    await navigator.clipboard.writeText(downstreamMinimalConfig.value)
    appStore.showSuccess('已复制最小配置')
  } catch {
    appStore.showError('复制失败，请手动选中文本')
  }
}

// Base URL 速查表：客户端 ↔ 应该填的 Base URL 映射
const baseUrlMatrix = computed(() => [
  {
    client: 'Claude Code',
    tagClass: 'bg-warning-soft text-warning dark:bg-warning/15 dark:text-warning-soft',
    protocol: 'Anthropic Messages',
    url: siteRootUrl.value,
    urlClass: 'text-success dark:text-success-soft',
    note: '不带 /v1，客户端会自动拼 /v1/messages'
  },
  {
    client: 'Cursor',
    tagClass: 'bg-info-soft text-info dark:bg-info/15 dark:text-info-soft',
    protocol: 'OpenAI Chat',
    url: openAIBaseUrl.value,
    urlClass: 'text-success dark:text-success-soft',
    note: '带 /v1，Override OpenAI Base URL 字段'
  },
  {
    client: 'OpenAI SDK',
    tagClass: 'bg-info-soft text-info dark:bg-info/15 dark:text-info-soft',
    protocol: 'OpenAI Chat',
    url: openAIBaseUrl.value,
    urlClass: 'text-success dark:text-success-soft',
    note: 'baseURL / base_url 参数填这个'
  },
  {
    client: 'Cline',
    tagClass: 'bg-info-soft text-info dark:bg-info/15 dark:text-info-soft',
    protocol: 'OpenAI Chat',
    url: openAIBaseUrl.value,
    urlClass: 'text-success dark:text-success-soft',
    note: 'OpenAI Compatible Provider 的 Base URL'
  },
  {
    client: 'Cherry Studio',
    tagClass: 'bg-info-soft text-info dark:bg-info/15 dark:text-info-soft',
    protocol: 'OpenAI Chat',
    url: openAIBaseUrl.value,
    urlClass: 'text-success dark:text-success-soft',
    note: 'API 地址字段（OpenAI 服务商）'
  },
  {
    client: 'Continue.dev',
    tagClass: 'bg-info-soft text-info dark:bg-info/15 dark:text-info-soft',
    protocol: 'OpenAI Chat',
    url: openAIBaseUrl.value,
    urlClass: 'text-success dark:text-success-soft',
    note: 'apiBase 字段'
  },
  {
    client: 'JetBrains AI',
    tagClass: 'bg-info-soft text-info dark:bg-info/15 dark:text-info-soft',
    protocol: 'OpenAI Chat',
    url: openAIBaseUrl.value,
    urlClass: 'text-success dark:text-success-soft',
    note: 'Custom Model URL 字段'
  },
  {
    client: 'LobeChat',
    tagClass: 'bg-info-soft text-info dark:bg-info/15 dark:text-info-soft',
    protocol: 'OpenAI Chat',
    url: openAIBaseUrl.value,
    urlClass: 'text-success dark:text-success-soft',
    note: 'API 代理地址（OpenAI Provider）'
  },
  {
    client: 'Codex CLI',
    tagClass: 'bg-success-soft text-success dark:bg-success/15 dark:text-success-soft',
    protocol: 'OpenAI Responses',
    url: openAIBaseUrl.value,
    urlClass: 'text-success dark:text-success-soft',
    note: 'config.toml 的 base_url；wire_api = "responses"'
  },
  {
    client: 'CC Switch / Codex',
    tagClass: 'bg-success-soft text-success dark:bg-success/15 dark:text-success-soft',
    protocol: 'OpenAI Responses',
    url: openAIBaseUrl.value,
    urlClass: 'text-success dark:text-success-soft',
    note: '供应商 API 地址填 /v1，config.toml 必须有 model_providers 块'
  },
  {
    client: 'Gemini SDK',
    tagClass: 'bg-brand-50 text-brand-700 dark:bg-brand-500/15 dark:text-brand-300',
    protocol: 'Gemini Native',
    url: geminiBaseUrl.value,
    urlClass: 'text-success dark:text-success-soft',
    note: 'Gemini 兼容 SDK 的 base URL'
  }
])

const quickStart = [
  { no: '1', title: '创建 API 密钥', desc: '进入 API 密钥页面创建密钥，复制以 sk- 开头的 Key，并确认密钥有可访问分组。', to: '/keys' },
  { no: '2', title: '确认可用模型', desc: '进入可用渠道或模型广场，查看你的分组能使用哪些模型，以及模型当前是否可用。', to: '/available-channels' },
  { no: '3', title: '填写 Base URL', desc: '根据客户端类型填写根地址、/v1、/v1beta 或 Codex 专用路径，填错路径通常会 404。', to: '' },
]

// endpoints 数据已合并到 baseUrlMatrix 速查表里，旧字段保留给可能的外部引用但前端不再渲染

const snippets = computed(() => [
  {
    title: 'Claude Code',
    tag: 'Anthropic',
    desc: '按你的系统选对应代码块复制到终端执行，注意 export 关键字不能漏。一次性环境变量只对当前终端窗口有效，关掉终端就失效。',
    code: [
      '# ========== macOS / Linux (bash / zsh) ==========',
      `export ANTHROPIC_BASE_URL="${siteRootUrl.value}"`,
      'export ANTHROPIC_AUTH_TOKEN="sk-你的密钥"',
      'claude   # 启动 Claude Code',
      '',
      '# ========== Windows PowerShell ==========',
      `$env:ANTHROPIC_BASE_URL = "${siteRootUrl.value}"`,
      '$env:ANTHROPIC_AUTH_TOKEN = "sk-你的密钥"',
      'claude   # 启动 Claude Code',
      '',
      '# ========== Windows CMD ==========',
      `set ANTHROPIC_BASE_URL=${siteRootUrl.value}`,
      'set ANTHROPIC_AUTH_TOKEN=sk-你的密钥',
      'claude',
      '',
      '# ========== 永久配置（推荐）==========',
      '# 方式 1：写入 shell 配置文件',
      '#   bash:  ~/.bashrc',
      '#   zsh:   ~/.zshrc',
      '#   把上面 export 两行加进去后重启终端',
      '',
      '# 方式 2：Claude Code 自带 settings.json',
      '# 路径：~/.claude/settings.json （Windows 是 %USERPROFILE%\\.claude\\settings.json）',
      '{',
      '  "env": {',
      `    "ANTHROPIC_BASE_URL": "${siteRootUrl.value}",`,
      '    "ANTHROPIC_AUTH_TOKEN": "sk-你的密钥"',
      '  }',
      '}',
    ].join('\n'),
  },
  {
    title: 'Codex CLI',
    tag: 'OpenAI 兼容',
    desc: '环境变量法跟 Claude Code 一样要分平台写 export / $env / set。推荐用 ~/.codex/config.toml 永久配置 provider，避免每次重开终端都要重新设置。',
    code: [
      '# ========== 临时环境变量（仅当前终端窗口）==========',
      '# macOS / Linux',
      'export OPENAI_API_KEY="sk-你的密钥"',
      `export OPENAI_BASE_URL="${openAIBaseUrl.value}"`,
      'codex   # 启动 Codex',
      '',
      '# Windows PowerShell',
      '$env:OPENAI_API_KEY = "sk-你的密钥"',
      `$env:OPENAI_BASE_URL = "${openAIBaseUrl.value}"`,
      'codex',
      '',
      '# ========== 永久配置：~/.codex/config.toml（推荐）==========',
      '[model_providers.sub2api]',
      'name = "Sub2API"',
      `base_url = "${openAIBaseUrl.value}"`,
      'env_key = "OPENAI_API_KEY"',
      'wire_api = "responses"',
      '',
      'model_provider = "sub2api"',
      '',
      '# 然后还是需要导出 OPENAI_API_KEY（密钥不能直接写 toml 避免明文泄露）',
    ].join('\n'),
  },
  {
    title: 'CC Switch / Codex',
    tag: '图形界面',
    desc: '适合让客户在 CC Switch 图形界面里新增 Codex 供应商。最关键的是 config.toml 不能只写顶层 base_url，必须通过 model_provider 指向 model_providers.OpenAI，并设置 wire_api = "responses"。',
    code: [
      '# ========== CC Switch 基本字段 ==========',
      '供应商名称：任意，例如 velorix',
      `API 请求地址：${openAIBaseUrl.value}`,
      'API Key：sk-你的密钥',
      '模型名称：gpt-5.4（以 /v1/models 实际返回为准）',
      '',
      '# ========== auth.json ==========',
      '{',
      '  "OPENAI_API_KEY": "sk-你的密钥",',
      '  "auth_mode": "apikey"',
      '}',
      '',
      '# ========== config.toml ==========',
      '# 这几行顶层配置要放在 [projects]、[windows]、[plugins] 等表头之前',
      'model_provider = "OpenAI"',
      'model = "gpt-5.4"',
      'model_reasoning_effort = "xhigh"',
      'disable_response_storage = true',
      '',
      '[model_providers.OpenAI]',
      'name = "OpenAI"',
      `base_url = "${openAIBaseUrl.value}"`,
      'wire_api = "responses"',
      'requires_openai_auth = true',
      '',
      '# ========== 常见错误 ==========',
      '# 错误：只在顶层写 base_url',
      '# 结果：Codex 可能继续请求 api.openai.com，出现 401 或连接失败',
      '# 正确：base_url 必须写在 [model_providers.OpenAI] 块里',
      '',
      '# 错误：缺少 wire_api = "responses"',
      '# 结果：Codex 可能走 /chat/completions，检查失败或接口不匹配',
      '',
      '# ========== 验证 ==========',
      `curl ${openAIBaseUrl.value}/models -H "Authorization: Bearer sk-你的密钥"`,
    ].join('\n'),
  },
  {
    title: 'Cursor',
    tag: 'OpenAI 兼容',
    desc: 'Cursor 设置全程 GUI 操作。注意：开启 OpenAI Base URL 后必须关掉 Anthropic API Key，否则 claude-* 模型会走 Cursor 默认 Anthropic 通道而绕过本站。',
    code: [
      '# Cursor 配置步骤（GUI）',
      '# 1. 打开 Settings → Models',
      '',
      '# 2. OpenAI API Key 区域',
      `   - Override OpenAI Base URL → 开启，填：${openAIBaseUrl.value}`,
      '   - OpenAI API Key → 填：sk-你的密钥',
      '   - 点击 Verify 验证（看到 success 才算成功）',
      '',
      '# 3. Anthropic API Key 区域',
      '   - 务必关闭（重要避坑点）',
      '   - 否则 claude-* 模型会绕过本站直连官方',
      '',
      '# 4. Add Model：添加你后台分组里实际可用的模型名',
      '   - 例：claude-sonnet-4-5、gpt-4o、gemini-2.0-flash',
      '   - 关掉 Cursor 自带的同名模型避免冲突',
      '',
      '# 5. 在 Chat（Ctrl+L）或 Inline Edit（Ctrl+K）选择该模型测试',
      '',
      '# 已知限制',
      '#   ✓ Chat 对话、Inline Edit 可用',
      '#   ✗ Tab 自动补全（走 cursor-small 内部模型，不可替换）',
      '#   ✗ Apply 应用代码（走 Cursor 内部模型）',
      '#   ⚠ Agent / Composer：取决于模型 function calling 支持',
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
    title: 'Cline',
    tag: 'VSCode 编码 Agent',
    desc: 'Cline 是开源 VSCode 编码 Agent，支持 OpenAI Compatible Provider。安装扩展后右下角设置图标进入配置。',
    code: [
      '# Cline 配置（VSCode 扩展）',
      '# 1. 安装：VSCode 扩展商店搜 "Cline"',
      '# 2. 点击 Cline 右上角设置图标',
      '',
      '# API Provider 选 OpenAI Compatible',
      `Base URL:     ${openAIBaseUrl.value}`,
      'API Key:      sk-你的密钥',
      'Model ID:     claude-sonnet-4-5  # 或 gpt-4o 等后台分组里的模型',
      '',
      '# 4. 启用 Plan/Act 模式',
      '#    - Plan 模式：让 AI 先列出方案，确认后才动手',
      '#    - Act 模式：直接执行编辑/命令',
      '',
      '# 5. 建议开启 Auto Approve 但仅限读操作；',
      '#    Write/Execute 操作保持人工确认避免误操作',
    ].join('\n'),
  },
  {
    title: 'Cherry Studio',
    tag: '桌面 GUI 客户端',
    desc: '开源跨平台 AI 桌面应用，支持多模型多对话、知识库、本地数据。配置一次可同时使用多家模型。',
    code: [
      '# Cherry Studio 配置（桌面 GUI）',
      '# 1. 下载：cherry-ai.com 或 GitHub Release',
      '# 2. 设置 → 模型服务 → 添加 → OpenAI',
      '',
      `API 地址:  ${openAIBaseUrl.value}`,
      'API 密钥:  sk-你的密钥',
      '',
      '# 3. 模型列表里添加你需要用到的模型',
      '#    - claude-sonnet-4-5 / gpt-4o / gemini-2.0-flash',
      '',
      '# 4. 默认助手 → 选择模型即可开始对话',
      '',
      '# 提示：Cherry Studio 支持流式、图像输入、知识库 RAG',
    ].join('\n'),
  },
  {
    title: 'Continue.dev',
    tag: 'VSCode / JetBrains',
    desc: 'VSCode 和 JetBrains 双平台编码助手。在 ~/.continue/config.json 配置 models 数组。',
    code: [
      '// ~/.continue/config.json',
      '{',
      '  "models": [',
      '    {',
      '      "title": "Claude via Sub2API",',
      '      "provider": "openai",',
      '      "model": "claude-sonnet-4-5",',
      `      "apiBase": "${openAIBaseUrl.value}",`,
      '      "apiKey": "sk-你的密钥"',
      '    },',
      '    {',
      '      "title": "GPT-4o via Sub2API",',
      '      "provider": "openai",',
      '      "model": "gpt-4o",',
      `      "apiBase": "${openAIBaseUrl.value}",`,
      '      "apiKey": "sk-你的密钥"',
      '    }',
      '  ],',
      '  "tabAutocompleteModel": {',
      '    "title": "Autocomplete",',
      '    "provider": "openai",',
      '    "model": "gpt-4o-mini",',
      `    "apiBase": "${openAIBaseUrl.value}",`,
      '    "apiKey": "sk-你的密钥"',
      '  }',
      '}',
    ].join('\n'),
  },
  {
    title: 'JetBrains AI Assistant',
    tag: 'IntelliJ / PyCharm / WebStorm',
    desc: 'JetBrains 全家桶官方 AI 插件，从 2024.3 起支持 Custom OpenAI Provider。',
    code: [
      '# JetBrains AI Assistant 配置（GUI）',
      '# 1. Settings → Tools → AI Assistant → Models',
      '# 2. 在 Third-party providers 区域：',
      '#    点击 "+" → Add custom model',
      '',
      '# 3. Provider 类型选 OpenAI 兼容',
      `URL:        ${openAIBaseUrl.value}`,
      'API Key:    sk-你的密钥',
      'Model:      gpt-4o          # 或后台分组里其他模型名',
      '',
      '# 4. 在 AI Chat 窗口右上角选择该 custom model',
      '',
      '# 限制：',
      '#   - 代码补全 (inline completion) 走 JetBrains 自己的模型',
      '#   - Edit Code with AI 可用自定义模型',
      '#   - 部分 Agent 能力依赖 JetBrains 协议，可能不可用',
    ].join('\n'),
  },
  {
    title: 'LobeChat',
    tag: '开源 Web UI',
    desc: 'LobeChat 是开源 Web 端 ChatGPT 替代品，支持 PWA、Docker 部署。在设置里加自定义 OpenAI Provider。',
    code: [
      '# LobeChat 配置（Web GUI 或 Docker）',
      '# 1. 设置 → 语言模型 → OpenAI',
      '',
      `API 代理地址:  ${openAIBaseUrl.value}`,
      'API Key:       sk-你的密钥',
      '',
      '# 2. 自定义模型名（用 + 添加）',
      '#    claude-sonnet-4-5',
      '#    gpt-4o',
      '#    gemini-2.0-flash',
      '',
      '# 3. Docker 部署可通过环境变量预设：',
      `# OPENAI_API_KEY=sk-你的密钥`,
      `# OPENAI_PROXY_URL=${openAIBaseUrl.value}`,
      `# OPENAI_MODEL_LIST=claude-sonnet-4-5,gpt-4o`,
    ].join('\n'),
  },
  {
    title: 'HTTP / curl',
    tag: '原始请求',
    desc: '用 curl 确认密钥、模型名和 Base URL 都正确，是排障第一步。',
    code: [
      '# OpenAI 兼容 Chat Completions',
      `curl ${openAIBaseUrl.value}/chat/completions \\`,
      '  -H "Authorization: Bearer sk-你的密钥" \\',
      '  -H "Content-Type: application/json" \\',
      '  -d \'{"model":"gpt-4o-mini","messages":[{"role":"user","content":"hello"}]}\'',
      '',
      '# Anthropic Messages（Claude Code 协议）',
      `curl ${siteRootUrl.value}/v1/messages \\`,
      '  -H "x-api-key: sk-你的密钥" \\',
      '  -H "anthropic-version: 2023-06-01" \\',
      '  -H "Content-Type: application/json" \\',
      '  -d \'{"model":"claude-sonnet-4-5","max_tokens":256,"messages":[{"role":"user","content":"hello"}]}\'',
      '',
      '# 流式响应（添加 -N 不缓冲 + stream=true）',
      `curl -N ${openAIBaseUrl.value}/chat/completions \\`,
      '  -H "Authorization: Bearer sk-你的密钥" \\',
      '  -H "Content-Type: application/json" \\',
      '  -d \'{"model":"gpt-4o","stream":true,"messages":[{"role":"user","content":"讲个笑话"}]}\'',
    ].join('\n'),
  },
])

// 当前激活的 snippet（响应 activeSnippetIdx 切换）
const activeSnippet = computed(() => snippets.value[activeSnippetIdx.value])

const commonErrors = [
  { code: '401', title: '密钥无效', desc: '检查 API Key 是否复制完整，Header 是否使用 Bearer 或 x-api-key，密钥是否被禁用。' },
  { code: '403', title: '没有权限或触发风控', desc: '检查密钥分组、订阅权益、模型范围；如果是安全提醒，请调整问题后重试。' },
  { code: '404', title: 'Base URL 填错', desc: 'Claude Code 通常填根地址，OpenAI SDK / Cursor 通常填 `/v1`，Gemini 填 `/v1beta`。' },
  { code: '422', title: 'Cursor 中 Claude 报 422', desc: 'Cursor 开启 Override OpenAI Base URL 后 claude-* 模型容易 422，需要把 Anthropic API Key 关掉，让所有模型都走 OpenAI 兼容通道。' },
  { code: 'model_not_found', title: '模型不可用', desc: '检查模型名是否和可用渠道展示一致，以及你的分组是否有这个模型。' },
  { code: 'insufficient_quota', title: '额度不足', desc: '检查余额、订阅是否到期、兑换码权益是否已生效。' },
  { code: 'still official api', title: '仍在请求官方接口', desc: '修改环境变量后重启终端或客户端，确认配置文件路径被当前客户端读取。' },
]

// 模型选择指南：场景 → 推荐模型，实际模型名以后台分组配置为准
const modelGuide = [
  {
    scene: '日常对话 / 翻译 / 文档总结',
    desc: '快速响应、便宜，适合高频轻量场景',
    icon: 'chat',
    iconClass: 'bg-success-soft text-success dark:bg-success/15 dark:text-success-soft',
    models: ['gpt-4o-mini', 'claude-haiku-4-5', 'gemini-2.0-flash']
  },
  {
    scene: '编码 / 代码审查',
    desc: '需要长上下文 + 高准确率，强烈推荐 Claude Sonnet 系列',
    icon: 'terminal',
    iconClass: 'bg-info-soft text-info dark:bg-info/15 dark:text-info-soft',
    models: ['claude-sonnet-4-5', 'claude-opus-4-5', 'gpt-4o']
  },
  {
    scene: '复杂推理 / 数学 / 决策分析',
    desc: '深度推理任务，回复较慢但质量高',
    icon: 'lightbulb',
    iconClass: 'bg-warning-soft text-warning dark:bg-warning/15 dark:text-warning-soft',
    models: ['o1', 'o1-mini', 'claude-opus-4-5']
  },
  {
    scene: '图像理解 / OCR / 视觉问答',
    desc: '上传图片让模型描述/分析，需要 multimodal 模型',
    icon: 'eye',
    iconClass: 'bg-brand-50 text-brand-600 dark:bg-brand-500/15 dark:text-brand-300',
    models: ['gpt-4o', 'claude-sonnet-4-5', 'gemini-2.0-flash']
  },
  {
    scene: 'Agent / Function Calling',
    desc: '工具调用、Cline/Cursor Agent 等场景，模型要支持 function calling',
    icon: 'bolt',
    iconClass: 'bg-danger-soft text-danger dark:bg-danger/15 dark:text-danger-soft',
    models: ['claude-sonnet-4-5', 'gpt-4o', 'gpt-4o-mini']
  },
  {
    scene: '超长上下文 / 长文档分析',
    desc: '200K+ tokens 上下文，可一次性塞入整本书或大代码库',
    icon: 'document',
    iconClass: 'bg-success-soft text-success dark:bg-success/15 dark:text-success-soft',
    models: ['claude-sonnet-4-5', 'gemini-2.0-flash', 'gpt-4o']
  }
]

// 高级用法：流式、图像、工具调用、缓存
const advancedFeatures = [
  {
    icon: 'bolt',
    title: '流式响应（Streaming）',
    desc: '请求体加 `"stream": true`，响应会以 SSE 格式逐 token 推送。前端逐块拼接 message_delta / content_block_delta。',
    example: [
      '// OpenAI SDK Node.js',
      'const stream = await client.chat.completions.create({',
      '  model: "gpt-4o",',
      '  stream: true,',
      '  messages: [{ role: "user", content: "讲个笑话" }],',
      '})',
      'for await (const chunk of stream) {',
      '  process.stdout.write(chunk.choices[0]?.delta?.content || "")',
      '}',
    ].join('\n')
  },
  {
    icon: 'eye',
    title: '图像输入（Vision）',
    desc: '把图片以 URL 或 base64 dataURL 形式放入 content 数组。模型要支持 vision（如 gpt-4o、claude-sonnet-4-5、gemini-*）。',
    example: [
      'curl https://your-site.com/v1/chat/completions \\',
      '  -H "Authorization: Bearer sk-你的密钥" \\',
      '  -d \'{',
      '    "model": "gpt-4o",',
      '    "messages": [{',
      '      "role": "user",',
      '      "content": [',
      '        {"type": "text", "text": "这张图里有什么？"},',
      '        {"type": "image_url", "image_url": {"url": "https://example.com/cat.jpg"}}',
      '      ]',
      '    }]',
      '  }\'',
    ].join('\n')
  },
  {
    icon: 'cog',
    title: '工具调用（Function Calling / Tools）',
    desc: '让模型在回答时调用预定义函数。Agent 类客户端（Cline/Cursor Agent）依赖这个能力。',
    example: [
      '{',
      '  "model": "claude-sonnet-4-5",',
      '  "messages": [{"role": "user", "content": "北京今天天气"}],',
      '  "tools": [{',
      '    "type": "function",',
      '    "function": {',
      '      "name": "get_weather",',
      '      "description": "查询城市天气",',
      '      "parameters": {',
      '        "type": "object",',
      '        "properties": {"city": {"type": "string"}},',
      '        "required": ["city"]',
      '      }',
      '    }',
      '  }]',
      '}',
    ].join('\n')
  },
  {
    icon: 'database',
    title: '提示词缓存（Prompt Caching）',
    desc: 'Claude 系列支持长 system prompt 缓存。重复请求时缓存命中部分按 0.1x 计费，大幅降低成本。在 Anthropic 协议下用 cache_control 标记缓存断点。',
    example: [
      '{',
      '  "model": "claude-sonnet-4-5",',
      '  "system": [',
      '    {',
      '      "type": "text",',
      '      "text": "你是代码审查助手...（长 prompt）",',
      '      "cache_control": {"type": "ephemeral"}',
      '    }',
      '  ],',
      '  "messages": [...]',
      '}',
    ].join('\n')
  },
  {
    icon: 'shield',
    title: '速率限制与并发',
    desc: '后台分组可能限制每分钟请求数（RPM）和并发。批量任务建议加 sleep 或队列，避免 429。可以在密钥页面查看当前限额。',
    example: ''
  }
]

// 使用安全提醒列表
const securityTips = [
  '不要把 API 密钥发给他人，也不要放到公开仓库、前端网页或截图里。密钥泄露后，别人可以直接消耗你的额度。',
  '一密钥一用途：建议为不同客户端/项目创建独立密钥，方便单独禁用或追踪用量。',
  '修改环境变量或配置文件后，重启终端或客户端，确认新配置已经生效。',
  '如果页面提示内容可能触发安全限制，请调整问题再发送。明显违规或重复命中可能会被管理员复核。',
  '客户端 IDE / 编辑器更新后可能恢复默认 Base URL，更新后建议复查设置。'
]

const sources = [
  { label: 'Claude Code LLM Gateway', href: 'https://code.claude.com/docs/zh-CN/llm-gateway' },
  { label: 'Claude Code settings', href: 'https://docs.claude.com/en/docs/claude-code/settings' },
  { label: 'OpenAI Codex configuration', href: 'https://developers.openai.com/codex/config' },
  { label: 'OpenAI API reference', href: 'https://platform.openai.com/docs/api-reference' },
  { label: 'Cursor 自定义 OpenAI Base URL', href: 'https://docs.cursor.com/settings/models' },
  { label: 'Cline OpenAI Compatible Provider', href: 'https://docs.cline.bot/provider-config/openai-compatible' },
  { label: 'Continue.dev 配置', href: 'https://docs.continue.dev/setup/select-provider' },
  { label: 'JetBrains AI Assistant 自定义模型', href: 'https://www.jetbrains.com/help/ai-assistant/use-custom-models.html' },
  { label: 'Cherry Studio 文档', href: 'https://docs.cherry-ai.com/' },
  { label: 'LobeChat 文档', href: 'https://lobehub.com/docs' }
]

// 轻量语法高亮：覆盖 shell / json / toml 混合内容，无依赖
// 高亮规则：注释（# 或 //）→ 灰、字符串 → 翠绿、URL → 青蓝、关键字 → 粉紫、数字 → 琥珀
function escapeHtml(s: string): string {
  return s.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;')
}

function highlightLine(raw: string): string {
  const line = escapeHtml(raw)
  // 整行注释（# 或 //）— 把整行染灰，不再做其它替换
  if (/^\s*(#|\/\/)/.test(line)) {
    return `<span class="hl-c">${line}</span>`
  }
  // 行内 token 替换（先字符串再 URL 再关键字，避免互相覆盖）
  let out = line
  // 字符串字面量（双/单引号包裹），允许内部转义
  out = out.replace(/(&quot;[^&\n]*?&quot;|&#39;[^&\n]*?&#39;)/g, '<span class="hl-s">$1</span>')
  // URL（http/https）
  out = out.replace(/(https?:\/\/[^\s&"'<]+)/g, '<span class="hl-u">$1</span>')
  // shell 关键字
  out = out.replace(/\b(export|set|unset)\b/g, '<span class="hl-k">$1</span>')
  // PowerShell 变量 $env:VAR
  out = out.replace(/(\$env:[A-Za-z_][A-Za-z0-9_]*)/g, '<span class="hl-v">$1</span>')
  // JSON / TOML 字段名
  out = out.replace(/^(\s*)([A-Z_][A-Z0-9_]+)(=)/g, '$1<span class="hl-p">$2</span>$3')
  // 数字
  out = out.replace(/\b(\d+)\b/g, '<span class="hl-n">$1</span>')
  return out
}

function highlightCode(code: string): string {
  return code.split('\n').map(highlightLine).join('\n')
}

// 端点地址复制
const copiedEndpoint = ref<string | null>(null)
let endpointResetTimer: ReturnType<typeof setTimeout> | null = null
async function copyEndpoint(url: string): Promise<void> {
  try {
    await navigator.clipboard.writeText(url)
    copiedEndpoint.value = url
    if (endpointResetTimer) clearTimeout(endpointResetTimer)
    endpointResetTimer = setTimeout(() => { copiedEndpoint.value = null }, 2000)
  } catch (error) {
    console.error('Failed to copy endpoint:', error)
  }
}
</script>

<style scoped>
/* ============================================================================
 * 使用文档：与全站 Notion 风一致，仅保留 snippet tab 与 code 标签的局部样式
 * ========================================================================= */

/* code 内联标签：黛蓝信息调代码片段 */
.docs-hint-code {
  display: inline-block;
  background: rgb(255 255 255 / 0.7);
  border-radius: 0.25rem;
  padding: 0.05rem 0.35rem;
  font-family: ui-monospace, monospace;
  font-size: 0.75rem;
  color: rgb(58 85 112); /* info */
}

:global(:root.dark) .docs-hint-code {
  background: rgb(52 48 42 / 0.7); /* dark-800 */
  color: rgb(227 232 238); /* info-soft */
}

/* ============ 接入示例 Tab 切换：黑底 active 与全站时间窗口风格一致 ============ */
/* mask 渐变：tab 横滚时左右淡出，暗示"还能往那边滚" */
.snippet-tabs {
  display: flex;
  gap: 0;
  overflow-x: auto;
  border-bottom: 1px solid rgb(221 213 196 / 0.6); /* gray-200 */
  background: rgb(246 242 233 / 0.5); /* gray-50 */
  padding: 0 0.75rem;
  scrollbar-width: none;
  -webkit-mask-image: linear-gradient(to right, transparent 0, black 1.5rem, black calc(100% - 1.5rem), transparent 100%);
  mask-image: linear-gradient(to right, transparent 0, black 1.5rem, black calc(100% - 1.5rem), transparent 100%);
}

.snippet-tabs::-webkit-scrollbar {
  display: none;
}

:global(:root.dark) .snippet-tabs {
  border-bottom-color: rgb(70 65 58 / 0.6); /* dark-700 */
  background: rgb(52 48 42 / 0.4); /* dark-800 */
}

.snippet-tab {
  position: relative;
  display: inline-flex;
  align-items: center;
  gap: 0.375rem;
  white-space: nowrap;
  padding: 0.75rem 1rem;
  font-size: 0.8125rem;
  font-weight: 500;
  color: rgb(138 130 117); /* gray-500 */
  border-bottom: 2px solid transparent;
  margin-bottom: -1px;
  transition: color 0.15s, border-color 0.15s;
}

.snippet-tab:hover:not(.snippet-tab-active) {
  color: rgb(42 39 34); /* gray-900 */
}

.snippet-tab-active {
  color: rgb(42 39 34); /* gray-900 */
  border-bottom-color: rgb(42 39 34); /* gray-900 */
}

:global(:root.dark) .snippet-tab {
  color: rgb(180 171 152); /* dark-300 */
}

:global(:root.dark) .snippet-tab:hover:not(.snippet-tab-active) {
  color: rgb(255 255 255);
}

:global(:root.dark) .snippet-tab-active {
  color: rgb(255 255 255);
  border-bottom-color: rgb(255 255 255);
}

/* ============ 代码块语法高亮配色（深底主题）============ */
.snippet-pre :deep(.hl-c) { color: rgb(100 116 139); font-style: italic; }   /* 注释：slate-500 */
.snippet-pre :deep(.hl-s) { color: rgb(134 239 172); }                       /* 字符串：emerald-300 */
.snippet-pre :deep(.hl-u) { color: rgb(125 211 252); text-decoration: underline; }  /* URL：sky-300 */
.snippet-pre :deep(.hl-k) { color: rgb(244 114 182); font-weight: 600; }     /* 关键字：pink-400 */
.snippet-pre :deep(.hl-v) { color: rgb(251 191 36); }                        /* 变量：amber-400 */
.snippet-pre :deep(.hl-p) { color: rgb(196 181 253); }                       /* 字段名：violet-300 */
.snippet-pre :deep(.hl-n) { color: rgb(252 165 165); }                       /* 数字：red-300 */

</style>
