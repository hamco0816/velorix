# 宣纸墨色 设计令牌规范（INK-DESIGN）

> 适用范围：sub2api 网页端（Tailwind）+ velorix 桌面端（原生 CSS 变量）。
> 改造性质：**色彩材质换肤**——只换 色板 / 字体 / 圆角 / 阴影 / 纹理 / 点睛，**不动布局、组件结构、信息架构**。
> 源参考：`D:/Project/vue/chaye/src/styles/design-tokens.css`（chaye 宋韵编辑风）。
>
> **评审修正已落地（v2）**：① 夜墨 `dark` 阶改回 Tailwind 正向（50 最浅 → 950 最深），避免暗色整页发白；② 占位 / 链接 / 实心 warning 统一用更深档（达 WCAG AA），`#b4ab98` 仅作纯装饰；③ focus ring 用 brand-600 + 2px；④ 夜墨 danger 提亮到 `#d8786a`。`ink-preview.html` 样板已按此渲染。
>
> **提饱和修订（v3，2026-06-16）**：v1/v2 的语义色与 brand/tea 色阶饱和度过低，落到状态徽章 / 分组标签 / 计费标签 / 统计卡图标 / 迷你折线上整体「发灰发土、像没有彩色」，违背了 §8.1「功能性多色保留不收敛」。已**保持各色相不变、整体提亮提饱和**（仍是宣纸土系，不做成 NewAPI 霓虹）。中性骨架 `primary/gray/dark`、字体、圆角、投影**完全不动**。新值如下（sub2api `tailwind.config.js` 已落地，下方 §2.x / §4 / §5 旧表为 v2 历史值，以本节与实际 config 为准；velorix 桌面端 CSS 变量待同步）：
>
> | 令牌 | v2（旧·偏灰） | v3（新·提饱和） |
> |------|------|------|
> | success DEFAULT / soft / deep | `#3d5a45` / `#e7ece2` / `#2c4232` | `#3a7d52` / `#d8ecdc` / `#2f6644` |
> | warning DEFAULT / soft / deep | `#c56b3e` / `#f4e6d8` / `#a9542c` | `#cf6f2c` / `#f8e3cd` / `#b1561f` |
> | danger / seal DEFAULT / soft / deep | `#a23b2e` / `#f3e1dc` / `#872f24` | `#c0392b` / `#f8ddd5` / `#9c2b1f` |
> | info DEFAULT / soft / deep | `#3a5570` / `#e3e8ee` / `#2c4256` | `#3f6ea6` / `#dde6f2` / `#2d5481` |
> | brand 50→950 | `#fbf3ea…#3a1d11`（500 `#c56b3e`） | `#fdf3e9 #f8e3cd #f0c79e #e7a86f #dd8a44 #cf6f2c #b1561f #8f4519 #723715 #562a12 #3a1c0c` |
> | tea 50→950 | `#eef2ec…#19251c`（600 `#3d5a45`） | `#edf5ef #d8ecdc #b8dcc0 #8fcaa0 #5fb27e #43965f #3a7d52 #2f6644 #285436 #21422c #16301e` |
>
> 暗色文字走 `dark:text-tea-300/400`（成功）与 `dark:text-brand-300/400`（警告），已随 tea/brand 色阶提亮自动变亮；浅色走 4 个语义令牌，随 config 提亮自动覆盖。另有 7 个文件共 26 处 config 够不到的硬编码 hex（迷你折线 / 图标渐变 / 仪表盘表盘 / 文档图标 / 图表色带）已按同一映射定点替换。
>
> **对比度修订（v3.1，2026-06-17，以本节为准）**：v3 提饱和后浅色仍有 WCAG AA 失败（实测：muted `#8a8275` 3.40 / 占位 `#b4ab98` 2.04 / 徽章文字落自身 soft 底 warning 2.83·success 4.00·danger 4.22 / 链接 brand-600 `#b1561f` 4.45，均 <4.5；**夜墨全达标，不动**）。已修：① muted = primary/gray-500 `#8a8275 → #736a5c`（4.76）；② 占位符 `gray-400 → gray-500`（共 8 处）；③ 正文链接 brand-600 `#b1561f → #a44f1c`（5.06）；④ 4 个语义色 **DEFAULT 调到「在自身 soft 底也 ≥4.5」**（因 sub2api 散落 473+ 处 `bg-X-soft text-X` 无法逐个改组件，走令牌级）：success `#3a7d52→#347049`、danger `#c0392b→#b23425`、info `#3f6ea6→#3a6499`、warning `#cf6f2c→#9c5712`（soft `#f8e3cd→#fbecd2`，**茶橘→深琥珀，与 brand 点睛区分**），seal 同 danger `#b23425`；⑤ velorix `--surface #fbf8f1→#fdfcf8`（卡片提净白增强层次）。注意：warning `#cf6f2c` 亮度过高，在任何背景上对比上限仅 3.89，故警告文字必须用更深档。

---

## 1. 设计理念

把现有「白底 + 冷黑 CTA + 日落橙点睛」的 Zinc 控制台，平滑迁移到「宣纸暖墨」体系：页面底色从纯白换成宣纸米白 `#f6f2e9`，中性灰阶整体注入暖褐调（**绝不回到冷中性灰**，这是宣纸气质的命门），主 CTA 把冷黑换成 chaye 暖墨黑 `#22201c`（primary-950 档），原日落橙顺势调成 chaye 茶汤橘 `#c56b3e` 当主点睛（选中态 / focus ring / 关键徽章 / 营销点睛；正文链接用其 600 档 #a9542c），竹青绿 `#3d5a45` 做辅助强调（成功态、次级图标）。圆角收到近方角（默认 3px），阴影改用极淡的暖墨投影而非黑，body 叠一层 opacity 0.04 的 feTurbulence 宣纸噪点。暗色模式新做一套自洽的「夜墨」：深暖墨底 + 宣纸色文字 + 提亮的竹青 / 茶橘点睛，与浅色版逐档对应。功能性多色（模型 / 支付品牌色、图表数据系列、终端 ANSI）一律**保留原样不收敛**。

