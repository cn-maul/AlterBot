import axios from 'axios'

const client = axios.create({
  baseURL: '/api/v1',
  timeout: 10000,
})

const rootClient = axios.create({
  baseURL: '/api',
  timeout: 10000,
})

// 获取所有监控器
export function fetchMonitors() {
  return client.get('/monitors/').then(r => r.data)
}

// 获取所有分组
export function fetchGroups() {
  return rootClient.get('/groups').then(r => r.data)
}

// 获取单个监控器
export function fetchMonitor(name) {
  return client.get(`/monitors/${encodeURIComponent(name)}`).then(r => r.data)
}

// 新增监控器
export function createMonitor(config) {
  return client.post('/monitors/', config).then(r => r.data)
}

// 更新监控器
export function updateMonitor(name, config) {
  return client.put(`/monitors/${encodeURIComponent(name)}`, config).then(r => r.data)
}

// 删除监控器
export function deleteMonitor(name) {
  return client.delete(`/monitors/${encodeURIComponent(name)}`).then(r => r.data)
}

// 启动监控器
export function startMonitor(name) {
  return client.post(`/monitors/${encodeURIComponent(name)}/start`).then(r => r.data)
}

// 停止监控器
export function stopMonitor(name) {
  return client.post(`/monitors/${encodeURIComponent(name)}/stop`).then(r => r.data)
}

// 获取更新历史
export function fetchUpdates(name) {
  return client.get(`/monitors/${encodeURIComponent(name)}/updates`).then(r => r.data)
}

// 获取监控器完整配置（用于编辑模式）
export function fetchMonitorConfig(name) {
  return client.get(`/monitors/${encodeURIComponent(name)}/config`).then(r => r.data)
}

// 健康检查
export function healthCheck() {
  return rootClient.get('/health').then(r => r.data)
}

// 获取通知设置
export function fetchNotificationSettings() {
  return rootClient.get('/settings/notifications').then(r => r.data)
}

// 更新通知设置
export function updateNotificationSettings(settings) {
  return rootClient.put('/settings/notifications', settings).then(r => r.data)
}