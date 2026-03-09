<template>
  <header class="topbar">
    <!-- Live metrics ticker -->
    <div class="metrics-strip">
      <div class="metric-chip">
        <span class="metric-key">BAL</span>
        <span class="metric-val num-glow">${{ fmt2(overview.balance) }}</span>
      </div>
      <div class="metric-sep">│</div>
      <div class="metric-chip">
        <span class="metric-key">ORDERS</span>
        <span class="metric-val">{{ overview.orders ?? 0 }}</span>
      </div>
      <div class="metric-sep">│</div>
      <div class="metric-chip">
        <span class="metric-key">POS</span>
        <span class="metric-val">{{ overview.positions ?? 0 }}</span>
      </div>
      <div class="metric-sep">│</div>
      <div class="metric-chip">
        <span class="metric-key">WALLET</span>
        <span class="metric-val addr-val">{{ shortAddr }}</span>
      </div>
    </div>

    <!-- Right controls -->
    <div class="topbar-right">
      <div class="ws-pill" :class="connected ? 'ws-pill--on' : 'ws-pill--off'">
        <span class="ws-dot" />
        <span class="ws-label">{{ connected ? 'LIVE' : 'DISC' }}</span>
      </div>
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
import { storeToRefs } from 'pinia'

const { locale } = useI18n()
const router = useRouter()
const auth = useAuthStore()
const app = useAppStore()
const { overview, connected } = storeToRefs(app)

const currentLang = ref(locale.value)

const shortAddr = computed(() => {
  const w = overview.value.wallet
  if (!w) return '—'
  return w.slice(0, 6) + '…' + w.slice(-4)
})

function fmt2(n) { return (+(n || 0)).toFixed(2) }
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
  gap: 1rem;
}

/* Metrics strip */
.metrics-strip {
  display: flex;
  align-items: center;
  gap: 0.1rem;
  flex: 1;
  overflow: hidden;
}

.metric-chip {
  display: flex;
  align-items: center;
  gap: 0.35rem;
  padding: 0 0.5rem;
  white-space: nowrap;
}

.metric-key {
  font-size: 0.86rem;
  font-weight: 600;
  color: var(--text-secondary);
  letter-spacing: 0.10em;
  text-transform: uppercase;
}

.metric-val {
  font-size: 0.90rem;
  font-weight: 600;
  color: var(--text-primary);
  font-family: var(--font-mono);
}

.num-glow {
  color: var(--price-bright);
  text-shadow: 0 0 10px rgba(251,191,36,0.35);
}

.addr-val {
  color: var(--text-secondary);
  font-size: 0.94rem;
}

.metric-sep {
  color: var(--border);
  font-size: 0.90rem;
  padding: 0 0.1rem;
  user-select: none;
}

/* Right controls */
.topbar-right {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  flex-shrink: 0;
}

/* WebSocket status pill */
.ws-pill {
  display: flex;
  align-items: center;
  gap: 0.3rem;
  padding: 0.30rem 0.70rem;
  border-radius: 2px;
  border: 1px solid var(--border);
  font-size: 1.00rem;
  font-weight: 700;
  letter-spacing: 0.10em;
}

.ws-pill--on  { border-color: rgba(16,217,148,0.35); color: var(--success); background: rgba(16,217,148,0.06); }
.ws-pill--off { border-color: rgba(245,158,11,0.35);  color: var(--warning); background: rgba(245,158,11,0.06); }

.ws-dot {
  width: 5px; height: 5px;
  border-radius: 50%;
  background: currentColor;
}
.ws-pill--on  .ws-dot { animation: pulse-dot 2s ease infinite; }
.ws-pill--off .ws-dot { animation: pulse-dot 0.8s ease infinite; }

/* Language select */
.lang-select {
  background: var(--bg-hover);
  border: 1px solid var(--border);
  color: var(--text-secondary);
  border-radius: var(--radius);
  padding: 0.2rem 0.4rem;
  font-size: 0.92rem;
  font-family: var(--font-mono);
  cursor: pointer;
  outline: none;
  transition: border-color var(--transition);
}
.lang-select:focus { border-color: var(--accent); }

/* Logout button */
.btn-logout {
  background: none;
  border: 1px solid var(--border);
  color: var(--text-secondary);
  border-radius: var(--radius);
  padding: 0.2rem 0.5rem;
  font-size: 0.96rem;
  cursor: pointer;
  line-height: 1;
  transition: all var(--transition);
}
.btn-logout:hover {
  background: var(--danger-dim);
  color: var(--danger);
  border-color: var(--danger);
}

@media (max-width: 640px) {
  .metric-chip:nth-child(n+5) { display: none; }
  .metric-sep:nth-child(n+4) { display: none; }
}
</style>
