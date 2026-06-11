// i18n key 审计：扫描 src 下所有字面量 t('key') 引用，检查 zh/en 词典是否缺 key
// 用法：node scripts/check-i18n-keys.mjs（以 _ 结尾的 key 视为动态拼接前缀，跳过）
import { readFileSync, readdirSync, statSync, rmSync, realpathSync } from 'node:fs'
import { join, dirname } from 'node:path'
import { fileURLToPath, pathToFileURL } from 'node:url'
import { createRequire } from 'node:module'

const root = join(dirname(fileURLToPath(import.meta.url)), '..')
// pnpm 严格依赖下 esbuild 不在顶层，借 vite 的依赖解析拿到它
const requireFromVite = createRequire(realpathSync(join(root, 'node_modules/vite/package.json')))
const { build } = requireFromVite('esbuild')

async function loadLocale(name) {
  const out = join(root, `scripts/.tmp-${name}.mjs`)
  await build({
    entryPoints: [join(root, `src/i18n/locales/${name}.ts`)],
    bundle: true,
    format: 'esm',
    outfile: out,
    logLevel: 'silent'
  })
  const mod = await import(pathToFileURL(out).href)
  rmSync(out)
  return mod.default
}

function flatten(obj, prefix, into) {
  for (const [k, v] of Object.entries(obj)) {
    const key = prefix ? `${prefix}.${k}` : k
    if (v && typeof v === 'object') flatten(v, key, into)
    else into.add(key)
  }
  return into
}

function walk(dir, files) {
  for (const name of readdirSync(dir)) {
    const full = join(dir, name)
    const st = statSync(full)
    if (st.isDirectory()) {
      if (name === 'node_modules' || name === '__tests__') continue
      walk(full, files)
    } else if (/\.(vue|ts)$/.test(name) && !full.includes('i18n')) {
      files.push(full)
    }
  }
  return files
}

const zh = flatten(await loadLocale('zh'), '', new Set())
const en = flatten(await loadLocale('en'), '', new Set())

// 两类引用：t('key') 字面量调用 + 路由 meta 的 titleKey/descriptionKey
const keyRe = /(?:[^\w$]|^)\$?t\(\s*['"]([a-z][\w-]*(?:\.[\w-]+)+)['"]/g
const metaKeyRe = /(?:titleKey|descriptionKey):\s*['"]([a-z][\w-]*(?:\.[\w-]+)+)['"]/g
const used = new Map() // key -> [files]
for (const file of walk(join(root, 'src'), [])) {
  const text = readFileSync(file, 'utf8')
  for (const re of [keyRe, metaKeyRe]) {
    for (const m of text.matchAll(re)) {
      const key = m[1]
      if (!used.has(key)) used.set(key, [])
      used.get(key).push(file.replace(root, ''))
    }
  }
}

let missing = 0
for (const [key, files] of [...used].sort()) {
  // 形如 t('a.b.capSource_' + x) 的动态拼接前缀（约定以 _ 结尾），无法静态校验，跳过
  if (key.endsWith('_')) continue
  const inZh = zh.has(key)
  const inEn = en.has(key)
  if (!inZh || !inEn) {
    missing++
    console.log(`MISSING ${!inZh ? '[zh]' : ''}${!inEn ? '[en]' : ''} ${key}`)
    console.log(`   used: ${[...new Set(files)].slice(0, 3).join(', ')}`)
  }
}
console.log(missing === 0 ? 'OK: all literal t() keys exist in zh+en' : `TOTAL MISSING: ${missing}`)
