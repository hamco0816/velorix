/**
 * 套餐限额可视性 / 校验工具
 *
 * 核心问题：admin 给套餐配 daily / weekly / monthly 三个时间窗口的额度上限，
 * 这三个窗口的关系是「短窗口数学上限会蚕食长窗口」：
 *
 *   daily $40 → 7 天理论最多花 $280，所以 weekly $300 等价于"无周限"，是废设置
 *   daily $40 → 30 天理论最多花 $1200，weekly $220 → 30 天最多 weekly × 30/7 ≈ $943
 *   所以 monthly 必须 < min($1200, $943) = $943 才真起约束
 *
 * 用户侧：废限额展示出来反而混淆用户（"明明每天 $40 为啥还有周 $280？"）→ 自动隐藏。
 * 管理员侧：admin 配错时下方红字提醒，但不阻断提交（允许故意配置）。
 */

const MONTH_DAYS = 30
const WEEK_DAYS = 7
// 月折算到周：30 / 7 ≈ 4.2857
const WEEK_TO_MONTH_RATIO = MONTH_DAYS / WEEK_DAYS

/** 取自 plan 的三个限额（USD），同结构出现在 user-facing card 和 admin 表单里 */
export interface LimitsInput {
  daily_limit_usd?: number | null
  weekly_limit_usd?: number | null
  monthly_limit_usd?: number | null
}

export interface LimitVisibility {
  showDaily: boolean
  showWeekly: boolean
  showMonthly: boolean
}

/**
 * 决定哪些限额"值得展示"——即不会被更紧的限额吃掉。
 *
 * 规则：
 *   - daily 设了就展示（它是 burst 防护，永远有意义）
 *   - weekly 仅当 < daily × 7 时展示（否则永远撞不到，是废设置）
 *   - monthly 仅当 < min(daily × 30, weekly × 30/7) 时展示
 *
 * 语义约定：
 *   - null    = 该字段未在 plan 层覆盖，沿用分组默认 → 不展示
 *   - <= 0    = 该字段被显式设为"无限制" → 展示（前端会渲染"无限制"标签）
 *   - > 0     = 具体额度 → 检查是否被更紧的限额废掉
 *
 * 注：只对 > 0 的有限值做废设置检查，"unlimited" (0) 永远不构成约束，
 * 也不会让其它限额变废。
 */
export function getEffectiveLimitVisibility(limits: LimitsInput): LimitVisibility {
  const daily = limits.daily_limit_usd
  const weekly = limits.weekly_limit_usd
  const monthly = limits.monthly_limit_usd

  const showDaily = daily != null

  let showWeekly = weekly != null
  if (showWeekly && weekly! > 0 && daily != null && daily > 0) {
    // weekly 是有限值，且 daily 也是有限值 → 比较 weekly vs daily × 7
    if (weekly! >= daily * WEEK_DAYS) showWeekly = false
  }

  let showMonthly = monthly != null
  if (showMonthly && monthly! > 0) {
    // 收集所有"更紧的限额折算到月"的上限，取最小
    const bounds: number[] = []
    if (daily != null && daily > 0) bounds.push(daily * MONTH_DAYS)
    if (weekly != null && weekly > 0) bounds.push(weekly * WEEK_TO_MONTH_RATIO)
    if (bounds.length > 0 && monthly! >= Math.min(...bounds)) {
      showMonthly = false
    }
  }

  return { showDaily, showWeekly, showMonthly }
}

export interface LimitWarning {
  /** 哪个字段废了 */
  field: 'weekly' | 'monthly'
  /** 警告文本（已带具体上限数字，可直接 v-text）*/
  message: string
  /** 计算出来的"真实上限"，给警告里展示用 */
  effectiveMax: number
  /** 是哪个更紧的限额造成的，用于解释 */
  cappedBy: 'daily' | 'weekly' | 'both'
}

/**
 * 校验三个限额的相互关系，返回每个废字段的警告。
 * 给 admin 表单实时显示用——不阻断提交，只是提醒。
 *
 * 返回值 keyed by field：
 *   { weekly: ..., monthly: ... }
 * 没有警告的字段 key 缺省（不要返回 null 字符串，让 v-if 判断更清爽）
 */
export function validatePlanLimits(
  daily: number | null | undefined,
  weekly: number | null | undefined,
  monthly: number | null | undefined,
): Partial<Record<'weekly' | 'monthly', LimitWarning>> {
  const result: Partial<Record<'weekly' | 'monthly', LimitWarning>> = {}

  // weekly redundant 判定
  if (daily != null && daily > 0 && weekly != null && weekly > 0) {
    const dailyMaxWeek = daily * WEEK_DAYS
    if (weekly >= dailyMaxWeek) {
      result.weekly = {
        field: 'weekly',
        message: '', // 由调用方用 i18n 填充
        effectiveMax: dailyMaxWeek,
        cappedBy: 'daily',
      }
    }
  }

  // monthly redundant 判定
  if (monthly != null && monthly > 0) {
    const bounds: { value: number; from: 'daily' | 'weekly' }[] = []
    if (daily != null && daily > 0) bounds.push({ value: daily * MONTH_DAYS, from: 'daily' })
    if (weekly != null && weekly > 0) bounds.push({ value: weekly * WEEK_TO_MONTH_RATIO, from: 'weekly' })
    if (bounds.length > 0) {
      // 取最紧的那个边界
      bounds.sort((a, b) => a.value - b.value)
      const tightest = bounds[0]
      if (monthly >= tightest.value) {
        result.monthly = {
          field: 'monthly',
          message: '',
          effectiveMax: tightest.value,
          // 如果两个边界值很接近（差异 < 5%），认为是 both；否则归到更紧的那个
          cappedBy:
            bounds.length > 1 && (bounds[1].value - tightest.value) / tightest.value < 0.05
              ? 'both'
              : tightest.from,
        }
      }
    }
  }

  return result
}
