<template>
  <AppLayout wide>
    <div class="risk-page space-y-5">
      <!-- 简洁说明 chip 替代 hero（页面标题已由 AppHeader 提供） -->
      <div class="flex items-start gap-2.5 rounded-xl border border-amber-200/60 bg-amber-50/60 px-4 py-3 dark:border-amber-500/20 dark:bg-amber-500/5">
        <Icon name="shield" size="sm" class="mt-0.5 flex-shrink-0 text-amber-600 dark:text-amber-400" />
        <p class="text-sm leading-6 text-amber-900 dark:text-amber-200">
          风控拦截会在请求发送到上游前执行。这里记录命中规则、是否经过 AI 审核以及管理员复核状态，便于排查风险和清空用户管控记录。
        </p>
      </div>

      <section class="surface-card p-4 sm:p-5">
        <div class="grid grid-cols-1 gap-3 sm:grid-cols-2 sm:gap-4 xl:grid-cols-[minmax(190px,1fr)_minmax(190px,1fr)_minmax(190px,1fr)_minmax(320px,1.25fr)_auto]">
          <label class="space-y-2">
            <span class="filter-label">状态</span>
            <Select
              v-model="selectedStatus"
              :options="statusFilterOptions"
              class="risk-select"
            />
          </label>

          <label class="space-y-2">
            <span class="filter-label">时间</span>
            <Select
              v-model="selectedTimeRange"
              :options="timeRangeOptions"
              class="risk-select"
            />
          </label>

          <label class="space-y-2">
            <span class="filter-label">用户 ID 可选</span>
            <input
              v-model.trim="userIdInput"
              type="number"
              min="1"
              class="risk-filter-control"
              placeholder="请输入用户 ID"
            />
          </label>

          <label class="space-y-2">
            <span class="filter-label">搜索用户/邮箱、API Key、请求</span>
            <span class="relative block">
              <input
                v-model.trim="filters.q"
                class="risk-filter-control pr-10"
                placeholder="请输入关键词"
                @keyup.enter="applyFilters"
              />
              <Icon name="search" size="sm" class="pointer-events-none absolute right-3 top-1/2 -translate-y-1/2 text-gray-400" />
            </span>
          </label>

          <div class="flex items-end">
            <button class="btn btn-secondary h-10 w-full justify-center gap-1.5 xl:w-auto" :disabled="loading" @click="loadEvents">
              <Icon name="refresh" size="sm" />
              <span>刷新</span>
            </button>
          </div>
        </div>

        <div class="mt-5 flex flex-wrap items-center gap-2">
          <button class="btn btn-primary min-w-[88px]" :disabled="loading" @click="applyFilters">查询</button>
          <button class="btn btn-secondary min-w-[88px]" :disabled="loading" @click="resetFilters">重置</button>
        </div>
      </section>

      <div class="grid grid-cols-2 gap-3 sm:grid-cols-3 lg:grid-cols-5">
        <div v-for="stat in riskStats" :key="stat.key" class="metric-card">
          <span class="risk-stat-icon" :class="stat.iconClass">
            <Icon :name="stat.icon" size="sm" :stroke-width="1.75" />
          </span>
          <div class="min-w-0 flex-1">
            <p class="text-[13px] font-medium text-gray-500 dark:text-dark-400">{{ stat.label }}</p>
            <p class="mt-1 text-xl font-semibold leading-tight tabular-nums text-gray-900 sm:text-[24px] dark:text-white">{{ stat.value }}</p>
            <p class="mt-1 text-xs text-gray-400 dark:text-dark-500">{{ stat.hint }}</p>
          </div>
        </div>
      </div>

      <!-- 白名单管理 + 规则统计：默认折叠，admin 可点开 -->
      <section class="surface-card overflow-hidden">
        <div class="flex flex-wrap items-center gap-2 border-b border-gray-100 px-4 py-3 dark:border-dark-700">
          <button
            type="button"
            class="rounded-md px-3 py-1.5 text-sm font-medium transition-colors"
            :class="activePanel === 'allowlist' ? 'bg-gray-900 text-white dark:bg-white dark:text-gray-900' : 'text-gray-600 hover:bg-gray-100 dark:text-gray-300 dark:hover:bg-dark-700'"
            @click="togglePanel('allowlist')"
          >
            白名单管理 ({{ allowlistDetail.length }})
          </button>
          <button
            type="button"
            class="rounded-md px-3 py-1.5 text-sm font-medium transition-colors"
            :class="activePanel === 'stats' ? 'bg-gray-900 text-white dark:bg-white dark:text-gray-900' : 'text-gray-600 hover:bg-gray-100 dark:text-gray-300 dark:hover:bg-dark-700'"
            @click="togglePanel('stats')"
          >
            规则命中统计
          </button>
          <!-- 批量 AI 复核：把当前列表里所有 ai_reviewed=false 的事件依次同步复核 -->
          <button
            v-if="pendingAIReviewCount > 0"
            type="button"
            class="rounded-md bg-violet-100 px-3 py-1.5 text-sm font-medium text-violet-700 hover:bg-violet-200 disabled:opacity-50 dark:bg-violet-900/30 dark:text-violet-300 dark:hover:bg-violet-900/50"
            :disabled="batchAIReviewing"
            title="对当前页所有未经过 AI 的事件依次跑 AI 审核（顺序执行避免限流）"
            @click="batchAIReview"
          >
            {{ batchAIReviewing ? `AI 复核中… ${batchProgress.done}/${batchProgress.total}` : `批量 AI 复核当前页 (${pendingAIReviewCount})` }}
          </button>
          <button
            v-if="activePanel"
            type="button"
            class="ml-auto rounded-md p-1.5 text-gray-400 hover:bg-gray-100 hover:text-gray-700 dark:hover:bg-dark-700 dark:hover:text-gray-200"
            title="收起"
            @click="activePanel = ''"
          >
            <Icon name="x" size="sm" />
          </button>
        </div>

        <!-- 白名单管理面板 -->
        <div v-if="activePanel === 'allowlist'" class="px-4 py-4">
          <p class="mb-3 text-xs text-gray-500 dark:text-gray-400">
            在白名单的用户所有请求会跳过敏感词检测。可从下方搜索框直接添加用户，或在事件列表点击"加入白名单"。
          </p>
          <!-- 直接添加用户：输入 ID / 邮箱 / 用户名实时搜索 → 点击列表项添加 -->
          <div class="mb-4 rounded-lg border border-gray-200 bg-gray-50/50 p-3 dark:border-dark-700 dark:bg-dark-800/40">
            <label class="mb-1.5 block text-xs font-medium text-gray-600 dark:text-gray-300">
              添加用户到白名单
            </label>
            <div class="relative">
              <input
                v-model="allowlistSearchQuery"
                type="text"
                placeholder="输入用户 ID / 邮箱 / 用户名 搜索"
                class="input w-full text-sm"
                @input="onAllowlistSearchInput"
                @focus="allowlistSearchFocused = true"
                @blur="onAllowlistSearchBlur"
              />
              <!-- 搜索下拉 -->
              <div
                v-if="allowlistSearchFocused && allowlistSearchResults.length > 0"
                class="absolute z-10 mt-1 max-h-60 w-full overflow-auto rounded-md border border-gray-200 bg-white shadow-lg dark:border-dark-600 dark:bg-dark-800"
              >
                <button
                  v-for="user in allowlistSearchResults"
                  :key="user.id"
                  type="button"
                  class="block w-full px-3 py-2 text-left text-sm hover:bg-gray-100 disabled:opacity-50 dark:hover:bg-dark-700"
                  :disabled="allowlistedUserIds.has(user.id) || allowlistAddingId === user.id"
                  @mousedown.prevent="addAllowlistFromSearch(user)"
                >
                  <span class="font-mono text-gray-500">#{{ user.id }}</span>
                  <span class="ml-2 font-medium text-gray-900 dark:text-white">{{ user.username || user.email }}</span>
                  <span v-if="user.username && user.email" class="ml-2 text-xs text-gray-500">{{ user.email }}</span>
                  <span v-if="allowlistedUserIds.has(user.id)" class="ml-2 text-xs text-emerald-600 dark:text-emerald-400">已在白名单</span>
                  <span v-else-if="allowlistAddingId === user.id" class="ml-2 text-xs text-gray-500">添加中…</span>
                </button>
              </div>
            </div>
            <p v-if="allowlistSearchQuery && allowlistSearchResults.length === 0 && !allowlistSearchLoading" class="mt-1 text-xs text-amber-600 dark:text-amber-400">
              没有匹配的用户
            </p>
            <p v-else class="mt-1 text-[11px] text-gray-400">
              点击搜索结果直接加入白名单；白名单用户后续所有请求跳过敏感词检测
            </p>
          </div>
          <div v-if="allowlistLoading" class="py-4 text-center text-sm text-gray-400">加载中...</div>
          <div v-else-if="allowlistDetail.length === 0" class="rounded-lg border border-dashed border-gray-300 py-8 text-center text-sm text-gray-500 dark:border-dark-600">
            白名单为空
          </div>
          <div v-else class="overflow-x-auto">
            <table class="w-full min-w-[480px] divide-y divide-gray-200 dark:divide-dark-700">
              <thead class="bg-gray-50 dark:bg-dark-800">
                <tr>
                  <th class="table-th">User ID</th>
                  <th class="table-th">邮箱</th>
                  <th class="table-th text-right">操作</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-gray-100 dark:divide-dark-700">
                <tr v-for="entry in allowlistDetail" :key="entry.user_id" class="text-sm hover:bg-gray-50 dark:hover:bg-dark-800">
                  <td class="table-td font-mono">#{{ entry.user_id }}</td>
                  <td class="table-td">{{ entry.email || '-' }}</td>
                  <td class="table-td text-right">
                    <button class="btn btn-secondary btn-sm" @click="removeFromAllowlist(entry.user_id)">
                      移除
                    </button>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>

        <!-- 规则统计面板 -->
        <div v-if="activePanel === 'stats'" class="px-4 py-4">
          <div class="mb-3 flex flex-wrap items-center gap-2">
            <span class="text-xs text-gray-500 dark:text-gray-400">时间窗口：</span>
            <button
              v-for="opt in ruleStatsRangeOptions"
              :key="opt.hours"
              type="button"
              class="rounded-md px-2.5 py-1 text-xs font-medium"
              :class="ruleStatsHours === opt.hours ? 'bg-gray-900 text-white dark:bg-white dark:text-gray-900' : 'bg-gray-100 text-gray-600 hover:bg-gray-200 dark:bg-dark-700 dark:text-gray-300'"
              @click="setRuleStatsHours(opt.hours)"
            >
              {{ opt.label }}
            </button>
            <span class="ml-auto text-xs text-gray-400">显示 top {{ ruleStats.length }}，按命中数降序</span>
          </div>
          <p class="mb-3 rounded-md bg-blue-50 px-3 py-2 text-xs text-blue-700 dark:bg-blue-900/20 dark:text-blue-300">
            提示：如果某条规则的 AI Pass 率很高 → 大概率是误报源头，建议把它从 block 改为 warn 或删除
          </p>
          <div v-if="ruleStatsLoading" class="py-4 text-center text-sm text-gray-400">加载中...</div>
          <div v-else-if="ruleStats.length === 0" class="rounded-lg border border-dashed border-gray-300 py-8 text-center text-sm text-gray-500 dark:border-dark-600">
            时间窗口内无规则命中数据
          </div>
          <div v-else class="overflow-x-auto">
            <table class="w-full min-w-[920px] divide-y divide-gray-200 dark:divide-dark-700">
              <thead class="bg-gray-50 dark:bg-dark-800">
                <tr>
                  <th class="table-th">规则词</th>
                  <th class="table-th">来源</th>
                  <th class="table-th text-right">命中</th>
                  <th class="table-th text-right">block</th>
                  <th class="table-th text-right">warn</th>
                  <th class="table-th text-right">已归档</th>
                  <th class="table-th text-right">AI pass</th>
                  <th class="table-th text-right">AI reject</th>
                  <th class="table-th text-right">AI flag</th>
                  <th class="table-th text-right">误报率*</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-gray-100 dark:divide-dark-700">
                <tr v-for="row in ruleStats" :key="row.rule_word + row.rule_source" class="text-sm hover:bg-gray-50 dark:hover:bg-dark-800">
                  <td class="table-td">
                    <div class="font-mono text-xs">{{ row.rule_word }}</div>
                  </td>
                  <td class="table-td">
                    <span class="badge" :class="row.rule_source === 'custom' ? 'badge-blue' : 'badge-gray'">
                      {{ row.rule_source === 'custom' ? '自定义' : '内置' }}
                    </span>
                  </td>
                  <td class="table-td text-right font-mono tabular-nums">{{ row.total_hits }}</td>
                  <td class="table-td text-right font-mono tabular-nums">{{ row.blocked_count }}</td>
                  <td class="table-td text-right font-mono tabular-nums">{{ row.warned_count }}</td>
                  <td class="table-td text-right font-mono tabular-nums">{{ row.cleared_count }}</td>
                  <td class="table-td text-right font-mono tabular-nums text-emerald-600 dark:text-emerald-400">{{ row.ai_pass_count }}</td>
                  <td class="table-td text-right font-mono tabular-nums text-red-600 dark:text-red-400">{{ row.ai_reject_count }}</td>
                  <td class="table-td text-right font-mono tabular-nums text-amber-600 dark:text-amber-400">{{ row.ai_flag_count }}</td>
                  <td class="table-td text-right">
                    <span :class="['font-mono tabular-nums', falseRateClass(row)]">
                      {{ falseRate(row) }}
                    </span>
                  </td>
                </tr>
              </tbody>
            </table>
            <p class="mt-2 text-[11px] text-gray-400">
              * 误报率 = AI pass / 总命中。30%+ 说明该规则容易误报；启用 AI 审核后才有数据。
            </p>
          </div>
        </div>
      </section>

      <section class="surface-card overflow-hidden">
        <div class="overflow-x-auto px-4 pt-4">
          <table class="w-full min-w-[960px] divide-y divide-gray-200 dark:divide-dark-700">
            <thead class="bg-gray-50 dark:bg-dark-800">
              <tr>
                <th class="table-th">时间</th>
                <th class="table-th">用户 / Key</th>
                <th class="table-th">路径</th>
                <th class="table-th">命中规则</th>
                <th class="table-th">AI 审核</th>
                <th class="table-th">状态</th>
                <th class="table-th">请求预览</th>
                <th class="table-th text-right">操作</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-gray-100 bg-white dark:divide-dark-700 dark:bg-dark-900">
              <tr v-if="loading">
                <td colspan="8" class="px-4 py-16 text-center text-sm text-gray-500 dark:text-gray-400">加载中...</td>
              </tr>
              <tr v-else-if="events.length === 0">
                <td colspan="8">
                  <div class="risk-empty-state">
                    <div class="risk-empty-icon">
                      <Icon name="inbox" size="xl" />
                    </div>
                    <div class="text-base font-semibold text-gray-800 dark:text-gray-100">暂无风控日志</div>
                    <div class="mt-1 text-sm text-gray-500 dark:text-gray-400">当前没有符合条件的日志记录</div>
                  </div>
                </td>
              </tr>
              <template v-else>
                <tr v-for="item in events" :key="item.id" class="transition hover:bg-gray-50 dark:hover:bg-dark-800/70">
                  <td class="table-td whitespace-nowrap">
                    <div class="text-sm text-gray-900 dark:text-gray-100">{{ formatDate(item.created_at) }}</div>
                    <div class="text-xs text-gray-500">#{{ item.id }}</div>
                  </td>
                  <td class="table-td min-w-[190px]">
                    <div class="text-sm font-medium text-gray-900 dark:text-gray-100">
                      {{ item.user_email || formatID('用户', item.user_id) }}
                    </div>
                    <div class="text-xs text-gray-500 dark:text-gray-400">
                      {{ item.api_key_name || formatID('Key', item.api_key_id) }}
                    </div>
                  </td>
                  <td class="table-td min-w-[190px]">
                    <div class="text-sm text-gray-900 dark:text-gray-100">{{ item.method }} {{ item.path }}</div>
                    <div class="text-xs text-gray-500 dark:text-gray-400">{{ item.client_ip || '-' }}</div>
                  </td>
                  <td class="table-td min-w-[180px]">
                    <div class="flex flex-wrap items-center gap-1.5">
                      <span class="badge" :class="item.rule_source === 'custom' ? 'badge-blue' : 'badge-gray'">
                        {{ sourceText(item.rule_source) }}
                      </span>
                      <span class="badge badge-amber">{{ item.action || 'blocked' }}</span>
                    </div>
                    <div class="mt-1 text-sm text-gray-900 dark:text-gray-100">{{ item.rule_word || '-' }}</div>
                    <div class="text-xs text-gray-500 dark:text-gray-400">{{ item.rule_path || '-' }}</div>
                  </td>
                  <td class="table-td whitespace-nowrap">
                    <!-- AI 审核结果：未审核显示灰；已审核解析 ai_review_result JSON 显示彩色 verdict badge
                         pass=绿 / flag=黄 / reject=红 / raw=灰 / error=红边框
                         hover 时通过 title 显示完整 reason -->
                    <template v-if="!item.ai_reviewed">
                      <span class="badge badge-gray">未经过 AI</span>
                    </template>
                    <template v-else>
                      <span :class="['badge', aiVerdictBadgeClass(aiParse(item.ai_review_result).verdict)]" :title="aiParse(item.ai_review_result).reason">
                        {{ aiVerdictLabel(aiParse(item.ai_review_result).verdict) }}
                      </span>
                      <div v-if="aiParse(item.ai_review_result).reason" class="mt-1 max-w-[180px] truncate text-xs text-gray-500 dark:text-gray-400" :title="aiParse(item.ai_review_result).reason">
                        {{ aiParse(item.ai_review_result).reason }}
                      </div>
                      <div v-if="aiParse(item.ai_review_result).category" class="text-[10px] text-gray-400">
                        {{ aiParse(item.ai_review_result).category }}
                      </div>
                    </template>
                  </td>
                  <td class="table-td whitespace-nowrap">
                    <span class="badge" :class="statusClass(item.status)">{{ statusText(item.status) }}</span>
                    <div v-if="item.reviewed_at" class="mt-1 text-xs text-gray-500">{{ formatDate(item.reviewed_at) }}</div>
                  </td>
                  <td class="table-td max-w-[420px] align-top">
                    <!-- 行内仍 line-clamp-3 保持表格紧凑；点击"查看完整内容"弹独立 modal 显示
                         完整 prompt（后端已截到 1000 字符）。比展开整行更不打乱表格布局。 -->
                    <div v-if="item.prompt_preview" class="text-sm text-gray-700 dark:text-gray-300">
                      <div class="line-clamp-3">{{ item.prompt_preview }}</div>
                      <button
                        v-if="(item.prompt_preview?.length || 0) > 80"
                        type="button"
                        class="mt-1 inline-flex items-center gap-0.5 text-xs font-medium text-primary-600 hover:text-primary-700 hover:underline dark:text-primary-400"
                        @click="openPreviewDialog(item)"
                      >
                        <Icon name="eye" size="xs" /> 查看完整内容
                      </button>
                    </div>
                    <div v-else class="text-sm text-gray-400">-</div>
                    <div v-if="item.request_id || item.client_request_id" class="mt-1 text-xs text-gray-500">
                      {{ item.request_id || item.client_request_id }}
                    </div>
                  </td>
                  <td class="table-td whitespace-nowrap text-right">
                    <button
                      v-if="item.status !== 'reviewed'"
                      class="btn btn-secondary btn-sm mr-2"
                      :disabled="loading"
                      @click="markReviewed(item)"
                    >
                      标记复核
                    </button>
                    <!-- AI 复核：仅对"未经过 AI"事件显示，admin 点击后同步跑一次审核
                         结果直接写回事件，pass 自动归档；进行中按钮 disabled + 文案切换 -->
                    <button
                      v-if="!item.ai_reviewed"
                      class="btn btn-secondary btn-sm mr-2"
                      :disabled="loading || aiReviewingIds.has(item.id)"
                      :title="'用 AI 模型对该事件做二次审核（消耗配置的 ApiKey 额度）'"
                      @click="triggerAIReview(item)"
                    >
                      {{ aiReviewingIds.has(item.id) ? 'AI 审核中…' : 'AI 复核' }}
                    </button>
                    <button
                      v-if="item.user_id && !allowlistedUserIds.has(item.user_id)"
                      class="btn btn-secondary btn-sm mr-2"
                      :disabled="loading"
                      :title="'加入白名单：该用户后续请求跳过敏感词检测'"
                      @click="addToAllowlist(item)"
                    >
                      加入白名单
                    </button>
                    <span
                      v-else-if="item.user_id && allowlistedUserIds.has(item.user_id)"
                      class="mr-2 inline-flex items-center rounded bg-emerald-100 px-2 py-0.5 text-xs font-medium text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-300"
                      title="该用户已在风控白名单"
                    >
                      已白名单
                    </span>
                    <button
                      v-if="item.user_id && item.status !== 'cleared'"
                      class="btn btn-danger btn-sm"
                      :disabled="loading"
                      @click="openClearUser(item)"
                    >
                      清空用户
                    </button>
                  </td>
                </tr>
              </template>
            </tbody>
          </table>
        </div>

        <div v-if="pagination.total > 0" class="border-t border-gray-100 px-4 py-3 dark:border-dark-700">
          <Pagination
            :page="pagination.page"
            :total="pagination.total"
            :page-size="pagination.page_size"
            @update:page="handlePageChange"
            @update:pageSize="handlePageSizeChange"
          />
        </div>
      </section>
    </div>
  </AppLayout>

  <ConfirmDialog
    :show="clearDialogVisible"
    title="清空用户警告"
    :message="clearDialogMessage"
    confirm-text="确认清空"
    cancel-text="取消"
    danger
    @confirm="clearSelectedUser"
    @cancel="clearDialogVisible = false"
  />

  <!-- 请求预览完整内容 modal：复核违规时需要完整 prompt，弹窗比行内展开更不打乱表格布局 -->
  <BaseDialog
    :show="!!previewDialogEvent"
    :title="previewDialogTitle"
    width="wide"
    @close="previewDialogEvent = null"
  >
    <div v-if="previewDialogEvent" class="space-y-3">
      <!-- 元信息条 -->
      <div class="grid grid-cols-1 gap-2 rounded-lg bg-gray-50 p-3 text-xs dark:bg-dark-800 sm:grid-cols-2">
        <div>
          <span class="text-gray-500">用户：</span>
          <span class="font-medium">{{ previewDialogEvent.user_email || '-' }}</span>
        </div>
        <div>
          <span class="text-gray-500">时间：</span>
          <span class="font-medium">{{ formatDate(previewDialogEvent.created_at) }}</span>
        </div>
        <div>
          <span class="text-gray-500">路径：</span>
          <span class="font-mono">{{ previewDialogEvent.method }} {{ previewDialogEvent.path }}</span>
        </div>
        <div>
          <span class="text-gray-500">命中规则：</span>
          <span class="font-medium">{{ previewDialogEvent.rule_word || '-' }}</span>
          <span v-if="previewDialogEvent.rule_path" class="ml-1 font-mono text-gray-500">{{ previewDialogEvent.rule_path }}</span>
        </div>
        <div v-if="previewDialogEvent.request_id || previewDialogEvent.client_request_id" class="sm:col-span-2">
          <span class="text-gray-500">Request ID：</span>
          <span class="font-mono">{{ previewDialogEvent.request_id || previewDialogEvent.client_request_id }}</span>
        </div>
      </div>
      <!-- 完整 prompt 内容 -->
      <div>
        <div class="mb-1.5 flex items-center justify-between">
          <label class="text-xs font-medium text-gray-500 dark:text-gray-400">请求内容（后端已截至 1000 字符）</label>
          <button
            type="button"
            class="inline-flex items-center gap-1 text-xs text-gray-500 hover:text-primary-600 dark:text-gray-400 dark:hover:text-primary-400"
            @click="copyPreviewToClipboard"
          >
            <Icon name="copy" size="xs" />
            {{ previewCopied ? '已复制' : '复制' }}
          </button>
        </div>
        <pre class="max-h-[60vh] overflow-auto whitespace-pre-wrap break-words rounded-lg bg-gray-50 p-3 text-sm text-gray-800 dark:bg-dark-800 dark:text-gray-200">{{ previewDialogEvent.prompt_preview || '-' }}</pre>
      </div>
    </div>
    <template #footer>
      <div class="flex justify-end">
        <button class="btn btn-secondary" @click="previewDialogEvent = null">关闭</button>
      </div>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import AppLayout from '@/components/layout/AppLayout.vue'
