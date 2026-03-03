<template>
  <header class="app-header">
    <div class="brand">
      <span class="brand-icon">◈</span>
      <span class="brand-name">Polytrade</span>
    </div>

    <nav class="nav-tabs">
      <RouterLink
        v-for="item in navItems"
        :key="item.to"
        :to="item.to"
        class="nav-tab"
        active-class="nav-tab--active"
      >{{ $t(item.label) }}</RouterLink>
    </nav>

    <div class="header-actions">
      <span class="ws-dot" :class="connected ? 'ws-dot--on' : 'ws-dot--off'" />
      <span class="ws-label">{{ connected ? $t('common.connected') : $t('common.disconnected') }}</span>

      <button class="btn-icon" @click="toggleTheme" :title="isDark ? 'Light mode' : 'Dark mode'">
        {{ isDark ? '☀' : '☾' }}
      </button>

      <select class="lang-select" v-model="currentLang" @change="changeLang">
        <option v-for="l in LANGS" :key="l" :value="l">{{ l.toUpperCase() }}</option>
      </select>

      <button class="btn-logout" @click="logout">{{ $t('common.cancel') }}</button>
    </div>
  </header>
</template>

<script setup>
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useAppStore } from '@/stores/app'
import { LANGS } from '@/i18n'

const { locale } = useI18n()
const router = useRouter()
const auth = useAuthStore()
const app = useAppStore()
const connected = computed(() => app.connected)

const isDark = ref(document.documentElement.getAttribute('data-theme') !== 'light')
const currentLang = ref(locale.value)

const navItems = [
  { to: '/overview',     label: 'nav.overview' },
  { to: '/orders',       label: 'nav.orders' },
  { to: '/positions',    label: 'nav.positions' },
  { to: '/copytrading',  label: 'nav.copytrading' },
  { to: '/wallets',      label: 'nav.wallets' },
  { to: '/logs',         label: 'nav.logs' },
  { to: '/settings',     label: 'nav.settings' },
]

function toggleTheme() {
  isDark.value = !isDark.value
  document.documentElement.setAttribute('data-theme', isDark.value ? 'dark' : 'light')
  localStorage.setItem('theme', isDark.value ? 'dark' : 'light')
}

function changeLang() {
  locale.value = currentLang.value
  localStorage.setItem('lang', currentLang.value)
}

function logout() {
  auth.logout()
  router.push('/login')
}
</script>

<style scoped>
.app-header {
  display: flex;
  align-items: center;
  gap: 1rem;
  padding: 0 1.5rem;
  height: 52px;
  background: var(--bg-secondary);
  border-bottom: 1px solid var(--border);
  position: sticky;
  top: 0;
  z-index: 100;
  font-family: var(--font-ui);
}

.brand {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  font-weight: 700;
  font-size: 1.1rem;
  color: var(--accent);
  white-space: nowrap;
}
.brand-icon { font-size: 1.3rem; }

.nav-tabs {
  display: flex;
  gap: 0.25rem;
  flex: 1;
  overflow-x: auto;
}

.nav-tab {
  padding: 0.3rem 0.75rem;
  border-radius: var(--radius);
  color: var(--text-secondary);
  text-decoration: none;
  font-size: 0.875rem;
  white-space: nowrap;
  transition: color var(--transition), background var(--transition);
}
.nav-tab:hover { color: var(--text-primary); background: var(--bg-hover); }
.nav-tab--active { color: var(--accent); background: var(--bg-hover); }

.header-actions {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.ws-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  flex-shrink: 0;
}
.ws-dot--on  { background: var(--success); box-shadow: 0 0 6px var(--success); }
.ws-dot--off { background: var(--text-muted); }
.ws-label { font-size: 0.75rem; color: var(--text-secondary); }

.btn-icon {
  background: none;
  border: none;
  cursor: pointer;
  font-size: 1rem;
  padding: 0.25rem 0.4rem;
  border-radius: var(--radius);
  color: var(--text-secondary);
  transition: background var(--transition);
}
.btn-icon:hover { background: var(--bg-hover); color: var(--text-primary); }

.lang-select {
  background: var(--bg-hover);
  border: 1px solid var(--border);
  color: var(--text-secondary);
  border-radius: var(--radius);
  padding: 0.2rem 0.4rem;
  font-size: 0.75rem;
  cursor: pointer;
}

.btn-logout {
  background: none;
  border: 1px solid var(--border);
  color: var(--text-secondary);
  border-radius: var(--radius);
  padding: 0.25rem 0.6rem;
  font-size: 0.8rem;
  cursor: pointer;
  transition: background var(--transition), color var(--transition);
}
.btn-logout:hover { background: var(--danger); color: white; border-color: var(--danger); }
</style>
