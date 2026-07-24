# API 文档

## 基本约定

- API 前缀：`/api`
- 监控器接口前缀：`/api/v1/monitors`
- 内容类型：`application/json`
- 成功响应：`{"code":0,"message":"success","data":...}`
- 失败响应：`{"code":<错误码>,"message":"<错误信息>","data":null}`

如果设置了 `ALTERBOT_AUTH_TOKEN`，所有 `/api` 请求都需要提供：

```http
Authorization: Bearer <token>
```

## 系统接口

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| `GET` | `/api/health` | 健康检查 |
| `GET` | `/api/stats` | 系统统计 |
| `GET` | `/api/groups` | 监控分组 |
| `GET` | `/api/settings/notifications` | 获取全局通知设置 |
| `PUT` | `/api/settings/notifications` | 更新全局通知设置 |

## 监控器接口

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| `GET` | `/api/v1/monitors/` | 获取监控器列表 |
| `POST` | `/api/v1/monitors/` | 创建监控器 |
| `GET` | `/api/v1/monitors/:name` | 获取运行状态 |
| `GET` | `/api/v1/monitors/:name/config` | 获取完整配置 |
| `PUT` | `/api/v1/monitors/:name` | 更新配置 |
| `DELETE` | `/api/v1/monitors/:name` | 删除监控器及关联状态 |
| `POST` | `/api/v1/monitors/:name/start` | 启动监控器 |
| `POST` | `/api/v1/monitors/:name/stop` | 停止监控器 |
| `POST` | `/api/v1/monitors/:name/check` | 立即检查 |
| `POST` | `/api/v1/monitors/:name/baseline` | 重置基线 |
| `GET` | `/api/v1/monitors/:name/updates` | 获取旧版新增记录 |
| `GET` | `/api/v1/monitors/:name/events` | 获取变化事件 |
| `GET` | `/api/v1/monitors/:name/snapshots` | 获取当前快照 |
| `PUT` | `/api/v1/monitors/:name/notify-accounts` | 更新通知账户 |
| `PUT` | `/api/v1/monitors/:name/mark-all-notified` | 标记全部已通知 |
| `POST` | `/api/v1/monitors/:name/mark-read` | 标记记录已读 |

## 配置辅助接口

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| `POST` | `/api/v1/monitors/validate` | 抓取并验证监控配置，不写入基线 |
| `POST` | `/api/v1/monitors/preview` | 扫描网页候选区域 |
| `POST` | `/api/v1/monitors/smart-create` | 根据扫描结果创建监控 |

## 通知账户

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| `GET` | `/api/settings/notification-accounts` | 获取通知账户 |
| `POST` | `/api/settings/notification-accounts` | 创建通知账户 |
| `PUT` | `/api/settings/notification-accounts/:id` | 更新通知账户 |
| `DELETE` | `/api/settings/notification-accounts/:id` | 删除通知账户 |
| `GET` | `/api/settings/notification-providers` | 获取通知服务元数据 |

## 扫描规则模板

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| `GET` | `/api/settings/scan-rules` | 获取模板列表 |
| `POST` | `/api/settings/scan-rules` | 创建模板 |
| `PUT` | `/api/settings/scan-rules/:id` | 更新模板 |
| `DELETE` | `/api/settings/scan-rules/:id` | 删除模板 |
| `POST` | `/api/settings/scan-rules/:id/test` | 测试模板 |

## 到价提醒策略示例

以下片段表示价格从目标价以上降到 `199.00` 或以下时产生事件：

```json
{
  "strategy_type": "field_transition",
  "strategy_config": {
    "type": "field_transition",
    "identity": { "source": "source_url" },
    "conditions": [
      {
        "field": "price",
        "value_type": "money",
        "operator": "at_or_below",
        "threshold": { "value": "199.00" }
      }
    ],
    "on_first_baseline": "silent"
  },
  "field_data_types": {
    "price": "money"
  }
}
```

创建和更新请求还需要包含名称、URL、选择器、字段和检查间隔等完整监控配置。推荐先调用 `/validate`，或者直接通过 Web 管理界面创建。
