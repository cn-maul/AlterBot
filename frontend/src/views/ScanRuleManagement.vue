<template>
  <div class="page scan-rules-page">
    <div v-if="successMsg" class="toast toast-success">{{ successMsg }}</div>
    <div v-if="pageErrorMsg" class="toast toast-error">{{ pageErrorMsg }}</div>

    <header class="page-header">
      <div>
        <h1>扫描规则</h1>
        <p>预扫描并保存可复用的网页提取规则</p>
      </div>
      <div class="header-actions">
        <input ref="fileInput" class="file-input" type="file" accept="application/json,.json" @change="handleImportFile" />
        <button class="btn btn-ghost btn-sm" :disabled="importing" @click="fileInput?.click()">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" aria-hidden="true"><path d="M12 3v12"/><path d="m7 10 5 5 5-5"/><path d="M5 21h14"/></svg>
          {{ importing ? '导入中...' : '导入' }}
        </button>
        <button class="btn btn-ghost btn-sm" :disabled="exporting || rules.length === 0" @click="handleExport">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" aria-hidden="true"><path d="M12 21V9"/><path d="m17 14-5-5-5 5"/><path d="M5 3h14"/></svg>
          {{ exporting ? '导出中...' : '导出' }}
        </button>
      </div>
    </header>

    <section class="builder-section" aria-labelledby="quick-rule-title">
      <div class="section-title-row">
        <h2 id="quick-rule-title">快速保存</h2>
        <span v-if="scanResult" class="result-count">{{ candidates.length }} 个候选</span>
      </div>

      <div class="scan-form">
        <div class="form-group url-field">
          <label for="rule-url">URL</label>
          <input id="rule-url" v-model="url" class="form-input" placeholder="https://example.com/announcements/" @keyup.enter="handleScan" />
        </div>
        <div class="form-group keyword-field">
          <label for="rule-keywords">关键词</label>
          <input id="rule-keywords" v-model="keywords" class="form-input" placeholder="公告, 招聘, 公示" @keyup.enter="handleScan" />
        </div>
        <button class="btn btn-primary scan-button" :disabled="!url.trim() || scanning" @click="handleScan">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" aria-hidden="true"><circle cx="11" cy="11" r="7"/><path d="m20 20-3.5-3.5"/></svg>
          {{ scanning ? '扫描中...' : '预扫描' }}
        </button>
      </div>

      <div v-if="scanError" class="inline-error">{{ scanError }}</div>
      <div v-if="scanning" class="scan-loading"><div class="spinner" /><span>正在扫描网页</span></div>

      <div v-else-if="candidates.length" class="candidate-list">
        <button
          v-for="(candidate, index) in candidates"
          :key="candidateKey(candidate, index)"
          type="button"
          class="candidate-row"
          :class="{ selected: selectedIndex === index }"
          @click="selectedIndex = index"
        >
          <span class="choice-mark" aria-hidden="true">
            <svg v-if="selectedIndex === index" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3"><path d="m5 12 4 4L19 6"/></svg>
          </span>
          <span class="candidate-content">
            <span class="candidate-heading">
              <strong>候选 {{ index + 1 }}</strong>
              <span>{{ candidate.item_count }} 条</span>
              <span v-if="candidate.keyword_hits">关键词命中 {{ candidate.keyword_hits }}</span>
              <span class="candidate-strategy" :class="strategyClass(candidate.strategy)" v-if="candidate.strategy">{{ strategyLabel(candidate.strategy) }}</span>
            </span>
            <span class="sample-list">
              <span v-for="(item, itemIndex) in (candidate.sample_items || []).slice(0, 4)" :key="itemIndex" class="sample-line">
                <span class="sample-title">{{ item.title || item.url || '未命名内容' }}</span>
                <span v-if="item.date" class="sample-date">{{ item.date }}</span>
              </span>
            </span>
            <span class="selector-line">
              <code>{{ candidate.config?.container }}</code>
              <span>/</span>
              <code>{{ candidate.config?.item }}</code>
            </span>
          </span>
        </button>
      </div>

      <div v-else-if="scanned" class="empty-result">没有找到可保存的内容区域</div>

      <div v-if="selectedCandidate" class="save-panel">
        <div class="form-group name-field">
          <label for="rule-name">规则名称</label>
          <input id="rule-name" v-model="ruleName" class="form-input" placeholder="例如：殷都区招聘公告列表" @keyup.enter="handleSave" />
        </div>
        <div class="scope-field">
          <span class="field-label">适用范围</span>
          <div class="scope-control">
            <button type="button" :class="{ active: scopeType === 'exact' }" @click="scopeType = 'exact'">当前页面</button>
            <button type="button" :disabled="!routeScopeAvailable" :class="{ active: scopeType === 'route' }" title="匹配当前路径及其子路径" @click="scopeType = 'route'">当前路由</button>
            <button type="button" :class="{ active: scopeType === 'global' }" title="跨网站按相同页面结构匹配" @click="scopeType = 'global'">通用结构</button>
          </div>
        </div>
        <div class="scope-summary">{{ scopeSummary }}</div>
        <button class="btn btn-primary save-button" :disabled="saving || !ruleName.trim()" @click="handleSave">
          {{ saving ? '保存中...' : '保存规则' }}
        </button>
      </div>
    </section>

    <section class="library-section" aria-labelledby="rule-library-title">
      <div class="section-title-row library-title">
        <h2 id="rule-library-title">已保存规则</h2>
        <span>{{ rules.length }}</span>
      </div>

      <div v-if="loading" class="list-state">正在加载规则</div>
      <div v-else-if="rules.length === 0" class="list-state">暂无已保存规则</div>
      <div v-else class="rule-list">
        <article v-for="rule in rules" :key="rule.id" class="rule-row">
          <div class="rule-main">
            <div class="rule-heading">
              <strong>{{ rule.name }}</strong>
              <span class="scope-badge" :class="`scope-${rule.scope_type || 'legacy'}`">{{ scopeName(rule) }}</span>
              <span v-if="!rule.enabled" class="disabled-badge">已禁用</span>
            </div>
            <div class="rule-target">{{ scopeTarget(rule) }}</div>
            <div class="rule-structure">
              <code>{{ rule.container }}</code>
              <span>/</span>
              <code>{{ rule.item }}</code>
              <span class="field-count">{{ (rule.fields || []).length }} 个字段</span>
            </div>
          </div>
          <button class="icon-button danger" title="删除规则" @click="handleDelete(rule)">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" aria-hidden="true"><path d="M3 6h18"/><path d="M8 6V4h8v2"/><path d="m19 6-1 14H6L5 6"/><path d="M10 11v5M14 11v5"/></svg>
          </button>
        </article>
      </div>
    </section>
  </div>
