<template>
  <span class="status-badge" :class="`badge-${type}`">
    <span class="badge-dot" />
    {{ label }}
  </span>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  status: { type: String, default: 'unknown' },
})

const type = computed(() => {
  const s = props.status
  if (s === 'running') return 'success'
  if (s === 'stopped') return 'default'
  if (s === 'error') return 'error'
  return 'default'
})

const label = computed(() => {
  const s = props.status
  if (s === 'running') return '运行中'
  if (s === 'stopped') return '已停止'
  if (s === 'error') return '异常'
  return '未知'
})
</script>

<style scoped>
.status-badge {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  padding: 0.2rem 0.6rem;
  border-radius: 20px;
  font-size: 0.78rem;
  font-weight: 500;
}

.badge-dot {
  width: 7px;
  height: 7px;
  border-radius: 50%;
}

.badge-success {
  background: var(--success-bg, #ecfdf5);
  color: var(--success, #10b981);
}
.badge-success .badge-dot {
  background: var(--success, #10b981);
}

.badge-error {
  background: var(--error-bg, #fef2f2);
  color: var(--error, #ef4444);
}
.badge-error .badge-dot {
  background: var(--error, #ef4444);
}

.badge-default {
  background: var(--bg-hover, #f1f5f9);
  color: var(--text-muted, #94a3b8);
}
.badge-default .badge-dot {
  background: var(--text-muted, #94a3b8);
}
</style>