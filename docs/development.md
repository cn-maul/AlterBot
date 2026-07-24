# 开发指南

## 环境

- Go 1.26
- Node.js 22 或兼容版本
- pnpm
- SQLite 由 Go 驱动内置，不需要单独安装数据库服务

## 启动开发环境

后端：

```bash
make dev
```

前端：

```bash
cd frontend
pnpm install
pnpm run dev
```

后端默认监听 `8080`，Vite 开发服务器默认监听 `5173`。

## 构建

```bash
make build
```

构建顺序是：安装前端依赖、生成 `frontend/dist`、编译 Go 二进制并嵌入前端资源。

只构建前端：

```bash
cd frontend
pnpm run build
```

## 测试与静态检查

```bash
go test ./...
go vet ./...
cd frontend
pnpm test
pnpm run build
```

修改监控规则、快照或通知投递逻辑时，应同时覆盖首次基线、幂等去重、异常价格、币种边界、通知终态和配置编辑回填。

## 核心链路

```text
网页抓取
  → CSS 字段提取
  → 类型规范化
  → 健康检查
  → 加载上次快照
  → Detector 比较
  → 事务保存快照、事件和投递任务
  → Delivery worker 异步通知
```

## 目录结构

```text
database/   SQLite 模型、迁移和仓储
fetcher/    HTTP 抓取
monitor/    提取、规范化、Detector、快照、事件和投递服务
notify/     PushPlus、Webhook、Server酱等通知实现
web/        Gin API、配置验证和前端静态资源服务
frontend/   Vue 3 管理界面
docs/       使用、部署、API、开发和设计文档
main.go     应用入口与服务生命周期
```

## 监控策略

- `presence`：检测稳定身份的新条目。
- `field_transition`：比较同一条目的结构化字段，目前用于价格下降和到价提醒。

策略定义、字段类型和身份配置由前后端共同校验。涉及检测语义的配置变化会递增配置版本并重建基线。

## 数据模型

核心持久化对象包括：

- `Site`、`SiteField`：监控定义；
- `MonitorSnapshot`：每个稳定条目的当前比较基线；
- `MonitorEvent`：检测到的变化事件；
- `NotificationDelivery`：每个事件对应的异步通知任务；
- `UpdateRecord`：旧版新增监控兼容记录。

详细的引擎设计和历史审查记录位于[设计档案](design/)。