承接关系一句话：**旧「黑 CTA + 橙点睛」→ 新「墨 CTA（暖墨黑）+ 茶橘点睛 + 竹青辅助」**，品牌力度与连续性都保住。

---

## 2. 浅色（宣纸）完整令牌表

### 2.1 中性灰阶 primary / gray / dark（11 档，三组同源，从宣纸过渡到暖墨）

> 关键：三组逐档 hex 必须**完全一致**（维持明暗切换色温统一），中间档全部带暖褐调。
> 50–200 取自 chaye 宣纸 / 线条系，300–950 是 paper→ink 的暖墨插值。

| 档位 | hex | 暖墨色温定位 | 典型用途 |
|------|------|------|------|
| 50  | `#f6f2e9` | 宣纸 paper | 页面主背景 |
| 100 | `#efe8d8` | 宣纸 paper-deep | 区块分隔 / 次级背景 |
| 200 | `#ddd5c4` | 线 line | 常规边框 / 分隔线 |
| 300 | `#c8bfa9` | 线 line-strong | 强调边框 / 输入框描边 |
| 400 | `#b4ab98` | 墨 ink-thin | 最弱文字 / 禁用 / 图标 |
| 500 | `#8a8275` | 墨 ink-faint | 辅助 / 占位文字 |
| 600 | `#6b6356` | 暖褐过渡 | 次级正文 |
| 700 | `#4a453d` | 墨 ink-soft | 次要文字 |
| 800 | `#34302a` | 暖墨过渡 | 深色面板 / 强调文字 |
| 900 | `#2a2722` | 近墨 | 标题 / 深背景 |
| 950 | `#22201c` | 墨 ink（暖墨黑） | 主文字 / **墨 CTA 底色** |

### 2.2 品牌点睛 brand（茶汤橘，11 档）

> 以 chaye `--amber #c56b3e` 为 500 档锚点，向两端展开为可用色阶。
> 用途：选中态 / focus ring / 关键徽章 / 营销点睛。**正文链接固定 600 档 `#a9542c`**（500 `#c56b3e` 在纸底仅 3.39:1，正文不达 AA）。

| 档位 | hex | 说明 |
|------|------|------|
| 50  | `#fbf3ea` | 极淡茶橘底（接近 amber-wash） |
| 100 | `#f4e6d8` | chaye amber-wash · 标签淡底 |
| 200 | `#ecd2b8` | 浅橘描边 |
| 300 | `#dfb088` | 浅橘 |
| 400 | `#d18f5e` | 亮茶橘 |
| 500 | `#c56b3e` | **chaye amber · 主点睛** |
| 600 | `#a9542c` | chaye amber-deep · hover / 链接 |
| 700 | `#8c4424` | 按下态 |
| 800 | `#6f361d` | 深橘 |
| 900 | `#552a18` | 极深橘 |
| 950 | `#3a1d11` | 暗角橘 |

### 2.3 辅助强调 tea（竹青绿，11 档）

> 以 chaye `--tea #3d5a45` 为 600 档锚点。用途：成功态、次级强调、部分图标。

| 档位 | hex | 说明 |
|------|------|------|
| 50  | `#eef2ec` | 竹青极淡底 |
| 100 | `#e7ece2` | chaye tea-wash · 成功淡底 |
| 200 | `#d2dccb` | 浅竹青描边 |
| 300 | `#aebfa6` | 浅竹青 |
| 400 | `#88a081` | 亮竹青 |
| 500 | `#6e8a72` | chaye tea-soft |
| 600 | `#3d5a45` | **chaye tea · 主强调** |
| 700 | `#34503c` | hover |
| 800 | `#2c4232` | chaye tea-deep · 按下 |
| 900 | `#243528` | 深竹青 |
| 950 | `#19251c` | 暗角竹青 |

### 2.4 语义色（浅色）

| 语义 | 主色 | soft 底（徽章/提示背景） | deep（hover/边框） | 来源 |
|------|------|------|------|------|
| success 成功 | `#3d5a45` | `#e7ece2` | `#2c4232` | 竹青绿系 |
| warning 警告 | `#c56b3e` | `#f4e6d8` | `#a9542c` | 茶橘 / 琥珀系 |
| danger 错误 | `#a23b2e` | `#f3e1dc` | `#872f24` | 印章红系（chaye seal） |
| info 信息 | `#3a5570` | `#e3e8ee` | `#2c4256` | 低饱和黛蓝（非高饱和蓝紫，守住宣纸氛围 + 过 lint） |

> info 的 `#3a5570` 是去饱和的靛青 / 黛蓝，避开 `blue/indigo/violet/purple` 关键字 lint，也不破坏宣纸气质。

### 2.5 表面 / 边框 / 文字语义（浅色，velorix 直用 + sub2api 组件参考）

| 角色 | hex | chaye 对应 |
|------|------|------|
| bg 页面背景 | `#f6f2e9` | paper |
| bg-subtle 次级背景 | `#efe8d8` | paper-deep |
| bg-page 内容区底（比卡片深一层） | `#efe8d8` | paper-deep |
| surface 卡片 / 浮层 | `#fbf8f1` | paper-warm |
| border 常规线 | `#ddd5c4` | line |
| border-strong 强调线 | `#c8bfa9` | line-strong |
| ink 主文字 | `#22201c` | ink |
| ink-soft 次要文字 | `#4a453d` | ink-soft |
| muted 辅助文字 | `#8a8275` | ink-faint |
| muted-soft 纯装饰图标 | `#b4ab98` | ink-thin（仅 2.04:1，**禁用于任何需读文字**；占位文字用 muted `#8a8275` 起、关键占位用 primary-600 `#6b6356`） |

