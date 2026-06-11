# Product

## Register

product

## Users

- **终端开发者/AI 重度用户（前台用户端）**：购买 API 配额、生成 API Key、查看用量与账单。多为程序员，熟悉 Stripe/Vercel/Linear 这类工具的质感标准；使用场景是工作中的快速进出（充值、查 key、看用量），停留时间短、目标明确。
- **站长/运营者（管理端）**：基于 Sub2API 自建中转站点，管理上游账号、渠道、定价、订单与用户。长时间停留的运营工作台，信息密度优先。
- 中英双语用户（i18n zh/en 全覆盖），桌面为主、移动端需可用。

## Product Purpose

Sub2API 是开源 AI API 网关平台：分发和管理 AI 产品订阅的 API 配额，用户通过平台生成的 Key 调用上游 AI 服务（Claude/OpenAI/Gemini/DeepSeek），平台负责鉴权、计费、负载均衡与转发，内置支付（易支付/支付宝/微信/Stripe）。成功标准：用户信任并完成充值转化；站长部署后无需二次美化即可对外运营。

## Brand Personality

可信、专业、克制。三个词：**Trustworthy / Precise / Calm**。这是一个「替用户管钱和配额」的平台，界面要传达的核心情绪是**财务级的可靠感**——像 Stripe 后台一样让人放心把钱放进来。点睛的暖橙带来一点温度与活力，避免冷漠。

## Anti-references

- AI 味套壳站：紫色渐变、玻璃拟态滥用、emoji 当图标、每张卡片渐变边框、无意义动画。
- 上一代国产管理后台模板（Element 默认蓝、密密麻麻彩色按钮、彩虹色统计卡）。
- 营销页的「币圈感」：夸张大字报、跑马灯数字、闪烁徽章。

## Design Principles

1. **信任先行**：涉及钱与配额的数字（余额、价格、用量）永远清晰、等宽、可核对；状态透明（实时更新时间、错误可重试）。
2. **黑白为骨、橙为点睛**：白底 + 纯黑 CTA 是骨架；品牌橙只用于选中态、关键点睛与营销强调，永不铺面。
3. **工具消失于任务**：用户端高频动作（充值、建 Key、查用量）必须 ≤2 步可达；熟悉的范式优于新奇的发明。
4. **状态完备**：每个界面必有 loading（骨架）、empty（教学式）、error（可重试）三态；每个控件必有 hover/focus/disabled 态。
5. **动效传达状态**：150–250ms、ease-out；动效只为反馈与层级服务，永不为装饰；尊重 prefers-reduced-motion。

## Accessibility & Inclusion

- 正文对比度 ≥4.5:1，大字 ≥3:1；焦点环全站统一可见。
- 弹窗有 focus trap 与弹窗栈管理（已建于 BaseDialog）；表单控件有 label/aria-invalid/aria-describedby。
- 全量动效提供 reduced-motion 降级；暗色模式与浅色模式同等完成度。
