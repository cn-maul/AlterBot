<template>
  <div class="monitor-detail">
    <div class="page-header">
      <div>
        <router-link to="/" class="back-link">← 返回</router-link>
        <h1>{{ monitor ? monitor.name : '加载中...' }}</h1>
      </div>
      <div class="header-actions" v-if="monitor">
        <button class="btn btn-action" :class="monitor.is_running ? 'btn-pause' : 'btn-play'" @click="toggleRun" :disabled="actionLoading">
          <svg v-if="monitor.is_running" viewBox="0 0 24 24" fill="currentColor" width="18" height="18">
            <rect x="6" y="4" width="4" height="16" rx="1"/>
            <rect x="14" y="4" width="4" height="16" rx="1"/>
          </svg>
          <svg v-else viewBox="0 0 24 24" fill="currentColor" width="18" height="18">
            <path d="M8 5v14l11-7z"/>
          </svg>
          {{ actionLoading ? '处理中...' : (monitor.is_running ? '暂停' : '启动') }}
        </button>
        <router-link :to="`/edit/${encodeURIComponent(monitor.name)}`" class="btn btn-ghost">
          ✏️ 编辑
        </router-link>
        <button class="btn btn-danger" @click="confirmDelete">
          🗑 删除
        </button>
      </div>
    </div>

    <!-- 内联通知 -->
    <div class="toast toast-success" v-if="successMsg">{{ successMsg }}</div>
    <div class="toast toast-warning" v-if="pageErrorMsg">{{ pageErrorMsg }}</div>

    <!-- 加载态 -->
    <div class="loading" v-if="loading">
      <div class="spinner" />
      <p>加载监控器详情...</p>
    </div>

    <!-- 错误态 -->
    <div class="empty" v-else-if="error">
      <div class="empty-icon">❌</div>
      <p>{{ error }}</p>
      <button class="btn btn-primary btn-sm" style="margin-top: 1rem;" @click="loadData">重试</button>
    </div>

    <!-- 删除确认对话框 -->
    <div class="modal-overlay" v-if="showDeleteConfirm" @click.self="showDeleteConfirm = false">
      <div class="modal-container">
        <div class="modal-header">
          <h2>确认删除</h2>
          <button class="modal-close" @click="showDeleteConfirm = false">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
          </button>
        </div>
        <div class="modal-body">
          <p>确定要删除监控器「{{ monitor?.name }}」吗？</p>
          <p style="color: var(--text-muted); font-size: 0.85rem; margin-top: 0.5rem;">删除后无法恢复，相关更新记录也会被清除。</p>
        </div>
        <div class="modal-footer">
          <button class="btn btn-ghost" @click="showDeleteConfirm = false">取消</button>
          <button class="btn btn-danger" @click="handleDelete" :disabled="actionLoading">{{ actionLoading ? '删除中...' : '确认删除' }}</button>
        </div>
      </div>
    </div>

    <template v-else-if="monitor">
      <!-- 状态面板 -->
      <div class="status-panel settings-section">
        <div class="status-row">
          <StatusBadge :status="monitor.is_running ? 'running' : (monitor.last_error ? 'error' : 'stopped')" />
          <span class="interval-badge">间隔 {{ formatInterval(monitor.check_interval) }}</span>
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

      <!-- 更新历史 -->
      <div class="settings-section">
        <div class="section-header">
          <h2>更新历史</h2>
        </div>
        <UpdateTable :records="records" :loading="updatesLoading" />
      </div>
    </template>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { fetchMonitor, fetchUpdates, startMonitor, stopMonitor, deleteMonitor } from '../api/monitors'
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
const successMsg = ref('')
const pageErrorMsg = ref('')
const showDeleteConfirm = ref(false)

let msgTimer = null

function showSuccess(msg) {
  successMsg.value = msg
  clearTimeout(msgTimer)
  msgTimer = setTimeout(() => { successMsg.value = '' }, 3000)
}

function showError(msg) {
  pageErrorMsg.value = msg
  clearTimeout(msgTimer)
  msgTimer = setTimeout(() => { pageErrorMsg.value = '' }, 5000)
}

onMounted(loadData)

async function loadData() {
  loading.value = true
  error.value = null
  try {
    const res = await fetchMonitor(route.params.name)
    if (res.code === 0) {
      monitor.value = res.data
      loadUpdates()
    } else {
      error.value = res.message || '监控器不存在'
    }
  } catch (e) {
    error.value = e.response?.data?.message || e.message
  } finally {
    loading.value = false
  }
}