---

## 3. 夜墨（暗色）完整令牌表

> chaye 无暗色，这是新设计的一套自洽夜墨：深暖墨底 + 宣纸色文字 + 提亮点睛。
> 原则：背景档暖墨（不是纯黑、不是冷 slate），文字用宣纸偏白的暖灰，点睛色**提亮一档**保证暗底对比度。

### 3.1 夜墨中性灰阶 dark（11 档，沿用 Tailwind 正向：50 最浅 → 950 最深）

> Tailwind `darkMode:'class'` 下，暗色用 `dark:` 前缀引用 `dark-*`。
> **必须沿用 Tailwind 约定（50 最浅、950 最深），只把每档换成暖墨值**——这样现有代码里 `dark:bg-dark-950`(最深) 仍是暗背景、`dark:text-dark-100`(最浅档) 仍是亮文字，组件 class 零改动。
> ⚠️ 切勿把顺序倒过来：若 `dark-950` 变成最浅，`dark:bg-dark-950` 会渲染成近白，暗色整页发白、文字落白底不可读。

| 档位 | hex | 夜墨定位 | 暗色场景用途（现有 class 怎么用的） |
|------|------|------|------|
| 50  | `#f1ebdc` | 宣纸白 | `dark:text-dark-50` 最高对比标题 |
| 100 | `#e3dccb` | 正文白 | `dark:text-dark-100` 主正文 |
| 200 | `#cfc7b6` | 次正文 | 次要正文 |
| 300 | `#b4ab98` | 辅助文字 | 辅助 / 次级文字 |
| 400 | `#8a8275` | 弱文字 | `dark:text-dark-400` 占位 / 弱文字 |
| 500 | `#6b6356` | 更弱 | 禁用文字 / 暗角 |
| 600 | `#5c5649` | 强边框 | `dark:border-dark-600` 强调分隔 |
| 700 | `#46413a` | 常规边框 | `dark:border-dark-700` 边框 |
| 800 | `#34302a` | 深面板 | `dark:bg-dark-800` 输入框 / 抬升面 |
| 900 | `#22201c` | 暖墨底 ink | `dark:bg-dark-900` 页面主背景 |
| 950 | `#1a1815` | 最深底 | `dark:bg-dark-950` 页面最底 / 凹陷区 |

### 3.2 夜墨 brand（提亮茶橘，11 档）

> 暗底需要更亮的橘以保对比，主点睛上移到 400 档亮度。

| 档位 | hex | 说明 |
|------|------|------|
| 50  | `#3a1d11` | 暗底徽章底 |
| 100 | `#552a18` | 深橘底 |
| 200 | `#6f361d` | |
| 300 | `#a9542c` | |
| 400 | `#d18f5e` | **夜墨主点睛（暗底用）** |
| 500 | `#dca878` | hover 提亮 |
| 600 | `#e6c099` | |
| 700 | `#eed3b6` | |
| 800 | `#f3e1cd` | 淡底文字 |
| 900 | `#f7ecdd` | |
| 950 | `#fbf3ea` | 最亮橘 |

### 3.3 夜墨 tea（提亮竹青，11 档）

| 档位 | hex | 说明 |
|------|------|------|
| 50  | `#19251c` | |
| 100 | `#243528` | |
| 200 | `#2c4232` | |
| 300 | `#3d5a45` | |
| 400 | `#6e8a72` | |
| 500 | `#88a081` | **夜墨主竹青（暗底用）** |
| 600 | `#aebfa6` | hover 提亮 |
| 700 | `#c6d2bf` | |
| 800 | `#d9e1d3` | |
| 900 | `#e7ece2` | |
| 950 | `#f0f3ed` | |

### 3.4 语义色（夜墨）

| 语义 | 主色（提亮） | soft 底（暗底徽章） | 说明 |
|------|------|------|------|
| success | `#88a081` | `#243528` | 竹青提亮 |
| warning | `#d18f5e` | `#3a2a1a` | 茶橘提亮 |
| danger | `#d8786a` | `#3a201c` | 印章红提亮（#cf6a5b 在 soft 底仅 4.18:1，提到 #d8786a 达 AA） |
| info | `#7d97b3` | `#1f2a36` | 黛蓝提亮 |

### 3.5 夜墨表面 / 边框 / 文字语义

| 角色 | hex |
|------|------|
| bg 页面背景 | `#22201c` |
| bg-subtle / bg-page | `#1a1815` |
| surface 卡片 / 浮层 | `#2a2722` |
| surface-raised 抬升 | `#34302a` |
| border 常规线 | `#46413a` |
| border-strong 强调线 | `#5c5649` |
| ink 主文字 | `#f1ebdc` |
| ink-soft 次要文字 | `#cfc7b6` |
| muted 辅助文字 | `#b4ab98` |
| muted-soft 占位 | `#8a8275` |

---

## 4. sub2api `tailwind.config.js` 逐令牌映射（可直接抄）

> 文件：`D:/Project/Go/sub2api/frontend/tailwind.config.js`。
> 只改 `theme.extend` 里的值，**类名全部不变**，组件无需改动。
> `primary` / `gray` / `dark` 三组保持 hex 完全一致（浅色阶）；暗色通过 `dark:` 前缀 + 下文 CSS 变量层切换。

