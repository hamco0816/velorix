/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{vue,js,ts,jsx,tsx}'],
  darkMode: 'class',
  theme: {
    extend: {
      colors: {
        // 主色调 - Zinc 中性纯黑灰，配纯黑按钮做"白底 + 黑 CTA + 暖橙点睛"的高级感
        primary: {
          50: '#fafafa',
          100: '#f4f4f5',
          200: '#e4e4e7',
          300: '#d4d4d8',
          400: '#a1a1aa',
          500: '#71717a',
          600: '#52525b',
          700: '#3f3f46',
          800: '#27272a',
          900: '#18181b',
          950: '#09090b'
        },
        // 点睛色 - Orange 日落橙，蓝橙互补的经典专业暖色（替代深棕琥珀，更鲜活吉利）
        brand: {
          50: '#fff7ed',
          100: '#ffedd5',
          200: '#fed7aa',
          300: '#fdba74',
          400: '#fb923c',
          500: '#f97316',
          600: '#ea580c',
          700: '#c2410c',
          800: '#9a3412',
          900: '#7c2d12',
          950: '#431407'
        },
        // 中性灰统一：gray（浅色模式主灰）与 dark（暗色模式背景/卡片）都重映射为
        // 与 primary 同源的 Zinc 色温，避免浅色 gray / 暗色 slate / 令牌 zinc 三种
        // 色温的灰混用导致明暗切换时观感不一致
        gray: {
          50: '#fafafa',
          100: '#f4f4f5',
          200: '#e4e4e7',
          300: '#d4d4d8',
          400: '#a1a1aa',
          500: '#71717a',
          600: '#52525b',
          700: '#3f3f46',
          800: '#27272a',
          900: '#18181b',
          950: '#09090b'
        },
        dark: {
          50: '#fafafa',
          100: '#f4f4f5',
          200: '#e4e4e7',
          300: '#d4d4d8',
          400: '#a1a1aa',
          500: '#71717a',
          600: '#52525b',
          700: '#3f3f46',
          800: '#27272a',
          900: '#18181b',
          950: '#09090b'
        }
      },
      fontSize: {
        // 微型字号：徽章/角标/表格密集场景专用，统一收编历史上的 text-[9~11px] 任意值
        '2xs': ['0.6875rem', { lineHeight: '1rem' }]
      },
      fontFamily: {
        sans: [
          'system-ui',
          '-apple-system',
          'BlinkMacSystemFont',
          'Segoe UI',
          'Roboto',
          'Helvetica Neue',
          'Arial',
          'PingFang SC',
          'Hiragino Sans GB',
          'Microsoft YaHei',
          'sans-serif'
        ],
        mono: ['ui-monospace', 'SFMono-Regular', 'Menlo', 'Monaco', 'Consolas', 'monospace']
      },
      boxShadow: {
        // 分层柔和阴影：贴地接触阴影 + 环境阴影，静止克制、有真实立体感（Stripe/Linear 质感）
        card: '0 1px 1px rgba(16, 24, 40, 0.03), 0 2px 5px -1px rgba(16, 24, 40, 0.05)',
        'card-hover': '0 10px 28px -8px rgba(16, 24, 40, 0.14), 0 3px 8px -3px rgba(16, 24, 40, 0.07)',
        // 浮层（下拉/弹窗）专用：更明显的悬浮投影，让浮层"脱离"页面
        elevated: '0 14px 36px -10px rgba(16, 24, 40, 0.20), 0 4px 12px -4px rgba(16, 24, 40, 0.10)',
        // 顶部内嵌高光，让深色按钮有金属质感（不靠外阴影发光）
        'inner-top': 'inset 0 1px 0 rgba(255, 255, 255, 0.08)'
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
      borderRadius: {
        '4xl': '2rem'
      }
    }
  },
  plugins: []
}
