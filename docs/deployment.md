# 部署指南

## Docker Compose

```bash
docker compose up -d --build
```

默认映射端口为 `8889:8889`，数据库保存在 `gentry-data` 卷中。

查看日志：

```bash
docker compose logs -f gentry
```

升级本地构建版本：

```bash
docker compose build --pull
docker compose up -d
```

## Docker

构建镜像：

```bash
docker build --build-arg GO_VERSION=1.26-alpine -t gentry:latest .
```

运行：

```bash
docker run -d \
  --name gentry \
  --restart unless-stopped \
  -p 8889:8889 \
  -v gentry-data:/app/data \
  -e TZ=Asia/Shanghai \
  gentry:latest
```

## 本地二进制

生产构建需要先生成 `frontend/dist`，因为 Go 会将该目录嵌入二进制：

```bash
make build
./gentry
```

本地二进制默认在当前工作目录创建 `gentry.db`。请在固定目录运行程序并定期备份数据库。

## 环境变量

| 变量 | 默认值 | 说明 |
| --- | --- | --- |
| `PORT` | 本地 `8080`，Docker `8889` | HTTP 服务端口 |
| `GIN_MODE` | Docker 中为 `release` | Gin 运行模式 |
| `TZ` | 系统默认时区 | 推荐设置为 `Asia/Shanghai` |
| `ALTERBOT_AUTH_TOKEN` | 空 | 可选 API Bearer Token；历史兼容命名 |
| `SCAN_RULES_FILE` | 空 | 可选的扫描规则文件路径 |

设置 `ALTERBOT_AUTH_TOKEN` 后，请求 `/api` 下的接口需要携带：

```text
Authorization: Bearer <token>
```

## 数据备份

核心状态存储在 SQLite 数据库中，包括监控定义、快照、事件和通知投递记录。

备份前建议暂停容器或程序，然后复制数据库文件。Docker 部署可将数据卷导出到宿主机备份系统。

## 健康检查

```bash
curl http://localhost:8889/api/health
```

如果配置了认证令牌，需要增加 `Authorization` 请求头。