```js
/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{vue,js,ts,jsx,tsx}'],
  darkMode: 'class',
  theme: {
    extend: {
      colors: {
        // 中性墨色灰：宣纸 → 暖墨，全程暖褐调，禁止冷中性灰
        primary: {
          50: '#f6f2e9', 100: '#efe8d8', 200: '#ddd5c4', 300: '#c8bfa9',
          400: '#b4ab98', 500: '#8a8275', 600: '#6b6356', 700: '#4a453d',
          800: '#34302a', 900: '#2a2722', 950: '#22201c'
        },
        // 点睛 - 茶汤橘（承接旧日落橙，500 为主点睛、950 接近墨）
        brand: {
          50: '#fbf3ea', 100: '#f4e6d8', 200: '#ecd2b8', 300: '#dfb088',
          400: '#d18f5e', 500: '#c56b3e', 600: '#a9542c', 700: '#8c4424',
          800: '#6f361d', 900: '#552a18', 950: '#3a1d11'
        },
        // 辅助强调 - 竹青绿（成功态 / 次级强调 / 图标，600 为主竹青）
        tea: {
          50: '#eef2ec', 100: '#e7ece2', 200: '#d2dccb', 300: '#aebfa6',
          400: '#88a081', 500: '#6e8a72', 600: '#3d5a45', 700: '#34503c',
          800: '#2c4232', 900: '#243528', 950: '#19251c'
        },
        // gray ≡ primary（浅色主灰，保持与 primary 同源）
        gray: {
          50: '#f6f2e9', 100: '#efe8d8', 200: '#ddd5c4', 300: '#c8bfa9',
          400: '#b4ab98', 500: '#8a8275', 600: '#6b6356', 700: '#4a453d',
          800: '#34302a', 900: '#2a2722', 950: '#22201c'
        },
        // dark - 夜墨阶（暗色背景/文字专用，沿用 Tailwind 正向：50 最浅 → 950 最深；
        //   切勿倒序，否则 dark:bg-dark-950 会变近白、暗色整页发白）
        dark: {
          50: '#f1ebdc', 100: '#e3dccb', 200: '#cfc7b6', 300: '#b4ab98',
          400: '#8a8275', 500: '#6b6356', 600: '#5c5649', 700: '#46413a',
          800: '#34302a', 900: '#22201c', 950: '#1a1815'
        },
        // 语义色（新增，浅色档；暗色由组件 dark: 前缀引用提亮值）
        success: { DEFAULT: '#3d5a45', soft: '#e7ece2', deep: '#2c4232' },
        warning: { DEFAULT: '#c56b3e', soft: '#f4e6d8', deep: '#a9542c' },
        danger:  { DEFAULT: '#a23b2e', soft: '#f3e1dc', deep: '#872f24' },
        info:    { DEFAULT: '#3a5570', soft: '#e3e8ee', deep: '#2c4256' },
        // 印章红（chaye seal，仅 danger / 关键标识，慎用）
        seal: { DEFAULT: '#a23b2e', deep: '#872f24' }
      },
      fontSize: {
        '2xs': ['0.6875rem', { lineHeight: '1rem' }]
      },
      fontFamily: {
        // 正文 / 界面：英文 Inter + 中文 思源黑体，系统栈兜底
        sans: [
          'Inter', '"Noto Sans SC"', 'system-ui', '-apple-system',
          '"Segoe UI"', '"PingFang SC"', '"Microsoft YaHei"', 'sans-serif'
        ],
        // 页面 / 卡片标题：英文 Cormorant Garamond(roman) + 中文 霞鹜文楷
        serif: [
          '"Cormorant Garamond"', '"LXGW WenKai Lite"', '"LXGW WenKai"', '"Noto Serif SC"', '"Songti SC"', 'serif'
        ],
        // 展示数字（统计卡/金额/用量/延迟）：Fraunces，配 lining-nums tabular-nums
        num: ['Fraunces', '"Noto Serif SC"', 'Georgia', 'serif'],
        // 品牌名 wordmark（站点名/Logo 文字）：Cormorant Garamond Italic（配 font-style:italic）
        brand: ['"Cormorant Garamond"', 'serif'],
        mono: ['ui-monospace', 'SFMono-Regular', 'Menlo', 'Monaco', 'Consolas', 'monospace']
      },
      // 圆角整体收紧到近方角（类名不变、只改值）
      borderRadius: {
        none: '0px',
        sm:   '2px',    // chaye radius-xs
        DEFAULT: '3px', // chaye radius（rounded）
        md:   '3px',
        lg:   '4px',    // 原 0.5rem → 收到 4px
        xl:   '6px',    // chaye radius-lg；原 0.75rem → 6px
        '2xl':'8px',    // 原 1rem → 8px
        '3xl':'10px',   // 原 1.5rem → 10px
        '4xl':'12px',   // 原 2rem → 12px
        full: '9999px'  // pill 不变
      },
      boxShadow: {
        // 极淡暖墨投影（rgba 34,32,28 暖墨，非冷黑）
        card:         '0 1px 1px rgba(34,32,28,0.04), 0 2px 6px -1px rgba(34,32,28,0.06)',
        'card-hover': '0 14px 30px -16px rgba(34,32,28,0.22), 0 4px 10px -4px rgba(34,32,28,0.10)',
        elevated:     '0 22px 48px -34px rgba(34,32,28,0.50), 0 6px 16px -6px rgba(34,32,28,0.14)',
        'inner-top':  'inset 0 1px 0 rgba(246,242,233,0.06)'
      },
      backgroundImage: {
        'gradient-radial': 'radial-gradient(var(--tw-gradient-stops))'
      },
      // animation / keyframes / backdropBlur 保持原样（与配色无关）
      animation: {
        'fade-in': 'fadeIn 0.3s ease-out',
        'slide-up': 'slideUp 0.3s ease-out',
        'slide-down': 'slideDown 0.3s ease-out',
        'slide-in-right': 'slideInRight 0.3s ease-out',
        'scale-in': 'scaleIn 0.2s ease-out',
        'pulse-slow': 'pulse 3s cubic-bezier(0.4, 0, 0.6, 1) infinite'
      },
      keyframes: {
        fadeIn: { '0%': { opacity: '0' }, '100%': { opacity: '1' } },
        slideUp: { '0%': { opacity: '0', transform: 'translateY(10px)' }, '100%': { opacity: '1', transform: 'translateY(0)' } },
        slideDown: { '0%': { opacity: '0', transform: 'translateY(-10px)' }, '100%': { opacity: '1', transform: 'translateY(0)' } },
        slideInRight: { '0%': { opacity: '0', transform: 'translateX(20px)' }, '100%': { opacity: '1', transform: 'translateX(0)' } },
        scaleIn: { '0%': { opacity: '0', transform: 'scale(0.98)' }, '100%': { opacity: '1', transform: 'scale(1)' } }
      },
      backdropBlur: { xs: '2px' }
    }
  },
  plugins: []
}
```

