<template>
  <div class="add-monitor">
    <div class="page-header">
      <div>
        <router-link to="/" class="back-btn">← 返回</router-link>
        <h1>{{ isEdit ? '编辑监控器' : '新增监控器' }}</h1>
      </div>
    </div>

    <!-- ===== 编辑模式 ===== -->
    <template v-if="isEdit">
      <div class="loading" v-if="loading">
        <div class="spinner" />
        <p>加载配置...</p>
      </div>

      <template v-else>
        <div class="settings-section">
          <div class="section-header">
            <h2>基础配置</h2>
          </div>

          <div class="form-group">
            <label>名称</label>
            <input v-model="form.name" class="form-input" placeholder="如 招录公告" />
          </div>

          <div class="form-group">
            <label>URL</label>
            <input v-model="form.url" class="form-input" placeholder="https://example.com/zlgg/" />
          </div>

          <div class="form-group">
            <label>分组</label>
            <input v-model="form.group" class="form-input" placeholder="默认" />
          </div>

          <div class="form-group">
            <label>检查间隔（秒）</label>
            <input v-model.number="form.check_interval" class="form-input" type="number" min="10" placeholder="3600（默认1小时）" />
          </div>
        </div>

        <div class="settings-section">
          <div class="section-header">
            <h2>提取配置</h2>
            <div class="section-actions">
              <button class="btn btn-sm btn-ghost" @click="togglePreview" :disabled="!form.url.trim()">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="14" height="14"><circle cx="11" cy="11" r="8"/><line x1="21" y1="21" x2="16.65" y2="16.65"/></svg>
                预览抓取
              </button>
            </div>
          </div>

          <div class="preview-panel" v-if="showPreview">
            <div class="form-group">
              <label>关键词（辅助验证，多个用逗号隔开）</label>
              <div class="preview-input-row">
                <input v-model="previewKeyword" class="form-input" placeholder="面试,录用,公示" @keyup.enter="runPreview" />
                <button class="btn btn-sm btn-primary" :disabled="previewLoading" @click="runPreview">{{ previewLoading ? '扫描中' : '扫描' }}</button>
              </div>
            </div>

            <div class="loading" v-if="previewLoading"><div class="spinner" /><p>扫描中...</p></div>

            <div class="preview-results" v-else-if="previewData && previewData.containers && previewData.containers.length > 0">
              <div v-for="(c, ci) in previewData.containers" :key="ci" class="preview-card">
                <div class="preview-card-header">
                  <span class="candidate-badge">{{ c.container_tag.toUpperCase() }}</span>
                  <span class="candidate-count">{{ c.item_count }} 条</span>
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
          </div>

          <div class="form-group">
            <label>容器选择器</label>
            <input v-model="form.container" class="form-input" placeholder="如 div.hap_infoBox" />
          </div>

          <div class="form-group">
            <label>列表项选择器（可选）</label>
            <input v-model="form.item" class="form-input" placeholder="如 div.hap_infoOne" />
          </div>

          <FieldEditor v-model="form.fields" />

          <div class="form-group">
            <label class="checkbox-label">
              <input v-model="form.is_active" type="checkbox" />
              保存后立即启动监控
            </label>
          </div>

          <div class="form-error" v-if="submitError">{{ submitError }}</div>

          <div class="form-actions">
            <router-link to="/" class="btn btn-ghost">取消</router-link>
            <button class="btn btn-primary" :disabled="submitting" @click="handleSubmit">
              {{ submitting ? '提交中...' : '保存修改' }}
            </button>
          </div>
        </div>
      </template>
    </template>

    <!-- ===== 新增模式 ===== -->
    <template v-else>
      <div class="mode-tabs">
        <button class="mode-tab" :class="{ active: mode === 'quick' }" @click="mode = 'quick'">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="16" height="16"><circle cx="11" cy="11" r="8"/><line x1="21" y1="21" x2="16.65" y2="16.65"/></svg>
          快速模式
        </button>
        <button class="mode-tab" :class="{ active: mode === 'advanced' }" @click="mode = 'advanced'">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="16" height="16"><path d="M12 15V3m0 12l-4-4m4 4l4-4M2 17l.621 2.485A2 2 0 004.561 21h14.878a2 2 0 001.94-1.515L22 17"/></svg>
          高级模式
        </button>
      </div>

      <!-- 快速模式 -->
      <template v-if="mode === 'quick'">
        <div class="settings-section" v-if="step === 1">
          <div class="section-header">
            <h2>输入网址和关键词</h2>
            <p class="section-desc">告诉系统要监控哪个网页，以及关注什么内容</p>
          </div>
          <div class="form-group">
            <label>网页 URL</label>
            <input v-model="quickForm.url" class="form-input" placeholder="https://example.com/announce/" />
          </div>
          <div class="form-group">
            <label>关键词（多个用逗号隔开）</label>
            <input v-model="quickForm.keywords" class="form-input" placeholder="面试,录用,公示" />
            <p class="hint">系统会根据关键词自动定位网页中的列表区域</p>
          </div>
          <div class="form-error" v-if="scanError">{{ scanError }}</div>
          <div class="form-actions">
            <router-link to="/" class="btn btn-ghost">取消</router-link>
            <button class="btn btn-primary" :disabled="scanning" @click="handleScan">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="16" height="16"><circle cx="11" cy="11" r="8"/><line x1="21" y1="21" x2="16.65" y2="16.65"/></svg>
              {{ scanning ? '扫描中...' : '预扫描' }}
            </button>
          </div>
        </div>

        <div class="settings-section" v-if="step === 2">
          <div class="step-indicator">
            <span class="step-dot done">①</span><span class="step-line"/><span class="step-dot active">②</span><span class="step-line"/><span class="step-dot">③</span>
          </div>
          <div class="section-header">
            <h2>选择抓取区域</h2>
            <p class="section-desc">系统发现 {{ scanResult.containers.length }} 个数据区域，选择符合预期的一个</p>
          </div>
          <div v-for="(container, ci) in scanResult.containers" :key="ci" class="candidate-card" :class="{ selected: selectedContainer === ci }" @click="selectedContainer = ci">
            <div class="candidate-header">
              <div class="candidate-info">
                <span class="candidate-badge">{{ container.container_tag.toUpperCase() }}</span>
                <span class="candidate-count">{{ container.item_count }} 条内容</span>
                <span class="candidate-hit">命中 {{ container.keyword_hits }} 个关键词</span>
              </div>
              <div class="candidate-check" v-if="selectedContainer === ci">
                <svg viewBox="0 0 24 24" fill="currentColor" width="20" height="20"><path d="M9 16.17L4.83 12l-1.42 1.41L9 19 21 7l-1.41-1.41z"/></svg>
              </div>
            </div>
            <div class="candidate-selector"><code>{{ container.container_css }}</code></div>
            <div class="sample-list">
              <div v-for="(item, ii) in container.sample_items" :key="ii" class="sample-item">
                <span class="sample-title">{{ item.title }}</span>
                <span class="sample-meta" v-if="item.date">{{ item.date }}</span>
              </div>
              <div class="sample-more" v-if="container.item_count > container.sample_items.length">...还有 {{ container.item_count - container.sample_items.length }} 条</div>
            </div>
          </div>
          <div class="empty" v-if="scanResult.containers.length === 0">
            <p>未找到匹配的内容区域</p>
            <p style="color: var(--text-muted); font-size: 0.8125rem; margin-top: 0.25rem;">试试不同的关键词，或切换到高级模式手动配置</p>
          </div>
          <div class="form-actions" v-if="scanResult.containers.length > 0">
            <button class="btn btn-ghost" @click="step = 1">重新扫描</button>
            <button class="btn btn-primary" @click="step = 3">下一步 →</button>
          </div>
        </div>

        <div class="settings-section" v-if="step === 3">
          <div class="step-indicator">
            <span class="step-dot done">①</span><span class="step-line"/><span class="step-dot done">②</span><span class="step-line"/><span class="step-dot active">③</span>
          </div>
          <div class="section-header">
            <h2>确认并保存</h2>
            <p class="section-desc">为这个监控器起个名字</p>
          </div>
          <div class="form-group">
            <label>监控器名称</label>
            <input v-model="quickForm.name" class="form-input" placeholder="如 招录公告" />
          </div>
          <div class="summary-card">
            <div class="summary-row"><span class="summary-label">URL</span><span class="summary-value">{{ quickForm.url }}</span></div>
            <div class="summary-row"><span class="summary-label">容器</span><code class="summary-code">{{ selectedContainerCss }}</code></div>
            <div class="summary-row"><span class="summary-label">关键词</span><span class="summary-value">{{ quickForm.keywords }}</span></div>
            <div class="summary-row"><span class="summary-label">预计条目</span><span class="summary-value">{{ selectedContainerCount }} 条</span></div>
          </div>
          <div class="form-error" v-if="createError">{{ createError }}</div>
          <div class="form-actions">
            <button class="btn btn-ghost" @click="step = 2">上一步</button>
            <button class="btn btn-primary" :disabled="creating" @click="handleCreate">{{ creating ? '创建中...' : '创建监控器' }}</button>
          </div>
        </div>
      </template>

      <!-- 高级模式 -->
      <template v-if="mode === 'advanced'">
        <div class="settings-section">
          <div class="section-header"><h2>基础配置</h2></div>
          <div class="form-group">
            <label>名称</label>
            <input v-model="form.name" class="form-input" placeholder="如 招录公告" />
          </div>
          <div class="form-group">
            <label>URL</label>
            <input v-model="form.url" class="form-input" placeholder="https://example.com/zlgg/" />
          </div>
          <div class="form-group">
            <label>分组</label>
            <input v-model="form.group" class="form-input" placeholder="默认" />
          </div>
          <div class="form-group">
            <label>检查间隔（秒）</label>
            <input v-model.number="form.check_interval" class="form-input" type="number" min="10" placeholder="3600（默认1小时）" />
          </div>
        </div>
        <div class="settings-section">
          <div class="section-header"><h2>提取配置</h2></div>
          <div class="form-group">
            <label>容器选择器</label>
            <input v-model="form.container" class="form-input" placeholder="如 div.hap_infoBox" />
          </div>
          <div class="form-group">
            <label>列表项选择器（可选）</label>
            <input v-model="form.item" class="form-input" placeholder="如 div.hap_infoOne" />
          </div>
          <FieldEditor v-model="form.fields" />
          <div class="form-group">
            <label class="checkbox-label">
              <input v-model="form.is_active" type="checkbox" />
              保存后立即启动监控
            </label>
          </div>
          <div class="form-error" v-if="submitError">{{ submitError }}</div>
          <div class="form-actions">
            <router-link to="/" class="btn btn-ghost">取消</router-link>
            <button class="btn btn-primary" :disabled="submitting" @click="handleSubmit">{{ submitting ? '提交中...' : '创建并启动' }}</button>
          </div>
        </div>
      </template>
    </template>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { createMonitor, updateMonitor, fetchMonitorConfig, previewScan, smartCreate } from '../api/monitors'
