<template>
  <div class="scan-rule-page">
    <div class="page-header">
      <div>
        <h1>扫描规则</h1>
        <p class="page-desc">管理可复用的扫描规则模板，优先于内置规则参与扫描。</p>
      </div>
      <div class="header-actions">
        <button class="btn btn-primary" @click="openCreate">新增规则</button>
      </div>
    </div>

    <div class="toast toast-success" v-if="successMsg">{{ successMsg }}</div>
    <div class="toast toast-warning" v-if="pageErrorMsg">{{ pageErrorMsg }}</div>

    <div class="loading" v-if="loading">
      <div class="spinner" />
      <p>加载扫描规则...</p>
    </div>

    <template v-else-if="rules.length === 0">
      <div class="empty">
        <div class="empty-icon">🧭</div>
        <p>还没有扫描规则</p>
        <p style="color: var(--text-muted); font-size: 0.8125rem; margin-top: 0.25rem;">
          创建规则后，扫描时会优先尝试命中这些模板。
        </p>
        <button class="btn btn-primary btn-sm" style="margin-top: 1rem;" @click="openCreate">新增规则</button>
      </div>
    </template>

    <template v-else>
      <div class="settings-section">
        <div class="section-header">
          <h2>规则模板（{{ rules.length }}）</h2>
          <p class="section-desc">按优先级排序，启用的规则会在扫描时优先参与候选生成。</p>
        </div>

        <div v-for="rule in rules" :key="rule.id" class="rule-card">
          <div class="rule-header">
            <div class="rule-info">
              <span class="rule-name">{{ rule.name }}</span>
              <span class="rule-badge">{{ rule.enabled ? '已启用' : '已禁用' }}</span>
              <span class="rule-priority">优先级 {{ rule.priority }}</span>
            </div>
            <div class="rule-actions">
              <button class="icon-btn" title="测试" @click="openTest(rule)">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="16" height="16"><polygon points="5 3 19 12 5 21 5 3"/></svg>
              </button>
              <button class="icon-btn" title="编辑" @click="openEdit(rule)">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="16" height="16"><path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"/><path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"/></svg>
              </button>
              <button class="icon-btn icon-btn-danger" title="删除" @click="confirmDelete(rule)">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="16" height="16"><polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/></svg>
              </button>
            </div>
          </div>

          <div class="rule-meta">
            <span>URL 包含：<code>{{ rule.url_contains }}</code></span>
            <span>Container：<code>{{ rule.container }}</code></span>
            <span>Item：<code>{{ rule.item }}</code></span>
          </div>

          <div class="rule-desc" v-if="rule.description">{{ rule.description }}</div>

          <div class="rule-fields" v-if="rule.fields && rule.fields.length">
            <div class="rule-field" v-for="field in rule.fields" :key="field.name + ':' + field.selector">
              <span class="field-name">{{ field.name }}</span>
              <code>{{ field.selector || '(当前项文本)' }}</code>
              <span class="field-type">{{ field.type }}</span>
            </div>
          </div>
        </div>
      </div>
    </template>

    <div class="modal-overlay" v-if="showModal" @click.self="showModal = false">
      <div class="modal-container modal-lg">
        <div class="modal-header">
          <h2>{{ editingRule ? '编辑扫描规则' : '新增扫描规则' }}</h2>
          <button class="modal-close" @click="showModal = false">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="18" height="18"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
          </button>
        </div>
        <div class="modal-body">
          <div class="form-group">
            <label>规则名称</label>
            <input v-model="form.name" class="form-input" placeholder="如 澎湃快讯时间线" />
          </div>
          <div class="form-group">
            <label>URL 包含</label>
            <input v-model="form.url_contains" class="form-input" placeholder="如 thepaper.cn/expressNews" />
          </div>
          <div class="form-group">
            <label>容器选择器</label>
            <input v-model="form.container" class="form-input" placeholder="如 ul.ant-timeline" />
          </div>
          <div class="form-group">
            <label>列表项选择器</label>
            <input v-model="form.item" class="form-input" placeholder="如 li.ant-timeline-item" />
          </div>
          <div class="form-group inline-grid">
            <div>
              <label>优先级</label>
              <input v-model.number="form.priority" class="form-input" type="number" min="1" placeholder="50" />
            </div>
            <div>
              <label>启用状态</label>
              <label class="checkbox-label"><input v-model="form.enabled" type="checkbox" /> 启用此规则</label>
            </div>
          </div>
          <div class="form-group">
            <label>说明（可选）</label>
            <textarea v-model="form.description" class="form-input" rows="3" placeholder="如：适用于澎湃快讯时间线结构" />
          </div>

          <FieldEditor v-model="form.fields" />

          <div class="form-error" v-if="modalError">{{ modalError }}</div>
        </div>
        <div class="modal-footer">
          <button class="btn btn-ghost" @click="showModal = false">取消</button>
          <button class="btn btn-primary" :disabled="modalSaving" @click="handleSaveRule">{{ modalSaving ? '保存中...' : '保存' }}</button>
        </div>
      </div>
    </div>

    <div class="modal-overlay" v-if="testTarget" @click.self="closeTest()">
      <div class="modal-container modal-lg">
        <div class="modal-header">
          <h2>测试扫描规则</h2>
          <button class="modal-close" @click="closeTest()">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="18" height="18"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
          </button>
        </div>
        <div class="modal-body">
          <div class="form-group">
            <label>测试 URL</label>
            <input v-model="testForm.url" class="form-input" placeholder="https://example.com/announce/" />
          </div>
          <div class="form-group">
            <label>关键词（多个用逗号隔开）</label>
            <input v-model="testForm.keywords" class="form-input" placeholder="面试,录用,公告" />
          </div>
          <div class="form-actions" style="justify-content: flex-start; margin-top: 0.5rem;">
            <button class="btn btn-primary" :disabled="testingRule" @click="handleTestRule">{{ testingRule ? '测试中...' : '开始测试' }}</button>
          </div>

          <div class="form-error" v-if="testError" style="margin-top: 0.75rem;">{{ testError }}</div>

          <div class="preview-results" v-if="testResult && testResult.containers && testResult.containers.length > 0" style="margin-top: 1rem;">
            <div v-for="(container, ci) in testResult.containers" :key="ci" class="preview-card">
              <div class="preview-card-header">
                <span class="candidate-badge">{{ container.strategy || 'candidate' }}</span>
                <span class="candidate-count">{{ container.item_count }} 条</span>
              </div>
              <div class="candidate-selector"><code>{{ container.config?.container }}</code> / <code>{{ container.config?.item }}</code></div>
              <div class="sample-list">
                <div v-for="(item, ii) in container.sample_items" :key="ii" class="sample-item">
                  <span class="sample-title">{{ item.title }}</span>
                  <span class="sample-meta" v-if="item.date">{{ item.date }}</span>
                </div>
              </div>
            </div>
          </div>
          <div class="empty" v-else-if="testRan" style="margin-top: 1rem;">
            <p>未匹配到结果</p>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn btn-ghost" @click="closeTest()">关闭</button>
        </div>
      </div>
    </div>

    <div class="modal-overlay" v-if="deleteTarget" @click.self="deleteTarget = null">
      <div class="modal-container" style="max-width: 400px;">
        <div class="modal-header">
          <h2>确认删除</h2>
          <button class="modal-close" @click="deleteTarget = null">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="18" height="18"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
          </button>
        </div>
        <div class="modal-body">
          <p>确定删除扫描规则「{{ deleteTarget.name }}」吗？</p>
        </div>
        <div class="modal-footer">
          <button class="btn btn-ghost" @click="deleteTarget = null">取消</button>
          <button class="btn btn-danger" @click="handleDeleteRule">确认删除</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { fetchScanRules, createScanRule, updateScanRule, deleteScanRule, testScanRule } from '../api/monitors'
