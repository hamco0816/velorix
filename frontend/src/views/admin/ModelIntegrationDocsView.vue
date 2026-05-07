<template>
  <AppLayout>
    <div class="mx-auto max-w-7xl space-y-6">
      <header class="space-y-4">
        <div class="flex flex-col gap-4 xl:flex-row xl:items-end xl:justify-between">
          <div>
            <div class="inline-flex items-center gap-2 rounded-md border border-primary-200 bg-primary-50 px-2.5 py-1 text-xs font-medium text-primary-700 dark:border-primary-800 dark:bg-primary-950/40 dark:text-primary-300">
              <Icon name="book" size="sm" />
              管理员文档
            </div>
            <h1 class="mt-3 text-2xl font-semibold text-gray-950 dark:text-white">配置词典与运维教程</h1>
            <p class="mt-2 max-w-3xl text-sm leading-6 text-gray-600 dark:text-gray-300">
              给管理员看的后台配置说明，重点解释小白容易混淆的专业术语、配置位置、风险点和上线前检查顺序。新增平台、支付、风控或用户文档时，先按这份清单核对。
            </p>
          </div>
          <div class="grid grid-cols-2 gap-2 text-center sm:grid-cols-4 xl:min-w-[520px]">
            <div v-for="stat in quickStats" :key="stat.label" class="rounded-lg border border-gray-200 bg-white px-3 py-2 dark:border-dark-700 dark:bg-dark-900">
              <div class="text-lg font-semibold text-gray-950 dark:text-white">{{ stat.value }}</div>
              <div class="text-xs text-gray-500 dark:text-gray-400">{{ stat.label }}</div>
            </div>
          </div>
        </div>

        <div class="rounded-lg border border-amber-200 bg-amber-50 px-4 py-3 text-sm leading-6 text-amber-900 dark:border-amber-800/70 dark:bg-amber-950/30 dark:text-amber-100">
          后台配置不要只看“能不能保存”。完整上线链路是：出口可控、账号可调度、分组能授权、渠道能展示、模型能计费、错误能归因、风控能拦截、日志能复核。
        </div>
      </header>

      <section class="space-y-3">
        <div class="flex items-center gap-2">
          <Icon name="checkCircle" size="md" class="text-green-600 dark:text-green-400" />
          <h2 class="text-lg font-semibold text-gray-950 dark:text-white">新平台上线顺序</h2>
        </div>
        <div class="grid grid-cols-1 gap-3 md:grid-cols-2 xl:grid-cols-4">
          <article
            v-for="step in setupSteps"
            :key="step.title"
            class="rounded-lg border border-gray-200 bg-white p-4 dark:border-dark-700 dark:bg-dark-900"
          >
            <div class="flex items-start gap-3">
              <span class="flex h-7 w-7 flex-shrink-0 items-center justify-center rounded-md bg-primary-600 text-xs font-semibold text-white">
                {{ step.no }}
              </span>
              <div class="min-w-0">
                <h3 class="text-sm font-semibold text-gray-950 dark:text-white">{{ step.title }}</h3>
                <p class="mt-1 text-xs leading-5 text-gray-600 dark:text-gray-300">{{ step.desc }}</p>
                <div class="mt-2 font-mono text-[11px] text-gray-500 dark:text-gray-400">{{ step.path }}</div>
              </div>
            </div>
          </article>
        </div>
      </section>

      <section class="card overflow-hidden">
        <div class="border-b border-gray-100 px-5 py-4 dark:border-dark-700">
          <div class="flex items-center gap-2">
            <Icon name="grid" size="md" class="text-primary-600 dark:text-primary-400" />
            <h2 class="text-lg font-semibold text-gray-950 dark:text-white">后台术语词典</h2>
          </div>
          <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
            这里解释配置项的业务含义，不替代表单校验。遇到问题时优先从“影响”和“常见错误”两列定位。
          </p>
        </div>
        <div class="overflow-x-auto">
          <table class="min-w-full divide-y divide-gray-100 text-sm dark:divide-dark-700">
            <thead class="bg-gray-50 text-xs uppercase text-gray-500 dark:bg-dark-800 dark:text-gray-400">
              <tr>
                <th class="px-4 py-3 text-left font-medium">配置/术语</th>
                <th class="px-4 py-3 text-left font-medium">白话解释</th>
                <th class="px-4 py-3 text-left font-medium">影响</th>
                <th class="px-4 py-3 text-left font-medium">常见错误</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-gray-100 bg-white dark:divide-dark-700 dark:bg-dark-900">
              <tr v-for="item in glossary" :key="item.term">
                <td class="whitespace-nowrap px-4 py-3 font-medium text-gray-950 dark:text-white">{{ item.term }}</td>
                <td class="min-w-[260px] px-4 py-3 text-gray-700 dark:text-gray-300">{{ item.meaning }}</td>
                <td class="min-w-[240px] px-4 py-3 text-gray-600 dark:text-gray-400">{{ item.impact }}</td>
                <td class="min-w-[240px] px-4 py-3 text-gray-600 dark:text-gray-400">{{ item.mistake }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </section>

      <section class="grid grid-cols-1 gap-4 xl:grid-cols-2">
        <article v-for="section in adminSections" :key="section.title" class="card p-5">
          <div class="flex items-center gap-2">
            <Icon :name="section.icon" size="md" :class="section.iconClass" />
            <h2 class="text-lg font-semibold text-gray-950 dark:text-white">{{ section.title }}</h2>
          </div>
          <div class="mt-4 divide-y divide-gray-100 dark:divide-dark-700">
            <div v-for="item in section.items" :key="item.title" class="py-3 first:pt-0 last:pb-0">
              <div class="flex flex-wrap items-center gap-2">
                <h3 class="text-sm font-semibold text-gray-950 dark:text-white">{{ item.title }}</h3>
                <span v-if="item.tag" class="rounded-md bg-gray-100 px-2 py-0.5 text-[11px] text-gray-600 dark:bg-dark-700 dark:text-gray-300">
                  {{ item.tag }}
                </span>
              </div>
              <p class="mt-1 text-xs leading-5 text-gray-600 dark:text-gray-300">{{ item.desc }}</p>
            </div>
          </div>
        </article>
      </section>

      <section class="card p-5">
        <div class="flex items-center gap-2">
          <Icon name="shield" size="md" class="text-red-600 dark:text-red-400" />
          <h2 class="text-lg font-semibold text-gray-950 dark:text-white">风控与敏感词处理建议</h2>
        </div>
        <div class="mt-4 grid grid-cols-1 gap-3 md:grid-cols-3">
          <div v-for="policy in riskPolicies" :key="policy.title" class="rounded-lg border border-gray-200 bg-gray-50 p-4 dark:border-dark-700 dark:bg-dark-800/70">
            <h3 class="text-sm font-semibold text-gray-950 dark:text-white">{{ policy.title }}</h3>
            <p class="mt-2 text-xs leading-5 text-gray-600 dark:text-gray-300">{{ policy.desc }}</p>
          </div>
        </div>
      </section>

      <section class="space-y-3">
        <div class="flex items-center gap-2">
          <Icon name="questionCircle" size="md" class="text-gray-600 dark:text-gray-300" />
          <h2 class="text-lg font-semibold text-gray-950 dark:text-white">常见排查路径</h2>
        </div>
        <div class="grid grid-cols-1 gap-3 md:grid-cols-2 xl:grid-cols-3">
          <div v-for="item in troubleshooting" :key="item.problem" class="rounded-lg border border-gray-200 bg-white p-3 dark:border-dark-700 dark:bg-dark-900">
            <h3 class="text-sm font-semibold text-gray-950 dark:text-white">{{ item.problem }}</h3>
            <p class="mt-1 text-xs leading-5 text-gray-600 dark:text-gray-300">{{ item.check }}</p>
          </div>
        </div>
      </section>

      <section class="rounded-lg border border-blue-200 bg-blue-50 px-4 py-3 text-sm leading-6 text-blue-900 dark:border-blue-800/70 dark:bg-blue-950/30 dark:text-blue-100">
        <div class="font-medium">外部接入依据</div>
        <div class="mt-1">
          Claude Code 官方网关文档要求网关暴露 Anthropic Messages 端点，并说明 `ANTHROPIC_AUTH_TOKEN` 的认证行为；Codex 官方文档使用 `config.toml` 管理 provider。用户接入页已经按这些行为写成可复制示例。
        </div>
      </section>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'

type DocIcon = 'server' | 'sparkles' | 'cpu' | 'creditCard'

const quickStats = [
  { value: '8', label: '上线步骤' },
  { value: '18', label: '核心术语' },
  { value: '7', label: '风险检查' },
  { value: '2', label: '文档页面' },
]

const setupSteps = [
  { no: '1', title: '配置出口 IP', desc: '先录入代理并做连通性/质量检测，确认出口 IP、地区和上游可达性。', path: '/admin/proxies' },
  { no: '2', title: '配置上游账号', desc: '选择平台、账号类型、凭据和代理，确认账号启用、未过期、可调度。', path: '/admin/accounts' },
  { no: '3', title: '创建分组边界', desc: '分组决定用户、密钥、订阅和账号之间的授权关系，先定边界再开放。', path: '/admin/groups' },
  { no: '4', title: '绑定账号分组', desc: '账号和分组必须平台一致，避免用户看得到模型但调度不到账号。', path: '/admin/accounts' },
  { no: '5', title: '配置渠道和映射', desc: '渠道聚合分组、模型映射和定价，是模型广场与路由的主要数据来源。', path: '/admin/channels/pricing' },
  { no: '6', title: '配置价格策略', desc: '按 token、按次、图片或倍率计费都要明确，否则前台只能显示未配置。', path: '/admin/channels/pricing' },
  { no: '7', title: '配置监控模板', desc: '为主模型和附加模型建立监控项，模型广场才能展示可用状态。', path: '/admin/channels/monitor' },
  { no: '8', title: '打开风控日志', desc: '敏感词、AI 审核、用户提醒和人工复核要有日志闭环。', path: '/admin/safety-risk' },
]

const glossary = [
  { term: '平台', meaning: 'OpenAI、Anthropic、Gemini、Bedrock、Antigravity 或兼容上游的业务归类。', impact: '账号、分组、渠道、模型广场都会按平台归类。', mistake: '把不同平台账号混到同一分组，导致调度失败或展示错乱。' },
  { term: '账号', meaning: '真正去调用上游的凭据实体，可以是 API Key、OAuth、Setup Token 或云账号。', impact: '决定请求实际从哪个上游账号发出。', mistake: '只录账号不绑定代理、不检查状态，后续容易限流或封号。' },
  { term: '代理/出口 IP', meaning: '账号访问上游时使用的网络出口。代理记录不等于真实出口 IP，真实 IP 需要检测。', impact: '影响账号风控、可达性、延迟和同平台账号隔离。', mistake: '同平台多个账号共用同一出口，或动态代理变更后不复测。' },
  { term: '分组', meaning: '把用户、密钥、订阅和账号池隔离开的授权边界。', impact: '决定用户能看到、能调用哪些模型和账号。', mistake: '给用户开了分组，但密钥没有绑定分组。' },
  { term: '渠道', meaning: '面向用户展示的能力入口，聚合平台、分组、模型映射和计费规则。', impact: '可用渠道、模型广场和前台价格都依赖渠道。', mistake: '渠道启用但没有用户可访问分组，前台仍然看不到。' },
  { term: '模型映射', meaning: '把用户请求的模型名映射到上游真实模型名。', impact: '影响路由、计费、监控和模型广场展示。', mistake: '通配符范围过大，导致错误模型被错误定价或调度。' },
  { term: '模型倍率', meaning: '按基础价格乘以倍率来计算费用的简化规则。', impact: '适合价格随模型等级变化但不想逐项维护时使用。', mistake: '倍率和固定价格同时配置，管理员误判实际扣费。' },
  { term: 'TLS 指纹模板', meaning: '控制客户端请求上游时的 TLS/HTTP 行为特征，用来更接近真实客户端或统一网关特征。', impact: '可能影响上游风控、连接兼容性和账号稳定性。', mistake: '随意切模板，导致同一账号短时间出现明显环境变化。' },
  { term: '请求头模板', meaning: '统一追加或覆盖上游请求 Header 的配置。', impact: '影响认证、Beta 功能、来源识别和网关兼容性。', mistake: '覆盖必要认证头，或没有转发 Claude Code 所需的 anthropic-* 头。' },
  { term: '错误透传规则', meaning: '决定上游错误原样返回给用户、改成友好提示、还是隐藏内部细节。', impact: '影响用户排障体验和后台错误归因。', mistake: '把上游封号/风控原因完整透给用户，或把所有错误都吞成“未知错误”。' },
  { term: '错误归属', meaning: '判断错误属于用户请求、平台账号、代理网络、系统配置还是上游服务。', impact: '决定是否扣费、是否禁用账号、是否写入监控异常。', mistake: '把用户参数错误当成账号故障，造成可用账号被误下线。' },
  { term: '跳过监控', meaning: '某些错误不参与渠道健康统计，只写日志或返回用户。', impact: '避免用户输入错误污染可用率。', mistake: '把真实上游故障也跳过，模型广场显示虚假可用。' },
  { term: '敏感词过滤', meaning: '请求进入上游前的本地规则拦截，可提示用户修改而不是直接 403。', impact: '保护上游账号，减少触发平台安全策略。', mistake: '只做前端提示，不在后端拦截。' },
  { term: 'AI 审核', meaning: '规则不确定时用轻量审核模型二次判断风险。', impact: '提高准确率，但会增加少量延迟和成本。', mistake: '所有请求都走 AI 审核，导致体验变慢。' },
  { term: '风控日志', meaning: '记录命中规则、处理动作、是否 AI 审核、用户提示和管理员复核状态。', impact: '管理员可复核、清空警告、识别恶意用户。', mistake: '只拦截不留证据，后续无法解释封禁或误杀。' },
  { term: '虎皮椒/易支付', meaning: '第三方聚合支付通道，用于微信和支付宝收款。', impact: '影响充值订阅、订单回调、金额校验和到账确认。', mistake: '只校验订单号不校验金额，存在金额绕过风险。' },
  { term: '自定义端点', meaning: '展示给用户的 API Base URL，例如 OpenAI `/v1`、Claude 根域名、Gemini `/v1beta`。', impact: '影响用户能否把密钥接入 Claude Code、Codex 或 SDK。', mistake: '所有客户端都填同一个路径，导致客户端自动拼接后 404。' },
  { term: 'Backend Mode', meaning: '后端模式下只保留必要入口，适合嵌入式或简化部署。', impact: '会影响侧边栏可见菜单和前台功能入口。', mistake: '以为功能没实现，其实是被模式或 feature flag 隐藏。' },
]

const adminSections: Array<{
  title: string
  icon: DocIcon
  iconClass: string
  items: Array<{ title: string; tag: string; desc: string }>
}> = [
  {
    title: '账号、代理与 IP',
    icon: 'server',
    iconClass: 'text-sky-600 dark:text-sky-400',
    items: [
      { title: '同平台账号尽量独立出口', tag: '账号安全', desc: '不同平台可以复用代理；同平台账号建议一账号一出口。真实出口 IP 只能靠检测、日志或代理商面板确认。' },
      { title: '动态代理要定期复测', tag: '稳定性', desc: '动态代理可能换出口，同平台账号可能在一段时间后出现 IP 撞车，建议监控最近检测 IP 和检测时间。' },
      { title: '账号异常先分层定位', tag: '排障', desc: '先看用户参数，再看分组授权，再看代理网络，最后才判断账号凭据失效，避免误禁可用账号。' },
    ],
  },
  {
    title: '渠道、模型广场与状态',
    icon: 'sparkles',
    iconClass: 'text-violet-600 dark:text-violet-400',
    items: [
      { title: '模型广场展示条件', tag: '展示', desc: '渠道启用、用户有可访问分组、存在模型映射和价格，才适合展示给用户。' },
      { title: '可用状态来自监控', tag: '状态', desc: '模型广场不应在用户打开页面时临时探测上游，应复用后台渠道监控结果，避免页面慢和额外消耗。' },
      { title: '主模型与附加模型', tag: '监控', desc: '主模型用于渠道整体状态，附加模型用于更细的模型级状态展示。没有监控结果应显示未监控，不要默认可用。' },
    ],
  },
  {
    title: 'TLS 指纹与错误透传',
    icon: 'cpu',
    iconClass: 'text-emerald-600 dark:text-emerald-400',
    items: [
      { title: 'TLS 指纹模板不要频繁换', tag: '风控', desc: '同一账号短时间切换明显不同的网络/客户端特征，会增加上游风险。变更后应小流量观察错误率。' },
      { title: '错误透传要分级', tag: '体验', desc: '用户参数错误可以友好提示；余额、限流、封禁、认证失败等内部细节应写日志，前台只给可理解提示。' },
      { title: '监控错误不要污染用户错误', tag: '归因', desc: '监控请求失败和真实用户请求失败要区分，否则会误判渠道健康或误导用户。' },
    ],
  },
  {
    title: '支付、订阅与金额校验',
    icon: 'creditCard',
    iconClass: 'text-amber-600 dark:text-amber-400',
    items: [
      { title: '订单金额必须后端校验', tag: '支付安全', desc: '虎皮椒、微信、支付宝回调都不能只看订单号，必须校验实际金额、计划金额、订单状态和回调签名。' },
      { title: '订阅权益绑定分组', tag: '授权', desc: '支付成功只代表订单完成，还要把用户权益同步到正确分组，否则用户充值后仍然无法调用模型。' },
      { title: '回调失败要可恢复', tag: '运维', desc: '支付回调、前端 resume token 和订单轮询都要能处理网络抖动，避免用户付了款但页面显示失败。' },
    ],
  },
]

const riskPolicies = [
  {
    title: '默认不直接 403',
    desc: '对普通用户给出“内容可能触发安全限制，请调整问题”的提示；只有明显滥用、重复命中或管理员策略要求时才强拦截。',
  },
  {
    title: '规则优先，AI 审核补充',
    desc: '高置信敏感词和已知越狱话术本地快速处理；低置信请求再走 AI 审核，避免所有请求变慢。',
  },
  {
    title: '日志必须可复核',
    desc: '记录用户、密钥、模型、命中词、处理动作、是否 AI 审核、管理员备注和复核状态，方便清空警告或确认封禁。',
  },
]

const troubleshooting = [
  { problem: '用户模型广场为空', check: '检查可用渠道开关、渠道状态、用户分组、密钥分组、渠道模型映射和价格配置。' },
  { problem: '模型显示可用但调用失败', check: '检查监控时间是否过旧、账号是否可调度、分组平台是否一致、代理出口是否可达。' },
  { problem: '价格显示未配置', check: '检查渠道自定义定价、模型名大小写、通配符展开和 LiteLLM 定价回落。' },
  { problem: 'Claude Code 接入 404', check: '检查客户端是否自动拼接 `/v1/messages`。Claude Code 通常填根域名，OpenAI 兼容 SDK 通常填 `/v1`。' },
  { problem: 'Codex 接入不生效', check: '检查 Codex 版本、`~/.codex/config.toml`、provider 的 `base_url`、`wire_api` 和 API key 环境变量。' },
  { problem: '上游错误透传不清楚', check: '检查错误归属、错误透传规则、是否跳过监控，以及用户前台是否需要友好提示。' },
]
</script>
