<template>
  <div class="settings-section">
    <div class="section-header">
      <h2>价格监控规则</h2>
      <p class="section-desc">配置商品场景、价格字段和触发条件</p>
    </div>

    <div class="subsection">
      <h3 class="subsection-title">页面场景</h3>
      <div class="filter-mode-row">
        <label class="radio-label" :class="{ active: form.rule.pageMode === 'single' }">
          <input type="radio" :checked="form.rule.pageMode === 'single'" @change="updatePageMode('single')" />
          单商品详情页
        </label>
        <label class="radio-label" :class="{ active: form.rule.pageMode === 'list' }">
          <input type="radio" :checked="form.rule.pageMode === 'list'" @change="updatePageMode('list')" />
          商品列表页
        </label>
      </div>
      <p class="subsection-desc" style="margin-top: 0.5rem;">
        单商品页可使用页面 URL 作为身份；商品列表必须配置列表项选择器和唯一身份字段。
      </p>
    </div>

    <div class="subsection">
      <h3 class="subsection-title">商品身份</h3>
      <p class="subsection-desc">用于关联两次检查中的同一件商品。该字段必须稳定且唯一，不能使用列表位置。</p>
      <IdentityFieldEditor
        :modelValue="form.rule.identity"
        @update:modelValue="updateIdentity"
        :fields="form.extraction.fields.filter(field => field.name !== form.rule.target.field)"
        :allowSourceUrl="form.rule.pageMode !== 'list'"
      />
    </div>

    <div class="subsection">
      <h3 class="subsection-title">被监控字段</h3>
      <div class="form-row">
        <div class="form-group">
          <label>字段名称</label>
          <select :value="form.rule.target.field" @change="updateTargetField($event.target.value)" class="form-input">
            <option value="">选择监控字段</option>
            <option v-for="f in form.extraction.fields" :key="f.name" :value="f.name">{{ f.name }}</option>
          </select>
        </div>
        <div class="form-group">
          <label>数据类型</label>
          <select value="money" class="form-input" disabled>
            <option value="money">金额</option>
          </select>
          <p class="type-hint">当前版本支持金额降价和到价提醒</p>
        </div>
      </div>
      <div class="form-group">
        <label>CSS 选择器</label>
        <p class="hint">通常与提取配置中的字段选择器相同。如果在提取配置中已定义了该字段的选择器，此处无需重复填写。</p>
      </div>
    </div>

    <div class="subsection">
      <h3 class="subsection-title">变化规则</h3>
      <div class="filter-mode-row rule-mode-row">
        <label class="radio-label" :class="{ active: form.rule.transition.operator === 'decreased' }">
          <input type="radio" :checked="form.rule.transition.operator === 'decreased'" @change="updateOperator('decreased')" />
          价格发生下降
        </label>
        <label class="radio-label" :class="{ active: form.rule.transition.operator === 'at_or_below' }">
          <input type="radio" :checked="form.rule.transition.operator === 'at_or_below'" @change="updateOperator('at_or_below')" />
          降到目标价及以下
        </label>
      </div>
      <ThresholdEditor
        v-if="form.rule.transition.operator === 'decreased'"
        :modelValue="form.rule.transition"
        @update:modelValue="updateTransition"
      />
      <div class="target-price-editor" v-else>
        <div class="form-group">
          <label>目标价格</label>
          <div class="input-with-suffix">
            <input
              :value="form.rule.transition.targetPrice"
              @input="updateTargetPrice($event.target.value)"
              class="form-input"
              type="number"
              min="0"
              step="0.01"
              placeholder="例如 199.00"
            />
            <span class="input-suffix">元</span>
          </div>
          <p class="hint">仅在价格从目标价以上降到目标价或以下时通知；持续低于目标价不会重复推送。</p>
        </div>
      </div>
    </div>

    <div class="baseline-notice">
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="16" height="16"><circle cx="12" cy="12" r="10"/><line x1="12" y1="16" x2="12" y2="12"/><line x1="12" y1="8" x2="12.01" y2="8"/></svg>
      <span>首次检查仅建立价格基线，不会发送通知。后续检测到符合价格规则的变化后才会推送。</span>
    </div>
  </div>