import FieldEditor from '../components/FieldEditor.vue'
import { useToastMessages } from '../composables/useToastMessages'

const { successMsg, pageErrorMsg, showSuccess, showError } = useToastMessages()

const loading = ref(true)
const rules = ref([])
const showModal = ref(false)
const editingRule = ref(null)
const deleteTarget = ref(null)
const modalSaving = ref(false)
const modalError = ref('')

// 测试弹窗
const testTarget = ref(null)
const testForm = ref({ url: '', keywords: '' })
const testingRule = ref(false)
const testError = ref('')
const testResult = ref(null)
const testRan = ref(false)

const form = ref({
  name: '',
  url_contains: '',
  container: '',
  item: '',
  priority: 50,
  enabled: true,
  description: '',
  fields: [{ name: 'title', selector: '', type: 'text', attr: '', transform: '' }],
})

function openTest(rule) {
  testTarget.value = rule
  testForm.value = { url: '', keywords: '' }
  testError.value = ''
  testResult.value = null
  testRan.value = false
}

function closeTest() {
  testTarget.value = null
  testResult.value = null
  testRan.value = false
}

async function handleTestRule() {
  if (!testForm.value.url.trim()) { testError.value = '请输入测试 URL'; return }
  // 关键词不再强制要求，允许无关键词进行扫描规则模板测试
  testingRule.value = true
  testError.value = ''
  testResult.value = null
  testRan.value = false
  try {
    const res = await testScanRule(testTarget.value.id, { url: testForm.value.url.trim() })
    if (res.code === 0) {
      testResult.value = res.data
    } else {
      testError.value = res.message || '扫描失败'
    }
  } catch (e) {
    testError.value = '测试失败: ' + (e.response?.data?.message || e.message)
  } finally {
    testingRule.value = false
    testRan.value = true
  }
}

function resetForm() {
  form.value = {
    name: '',
    url_contains: '',
    container: '',
    item: '',
    priority: 50,
    enabled: true,
    description: '',
    fields: [{ name: 'title', selector: '', type: 'text', attr: '', transform: '' }],
  }
}

