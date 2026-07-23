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
      <div class="detail-panel settings-section">
        <div class="detail-left">
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

        <div class="detail-right" v-if="allAccounts.length > 0">
          <div class="detail-divider"></div>
          <div class="accounts-header">
            <h3>推送账户</h3>
            <router-link to="/push" class="link-sm">管理账户</router-link>
          </div>
          <div class="accounts-list">
            <label v-for="acc in allAccounts" :key="acc.id" class="account-checkbox">
              <input
                type="checkbox"
                :value="acc.id"
                :checked="selectedAccountIDs.includes(acc.id)"
                @change="toggleAccount(acc.id)"
              />
              <span class="acc-name">{{ acc.name }}</span>
              <span class="acc-service-badge" :class="'badge-' + acc.service">{{ serviceLabel(acc.service) }}</span>
            </label>
          </div>
          <p class="hint" v-if="selectedAccountIDs.length === 0">未启用任何推送账户，发现更新时不会推送通知</p>
          <div class="accounts-actions">
            <button class="btn btn-primary btn-sm" :disabled="savingAccounts" @click="saveAccounts">
              {{ savingAccounts ? '保存中...' : '保存' }}
            </button>
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
        <div class="pagination" v-if="updatesTotal > updatesPageSize">
          <button class="btn btn-sm btn-ghost" :disabled="updatesPage <= 1 || updatesLoading" @click="changeUpdatesPage(updatesPage - 1)">上一页</button>
          <span>第 {{ updatesPage }} / {{ totalUpdatePages }} 页</span>
          <button class="btn btn-sm btn-ghost" :disabled="updatesPage >= totalUpdatePages || updatesLoading" @click="changeUpdatesPage(updatesPage + 1)">下一页</button>
        </div>
      </div>
    </template>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { fetchMonitor, fetchUpdates, startMonitor, stopMonitor, deleteMonitor, markAllNotified, markRead, fetchMonitorConfig, fetchAccounts, updateNotifyAccounts } from '../api/monitors'
import StatusBadge from '../components/StatusBadge.vue'
import UpdateTable from '../components/UpdateTable.vue'
import { useToastMessages } from '../composables/useToastMessages'

const route = useRoute()
const router = useRouter()
const { successMsg, pageErrorMsg, showSuccess, showError } = useToastMessages()

const monitor = ref(null)
const records = ref([])
const updatesPage = ref(1)
const updatesPageSize = 20
const updatesTotal = ref(0)
const totalUpdatePages = computed(() => Math.max(1, Math.ceil(updatesTotal.value / updatesPageSize)))
const loading = ref(true)
const updatesLoading = ref(false)
const error = ref(null)
const actionLoading = ref(false)
const markLoading = ref(false)
const showDeleteConfirm = ref(false)

// 推送账户选择
const allAccounts = ref([])
const selectedAccountIDs = ref([])
const accountDirty = ref(false)
const savingAccounts = ref(false)

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
      selectedAccountIDs.value = configRes.data.notify_account_ids || []
    }
  } catch (e) { error.value = e.response?.data?.message || e.message }
  finally { loading.value = false }
}

async function loadUpdates() {
  updatesLoading.value = true
  try {
    const res = await fetchUpdates(route.params.name, { page: updatesPage.value, size: updatesPageSize })
    if (res.code === 0 && res.data) {
      records.value = res.data.records || []
      updatesTotal.value = res.data.total || 0
      if (updatesPage.value > totalUpdatePages.value) {
        updatesPage.value = totalUpdatePages.value
        await loadUpdates()
      }
    }
  } catch { /* ignore */ }
  finally { updatesLoading.value = false }
}

async function changeUpdatesPage(page) {
  if (page < 1 || page > totalUpdatePages.value) return
  updatesPage.value = page
  await loadUpdates()
}

async function toggleRun() {
  actionLoading.value = true
  try {
    if (monitor.value.is_running) {
      await stopMonitor(route.params.name)
      showError('监控器已暂停')
    } else {
      await startMonitor(route.params.name)
      showSuccess('监控器已启动')
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

function toggleAccount(accountID) {
  const idx = selectedAccountIDs.value.indexOf(accountID)
  if (idx >= 0) {
    selectedAccountIDs.value.splice(idx, 1)
  } else {
    selectedAccountIDs.value.push(accountID)
  }
  accountDirty.value = true
}

async function saveAccounts() {
  savingAccounts.value = true
  try {
    await updateNotifyAccounts(route.params.name, selectedAccountIDs.value)
    showSuccess(selectedAccountIDs.value.length > 0 ? '推送账户已保存' : '已关闭所有推送账户')
    accountDirty.value = false
  } catch (e) {
    showError('保存失败: ' + (e.response?.data?.message || e.message))
  } finally {
    savingAccounts.value = false
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
.btn-pause:hover:not(:disabled) { background: var(--bg-hover); transform: scale(1.08); }

.header-actions { display: flex; gap: 0.5rem; align-items: center; }

/* 左右布局 */
.detail-panel { display: flex; gap: 0; }
.detail-left { flex: 1; min-width: 0; }
.detail-divider { width: 1px; background: var(--border-light); margin: 0 1.25rem; align-self: stretch; }
.detail-right { width: 240px; flex-shrink: 0; }

/* 推送账户列表 */
.accounts-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 0.75rem; }
.accounts-header h3 { font-size: 0.9375rem; font-weight: 700; color: var(--text); }
.link-sm { font-size: 0.75rem; color: var(--text-secondary); text-decoration: none; }
.link-sm:hover { color: var(--text); }
.accounts-list { display: flex; flex-direction: column; gap: 0.4rem; }
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
  font-size: 0.625rem; font-weight: 700; color: #000000;
  background: var(--green); padding: 0.1rem 0.4rem; border-radius: var(--radius-pill);
}
.badge-bark { background: #d32f2f; color: #fff; }
.badge-pushplus { background: #1976d2; color: #fff; }
.badge-serverchan { background: var(--green); color: #000; }
.accounts-actions { margin-top: 0.75rem; display: flex; justify-content: flex-end; }
.hint { font-size: 0.75rem; color: var(--text-muted); margin-top: 0.5rem; }
.pagination { display: flex; align-items: center; justify-content: center; gap: 0.75rem; margin-top: 1rem; font-size: 0.8125rem; color: var(--text-muted); }

@media (max-width: 768px) {
  .status-grid { grid-template-columns: 1fr; }
  .page-header { flex-direction: column; gap: 0.75rem; }
  .detail-panel { flex-direction: column; }
  .detail-divider { width: 100%; height: 1px; margin: 1rem 0; }
  .detail-right { width: 100%; }
}
</style>
