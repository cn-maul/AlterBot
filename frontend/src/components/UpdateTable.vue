<template>
  <div class="update-table">
    <div class="loading" v-if="loading">
      <div class="spinner" />
      <p>加载中...</p>
    </div>

    <div class="empty" v-else-if="records.length === 0">
      <p>暂无更新记录</p>
    </div>

    <template v-else>
      <table class="data-table">
        <thead>
          <tr>
            <th>时间</th>
            <th>标题</th>
            <th>通知状态</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="r in records" :key="r.ID">
            <td class="cell-time">{{ formatTime(r.CreatedAt) }}</td>
            <td>
              <a :href="r.URL" target="_blank" rel="noopener" class="record-link">{{ r.Title }}</a>
            </td>
            <td>
              <span class="notify-status" :class="r.Notified ? 'notified' : 'pending'">
                {{ r.Notified ? '✅ 已推送' : '⏳ 待推送' }}
              </span>
            </td>
          </tr>
        </tbody>
      </table>
    </template>
  </div>
</template>

<script setup>
defineProps({
  records: { type: Array, default: () => [] },
  loading: { type: Boolean, default: false },
})

function formatTime(t) {
  if (!t) return '—'
  const d = new Date(t)
  return d.toLocaleString('zh-CN', {
    month: '2-digit', day: '2-digit',
    hour: '2-digit', minute: '2-digit', second: '2-digit',
  })
}
</script>

<style scoped>
.record-link {
  color: var(--primary);
  text-decoration: none;
}

.record-link:hover {
  text-decoration: underline;
}

.notify-status {
  font-size: 0.8rem;
}

.cell-time {
  white-space: nowrap;
  color: var(--text-secondary);
  font-size: 0.8rem;
}
</style>