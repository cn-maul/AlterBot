<template>
  <div class="add-monitor">
    <div class="page-header">
      <div>
        <router-link to="/" class="back-btn">← 返回</router-link>
        <h1>{{ isEdit ? '编辑监控器' : '新增监控器' }}</h1>
      </div>
    </div>

    <!-- ===== Preview Panel ===== -->
    <div class="preview-panel" v-if="showPreview">
      <div class="section-header">
        <h2>预览抓取结果</h2>
        <button class="btn btn-sm btn-ghost" @click="showPreview = false">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="14" height="14"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
          关闭
        </button>
      </div>
      <div class="form-group">
        <label>关键词（辅助验证，多个用逗号隔开）</label>
        <div class="preview-input-row">
          <input v-model="previewKeyword" class="form-input" placeholder="公告" @keyup.enter="runPreview" />
          <button class="btn btn-sm btn-primary" :disabled="previewLoading" @click="runPreview">{{ previewLoading ? '扫描中' : '扫描' }}</button>
        </div>
      </div>
      <div class="loading" v-if="previewLoading"><div class="spinner" /><p>扫描中...</p></div>
      <div class="preview-results" v-else-if="previewData && previewData.containers && previewData.containers.length > 0">
        <div
          v-for="(c, ci) in previewData.containers"
          :key="ci"
          class="preview-card"
          :class="{ selected: selectedPreviewIndex === ci }"
          @click="selectedPreviewIndex = ci"
        >
          <div class="preview-card-header">
            <span class="candidate-badge">{{ c.container_tag.toUpperCase() }}</span>
            <span class="candidate-count">{{ c.item_count }} 条</span>
            <button class="btn btn-sm btn-primary apply-candidate" type="button" @click.stop="applyPreviewCandidate(c)">应用此配置</button>
          </div>
          <div class="candidate-selectors">
            <code>{{ c.config?.container || c.container_css }}</code>
            <span> / </span>
            <code>{{ c.config?.item || c.item_css || '单项' }}</code>
          </div>
          <div class="preview-samples">
            <div v-for="(item, ii) in c.sample_items" :key="ii" class="sample-item">
              <span class="sample-title">{{ item.title }}</span>
            </div>
          </div>
        </div>
      </div>
      <div class="empty" v-else-if="previewScanned">
        <p>未找到匹配内容，试试不同关键词或调整选择器</p>
      </div>
      <div class="form-error" v-if="previewError">{{ previewError }}</div>
    </div>

    <!-- ===== Loading ===== -->
    <div class="loading" v-if="isEdit && loading">
      <div class="spinner" />
      <p>加载配置...</p>
    </div>

    <!-- ===== Form ===== -->
    <template v-else>
      <MonitorForm
        :form="form"
        :showTypeSelector="true"
        :accounts="accounts"
        :error="submitError"
        :validationResult="validationResult"
        :validationLoading="validationLoading"
        :showBaselineWarning="baselineWarning"
        @preview="openPreview"
        @validate="runValidation"
        @update:form="onFormUpdate"
      >
        <template #actions>
          <button v-if="!isEdit" class="btn btn-ghost btn-sm" @click="handleSaveAsRule" :disabled="submitting" style="margin-right: auto;">另存为规则模板</button>
          <button class="btn btn-primary" :disabled="submitting" @click="handleSubmit">
            {{ submitting ? '提交中...' : (isEdit ? '保存修改' : '创建并启动') }}
          </button>
        </template>
      </MonitorForm>
    </template>
  </div>
</template>

