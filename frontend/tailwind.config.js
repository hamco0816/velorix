/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{vue,js,ts,jsx,tsx}'],
  darkMode: 'class',
  theme: {
    extend: {
      colors: {
        // 中性墨色灰：宣纸 → 暖墨，全程暖褐调，配暖墨黑按钮做"宣纸底 + 墨 CTA + 茶橘点睛"的高级感
        primary: {
          50: '#f6f2e9',
          100: '#efe8d8',
          200: '#ddd5c4',
          300: '#c8bfa9',
          400: '#b4ab98',
          500: '#736a5c',
          600: '#6b6356',
          700: '#4a453d',
          800: '#34302a',
          900: '#2a2722',
          950: '#22201c'
        },
        // 点睛色 - 茶汤橘（承接旧日落橙，500 为主点睛、950 接近墨）
        // 提饱和版：仍是茶橘色相，但整体提亮提饱和，让功能性点睛/警告/暗色文字更鲜明
        brand: {
          50: '#fdf3e9',
          100: '#f8e3cd',
          200: '#f0c79e',
          300: '#e7a86f',
          400: '#dd8a44',
          500: '#cf6f2c',
          600: '#a44f1c',
          700: '#8f4519',
          800: '#723715',
          900: '#562a12',
          950: '#3a1c0c'
        },
        // 辅助强调 - 竹青绿（成功态 / 次级强调 / 图标，600 为主竹青）
        // 提饱和版：仍是竹青色相，但明显更绿更亮，避免在卡片上发灰发土
        tea: {
          50: '#edf5ef',
          100: '#d8ecdc',
          200: '#b8dcc0',
          300: '#8fcaa0',
          400: '#5fb27e',
          500: '#43965f',
          600: '#3a7d52',
          700: '#2f6644',
          800: '#285436',
          900: '#21422c',
          950: '#16301e'
        },
        // 中性灰统一：gray（浅色模式主灰）与 primary 同源，逐档 hex 完全一致，
        // 避免浅色与令牌灰色温混用导致明暗切换观感不一致
        gray: {
          50: '#f6f2e9',
          100: '#efe8d8',
          200: '#ddd5c4',
          300: '#c8bfa9',
          400: '#b4ab98',
          500: '#736a5c',
          600: '#6b6356',
          700: '#4a453d',
          800: '#34302a',
          900: '#2a2722',
          950: '#22201c'
        },
        // dark - 夜墨阶（暗色背景/文字专用，沿用 Tailwind 正向：50 最浅 → 950 最深；
        //   切勿倒序，否则 dark:bg-dark-950 会变近白、暗色整页发白）
        dark: {
          50: '#f1ebdc',
          100: '#e3dccb',
          200: '#cfc7b6',
          300: '#b4ab98',
          400: '#8a8275',
          500: '#6b6356',
          600: '#5c5649',
          700: '#46413a',
          800: '#34302a',
          900: '#22201c',
          950: '#1a1815'
        },
        // 语义色（浅色档；暗色由组件 dark: 前缀引用 tea/brand 提亮档）
        // 提饱和版：保持宣纸色相（竹青/茶橘/印章红/黛蓝），但提亮提饱和，让状态/数据徽章鲜明可辨
        // 语义色（v3.1 对比度修订）：DEFAULT 调到在自身 soft 底上也达 WCAG AA(≥4.5)，徽章文字+图标统一可读；
        // warning 顺势从茶橘转深琥珀，与 brand 茶橘点睛区分（原二者同色易混）
        success: { DEFAULT: '#347049', soft: '#d8ecdc', deep: '#2f6644' },
        warning: { DEFAULT: '#9c5712', soft: '#fbecd2', deep: '#7a440e' },
        danger: { DEFAULT: '#b23425', soft: '#f8ddd5', deep: '#9c2b1f' },
        info: { DEFAULT: '#3a6499', soft: '#dde6f2', deep: '#2d5481' },
        // 印章红（chaye seal，仅 danger / 关键标识，慎用）
        seal: { DEFAULT: '#b23425', deep: '#9c2b1f' }
      },
      fontSize: {
        // 微型字号：徽章/角标/表格密集场景专用，统一收编历史上的 text-[9~11px] 任意值
        '2xs': ['0.6875rem', { lineHeight: '1rem' }]
      },
      fontFamily: {
        // 正文/界面：英文 Inter + 中文思源黑体，系统栈兜底
        sans: [
          'Inter',
          'Noto Sans SC',
          'system-ui',
          '-apple-system',
          'BlinkMacSystemFont',
          'Segoe UI',
          'Roboto',
          'PingFang SC',
          'Hiragino Sans GB',
          'Microsoft YaHei',
          'sans-serif'
        ],
        // 页面/卡片标题：英文 Cormorant Garamond（罗马正体）+ 中文霞鹜文楷
        serif: ['"Cormorant Garamond"', '"LXGW WenKai Lite"', '"LXGW WenKai"', '"Noto Serif SC"', 'Songti SC', 'Georgia', 'serif'],
        // 展示数字（统计/金额/用量/延迟）：Fraunces，配 lining-nums tabular-nums
        num: ['Fraunces', '"Noto Serif SC"', 'Georgia', 'serif'],
        // 品牌名 wordmark：Cormorant Garamond Italic（配 italic class）
        display: ['"Cormorant Garamond"', 'serif'],
        mono: ['ui-monospace', 'SFMono-Regular', 'Menlo', 'Monaco', 'Consolas', 'monospace']
      },
      boxShadow: {
        // 极淡暖墨投影（rgba 34,32,28 暖墨，非冷黑），静止克制、有真实立体感
        card: '0 1px 1px rgba(34,32,28,0.04), 0 2px 6px -1px rgba(34,32,28,0.06)',
        'card-hover': '0 14px 30px -16px rgba(34,32,28,0.22), 0 4px 10px -4px rgba(34,32,28,0.10)',
        // 浮层（下拉/弹窗）专用：更明显的悬浮投影，让浮层"脱离"页面
        elevated: '0 22px 48px -34px rgba(34,32,28,0.50), 0 6px 16px -6px rgba(34,32,28,0.14)',
        // 顶部内嵌高光，让深色按钮有金属质感（用宣纸白微光，不靠外阴影发光）
        'inner-top': 'inset 0 1px 0 rgba(246,242,233,0.06)'
      },
      backgroundImage: {
        'gradient-radial': 'radial-gradient(var(--tw-gradient-stops))'
      },
      animation: {
        'fade-in': 'fadeIn 0.3s ease-out',
        'slide-up': 'slideUp 0.3s ease-out',
        'slide-down': 'slideDown 0.3s ease-out',
        'slide-in-right': 'slideInRight 0.3s ease-out',
        'scale-in': 'scaleIn 0.2s ease-out',
        'pulse-slow': 'pulse 3s cubic-bezier(0.4, 0, 0.6, 1) infinite'
      },
      keyframes: {
        fadeIn: {
          '0%': { opacity: '0' },
          '100%': { opacity: '1' }
        },
        slideUp: {
          '0%': { opacity: '0', transform: 'translateY(10px)' },
          '100%': { opacity: '1', transform: 'translateY(0)' }
        },
        slideDown: {
          '0%': { opacity: '0', transform: 'translateY(-10px)' },
          '100%': { opacity: '1', transform: 'translateY(0)' }
        },
        slideInRight: {
          '0%': { opacity: '0', transform: 'translateX(20px)' },
          '100%': { opacity: '1', transform: 'translateX(0)' }
        },
        scaleIn: {
          '0%': { opacity: '0', transform: 'scale(0.98)' },
          '100%': { opacity: '1', transform: 'scale(1)' }
        }
      },
      backdropBlur: {
        xs: '2px'
      },
      // 圆角整体收紧到近方角（类名不变、只改值）
      borderRadius: {
        none: '0px',
        sm: '2px', // chaye radius-xs
        DEFAULT: '3px', // chaye radius（rounded）
        md: '3px',
        lg: '4px', // 原 0.5rem → 收到 4px
        xl: '6px', // chaye radius-lg；原 0.75rem → 6px
        '2xl': '8px', // 原 1rem → 8px
        '3xl': '10px', // 原 1.5rem → 10px
        '4xl': '12px', // 原 2rem → 12px
        full: '9999px' // pill 不变
      }
    }
  },
  plugins: []
}
