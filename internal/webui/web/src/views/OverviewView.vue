<template>
  <div class="view">
    <!-- Welcome typewriter (shown once per session) -->
    <div v-if="showWelcome" class="welcome-block">
      <div class="welcome-line wl-1">◈ POLYTRADE TERMINAL</div>
      <div class="welcome-line wl-2">━━━━━━━━━━━━━━━━━━━━━━━━━━━━</div>
      <div class="welcome-line wl-3">Wallet: <span class="mono">{{ overview.wallet || '—' }}</span></div>
      <div class="welcome-line wl-4">Balance: <span class="mono">${{ fmt2(overview.balance) }}</span></div>
      <div class="welcome-line wl-5">Subsystems: <span class="mono">{{ activeCount }}/{{ overview.subsystems?.length ?? 0 }} active</span></div>
    </div>

    <template v-else>
      <!-- Stat cards -->
      <div class="stat-grid">
        <div class="stat-card anim-in">
          <div class="stat-label">{{ $t('overview.balance') }}</div>
          <div class="stat-value">${{ fmt2(overview.balance) }}</div>
        </div>
        <div class="stat-card anim-in">
          <div class="stat-label">{{ $t('overview.wallet') }}</div>
          <div class="stat-value stat-addr mono">{{ overview.wallet || '—' }}</div>
        </div>
        <div class="stat-card anim-in">
          <div class="stat-label">Open Orders</div>
          <div class="stat-value">{{ overview.orders ?? 0 }}</div>
        </div>
        <div class="stat-card anim-in">
          <div class="stat-label">Positions</div>
          <div class="stat-value">{{ overview.positions ?? 0 }}</div>
        </div>
      </div>

      <!-- Equity curve chart -->
      <div class="chart-panel anim-in">
        <div class="panel-header">Session P&L</div>
        <div ref="chartEl" class="chart-canvas" />
      </div>

      <!-- Subsystems -->
      <div class="section-header anim-in">{{ $t('overview.subsystems') }}</div>
      <div class="subsystem-list anim-in">
        <div v-for="s in overview.subsystems" :key="s.name" class="subsystem-row">
          <span class="sub-dot" :class="s.active ? 'sub-dot--on' : 'sub-dot--off'" />
          <span class="sub-name">{{ s.name }}</span>
          <span class="sub-badge" :class="s.active ? 'badge--ok' : 'badge--off'">
            {{ s.active ? $t('overview.active') : $t('overview.inactive') }}
          </span>
        </div>
        <div v-if="!overview.subsystems?.length" class="empty">{{ $t('common.loading') }}</div>
      </div>

      <!-- Wallet Summary -->
      <template v-if="wallets.length">
        <div class="section-header anim-in">Wallets</div>
        <div class="wallet-summary anim-in">
          <div class="wallet-aggregate">
            <span class="ws-label">Total Balance:</span>
            <span class="ws-val mono">${{ fmt2(totalBalance) }}</span>
            <span class="ws-sep">│</span>
            <span class="ws-label">Total P&amp;L:</span>
            <span class="ws-val mono" :class="totalPnL >= 0 ? 'pnl-pos' : 'pnl-neg'">
              {{ totalPnL >= 0 ? '+' : '' }}{{ fmt2(totalPnL) }}
            </span>
            <span class="ws-sep">│</span>
            <span class="ws-label">Active:</span>
            <span class="ws-val mono">{{ activeWallets }}/{{ wallets.length }}</span>
          </div>
          <table class="wallet-table">
            <thead>
              <tr>
                <th>Label</th><th>Balance</th><th>P&amp;L</th><th>Status</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="w in wallets" :key="w.id">
                <td>{{ w.label || w.id }}</td>
                <td class="mono">${{ fmt2(w.balance_usd) }}</td>
                <td class="mono" :class="w.pnl_usd >= 0 ? 'pnl-pos' : 'pnl-neg'">
                  {{ w.pnl_usd >= 0 ? '+' : '' }}{{ fmt2(w.pnl_usd) }}
                </td>
                <td>
                  <span class="sub-dot" :class="w.enabled ? 'sub-dot--on' : 'sub-dot--off'" />
                  {{ w.enabled ? 'ON' : 'OFF' }}
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </template>
    </template>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import { storeToRefs } from 'pinia'
