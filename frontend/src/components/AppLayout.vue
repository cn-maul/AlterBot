<template>
  <div class="layout">
    <nav class="navbar">
      <div class="nav-inner">
        <router-link to="/" class="nav-brand">
          <span class="brand-icon">
            <svg viewBox="0 0 24 24" fill="currentColor" width="24" height="24">
              <circle cx="12" cy="12" r="3"/>
              <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm0 18c-4.41 0-8-3.59-8-8s3.59-8 8-8 8 3.59 8 8-3.59 8-8 8z"/>
              <path d="M12 6c-3.31 0-6 2.69-6 6s2.69 6 6 6 6-2.69 6-6-2.69-6-6-6zm0 10c-2.21 0-4-1.79-4-4s1.79-4 4-4 4 1.79 4 4-1.79 4-4 4z"/>
            </svg>
          </span>
          <span class="brand-text">Gentry</span>
        </router-link>
        <div class="nav-links">
          <router-link to="/" class="nav-link" :class="{ active: $route.path === '/' }">
            仪表盘
          </router-link>
          <router-link to="/add" class="nav-link" :class="{ active: $route.path === '/add' }">
            新增监控
          </router-link>
          <router-link to="/push" class="nav-link" :class="{ active: $route.path === '/push' }">
            推送管理
          </router-link>
          <router-link to="/scan-rules" class="nav-link" :class="{ active: $route.path === '/scan-rules' }">
            扫描规则
          </router-link>
        </div>
        <div class="nav-right">
          <button class="nav-link theme-toggle" :title="isDark ? '切换亮色模式' : '切换暗色模式'" @click="toggleTheme">
            <svg v-if="isDark" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="16" height="16"><circle cx="12" cy="12" r="5"/><line x1="12" y1="1" x2="12" y2="3"/><line x1="12" y1="21" x2="12" y2="23"/><line x1="4.22" y1="4.22" x2="5.64" y2="5.64"/><line x1="18.36" y1="18.36" x2="19.78" y2="19.78"/><line x1="1" y1="12" x2="3" y2="12"/><line x1="21" y1="12" x2="23" y2="12"/><line x1="4.22" y1="19.78" x2="5.64" y2="18.36"/><line x1="18.36" y1="5.64" x2="19.78" y2="4.22"/></svg>
            <svg v-else viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="16" height="16"><path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z"/></svg>
            <span>主题</span>
          </button>
        </div>
      </div>
    </nav>
    <main class="main-content">
      <slot />
    </main>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'

const STORAGE_KEY = 'gentry_theme'

const isDark = ref(false)

function applyTheme(dark) {
  document.documentElement.classList.toggle('dark', dark)
  isDark.value = dark
  localStorage.setItem(STORAGE_KEY, dark ? 'dark' : 'light')
}

function toggleTheme() {
  applyTheme(!isDark.value)
}

onMounted(() => {
  const saved = localStorage.getItem(STORAGE_KEY)
  if (saved === 'dark') {
    applyTheme(true)
  } else {
    applyTheme(false)
  }
})
</script>

<style scoped>
.layout {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  background: var(--bg-base);
}

.navbar {
  background: var(--bg-base);
  border-bottom: 1px solid var(--border);
  padding: 0 2rem;
  position: sticky;
  top: 0;
  z-index: 100;
}

.nav-inner {
  max-width: 1200px;
  margin: 0 auto;
  display: flex;
  align-items: center;
  height: 64px;
  gap: 2rem;
}

.nav-brand {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-weight: 700;
  font-size: 1.125rem;
  color: var(--text);
  text-decoration: none;
}

.brand-icon {
  color: var(--green);
  display: flex;
}

.nav-links {
  display: flex;
  gap: 0.25rem;
}

.nav-link {
  padding: 0.5rem 1rem;
  border-radius: var(--radius-pill);
  color: var(--text-secondary);
  text-decoration: none;
  font-size: 0.875rem;
  font-weight: 700;
  transition: var(--transition);
  display: inline-flex;
  align-items: center;
  gap: 0.4rem;
  background: none;
  border: none;
  cursor: pointer;
  text-transform: uppercase;
  letter-spacing: 1px;
}

.nav-link:hover {
  color: var(--text);
}

.nav-link.active {
  color: var(--text);
  background: var(--bg-elevated);
}

.nav-right {
  margin-left: auto;
  display: flex;
  align-items: center;
  gap: 0.25rem;
}

.nav-right .nav-link {
  font-size: 0.75rem;
  letter-spacing: 0.5px;
  text-transform: none;
}

.nav-right .nav-link svg {
  width: 16px;
  height: 16px;
}

.main-content {
  flex: 1;
  max-width: 1200px;
  width: 100%;
  margin: 0 auto;
  padding: 2rem;
}

@media (max-width: 768px) {
  .navbar { padding: 0 1rem; }
  .nav-inner { gap: 1rem; }
  .nav-link { padding: 0.4rem 0.7rem; font-size: 0.75rem; }
  .brand-text { display: none; }
}
</style>