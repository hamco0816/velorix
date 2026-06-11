# Design

## Theme

白底 + 纯黑 CTA + 暖橙点睛（Stripe/Vercel 质感）。浅色为主战场，暗色模式同等完成度（class 切换，`dark:` 前缀 + `:root.dark` 覆盖）。色彩策略：Restrained——中性 Zinc 灰阶承载界面，品牌橙 ≤10% 仅用于选中态/点睛/营销强调。

## Color

全部通过 `frontend/tailwind.config.js` 令牌使用，禁止新增裸色值。

- **primary（Zinc）**：50 `#fafafa` → 950 `#09090b`。界面骨架灰阶；`gray`、`dark` 两个别名同源重映射（历史兼容，勿改回多色温）。
- **brand（Orange）**：50 `#fff7ed` → 950 `#431407`，主档 500 `#f97316` / 600 `#ea580c`。用于：选中/激活态、focus ring、关键徽章、营销点睛。
- **语义色**：success=emerald、warning=amber、danger=red、info=sky（info 框统一 sky，不用蓝紫）。
- **保留多色的功能场景**：图表数据系列、平台/分组/状态语义徽章（功能性，勿收敛）。
- 收敛规则（已沉淀）：选中/激活态=品牌橙；分段切换器激活=中性 zinc；info 框=sky；正文链接=品牌橙。

## Typography

- 系统字体栈（system-ui / PingFang SC / Microsoft YaHei），单家族多字重，无 display 配对。
- 固定 rem 标尺：`2xs`(11px 徽章/角标) / xs / sm(正文主力) / base / 页面标题 `text-2xl font-bold tracking-tight`。
- 17px+ 展示级数字用 `tabular-nums tracking-tight`；金额、用量等可核对数字必须 tabular。
- 禁止 ≤16px 任意值字号（`scripts/check-style-discipline.mjs` 强制）。

## Components

组件类定义于 `frontend/src/style.css` @layer components；Vue 组件库在 `src/components/common/`。

- **按钮** `.btn` + `.btn-primary`(纯黑/暗色反白) `.btn-secondary`(白底描边) `.btn-ghost` `.btn-danger` `.btn-success` `.btn-warning`(品牌橙) ；尺寸 `.btn-sm/md/lg/icon`，min-height 严格等高。
- **输入** `.input`（36px 与按钮同高）、auth 专用 `.auth-input` 系列。
- **卡片** `.card`(rounded-xl + shadow-card) / `.surface-card`(rounded-2xl 数据面板) / `.card-hover`(可点浮起)；彩色氛围卡 `.card-amber` 等已统一为品牌暖橙微光，仅限单张主容器。
- **页面 hero** `.page-hero`（品牌暖橙微光渐变，历史七彩类名全部映射至此）。
- **弹窗** BaseDialog（focus trap + 弹窗栈）+ `.modal-*` 类；确认走 `useConfirm()` + ConfirmDialog，禁止原生 confirm。
- **表格** DataTable（含 error/@retry/骨架/空态）+ `.table-*` 类；**状态** StatusBadge/StatusTag/`.badge-*`；**空态** EmptyState；**错误** ErrorState；**加载** Skeleton / LoadingSpinner（spinner 只用于按钮内，内容区用骨架）。
- **图标** Icon.vue 统一单色描边图标库（继承文字色），品牌 logo 用 BrandIcon/PlatformIcon。

## Elevation & Depth

- `shadow-card`（静止卡片）→ `shadow-card-hover`（可点 hover）→ `shadow-elevated`（下拉/浮层）三级。
- 深色按钮用 `shadow-inner-top` 内嵌高光，不用彩色光晕。
- 玻璃效果仅限 modal 背景遮罩与吸顶导航。

## Motion

- 150–250ms，ease-out；预设 `animate-fade-in / slide-up / slide-down / slide-in-right / scale-in`。
- 动效只传达状态（hover 浮起、弹窗缩放进出、下拉 scale-in）；所有自定义动效必须配 `prefers-reduced-motion: reduce` 降级。

## Layout

- 应用壳：AppLayout（左侧 AppSidebar 264px 可收纳 + AppHeader 顶栏）；侧边栏激活态为黑底反白（与黑 CTA 同语言）。
- 页面头：PageHeader（标题 + 副标题 + meta + actions 插槽）。
- 间距节奏：页面 `space-y-8`，卡片体 `p-6`，表单 `gap-4`。
- 响应式靠结构（侧栏收纳、表格横滚、断点列数），不靠流体字号。

## Hard Rules

- 禁止装饰性蓝紫渐变（lint 强制）、禁止 side-stripe 彩条卡片、禁止 gradient text。
- 新增颜色一律走令牌；新增字号走标尺档位。
- 每个数据视图必须接 loading/empty/error 三态。