> 说明：`borderRadius` 这里显式列出全档位（含 Tailwind 默认档），把 `lg/xl/2xl/3xl/4xl` 整体调小，使现有大量 `rounded-lg/xl/2xl` 自动贴近方角，**组件 class 不动**。若担心 `sm`/`DEFAULT` 覆盖默认值影响过大，可只保留 `lg`/`xl`/`2xl`/`3xl`/`4xl` 五档覆盖，其余删去让 Tailwind 用默认。

### 4.1 主 CTA / 链接 / focus ring 落地约定（sub2api）

| 元素 | 旧（Zinc + 橙） | 新（墨 + 茶橘） |
|------|------|------|
| 主 CTA 底色 | `bg-primary-950`（冷黑 #09090b） | `bg-primary-950`（暖墨黑 #22201c）— 类名不变，值已换 |
| 主 CTA hover | `bg-black` | `bg-primary-900`（#2a2722） |
| 选中态 / focus ring | `ring-brand-600/40`（橙） | `ring-2 ring-brand-600/40`（#a9542c，2px，纸底约 4.7:1）— 现有 `.btn` 已是此写法 |
| 正文链接 | `text-brand-600` | `text-brand-600`（#a9542c） |
| 成功徽章 | green 内置 | `bg-success-soft text-success` |

---

## 5. velorix `styles.css` CSS 变量映射（可直接抄）

> 文件：`D:/Project/Electron/velorix-desktop/src/renderer/src/styles.css`（第 5–43 行 `:root`）。
> 替换整段 `:root`，并新增 `:root.dark`（或 `[data-theme="dark"]`）夜墨覆盖。类名 `.btn-primary` 等全部不变。

```css
:root {
  /* ---- 中性墨色（宣纸 → 暖墨） ---- */
  --bg: #f6f2e9;            /* 宣纸主背景 paper */
  --bg-subtle: #efe8d8;     /* paper-deep */
  --bg-page: #efe8d8;       /* 内容区底，比卡片深一层 */
  --surface: #fbf8f1;       /* 卡片/浮层 paper-warm */
  --surface-dark: #2a2722;  /* 深面板（原 #0b0b0d 统一到此） */
  --border: #ddd5c4;        /* line */
  --border-strong: #c8bfa9; /* line-strong */
  --ink: #22201c;           /* 主文字 ink（暖墨黑） */
  --ink-soft: #4a453d;      /* ink-soft */
  --muted: #8a8275;         /* ink-faint */
  --muted-soft: #b4ab98;    /* ink-thin */

  /* ---- 品牌茶汤橘（主点睛，承接旧橙） ---- */
  --brand: #c56b3e;         /* amber 主点睛（原 #ea580c） */
  --brand-bright: #d18f5e;  /* 提亮 */
  --brand-deep: #a9542c;    /* amber-deep · hover/按下 */
  --brand-soft: #f4e6d8;    /* amber-wash · 淡底 */
  --brand-ring: rgba(197, 107, 62, 0.4);

  /* ---- 竹青绿（辅助强调） ---- */
  --tea: #3d5a45;
  --tea-deep: #2c4232;
  --tea-soft: #6e8a72;
  --tea-wash: #e7ece2;

  /* ---- 语义色 ---- */
  --success: #3d5a45;       /* 竹青绿 */
  --success-soft: #e7ece2;
  --warning: #c56b3e;       /* 茶橘 */
  --warning-soft: #f4e6d8;
  --danger: #a23b2e;        /* 印章红 seal */
  --danger-soft: #f3e1dc;
  --danger-deep: #872f24;
  --info: #3a5570;          /* 低饱和黛蓝 */
  --info-soft: #e3e8ee;

  /* ---- 阴影（极淡暖墨投影） ---- */
  --shadow-card: 0 1px 1px rgba(34,32,28,0.04), 0 2px 6px -1px rgba(34,32,28,0.06);
  --shadow-pop:  0 22px 48px -34px rgba(34,32,28,0.50), 0 6px 16px -6px rgba(34,32,28,0.14);

  /* ---- 圆角（近方角） ---- */
  --radius: 6px;            /* 原 12px → chaye radius-lg */
  --radius-sm: 3px;         /* 原 8px → chaye radius */
  --radius-xs: 2px;
  --radius-pill: 999px;

  --ease-out: cubic-bezier(0.22, 1, 0.36, 1);

  /* ---- 字体 ---- */
  --font: 'Inter', 'Noto Sans SC', system-ui, -apple-system, 'Segoe UI', 'PingFang SC', 'Microsoft YaHei', sans-serif;
  --serif: 'Cormorant Garamond', 'LXGW WenKai Lite', 'LXGW WenKai', 'Noto Serif SC', 'Songti SC', serif;
  --num: 'Fraunces', 'Noto Serif SC', Georgia, serif;
  --wordmark: 'Cormorant Garamond', serif;   /* 品牌名 wordmark，配 font-style:italic */
  --mono: ui-monospace, 'SFMono-Regular', Menlo, Consolas, monospace;
}

/* ============ 夜墨暗色 ============ */
:root.dark {
  --bg: #22201c;
  --bg-subtle: #1a1815;
  --bg-page: #1a1815;
  --surface: #2a2722;
  --surface-dark: #1a1815;
  --border: #46413a;
  --border-strong: #5c5649;
  --ink: #f1ebdc;
  --ink-soft: #cfc7b6;
  --muted: #b4ab98;
  --muted-soft: #8a8275;

  --brand: #d18f5e;          /* 暗底提亮茶橘 */
  --brand-bright: #dca878;
  --brand-deep: #a9542c;
  --brand-soft: #3a2a1a;
  --brand-ring: rgba(209, 143, 94, 0.45);

  --tea: #88a081;            /* 暗底提亮竹青 */
  --tea-deep: #3d5a45;
  --tea-soft: #6e8a72;
  --tea-wash: #243528;

  --success: #88a081;       --success-soft: #243528;  --success-deep: #3d5a45;
  --warning: #d18f5e;       --warning-soft: #3a2a1a;  --warning-deep: #c56b3e;
  --danger: #d8786a;        --danger-soft: #3a201c;   --danger-deep: #cf6a5b;
  --info: #7d97b3;          --info-soft: #1f2a36;      --info-deep: #5b7794;

  --shadow-card: 0 1px 1px rgba(0,0,0,0.30), 0 2px 6px -1px rgba(0,0,0,0.40);
  --shadow-pop:  0 22px 48px -30px rgba(0,0,0,0.70), 0 6px 16px -6px rgba(0,0,0,0.45);
}
```

