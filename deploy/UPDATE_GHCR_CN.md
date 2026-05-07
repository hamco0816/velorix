# Velorix 生产站点镜像更新指南

本文档只适用于当前 Velorix 站点的生产部署。

当前生产镜像必须使用：

```bash
ghcr.io/hamco0816/velorix:latest
```

不要只使用 `docker-compose.local.yml` 更新生产站点。该文件默认镜像是：

```bash
weishaw/sub2api:latest
```

如果只执行 `docker compose -f docker-compose.local.yml pull sub2api`，会拉到默认镜像，线上可能变成旧版本或上游版本，导致新功能看不到。

## 第一次修正镜像来源

如果服务器还没有 `/opt/sub2api/deploy/docker-compose.override.yml`，先执行：

```bash
cd /opt/sub2api/deploy
printf '%s\n' 'services:' '  sub2api:' '    image: ghcr.io/hamco0816/velorix:latest' > docker-compose.override.yml
cat docker-compose.override.yml
```

逐行说明：

| 命令 | 含义 |
|------|------|
| `cd /opt/sub2api/deploy` | 进入服务器上的部署目录。所有 compose 命令都应该在这里执行。 |
| `printf '%s\n' ... > docker-compose.override.yml` | 创建覆盖配置文件，把 `sub2api` 服务的镜像改成 GHCR 的 Velorix 镜像。 |
| `cat docker-compose.override.yml` | 检查文件内容，确认镜像不是 `weishaw/sub2api:latest`。 |

正确内容应为：

```yaml
services:
  sub2api:
    image: ghcr.io/hamco0816/velorix:latest
```

## 日常更新命令

以后每次 GitHub 镜像构建完成后，只执行这一组命令：

```bash
cd /opt/sub2api/deploy
docker compose -f docker-compose.local.yml -f docker-compose.override.yml pull sub2api
docker compose -f docker-compose.local.yml -f docker-compose.override.yml up -d sub2api
docker compose -f docker-compose.local.yml -f docker-compose.override.yml ps
docker inspect sub2api --format '{{.Config.Image}}'
curl -fsS http://127.0.0.1:8080/health && echo
curl -fsS https://api.velorix.chat/health && echo
```

逐行说明：

| 命令 | 含义 |
|------|------|
| `cd /opt/sub2api/deploy` | 进入部署目录，确保 compose 能读取 `.env`、`docker-compose.local.yml` 和 `docker-compose.override.yml`。 |
| `docker compose -f docker-compose.local.yml -f docker-compose.override.yml pull sub2api` | 拉取 `sub2api` 服务的新镜像。`-f` 后面的文件会按顺序合并，后面的 `docker-compose.override.yml` 会覆盖前面的默认镜像。 |
| `docker compose -f docker-compose.local.yml -f docker-compose.override.yml up -d sub2api` | 只重建并后台启动应用容器 `sub2api`。不会重启 Postgres、Redis，也不会删除数据卷。 |
| `docker compose -f docker-compose.local.yml -f docker-compose.override.yml ps` | 查看容器是否正常运行，确认 `sub2api` 状态是 running 或 healthy。 |
| `docker inspect sub2api --format '{{.Config.Image}}'` | 查看当前容器实际使用的镜像。必须显示 `ghcr.io/hamco0816/velorix:latest`。 |
| `curl -fsS http://127.0.0.1:8080/health && echo` | 在服务器本机检查后端健康状态。返回 `{"status":"ok"}` 表示本机服务正常。 |
| `curl -fsS https://api.velorix.chat/health && echo` | 从公网域名检查反向代理和 HTTPS 是否正常。返回 `{"status":"ok"}` 表示公网访问正常。 |

## 判断是否拉错镜像

执行：

```bash
cd /opt/sub2api/deploy
docker compose -f docker-compose.local.yml config | grep -A3 "sub2api:"
docker inspect sub2api --format '{{.Config.Image}}'
```

如果看到：

```bash
image: weishaw/sub2api:latest
```

说明没有使用 override，镜像来源是错的。

正确检查方式是：

```bash
cd /opt/sub2api/deploy
docker compose -f docker-compose.local.yml -f docker-compose.override.yml config | grep -A5 "sub2api:"
docker inspect sub2api --format '{{.Config.Image}}'
```

正确结果必须包含：

```bash
image: ghcr.io/hamco0816/velorix:latest
```

## 常见问题

### 为什么不能只用 docker-compose.local.yml？

因为 `deploy/docker-compose.local.yml` 是通用部署文件，里面写的是上游默认镜像：

```yaml
image: weishaw/sub2api:latest
```

当前生产站点使用的是自己的 GHCR 镜像，所以必须通过 `docker-compose.override.yml` 覆盖镜像地址。

### 为什么不执行 docker compose down？

日常更新只需要替换应用容器，不需要停止数据库和 Redis。`up -d sub2api` 会只处理 `sub2api` 服务，更适合线上更新，风险更低。

### pull 提示没有权限怎么办？

先登录 GHCR：

```bash
docker login ghcr.io -u hamco0816
```

然后重新执行日常更新命令。