<script setup>
import { ref, computed, reactive, onMounted, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { createMonitor, updateMonitor, fetchMonitorConfig, fetchAccounts, previewScan, createScanRule, validateMonitorConfig } from '../api/monitors'
import MonitorForm from '../components/monitor/form/MonitorForm.vue'
import { createEmptyForm, toMonitorRequest, fromMonitorResponse, hasSemanticChange, validateForm, getDetectionFingerprint } from '../composables/useMonitorForm'
import { useToastMessages } from '../composables/useToastMessages'

const router = useRouter()
const route = useRoute()
const { showSuccess, showError } = useToastMessages()

const isEdit = computed(() => !!route.params.name)
const loading = ref(false)
const submitting = ref(false)
const submitError = ref(null)

const form = reactive(createEmptyForm())
const originalFormSnapshot = ref(null)

const accounts = ref([])

// Preview
const showPreview = ref(false)
const previewKeyword = ref('')
const previewLoading = ref(false)
const previewData = ref(null)
const previewScanned = ref(false)
const previewError = ref(null)
const selectedPreviewIndex = ref(null)

// Validation
const validationResult = ref(null)
const validationLoading = ref(false)
const validatedFingerprint = ref('')
const validationAttemptFingerprint = ref('')

// Baseline warning
const baselineWarning = ref(false)

function openPreview() {
  showPreview.value = true
}

async function runPreview() {
  if (!form.basic.url.trim()) return
  previewLoading.value = true
  previewScanned.value = false
  previewError.value = null
  selectedPreviewIndex.value = null
  try {
    const res = await previewScan({
      url: form.basic.url.trim(),
      keywords: previewKeyword.value || (form.monitorType === 'field_transition' ? '价格,售价,优惠' : '公告'),
    })
    if (res.code === 0 && res.data) {
      previewData.value = res.data
    }
  } catch (e) {
    previewError.value = e.response?.data?.message || e.message || '扫描失败'
  }
  previewScanned.value = true
  previewLoading.value = false
}

function normalizeScanFields(fields) {
  return (fields || []).map(field => ({
    name: field.name || '',
    selector: field.selector || '',
    type: field.type || 'text',
    attr: field.attr || '',
    transform: field.transform || '',
  }))
}

function ensurePriceField() {
  if (form.monitorType !== 'field_transition') return
  const currentTarget = form.rule.target.field || 'price'
  if (!form.extraction.fields.some(field => field.name === currentTarget)) {
    form.extraction.fields.push({ name: currentTarget, selector: '.price', type: 'text', attr: '', transform: '' })
  }
  form.rule.target.field = currentTarget
  form.rule.target.valueType = 'money'
  if (!['decreased', 'at_or_below'].includes(form.rule.transition.operator)) {
    form.rule.transition.operator = 'decreased'
  }
}

function applyPreviewCandidate(candidate) {
  const config = candidate.config || {}
  const fields = normalizeScanFields(config.fields)
  form.extraction.containerSelector = config.container || candidate.container_css || ''
  form.extraction.itemSelector = config.item || candidate.item_css || ''
  if (fields.length > 0) form.extraction.fields = fields
  if (form.monitorType === 'field_transition') {
    form.rule.pageMode = form.extraction.itemSelector ? 'list' : 'single'
    if (form.rule.pageMode === 'list') form.rule.identity.mode = 'field'
    ensurePriceField()
  }
  showPreview.value = false
  showSuccess('已应用扫描候选配置，可继续调整字段和规则')
}

function onFormUpdate(newForm) {
  Object.assign(form, newForm)
}

// Watch for semantic changes in edit mode
watch(() => form.monitorType, (monitorType) => {
  if (monitorType === 'field_transition') {
    form.rule.pageMode = form.extraction.itemSelector.trim() ? 'list' : 'single'
    if (form.rule.pageMode === 'list' && form.rule.identity.mode === 'source_url') {
      form.rule.identity = { mode: 'field', field: '' }
    }
    ensurePriceField()
  }
})

watch(
  () => getDetectionFingerprint(form),
  (fingerprint) => {
    if (validatedFingerprint.value && validatedFingerprint.value !== fingerprint) {
      validatedFingerprint.value = ''
      validationResult.value = null
    }
    baselineWarning.value = Boolean(
      isEdit.value && originalFormSnapshot.value && hasSemanticChange(originalFormSnapshot.value, form),
    )
  }
)

watch(
  () => JSON.stringify(toMonitorRequest(form)),
  (fingerprint) => {
    if (validationAttemptFingerprint.value && validationAttemptFingerprint.value !== fingerprint) {
      validationAttemptFingerprint.value = ''
      validationResult.value = null
    }
  }
)

// Load accounts and edit data
onMounted(async () => {
  try {
    const acctRes = await fetchAccounts()
    accounts.value = (acctRes.code === 0 ? acctRes.data : []) || []
  } catch { /* ignore */ }

  if (isEdit.value) {
    loading.value = true
    try {
      const res = await fetchMonitorConfig(route.params.name)
      if (res.code === 0 && res.data) {
        const loaded = fromMonitorResponse(res.data)
        Object.assign(form, loaded)
        originalFormSnapshot.value = JSON.parse(JSON.stringify(form))
      } else {
        submitError.value = res.message || '加载配置失败'
      }
    } catch (e) {
      submitError.value = '加载配置失败: ' + e.message
    } finally {
      loading.value = false
    }
  }
})

// Validate before submit
async function runValidation() {
  validationAttemptFingerprint.value = JSON.stringify(toMonitorRequest(form))
  const localError = validateForm(form)
  if (localError) {
    validationResult.value = { valid: false, errors: [localError], summary: '请先修正表单配置后再验证。' }
    submitError.value = localError
    return false
  }
  validationLoading.value = true
  submitError.value = null
  try {
    const payload = toMonitorRequest(form)
    const res = await validateMonitorConfig(payload)
    if (res.code === 0 && res.data) {
      validationResult.value = res.data
      validatedFingerprint.value = getDetectionFingerprint(form)
      return true
    }
    const message = res.message || '验证失败'
    validationResult.value = { valid: false, errors: [message], summary: '配置未通过验证。' }
    return false
  } catch (e) {
    const message = e.response?.data?.message || e.message || '验证失败'
    validationResult.value = { valid: false, errors: [message], summary: '配置未通过验证。' }
    return false
  } finally {
    validationLoading.value = false
  }
}

async function handleSubmit() {
  const err = validateForm(form)
  if (err) { submitError.value = err; return }
  const semanticChange = !isEdit.value || !originalFormSnapshot.value || hasSemanticChange(originalFormSnapshot.value, form)
  if (form.monitorType === 'field_transition' && semanticChange && validatedFingerprint.value !== getDetectionFingerprint(form)) {
    const valid = await runValidation()
    if (!valid) {
      submitError.value = '价格监控必须先通过配置验证才能保存'
      return
    }
  }
  submitError.value = null
  submitting.value = true
  try {
    const payload = toMonitorRequest(form)
    if (isEdit.value) {
      await updateMonitor(route.params.name, payload)
    } else {
      await createMonitor(payload)
    }
    router.push('/')
  } catch (e) {
    submitError.value = e.response?.data?.message || e.message
  } finally {
    submitting.value = false
  }
}

async function handleSaveAsRule() {
  if (!form.extraction.containerSelector) return
  const name = prompt('规则名称（如 澎湃快讯时间线）', form.basic.name.trim() + ' 规则')
  if (!name) return
  try {
    await createScanRule({
      name,
      url_contains: new URL(form.basic.url).hostname,
      container: form.extraction.containerSelector,
      item: form.extraction.itemSelector,
      priority: 50,
      enabled: true,
      description: '从表单保存',
      fields: form.extraction.fields.filter(f => f.name).map(f => ({
        name: f.name, selector: f.selector, type: f.type, attr: f.attr || '', transform: f.transform || '',
      })),
    })
    showSuccess('规则已保存')
  } catch (e) {
    showError('保存规则失败: ' + (e.response?.data?.message || e.message))
  }
}
</script>

<style scoped>
.page-header { margin-bottom: 1.5rem; }
.page-header h1 { font-size: 1.5rem; font-weight: 700; color: var(--text); margin-top: 0.5rem; }

.back-btn {
  display: inline-flex; align-items: center; gap: 0.3rem;
  padding: 0.35rem 0.85rem; border-radius: var(--radius-pill);
  font-size: 0.8125rem; font-weight: 700; color: var(--text-secondary);
  background: var(--bg-elevated); text-decoration: none;
  transition: var(--transition); margin-bottom: 0.5rem;
}
.back-btn:hover { background: var(--bg-hover); color: var(--text); }

.preview-panel {
  background: var(--bg-surface);
  border-radius: var(--radius-lg);
  padding: 1rem;
  margin-bottom: 1rem;
}
.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 0.75rem;
  padding-bottom: 0.5rem;
  border-bottom: 1px solid var(--border-light);
}
.section-header h2 { font-size: 0.9375rem; font-weight: 700; color: var(--text); }
.preview-input-row { display: flex; gap: 0.5rem; }
.preview-input-row .form-input { flex: 1; }
.preview-results { display: flex; flex-direction: column; gap: 0.5rem; }
.preview-card { background: var(--bg-card); border-radius: var(--radius-lg); padding: 0.75rem; }
.preview-card { border: 1px solid transparent; cursor: pointer; transition: var(--transition); }
.preview-card:hover, .preview-card.selected { border-color: var(--green); }
.preview-card-header { display: flex; align-items: center; gap: 0.5rem; margin-bottom: 0.5rem; }
.apply-candidate { margin-left: auto; }
.candidate-selectors { font-size: 0.75rem; color: var(--text-muted); margin-bottom: 0.5rem; word-break: break-all; }
.candidate-selectors code { color: var(--green); }
.candidate-badge { font-size: 0.6875rem; font-weight: 700; color: var(--text); background: var(--bg-elevated); padding: 0.15rem 0.5rem; border-radius: var(--radius-pill); }
.candidate-count { font-size: 0.75rem; color: var(--text-secondary); }
.preview-samples { display: flex; flex-direction: column; gap: 0.2rem; }
.preview-samples .sample-item {
  display: flex; padding: 0.2rem 0.5rem; border-radius: 4px; background: var(--bg-elevated); font-size: 0.8125rem;
}
</style>