### 5.1 velorix 关键类映射（值变、类名不变）

| 类 | 旧 | 新 |
|------|------|------|
| `.btn-primary` 底色 | `--ink`(#18181b) hover `#000` | `--ink`(#22201c) hover `--ink-soft`/`#2a2722` |
| `.btn-primary` 文字 | `#fff` | `#f6f2e9`（宣纸白，暖一点不刺眼） |
| `:focus-visible` ring | `2px solid var(--brand-ring)` | 同写法，值已换茶橘 ring |
| `.badge-brand` | `--brand-soft`/`--brand` | 同写法，值已换茶橘 |
| `.titlebar.dark` / 深面板 `#0b0b0d` | 散落硬编码 | 统一改引 `var(--surface-dark)` |

---

## 6. 字体自托管方案

两端**统一走 `@fontsource/*` 与 `lxgw-wenkai-lite-webfont` 分片自托管**（自带 `font-display: swap` + unicode-range 分片，Electron 离线可用，免手动子集），只引实际用到的字重。

最终体系按用途分四类：正文/界面、标题、展示数字、品牌名 wordmark，各自字栈与包名见下表。

### 6.1 最终字体总表

| 用途 | 英文字体 | 中文字体 | 字栈写法 | 包名 / @fontsource |
|------|---------|---------|---------|---------------------|
| 正文 / 界面 | Inter | 思源黑体 | `'Inter', 'Noto Sans SC', system-ui, -apple-system, 'Segoe UI', 'PingFang SC', 'Microsoft YaHei', sans-serif` | `@fontsource/inter`、`@fontsource/noto-sans-sc` |
| 页面 / 卡片标题 | Cormorant Garamond（roman·非斜体） | 霞鹜文楷 | `'Cormorant Garamond', 'LXGW WenKai Lite', 'LXGW WenKai', 'Noto Serif SC', 'Songti SC', serif` | `@fontsource/cormorant-garamond`、`lxgw-wenkai-lite-webfont`、`@fontsource/noto-serif-sc` |
| 展示数字（统计卡/金额/用量/延迟等可核对数字） | Fraunces | —（数字走 Fraunces，中文兜底宋体） | `'Fraunces', 'Noto Serif SC', Georgia, serif` + `font-variant-numeric: lining-nums tabular-nums` | `@fontsource/fraunces` |
| 品牌名 wordmark（站点名/Logo 文字） | Cormorant Garamond Italic | — | `'Cormorant Garamond', serif` + `font-style: italic` | `@fontsource/cormorant-garamond`（含 italic 分片） |

> 展示数字**必须**叠加 `font-variant-numeric: lining-nums tabular-nums`，保证等宽对齐、避免 Fraunces 默认旧体数字（oldstyle）影响核对。
> 品牌名 wordmark 用 Cormorant Garamond **斜体**，与标题的 roman 正体区分。

### 6.2 安装命令

新增 `@fontsource/fraunces`、`@fontsource/inter`、`lxgw-wenkai-lite-webfont`，原 `noto-sans-sc` / `noto-serif-sc` / `cormorant-garamond` 保留。

sub2api：

```bash
pnpm --dir frontend add @fontsource/inter @fontsource/noto-sans-sc lxgw-wenkai-lite-webfont @fontsource/noto-serif-sc @fontsource/cormorant-garamond @fontsource/fraunces
```

velorix：

```bash
pnpm add @fontsource/inter @fontsource/noto-sans-sc lxgw-wenkai-lite-webfont @fontsource/noto-serif-sc @fontsource/cormorant-garamond @fontsource/fraunces
```

### 6.3 main.ts import 清单（替换地基期那批，放在全局样式 import 之前）

sub2api `frontend/src/main.ts`（在 `import './style.css'` 之前）、velorix `src/renderer/src/main.ts`（在 `import './styles.css'` 之前）统一引入：

```ts
import '@fontsource/inter/400.css'
import '@fontsource/inter/500.css'
import '@fontsource/inter/600.css'
import '@fontsource/inter/700.css'
import '@fontsource/noto-sans-sc/400.css'
import '@fontsource/noto-sans-sc/500.css'
import '@fontsource/noto-sans-sc/700.css'
import 'lxgw-wenkai-lite-webfont/lxgwwenkailite-regular.css'   // 精简版(常用字)，标题用，~4MB 分片按需
import '@fontsource/fraunces/500.css'
import '@fontsource/fraunces/600.css'
import '@fontsource/cormorant-garamond/500.css'
import '@fontsource/cormorant-garamond/600.css'
import '@fontsource/cormorant-garamond/500-italic.css'
import '@fontsource/cormorant-garamond/600-italic.css'
```

`tailwind.config.js` 的 `fontFamily`（sans/serif/num/brand）已在第 4 节给出；CSS 变量 `--font` / `--serif` / `--num` / `--wordmark` 已在第 5 节接好。

### 6.4 体积与 FOUT 约束

- 中文**绝不全量**（8–10MB 单文件不可接受）；fontsource 与霞鹜文楷分片按需下载，首屏通常只拉几个分片共几百 KB。
- 统一 `font-display: swap`（fontsource / lxgw-wenkai-lite-webfont 内置），先系统字体渲染、到位再替换，规避 FOIT 空白。
- 仅给首屏点睛字体（Cormorant 标题、Fraunces 数字）做 preload，中文分片**不** preload。
- 备选（不加 npm 依赖）：手动 `pyftsubset` 生成霞鹜文楷 / Noto Serif SC 常用 3500 字子集（约 1–1.5MB），放 sub2api `frontend/public/fonts/`、velorix `src/renderer/public/fonts/`，`@font-face url(...)` 自写，绝对路径 `/fonts/x.woff2` 引用。

### 6.5 排版基线（两端 body / 标题 / 数字 / wordmark）

```css
body { font-family: var(--font); line-height: 1.7; }
h1, h2, h3, h4 { font-family: var(--serif); font-weight: 600; line-height: 1.25; letter-spacing: 0.01em; }
.num, [data-num] { font-family: var(--num); font-variant-numeric: lining-nums tabular-nums; }
.wordmark { font-family: var(--wordmark); font-style: italic; }
```

### 6.6 夜墨暗色修正

实心彩色按钮（warning / success / danger 一类，暗色下提亮为柔色底）在暗色下**文字必须改深墨字 `#22201c`**，不能用白字——白字落在提亮柔色底上对比不足，会发虚、发粉。

```css
:root.dark .btn-warning,
:root.dark .btn-success,
:root.dark .btn-danger { color: #22201c; }
```

---

## 7. 宣纸纹理实现

### 7.1 浅色（sub2api `style.css` / velorix `styles.css`，全局一次）

```css
body::before {
  content: '';
  position: fixed;
  inset: 0;
  z-index: 0;            /* 在内容之下；#app 需 position:relative; z-index:1 */
  pointer-events: none;
  opacity: 0.04;
  background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='160' height='160'%3E%3Cfilter id='n'%3E%3CfeTurbulence type='fractalNoise' baseFrequency='0.8' numOctaves='2' stitchTiles='stitch'/%3E%3C/filter%3E%3Crect width='100%25' height='100%25' filter='url(%23n)'/%3E%3C/svg%3E");
}
#app { position: relative; z-index: 1; }
```

> 注意 sub2api 用 Tailwind，`#app` 可能已有定位；确认 `body::before` 的 `z-index:0` 不盖住交互（已 `pointer-events:none`）。velorix 的 `#app` 已 `height:100%`，补 `position:relative; z-index:1`。

### 7.2 夜墨暗色纹理处理

暗底上同样的亮噪点会显脏，**降低不透明度并叠为亮色微光纹理**（或直接关闭）：

```css
:root.dark body::before,        /* velorix */
.dark body::before {            /* sub2api（darkMode:'class' 挂 html.dark） */
  opacity: 0.025;               /* 降到 0.02~0.03，暗底噪点更克制 */
  mix-blend-mode: screen;       /* 让噪点表现为极淡亮颗粒而非黑点，避免发脏 */
}
```

> 若暗色下仍觉脏，最稳妥是 `display: none` 关闭暗色纹理——夜墨靠纯暖墨底已足够有质感。

---

## 8. 语义色 / 圆角 / 阴影规范

### 8.1 语义色规范

- **成功 = 竹青绿系**（`#3d5a45` / soft `#e7ece2`）；**警告 = 茶橘 / 琥珀**（`#c56b3e` / soft `#f4e6d8`）；**错误 = 印章红**（`#a23b2e` / soft `#f3e1dc`）；**信息 = 低饱和黛蓝**（`#3a5570`，**禁高饱和蓝紫**，既守宣纸氛围又过 lint 规则2）。
- **功能性多色保留不收敛**：模型 / 厂商品牌色（BrandIcon / ModelIcon / ContactMethodIcon）、支付品牌色（Stripe / 支付宝 / 微信）、图表 12 色数据系列、终端 ANSI 调色板（xterm 不解析 `var()`，必须真实色名）——这些是 A 类功能色，换肤一律不动。
- 警告色与品牌点睛同为茶橘，需靠**形态 + 图标 + 文案**区分（warning 用三角警示图标 + soft 底块），避免语义混淆。

### 8.2 圆角规范（近方角 chaye 风）

| 令牌 | 值 | 用途 |
|------|------|------|
| xs / sm | `2px` | 徽章 / 角标 / 小标签 |
| DEFAULT | `3px` | 按钮 / 输入框 / 普通卡片 |
| lg | `4px` | 中卡片 |
| xl | `6px` | 大卡片 / 浮层（chaye radius-lg 上限） |
| 2xl / 3xl / 4xl | `8 / 10 / 12px` | 兜底大容器（仅极少数特大面板） |
| pill / full | `999px` | 胶囊标签 / 头像 / 开关 |

> sub2api 大量 `rounded-lg/xl/2xl` 通过第 4 节 `borderRadius` 改值整体收紧，**不改组件 class**。

### 8.3 阴影规范（极淡暖墨投影）

| 级别 | 值 | 用途 |
|------|------|------|
| card | `0 1px 1px rgba(34,32,28,0.04), 0 2px 6px -1px rgba(34,32,28,0.06)` | 静止卡片 |
| card-hover | `0 14px 30px -16px rgba(34,32,28,0.22), 0 4px 10px -4px rgba(34,32,28,0.10)` | 卡片 hover |
| elevated | `0 22px 48px -34px rgba(34,32,28,0.50), 0 6px 16px -6px rgba(34,32,28,0.14)` | 下拉 / 弹窗 / 浮层 |

> 投影色统一用暖墨 `rgba(34,32,28,…)`（即 ink `#22201c`），**不用冷黑 `16,24,40`**。暗色下投影改纯黑 `rgba(0,0,0,…)` 提高暗底可见度（见第 5 节夜墨）。chaye 风偏好「靠细线区分层级」，可在卡片上优先用 `border` + 极淡 `card` 阴影，而非厚重投影。

---

## 9. 两端关键迁移点与风险清单

1. **三组中性灰必须同步**：`primary` / `gray` / `dark`（浅色阶）逐档 hex 保持一致，否则明暗切换色温割裂。velorix 侧对应 `--bg/--ink/--border` 一族要同源。
2. **B 类硬编码是最大工作量**：sub2api 约 300+ 处把 Zinc/slate 裸 `rgb()` 写进 `<style>`（`SafetyRiskView.vue` ~75、`ModelIntegrationDocsView.vue` ~33、`ImageGenView.vue` ~32、`UserDocsView.vue` ~24、`AppSidebar.vue` ~18、`SettingsView.vue` ~16）。换肽只改 config 不动这些文件，**这些硬编码区域不会变色**，必须逐个替换为令牌或 `var()`。velorix 同理（`#fff` 当背景、`#fff7ed/#fed7aa/#c2410c` 旧橙、`#0b0b0d` 深面板散落）。
3. **A 类功能色不要误改**：模型/支付品牌色、图表 12 色、终端 ANSI 必须保留。误把品牌色令牌化会破坏识别性；xterm 主题用 `var()` 会直接失效（xterm 不解析 CSS 变量）。
4. **lint 红线（仅一条真陷阱）**：`check-style-discipline.mjs` 接入 CI（backend-ci.yml frontend job），拦 `text-[≤16px]` 任意值和 `from/via/to-blue|indigo|violet|purple` 渐变。换肤排版微调时**禁止写 `text-[11px]` 等小字号任意值**，统一走 `text-2xs/xs/sm/base`；info 色已避开蓝紫关键字。字体/纹理/圆角/阴影/serif 全在 lint 盲区，零风险。
5. **既有违规蓝紫渐变（借换肤一并清理）**：`ModelIcon.vue:273` 的 `linear-gradient(135deg,#6366f1,#8b5cf6)`（在 `<style>` 块、lint 抓不到）；另有 `badgeTone.ts` / `platformColors.ts` / `useChannelMonitorFormat.ts` 共约 9 处 `from/via/to-(blue|indigo|violet|purple)` 蓝紫渐变（在 `.ts` 里、不在 lint 扫描面、当前不报错但违背设计纪律）。统一换成 `brand` / `tea` / `info` 令牌。
6. **圆角整体变小的视觉风险**：sub2api 大量 `rounded-2xl`（原 16px → 8px）、`rounded-xl`（原 12px → 6px）会让卡片明显更方。先在 1–2 个高频页面（Dashboard、Sidebar）灰度验证观感，再全量。velorix `--radius` 12→6 同理。
7. **暗色提亮点睛的对比度**：夜墨 brand 主点睛上移到 400 档（`#d18f5e`）、tea 到 500 档（`#88a081`），务必在真实暗底校验 WCAG AA（正文链接 ≥4.5:1，大字/图标 ≥3:1）。
8. **CTA 文字色**：墨 CTA 上文字从纯白改宣纸白 `#f6f2e9` 更柔和；若对比不足（暖墨黑 #22201c 上宣纸白对比约 13:1，安全）保持即可。
9. **字体首屏体积**：中文 fontsource / 霞鹜文楷分片按需加载，但标题中文（霞鹜文楷）若用于大量正文会拉分片；建议**霞鹜文楷仅限标题、Fraunces 仅限数字点睛**，正文走 Inter + 思源黑体，控制首屏请求数。Cormorant 标题、Fraunces 数字可 preload，中文分片不 preload。
10. **velorix `--radius` 与组件圆角联动**：`.btn`(--radius-sm)、`.card`(--radius)、`.badge`(pill) 全靠变量，改 `:root` 即生效；但散落硬编码 `border-radius:12px` 的组件需单独排查替换。
11. **`surface-dark` 统一**：velorix 旧 `#0b0b0d` 在 `.titlebar.dark` / `LoginView` / `ChatView` 代码块 3+ 处重复，统一引 `var(--surface-dark)`，避免夜墨与这些深面板色温脱节。
12. **纹理性能/观感**：`body::before` 固定全屏 SVG 噪点开销极低，但暗色下务必降透明度或关闭（见 7.2），否则暗底显脏。i18n / 截图测试不受影响（check-i18n-keys.mjs 与视觉无关、且从不 exit 1）。

---

*令牌口径以本文件为准，组件内禁止再写死色值；新增颜色一律落到上述令牌或 `var()`。*
