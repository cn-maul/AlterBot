# Gentry — 网页变更监控系统

一个轻量级的网页变更监控系统，纯 Go 单二进制部署，内置 Vue 3 管理界面。支持 CSS 选择器提取网页内容、检测变化并通过多种渠道推送通知。

## 特性

- 🕵️ **CSS 选择器监控** — 使用 goquery 精确提取网页指定区域内容
- 🧠 **智能变更检测** — 基于标题+URL 组合的去重算法，避免重复通知
- 📡 **多种推送渠道** — PushPlus / Webhook / Server酱
- 🗄️ **数据持久化** — SQLite 存储配置和变更历史，重启不丢失
- 🌐 **Web 管理界面** — Vue 3 单页应用，零 UI 框架依赖
- 🐳 **Docker 一键部署** — 32MB Alpine 镜像，开箱即用
- ⚡ **运行时管理** — 通过 API/UI 动态增删改查监控器，无需重启

## 快速开始

### 方式一：Docker（推荐）

```bash
docker run -d --name gentry -p 8889:8889 -v gentry-data:/app/data ghcr.io/cn-maul/gentry:latest
```

打开 http://localhost:8889 即可使用。

### 方式二：直接运行

```bash
# 编译 + 构建前端
make build

# 运行
make run
```

### 方式三：开发模式

终端 1 — 后端：
```bash
make dev
```

终端 2 — 前端：
```bash
make dev-frontend
```

访问 http://localhost:5173 打开管理界面（API 代理到 :8080）。

## Docker 部署

### 构建镜像

```bash
docker build --build-arg GO_VERSION=1.26-alpine -t gentry:latest .
```

### 运行容器

```bash
docker run -d \
  --name gentry \
  -p 8889:8889 \
  -v gentry-data:/app/data \
  gentry:latest
```

### 环境变量

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `PORT` | `8889` | 监听端口 |
| `GIN_MODE` | `release` | Gin 运行模式 |
| `TZ` | `UTC` | 时区（如 `Asia/Shanghai`） |

### Docker Compose

```yaml
services:
  gentry:
    build:
      context: .
      dockerfile: Dockerfile
    image: gentry:latest
    container_name: gentry
    restart: unless-stopped
    ports:
      - "8889:8889"
    volumes:
      - gentry-data:/app/data
    environment:
      - GIN_MODE=release
      - TZ=Asia/Shanghai

volumes:
  gentry-data:
```

## 配置推送通知

通过 Web 管理界面（"推送管理"）配置通知渠道，支持：

### PushPlus

在 [pushplus.plus](https://www.pushplus.plus/) 注册获取 token，通过管理界面添加。

### Webhook 机器人

支持企业微信、飞书、钉钉等任意 Webhook URL。

### Server酱

通过 [sct.ftqq.com](https://sct.ftqq.com/) 获取 SendKey，通过管理界面添加。

## API 文档

### 基础信息

- **Base URL**: `/api/v1/monitors`
- **响应格式**: `{ "code": 0, "message": "success", "data": {...} }`

### 监控器管理

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/` | 获取所有监控器状态 |
| GET | `/:name` | 获取单个监控器详情 |
| POST | `/` | 新增监控器 |
| PUT | `/:name` | 更新监控器配置 |
| DELETE | `/:name` | 删除监控器 |
| POST | `/:name/start` | 启动监控器 |
| POST | `/:name/stop` | 停止监控器 |
| GET | `/:name/updates` | 获取更新历史 |

### 健康检查

```bash
curl http://localhost:8889/api/health
# {"code":0,"message":"success","data":{"database":true,"monitors":0,"status":"ok"}}
```

### 创建监控器示例

```bash
curl -X POST http://localhost:8889/api/v1/monitors/ \
  -H "Content-Type: application/json" \
  -d '{
    "name": "招录公告",
    "url": "https://example.com/zlgg/",
    "container": "div.news-list",
    "item": "div.news-item",
    "check_interval": 3600,
    "is_active": true,
    "fields": [
      { "name": "title", "selector": "a", "type": "text" },
      { "name": "url", "selector": "a", "type": "attr", "attr": "href" },
      { "name": "date", "selector": "span.date", "type": "text", "transform": "trim([])" }
    ]
  }'
```

## 架构

```
┌──────────────────────────────────────┐
│      Frontend (Vue 3 + Vite)        │
│   http://localhost:5173 (dev)        │
│   纯 CSS Variables 体系，零 UI 框架   │
│   Dashboard / Add / Detail / Edit    │
└──────────────┬───────────────────────┘
               │ HTTP/JSON
┌──────────────▼───────────────────────┐
│         Backend (Go + Gin)           │
│   http://localhost:8889              │
│   ┌─────────┐  ┌──────────────────┐ │
│   │ REST API│  │ Monitor Engine   │ │
│   │ CRUD    │  │ goroutine pool   │ │
│   │ Start/  │  │ ticker → check   │ │
│   │ Stop    │  │ → notify         │ │
│   └────┬────┘  └────────┬─────────┘ │
│        │                │            │
│   ┌────▼────────────────▼─────────┐  │
│   │      SQLite (GORM)            │  │
│   │  sites / updates / settings   │  │
│   └───────────────────────────────┘  │
└──────────────────────────────────────┘
```

## 项目结构

```
├── main.go                 # 入口
├── database/               # 数据库层 (GORM + SQLite)
│   ├── db.go               # 初始化 + 自动迁移
│   └── models.go           # Site / SiteField / UpdateRecord
├── fetcher/                # HTTP 抓取
│   ├── fetcher.go          # 核心抓取
│   ├── config.go           # 客户端配置
│   └── options.go          # 函数式选项
├── monitor/                # 监控核心
│   ├── monitor.go          # 主循环 + 变更检测 + 通知
│   ├── manager.go          # 全局注册表 + 状态管理
│   ├── extractor.go        # HTML 内容提取 (goquery)
│   └── scanner.go          # 智能扫描
├── notify/                 # 通知推送
│   ├── interface.go        # Notifier 接口
│   ├── registry.go         # 服务注册表
│   ├── manager.go          # 全局推送管理
│   ├── pushplus.go         # PushPlus 实现
│   ├── webhook.go          # Webhook 机器人实现
│   └── serverchan.go       # Server酱 实现
├── web/                    # Web API
│   ├── server.go           # Gin 服务器
│   ├── routes.go           # 路由 + CORS
│   ├── types.go            # 结构体定义
│   └── operations.go       # CRUD 操作
├── frontend/               # Vue 3 前端
│   ├── src/
│   │   ├── api/            # Axios API 调用
│   │   ├── views/          # Dashboard / AddMonitor / MonitorDetail
│   │   ├── components/     # MonitorCard / FieldEditor / StatusBadge
│   │   └── style.css       # CSS Variables 设计系统
│   └── vite.config.js
├── Dockerfile              # 多阶段构建
├── docker-compose.yml      # Docker Compose 编排
├── Makefile
└── README.md
```

## 开发

```bash
# 安装前端依赖
cd frontend && npm install

# 后端开发（需 Go 1.25+）
go run .

# 前端开发
cd frontend && npm run dev

# 全量构建
make build

# Docker 构建
docker build --build-arg GO_VERSION=1.26-alpine -t gentry:latest .
```

## 许可证

MIT