import { useAppStore } from '@/stores/app'
import { useApi } from '@/composables/useApi'
import { createChart } from 'lightweight-charts'

const app = useAppStore()
const { overview, walletsMap } = storeToRefs(app)
const api = useApi()

const wallets = computed(() => Object.values(walletsMap.value))
const totalBalance = computed(() => wallets.value.reduce((s, w) => s + (w.balance_usd || 0), 0))
const totalPnL = computed(() => wallets.value.reduce((s, w) => s + (w.pnl_usd || 0), 0))
const activeWallets = computed(() => wallets.value.filter(w => w.enabled).length)

const showWelcome = ref(true)
const chartEl = ref(null)
let chart = null
let series = null
const pnlHistory = []

const activeCount = computed(() => overview.value.subsystems?.filter(s => s.active).length ?? 0)
function fmt2(n) { return (+(n || 0)).toFixed(2) }

onMounted(async () => {
  try { app.overview = await api.getOverview() } catch {}
  setTimeout(() => { showWelcome.value = false }, 2200)
})

watch(showWelcome, (val) => {
  if (!val) setTimeout(() => initChart(), 80)
})

watch(() => overview.value.balance, (bal) => {
  if (bal == null) return
  const t = Math.floor(Date.now() / 1000)
  // Avoid duplicate timestamps (lightweight-charts requires strictly ascending time)
  const last = pnlHistory[pnlHistory.length - 1]
  if (last && last.time >= t) return
  pnlHistory.push({ time: t, value: Number((+(bal || 0)).toFixed(2)) })
  if (pnlHistory.length > 120) pnlHistory.shift()
  if (series) series.setData([...pnlHistory])
})

function initChart() {
  if (!chartEl.value) return
  chart = createChart(chartEl.value, {
    width: chartEl.value.clientWidth,
    height: 140,
    layout: { background: { color: 'transparent' }, textColor: '#8b7ec8' },
    grid: { vertLines: { color: '#2d2660' }, horzLines: { color: '#2d2660' } },
    rightPriceScale: { borderColor: '#2d2660' },
    timeScale: { borderColor: '#2d2660', timeVisible: true },
    crosshair: { mode: 1 },
    handleScroll: false,
    handleScale: false,
  })
  series = chart.addLineSeries({
    color: '#a78bfa',
    lineWidth: 2,
    priceLineVisible: false,
    lastValueVisible: true,
  })
  if (pnlHistory.length) series.setData([...pnlHistory])
}

onUnmounted(() => { if (chart) { chart.remove(); chart = null } })
</script>

<style scoped>
.view { display: flex; flex-direction: column; gap: 1.25rem; }

/* Welcome typewriter */
.welcome-block {
  display: flex; flex-direction: column; gap: 0.45rem;
  padding: 1.5rem; background: var(--bg-card);
  border: 1px solid var(--border); border-radius: var(--radius);
  font-family: var(--font-mono); overflow: hidden;
}
.welcome-line {
  font-size: 0.85rem; overflow: hidden; white-space: nowrap;
  animation: typewriter 0.5s steps(40, end) both;
  opacity: 0;
  animation-fill-mode: forwards;
}
.wl-1 { animation-delay: 0.1s; color: var(--accent-bright); font-weight: 600; }
.wl-2 { animation-delay: 0.5s; color: var(--text-muted); font-size: 0.72rem; }
.wl-3 { animation-delay: 0.85s; }
.wl-4 { animation-delay: 1.15s; }
.wl-5 { animation-delay: 1.45s; }