onMounted(loadAll)

async function loadAll() {
  loading.value = true
  try {
    const res = await fetchScanRules()
    if (res.code === 0) rules.value = res.data || []
  } catch (e) {
    showError('加载失败: ' + (e.response?.data?.message || e.message))
  } finally {
    loading.value = false
  }
}

function openCreate() {
  editingRule.value = null
  resetForm()
  modalError.value = ''
  showModal.value = true
}

function openEdit(rule) {
  editingRule.value = rule
  form.value = {
    name: rule.name,
    url_contains: rule.url_contains,
    container: rule.container,
    item: rule.item,
    priority: rule.priority,
    enabled: rule.enabled,
    description: rule.description || '',
    fields: (rule.fields && rule.fields.length > 0)
      ? rule.fields.map(f => ({ name: f.name || '', selector: f.selector || '', type: f.type || 'text', attr: f.attr || '', transform: f.transform || '' }))
      : [{ name: 'title', selector: '', type: 'text', attr: '', transform: '' }],
  }
  modalError.value = ''
  showModal.value = true
}

function confirmDelete(rule) {
  deleteTarget.value = rule
}

async function handleDeleteRule() {
  const rule = deleteTarget.value
  deleteTarget.value = null
  try {
    await deleteScanRule(rule.id)
    rules.value = rules.value.filter(r => r.id !== rule.id)
    showSuccess(`「${rule.name}」已删除`)
  } catch (e) {
    showError('删除失败: ' + (e.response?.data?.message || e.message))
  }
}

async function handleSaveRule() {
  if (!form.value.name.trim()) { modalError.value = '请输入规则名称'; return }
  if (!form.value.url_contains.trim()) { modalError.value = '请输入 URL 包含'; return }
  if (!form.value.container.trim()) { modalError.value = '请输入容器选择器'; return }
  if (!form.value.item.trim()) { modalError.value = '请输入列表项选择器'; return }
  if (!form.value.fields.some(f => f.name.trim() === 'title')) { modalError.value = '至少需要一个 title 字段'; return }

  modalError.value = ''
  modalSaving.value = true
  try {
    const payload = {
      name: form.value.name.trim(),
      url_contains: form.value.url_contains.trim(),
      container: form.value.container.trim(),
      item: form.value.item.trim(),
      priority: form.value.priority || 50,
      enabled: form.value.enabled,
      description: form.value.description || '',
      fields: form.value.fields.filter(f => f.name.trim() && f.type),
    }
    if (editingRule.value) {
      await updateScanRule(editingRule.value.id, payload)
      showSuccess('规则已更新')
    } else {
      await createScanRule(payload)
      showSuccess('规则已创建')
    }
    showModal.value = false
    await loadAll()
  } catch (e) {
    modalError.value = e.response?.data?.message || e.message
  } finally {
    modalSaving.value = false
  }
}
</script>

<style scoped>
.page-header { display: flex; justify-content: space-between; align-items: flex-start; margin-bottom: 1.5rem; }
.page-header h1 { font-size: 1.5rem; font-weight: 700; color: var(--text); margin-top: 0.5rem; }
.rule-card { background: var(--bg-card); border-radius: var(--radius-lg); padding: 1rem; margin-bottom: 0.75rem; }
.rule-header { display: flex; justify-content: space-between; align-items: center; gap: 1rem; }
.rule-info { display: flex; align-items: center; gap: 0.5rem; flex-wrap: wrap; }
.rule-name { font-weight: 700; font-size: 0.9375rem; color: var(--text); }
.rule-badge, .rule-priority { font-size: 0.6875rem; font-weight: 700; padding: 0.15rem 0.5rem; border-radius: var(--radius-pill); }
.rule-badge { color: var(--text); background: var(--green); }
.rule-priority { color: var(--text-secondary); background: var(--bg-elevated); }
.rule-actions { display: flex; gap: 0.25rem; }
.rule-meta { display: flex; flex-direction: column; gap: 0.25rem; margin-top: 0.5rem; font-size: 0.8125rem; color: var(--text-secondary); }
.rule-desc { margin-top: 0.5rem; color: var(--text-secondary); font-size: 0.875rem; }
.rule-fields { display: flex; flex-wrap: wrap; gap: 0.5rem; margin-top: 0.75rem; }
.rule-field { display: inline-flex; align-items: center; gap: 0.4rem; padding: 0.3rem 0.55rem; border-radius: var(--radius-pill); background: var(--bg-elevated); font-size: 0.75rem; }
.field-name { font-weight: 700; color: var(--text); }
.field-type { color: var(--text-muted); }
.inline-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 1rem; }
.modal-lg { max-width: 860px; }
.checkbox-label { display: inline-flex; align-items: center; gap: 0.5rem; margin-top: 0.5rem; }
@media (max-width: 720px) { .inline-grid { grid-template-columns: 1fr; } }
</style>
