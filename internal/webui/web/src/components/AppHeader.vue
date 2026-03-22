<template>
  <header class="app-header">
    <div class="header-left">
      <span class="page-title">{{ pageTitle }}</span>
    </div>
    <div class="header-right">
      <div class="status-pill" :class="connected ? 'pill-live' : 'pill-offline'">
        <span class="pill-dot"></span>
        <span>{{ connected ? 'LIVE' : 'OFFLINE' }}</span>
      </div>
      <div v-if="engineOn" class="status-pill pill-engine">ENGINE ON</div>
      <select class="lang-select" v-model="currentLang" @change="changeLang">
        <option v-for="l in LANGS" :key="l" :value="l">{{ l.toUpperCase() }}</option>
      </select>
      <button class="btn-logout" @click="logout" title="Logout">⏻</button>
    </div>
  </header>
</template>

<script setup>
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter, useRoute } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useAppStore } from '@/stores/app'
import { LANGS } from '@/i18n'
import { storeToRefs } from 'pinia'

const { locale } = useI18n()
const router = useRouter()
const route = useRoute()
const auth = useAuthStore()
const app = useAppStore()
const { connected, overview } = storeToRefs(app)

const currentLang = ref(locale.value)

const PAGE_TITLES = {
  '/': 'Overview',
  '/markets': 'Markets',
  '/orders': 'Orders & Positions',
  '/wallets': 'Wallets',
  '/strategies': 'Strategies',
  '/copytrading': 'Copytrading',
  '/logs': 'Logs',
  '/settings': 'Settings',
}
const pageTitle = computed(() => PAGE_TITLES[route.path] || route.name || 'PolyTrade')

const engineOn = computed(() => overview.value?.subsystems?.trading === 'running')

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
  height: var(--header-h-new, 54px);
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 24px;
  border-bottom: 1px solid var(--border);
  background: var(--bg-panel);
  flex-shrink: 0;
  z-index: 50;
}

.header-left { display: flex; align-items: center; gap: 10px; }
.page-title { font-size: var(--font-size-md, 15px); font-weight: 600; color: var(--fg); letter-spacing: 0.01em; }

.header-right { display: flex; align-items: center; gap: 8px; }

.status-pill {
  display: flex; align-items: center; gap: 5px;
  font-size: 10px; font-weight: 700; letter-spacing: 0.07em;
  padding: 3px 9px; border-radius: 3px; border: 1px solid;
}
.pill-live   { color: var(--success); border-color: rgba(16,217,148,0.40); background: rgba(16,217,148,0.08); }
.pill-offline{ color: var(--fg-muted); border-color: var(--border); background: transparent; }
.pill-engine { color: var(--accent-bright); border-color: rgba(124,58,237,0.40); background: rgba(124,58,237,0.08); }

.pill-dot {
  width: 5px; height: 5px; border-radius: 50%; background: currentColor;
}
.pill-live .pill-dot { animation: pulse-dot 2s ease infinite; }

.lang-select {
  background: rgba(255,255,255,0.04);
  border: 1px solid var(--border);
  color: var(--fg-muted);
  border-radius: var(--radius);
  padding: 2px 6px;
  font-size: 12px;
  font-family: var(--font-mono);
  cursor: pointer;
  outline: none;
}
.lang-select:focus { border-color: var(--accent); }

.btn-logout {
  background: none; border: 1px solid var(--border);
  color: var(--fg-muted); border-radius: var(--radius);
  padding: 3px 8px; font-size: 14px;
  cursor: pointer; line-height: 1;
  transition: all var(--transition);
}
.btn-logout:hover { background: var(--danger-dim); color: var(--danger); border-color: var(--danger); }
</style>
