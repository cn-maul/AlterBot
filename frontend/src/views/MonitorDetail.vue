<template>
  <div class="monitor-detail">
    <div class="page-header">
      <div>
        <router-link to="/" class="back-btn pill-link">← 返回</router-link>
        <h1>{{ monitor ? monitor.name : '加载中...' }}</h1>
      </div>
      <div class="header-actions" v-if="monitor">
        <button class="circle-btn" :class="monitor.is_running ? 'btn-pause' : 'btn-play'" @click="toggleRun" :disabled="actionLoading" :title="monitor.is_running ? '暂停' : '启动'">
          <svg v-if="monitor.is_running" viewBox="0 0 24 24" fill="currentColor" width="18" height="18"><rect x="6" y="4" width="4" height="16" rx="1"/><rect x="14" y="4" width="4" height="16" rx="1"/></svg>
          <svg v-else viewBox="0 0 24 24" fill="currentColor" width="18" height="18"><path d="M8 5v14l11-7z"/></svg>
        </button>
        <router-link :to="`/edit/${encodeURIComponent(monitor.name)}`" class="btn btn-ghost btn-sm">编辑</router-link>
        <button class="btn btn-danger btn-sm" @click="confirmDelete">删除</button>
      </div>
    </div>

    <div class="toast toast-success" v-if="successMsg">{{ successMsg }}</div>
    <div class="toast toast-warning" v-if="pageErrorMsg">{{ pageErrorMsg }}</div>

    <div class="loading" v-if="loading">
      <div class="spinner" />
      <p>加载监控器详情...</p>
    </div>

    <div class="empty" v-else-if="error">
      <div class="empty-icon">❌</div>
      <p>{{ error }}</p>
      <button class="btn btn-primary btn-sm" style="margin-top: 1rem;" @click="loadData">重试</button>
    </div>

    <div class="modal-overlay" v-if="showDeleteConfirm" @click.self="showDeleteConfirm = false">
      <div class="modal-container">
        <div class="modal-header">
          <h2>确认删除</h2>
          <button class="modal-close" @click="showDeleteConfirm = false">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="18" height="18"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
          </button>
        </div>
        <div class="modal-body">
          <p>确定要删除监控器「{{ monitor?.name }}」吗？</p>
          <p style="margin-top: 0.5rem;">删除后无法恢复，相关更新记录也会被清除。</p>
        </div>
        <div class="modal-footer">
          <button class="btn btn-ghost" @click="showDeleteConfirm = false">取消</button>
          <button class="btn btn-danger" @click="handleDelete" :disabled="actionLoading">{{ actionLoading ? '删除中...' : '确认删除' }}</button>
        </div>
      </div>
    </div>

    <template v-else-if="monitor">
      <div class="status-panel settings-section">
        <div class="status-row">
          <StatusBadge :status="monitor.is_running ? 'running' : (monitor.last_error ? 'error' : 'stopped')" />
          <span class="interval-badge">{{ formatInterval(monitor.check_interval) }}</span>
        </div>
        <div class="status-grid">
          <div class="status-item">
            <span class="status-label">网址</span>
            <span class="status-value">{{ monitor.url }}</span>
          </div>
          <div class="status-item" v-if="monitor.group">
            <span class="status-label">分组</span>
            <span class="status-value">{{ monitor.group }}</span>
          </div>
          <div class="status-item">
            <span class="status-label">上次检查</span>
            <span class="status-value">{{ monitor.last_check ? formatTime(monitor.last_check) : '—' }}</span>
          </div>
          <div class="status-item">
            <span class="status-label">检查耗时</span>
            <span class="status-value">{{ monitor.last_duration ? formatDuration(monitor.last_duration) : '—' }}</span>
          </div>
          <div class="status-item">
            <span class="status-label">更新次数</span>
            <span class="status-value">{{ monitor.updates_count || 0 }}</span>
          </div>
          <div class="status-item" v-if="monitor.last_error">
            <span class="status-label">错误信息</span>
            <span class="status-value error-text">{{ monitor.last_error }}</span>
          </div>
          <div class="status-item">
            <span class="status-label">下次检查</span>
            <span class="status-value">{{ monitor.next_check ? formatTime(monitor.next_check) : '—' }}</span>
          </div>
        </div>
      </div>

      <div class="settings-section">
        <div class="section-header">
          <h2>更新历史</h2>
          <button class="btn btn-sm btn-ghost" :disabled="markLoading" @click="handleMarkAll" v-if="records.length > 0">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="14" height="14"><polyline points="9 11 12 14 22 4"/><path d="M21 12v7a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h11"/></svg>
            全部标为已推送
          </button>
        </div>
        <UpdateTable :records="records" :loading="updatesLoading" />
      </div>

      <div class="settings-section">
        <div class="section-header">
          <h2>推送账户</h2>
          <router-link to="/push" class="btn btn-sm btn-ghost">管理账户</router-link>
        </div>

        <div class="empty" v-if="allAccounts.length === 0" style="padding: 1.5rem 0;">
          <p style="font-size: 0.875rem; color: var(--text-muted);">暂无可用的推送账户，先前往<router-link to="/push" style="color: var(--green);">推送管理</router-link>创建账户</p>
        </div>

        <div class="accounts-list" v-else>
          <label v-for="acc in allAccounts" :key="acc.ID" class="account-checkbox">
            <input
              type="checkbox"
              :value="acc.ID"
              :checked="selectedAccountIDs.includes(acc.ID)"
              @change="toggleAccount(acc.ID)"
            />
            <span class="acc-name">{{ acc.Name }}</span>
            <span class="acc-service-badge">{{ serviceLabel(acc.Service) }}</span>
          </label>
        </div>
        <p class="hint" v-if="allAccounts.length > 0 && selectedAccountIDs.length === 0">未启用任何推送账户，发现更新时不会推送通知</p>
      </div>
    </template>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { fetchMonitor, fetchUpdates, startMonitor, stopMonitor, deleteMonitor, markAllNotified, markRead, fetchMonitorConfig, fetchAccounts, updateNotifyAccounts } from '../api/monitors'
