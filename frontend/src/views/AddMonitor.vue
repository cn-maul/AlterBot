<template>
  <div class="add-monitor">
    <div class="page-header">
      <div>
        <router-link to="/" class="back-link">← 返回</router-link>
        <h1>{{ isEdit ? '编辑监控器' : '新增监控器' }}</h1>
      </div>
    </div>

    <div class="loading" v-if="isEdit && loading">
      <div class="spinner" />
      <p>加载配置...</p>
    </div>

    <div class="settings-section" v-else>
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

    <div class="settings-section" v-if="!isEdit || !loading">
      <div class="section-header">
        <h2>提取配置</h2>
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
          {{ submitting ? '提交中...' : (isEdit ? '保存修改' : '创建并启动') }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { createMonitor, updateMonitor, fetchMonitorConfig } from '../api/monitors'
import FieldEditor from '../components/FieldEditor.vue'

const router = useRouter()
const route = useRoute()

const isEdit = computed(() => !!route.params.name)
const loading = ref(false)
const submitting = ref(false)
const submitError = ref(null)

const form = ref({
  name: '',
  url: '',
  group: '',
  container: '',
  item: '',
  check_interval: 3600,
  is_active: true,
  fields: [{ name: 'title', selector: 'a', type: 'text', attr: '', transform: '' }],
})

onMounted(async () => {
  if (isEdit.value) {
    loading.value = true
    try {
      const res = await fetchMonitorConfig(route.params.name)
      if (res.code === 0 && res.data) {
        const data = res.data
        form.value.name = data.Name || ''
        form.value.url = data.URL || ''
        form.value.group = data.GroupName || ''
        form.value.container = data.Container || ''
        form.value.item = data.Item || ''
        form.value.check_interval = data.CheckInterval || 3600
        form.value.is_active = data.IsActive ?? true
        if (data.Fields && data.Fields.length > 0) {
          form.value.fields = data.Fields.map(f => ({
            name: f.Name || '',
            selector: f.Selector || '',
            type: f.Type || 'text',
            attr: f.Attr || '',
            transform: f.Transform || '',
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
      name: form.value.name.trim(),
      url: form.value.url.trim(),
      container: form.value.container.trim(),
      item: form.value.item.trim(),
      group: form.value.group.trim(),
      check_interval: form.value.check_interval || 3600,
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
</script>

<style scoped>
.page-header {
  margin-bottom: 2rem;
}

.page-header h1 {
  font-size: 1.5rem;
  font-weight: 700;
  color: var(--text);
  margin-top: 0.5rem;
}

.page-desc {
  color: var(--text-secondary);
  font-size: 0.8125rem;
}
</style>