</template>

<script setup>
import IdentityFieldEditor from './IdentityFieldEditor.vue'
import ThresholdEditor from './ThresholdEditor.vue'

const props = defineProps({
  form: { type: Object, required: true },
})
const emit = defineEmits(['update:form'])

function updatePageMode(pageMode) {
  const identity = pageMode === 'single'
    ? { mode: 'source_url', field: '' }
    : { mode: 'field', field: props.form.rule.identity.field || '' }
  emit('update:form', {
    ...props.form,
    rule: { ...props.form.rule, pageMode, identity },
  })
}

function updateIdentity(val) {
  emit('update:form', {
    ...props.form,
    rule: { ...props.form.rule, identity: val },
  })
}

function updateTarget(key, value) {
  emit('update:form', {
    ...props.form,
    rule: {
      ...props.form.rule,
      target: { ...props.form.rule.target, [key]: value },
    },
  })
}

function updateTargetField(fieldName) {
  updateTarget('field', fieldName)
}

function updateTransition(val) {
  emit('update:form', {
    ...props.form,
    rule: { ...props.form.rule, transition: val },
  })
}

function updateOperator(operator) {
  updateTransition({ ...props.form.rule.transition, operator })
}

function updateTargetPrice(targetPrice) {
  updateTransition({ ...props.form.rule.transition, targetPrice })
}
</script>

<style scoped>
.section-header { margin-bottom: 1.25rem; padding-bottom: 0.75rem; border-bottom: 1px solid var(--border-light); }
.section-header h2 { font-size: 1.125rem; font-weight: 700; color: var(--text); margin-bottom: 0.15rem; }
.section-desc { font-size: 0.8125rem; color: var(--text-secondary); }

.subsection {
  padding: 1rem;
  background: var(--bg-surface);
  border-radius: var(--radius-lg);
  margin-bottom: 1rem;
}
.subsection-title {
  font-size: 0.875rem;
  font-weight: 700;
  color: var(--text);
  margin-bottom: 0.25rem;
}
.subsection-desc {
  font-size: 0.75rem;
  color: var(--text-secondary);
  margin-bottom: 0.75rem;
}
.form-row { display: flex; gap: 1rem; }
.form-row .form-group { flex: 1; }
.type-hint { font-size: 0.6875rem; color: var(--warning); margin-top: 0.2rem; }
.hint { font-size: 0.75rem; color: var(--text-muted); margin-top: 0.2rem; }
.filter-mode-row { display: flex; gap: 0.5rem; flex-wrap: wrap; }
.rule-mode-row { margin-bottom: 0.75rem; }
.radio-label {
  display: flex; align-items: center; gap: 0.4rem;
  padding: 0.45rem 0.85rem; border-radius: var(--radius-pill);
  font-size: 0.8125rem; font-weight: 700; cursor: pointer;
  background: var(--bg-elevated); color: var(--text-secondary);
}
.radio-label.active { background: var(--green); color: #000; }
.radio-label input { display: none; }
.target-price-editor { margin-top: 0.5rem; }
.input-with-suffix { display: flex; align-items: center; gap: 0.5rem; }
.input-with-suffix .form-input { max-width: 220px; }
.input-suffix { color: var(--text-muted); font-size: 0.8125rem; font-weight: 700; }

.baseline-notice {
  display: flex;
  align-items: flex-start;
  gap: 0.5rem;
  padding: 0.75rem;
  background: var(--bg-elevated);
  border-radius: var(--radius-lg);
  font-size: 0.8125rem;
  color: var(--text-secondary);
  line-height: 1.4;
}
.baseline-notice svg { flex-shrink: 0; margin-top: 1px; color: var(--text-muted); }
</style>