/* Stat grid */
.stat-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(180px, 1fr)); gap: 0.75rem; }
.stat-card {
  background: var(--bg-card); border: 1px solid var(--border);
  border-radius: var(--radius); padding: 1rem 1.25rem;
  transition: border-color var(--transition);
}
.stat-card:hover { border-color: var(--accent); }
.stat-label {
  font-size: 0.65rem; text-transform: uppercase; letter-spacing: 0.08em;
  color: var(--text-secondary); margin-bottom: 0.4rem;
}
.stat-value {
  font-size: 1.35rem; font-weight: 600; color: var(--accent-bright);
  font-family: var(--font-mono);
}
.stat-addr { font-size: 0.72rem; word-break: break-all; }

/* Chart panel */
.chart-panel {
  background: var(--bg-card); border: 1px solid var(--border);
  border-radius: var(--radius); overflow: hidden;
}
.panel-header {
  padding: 0.45rem 1rem; font-size: 0.65rem; text-transform: uppercase;
  letter-spacing: 0.08em; color: var(--text-secondary);
  border-bottom: 1px solid var(--border);
}
.chart-canvas { width: 100%; }

/* Subsystems */
.section-header {
  font-size: 0.65rem; text-transform: uppercase; letter-spacing: 0.08em;
  color: var(--text-secondary);
}
.subsystem-list {
  background: var(--bg-card); border: 1px solid var(--border);
  border-radius: var(--radius); overflow: hidden;
}
.subsystem-row {
  display: flex; align-items: center; gap: 0.75rem;
  padding: 0.6rem 1rem; border-bottom: 1px solid var(--border-subtle);
  transition: background var(--transition);
}
.subsystem-row:last-child { border-bottom: none; }
.subsystem-row:hover { background: var(--bg-hover); }

.sub-dot { width: 7px; height: 7px; border-radius: 50%; flex-shrink: 0; }
.sub-dot--on  { background: var(--success); box-shadow: 0 0 6px var(--success); animation: pulse-dot 2.5s ease infinite; }
.sub-dot--off { background: var(--text-muted); }
.sub-name  { flex: 1; font-size: 0.85rem; }
.sub-badge { font-size: 0.65rem; font-weight: 600; padding: 0.15rem 0.45rem; border-radius: 999px; }
.badge--ok  { background: var(--success-dim); color: var(--success); }
.badge--off { background: var(--badge-bg);    color: var(--text-muted); }

.mono  { font-family: var(--font-mono); }
.empty { padding: 1rem; color: var(--text-muted); text-align: center; }

/* Wallet summary */
.wallet-summary {
  background: var(--bg-card); border: 1px solid var(--border);
  border-radius: var(--radius); overflow: hidden;
}
.wallet-aggregate {
  display: flex; align-items: center; gap: 0.75rem; flex-wrap: wrap;
  padding: 0.6rem 1rem; border-bottom: 1px solid var(--border);
  font-size: 0.82rem;
}
.ws-label { color: var(--text-secondary); font-size: 0.65rem; text-transform: uppercase; letter-spacing: 0.06em; }
.ws-val { font-weight: 600; font-family: var(--font-mono); }
.ws-sep { color: var(--border); }
.wallet-table { width: 100%; border-collapse: collapse; font-size: 0.82rem; }
.wallet-table th {
  padding: 0.4rem 1rem; text-align: left; font-size: 0.65rem;
  text-transform: uppercase; letter-spacing: 0.08em;
  color: var(--text-secondary); border-bottom: 1px solid var(--border);
}
.wallet-table td { padding: 0.4rem 1rem; border-bottom: 1px solid var(--border-subtle); }
.wallet-table tr:last-child td { border-bottom: none; }
.wallet-table tr:hover td { background: var(--bg-hover); }
.pnl-pos { color: var(--success); }
.pnl-neg { color: var(--danger); }
</style>
