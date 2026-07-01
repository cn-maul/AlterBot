<template>
  <div class="dashboard">
    <div class="page-header">
      <div>
        <h1>监控器</h1>
        <p class="page-desc">管理和监控网页内容变更</p>
      </div>
      <div class="header-actions">
        <router-link to="/add" class="btn btn-primary">＋ 新增监控器</router-link>
      </div>
    </div>

    <!-- 内联通知 -->
    <div class="toast toast-success" v-if="successMsg">{{ successMsg }}</div>
    <div class="toast toast-warning" v-if="pageErrorMsg">{{ pageErrorMsg }}</div>

    <!-- 加载态 -->
    <div class="loading" v-if="loading">
      <div class="spinner" />
      <p>加载中...</p>
    </div>

    <!-- 错误态 -->
    <div class="empty" v-else-if="error">
      <div class="empty-icon">❌</div>
      <p>加载失败</p>
      <p style="color: var(--text-muted); font-size: 0.85rem; margin-top: 0.25rem;">{{ error }}</p>
      <button class="btn btn-primary btn-sm" style="margin-top: 1rem;" @click="loadData">重试</button>
    </div>

    <!-- 空态 -->
    <div class="empty" v-else-if="monitors.length === 0">
      <div class="empty-icon">📡</div>
      <p>还没有监控器</p>
      <p style="color: var(--text-muted); font-size: 0.85rem; margin-top: 0.25rem;">
        点击上方按钮添加第一个监控器
      </p>
    </div>

    <!-- 删除确认对话框 -->
    <div class="modal-overlay" v-if="deleteTarget" @click.self="deleteTarget = null">
      <div class="modal-container">
        <div class="modal-header">
          <h2>确认删除</h2>
          <button class="modal-close" @click="deleteTarget = null">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
          </button>
        </div>
        <div class="modal-body">
          <p>确定要删除监控器「{{ deleteTarget }}」吗？</p>
          <p style="color: var(--text-muted); font-size: 0.85rem; margin-top: 0.5rem;">删除后无法恢复。</p>
        </div>
        <div class="modal-footer">
          <button class="btn btn-ghost" @click="deleteTarget = null">取消</button>
          <button class="btn btn-danger" @click="handleDelete">确认删除</button>
        </div>
      </div>
    </div>

    <!-- 分组列表 -->
    <template v-else>
      <div v-for="group in groupList" :key="group.name" class="group-section">
        <div class="group-header">
          <h2 class="group-title">{{ group.name }}</h2>
          <span class="group-count">{{ group.items.length }} 个监控器</span>
        </div>
        <div class="group-list">
          <MonitorCard
            v-for="m in group.items"
            :key="m.name"
            :monitor="m"
            @start="handleStart(m.name)"
            @stop="handleStop(m.name)"
            @edit="handleEdit(m.name)"
            @delete="deleteTarget = m.name"
            @view="handleView(m.name)"
          />
        </div>
      </div>
    </template>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { fetchMonitors, startMonitor, stopMonitor, deleteMonitor } from '../api/monitors'
import MonitorCard from '../components/MonitorCard.vue'

const router = useRouter()
const monitors = ref([])
const loading = ref(true)
const error = ref(null)
const successMsg = ref('')
const pageErrorMsg = ref('')
const deleteTarget = ref(null)

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

const groupList = computed(() => {
  const map = {}
  for (const m of monitors.value) {
    const g = m.group || '默认'
    if (!map[g]) map[g] = { name: g, items: [] }
    map[g].items.push(m)
  }
  const keys = Object.keys(map).sort((a, b) => {
    if (a === '默认') return -1
    if (b === '默认') return 1
    return a.localeCompare(b, 'zh')
  })
  return keys.map(k => map[k])
})

onMounted(loadData)

async function loadData() {
  loading.value = true
  error.value = null
  try {
    const res = await fetchMonitors()
    if (res.code === 0) {
      monitors.value = res.data || []
    } else {
      error.value = res.message || '加载失败'
    }
  } catch (e) {
    error.value = e.message || '网络错误'
  } finally {
    loading.value = false
  }
}

async function handleStart(name) {
  try {
    await startMonitor(name)
    showSuccess(`「${name}」已启动`)
    await loadData()
  } catch (e) {
    showError('启动失败: ' + (e.response?.data?.message || e.message))
  }
}

async function handleStop(name) {
  try {
    await stopMonitor(name)
    showError(`「${name}」已暂停`)
    await loadData()
  } catch (e) {
    showError('暂停失败: ' + (e.response?.data?.message || e.message))
  }
}

function handleEdit(name) {
  router.push(`/edit/${encodeURIComponent(name)}`)
}

async function handleDelete() {
  const name = deleteTarget.value
  deleteTarget.value = null
  try {
    await deleteMonitor(name)
    showSuccess(`「${name}」已删除`)
    await loadData()
  } catch (e) {
    showError('删除失败: ' + (e.response?.data?.message || e.message))
  }
}

function handleView(name) {
  router.push(`/monitor/${encodeURIComponent(name)}`)
}
</script>

<style scoped>
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 2rem;
}

.page-header h1 {
  font-size: 1.5rem;
  font-weight: 700;
  color: var(--text);
  margin-bottom: 0.25rem;
}

.page-desc {
  color: var(--text-muted);
  font-size: 0.85rem;
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

.group-section {
  margin-bottom: 1.5rem;
}

.group-header {
  display: flex;
  align-items: baseline;
  gap: 0.75rem;
  margin-bottom: 0.75rem;
  padding-bottom: 0.5rem;
  border-bottom: 1px solid var(--border-light);
}

.group-title {
  font-size: 1rem;
  font-weight: 600;
  color: var(--text);
}

.group-count {
  font-size: 0.8rem;
  color: var(--text-muted);
}

.group-list {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

@media (max-width: 768px) {
  .page-header {
    flex-direction: column;
    gap: 0.75rem;
  }
}
</style>