</template>

<script setup>
import { computed, onMounted, ref, watch } from 'vue'
import { deleteScanRule, exportScanRules, fetchScanRules, importScanRules, previewScan, quickCreateScanRule } from '../api/monitors'
import { useToastMessages } from '../composables/useToastMessages'

const { successMsg, pageErrorMsg, showSuccess, showError } = useToastMessages()

const loading = ref(true)
const rules = ref([])
const url = ref('')
const keywords = ref('')
const scanning = ref(false)
const scanned = ref(false)
const scanError = ref('')
const scanResult = ref(null)
const selectedIndex = ref(null)
const ruleName = ref('')
const scopeType = ref('exact')
const saving = ref(false)
const importing = ref(false)
const exporting = ref(false)
const fileInput = ref(null)

const candidates = computed(() => scanResult.value?.containers || [])
const selectedCandidate = computed(() => selectedIndex.value === null ? null : candidates.value[selectedIndex.value])
const parsedURL = computed(() => {
  try { return new URL(url.value.trim()) } catch { return null }
})
const routeScopeAvailable = computed(() => Boolean(parsedURL.value && (parsedURL.value.pathname !== '/' || parsedURL.value.search)))
const scopeSummary = computed(() => {
  if (scopeType.value === 'global') return '所有网站中结构相同的页面'
  if (!parsedURL.value) return ''
  if (scopeType.value === 'route') return `${parsedURL.value.host}${parsedURL.value.pathname}${parsedURL.value.search}`
  return parsedURL.value.href
})

watch([url, keywords], () => {
  scanResult.value = null
  selectedIndex.value = null
  scanned.value = false
  scanError.value = ''
  ruleName.value = ''
  scopeType.value = 'exact'
})

