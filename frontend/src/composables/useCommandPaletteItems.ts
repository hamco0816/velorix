/**
 * 命令面板条目构建：按当前用户角色、简单模式与功能开关，生成与侧边栏一致的
 * 可搜索导航清单 + 快捷操作（充值 / 新建密钥 / 切主题 / 切语言）。
 */
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useAppStore, useAuthStore } from '@/stores'
import { useAdminSettingsStore } from '@/stores/adminSettings'
import { FeatureFlags, isFeatureFlagEnabled } from '@/utils/featureFlags'
import { getLocale, setLocale } from '@/i18n'
import { useTheme } from '@/composables/useTheme'

/** 仅允许 Icon.vue 中实际存在的图标名 */
export type PaletteIconName =
  | 'grid' | 'key' | 'sparkles' | 'chart' | 'creditCard' | 'gift' | 'document'
  | 'badge' | 'globe' | 'bolt' | 'book' | 'users' | 'user' | 'dollar'
  | 'sun' | 'moon' | 'cog' | 'server' | 'shield' | 'bell' | 'chat'
  | 'calculator' | 'inbox' | 'externalLink'

export type PaletteGroupKey = 'navigation' | 'admin' | 'actions'

export interface PaletteItem {
  id: string
  label: string
  icon: PaletteIconName
  /** 额外检索词（路径 + 中英同义词），与 label 一起参与模糊匹配 */
  keywords: string
  group: PaletteGroupKey
  perform: () => void
}

export interface PaletteGroup {
  key: PaletteGroupKey
  label: string
  items: PaletteItem[]
}

