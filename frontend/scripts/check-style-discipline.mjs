// 样式纪律检查：在 lint 阶段拦截违反设计系统的写法，防止视觉收敛成果回退。
// 规则 1：禁止 ≤16px 的任意值字号（text-[10px] 等），必须用 text-2xs/xs/sm/base 标准档
// 规则 2：禁止装饰性蓝紫渐变（from/via/to-blue|indigo|violet|purple），与"黑 CTA + 暖橙点睛"冲突
import { readdirSync, readFileSync, statSync } from 'node:fs'
import { join, relative } from 'node:path'
import process from 'node:process'

const SRC_DIR = new URL('../src', import.meta.url).pathname.replace(/^\/([A-Za-z]:)/, '$1')

const RULES = [
  {
    name: '任意值小字号（请改用 text-2xs / text-xs / text-sm / text-base）',
    pattern: /text-\[(?:[0-9]|1[0-6])px\]/g
  },
  {
    name: '装饰性蓝紫渐变（与品牌"黑 CTA + 暖橙点睛"冲突）',
    pattern: /(?:from|via|to)-(?:blue|indigo|violet|purple)-\d+/g
  }
]

// 递归收集需要检查的源文件（.vue 模板与全局样式）
function collectFiles(dir) {
  const results = []
  for (const name of readdirSync(dir)) {
    if (name === 'node_modules' || name === '__tests__') continue
    const full = join(dir, name)
    if (statSync(full).isDirectory()) {
      results.push(...collectFiles(full))
    } else if (name.endsWith('.vue') || name === 'style.css') {
      results.push(full)
    }
  }
  return results
}

const violations = []
for (const file of collectFiles(SRC_DIR)) {
  const content = readFileSync(file, 'utf8')
  const lines = content.split('\n')
  for (const rule of RULES) {
    lines.forEach((line, index) => {
      const matches = line.match(rule.pattern)
      if (matches) {
        violations.push(
          `${relative(process.cwd(), file)}:${index + 1} [${rule.name}] ${matches.join(', ')}`
        )
      }
    })
  }
}

if (violations.length > 0) {
  console.error(`样式纪律检查失败，共 ${violations.length} 处违规：\n`)
  for (const violation of violations) console.error('  ' + violation)
  process.exit(1)
}
console.log('样式纪律检查通过')
