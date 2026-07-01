<template>
  <div class="field-editor">
    <div class="section-header">
      <h2>提取字段</h2>
    </div>

    <div class="fields-list">
      <div v-for="(field, index) in modelValue" :key="index" class="field-row">
        <div class="field-inputs">
          <div class="form-group">
            <label>名称</label>
            <input v-model="field.name" class="form-input" placeholder="如 title" />
          </div>
          <div class="form-group">
            <label>选择器</label>
            <input v-model="field.selector" class="form-input" placeholder="如 a.title" />
          </div>
          <div class="form-group">
            <label>类型</label>
            <select v-model="field.type" class="form-input">
              <option value="text">文本 (text)</option>
              <option value="attr">属性 (attr)</option>
            </select>
          </div>
          <div class="form-group" v-if="field.type === 'attr'">
            <label>属性名</label>
            <input v-model="field.attr" class="form-input" placeholder="默认 href" />
          </div>
          <div class="form-group">
            <label>转换</label>
            <input v-model="field.transform" class="form-input" placeholder="如 trim([])" />
          </div>
        </div>
        <button class="btn-icon btn-icon-danger" title="删除字段" @click="removeField(index)">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
        </button>
      </div>
    </div>

    <div class="empty" v-if="modelValue.length === 0">
      <p>暂无字段，点击下方按钮添加</p>
    </div>

    <button class="btn btn-ghost btn-sm" @click="addField" style="margin-top: 0.5rem;">
      ＋ 添加字段
    </button>
  </div>
</template>

<script setup>
const props = defineProps({
  modelValue: { type: Array, default: () => [] },
})
const emit = defineEmits(['update:modelValue'])

function addField() {
  emit('update:modelValue', [
    ...props.modelValue,
    { name: '', selector: '', type: 'text', attr: '', transform: '' },
  ])
}

function removeField(index) {
  const copy = [...props.modelValue]
  copy.splice(index, 1)
  emit('update:modelValue', copy)
}
</script>

<style scoped>
.field-editor {
  margin-bottom: 1rem;
}

.fields-list {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.field-row {
  display: flex;
  gap: 0.5rem;
  align-items: flex-start;
  padding: 0.75rem;
  background: var(--bg);
  border-radius: var(--radius);
  border: 1px solid var(--border-light);
}

.field-inputs {
  flex: 1;
  display: grid;
  grid-template-columns: 1fr 1fr 1fr 1fr 1fr;
  gap: 0.5rem;
}

@media (max-width: 768px) {
  .field-inputs {
    grid-template-columns: 1fr 1fr;
  }
}
</style>