<template>
  <div class="layout">
    <aside class="sidebar-left">
      <div class="sidebar-brand">
        <router-link to="/" class="brand-link">
          <span class="brand-icon">
            <svg viewBox="0 0 24 24" fill="currentColor" width="22" height="22">
              <circle cx="12" cy="12" r="3"/>
              <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm0 18c-4.41 0-8-3.59-8-8s3.59-8 8-8 8 3.59 8 8-3.59 8-8 8z"/>
              <path d="M12 6c-3.31 0-6 2.69-6 6s2.69 6 6 6 6-2.69 6-6-2.69-6-6-6zm0 10c-2.21 0-4-1.79-4-4s1.79-4 4-4 4 1.79 4 4-1.79 4-4 4z"/>
            </svg>
          </span>
          <span class="brand-text">Gentry</span>
        </router-link>
      </div>

      <nav class="sidebar-nav">
        <router-link to="/" class="nav-item" :class="{ active: $route.path === '/' }">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="18" height="18"><rect x="3" y="3" width="7" height="7" rx="1"/><rect x="14" y="3" width="7" height="7" rx="1"/><rect x="3" y="14" width="7" height="7" rx="1"/><rect x="14" y="14" width="7" height="7" rx="1"/></svg>
          <span>仪表盘</span>
        </router-link>
        <router-link to="/add" class="nav-item" :class="{ active: $route.path === '/add' }">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="18" height="18"><circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="16"/><line x1="8" y1="12" x2="16" y2="12"/></svg>
          <span>新增监控</span>
        </router-link>
        <router-link to="/push" class="nav-item" :class="{ active: $route.path === '/push' }">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="18" height="18"><path d="M18 8A6 6 0 0 0 6 8c0 7-3 9-3 9h18s-3-2-3-9"/><path d="M13.73 21a2 2 0 0 1-3.46 0"/></svg>
          <span>推送管理</span>
        </router-link>
        <router-link to="/scan-rules" class="nav-item" :class="{ active: $route.path === '/scan-rules' }">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="18" height="18"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><polyline points="14 2 14 8 20 8"/><line x1="16" y1="13" x2="8" y2="13"/><line x1="16" y1="17" x2="8" y2="17"/></svg>
          <span>扫描规则</span>
        </router-link>
      </nav>

      <div class="sidebar-footer">
        <button class="nav-item theme-btn" :title="isDark ? '切换亮色模式' : '切换暗色模式'" @click="toggleTheme">
          <svg v-if="isDark" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="18" height="18"><circle cx="12" cy="12" r="5"/><line x1="12" y1="1" x2="12" y2="3"/><line x1="12" y1="21" x2="12" y2="23"/><line x1="4.22" y1="4.22" x2="5.64" y2="5.64"/><line x1="18.36" y1="18.36" x2="19.78" y2="19.78"/><line x1="1" y1="12" x2="3" y2="12"/><line x1="21" y1="12" x2="23" y2="12"/><line x1="4.22" y1="19.78" x2="5.64" y2="18.36"/><line x1="18.36" y1="5.64" x2="19.78" y2="4.22"/></svg>
          <svg v-else viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="18" height="18"><path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z"/></svg>
          <span>{{ isDark ? '亮色' : '暗色' }}</span>
        </button>
      </div>
    </aside>

    <main class="main-content">
      <slot />
    </main>

    <aside class="sidebar-right">
      <div class="stats-header">
        <h3>系统概览</h3>
      </div>
      <div class="stats-list">
        <div class="stat-item">
          <div class="stat-value">{{ stats.total_monitors }}</div>
          <div class="stat-label">监控器总数</div>
        </div>
        <div class="stat-item">
          <div class="stat-value stat-green">{{ stats.running_monitors }}</div>
          <div class="stat-label">运行中</div>
        </div>
        <div class="stat-item">
          <div class="stat-value">{{ stats.total_updates }}</div>
          <div class="stat-label">变更记录</div>
        </div>
        <div class="stat-item">
          <div class="stat-value stat-blue">{{ stats.updates_last_hour }}</div>
          <div class="stat-label">近1小时更新</div>
        </div>
        <div class="stat-item">
          <div class="stat-value stat-orange">{{ stats.unnotified_updates }}</div>
          <div class="stat-label">待推送</div>
        </div>
        <div class="stat-item">
          <div class="stat-value">{{ stats.total_accounts }}</div>
          <div class="stat-label">推送账户</div>
        </div>
      </div>
    </aside>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, onUnmounted } from 'vue'
import { fetchStats } from '../api/monitors'

const STORAGE_KEY = 'gentry_theme'
const isDark = ref(false)
const stats = reactive({
  total_monitors: 0,
  running_monitors: 0,
  total_updates: 0,
  updates_last_hour: 0,
  unnotified_updates: 0,
  total_accounts: 0,
})

let statsTimer = null