import FieldEditor from '../components/FieldEditor.vue'

const router = useRouter()
const route = useRoute()

const isEdit = computed(() => !!route.params.name)
const submitting = ref(false)
const submitError = ref(null)

// 高级模式表单（新增 + 编辑共用）
const loading = ref(false)
const form = ref({
  name: '', url: '', group: '', container: '', item: '',
  check_interval: 3600, is_active: true,
  fields: [{ name: 'title', selector: 'a', type: 'text', attr: '', transform: '' }],
})

// 编辑模式预览
const showPreview = ref(false)
const previewKeyword = ref('')
const previewLoading = ref(false)
const previewData = ref(null)
const previewScanned = ref(false)

function togglePreview() {
  showPreview.value = !showPreview.value
  if (!showPreview.value) {
    previewData.value = null
    previewScanned.value = false
  }
}

async function runPreview() {
  if (!form.value.url.trim()) return
  previewLoading.value = true
  previewScanned.value = false
  try {
    const res = await previewScan({
      url: form.value.url.trim(),
      keywords: previewKeyword.value || '公告',
    })
    if (res.code === 0 && res.data) {
      previewData.value = res.data
    }
  } catch { /* ignore */ }
  previewScanned.value = true
  previewLoading.value = false
}

// 加载编辑数据
onMounted(async () => {
  if (isEdit.value) {
    loading.value = true
    try {
      const res = await fetchMonitorConfig(route.params.name)
      if (res.code === 0 && res.data) {
        const d = res.data
        form.value.name = d.Name || ''
        form.value.url = d.URL || ''
        form.value.group = d.GroupName || ''
        form.value.container = d.Container || ''
        form.value.item = d.Item || ''
        form.value.check_interval = d.CheckInterval || 3600
        form.value.is_active = d.IsActive ?? true
        if (d.Fields && d.Fields.length > 0) {
          form.value.fields = d.Fields.map(f => ({
            name: f.Name || '', selector: f.Selector || '',
            type: f.Type || 'text', attr: f.Attr || '', transform: f.Transform || '',
          }))
        }
      }
    } catch (e) {
      submitError.value = '加载配置失败: ' + e.message
    } finally {
      loading.value = false
    }
  }
})

