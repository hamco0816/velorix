# 下游 API 对接文档

本文档面向需要接入本服务的下游客户、代理商或第三方系统。下游只需要拿到平台分配的 API Key，然后按 OpenAI、Claude 或 Gemini 兼容协议调用网关接口。

> 当前生产域名示例使用 `https://api.velorix.chat`。如果是私有部署，把示例里的域名替换成你的部署域名即可。

## 1. 对接范围

下游对接的是模型网关接口，不是后台管理接口。

- 下游使用：用户 API Key，例如 `sk-xxxxxxxx`
- 下游不要使用：后台 JWT、管理员账号密码、数据库账号
- 推荐调用方式：服务端调用，不建议把 API Key 暴露在浏览器前端

## 2. Base URL

| 场景 | Base URL | 说明 |
| --- | --- | --- |
| OpenAI 兼容接口 | `https://api.velorix.chat/v1` | 适合 OpenAI SDK、Chat Completions、Responses |
| Claude Messages 兼容接口 | `https://api.velorix.chat/v1` | 适合 Anthropic Messages 格式 |
| OpenAI Responses 别名接口 | `https://api.velorix.chat` | 支持 `/responses`、`/chat/completions` 等无 `/v1` 前缀路径 |
| Gemini 原生接口 | `https://api.velorix.chat/v1beta` | 适合 Gemini `generateContent` 格式 |
| Codex 直连接口 | `https://api.velorix.chat` | 完整路径 `/backend-api/codex/responses`，适合需要 Codex 风格 backend API 的客户端 |

优先建议下游使用 `https://api.velorix.chat/v1`，除非客户端明确要求 Gemini 原生路径或 Codex 专用路径。

## 3. 鉴权方式

推荐使用 `Authorization` 请求头：

```http
Authorization: Bearer sk-xxxxxxxx
```

也支持以下请求头：

```http
x-api-key: sk-xxxxxxxx
x-goog-api-key: sk-xxxxxxxx
```

Gemini 原生接口可以使用 `x-goog-api-key`。

⚠️ **不要把 API Key 放在 URL 查询参数里**：`?key=sk-xxx` 或 `?api_key=sk-xxx` 会被网关直接拒绝并返回 `400 api_key_in_query_deprecated`。

API Key 必须满足以下条件：

- Key 未禁用、未过期
- Key 所属用户有可用余额、套餐或额度
- Key 已分配到可用分组
- 分组绑定的上游平台支持当前请求的接口类型

## 4. 支持的主要接口

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| `GET` | `/v1/models` | 获取当前 Key 在所属分组下可用的模型列表 |
| `GET` | `/v1/usage` | 查询当前 Key 的用量、额度、速率窗口信息 |
| `POST` | `/v1/chat/completions` | OpenAI Chat Completions 兼容接口 |
| `POST` | `/v1/responses` | OpenAI Responses 兼容接口，推荐新项目优先使用 |
| `POST` | `/v1/messages` | Claude Messages 兼容接口 |
| `POST` | `/v1/messages/count_tokens` | Claude token 统计接口，OpenAI 分组返回 404 |
| `POST` | `/v1/images/generations` | 图片生成，仅 OpenAI 类型分组支持（其他分组返回 404）|
| `POST` | `/v1/images/edits` | 图片编辑，仅 OpenAI 类型分组支持 |
| `GET` | `/v1beta/models` | Gemini 原生模型列表 |
| `POST` | `/v1beta/models/{model}:generateContent` | Gemini 原生内容生成 |
| `POST` | `/backend-api/codex/responses` | Codex 直连 Responses 接口 |

同一个 API Key 绑定到不同分组时，网关会按分组上游平台自动路由：

- OpenAI 类型分组 → `/v1/chat/completions` 和 `/v1/responses` 走 OpenAI 实现
- Claude/Anthropic 类型分组 → `/v1/messages` 走 Anthropic 实现，`/v1/chat/completions` 会做协议转换
- Gemini 类型分组 → 推荐用 `/v1beta` 原生路径

## 5. 如何确定可用模型名

**不要凭印象写模型名**。每个 API Key 实际可调用的模型由后台分组配置决定，不同分组开放的模型可能不同。可以通过 `/v1/models` 端点查询当前 Key 实际可用模型：

```bash
curl https://api.velorix.chat/v1/models \
  -H "Authorization: Bearer sk-xxxxxxxx"
```

返回的 `data[].id` 字段就是可以填入 `model` 参数的名称。如果客户端要求兼容模型名（例如把 `gpt-4` 映射到分组里的 `gpt-5.5`），可以在后台分组的"模型映射"中配置。

## 6. 快速测试

```bash
curl https://api.velorix.chat/v1/models \
  -H "Authorization: Bearer sk-xxxxxxxx"
```

