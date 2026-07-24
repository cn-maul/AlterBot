<template>
  <div class="update-panel">
    <button class="version-btn" :disabled="checking" @click="handleClick">
      <span v-if="state === 'idle'">{{ version }}</span>
      <span v-else-if="state === 'checking'">检查中...</span>
      <span v-else-if="state === 'uptodate'">已是最新</span>
      <span v-else-if="state === 'error'">检查失败</span>
    </button>

    <div class="update-actions" v-if="state === 'available'">
      <button class="update-btn" :disabled="updating" @click="handleUpdate">
        {{ updating ? '更新中...' : '升级到 ' + latestVersion }}
      </button>
      <div class="progress-bar" v-if="updating"><div class="progress-fill fill-anim" /></div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { fetchVersion, checkUpdate, applyUpdate } from '../api/monitors'

const version = ref('')
const state = ref('idle') // idle | checking | uptodate | available | error
const latestVersion = ref('')
const downloadURL = ref('')
const updating = ref(false)

onMounted(async () => {
  try {
    const res = await fetchVersion()
    if (res.code === 0 && res.data) {
      version.value = res.data.version
    }
  } catch {}
})

async function handleClick() {
  state.value = 'checking'
  try {
    const res = await checkUpdate()
    if (res.code === 0 && res.data) {
      if (res.data.has_update) {
        latestVersion.value = res.data.latest_version
        downloadURL.value = res.data.download_url
        state.value = 'available'
      } else {
        state.value = 'uptodate'
        setTimeout(() => { state.value = 'idle' }, 2000)
      }
    } else {
      state.value = 'error'
      setTimeout(() => { state.value = 'idle' }, 2000)
    }
  } catch {
    state.value = 'error'
    setTimeout(() => { state.value = 'idle' }, 2000)
  }
}

async function handleUpdate() {
  if (!downloadURL.value) return
  updating.value = true
  try {
    await applyUpdate(downloadURL.value)
  } catch {
    updating.value = false
  }
}
</script>

<style scoped>
.update-panel {
  border-top: 1px solid var(--border-light);
  padding: 0.5rem;
  display: flex;
  flex-direction: column;
  gap: 0.4rem;
}

.version-btn {
  width: 100%;
  padding: 0.4rem;
  border: none;
  border-radius: var(--radius);
  background: var(--bg-elevated);
  color: var(--text-muted);
  font-size: 0.6875rem;
  font-weight: 600;
  cursor: pointer;
  transition: var(--transition);
  text-align: center;
}
.version-btn:hover { color: var(--text); background: var(--bg-hover); }
.version-btn:disabled { opacity: 0.7; cursor: default; }

.update-btn {
  width: 100%;
  padding: 0.4rem;
  border: none;
  border-radius: var(--radius);
  background: var(--green);
  color: #000;
  font-size: 0.75rem;
  font-weight: 700;
  cursor: pointer;
  transition: var(--transition);
}
.update-btn:hover { opacity: 0.9; }
.update-btn:disabled { opacity: 0.5; cursor: not-allowed; }

.progress-bar {
  height: 3px;
  background: var(--bg-elevated);
  border-radius: 2px;
  overflow: hidden;
  margin-top: 0.25rem;
}
.progress-fill {
  height: 100%;
  background: var(--green);
  border-radius: 2px;
}
.fill-anim {
  width: 60%;
  animation: progress-indeterminate 1.5s ease-in-out infinite;
}
@keyframes progress-indeterminate {
  0% { transform: translateX(-100%); }
  100% { transform: translateX(200%); }
}
</style>