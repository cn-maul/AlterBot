<template>
  <div class="push-page">
    <div class="page-header">
      <h1>推送管理</h1>
      <p class="page-desc">配置推送通知渠道，网页变更时自动发送通知</p>
    </div>

    <!-- 加载态 -->
    <div class="loading" v-if="loading">
      <div class="spinner" />
      <p>加载中...</p>
    </div>

    <template v-else>
      <!-- 推送通知设置 -->
      <div class="settings-section">
        <div class="section-header">
          <h2>推送通知</h2>
          <p class="section-desc">发现网页变更时通过以下渠道发送通知</p>
        </div>

        <div class="setting-item">
          <div class="setting-info">
            <div class="setting-label">启用推送</div>
            <div class="setting-desc">开启后，检测到变更时会发送通知</div>
          </div>
          <div class="setting-control">
            <label class="toggle">
              <input type="checkbox" v-model="form.enabled" />
              <span class="toggle-track"></span>
            </label>
          </div>
        </div>

        <div class="setting-item">
          <div class="setting-info">
            <div class="setting-label">推送服务</div>
            <div class="setting-desc">选择通知渠道</div>
          </div>
          <div class="setting-control">
            <select v-model="form.service" class="form-input" style="width: auto;">
              <option value="pushplus">PushPlus</option>
              <option value="webhook">Webhook 机器人</option>
            </select>
          </div>
        </div>

        <!-- PushPlus 配置 -->
        <template v-if="form.service === 'pushplus'">
          <div class="setting-item">
            <div class="setting-info">
              <div class="setting-label">Token</div>
              <div class="setting-desc">PushPlus 用户令牌</div>
            </div>
            <div class="setting-control">
              <input v-model="form.config.token" class="form-input" style="width: 280px;" placeholder="输入 token" />
            </div>
          </div>
          <div class="setting-item">
            <div class="setting-info">
              <div class="setting-label">Channel</div>
              <div class="setting-desc">推送渠道（如 mail、wechat、sms）</div>
            </div>
            <div class="setting-control">
              <input v-model="form.config.channel" class="form-input" style="width: 200px;" placeholder="mail" />
            </div>
          </div>
        </template>

        <!-- Webhook 配置 -->
        <template v-if="form.service === 'webhook'">
          <div class="setting-item">
            <div class="setting-info">
              <div class="setting-label">Webhook URL</div>
              <div class="setting-desc">企业微信/飞书/钉钉机器人地址</div>
            </div>
            <div class="setting-control">
              <input v-model="form.config.url" class="form-input" style="width: 400px;" placeholder="https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=xxx" />
            </div>
          </div>
        </template>

        <!-- 错误提示 -->
        <div class="toast toast-error" v-if="saveError">{{ saveError }}</div>

        <!-- 操作按钮 -->
        <div class="form-actions">
          <span class="save-hint" v-if="saved">✅ 已保存</span>
          <button class="btn btn-primary" :disabled="saving" @click="handleSave">
            {{ saving ? '保存中...' : '保存设置' }}
          </button>
        </div>
      </div>
    </template>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { fetchNotificationSettings, updateNotificationSettings } from '../api/monitors'

const loading = ref(true)
const saving = ref(false)
const saved = ref(false)
const saveError = ref('')

const form = ref({
  enabled: false,
  service: 'pushplus',
  config: {
    token: '',
    channel: 'mail',
    url: '',
  },
})

onMounted(loadSettings)

async function loadSettings() {
  loading.value = true
  try {
    const res = await fetchNotificationSettings()
    if (res.code === 0 && res.data) {
      form.value.enabled = res.data.enabled || false
      form.value.service = res.data.service || 'pushplus'
      form.value.config = {
        token: res.data.config?.token || '',
        channel: res.data.config?.channel || 'mail',
        url: res.data.config?.url || '',
      }
    }
  } catch (e) {
    saveError.value = '加载设置失败: ' + (e.response?.data?.message || e.message)
  } finally {
    loading.value = false
  }
}

async function handleSave() {
  saving.value = true
  saved.value = false
  saveError.value = ''
  try {
    const payload = {
      enabled: form.value.enabled,
      service: form.value.service,
      config: form.value.config,
    }
    await updateNotificationSettings(payload)
    saved.value = true
    setTimeout(() => { saved.value = false }, 2000)
  } catch (e) {
    saveError.value = '保存失败: ' + (e.response?.data?.message || e.message)
  } finally {
    saving.value = false
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
  margin-bottom: 0.25rem;
}

.page-desc {
  color: var(--text-muted);
  font-size: 0.85rem;
}

.toast {
  padding: 0.6rem 1rem;
  border-radius: var(--radius);
  font-size: 0.85rem;
  margin-bottom: 1rem;
  animation: fadeIn 0.2s ease;
}

.toast-error {
  background: var(--error-bg);
  color: var(--error);
  border: 1px solid var(--error);
}

@keyframes fadeIn {
  from { opacity: 0; transform: translateY(-4px); }
  to { opacity: 1; transform: translateY(0); }
}

.form-actions {
  display: flex;
  gap: 0.5rem;
  justify-content: flex-end;
  margin-top: 1rem;
  padding-top: 1rem;
  border-top: 1px solid var(--border-light);
  align-items: center;
}

.toggle {
  display: inline-flex;
  align-items: center;
  cursor: pointer;
}

.toggle input {
  display: none;
}

.toggle-track {
  width: 44px;
  height: 24px;
  background: var(--border);
  border-radius: 12px;
  position: relative;
  transition: background 0.2s;
}

.toggle-track::after {
  content: '';
  position: absolute;
  top: 2px;
  left: 2px;
  width: 20px;
  height: 20px;
  background: white;
  border-radius: 50%;
  transition: transform 0.2s;
  box-shadow: 0 1px 3px rgba(0,0,0,0.2);
}

.toggle input:checked + .toggle-track {
  background: var(--primary);
}

.toggle input:checked + .toggle-track::after {
  transform: translateX(20px);
}

.save-hint {
  color: var(--success);
  font-size: 0.85rem;
  display: flex;
  align-items: center;
}
</style>