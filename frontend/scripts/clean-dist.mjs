// 在 vite build 前清空后端嵌入的 dist 目录，带 Windows 友好的 retry。
//
// 背景：vite 默认 emptyOutDir 会在构建开始时 unlink 旧产物；但 Windows 上若
// dist/ 文件被进程（dev 服务器、IDE 索引、防病毒扫描）占用，unlink 会抛
// "EPERM: operation not permitted, unlink ..."，整个 build 直接失败。
// 跨平台脚本提前清空，可以让单次失败重试避开瞬时锁，并把后续 vite emptyOutDir
// 变成 noop，降低构建被本地状态干扰的概率。
//
// 注意：只能本仓库内删 internal/web/dist；脚本里写死目标路径以防误删。

import { fileURLToPath } from 'node:url'
import { dirname, resolve } from 'node:path'
import { rm, mkdir } from 'node:fs/promises'

// 顶层兜底：任何未处理的 promise rejection 都打 warn 后 exit 0，
// 防止脚本意外让 npm/pnpm prebuild 钩子失败（Docker CI 上观察到 exit code 1）。
process.on('unhandledRejection', (err) => {
  console.warn('[clean-dist] unhandled rejection (non-fatal):', err && (err.message || err))
  process.exit(0)
})
process.on('uncaughtException', (err) => {
  console.warn('[clean-dist] uncaught exception (non-fatal):', err && (err.message || err))
  process.exit(0)
})

const __dirname = dirname(fileURLToPath(import.meta.url))
const target = resolve(__dirname, '../../backend/internal/web/dist')

async function sleep(ms) {
  return new Promise((r) => setTimeout(r, ms))
}

async function tryClean() {
  // maxRetries=5、间隔指数退避，覆盖防病毒/索引服务的短时锁
  const maxRetries = 5
  for (let i = 0; i < maxRetries; i++) {
    try {
      await rm(target, { recursive: true, force: true })
      return true
    } catch (err) {
      if (err && err.code === 'ENOENT') return true
      const wait = 200 * Math.pow(2, i)
      console.warn(`[clean-dist] attempt ${i + 1}/${maxRetries} failed: ${err.code || err.message}; retrying in ${wait}ms`)
      await sleep(wait)
    }
  }
  return false
}

async function main() {
  // 1) 清理旧产物
  try {
    const ok = await tryClean()
    if (!ok) {
      console.warn(
        '[clean-dist] could not clean dist after retries — likely another process (dev server / IDE indexers / antivirus) holds a file in ' +
          target +
          '. Falling back to vite emptyOutDir.'
      )
    }
  } catch (err) {
    console.warn('[clean-dist] clean step error (non-fatal):', err && (err.message || err))
  }

  // 2) 预创建空目录（Windows 上抗 EPERM mkdir / Docker linux 容器抗父目录不存在）
  let mkdirOk = false
  for (let i = 0; i < 5; i++) {
    try {
      await mkdir(target, { recursive: true })
      mkdirOk = true
      break
    } catch (err) {
      const wait = 200 * Math.pow(2, i)
      console.warn(`[clean-dist] mkdir attempt ${i + 1}/5 failed: ${err.code || err.message}; retrying in ${wait}ms`)
      await sleep(wait)
    }
  }
  if (!mkdirOk) {
    console.warn(
      '[clean-dist] could not pre-create ' +
        target +
        ' — vite will try again itself.'
    )
  }
}

main()
  .catch((err) => {
    // 任何 main 内冒泡上来的错误都不让脚本 exit 1（pnpm prebuild 失败会阻塞整个 build）
    console.warn('[clean-dist] main error (non-fatal):', err && (err.message || err))
  })
  .finally(() => {
    // 显式 exit 0，避免 Node 因 stdio 缓冲 / event loop 残留导致非零退出码
    process.exit(0)
  })