import StatusBadge from '../components/StatusBadge.vue'
import UpdateTable from '../components/UpdateTable.vue'

const route = useRoute()
const router = useRouter()

const monitor = ref(null)
const records = ref([])
const loading = ref(true)
const updatesLoading = ref(false)
const error = ref(null)
const actionLoading = ref(false)
const markLoading = ref(false)
const successMsg = ref('')
const pageErrorMsg = ref('')
const showDeleteConfirm = ref(false)

// 推送账户选择
const allAccounts = ref([])
const selectedAccountIDs = ref([])
const togglingAccount = ref(false)

let msgTimer = null

function showSuccess(msg) { successMsg.value = msg; clearTimeout(msgTimer); msgTimer = setTimeout(() => { successMsg.value = '' }, 3000) }
function showError(msg) { pageErrorMsg.value = msg; clearTimeout(msgTimer); msgTimer = setTimeout(() => { pageErrorMsg.value = '' }, 5000) }

onMounted(loadData)

async function loadData() {
  loading.value = true; error.value = null
  try {
    const [res, configRes, acctsRes] = await Promise.all([
      fetchMonitor(route.params.name),
      fetchMonitorConfig(route.params.name).catch(() => null),
      fetchAccounts().catch(() => ({ data: [] })),
    ])
    if (res.code === 0) {
      monitor.value = res.data
      loadUpdates()
      markRead(route.params.name).catch(() => {})
    } else { error.value = res.message || '监控器不存在' }

    // 加载推送账户列表
    allAccounts.value = (acctsRes.code === 0 ? acctsRes.data : []) || []

    // 解析当前监控器的启用账户 ID
    if (configRes && configRes.code === 0 && configRes.data) {
      const d = configRes.data
      if (d.NotifyAccountIDs) {
        try { selectedAccountIDs.value = JSON.parse(d.NotifyAccountIDs) } catch { selectedAccountIDs.value = [] }
      } else { selectedAccountIDs.value = [] }
    }
  } catch (e) { error.value = e.response?.data?.message || e.message }
  finally { loading.value = false }
}

async function loadUpdates() {
  updatesLoading.value = true
  try {
    const res = await fetchUpdates(route.params.name)
    if (res.code === 0 && res.data) { records.value = res.data.records || [] }
  } catch { /* ignore */ }
  finally { updatesLoading.value = false }
}

async function toggleRun() {
  actionLoading.value = true
  try {
    if (monitor.value.is_running) {
      await stopMonitor(route.params.name)
      pageErrorMsg.value = '监控器已暂停'
      clearTimeout(msgTimer); msgTimer = setTimeout(() => { pageErrorMsg.value = '' }, 3000)
    } else {
      await startMonitor(route.params.name)
      successMsg.value = '监控器已启动'
      clearTimeout(msgTimer); msgTimer = setTimeout(() => { successMsg.value = '' }, 3000)
    }
    await loadData()
  } catch (e) { showError('操作失败: ' + (e.response?.data?.message || e.message)) }
  finally { actionLoading.value = false }
}

function confirmDelete() { showDeleteConfirm.value = true }

async function handleDelete() {
  actionLoading.value = true
  try {
    await deleteMonitor(route.params.name)
    router.push('/')
  } catch (e) {
    showError('删除失败: ' + (e.response?.data?.message || e.message))
    showDeleteConfirm.value = false; actionLoading.value = false
  }
}

async function handleMarkAll() {
  markLoading.value = true
  try {
    const res = await markAllNotified(route.params.name)
    const n = res.data?.updated || 0
    showSuccess(`已将 ${n} 条记录标为已推送`)
    await loadUpdates()
  } catch (e) { showError('操作失败: ' + (e.message)) }
  finally { markLoading.value = false }
}