async function loadUpdates() {
  updatesLoading.value = true
  try {
    const res = await fetchUpdates(route.params.name)
    if (res.code === 0 && res.data) {
      records.value = res.data.records || []
    }
  } catch {
    // ignore
  } finally {
    updatesLoading.value = false
  }
}

async function toggleRun() {
  actionLoading.value = true
  try {
    if (monitor.value.is_running) {
      await stopMonitor(route.params.name)
      pageErrorMsg.value = '监控器已暂停'
      clearTimeout(msgTimer)
      msgTimer = setTimeout(() => { pageErrorMsg.value = '' }, 3000)
    } else {
      await startMonitor(route.params.name)
      successMsg.value = '监控器已启动'
      clearTimeout(msgTimer)
      msgTimer = setTimeout(() => { successMsg.value = '' }, 3000)
    }
    await loadData()
  } catch (e) {
    showError('操作失败: ' + (e.response?.data?.message || e.message))
  } finally {
    actionLoading.value = false
  }
}

function confirmDelete() {
  showDeleteConfirm.value = true
}

async function handleDelete() {
  actionLoading.value = true
  try {
    await deleteMonitor(route.params.name)
    router.push('/')
  } catch (e) {
    showError('删除失败: ' + (e.response?.data?.message || e.message))
    showDeleteConfirm.value = false
    actionLoading.value = false
  }
}

function formatTime(t) {
  if (!t) return '—'
  return new Date(t).toLocaleString('zh-CN')
}

function formatInterval(ns) {
  if (!ns) return '—'
  const s = Math.floor(ns / 1e9)
  if (s >= 3600) return `${Math.round(s / 3600)} 小时`
  if (s >= 60) return `${Math.round(s / 60)} 分钟`
  return `${s} 秒`
}

function formatDuration(ns) {
  if (!ns) return '—'
  const ms = Math.round(ns / 1e6)
  if (ms >= 1000) return `${(ms / 1000).toFixed(1)}s`
  return `${ms}ms`
}
</script>

<style scoped>
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 1.5rem;
}

.page-header h1 {
  font-size: 1.5rem;
  font-weight: 700;
  color: var(--text);
  margin-top: 0.5rem;
}

.back-link {
  color: var(--text-secondary);
  font-size: 0.9rem;
  text-decoration: none;
}

.back-link:hover {
  color: var(--primary);
}

.toast {
  padding: 0.6rem 1rem;
  border-radius: var(--radius);
  font-size: 0.85rem;
  margin-bottom: 1rem;
  animation: fadeIn 0.2s ease;
}

.toast-success {
  background: var(--success-bg);
  color: var(--success);
  border: 1px solid var(--success);
}

.toast-warning {
  background: var(--warning-bg);
  color: var(--warning);
  border: 1px solid var(--warning);
}

.toast-error {
  background: var(--error-bg);
  color: var(--error);
  border: 1px solid var(--error);
}

@keyframes fadeIn {
  from { opacity: 0; transform: translateY(-4px); }
  to { opacity: 1; transform: translateY(0); }
}

.status-row {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  margin-bottom: 1rem;
}

.interval-badge {
  font-size: 0.8rem;
  color: var(--text-muted);
}

.status-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 0.75rem;
}

.status-item {
  display: flex;
  flex-direction: column;
  gap: 0.15rem;
}

.status-label {
  font-size: 0.78rem;
  color: var(--text-muted);
  text-transform: uppercase;
  letter-spacing: 0.03em;
}

.status-value {
  font-size: 0.9rem;
  color: var(--text);
  word-break: break-all;
}

.error-text {
  color: var(--error);
}

.header-actions {
  display: flex;
  gap: 0.6rem;
  align-items: center;
}

.header-actions .btn {
  padding: 0.6rem 1.2rem;
  font-size: 0.9rem;
}

.btn-action {
  display: inline-flex;
  align-items: center;
  gap: 0.4rem;
  border: none;
  border-radius: var(--radius);
  cursor: pointer;
  transition: var(--transition);
  padding: 0.6rem 1.2rem;
  font-weight: 500;
  font-size: 0.9rem;
}

.btn-play {
  background: var(--success);
  color: white;
}

.btn-play:hover:not(:disabled) {
  background: #059669;
}

.btn-play svg,
.btn-pause svg {
  width: 18px;
  height: 18px;
  flex-shrink: 0;
}

.btn-pause {
  background: var(--warning);
  color: white;
}

.btn-pause:hover:not(:disabled) {
  background: #d97706;
}

@media (max-width: 768px) {
  .status-grid {
    grid-template-columns: 1fr;
  }
  .page-header {
    flex-direction: column;
    gap: 0.75rem;
  }
}
</style>