如果返回模型列表（JSON 形如 `{"object":"list","data":[...]}`），说明域名、网络和 API Key 基本正常。

## 7. OpenAI Chat Completions 示例

```bash
curl https://api.velorix.chat/v1/chat/completions \
  -H "Authorization: Bearer sk-xxxxxxxx" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-5.5",
    "messages": [
      {
        "role": "user",
        "content": "你好，简单介绍一下你自己"
      }
    ],
    "stream": true
  }'
```

Python SDK 示例：

```python
from openai import OpenAI

client = OpenAI(
    api_key="sk-xxxxxxxx",
    base_url="https://api.velorix.chat/v1",
)

stream = client.chat.completions.create(
    model="gpt-5.5",
    messages=[
        {"role": "user", "content": "写一个 Python 快速排序示例"}
    ],
    stream=True,
)

for event in stream:
    delta = event.choices[0].delta.content
    if delta:
        print(delta, end="")
```

Node.js SDK 示例：

```javascript
import OpenAI from "openai";

const client = new OpenAI({
  apiKey: "sk-xxxxxxxx",
  baseURL: "https://api.velorix.chat/v1",
});

const stream = await client.chat.completions.create({
  model: "gpt-5.5",
  messages: [
    { role: "user", content: "写一个 JavaScript 防抖函数" },
  ],
  stream: true,
});

for await (const part of stream) {
  const text = part.choices?.[0]?.delta?.content;
  if (text) process.stdout.write(text);
}
```

## 8. OpenAI Responses 示例

```bash
curl https://api.velorix.chat/v1/responses \
  -H "Authorization: Bearer sk-xxxxxxxx" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-5.5",
    "input": "用 Go 写一个 HTTP 服务示例",
    "stream": true
  }'
```

如果客户端只能配置根域名，也可以使用别名路径：

```bash
curl https://api.velorix.chat/responses \
  -H "Authorization: Bearer sk-xxxxxxxx" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-5.5",
    "input": "解释一下 SSE 流式响应",
    "stream": true
  }'
```

## 9. Claude Messages 示例

Claude 兼容接口使用 Anthropic Messages 请求格式：

```bash
curl https://api.velorix.chat/v1/messages \
  -H "Authorization: Bearer sk-xxxxxxxx" \
  -H "Content-Type: application/json" \
  -H "anthropic-version: 2023-06-01" \
  -d '{
    "model": "claude-sonnet-4-5",
    "max_tokens": 1024,
    "messages": [
      {
        "role": "user",
        "content": "帮我审查这段代码有什么问题"
      }
    ],
    "stream": true
  }'
```

如果使用 Anthropic SDK，需要确认 SDK 支持自定义 base URL，并把 base URL 指向：

```text
https://api.velorix.chat
```

实际请求路径仍然是 `/v1/messages`。

## 10. Gemini 原生接口示例

Gemini 原生接口推荐使用 `x-goog-api-key` 请求头：

```bash
curl https://api.velorix.chat/v1beta/models \
  -H "x-goog-api-key: sk-xxxxxxxx"
```

生成内容：

```bash
curl https://api.velorix.chat/v1beta/models/gemini-2.5-pro:generateContent \
  -H "x-goog-api-key: sk-xxxxxxxx" \
  -H "Content-Type: application/json" \
  -d '{
    "contents": [
      {
        "parts": [
          {
            "text": "用三句话解释什么是 API 网关"
          }
        ]
      }
    ]
  }'
```

## 11. Codex 或代码生成客户端

如果客户端支持 OpenAI 兼容配置，优先使用：

```text
Base URL: https://api.velorix.chat/v1
API Key: sk-xxxxxxxx
Endpoint: /responses
```

如果客户端要求 Codex 风格的 backend API，**完整请求路径是 `/backend-api/codex/responses`**。常见客户端配置写法：

```text
Base URL: https://api.velorix.chat
Endpoint: /backend-api/codex/responses
API Key: sk-xxxxxxxx
```

部分客户端会把 Base URL 与 Endpoint 拼接，确认最终请求 URL 是 `https://api.velorix.chat/backend-api/codex/responses` 即可。

代码生成任务通常时间较长，建议客户端和中间代理把读取超时设置得更长，例如 10 到 60 分钟。流式请求过程中不要主动断开连接，否则客户端可能会看到类似 `stream disconnected before completion` 的错误。

### 11.1 用 cc-switch 图形化配置 Codex（图形界面用户推荐）