import Pagination from '@/components/common/Pagination.vue'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Select, { type SelectOption } from '@/components/common/Select.vue'
import Icon from '@/components/icons/Icon.vue'
import { useAppStore } from '@/stores/app'
import {
  clearSafetyRiskEventsForUser,
  listSafetyRiskEvents,
  reviewSafetyRiskEvent,
  listSafetyAllowlist,
  addSafetyAllowlist,
  removeSafetyAllowlist,
  listSafetyRiskRuleStats,
  reviewSafetyRiskEventWithAI,
  type SafetyRiskEvent,
  type SafetyRiskQueryParams,
  type SafetyRiskRuleStat,
} from '@/api/admin/safetyRisk'
import { adminAPI } from '@/api/admin'
import { getPersistedPageSize } from '@/composables/usePersistedPageSize'

type StatIcon = 'document' | 'shield' | 'brain' | 'checkCircle' | 'xCircle'

interface RiskStat {
  key: string
  label: string
  value: number
  hint: string
  icon: StatIcon
  iconClass: string
}

const appStore = useAppStore()
const loading = ref(false)
const events = ref<SafetyRiskEvent[]>([])
// 请求预览弹窗：admin 点击行的"查看完整内容"按钮时弹出展示完整 prompt + 关键元信息
const previewDialogEvent = ref<SafetyRiskEvent | null>(null)
const previewCopied = ref(false)
const previewDialogTitle = computed(() => {
  if (!previewDialogEvent.value) return '请求预览'
  return `请求预览 · #${previewDialogEvent.value.id}`
})
function openPreviewDialog(item: SafetyRiskEvent) {
  previewDialogEvent.value = item
  previewCopied.value = false
}
async function copyPreviewToClipboard() {
  const content = previewDialogEvent.value?.prompt_preview || ''
  if (!content) return
  try {
    await navigator.clipboard.writeText(content)
    previewCopied.value = true
    setTimeout(() => { previewCopied.value = false }, 1500)
  } catch {
    /* clipboard 权限被拒，静默处理 */
  }
}
// 风控白名单缓存：在白名单的 user 后续请求直接跳过 sensitive_filter；
// 列表中行级 "加入白名单" / "已白名单" 角标基于这个 Set 渲染
const allowlistedUserIds = ref<Set<number>>(new Set())
async function refreshAllowlist() {
  try {
    const ids = await listSafetyAllowlist()
    allowlistedUserIds.value = new Set(ids)
  } catch {
    /* 拉失败不阻断主流程；事件列表本身仍能用 */
  }
}
async function addToAllowlist(event: SafetyRiskEvent) {
  if (!event.user_id || event.user_id <= 0) return
  try {
    await addSafetyAllowlist(event.user_id)
    allowlistedUserIds.value = new Set([...allowlistedUserIds.value, event.user_id])
    appStore.showSuccess(`已把 ${event.user_email || `用户#${event.user_id}`} 加入风控白名单，后续请求将跳过敏感词检测`)
  } catch (err) {
    appStore.showError('加入白名单失败：' + (err as Error).message)
  }
}
// 行级 AI 复核：admin 在事件列表点"AI 复核"按钮同步触发，期间该行 disabled
const aiReviewingIds = ref<Set<number>>(new Set())
const batchAIReviewing = ref(false)
const batchProgress = ref({ done: 0, total: 0 })

// 当前页内还没经过 AI 复核的事件数（驱动顶部"批量复核"按钮显示和计数）
const pendingAIReviewCount = computed(() => events.value.filter((e) => !e.ai_reviewed).length)

async function triggerAIReview(item: SafetyRiskEvent) {
  if (aiReviewingIds.value.has(item.id)) return
  aiReviewingIds.value = new Set([...aiReviewingIds.value, item.id])
  try {
    const result = await reviewSafetyRiskEventWithAI(item.id)
    // 局部更新该行，避免重新拉整页
    item.ai_reviewed = true
    item.ai_review_result = JSON.stringify(result)
    if (result.verdict === 'pass') {
      item.status = 'cleared'
    }
    appStore.showSuccess(`AI 判定：${result.verdict}${result.reason ? '（' + result.reason + '）' : ''}`)
  } catch (err) {
    const e = err as { response?: { data?: { message?: string } }; message?: string }
    appStore.showError('AI 复核失败：' + (e.response?.data?.message || e.message || '未知错误'))
  } finally {
    const next = new Set(aiReviewingIds.value)
    next.delete(item.id)
    aiReviewingIds.value = next
  }
}

// 批量 AI 复核：顺序执行（不并发），避免触发 ApiKey 并发限流，每条之间 500ms 间隔
async function batchAIReview() {
  const pending = events.value.filter((e) => !e.ai_reviewed)
  if (pending.length === 0 || batchAIReviewing.value) return
  batchAIReviewing.value = true
  batchProgress.value = { done: 0, total: pending.length }
  for (const item of pending) {
    await triggerAIReview(item)
    batchProgress.value.done += 1
    if (batchProgress.value.done < batchProgress.value.total) {
      await new Promise((r) => setTimeout(r, 500))
    }
  }
  batchAIReviewing.value = false
  appStore.showSuccess(`批量 AI 复核完成，共处理 ${batchProgress.value.total} 条`)
}

// 折叠面板：'' 表示都收起，'allowlist' / 'stats' 分别打开两个管理面板。
// 切换 panel 时按需 lazy load 数据，避免每次进风控页就拉多个 API。
type AdminPanel = '' | 'allowlist' | 'stats'
const activePanel = ref<AdminPanel>('')
const allowlistDetail = ref<Array<{ user_id: number; email: string }>>([])
const allowlistLoading = ref(false)
const ruleStats = ref<SafetyRiskRuleStat[]>([])
const ruleStatsLoading = ref(false)
const ruleStatsHours = ref(168)
const ruleStatsRangeOptions = [
  { hours: 24, label: '24h' },
  { hours: 72, label: '3d' },
  { hours: 168, label: '7d' },
  { hours: 720, label: '30d' },
]

function togglePanel(panel: AdminPanel) {
  if (activePanel.value === panel) {
    activePanel.value = ''
    return
  }
  activePanel.value = panel
  if (panel === 'allowlist') {
    loadAllowlistDetail()
  } else if (panel === 'stats') {
    loadRuleStats()
  }
}

// 白名单"直接添加用户"搜索：输入 ID/邮箱/用户名 → 防抖 300ms → 拉 admin users API
// 复用独享池赠送同款 UX（dropdown 列表 + 已在白名单状态提示）
const allowlistSearchQuery = ref('')
const allowlistSearchResults = ref<Array<{ id: number; email?: string; username?: string }>>([])
const allowlistSearchLoading = ref(false)
const allowlistSearchFocused = ref(false)
const allowlistAddingId = ref<number | null>(null)
let allowlistSearchTimer: ReturnType<typeof setTimeout> | null = null
let allowlistSearchSeq = 0

function onAllowlistSearchInput() {
  if (allowlistSearchTimer) clearTimeout(allowlistSearchTimer)
  const q = allowlistSearchQuery.value.trim()
  if (!q) {
    allowlistSearchResults.value = []
    return
  }
  allowlistSearchLoading.value = true
  const seq = ++allowlistSearchSeq
  allowlistSearchTimer = setTimeout(async () => {
    try {
      const res = await adminAPI.users.list(1, 8, { search: q })
      if (seq === allowlistSearchSeq) {
        allowlistSearchResults.value = (res?.items || []).map((u) => ({
          id: u.id,
          email: u.email,
          username: u.username,
        }))
      }
    } catch {
      if (seq === allowlistSearchSeq) allowlistSearchResults.value = []
    } finally {
      if (seq === allowlistSearchSeq) allowlistSearchLoading.value = false
    }
  }, 300)
}

function onAllowlistSearchBlur() {
  // 延迟关闭让 mousedown 选项点击事件先触发
  setTimeout(() => { allowlistSearchFocused.value = false }, 200)
}

async function addAllowlistFromSearch(user: { id: number; email?: string; username?: string }) {
  if (allowlistedUserIds.value.has(user.id) || allowlistAddingId.value === user.id) return
  allowlistAddingId.value = user.id
  try {
    await addSafetyAllowlist(user.id)
    allowlistedUserIds.value = new Set([...allowlistedUserIds.value, user.id])
    // 局部更新列表，不重新拉整个 detail（避免抖动）
    allowlistDetail.value = [
      ...allowlistDetail.value,
      { user_id: user.id, email: user.email || '' },
    ]
    appStore.showSuccess(`已把 ${user.username || user.email || `用户#${user.id}`} 加入风控白名单`)
    // 清空搜索状态方便继续加下一个
    allowlistSearchQuery.value = ''
    allowlistSearchResults.value = []
  } catch (err) {
    appStore.showError('加入白名单失败：' + (err as Error).message)
  } finally {
    allowlistAddingId.value = null
  }
}

// 拉白名单的详细信息：先 listSafetyAllowlist 拿 user_id 数组，
// 再用 admin users API 批量取邮箱（per-id 单独查，简单实现；规模 < 100 用户 OK）
async function loadAllowlistDetail() {
  allowlistLoading.value = true
  try {
    const ids = await listSafetyAllowlist()
    allowlistedUserIds.value = new Set(ids)
    if (ids.length === 0) {
      allowlistDetail.value = []
      return
    }
    const results = await Promise.allSettled(
      ids.map((id) => adminAPI.users.getById(id)),
    )
    allowlistDetail.value = ids.map((id, idx) => {
      const res = results[idx]
      const email = res.status === 'fulfilled' ? res.value?.email || '' : ''
      return { user_id: id, email }
    })
  } catch {
    allowlistDetail.value = []
  } finally {
    allowlistLoading.value = false
  }
}

async function removeFromAllowlist(userId: number) {
  try {
    await removeSafetyAllowlist(userId)
    allowlistDetail.value = allowlistDetail.value.filter((e) => e.user_id !== userId)
    const next = new Set(allowlistedUserIds.value)
    next.delete(userId)
    allowlistedUserIds.value = next
    appStore.showSuccess(`已把用户 #${userId} 从白名单移除`)
  } catch (err) {
    appStore.showError('移除失败：' + (err as Error).message)
  }
}

// 规则统计：按时间窗口拉
async function loadRuleStats() {
  ruleStatsLoading.value = true
  try {
    const res = await listSafetyRiskRuleStats(ruleStatsHours.value, 50)
    ruleStats.value = res.stats || []
  } catch {
    ruleStats.value = []
  } finally {
    ruleStatsLoading.value = false
  }
}
function setRuleStatsHours(hours: number) {
  ruleStatsHours.value = hours
  loadRuleStats()
}

// 误报率：AI pass / 总命中。30%+ 标红，10-30% 标黄，< 10% 灰
function falseRate(row: SafetyRiskRuleStat): string {
  if (!row.total_hits) return '-'
  const aiTotal = row.ai_pass_count + row.ai_reject_count + row.ai_flag_count
  if (aiTotal === 0) return '—（未启用 AI）'
  const rate = row.ai_pass_count / aiTotal
  return `${(rate * 100).toFixed(0)}%`
}
function falseRateClass(row: SafetyRiskRuleStat): string {
  if (!row.total_hits) return 'text-gray-400'
  const aiTotal = row.ai_pass_count + row.ai_reject_count + row.ai_flag_count
  if (aiTotal === 0) return 'text-gray-400'
  const rate = row.ai_pass_count / aiTotal
  if (rate >= 0.3) return 'text-red-600 dark:text-red-400 font-semibold'
  if (rate >= 0.1) return 'text-amber-600 dark:text-amber-400'
  return 'text-gray-500 dark:text-gray-400'
}

const selectedClearEvent = ref<SafetyRiskEvent | null>(null)
const clearDialogVisible = ref(false)
const userIdInput = ref('')

const filters = reactive<SafetyRiskQueryParams>({
  status: 'pending',
  time_range: '24h',
  q: '',
})

const pagination = reactive({
  page: 1,
  page_size: getPersistedPageSize(),
  total: 0,
})

const statusFilterOptions: SelectOption[] = [
  { value: '', label: '全部' },
  { value: 'pending', label: '待复核' },
  { value: 'reviewed', label: '已复核' },
  { value: 'cleared', label: '已清空' },
]

const timeRangeOptions: SelectOption[] = [
  { value: '5m', label: '最近 5 分钟' },
  { value: '30m', label: '最近 30 分钟' },
  { value: '1h', label: '最近 1 小时' },
  { value: '6h', label: '最近 6 小时' },
  { value: '24h', label: '最近 24 小时' },
  { value: '7d', label: '最近 7 天' },
  { value: '30d', label: '最近 30 天' },
]

const selectedStatus = computed<string | number | boolean | null>({
  get: () => filters.status ?? '',
  set: (value) => {
    filters.status = typeof value === 'string' ? value : ''
  },
})

const selectedTimeRange = computed<string | number | boolean | null>({
  get: () => filters.time_range ?? '24h',
  set: (value) => {
    const nextValue = typeof value === 'string' ? value : '24h'
    filters.time_range = nextValue as SafetyRiskQueryParams['time_range']
  },
})

const riskStats = computed<RiskStat[]>(() => [
  {
    key: 'total',
    label: '总拦截数',
    value: pagination.total,
    hint: timeRangeText(filters.time_range),
    icon: 'document',
    iconClass: 'bg-blue-50 text-blue-600 dark:bg-blue-900/30 dark:text-blue-300',
  },
  {
    key: 'pending',
    label: '待复核',
    value: countByStatus('pending'),
    hint: '当前页',
    icon: 'shield',
    iconClass: 'bg-violet-50 text-violet-600 dark:bg-violet-900/30 dark:text-violet-300',
  },
  {
    key: 'ai',
    label: 'AI 审核中',
    value: events.value.filter(isAIReviewing).length,
    hint: '当前页',
    icon: 'brain',
    iconClass: 'bg-orange-50 text-orange-600 dark:bg-orange-900/30 dark:text-orange-300',
  },
  {
    key: 'approved',
    label: '已通过',
    value: events.value.filter(isApproved).length,
    hint: '当前页',
    icon: 'checkCircle',
    iconClass: 'bg-emerald-50 text-emerald-600 dark:bg-emerald-900/30 dark:text-emerald-300',
  },
  {
    key: 'rejected',
    label: '已拒绝',
    value: events.value.filter(isRejected).length,
    hint: '当前页',
    icon: 'xCircle',
    iconClass: 'bg-rose-50 text-rose-600 dark:bg-rose-900/30 dark:text-rose-300',
  },
])

const clearDialogMessage = computed(() => {
  const item = selectedClearEvent.value
  if (!item) return ''
  const user = item.user_email || formatID('用户', item.user_id)
  return `将把 ${user} 的未清空风控警告全部标记为已清空。该操作只清空警告记录，不会删除日志。`
})

async function loadEvents() {
  loading.value = true
  try {
    const params: SafetyRiskQueryParams = {
      ...filters,
      page: pagination.page,
      page_size: pagination.page_size,
    }
    const userId = Number(userIdInput.value)
    if (Number.isFinite(userId) && userId > 0) {
      params.user_id = userId
    }
    if (!params.q) delete params.q
    const data = await listSafetyRiskEvents(params)
    events.value = data.items
    pagination.total = data.total
    pagination.page = data.page
    pagination.page_size = data.page_size
  } catch (error: any) {
    appStore.showError(error?.message || '加载风控日志失败')
  } finally {
    loading.value = false
  }
}

function applyFilters() {
  pagination.page = 1
  loadEvents()
}

function resetFilters() {
  filters.status = 'pending'
  filters.time_range = '24h'
  filters.q = ''
  userIdInput.value = ''
  applyFilters()
}

async function markReviewed(item: SafetyRiskEvent) {
  try {
    await reviewSafetyRiskEvent(item.id, {
      status: 'reviewed',
      review_note: 'manual reviewed',
    })
    appStore.showSuccess('已标记为复核')
    await loadEvents()
  } catch (error: any) {
    appStore.showError(error?.message || '复核失败')
  }
}

function openClearUser(item: SafetyRiskEvent) {
  selectedClearEvent.value = item
  clearDialogVisible.value = true
}

async function clearSelectedUser() {
  const item = selectedClearEvent.value
  if (!item?.user_id) {
    clearDialogVisible.value = false
    return
  }
  try {
    const result = await clearSafetyRiskEventsForUser(item.user_id, 'manual clear after review')
    appStore.showSuccess(`已清空 ${result.cleared} 条警告记录`)
    clearDialogVisible.value = false
    selectedClearEvent.value = null
    await loadEvents()
  } catch (error: any) {
    appStore.showError(error?.message || '清空用户警告失败')
  }
}

function handlePageChange(page: number) {
  pagination.page = page
  loadEvents()
}

function handlePageSizeChange(pageSize: number) {
  pagination.page = 1
  pagination.page_size = pageSize
  loadEvents()
}

function countByStatus(status: string): number {
  return events.value.filter((item) => item.status === status).length
}

function normalizeReviewResult(value?: string | null): string {
  return String(value || '').trim().toLowerCase()
}

function isAIReviewing(item: SafetyRiskEvent): boolean {
  const result = normalizeReviewResult(item.ai_review_result)
  return item.ai_reviewed && item.status === 'pending' && (!result || result === 'pending' || result === 'reviewing')
}

function isApproved(item: SafetyRiskEvent): boolean {
  const status = String(item.status || '').toLowerCase()
  const result = normalizeReviewResult(item.ai_review_result)
  return ['approved', 'passed', 'allowed'].includes(status) || ['approve', 'approved', 'pass', 'passed', 'allow', 'allowed'].includes(result)
}

function isRejected(item: SafetyRiskEvent): boolean {
  const status = String(item.status || '').toLowerCase()
  const result = normalizeReviewResult(item.ai_review_result)
  return ['rejected', 'blocked', 'denied'].includes(status) || ['reject', 'rejected', 'block', 'blocked', 'deny', 'denied'].includes(result)
}

function timeRangeText(range?: string): string {
  switch (range) {
    case '5m':
      return '最近 5 分钟'
    case '30m':
      return '最近 30 分钟'
    case '1h':
      return '最近 1 小时'
    case '6h':
      return '最近 6 小时'
    case '7d':
      return '最近 7 天'
    case '30d':
      return '最近 30 天'
    case '24h':
    default:
      return '最近 24 小时'
  }
}

function formatDate(value?: string | null): string {
  if (!value) return '-'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value
  return date.toLocaleString()
}

function formatID(label: string, value?: number | null): string {
  return value ? `${label} #${value}` : '-'
}

function sourceText(source: string): string {
  if (source === 'custom') return '自定义'
  if (source === 'builtin') return '内置'
  if (source === 'ai') return 'AI'
  return source || '本地'
}

function statusText(status: string): string {
  switch (status) {
    case 'pending':
      return '待复核'
    case 'reviewed':
      return '已复核'
    case 'cleared':
      return '已清空'
    case 'rejected':
      return '已拒绝'
    case 'approved':
    case 'passed':
      return '已通过'
    default:
      return status || '-'
  }
}

function statusClass(status: string): string {
  switch (status) {
    case 'pending':
      return 'badge-amber'
    case 'reviewed':
      return 'badge-blue'
    case 'cleared':
    case 'approved':
    case 'passed':
      return 'badge-green'
    case 'rejected':
      return 'badge-red'
    default:
      return 'badge-gray'
  }
}

// aiParse 解析 ai_review_result（JSON 字符串）为对象。
// 兼容：1) 新格式 {"verdict","category","reason"} 2) 旧格式 raw 字符串（兜底显示）
// 失败时 verdict='' 表示未识别
interface AIReviewParsed {
  verdict: string
  category: string
  reason: string
  raw?: string
}
function aiParse(raw: string | null | undefined): AIReviewParsed {
  if (!raw) return { verdict: '', category: '', reason: '' }
  const s = raw.trim()
  if (!s || s === 'not_used') return { verdict: '', category: '', reason: '' }
  // 尝试 JSON 解析
  if (s.startsWith('{')) {
    try {
      const obj = JSON.parse(s)
      return {
        verdict: String(obj.verdict || '').toLowerCase(),
        category: String(obj.category || ''),
        reason: String(obj.reason || ''),
        raw: obj.raw,
      }
    } catch {
      /* 落到 fallback */
    }
  }
  // 旧格式 fallback：'pass|category|reason' 拼接（之前版本格式）
  if (s.includes('|')) {
    const [verdict, category, ...rest] = s.split('|')
    return {
      verdict: (verdict || '').toLowerCase(),
      category: category || '',
      reason: rest.join('|') || '',
    }
  }
  // 兜底：错误信息 / raw string
  return { verdict: 'raw', category: '', reason: s }
}

function aiVerdictLabel(verdict: string): string {
  switch (verdict) {
    case 'pass':
      return '通过'
    case 'flag':
      return '需复核'
    case 'reject':
      return '违规'
    case 'raw':
      return '原始'
    case 'error':
      return '错误'
    default:
      return '已经过 AI'
  }
}

function aiVerdictBadgeClass(verdict: string): string {
  switch (verdict) {
    case 'pass':
      return 'badge-green'
    case 'flag':
      return 'badge-amber'
    case 'reject':
      return 'badge-red'
    case 'error':
      return 'badge-red'
    default:
      return 'badge-gray'
  }
}

onMounted(() => {
  loadEvents()
  refreshAllowlist()
})
</script>

<style scoped>
/* 风控页全局：所有 .card 用兑换码同款远距离阴影，与全站视觉调性统一 */
.risk-page :deep(.card) {
  box-shadow: 0 18px 44px -34px rgb(15 23 42 / 0.55);
}

:global(:root.dark) .risk-page :deep(.card) {
  box-shadow: none;
}

/* 信息提示条：sky 调，与登录注册的辅助提示一致 */
.risk-info-banner {
  display: flex;
  align-items: center;
  gap: 0.625rem;
  border: 1px solid rgb(186 230 253);
  border-radius: 0.5rem;
  background: rgb(240 249 255 / 0.7);
  padding: 0.875rem 1rem;
  color: rgb(2 132 199);
  font-size: 0.875rem;
  line-height: 1.6;
}

:global(:root.dark) .risk-info-banner {
  border-color: rgb(7 89 133 / 0.6);
  background: rgb(8 47 73 / 0.25);
  color: rgb(125 211 252);
}

.risk-info-icon {
  display: inline-flex;
  flex: 0 0 auto;
}

.filter-label {
  display: block;
  color: rgb(55 65 81);
  font-size: 0.8125rem;
  font-weight: 500;
}

.dark .filter-label {
  color: rgb(209 213 219);
}

.risk-select :deep(.select-trigger) {
  height: 2.75rem;
  border-color: rgb(209 213 219);
  border-radius: 0.5rem;
  padding: 0.625rem 0.875rem;
  box-shadow: 0 1px 2px rgb(15 23 42 / 0.04);
}

.risk-select :deep(.select-trigger-open) {
  border-color: rgb(59 130 246);
  box-shadow: 0 0 0 3px rgb(59 130 246 / 0.14);
}

.dark .risk-select :deep(.select-trigger) {
  border-color: rgb(75 85 99);
  box-shadow: none;
}

.risk-filter-control {
  display: block;
  width: 100%;
  height: 2.75rem;
  border: 1px solid rgb(209 213 219);
  border-radius: 0.5rem;
  background-color: rgb(255 255 255);
  padding: 0.625rem 0.875rem;
  color: rgb(17 24 39);
  font-size: 0.875rem;
  line-height: 1.25rem;
  box-shadow: 0 1px 2px rgb(15 23 42 / 0.04);
  transition: border-color 0.15s ease, box-shadow 0.15s ease, background-color 0.15s ease;
}

.risk-filter-control::placeholder {
  color: rgb(156 163 175);
}

.risk-filter-control:focus {
  border-color: rgb(59 130 246);
  box-shadow: 0 0 0 3px rgb(59 130 246 / 0.14);
  outline: none;
}

.dark .risk-filter-control {
  border-color: rgb(75 85 99);
  background-color: rgb(17 24 39);
  color: rgb(243 244 246);
  box-shadow: none;
}

.dark .risk-filter-control::placeholder {
  color: rgb(107 114 128);
}

.risk-action-button,
.risk-icon-button {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  border-radius: 0.5rem;
  font-size: 0.875rem;
  font-weight: 600;
  line-height: 1.25rem;
  transition: background-color 0.15s ease, border-color 0.15s ease, color 0.15s ease, box-shadow 0.15s ease;
}

.risk-action-button {
  height: 2.5rem;
  padding: 0 1.25rem;
}

.risk-action-primary {
  border: 1px solid rgb(37 99 235);
  background: linear-gradient(180deg, rgb(37 99 235), rgb(29 78 216));
  color: white;
  box-shadow: 0 8px 18px rgb(37 99 235 / 0.18);
}

.risk-action-primary:hover:not(:disabled) {
  border-color: rgb(29 78 216);
  background: linear-gradient(180deg, rgb(29 78 216), rgb(30 64 175));
}

.risk-action-secondary,
.risk-icon-button {
  border: 1px solid rgb(209 213 219);
  background: rgb(255 255 255);
  color: rgb(55 65 81);
  box-shadow: 0 1px 2px rgb(15 23 42 / 0.04);
}

.risk-icon-button {
  padding: 0 1rem;
}

.risk-action-secondary:hover:not(:disabled),
.risk-icon-button:hover:not(:disabled) {
  border-color: rgb(156 163 175);
  background: rgb(249 250 251);
  color: rgb(17 24 39);
}

.risk-action-button:disabled,
.risk-icon-button:disabled {
  cursor: not-allowed;
  opacity: 0.65;
}

.dark .risk-action-secondary,
.dark .risk-icon-button {
  border-color: rgb(75 85 99);
  background: rgb(17 24 39);
  color: rgb(229 231 235);
  box-shadow: none;
}

.dark .risk-action-secondary:hover:not(:disabled),
.dark .risk-icon-button:hover:not(:disabled) {
  border-color: rgb(107 114 128);
  background: rgb(31 41 55);
  color: white;
}

/* metric-card 内的图标块：2.5rem 与 ops 系统资源卡同款尺寸 */
.risk-stat-icon {
  display: inline-flex;
  height: 2.5rem;
  width: 2.5rem;
  flex: 0 0 auto;
  align-items: center;
  justify-content: center;
  border-radius: 0.75rem;
}

/* metric-card 局部样式：与仪表盘统一的 Notion 风 */
.metric-card {
  display: flex;
  align-items: flex-start;
  gap: 0.75rem;
  border-radius: 1rem;
  border: 1px solid rgb(229 231 235 / 0.7);
  background: white;
  padding: 1rem;
  box-shadow: 0 1px 2px rgba(15, 23, 42, 0.04);
  transition: all 0.2s;
}
.metric-card:hover {
  border-color: rgb(209 213 219);
  box-shadow: 0 4px 16px rgba(15, 23, 42, 0.06);
  transform: translateY(-2px);
}
.dark .metric-card {
  border-color: rgb(55 65 81 / 0.6);
  background: rgb(31 41 55 / 0.4);
}
.dark .metric-card:hover {
  border-color: rgb(75 85 99);
}

.risk-empty-state {
  display: flex;
  min-height: 18rem;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 3rem 1rem;
  text-align: center;
}

.risk-empty-icon {
  display: inline-flex;
  height: 4rem;
  width: 4rem;
  align-items: center;
  justify-content: center;
  border-radius: 1rem;
  background: rgb(243 244 246);
  color: rgb(156 163 175);
  margin-bottom: 1rem;
}

.dark .risk-empty-icon {
  background: rgb(55 65 81);
  color: rgb(107 114 128);
}

.table-th {
  padding: 0.875rem 1rem;
  text-align: left;
  font-size: 0.8125rem;
  font-weight: 700;
  letter-spacing: 0;
  color: rgb(17 24 39);
  white-space: nowrap;
}

.dark .table-th {
  color: rgb(229 231 235);
}

.table-td {
  padding: 0.875rem 1rem;
  vertical-align: top;
}

.badge {
  display: inline-flex;
  align-items: center;
  border-radius: 9999px;
  padding: 0.125rem 0.5rem;
  font-size: 0.75rem;
  font-weight: 600;
  line-height: 1.25rem;
}

.badge-gray {
  background: rgb(243 244 246);
  color: rgb(55 65 81);
}

.badge-blue {
  background: rgb(219 234 254);
  color: rgb(30 64 175);
}

.badge-amber {
  background: rgb(254 243 199);
  color: rgb(146 64 14);
}

.badge-green {
  background: rgb(220 252 231);
  color: rgb(22 101 52);
}

.badge-red {
  background: rgb(254 226 226);
  color: rgb(153 27 27);
}

.dark .badge-gray {
  background: rgb(55 65 81);
  color: rgb(229 231 235);
}

.dark .badge-blue {
  background: rgb(30 58 138 / 0.35);
  color: rgb(147 197 253);
}

.dark .badge-amber {
  background: rgb(120 53 15 / 0.35);
  color: rgb(252 211 77);
}

.dark .badge-green {
  background: rgb(20 83 45 / 0.35);
  color: rgb(134 239 172);
}

.dark .badge-red {
  background: rgb(127 29 29 / 0.35);
  color: rgb(252 165 165);
}
</style>
