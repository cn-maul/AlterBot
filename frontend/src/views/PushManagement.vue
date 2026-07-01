<template>
  <div class="push-page">
    <div class="page-header">
      <div>
        <h1>推送管理</h1>
        <p class="page-desc">配置推送通知渠道，网页变更时自动发送通知</p>
      </div>
    </div>

    <div class="loading" v-if="loading">
      <div class="spinner" />
      <p>加载中...</p>
    </div>

    <template v-else>
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
            <select v-model="form.service" class="form-input" style="width: auto; min-width: 200px;">
              <option value="pushplus">PushPlus</option>
              <option value="webhook">Webhook 机器人</option>
            </select>
          </div>
        </div>

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

        <template v-if="form.service === 'webhook'">
          <div class="setting-item">
            <div class="setting-info">
              <div class="setting-label">Webhook URL</div>
              <div class="setting-desc">企业微信/飞书/钉钉机器人地址</div>
            </div>
            <div class="setting-control">
              <input v-model="form.config.url" class="form-input" style="min-width: 350px;" placeholder="https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=xxx" />
            </div>
          </div>
        </template>

        <div class="toast toast-error" v-if="saveError">{{ saveError }}</div>

        <div class="form-actions">
          <span class="save-hint" v-if="saved">已保存</span>
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
.save-hint {
  color: var(--green);
  font-size: 0.8125rem;
  font-weight: 600;
}
</style>