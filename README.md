# AlterBot 网页更新通知

## 1、使用
在你的项目根目录创建目录`config`用于存放监视器的配置文件，创建配置文件`default.json`，编译时需要设置`GOEXPERIMENT=jsonv2,greenteagc`
```json
{
  "notification": {
    "service": "pushplus",
    "config": {
      "token": "yourtoken",
      "channel": "mail"
    }
  },
  "sites": [
    {
      "name": "招录公告",
      "url": "https://xxx.cn//zlgg/",
      "storage": "storage/xxx.json",
      "selectors": {
        "container": "div.hap_infoBox",
        "item": "div.hap_infoOne",
        "fields": [
          {
            "name": "title",
            "selector": "a",
            "type": "text"
          },
          {
            "name": "url",
            "selector": "a",
            "attr": "href",
            "type": "attr"
          },
          {
            "name": "date",
            "selector": "span.hap_infoDate",
            "type": "text",
            "transform": "trim([], '[]')"
          }
        ]
      },
      "check_interval": 30
    }
  ]
}
```

首先加载配置文件
```go
cfg, err := config.LoadConfig("config/default.json")
if err != nil {
log.Fatalf("加载配置失败: %v", err)
}
```
初始化推送服务
```go
if cfg.Notification != nil {
		if err := notify.InitGlobalNotifier(
			cfg.Notification.Service,
			cfg.Notification.Config,
		); err != nil {
			log.Fatal("推送服务初始化失败:", err)
		}
	}
```
启动监控goroutine
```go
var wg sync.WaitGroup
	stopCh := make(chan struct{})
	
	for _, site := range cfg.Sites {
		wg.Go(func() { // Go 1.25新语法
			monitor.Start(&site, stopCh)
		})
	}
```
等待中断信号,关闭所有监控器
```go
sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	
	close(stopCh)
	wg.Wait()
	log.Println("所有监控器已停止")
```
## 2、API

以下是完整的 API 文档，你可以直接添加到你的 `README.md` 文件中：

# API 文档

## 基础信息

- **Base URL**: `/api/v1/monitors`
- **认证**: 无（如需认证可后续添加）
- **响应格式**: JSON

```json
{
  "code": 0,
  "message": "success",
  "data": {} // 实际数据
}
```

## 监控器管理 API

### 1. 获取所有监控器状态

**Endpoint**: `GET /`

**描述**: 获取当前所有监控器的状态信息

**请求示例**:
```bash
curl -X GET http://localhost:8080/api/v1/monitors/
```

**成功响应**:
```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "name": "示例监控",
      "url": "https://example.com",
      "is_running": true,
      "last_check": "2023-10-01T12:00:00Z",
      "last_duration": 1500000000,
      "last_error": "",
      "last_update": "2023-10-01T11:58:00Z",
      "updates_count": 3,
      "next_check": "2023-10-01T12:05:00Z",
      "check_interval": 300000000000
    }
  ]
}
```

### 2. 获取单个监控器状态

**Endpoint**: `GET /:name`

**参数**:
- `name` (URL参数): 监控器名称

**请求示例**:
```bash
curl -X GET http://localhost:8080/api/v1/monitors/示例监控
```

**成功响应**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "name": "示例监控",
    "url": "https://example.com",
    "is_running": true,
    "last_check": "2023-10-01T12:00:00Z",
    "last_duration": 1500000000,
    "last_error": "",
    "last_update": "2023-10-01T11:58:00Z",
    "updates_count": 3,
    "next_check": "2023-10-01T12:05:00Z",
    "check_interval": 300000000000
  }
}
```

**监控器不存在响应**:
```json
{
  "code": 404,
  "message": "monitor not found"
}
```

### 3. 添加监控器

**Endpoint**: `POST /`

**请求体**:
```json
{
  "name": "新监控",
  "url": "https://new-site.com",
  "storage": "data/new-site.json",
  "interval": "5m",
  "selectors": {
    "container": ".news-list",
    "item": ".news-item",
    "fields": [
      {
        "name": "title",
        "selector": "h3",
        "type": "text"
      },
      {
        "name": "url",
        "selector": "a",
        "type": "attr",
        "attr": "href"
      }
    ]
  }
}
```

**请求示例**:
```bash
curl -X POST http://localhost:8080/api/v1/monitors/ \
  -H "Content-Type: application/json" \
  -d '{
    "name": "新监控",
    "url": "https://new-site.com",
    "storage": "data/new-site.json",
    "interval": "5m",
    "selectors": {
      "container": ".news-list",
      "item": ".news-item",
      "fields": [
        {
          "name": "title",
          "selector": "h3",
          "type": "text"
        },
        {
          "name": "url",
          "selector": "a",
          "type": "attr",
          "attr": "href"
        }
      ]
    }
  }'
```

**成功响应**:
```json
{
  "code": 0,
  "message": "monitor created",
  "data": "新监控"
}
```

**错误响应**:
- 400: 请求体无效
- 409: 监控器已存在

### 4. 删除监控器

**Endpoint**: `DELETE /:name`

**参数**:
- `name` (URL参数): 监控器名称

**请求示例**:
```bash
curl -X DELETE http://localhost:8080/api/v1/monitors/新监控
```

**成功响应**:
```json
{
  "code": 0,
  "message": "monitor removed"
}
```

**监控器不存在响应**:
```json
{
  "code": 404,
  "message": "monitor not found"
}
```

### 5. 启动监控器

**Endpoint**: `POST /:name/start`

**参数**:
- `name` (URL参数): 监控器名称

**请求示例**:
```bash
curl -X POST http://localhost:8080/api/v1/monitors/示例监控/start
```

**成功响应**:
```json
{
  "code": 0,
  "message": "monitor started"
}
```

**错误响应**:
- 404: 监控器不存在
- 200: 监控器已在运行 (code=0)

### 6. 停止监控器

**Endpoint**: `POST /:name/stop`

**参数**:
- `name` (URL参数): 监控器名称

**请求示例**:
```bash
curl -X POST http://localhost:8080/api/v1/monitors/示例监控/stop
```

**成功响应**:
```json
{
  "code": 0,
  "message": "monitor stopped"
}
```

**错误响应**:
- 404: 监控器不存在
- 200: 监控器已停止 (code=0)

## 状态码说明

| 状态码 | 说明 |
|--------|------|
| 0      | 成功 |
| 400    | 请求参数错误 |
| 404    | 资源不存在 |
| 409    | 资源已存在 |
| 500    | 服务器内部错误 |

## 示例配置

完整的监控器配置参考：
```json
{
  "name": "示例监控",
  "url": "https://example.com",
  "storage": "data/example.json",
  "interval": "5m",
  "selectors": {
    "container": ".news-list",
    "item": ".news-item",
    "fields": [
      {
        "name": "title",
        "selector": "h3",
        "type": "text"
      },
      {
        "name": "url",
        "selector": "a",
        "type": "attr",
        "attr": "href"
      }
    ]
  }
}
```




