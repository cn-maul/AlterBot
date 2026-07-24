<template>
  <div class="settings-section">
    <div class="section-header">
      <h2>提取配置</h2>
      <div class="section-actions">
        <button class="btn btn-sm btn-ghost" @click="$emit('preview')" :disabled="!url">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="14" height="14"><circle cx="11" cy="11" r="8"/><line x1="21" y1="21" x2="16.65" y2="16.65"/></svg>
          预览抓取
        </button>
      </div>
    </div>

    <div class="form-group">
      <label>容器选择器</label>
      <input :value="modelValue.containerSelector" @input="update('containerSelector', $event.target.value)" class="form-input" placeholder="如 div.hap_infoBox" />
    </div>

    <div class="form-group">
      <label>列表项选择器（可选）</label>
      <input :value="modelValue.itemSelector" @input="update('itemSelector', $event.target.value)" class="form-input" placeholder="如 div.hap_infoOne" />
    </div>

    <div class="fields-section">
      <div class="fields-header">
        <span class="fields-title">提取字段</span>
        <button class="btn btn-sm btn-ghost" @click="addField">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="14" height="14"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
          添加字段
        </button>
      </div>

      <div class="fields-list">
        <div v-for="(field, index) in modelValue.fields" :key="index" class="field-card">
          <div class="field-grid">
            <div class="form-group">
              <label>名称</label>
              <input :value="field.name" @input="updateField(index, 'name', $event.target.value)" class="form-input" placeholder="如 title" />
            </div>
            <div class="form-group">
              <label>选择器</label>
              <input :value="field.selector" @input="updateField(index, 'selector', $event.target.value)" class="form-input" placeholder="如 a.title" />
            </div>
            <div class="form-group">
              <label>类型</label>
              <select :value="field.type" @change="updateField(index, 'type', $event.target.value)" class="form-input">
                <option value="text">文本 (text)</option>
                <option value="attr">属性 (attr)</option>
              </select>
            </div>
            <div class="form-group" v-if="field.type === 'attr'">
              <label>属性名</label>
              <input :value="field.attr" @input="updateField(index, 'attr', $event.target.value)" class="form-input" placeholder="默认 href" />
            </div>
            <div class="form-group">
              <label>转换</label>
              <input :value="field.transform" @input="updateField(index, 'transform', $event.target.value)" class="form-input" placeholder="如 trim([])" />
            </div>
          </div>
          <button class="field-remove-btn" title="删除字段" @click="removeField(index)">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="16" height="16"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
          </button>
        </div>
      </div>

      <div class="empty" v-if="modelValue.fields.length === 0">
        <p>暂无字段，点击上方按钮添加</p>
      </div>
    </div>
  </div>
</template>

<script setup>
const props = defineProps({
  modelValue: { type: Object, required: true },
  url: { type: String, default: '' },
})
const emit = defineEmits(['update:modelValue', 'preview'])

function update(key, value) {
  emit('update:modelValue', { ...props.modelValue, [key]: value })
}

function updateField(index, key, value) {
  const fields = [...props.modelValue.fields]
  fields[index] = { ...fields[index], [key]: value }
  emit('update:modelValue', { ...props.modelValue, fields })
}

function addField() {
  const fields = [...props.modelValue.fields, { name: '', selector: '', type: 'text', attr: '', transform: '' }]
  emit('update:modelValue', { ...props.modelValue, fields })
}

function removeField(index) {
  const fields = [...props.modelValue.fields]
  fields.splice(index, 1)
  emit('update:modelValue', { ...props.modelValue, fields })
}
</script>

<style scoped>
.section-header { display: flex; justify-content: space-between; align-items: center; }
.section-actions { flex-shrink: 0; }

.fields-section { margin-top: 0.5rem; }
.fields-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 0.75rem;
}
.fields-title {
  font-size: 0.75rem;
  font-weight: 700;
  color: var(--text-secondary);
  text-transform: uppercase;
  letter-spacing: 1px;
}
.fields-list { display: flex; flex-direction: column; gap: 0.5rem; }
.field-card {
  display: flex;
  gap: 0.5rem;
  align-items: flex-start;
  padding: 0.75rem;
  background: var(--bg-surface);
  border-radius: var(--radius-lg);
}
.field-grid {
  flex: 1;
  display: grid;
  grid-template-columns: 1fr 1fr 1fr 1fr 1fr;
  gap: 0.5rem;
}
.field-remove-btn {
  background: none;
  border: none;
  cursor: pointer;
  padding: 0.4rem;
  border-radius: var(--radius-circle);
  transition: var(--transition);
  color: var(--text-muted);
  display: inline-flex;
  align-items: center;
  flex-shrink: 0;
  margin-top: 1.5rem;
}
.field-remove-btn:hover { background: var(--error-bg); color: var(--error); }

@media (max-width: 768px) {
  .field-grid { grid-template-columns: 1fr 1fr; }
}
</style>
