<template>
  <div class="type-selector">
    <div
      class="type-card"
      :class="{ selected: modelValue === 'presence' }"
      @click="$emit('update:modelValue', 'presence')"
    >
      <div class="type-icon type-icon-presence">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="28" height="28">
          <circle cx="12" cy="12" r="10"/>
          <line x1="12" y1="8" x2="12" y2="16"/>
          <line x1="8" y1="12" x2="16" y2="12"/>
        </svg>
      </div>
      <div class="type-info">
        <h3 class="type-title">网页新增监控</h3>
        <p class="type-desc">检测页面中出现的新条目</p>
        <div class="type-examples">
          <span class="type-tag">公告</span>
          <span class="type-tag">新闻</span>
          <span class="type-tag">更新日志</span>
          <span class="type-tag">商品上新</span>
        </div>
      </div>
      <div class="type-check" v-if="modelValue === 'presence'">
        <svg viewBox="0 0 24 24" fill="currentColor" width="20" height="20"><path d="M9 16.17L4.83 12l-1.42 1.41L9 19 21 7l-1.41-1.41z"/></svg>
      </div>
    </div>

    <div
      class="type-card"
      :class="{ selected: modelValue === 'field_transition' }"
      @click="$emit('update:modelValue', 'field_transition')"
    >
      <div class="type-icon type-icon-numeric">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="28" height="28">
          <polyline points="22 12 18 12 15 21 9 3 6 12 2 12"/>
        </svg>
      </div>
      <div class="type-info">
        <h3 class="type-title">价格监控</h3>
        <p class="type-desc">监控价格下降幅度或降到目标价</p>
        <div class="type-examples">
          <span class="type-tag">价格下降</span>
          <span class="type-tag">降价阈值</span>
          <span class="type-tag">到价提醒</span>
        </div>
      </div>
      <div class="type-check" v-if="modelValue === 'field_transition'">
        <svg viewBox="0 0 24 24" fill="currentColor" width="20" height="20"><path d="M9 16.17L4.83 12l-1.42 1.41L9 19 21 7l-1.41-1.41z"/></svg>
      </div>
    </div>
  </div>
</template>

<script setup>
defineProps({
  modelValue: { type: String, default: 'presence' },
})
defineEmits(['update:modelValue'])
</script>

<style scoped>
.type-selector {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 1rem;
  margin-bottom: 1.5rem;
}

.type-card {
  position: relative;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.75rem;
  padding: 1.5rem 1rem;
  background: var(--bg-card);
  border: 2px solid transparent;
  border-radius: var(--radius-lg);
  cursor: pointer;
  transition: var(--transition);
  text-align: center;
}

.type-card:hover {
  background: var(--bg-hover);
  border-color: var(--border);
}

.type-card.selected {
  border-color: var(--green);
  background: rgba(29, 185, 84, 0.04);
}

.type-icon {
  width: 56px;
  height: 56px;
  border-radius: var(--radius-circle);
  display: flex;
  align-items: center;
  justify-content: center;
}

.type-icon-presence {
  background: var(--success-bg);
  color: var(--green);
}

.type-icon-numeric {
  background: #e3f2fd;
  color: #1976d2;
}

.dark .type-icon-numeric {
  background: #0d2137;
  color: #64b5f6;
}

.type-info {
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
}

.type-title {
  font-size: 0.9375rem;
  font-weight: 700;
  color: var(--text);
}

.type-desc {
  font-size: 0.75rem;
  color: var(--text-secondary);
  line-height: 1.4;
}

.type-examples {
  display: flex;
  flex-wrap: wrap;
  gap: 0.3rem;
  justify-content: center;
  margin-top: 0.25rem;
}

.type-tag {
  font-size: 0.625rem;
  font-weight: 700;
  color: var(--text-muted);
  background: var(--bg-elevated);
  padding: 0.1rem 0.4rem;
  border-radius: var(--radius-pill);
}

.type-check {
  position: absolute;
  top: 0.75rem;
  right: 0.75rem;
  color: var(--green);
}

@media (max-width: 600px) {
  .type-selector {
    grid-template-columns: 1fr;
  }
}
</style>