export function useCommandPaletteItems() {
  const { t } = useI18n()
  const router = useRouter()
  const appStore = useAppStore()
  const authStore = useAuthStore()
  const adminSettingsStore = useAdminSettingsStore()
  const { isDark, toggleTheme } = useTheme()

  const groups = computed<PaletteGroup[]>(() => {
    const simple = authStore.isSimpleMode
    const isAdmin = authStore.isAdmin
    const payment = isFeatureFlagEnabled(FeatureFlags.payment)
    const go = (path: string, query?: Record<string, string>) => () => {
      router.push(query ? { path, query } : path)
    }

    // —— 用户端导航：与 AppSidebar.buildSelfNavItems 的可见性规则保持一致 ——
    const nav: PaletteItem[] = []
    const addNav = (
      visible: boolean,
      id: string,
      labelKey: string,
      icon: PaletteIconName,
      path: string,
      keywords: string
    ) => {
      if (!visible) return
      nav.push({ id, label: t(labelKey), icon, keywords: `${path} ${keywords}`, group: 'navigation', perform: go(path) })
    }

    addNav(!isAdmin, 'nav-dashboard', 'nav.dashboard', 'grid', '/dashboard', 'dashboard 仪表盘 概览 overview')
    addNav(true, 'nav-keys', 'nav.apiKeys', 'key', '/keys', 'api key token 密钥')
    addNav(!simple, 'nav-image-gen', 'nav.imageGen', 'sparkles', '/image-gen', 'image 绘图 画图 生成')
    addNav(!simple, 'nav-usage', 'nav.usage', 'chart', '/usage', 'usage logs 用量 使用记录')
    addNav(!simple && payment, 'nav-purchase', 'nav.buySubscription', 'creditCard', '/purchase', 'purchase recharge 充值 订阅 购买')
    addNav(!simple, 'nav-redeem', 'nav.redeem', 'gift', '/redeem', 'redeem code 兑换码')
    addNav(!simple && payment, 'nav-orders', 'nav.myOrders', 'document', '/orders', 'orders 订单')
    addNav(!simple && isFeatureFlagEnabled(FeatureFlags.invoice), 'nav-invoices', 'nav.myInvoices', 'document', '/invoices', 'invoice 发票')
    addNav(!simple, 'nav-subscriptions', 'nav.mySubscriptions', 'creditCard', '/subscriptions', 'subscriptions 订阅')
    addNav(!simple && payment, 'nav-seats', 'nav.mySeats', 'badge', '/seats', 'seats 独享 专属')
    addNav(!simple && isFeatureFlagEnabled(FeatureFlags.availableChannels), 'nav-pricing', 'nav.pricing', 'globe', '/pricing', 'pricing channels 计费 价格 渠道')
    addNav(isFeatureFlagEnabled(FeatureFlags.channelMonitor), 'nav-monitor', 'nav.channelStatus', 'bolt', '/monitor', 'status monitor 状态 监控')
    addNav(!simple, 'nav-docs', 'nav.docs', 'book', '/docs', 'docs documentation 文档')
    addNav(!simple && isFeatureFlagEnabled(FeatureFlags.affiliate), 'nav-affiliate', 'nav.affiliate', 'users', '/affiliate', 'affiliate invite 邀请 返利')
    addNav(true, 'nav-profile', 'nav.profile', 'user', '/profile', 'profile account 个人资料 账户 设置')

    // 自定义菜单（站长配置的外部页面）
    const customItems = (appStore.cachedPublicSettings?.custom_menu_items ?? [])
      .filter((item) => item.visibility === 'user')
    for (const item of customItems) {
      nav.push({
        id: `nav-custom-${item.id}`,
        label: item.label,
        icon: 'externalLink',
        keywords: `/custom/${item.id}`,
        group: 'navigation',
        perform: go(`/custom/${item.id}`)
      })
    }

    // —— 管理后台导航（仅管理员可见，可见性与 adminNavItems 同步） ——
    const admin: PaletteItem[] = []
    if (isAdmin) {
      const addAdmin = (
        visible: boolean,
        id: string,
        labelKey: string,
        icon: PaletteIconName,
        path: string,
        keywords: string
      ) => {
        if (!visible) return
        admin.push({ id, label: t(labelKey), icon, keywords: `${path} ${keywords}`, group: 'admin', perform: go(path) })
      }
      const adminPayment = adminSettingsStore.paymentEnabled

      addAdmin(true, 'admin-dashboard', 'nav.dashboard', 'grid', '/admin/dashboard', 'admin dashboard 管理 仪表盘')
      addAdmin(adminSettingsStore.opsMonitoringEnabled, 'admin-ops', 'nav.ops', 'chart', '/admin/ops', 'ops monitoring 运维 监控')
      addAdmin(true, 'admin-safety', 'nav.safetyRisk', 'shield', '/admin/safety-risk', 'safety risk 风控 安全')
      addAdmin(!simple, 'admin-users', 'nav.users', 'users', '/admin/users', 'users 用户管理')
      addAdmin(!simple, 'admin-groups', 'nav.groups', 'inbox', '/admin/groups', 'groups 分组')
      addAdmin(true, 'admin-accounts', 'nav.accounts', 'globe', '/admin/accounts', 'accounts 账号')
      addAdmin(!simple, 'admin-channel-pricing', 'nav.channelPricing', 'dollar', '/admin/channels/pricing', 'channel pricing 渠道定价')
      addAdmin(!simple && isFeatureFlagEnabled(FeatureFlags.channelMonitor), 'admin-channel-monitor', 'nav.channelMonitor', 'bolt', '/admin/channels/monitor', 'channel monitor 渠道监控')
      addAdmin(true, 'admin-proxies', 'nav.proxies', 'server', '/admin/proxies', 'proxy ip 代理')
      addAdmin(!simple && adminPayment, 'admin-orders', 'nav.orderManagement', 'document', '/admin/orders', 'orders 订单管理')
      addAdmin(!simple && adminPayment, 'admin-plans', 'nav.paymentPlans', 'creditCard', '/admin/orders/plans', 'plans 套餐')
      addAdmin(!simple, 'admin-subscriptions', 'nav.subscriptions', 'creditCard', '/admin/subscriptions', 'subscriptions 订阅管理')
      addAdmin(!simple, 'admin-model-pricing', 'nav.modelPricing', 'calculator', '/admin/pricing/models', 'model pricing 模型定价')
      addAdmin(true, 'admin-usage', 'nav.usage', 'chart', '/admin/usage', 'usage 使用记录')
      addAdmin(!simple, 'admin-redeem', 'nav.redeemCodes', 'gift', '/admin/redeem', 'redeem codes 兑换码')
      addAdmin(!simple, 'admin-promo', 'nav.promoCodes', 'gift', '/admin/promo-codes', 'promo codes 优惠码')
      addAdmin(true, 'admin-announcements', 'nav.announcements', 'bell', '/admin/announcements', 'announcements 公告')
      addAdmin(true, 'admin-support', 'nav.onlineSupport', 'chat', '/admin/support', 'support 客服 工单')
      addAdmin(true, 'admin-settings', 'nav.settings', 'cog', '/admin/settings', 'settings 系统设置')
    }

    // —— 快捷操作 ——
    const actions: PaletteItem[] = []
    if (!simple && payment) {
      actions.push({
        id: 'action-recharge',
        label: t('commandPalette.actions.recharge'),
        icon: 'dollar',
        keywords: 'recharge topup balance 充值 余额',
        group: 'actions',
        perform: go('/purchase', { tab: 'recharge' })
      })
    }
    actions.push(
      {
        id: 'action-create-key',
        label: t('commandPalette.actions.createKey'),
        icon: 'key',
        keywords: 'create new api key 新建 创建 密钥',
        group: 'actions',
        perform: go('/keys', { create: '1' })
      },
      {
        id: 'action-toggle-theme',
        label: t('commandPalette.actions.toggleTheme'),
        icon: isDark.value ? 'sun' : 'moon',
        keywords: 'theme dark light 主题 暗色 亮色 夜间',
        group: 'actions',
        perform: () => toggleTheme()
      },
      {
        id: 'action-switch-language',
        label: t('commandPalette.actions.switchLanguage'),
        icon: 'globe',
        keywords: 'language locale english chinese 语言 中文 英文',
        group: 'actions',
        perform: () => {
          void setLocale(getLocale() === 'zh' ? 'en' : 'zh')
        }
      }
    )

    const result: PaletteGroup[] = [
      { key: 'navigation', label: t('commandPalette.groups.navigation'), items: nav }
    ]
    if (admin.length > 0) {
      result.push({ key: 'admin', label: t('commandPalette.groups.admin'), items: admin })
    }
    result.push({ key: 'actions', label: t('commandPalette.groups.actions'), items: actions })
    return result
  })

  return { groups }
}
