<template>
  <div class="layout">
    <nav class="navbar">
      <div class="nav-inner">
        <router-link to="/" class="nav-brand">
          <span class="brand-icon">📡</span>
          <span class="brand-text">AlterBot</span>
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
        </div>
        <div class="nav-right">
          <button class="btn-icon theme-toggle" :title="isDark ? '切换亮色模式' : '切换暗色模式'" @click="toggleTheme">
            <svg v-if="isDark" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="5"/><line x1="12" y1="1" x2="12" y2="3"/><line x1="12" y1="21" x2="12" y2="23"/><line x1="4.22" y1="4.22" x2="5.64" y2="5.64"/><line x1="18.36" y1="18.36" x2="19.78" y2="19.78"/><line x1="1" y1="12" x2="3" y2="12"/><line x1="21" y1="12" x2="23" y2="12"/><line x1="4.22" y1="19.78" x2="5.64" y2="18.36"/><line x1="18.36" y1="5.64" x2="19.78" y2="4.22"/></svg>
            <svg v-else viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z"/></svg>
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

const STORAGE_KEY = 'alterbot_theme'

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
  }
})
</script>

<style scoped>
.layout {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
}

.navbar {
  background: var(--bg-card);
  border-bottom: 1px solid var(--border);
  padding: 0 2rem;
  position: sticky;
  top: 0;
  z-index: 100;
}

.nav-inner {
  max-width: 1100px;
  margin: 0 auto;
  display: flex;
  align-items: center;
  height: 56px;
  gap: 2rem;
}

.nav-right {
  margin-left: auto;
  display: flex;
  align-items: center;
  gap: 0.25rem;
}

.nav-brand {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-weight: 700;
  font-size: 1rem;
  color: var(--text);
  text-decoration: none;
}

.brand-icon {
  font-size: 1.2rem;
}

.nav-links {
  display: flex;
  gap: 0.25rem;
}

.nav-link {
  padding: 0.5rem 1rem;
  border-radius: var(--radius);
  color: var(--text-secondary);
  text-decoration: none;
  font-size: 0.9rem;
  font-weight: 500;
  transition: var(--transition);
}

.nav-link:hover {
  background: var(--bg-hover);
  color: var(--text);
}

.nav-link.active {
  background: var(--primary-bg);
  color: var(--primary);
}

.theme-toggle {
  width: 34px;
  height: 34px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.theme-toggle svg {
  width: 18px;
  height: 18px;
}

.main-content {
  flex: 1;
  max-width: 1100px;
  width: 100%;
  margin: 0 auto;
  padding: 2rem;
}
</style>
