<template>
  <header class="topbar">
    <div class="brand">
      <span class="brand-glyph">◈</span>
      <span class="brand-name">POLYTRADE</span>
    </div>
    <div class="topbar-right">
      <span class="ws-indicator" :class="connected ? 'ws--on' : 'ws--off'">
        <span class="ws-dot" />
        <span class="ws-text">{{ connected ? $t('common.connected') : $t('common.disconnected') }}</span>
      </span>
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
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useAppStore } from '@/stores/app'
import { LANGS } from '@/i18n'

const { locale } = useI18n()
const router = useRouter()
const auth = useAuthStore()
const app = useAppStore()
const connected = computed(() => app.connected)
const currentLang = ref(locale.value)

function changeLang() { locale.value = currentLang.value; localStorage.setItem('lang', currentLang.value) }
function logout() { auth.logout(); router.push('/login') }
</script>

<style scoped>
.topbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 1rem;
  height: var(--topbar-h);
  background: var(--bg-sidebar);
  border-bottom: 1px solid var(--border);
  flex-shrink: 0;
  z-index: 100;
}

.brand { display: flex; align-items: center; gap: 0.5rem; }
.brand-glyph { color: var(--accent-bright); font-size: 1rem; }
.brand-name  { font-size: 0.75rem; font-weight: 700; letter-spacing: 0.16em; color: var(--accent-bright); }

.topbar-right { display: flex; align-items: center; gap: 0.75rem; }

.ws-indicator {
  display: flex; align-items: center; gap: 0.3rem;
  font-size: 0.7rem; color: var(--text-secondary);
}
.ws-dot { width: 6px; height: 6px; border-radius: 50%; background: var(--text-muted); flex-shrink: 0; }
.ws--on  .ws-dot { background: var(--success); box-shadow: 0 0 5px var(--success); animation: pulse-dot 2.5s ease infinite; }
.ws--off .ws-dot { background: var(--warning); animation: pulse-dot 1s ease infinite; }
.ws-text { display: none; }
@media (min-width: 640px) { .ws-text { display: inline; } }

.lang-select {
  background: var(--bg-hover); border: 1px solid var(--border);
  color: var(--text-secondary); border-radius: var(--radius);
  padding: 0.15rem 0.35rem; font-size: 0.7rem; cursor: pointer;
  font-family: var(--font-mono);
}

.btn-logout {
  background: none; border: 1px solid var(--border); color: var(--text-secondary);
  border-radius: var(--radius); padding: 0.2rem 0.45rem; font-size: 0.85rem;
  cursor: pointer; transition: background var(--transition), color var(--transition);
  line-height: 1;
}
.btn-logout:hover { background: var(--danger-dim); color: var(--danger); border-color: var(--danger); }
</style>