async function toggleAccount(accountID) {
  if (togglingAccount.value) return
  togglingAccount.value = true
  const idx = selectedAccountIDs.value.indexOf(accountID)
  if (idx >= 0) {
    selectedAccountIDs.value.splice(idx, 1)
  } else {
    selectedAccountIDs.value.push(accountID)
  }
  try {
    await updateNotifyAccounts(route.params.name, JSON.stringify(selectedAccountIDs.value))
    showSuccess(selectedAccountIDs.value.length > 0 ? '推送账户已更新' : '已关闭所有推送账户')
  } catch (e) {
    showError('保存失败: ' + (e.response?.data?.message || e.message))
    // 还原
    if (idx >= 0) selectedAccountIDs.value.push(accountID)
    else selectedAccountIDs.value.splice(selectedAccountIDs.value.indexOf(accountID), 1)
  } finally {
    togglingAccount.value = false
  }
}

function formatTime(t) { if (!t) return '—'; return new Date(t).toLocaleString('zh-CN') }
function formatInterval(ns) { if (!ns) return '—'; const s = Math.floor(ns / 1e9); if (s >= 3600) return `${Math.round(s / 3600)} 小时`; if (s >= 60) return `${Math.round(s / 60)} 分钟`; return `${s} 秒` }
function formatDuration(ns) { if (!ns) return '—'; const ms = Math.round(ns / 1e6); if (ms >= 1000) return `${(ms / 1000).toFixed(1)}s`; return `${ms}ms` }
function serviceLabel(s) {
  if (s === 'pushplus') return 'PushPlus'
  if (s === 'webhook') return 'Webhook'
  if (s === 'serverchan') return 'Server酱'
  return s
}
</script>

<style scoped>
.status-row { display: flex; align-items: center; gap: 0.75rem; margin-bottom: 1rem; }
.interval-badge { font-size: 0.75rem; color: var(--text-muted); background: var(--bg-elevated); padding: 0.15rem 0.5rem; border-radius: var(--radius-pill); }
.status-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 0.75rem; }
.status-item { display: flex; flex-direction: column; gap: 0.1rem; }
.status-label { font-size: 0.6875rem; font-weight: 700; color: var(--text-muted); text-transform: uppercase; letter-spacing: 0.8px; }
.status-value { font-size: 0.875rem; color: var(--text); word-break: break-all; }
.error-text { color: var(--error); }

/* 统一返回按钮 */
.back-btn {
  display: inline-flex; align-items: center; gap: 0.3rem;
  padding: 0.35rem 0.85rem; border-radius: var(--radius-pill);
  font-size: 0.8125rem; font-weight: 700; color: var(--text-secondary);
  background: var(--bg-elevated); text-decoration: none;
  transition: var(--transition);
}
.back-btn:hover { background: var(--bg-hover); color: var(--text); }

/* Page header */
.page-header { display: flex; justify-content: space-between; align-items: flex-start; margin-bottom: 1.5rem; }
.page-header h1 { font-size: 1.5rem; font-weight: 700; color: var(--text); margin-top: 0.5rem; }

/* Circular button */
.circle-btn {
  width: 44px; height: 44px; border: none; border-radius: var(--radius-circle);
  cursor: pointer; transition: var(--transition);
  display: flex; align-items: center; justify-content: center; padding: 0;
}
.circle-btn:disabled { opacity: 0.4; cursor: not-allowed; }
.btn-play { background: var(--green); color: #000000; }
.btn-play:hover:not(:disabled) { background: var(--green-hover); transform: scale(1.08); }
.btn-pause { background: var(--bg-elevated); color: var(--text); }
.btn-pause:hover:not(:disabled) { background: #333; transform: scale(1.08); }

.header-actions { display: flex; gap: 0.5rem; align-items: center; }

@media (max-width: 768px) {
  .status-grid { grid-template-columns: 1fr; }
  .page-header { flex-direction: column; gap: 0.75rem; }
}

/* 推送账户列表 */
.accounts-list {
  display: flex; flex-direction: column; gap: 0.4rem;
}
.account-checkbox {
  display: flex; align-items: center; gap: 0.5rem;
  padding: 0.5rem 0.75rem;
  border-radius: var(--radius-lg);
  background: var(--bg-card);
  cursor: pointer;
  transition: var(--transition);
}
.account-checkbox:hover { background: var(--bg-hover); }
.account-checkbox input { accent-color: var(--green); }
.acc-name { font-size: 0.875rem; font-weight: 700; color: var(--text); }
.acc-service-badge {
  font-size: 0.625rem; font-weight: 700; color: #000;
  background: var(--green); padding: 0.1rem 0.4rem; border-radius: var(--radius-pill);
}
.hint { font-size: 0.75rem; color: var(--text-muted); margin-top: 0.5rem; }
</style>