onMounted(loadRules)

async function loadRules() {
  loading.value = true
  try {
    const response = await fetchScanRules()
    rules.value = response.code === 0 ? (response.data || []) : []
  } catch (error) {
    showError('加载规则失败: ' + errorMessage(error))
  } finally {
    loading.value = false
  }
}

async function handleScan() {
  if (!url.value.trim()) return
  scanning.value = true
  scanned.value = false
  scanError.value = ''
  scanResult.value = null
  selectedIndex.value = null
  try {
    const response = await previewScan({ url: url.value.trim(), keywords: keywords.value.trim() })
    if (response.code === 0) scanResult.value = response.data
    else scanError.value = response.message || '扫描失败'
  } catch (error) {
    scanError.value = errorMessage(error)
  } finally {
    scanning.value = false
    scanned.value = true
  }
}

async function handleSave() {
  if (!selectedCandidate.value || !ruleName.value.trim()) return
  saving.value = true
  try {
    await quickCreateScanRule({
      name: ruleName.value.trim(),
      url: url.value.trim(),
      keywords: keywords.value.trim(),
      scope_type: scopeType.value,
      config: selectedCandidate.value.config,
    })
    showSuccess('规则已保存')
    resetBuilder()
    await loadRules()
  } catch (error) {
    showError('保存规则失败: ' + errorMessage(error))
  } finally {
    saving.value = false
  }
}

async function handleDelete(rule) {
  if (!window.confirm(`确定删除规则「${rule.name}」吗？`)) return
  try {
    await deleteScanRule(rule.id)
    rules.value = rules.value.filter(item => item.id !== rule.id)
    showSuccess('规则已删除')
  } catch (error) {
    showError('删除规则失败: ' + errorMessage(error))
  }
}

async function handleExport() {
  exporting.value = true
  try {
    const response = await exportScanRules()
    if (response.code !== 0 || !response.data) throw new Error(response.message || '导出失败')
    const blob = new Blob([JSON.stringify(response.data, null, 2)], { type: 'application/json;charset=utf-8' })
    const link = document.createElement('a')
    link.href = URL.createObjectURL(blob)
    link.download = `gentry-scan-rules-${new Date().toISOString().slice(0, 10)}.json`
    link.click()
    URL.revokeObjectURL(link.href)
    showSuccess(`已导出 ${rules.value.length} 条规则`)
  } catch (error) {
    showError('导出规则失败: ' + errorMessage(error))
  } finally {
    exporting.value = false
  }
}

async function handleImportFile(event) {
  const file = event.target.files?.[0]
  if (!file) return
  importing.value = true
  try {
    const document = JSON.parse(await file.text())
    const response = await importScanRules(document)
    if (response.code !== 0) throw new Error(response.message || '导入失败')
    const imported = response.data?.imported || 0
    const skipped = response.data?.skipped || 0
    showSuccess(`已导入 ${imported} 条${skipped ? `，跳过 ${skipped} 条同名规则` : ''}`)
    await loadRules()
  } catch (error) {
    showError('导入规则失败: ' + errorMessage(error))
  } finally {
    importing.value = false
    event.target.value = ''
  }
}

function resetBuilder() {
  url.value = ''
  keywords.value = ''
  scanResult.value = null
  selectedIndex.value = null
  scanned.value = false
  ruleName.value = ''
  scopeType.value = 'exact'
}

function candidateKey(candidate, index) {
  return `${candidate.config?.container || ''}:${candidate.config?.item || ''}:${index}`
}

function scopeName(rule) {
  if (rule.scope_type === 'exact') return '页面'
  if (rule.scope_type === 'route') return '路由'
  if (rule.scope_type === 'global') return '通用'
  return '旧版'
}

function scopeTarget(rule) {
  if (rule.scope_type === 'global') return '所有网站中结构相同的页面'
  return rule.source_url || `URL 包含 ${rule.url_contains}`
}

function strategyLabel(strategy) {
  if (!strategy) return ''
  if (strategy.startsWith('template_')) return `规则「${strategy.slice(9)}」`
  const labels = {
    keyword_ancestor: '关键词定位',
    repeated_list: '重复列表',
    link_cluster: '链接簇',
    table_rows: '表格检测',
  }
  return labels[strategy] || strategy
}