function validate() {
  if (!form.value.name.trim()) return '名称不能为空'
  if (!form.value.url.trim()) return 'URL不能为空'
  if (!form.value.container.trim()) return '容器选择器不能为空'
  for (const f of form.value.fields) {
    if (!f.name.trim()) return '字段名称不能为空'
    if (!f.selector.trim()) return '字段选择器不能为空'
  }
  return null
}

async function handleSubmit() {
  const err = validate()
  if (err) { submitError.value = err; return }
  submitError.value = null
  submitting.value = true
  try {
    const payload = {
      name: form.value.name.trim(), url: form.value.url.trim(),
      container: form.value.container.trim(), item: form.value.item.trim(),
      group: form.value.group.trim(), check_interval: form.value.check_interval || 3600,
      is_active: form.value.is_active,
      fields: form.value.fields.filter(f => f.name.trim() && f.selector.trim()),
    }
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

// ===== 新增模式 =====
const mode = ref('quick')
const step = ref(1)
const scanning = ref(false)
const creating = ref(false)
const scanError = ref('')
const createError = ref('')
const scanResult = ref({ containers: [] })
const selectedContainer = ref(null)

const quickForm = ref({ url: '', keywords: '', name: '' })

const selectedContainerCss = computed(() => {
  if (selectedContainer.value === null || !scanResult.value.containers[selectedContainer.value]) return ''
  return scanResult.value.containers[selectedContainer.value].container_css
})

const selectedContainerCount = computed(() => {
  if (selectedContainer.value === null || !scanResult.value.containers[selectedContainer.value]) return 0
  return scanResult.value.containers[selectedContainer.value].item_count
})

async function handleScan() {
  if (!quickForm.value.url.trim()) { scanError.value = '请输入 URL'; return }
  if (!quickForm.value.keywords.trim()) { scanError.value = '请输入至少一个关键词'; return }
  scanning.value = true
  scanError.value = ''
  selectedContainer.value = null
  try {
    const res = await previewScan({ url: quickForm.value.url.trim(), keywords: quickForm.value.keywords.trim() })
    if (res.code === 0 && res.data) {
      scanResult.value = res.data
      if (res.data.containers && res.data.containers.length > 0) {
        selectedContainer.value = 0
        step.value = 2
      } else {
        scanError.value = '未找到匹配的内容。试试不同的关键词，或切换到高级模式。'
      }
    } else {
      scanError.value = res.message || '扫描失败'
    }
  } catch (e) {
    scanError.value = '扫描失败: ' + (e.response?.data?.message || e.message)
  } finally {
    scanning.value = false
  }
}

async function handleCreate() {
  if (!quickForm.value.name.trim()) { createError.value = '请输入监控器名称'; return }
  creating.value = true
  createError.value = ''
  try {
    const res = await smartCreate({ name: quickForm.value.name.trim(), url: quickForm.value.url.trim(), container_css: selectedContainerCss.value })
    if (res.code === 0) { router.push('/') }
    else { createError.value = res.message || '创建失败' }
  } catch (e) {
    createError.value = '创建失败: ' + (e.response?.data?.message || e.message)
  } finally {
    creating.value = false
  }
}
</script>

<style scoped>
.page-header { margin-bottom: 1.5rem; }
.page-header h1 { font-size: 1.5rem; font-weight: 700; color: var(--text); margin-top: 0.5rem; }

.section-header { display: flex; justify-content: space-between; align-items: center; }
.section-actions { flex-shrink: 0; }

/* Mode tabs */
.mode-tabs {
  display: flex; gap: 0.25rem; background: var(--bg-surface);
  border-radius: var(--radius-pill); padding: 3px; margin-bottom: 1.5rem; width: fit-content;
}
.mode-tab {
  display: inline-flex; align-items: center; gap: 0.4rem;
  padding: 0.5rem 1.25rem; border: none; border-radius: var(--radius-pill);
  background: transparent; color: var(--text-secondary);
  font-size: 0.8125rem; font-weight: 700; cursor: pointer; transition: var(--transition);
}
.mode-tab:hover { color: var(--text); }
.mode-tab.active { background: var(--green); color: #000000; }

/* Preview */
.preview-panel { background: var(--bg-base); border-radius: var(--radius-lg); padding: 1rem; margin-bottom: 1rem; }
.preview-input-row { display: flex; gap: 0.5rem; }
.preview-input-row .form-input { flex: 1; }
.preview-results { display: flex; flex-direction: column; gap: 0.5rem; }
.preview-card { background: var(--bg-surface); border-radius: var(--radius-lg); padding: 0.75rem; }
.preview-card-header { display: flex; align-items: center; gap: 0.5rem; margin-bottom: 0.5rem; }
.preview-samples { display: flex; flex-direction: column; gap: 0.2rem; }
.preview-samples .sample-item {
  display: flex; padding: 0.2rem 0.5rem; border-radius: 4px; background: var(--bg-elevated); font-size: 0.8125rem;
}

/* Step */
.step-indicator { display: flex; align-items: center; gap: 0.5rem; margin-bottom: 1.5rem; }
.step-dot { width: 28px; height: 28px; border-radius: var(--radius-circle); display: flex; align-items: center; justify-content: center; font-size: 0.75rem; font-weight: 700; background: var(--bg-elevated); color: var(--text-muted); flex-shrink: 0; }
.step-dot.done { background: var(--green); color: #000000; }
.step-dot.active { background: var(--bg-elevated); color: var(--text); box-shadow: 0 0 0 2px var(--green); }
.step-line { flex: 1; height: 2px; background: #333333; max-width: 60px; }

/* Candidate cards */
.candidate-card { background: var(--bg-card); border-radius: var(--radius-lg); padding: 1rem; margin-bottom: 0.75rem; cursor: pointer; transition: var(--transition); border: 2px solid transparent; }
.candidate-card:hover { background: var(--bg-hover); }
.candidate-card.selected { border-color: var(--green); background: rgba(30, 215, 96, 0.05); }
.candidate-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 0.5rem; }
.candidate-info { display: flex; align-items: center; gap: 0.5rem; }
.candidate-badge { font-size: 0.6875rem; font-weight: 700; color: var(--text); background: var(--bg-elevated); padding: 0.15rem 0.5rem; border-radius: var(--radius-pill); }
.candidate-count { font-size: 0.75rem; color: var(--text-secondary); }
.candidate-hit { font-size: 0.75rem; color: var(--green); }
.candidate-check { color: var(--green); }
.candidate-selector { margin-bottom: 0.75rem; }
.candidate-selector code { font-size: 0.75rem; color: var(--text-muted); background: var(--bg-elevated); padding: 0.15rem 0.4rem; border-radius: 4px; }
.sample-list { display: flex; flex-direction: column; gap: 0.25rem; }
.sample-item { display: flex; justify-content: space-between; padding: 0.3rem 0.5rem; border-radius: 4px; background: var(--bg-base); font-size: 0.8125rem; }
.sample-title { color: var(--text); flex: 1; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.sample-meta { color: var(--text-muted); font-size: 0.75rem; margin-left: 0.5rem; flex-shrink: 0; }
.sample-more { font-size: 0.75rem; color: var(--text-muted); text-align: center; padding: 0.3rem; }

/* Summary */
.summary-card { background: var(--bg-surface); border-radius: var(--radius-lg); padding: 1rem; margin-bottom: 1rem; }
.summary-row { display: flex; align-items: center; padding: 0.5rem 0; border-bottom: 1px solid #2a2a2a; gap: 1rem; }
.summary-row:last-child { border-bottom: none; }
.summary-label { font-size: 0.75rem; font-weight: 700; color: var(--text-muted); text-transform: uppercase; letter-spacing: 0.8px; min-width: 60px; flex-shrink: 0; }
.summary-value { font-size: 0.875rem; color: var(--text); word-break: break-all; }
.summary-code { font-size: 0.75rem; color: var(--green); background: var(--bg-elevated); padding: 0.15rem 0.4rem; border-radius: 4px; }

.hint { font-size: 0.75rem; color: var(--text-muted); margin-top: 0.25rem; }

/* 统一返回按钮 */
.back-btn {
  display: inline-flex; align-items: center; gap: 0.3rem;
  padding: 0.35rem 0.85rem; border-radius: var(--radius-pill);
  font-size: 0.8125rem; font-weight: 700; color: var(--text-secondary);
  background: var(--bg-elevated); text-decoration: none;
  transition: var(--transition); margin-bottom: 0.5rem;
}
.back-btn:hover { background: var(--bg-hover); color: var(--text); }
</style>