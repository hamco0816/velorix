// 把 ../docs 下的对外文档同步到 public/docs，让前端能在 /docs/xxx 路径下访问。
// 唯一权威源在 ../docs/，public/docs/ 只是构建产物，但因为 dev 模式也读 public，所以也提交进 git 保证开发体验。
import { promises as fs } from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const __dirname = path.dirname(fileURLToPath(import.meta.url))
const frontendRoot = path.resolve(__dirname, '..')
const projectRoot = path.resolve(frontendRoot, '..')
const sourceDir = path.join(projectRoot, 'docs')
const targetDir = path.join(frontendRoot, 'public', 'docs')

// 需要同步到前端的文档清单：[源文件名, 目标文件名]
const FILES = [
  ['DOWNSTREAM_API_INTEGRATION.md', 'downstream-api-integration.md'],
]

async function main() {
  await fs.mkdir(targetDir, { recursive: true })
  for (const [src, dest] of FILES) {
    const srcPath = path.join(sourceDir, src)
    const destPath = path.join(targetDir, dest)
    const content = await fs.readFile(srcPath, 'utf8')
    await fs.writeFile(destPath, content, 'utf8')
    console.log(`[sync-public-docs] ${src} -> public/docs/${dest}`)
  }
}

main().catch((err) => {
  console.error('[sync-public-docs] failed:', err)
  process.exitCode = 1
})
