# 快速开始

## Docker Compose

项目根目录已经提供 `docker-compose.yml`：

```bash
docker compose up -d --build
docker compose logs -f
```

启动后访问 [http://localhost:8889](http://localhost:8889)。停止服务使用：

```bash
docker compose down
```

普通停止不会删除数据库。只有执行 `docker compose down -v` 才会同时删除 `gentry-data` 数据卷。

## 本地运行

环境要求：

- Go 1.26
- Node.js 22 或兼容版本
- pnpm

构建前端并编译后端：

```bash
make build
make run
```

Windows 也可以运行根目录的 `build-windows.bat`。

默认地址为 [http://localhost:8080](http://localhost:8080)。可通过 `PORT` 环境变量修改端口。

## 开发模式

后端：

```bash
make dev
```

另一个终端启动前端：

```bash
cd frontend
pnpm install
pnpm run dev
```

开发界面默认运行在 [http://localhost:5173](http://localhost:5173)。

## 创建网页新增监控

1. 打开“新增监控”。
2. 选择“网页新增监控”。
3. 填写网页 URL、容器选择器和列表项选择器。
4. 配置标题、链接、日期等提取字段。
5. 执行配置验证，确认能够稳定提取条目。
6. 选择通知账户并保存。

首次检查只记录当前条目，之后出现新的稳定条目标识时才产生通知。

## 创建价格监控

1. 选择“价格监控”。
2. 选择单商品详情页或商品列表页。
3. 配置价格字段，并确保它能被解析为金额。
4. 单商品页可使用页面 URL 作为商品身份；列表页必须配置稳定且唯一的身份字段，例如 SKU 或商品链接。
5. 选择“价格发生下降”或“降到目标价及以下”。
6. 验证价格样本后保存。

首次检查只建立价格基线，不会因为页面当前已经是低价而立即推送。
