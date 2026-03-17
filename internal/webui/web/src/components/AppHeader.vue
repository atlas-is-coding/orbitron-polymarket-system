<template>
  <header class="topbar">
    <!-- Logo -->
    <div class="brand">
      <span class="brand-glyph">◈</span>
      <span class="brand-name">POLYTRADE</span>
    </div>

    <!-- Live metrics ticker -->
    <div class="metrics-strip">
      <div class="metric-chip">
        <span class="metric-key">BAL</span>
        <span class="metric-val num-glow">${{ fmt2(overview.balance) }}</span>
      </div>
      <div class="metric-sep">│</div>
      <div class="metric-chip">
        <span class="metric-key">P&amp;L</span>
        <span class="metric-val" :class="pnlClass">{{ pnlSign }}${{ fmt2(Math.abs(pnl)) }}</span>
      </div>
      <div class="metric-sep">│</div>
      <div class="metric-chip">
        <span class="metric-key">ORDERS</span>
        <span class="metric-val">{{ overview.orders?.length ?? 0 }}</span>
      </div>
      <div class="metric-sep">│</div>
      <div class="metric-chip">
        <span class="metric-key">POS</span>
        <span class="metric-val">{{ overview.positions?.length ?? 0 }}</span>
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

const pnl = computed(() => +(overview.value.pnl ?? overview.value.pnl_usd ?? 0))
const pnlClass = computed(() => pnl.value >= 0 ? 'pnl-pos' : 'pnl-neg')
const pnlSign = computed(() => pnl.value >= 0 ? '+' : '-')

const shortAddr = computed(() => {
  const w = overview.value.wallet
  if (!w) return '—'
  return w.slice(0, 6) + '…' + w.slice(-4)
})

function fmt2(n) {
  if (n === null || n === undefined || isNaN(n)) return '---'
  return (+(n || 0)).toFixed(2)
}
function changeLang() { locale.value = currentLang.value; localStorage.setItem('lang', currentLang.value) }
function logout() { auth.logout(); router.push('/login') }
</script>

<style scoped>
.topbar {
  display: flex;
  align-items: center;
  padding: 0 1rem;
  height: var(--header-h);
  background: rgba(255,255,255,0.02);
  border-bottom: 1px solid var(--border);
  flex-shrink: 0;
  z-index: 100;
  gap: 1rem;
}

/* Logo */
.brand {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  flex-shrink: 0;
}
.brand-glyph {
  font-size: 1rem;
  color: var(--accent-bright);
  text-shadow: 0 0 14px rgba(124,58,237,0.60);
}
.brand-name {
  font-size: 0.82rem;
  font-weight: 800;
  letter-spacing: 0.18em;
  color: var(--accent-bright);
  line-height: 1;
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
  padding: 0 0.45rem;
  white-space: nowrap;
}

.metric-key {
  font-size: 0.78rem;
  font-weight: 600;
  color: var(--text-secondary);
  letter-spacing: 0.10em;
  text-transform: uppercase;
}

.metric-val {
  font-size: 0.86rem;
  font-weight: 600;
  color: var(--text-primary);
  font-family: var(--font-mono);
}

.num-glow {
  color: var(--price-bright);
  text-shadow: 0 0 10px rgba(251,191,36,0.35);
}

.pnl-pos { color: var(--success); text-shadow: 0 0 8px rgba(16,217,148,0.30); }
.pnl-neg { color: var(--danger);  text-shadow: 0 0 8px rgba(248,113,113,0.25); }

.addr-val {
  color: var(--text-secondary);
  font-size: 0.88rem;
}

.metric-sep {
  color: var(--border);
  font-size: 0.84rem;
  padding: 0 0.05rem;
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
  padding: 0.22rem 0.55rem;
  border-radius: 2px;
  border: 1px solid var(--border);
  font-size: 0.78rem;
  font-weight: 700;
  letter-spacing: 0.10em;
}
.ws-pill--on  { border-color: rgba(52,211,153,0.35); color: var(--success); background: rgba(52,211,153,0.06); }
.ws-pill--off { border-color: rgba(251,191,36,0.35);  color: var(--warning); background: rgba(251,191,36,0.06); }

.ws-dot {
  width: 5px; height: 5px;
  border-radius: 50%;
  background: currentColor;
}
.ws-pill--on  .ws-dot { animation: pulse-dot 2s ease infinite; }
.ws-pill--off .ws-dot { animation: pulse-dot 0.8s ease infinite; }

/* Language select */
.lang-select {
  background: rgba(255,255,255,0.04);
  border: 1px solid var(--border);
  color: var(--text-secondary);
  border-radius: var(--radius);
  padding: 0.15rem 0.35rem;
  font-size: 0.82rem;
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
  padding: 0.15rem 0.45rem;
  font-size: 0.90rem;
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
  .brand-name { display: none; }
  .metric-chip:nth-child(n+7) { display: none; }
  .metric-sep:nth-child(n+6) { display: none; }
}
</style>
