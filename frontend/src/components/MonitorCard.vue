<template>
  <div class="monitor-row" :class="{ 'row-error': isError, 'row-stopped': !isRunning }">
    <div class="row-left">
      <StatusBadge :status="statusText" />
      <span class="row-name">{{ monitor.name }}</span>
    </div>

    <div class="row-center">
      <span class="row-url" :title="monitor.url">{{ monitor.url }}</span>
    </div>

    <div class="row-meta">
      <span class="meta-item" v-if="monitor.last_check">
        {{ formatTime(monitor.last_check) }}
      </span>
      <span class="meta-item" v-if="monitor.updates_count > 0" style="color: var(--primary);">
        {{ monitor.updates_count }} 条更新
      </span>
      <span class="meta-item meta-error" v-if="monitor.last_error" :title="monitor.last_error">
        错误
      </span>
    </div>

    <div class="row-actions">
      <button class="btn btn-action" :class="isRunning ? 'btn-pause' : 'btn-play'" @click="$emit(isRunning ? 'stop' : 'start')" :title="isRunning ? '暂停' : '启动'">
        <!-- 暂停图标（两条竖杠） -->
        <svg v-if="isRunning" viewBox="0 0 24 24" fill="currentColor" width="18" height="18">
          <rect x="6" y="4" width="4" height="16" rx="1"/>
          <rect x="14" y="4" width="4" height="16" rx="1"/>
        </svg>
        <!-- 播放图标（三角形） -->
        <svg v-else viewBox="0 0 24 24" fill="currentColor" width="18" height="18">
          <path d="M8 5v14l11-7z"/>
        </svg>
      </button>
      <button class="btn-icon btn-icon-edit" title="编辑" @click="$emit('edit')">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"/><path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"/></svg>
      </button>
      <button class="btn-icon btn-icon-view" title="详情" @click="$emit('view')">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"/><circle cx="12" cy="12" r="3"/></svg>
      </button>
      <button class="btn-icon btn-icon-danger" title="删除" @click="$emit('delete')">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/></svg>
      </button>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import StatusBadge from './StatusBadge.vue'

const props = defineProps({
  monitor: { type: Object, required: true },
})

defineEmits(['start', 'stop', 'edit', 'delete', 'view'])

const isRunning = computed(() => props.monitor.is_running)
const isError = computed(() => !!props.monitor.last_error)
const statusText = computed(() => {
  if (isError.value) return 'error'
  return isRunning.value ? 'running' : 'stopped'
})

function formatTime(t) {
  if (!t) return '—'
  const d = new Date(t)
  const now = new Date()
  const sameDay = d.toDateString() === now.toDateString()
  if (sameDay) return d.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' })
  return d.toLocaleDateString('zh-CN', { month: '2-digit', day: '2-digit' })
}
</script>

<style scoped>
.monitor-row {
  display: flex;
  align-items: center;
  gap: 1rem;
  padding: 0.75rem 1rem;
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  transition: var(--transition);
}

.monitor-row:hover {
  border-color: var(--primary-light);
  box-shadow: var(--shadow-sm);
}

.row-error {
  border-color: var(--error);
  border-left: 3px solid var(--error);
}

.row-stopped {
  opacity: 0.65;
}

.row-left {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  min-width: 150px;
  flex-shrink: 0;
}

.row-name {
  font-weight: 600;
  font-size: 0.9rem;
  color: var(--text);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.row-center {
  flex: 1;
  min-width: 0;
}

.row-url {
  font-size: 0.8rem;
  color: var(--text-muted);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  display: block;
}

.row-meta {
  display: flex;
  gap: 0.75rem;
  flex-shrink: 0;
}

.meta-item {
  font-size: 0.78rem;
  color: var(--text-secondary);
  white-space: nowrap;
}

.meta-error {
  color: var(--error);
}

.row-actions {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  flex-shrink: 0;
}

/* 启动/暂停大按钮 */
.btn-action {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 38px;
  height: 38px;
  border: none;
  border-radius: var(--radius);
  cursor: pointer;
  transition: var(--transition);
  padding: 0;
}

.btn-play {
  background: var(--success);
  color: white;
}

.btn-play:hover {
  background: #059669;
}

.btn-pause {
  background: var(--warning);
  color: white;
}

.btn-pause:hover {
  background: #d97706;
}

/*图标按钮增大*/
.btn-icon {
  background: none;
  border: none;
  cursor: pointer;
  padding: 0.5rem;
  border-radius: var(--radius-sm);
  transition: var(--transition);
  color: var(--text-muted);
  display: inline-flex;
  align-items: center;
}

.btn-icon svg {
  width: 18px;
  height: 18px;
}

.btn-icon:hover {
  background: var(--bg-hover);
  color: var(--primary);
}

.btn-icon-edit:hover {
  color: var(--primary);
}

.btn-icon-view:hover {
  color: var(--primary);
}

.btn-icon-danger:hover {
  background: var(--error-bg);
  color: var(--error);
}

@media (max-width: 768px) {
  .monitor-row {
    flex-wrap: wrap;
    gap: 0.5rem;
  }
  .row-left {
    min-width: 0;
    flex: 1;
  }
  .row-meta {
    order: 10;
    width: 100%;
  }
}
</style>