function strategyClass(strategy) {
  return strategy?.startsWith('template_') ? 'strategy-rule' : 'strategy-heuristic'
}

function errorMessage(error) {
  return error.response?.data?.message || error.message || '操作失败'
}
</script>

<style scoped>
.scan-rules-page { max-width: 1120px; }
.page-header { display: flex; align-items: flex-start; justify-content: space-between; gap: 1rem; margin-bottom: 1.75rem; }
.page-header h1 { margin: 0; color: var(--text); font-size: 1.5rem; font-weight: 700; }
.page-header p { margin-top: 0.3rem; color: var(--text-secondary); font-size: 0.8125rem; }
.header-actions { display: flex; gap: 0.5rem; flex-shrink: 0; }
.header-actions svg, .scan-button svg { width: 16px; height: 16px; }
.file-input { display: none; }

.builder-section, .library-section { padding: 0 0 1.75rem; }
.builder-section { border-bottom: 1px solid var(--border); }
.library-section { padding-top: 1.75rem; }
.section-title-row { display: flex; align-items: center; justify-content: space-between; gap: 1rem; margin-bottom: 1rem; }
.section-title-row h2 { margin: 0; color: var(--text); font-size: 1rem; font-weight: 700; }
.result-count, .library-title > span { color: var(--text-muted); font-size: 0.75rem; }

.scan-form { display: grid; grid-template-columns: minmax(280px, 2fr) minmax(200px, 1fr) auto; align-items: end; gap: 0.75rem; }
.scan-form .form-group { min-width: 0; margin: 0; }
.scan-button { height: 38px; padding-inline: 1.2rem; }
.inline-error { margin-top: 0.75rem; color: var(--error); font-size: 0.8125rem; }
.scan-loading { display: flex; align-items: center; justify-content: center; gap: 0.75rem; min-height: 120px; color: var(--text-secondary); font-size: 0.8125rem; }
.scan-loading .spinner { width: 22px; height: 22px; margin: 0; border-width: 2px; }
.empty-result { margin-top: 1rem; padding: 1.5rem 0; border-top: 1px solid var(--border-light); color: var(--text-secondary); text-align: center; font-size: 0.8125rem; }