function applyTheme(dark) {
  document.documentElement.classList.toggle('dark', dark)
  isDark.value = dark
  localStorage.setItem(STORAGE_KEY, dark ? 'dark' : 'light')
}

function toggleTheme() {
  applyTheme(!isDark.value)
}

async function loadStats() {
  try {
    const res = await fetchStats()
    if (res.code === 0 && res.data) {
      Object.assign(stats, res.data)
    }
  } catch (_) {}
}

onMounted(() => {
  const saved = localStorage.getItem(STORAGE_KEY)
  applyTheme(saved === 'dark')
  loadStats()
  statsTimer = setInterval(loadStats, 15000)
})

onUnmounted(() => {
  if (statsTimer) clearInterval(statsTimer)
})
</script>

<style scoped>
.layout {
  display: flex;
  min-height: 100vh;
  background: var(--bg-base);
}

.sidebar-left {
  width: 200px;
  min-width: 200px;
  background: var(--bg-surface);
  border-right: 1px solid var(--border);
  display: flex;
  flex-direction: column;
  position: fixed;
  top: 0;
  left: 0;
  bottom: 0;
  z-index: 100;
}

.sidebar-brand {
  padding: 1.25rem 1rem;
  border-bottom: 1px solid var(--border-light);
}

.brand-link {
  display: flex;
  align-items: center;
  gap: 0.6rem;
  text-decoration: none;
  color: var(--text);
}

.brand-icon {
  color: var(--green);
  display: flex;
}

.brand-text {
  font-size: 1rem;
  font-weight: 700;
  letter-spacing: -0.3px;
}

.sidebar-nav {
  flex: 1;
  padding: 0.75rem 0.5rem;
  display: flex;
  flex-direction: column;
  gap: 0.15rem;
}

.nav-item {
  display: flex;
  align-items: center;
  gap: 0.65rem;
  padding: 0.55rem 0.75rem;
  border-radius: var(--radius);
  color: var(--text-secondary);
  text-decoration: none;
  font-size: 0.8125rem;
  font-weight: 600;
  transition: var(--transition);
  border: none;
  background: none;
  cursor: pointer;
  width: 100%;
  text-align: left;
}

.nav-item:hover {
  color: var(--text);
  background: var(--bg-hover);
}

.nav-item.active {
  color: var(--text);
  background: var(--bg-elevated);
}

.nav-item svg {
  flex-shrink: 0;
  opacity: 0.7;
}

.nav-item.active svg {
  opacity: 1;
}

.sidebar-footer {
  padding: 0.5rem;
  border-top: 1px solid var(--border-light);
}

.theme-btn {
  font-size: 0.75rem;
  color: var(--text-muted);
}

.theme-btn:hover {
  color: var(--text-secondary);
}

.main-content {
  flex: 1;
  margin-left: 200px;
  margin-right: 220px;
  padding: 2rem;
  min-width: 0;
}

.sidebar-right {
  width: 220px;
  min-width: 220px;
  background: var(--bg-surface);
  border-left: 1px solid var(--border);
  position: fixed;
  top: 0;
  right: 0;
  bottom: 0;
  overflow-y: auto;
  padding: 1.25rem 1rem;
}

.stats-header h3 {
  font-size: 0.6875rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 1.2px;
  color: var(--text-muted);
  margin-bottom: 1rem;
}

.stats-list {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.stat-item {
  padding: 0.65rem 0.75rem;
  border-radius: var(--radius);
  transition: var(--transition);
}

.stat-item:hover {
  background: var(--bg-hover);
}

.stat-value {
  font-size: 1.5rem;
  font-weight: 700;
  color: var(--text);
  line-height: 1.2;
  font-variant-numeric: tabular-nums;
}

.stat-green { color: var(--green); }
.stat-blue { color: #1976d2; }
.stat-orange { color: var(--warning); }

.stat-label {
  font-size: 0.6875rem;
  color: var(--text-muted);
  margin-top: 0.1rem;
  font-weight: 500;
}

@media (max-width: 1100px) {
  .sidebar-right { display: none; }
  .main-content { margin-right: 0; }
}

@media (max-width: 768px) {
  .sidebar-left {
    position: fixed;
    bottom: 0;
    top: auto;
    left: 0;
    right: 0;
    width: 100%;
    min-width: 100%;
    height: 56px;
    flex-direction: row;
    border-right: none;
    border-top: 1px solid var(--border);
    z-index: 200;
  }

  .sidebar-brand,
  .sidebar-footer { display: none; }

  .sidebar-nav {
    flex-direction: row;
    justify-content: space-around;
    padding: 0;
    gap: 0;
    width: 100%;
    align-items: center;
  }

  .nav-item {
    flex-direction: column;
    gap: 0.15rem;
    padding: 0.4rem 0.5rem;
    font-size: 0.625rem;
    justify-content: center;
  }

  .main-content {
    margin-left: 0;
    margin-right: 0;
    padding: 1rem;
    padding-bottom: 72px;
  }
}
</style>