[cc-switch](https://github.com/farion1231/cc-switch) 是常用的多供应商切换工具，可以在图形界面里管理 Codex、Claude Code 等客户端的配置。用它接入本服务时，**最容易出错的就是 `config.toml` 的写法**，下面给出可直接套用的配置。

**第一步：新建 Codex 供应商，填写基本字段**

| 字段 | 填写内容 |
| --- | --- |
| 供应商名称 | 任意，例如 `velorix` |
| API 请求地址 | `https://api.velorix.chat/v1` |
| API Key | `sk-xxxxxxxx`（平台分配的 Key） |
| 模型名称 | `gpt-5.4`（以 `/v1/models` 实际返回为准） |

**第二步：填写 `auth.json`**

```json
{
  "OPENAI_API_KEY": "sk-xxxxxxxx",
  "auth_mode": "apikey"
}
```

**第三步：填写 `config.toml`（关键，写错就连不通）**

```toml
model_provider = "OpenAI"
model = "gpt-5.4"
model_reasoning_effort = "xhigh"
disable_response_storage = true

[model_providers.OpenAI]
name = "OpenAI"
base_url = "https://api.velorix.chat/v1"
wire_api = "responses"
requires_openai_auth = true
```

> 如果你的 `~/.codex/config.toml` 里原本还有 `[projects]`、`[plugins]`、`[windows]` 等个人设置，保留它们即可，只要把上面这段 `model_provider` 加 `[model_providers.OpenAI]` 补进去就行。注意 TOML 规则：所有顶层 `key = value` 必须写在第一个 `[xxx]` 表头之前。

**为什么必须有 `[model_providers.OpenAI]` 这一段？**

Codex **只认 `[model_providers.X]` 块里的 `base_url`**，并且要靠顶层 `model_provider = "X"` 指过去才会生效。如果只在 `config.toml` 顶层写 `base_url = "https://api.velorix.chat/v1"`、却没有对应的 provider 块，Codex 会忽略它、继续把请求发到官方 `https://api.openai.com`，于是你填的 Key 在官方那边鉴权失败，表现就是"连接不通 / 401"。`wire_api = "responses"` 也必须有，否则 Codex 会走 `/chat/completions` 而不是本服务推荐的 `/responses`。

**常见错误对照**

| 错误写法 | 现象 | 修正 |
| --- | --- | --- |
| 只在顶层写 `base_url`，没有 `[model_providers]` 块 | 请求发到 api.openai.com，401 / 连不通 | 按上面补全 provider 块 |
| 有 provider 块但缺 `wire_api = "responses"` | 走错接口，报 404 或格式错误 | 加上 `wire_api = "responses"` |
| `model_provider` 的值和 `[model_providers.X]` 的 X 不一致 | Codex 找不到 provider，回退官方 | 两处名称必须完全一致 |

**第四步：验证**

切换到该供应商后，先用一条命令确认网关侧是通的：

```bash
curl https://api.velorix.chat/v1/models -H "Authorization: Bearer sk-xxxxxxxx"
```

能返回模型列表 JSON 就说明 Key 和网络正常；之后在终端运行 `codex` 发一句话测试即可。如果 `curl` 正常但 `codex` 仍连不通，基本就是 `config.toml` 的 provider 块没写对，回到第三步检查。

## 12. 计费与限流（理解 429 的关键）

下游遇到 `429` 时，可能是触发了以下任意一种限制：

| 限制类型 | 触发条件 | 处理方式 |
| --- | --- | --- |
| 余额不足 | 用户钱包余额为 0 或不够支付本次请求 | 充值，或换用配套套餐 |
| 套餐额度耗尽 | 订阅模式下，5h / 1d / 7d 滚动窗口的额度用尽 | 等待窗口刷新，或升级套餐 |
| Key 额度耗尽 | API Key 自身设置了独立 quota 且已用尽 | 后台重置 Key 配额，或换 Key |
| 并发超限 | 同一用户/Key 当前活跃请求数超过限制 | 降低并发度，或后台调高并发上限 |
| RPM 超限 | 每分钟请求数超过 Key 或用户的 RPM 上限 | 降低请求频率 |

计费按 token 进行：

- 输入 token、输出 token、缓存创建 token、缓存读取 token 各有独立单价
- 缓存读取价格通常远低于普通输入价（具体倍率以"计费标准"页为准）
- 每次请求结束后，从用户余额中扣除（或从套餐额度中累加）

下游可以通过 `/v1/usage` 主动查询当前 Key 的用量、剩余额度和速率窗口状态。

## 13. /v1/usage 查询用量

```bash
curl https://api.velorix.chat/v1/usage \
  -H "Authorization: Bearer sk-xxxxxxxx"
```

返回结构（节选）：

```json
{
  "mode": "quota_limited",
  "isValid": true,
  "status": "active",
  "quota": {
    "limit": 100.0,
    "used": 23.45,
    "remaining": 76.55,
    "unit": "USD"
  },
  "rate_limits": [
    {
      "window": "5h",
      "limit": 50000,
      "used": 12340,
      "remaining": 37660,
      "window_start": "2026-05-13T08:00:00Z",
      "reset_at": "2026-05-13T13:00:00Z"
    }
  ],
  "usage": {
    "today": {
      "requests": 142,
      "input_tokens": 32100,
      "output_tokens": 15800,
      "cache_creation_tokens": 0,
      "cache_read_tokens": 4200,
      "total_tokens": 52100,
      "actual_cost": 0.34
    },
    "total": { /* 同 today 结构 */ },
    "average_duration_ms": 1820,
    "rpm": 6,
    "tpm": 2150
  }
}
```

字段说明：

- `mode`: `quota_limited`（Key 有额度/速率限制）或 `unrestricted`（无限制，按余额扣费）
- `quota`: 仅在 Key 设置了独立 quota 时返回
- `rate_limits`: 仅在 Key 设置了 5h/1d/7d 速率窗口时返回；`window_start` 是窗口起点，`reset_at` 是当前窗口耗尽后多久重置
- `usage.today.actual_cost`: 当日实际扣费（已应用分组倍率/促销），单位 USD

`/v1/usage` 不计费、不消耗额度，可以高频轮询。

## 14. 错误响应格式

所有网关错误返回 JSON 包含 `type` 和 `error` 两个顶层字段：

```json
{
  "type": "error",
  "error": {
    "type": "authentication_error",
    "message": "Invalid API key"
  }
}
```

常见错误码：

| HTTP 状态 | `error.type` | 常见原因 | 处理方式 |
| --- | --- | --- | --- |
| `400` | `invalid_request_error` | 请求体格式错误、缺少必填字段 | 检查 JSON 结构 |
| `400` | `api_key_in_query_deprecated` | API Key 放在了 URL 查询参数里 | 改用 `Authorization: Bearer` 请求头 |
| `401` | `authentication_error` | 未传 API Key、Key 无效或已被吊销 | 检查请求头和 Key 是否复制完整 |
| `403` | `permission_error` | Key 被禁用、过期、用户无权访问该分组 | 在后台检查 Key 状态、用户状态和分组 |
| `404` | `not_found_error` | 当前分组不支持该接口（例如非 OpenAI 分组调用图片接口） | 换成支持该接口的平台分组 |
| `413` | `invalid_request_error` | 请求体超过网关 body size 限制 | 减小 payload 或在后台调高 `MaxBodySize` |
| `429` | `rate_limit_error` | 触发并发/RPM/额度/余额限制 | 参考第 12 节 |
| `503` | `overloaded_error` | 网关过载保护（CPU/内存/磁盘门控）或上游不可用 | 稍后重试，必要时切换上游或扩容 |

## 15. 流式响应与超时建议

下游如果用于写代码、长文本生成或 Agent 场景，建议启用流式响应：

```json
{
  "stream": true
}
```

流式响应使用 Server-Sent Events (SSE) 格式，每个事件以 `data: <json>\n\n` 分隔，最后以 `data: [DONE]` 结束（Claude Messages 格式略有不同，参考 Anthropic 官方文档）。

客户端侧建议：

- HTTP 读取超时设置为 10 到 60 分钟
- 反向代理读取超时也要同步调大
- 不要在流式响应未完成前关闭连接
- 遇到网络中断时由业务侧决定是否重试

如果经过 Nginx、Caddy、Cloudflare 或其他代理，需要确认代理层没有较短的 `read_timeout`、`response_header_timeout` 或类似限制。

## 16. 下游上线前检查清单

- 已拿到正确的 API Key
- 已确认使用的 Base URL
- 已用 `/v1/models` 测试 Key 可用并获取实际可用模型名
- 已确认 Key 所属用户余额、套餐或额度正常
- 已确认 Key 绑定的分组支持目标接口
- 客户端和代理层超时已满足长任务需求
- 生产环境不要把 API Key 写死在前端页面或公开仓库
- 已了解 429 的几种触发条件和应对方式（第 12 节）

## 17. 提供给下游的最小配置

给下游交付时，至少提供以下信息：

```text
Base URL: https://api.velorix.chat/v1
API Key: sk-xxxxxxxx
推荐接口: /v1/responses 或 /v1/chat/completions
模型名称: 调用 /v1/models 查询实际可用模型
是否支持流式: 支持，设置 stream=true
用量查询: GET /v1/usage
```

如果下游使用 Gemini 原生客户端：

```text
Base URL: https://api.velorix.chat/v1beta
API Key Header: x-goog-api-key
API Key: sk-xxxxxxxx
```

如果下游使用 Claude Messages 格式：

```text
Base URL: https://api.velorix.chat
Endpoint: /v1/messages
API Key: sk-xxxxxxxx
anthropic-version: 2023-06-01
```