.candidate-list { display: grid; gap: 0.6rem; margin-top: 1rem; }
.candidate-row { display: grid; grid-template-columns: 24px minmax(0, 1fr); gap: 0.75rem; width: 100%; padding: 0.85rem; border: 1px solid var(--border); border-radius: 6px; background: var(--bg-surface); color: var(--text); text-align: left; cursor: pointer; transition: var(--transition); }
.candidate-row:hover { border-color: var(--text-muted); background: var(--bg-hover); }
.candidate-row.selected { border-color: var(--green); box-shadow: 0 0 0 1px var(--green) inset; }
.choice-mark { display: inline-flex; align-items: center; justify-content: center; width: 20px; height: 20px; margin-top: 1px; border: 1px solid var(--border); border-radius: 50%; color: #000; background: var(--bg-elevated); }
.candidate-row.selected .choice-mark { border-color: var(--green); background: var(--green); }
.choice-mark svg { width: 13px; height: 13px; }
.candidate-content { min-width: 0; }
.candidate-heading { display: flex; align-items: center; gap: 0.65rem; margin-bottom: 0.55rem; font-size: 0.75rem; color: var(--text-secondary); }
.candidate-heading strong { color: var(--text); font-size: 0.875rem; }
.candidate-strategy { font-size: 0.6875rem; padding: 0.1rem 0.5rem; border-radius: var(--radius-pill); white-space: nowrap; }
.candidate-strategy.strategy-rule { color: #fff; background: #e74c3c; font-weight: 700; }
.candidate-strategy.strategy-heuristic { color: var(--accent); background: var(--bg-elevated); }
.sample-list { display: grid; gap: 0.2rem; }
.sample-line { display: flex; justify-content: space-between; gap: 1rem; min-width: 0; padding: 0.28rem 0.5rem; border-radius: 4px; background: var(--bg-elevated); font-size: 0.8125rem; }
.sample-title { min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.sample-date { flex-shrink: 0; color: var(--text-muted); font-size: 0.75rem; }
.selector-line { display: flex; gap: 0.4rem; min-width: 0; margin-top: 0.55rem; color: var(--text-muted); font-size: 0.6875rem; }
.selector-line code { overflow: hidden; color: var(--text-secondary); text-overflow: ellipsis; white-space: nowrap; }

.save-panel { display: grid; grid-template-columns: minmax(220px, 1fr) auto minmax(180px, 1fr) auto; align-items: end; gap: 0.75rem; margin-top: 1rem; padding-top: 1rem; border-top: 1px solid var(--border-light); }
.save-panel .form-group { margin: 0; }
.field-label { display: block; margin-bottom: 0.35rem; color: var(--text-secondary); font-size: 0.75rem; font-weight: 700; text-transform: uppercase; letter-spacing: 1px; }
.scope-control { display: grid; grid-template-columns: repeat(3, auto); overflow: hidden; border: 1px solid var(--border); border-radius: 6px; }
.scope-control button { min-height: 38px; padding: 0 0.8rem; border: 0; border-right: 1px solid var(--border); background: var(--bg-elevated); color: var(--text-secondary); cursor: pointer; font-size: 0.75rem; white-space: nowrap; }
.scope-control button:last-child { border-right: 0; }
.scope-control button.active { background: var(--green); color: #000; font-weight: 700; }
.scope-control button:disabled { opacity: 0.4; cursor: not-allowed; }
.scope-summary { align-self: center; min-width: 0; overflow: hidden; color: var(--text-muted); font-size: 0.75rem; text-overflow: ellipsis; white-space: nowrap; }
.save-button { height: 38px; }

.rule-list { border-top: 1px solid var(--border); }
.rule-row { display: flex; align-items: center; justify-content: space-between; gap: 1rem; padding: 1rem 0; border-bottom: 1px solid var(--border-light); }
.rule-main { min-width: 0; }
.rule-heading { display: flex; align-items: center; flex-wrap: wrap; gap: 0.5rem; }
.rule-heading strong { font-size: 0.875rem; }
.scope-badge, .disabled-badge { padding: 0.14rem 0.45rem; border-radius: 4px; font-size: 0.6875rem; font-weight: 700; }
.scope-badge { background: var(--bg-elevated); color: var(--text-secondary); }
.scope-global { color: var(--green); }
.disabled-badge { color: var(--warning); background: var(--warning-bg); }
.rule-target { margin-top: 0.35rem; overflow: hidden; color: var(--text-secondary); font-size: 0.8125rem; text-overflow: ellipsis; white-space: nowrap; }
.rule-structure { display: flex; gap: 0.4rem; min-width: 0; margin-top: 0.35rem; color: var(--text-muted); font-size: 0.6875rem; }
.rule-structure code { max-width: 280px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.field-count { flex-shrink: 0; margin-left: 0.35rem; }
.icon-button { display: inline-flex; align-items: center; justify-content: center; width: 34px; height: 34px; flex: 0 0 34px; border: 0; border-radius: 50%; background: transparent; color: var(--text-muted); cursor: pointer; }
.icon-button:hover { background: var(--error-bg); color: var(--error); }
.icon-button svg { width: 17px; height: 17px; }
.list-state { padding: 2.5rem 0; color: var(--text-secondary); text-align: center; font-size: 0.8125rem; }

@media (max-width: 900px) {
  .scan-form { grid-template-columns: 1fr 1fr; }
  .scan-button { grid-column: 1 / -1; justify-self: start; }
  .save-panel { grid-template-columns: 1fr 1fr; }
  .scope-summary { order: 3; }
  .save-button { order: 4; justify-self: end; }
}

@media (max-width: 640px) {
  .page-header { align-items: stretch; flex-direction: column; }
  .header-actions { align-self: flex-start; }
  .scan-form, .save-panel { grid-template-columns: 1fr; }
  .scan-button, .save-button { width: 100%; justify-self: stretch; }
  .scope-control { grid-template-columns: repeat(3, minmax(0, 1fr)); }
  .scope-control button { padding-inline: 0.35rem; white-space: normal; }
  .scope-summary, .save-button { order: initial; }
  .candidate-heading { align-items: flex-start; flex-wrap: wrap; }
  .sample-date { display: none; }
  .rule-structure code { max-width: 120px; }
}
</style>
