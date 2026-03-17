<template>
  <div class="view">
    <!-- Boot terminal (shown once) -->
    <div v-if="showWelcome" class="boot-terminal anim-in">
      <div class="boot-line bl-1">◈ POLYTRADE NEXUS TERMINAL — SYSTEM READY</div>
      <div class="boot-line bl-2">━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━</div>
      <div class="boot-line bl-3">Wallet    <span class="bval">{{ overview.wallet_address || overview.wallet || 'loading...' }}</span></div>
      <div class="boot-line bl-4">Balance   <span class="bval num-glow">${{ fmt2(overview.balance) }}</span></div>
      <div class="boot-line bl-5">Subsystems  <span class="bval">{{ activeCount }}/{{ overview.subsystems?.length ?? 0 }} online</span></div>
    </div>

    <template v-else>
      <!-- KPI row -->
      <div class="kpi-row">
        <div class="kpi-card anim-in" v-for="(kpi, i) in kpis" :key="i">
          <div class="kpi-label">{{ kpi.label }}</div>
          <div class="kpi-value" :class="kpi.cls">{{ kpi.value }}</div>
          <div v-if="kpi.sub" class="kpi-sub">{{ kpi.sub }}</div>
        </div>
      </div>

      <!-- Chart + Subsystems row -->
      <div class="mid-row">
        <!-- Equity chart -->
        <div class="panel flex-2 anim-in">
          <div class="panel-header">
            <span>SESSION P&amp;L</span>
            <span class="ph-right">{{ $t('overview.balance') }}: <span class="num-glow">${{ fmt2(overview.balance) }}</span></span>
          </div>
          <div ref="chartEl" class="chart-canvas" />
        </div>

        <!-- Subsystems -->
        <div class="panel flex-1 anim-in">
          <div class="panel-header"><span>{{ $t('overview.subsystems') }}</span></div>
          <div class="subsystem-list">
            <div v-for="s in overview.subsystems" :key="s.name" class="sub-row">
              <span class="status-dot" :class="s.active ? 'status-dot--on' : 'status-dot--off'" />
              <span class="sub-name">{{ s.name }}</span>
              <span class="badge" :class="s.active ? 'badge--ok' : 'badge--off'">
                {{ s.active ? $t('overview.active') : $t('overview.inactive') }}
              </span>
            </div>
            <div v-if="!overview.subsystems?.length" class="empty-state">{{ $t('common.loading') }}</div>
          </div>
        </div>
      </div>

      <!-- API Health Card -->
      <div class="panel health-panel anim-in">
        <div class="panel-header"><span>{{ $t('health.title') }}</span></div>
        <template v-if="healthStore.snapshot">
          <div
            v-for="svc in healthStore.snapshot.services"
            :key="svc.name"
            class="health-row"
          >
            <span class="health-dot" :class="svc.status">●</span>
            <span class="health-name">{{ svc.name }}</span>
            <span class="health-lat" :class="svc.status">{{ fmtLatency(svc.latency_ms) }}</span>
            <span v-if="svc.error" class="health-err">{{ svc.error }}</span>
          </div>
          <div v-if="healthStore.snapshot.geo" class="health-row health-geo">
            <span class="health-dot" :class="healthStore.snapshot.geo.blocked ? 'down' : 'ok'">●</span>
            <span class="health-name">Geoblock</span>
            <span :class="healthStore.snapshot.geo.blocked ? 'geo-blocked' : 'geo-ok'">
              {{
                healthStore.snapshot.geo.blocked
                  ? `⚠ ${$t('health.blocked')} · ${healthStore.snapshot.geo.country} · ${healthStore.snapshot.geo.ip}`
                  : `${$t('health.allowed')} · ${healthStore.snapshot.geo.country}`
              }}
            </span>
          </div>
          <div class="health-updated">{{ updatedAgo }}</div>
        </template>
        <div v-else class="health-empty muted">{{ $t('health.never') }}</div>
      </div>

      <!-- Wallet summary -->
      <template v-if="wallets.length">
        <div class="section-header anim-in">Wallets</div>
        <div class="panel anim-in">
          <!-- Aggregate row -->
          <div class="wallet-agg">
            <div class="agg-item">
              <span class="agg-label">TOTAL BAL</span>
              <span class="agg-val num-glow">${{ fmt2(totalBalance) }}</span>
            </div>
            <div class="agg-sep">│</div>
            <div class="agg-item">
              <span class="agg-label">TOTAL P&amp;L</span>
              <span class="agg-val" :class="totalPnL >= 0 ? 'num-success' : 'num-danger'">
                {{ totalPnL >= 0 ? '+' : '' }}{{ fmt2(totalPnL) }}
              </span>
            </div>
            <div class="agg-sep">│</div>
            <div class="agg-item">
              <span class="agg-label">ACTIVE</span>
              <span class="agg-val">{{ activeWallets }}/{{ wallets.length }}</span>
            </div>
          </div>
          <!-- Table -->
          <table class="data-table">
            <thead>
              <tr>
                <th>LABEL</th><th>ADDRESS</th><th>BALANCE</th><th>P&amp;L</th><th>STATUS</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="w in wallets" :key="w.id">
                <td>{{ w.label || w.id }}</td>
                <td class="mono addr-cell">{{ w.address ? w.address.slice(0, 8) + '…' + w.address.slice(-4) : '—' }}</td>
                <td class="mono">${{ fmt2(w.balance_usd) }}</td>
                <td class="mono" :class="w.pnl_usd >= 0 ? 'num-success' : 'num-danger'">
                  {{ w.pnl_usd >= 0 ? '+' : '' }}{{ fmt2(w.pnl_usd) }}
                </td>
                <td>
                  <span class="status-dot" :class="w.enabled ? 'status-dot--on' : 'status-dot--off'" />
                  <span class="status-text">{{ w.enabled ? 'ON' : 'OFF' }}</span>
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
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { useHealthStore } from '@/stores/health'
import { useApi } from '@/composables/useApi'
import { createChart } from 'lightweight-charts'

const { t } = useI18n()
const app = useAppStore()
const { overview, walletsMap } = storeToRefs(app)
const api = useApi()
const healthStore = useHealthStore()

const wallets = computed(() => Object.values(walletsMap.value))
const totalBalance = computed(() => overview.value.balance || 0)
const totalPnL = computed(() => overview.value.pnl || 0)
const activeWallets = computed(() => wallets.value.filter(w => w.enabled).length)
const activeCount = computed(() => overview.value.subsystems?.filter(s => s.active).length ?? 0)

const showWelcome = ref(true)
const chartEl = ref(null)
let chart = null
let series = null
const pnlHistory = []

function fmt2(n) {
  if (n === null || n === undefined || isNaN(n)) return '---'
  return (+(n || 0)).toFixed(2)
}

const kpis = computed(() => [
  { label: 'WALLET', value: (overview.value.wallet_address || overview.value.wallet || '—').slice(0,10)+'…', cls: 'mono-sm', sub: null },
  { label: 'BALANCE',   value: '$' + fmt2(overview.value.balance), cls: 'num-glow' },
  { label: 'OPEN ORDERS',  value: overview.value.orders?.length ?? 0, cls: 'val-neutral' },
  { label: 'POSITIONS', value: overview.value.positions?.length ?? 0, cls: 'val-neutral' },
  { label: 'SUBSYSTEMS', value: activeCount.value + '/' + (overview.value.subsystems?.length ?? 0), cls: 'val-neutral', sub: 'online' },
])

const updatedAgo = computed(() => {
  if (!healthStore.snapshot?.updated_at) return ''
  const s = Math.floor((Date.now() - new Date(healthStore.snapshot.updated_at)) / 1000)
  return t('health.updated').replace('{s}', s)
})

function fmtLatency(ms) {
  if (!ms && ms !== 0) return ''
  return ms < 1000 ? `${ms}ms` : `${(ms / 1000).toFixed(1)}s`
}

onMounted(async () => {
  try { app.overview = await api.getOverview() } catch {}
  healthStore.fetchHealth()
  setTimeout(() => { showWelcome.value = false }, 2200)
})

watch(showWelcome, (val) => {
  if (!val) setTimeout(() => initChart(), 80)
})

watch(() => overview.value.balance, (bal) => {
  if (bal == null) return
  const t = Math.floor(Date.now() / 1000)
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
    height: 150,
    layout: { background: { color: 'transparent' }, textColor: '#3d6080' },
    grid: { vertLines: { color: '#0e1a2a' }, horzLines: { color: '#0e1a2a' } },
    rightPriceScale: { borderColor: '#162035' },
    timeScale: { borderColor: '#162035', timeVisible: true },
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
.view { display: flex; flex-direction: column; gap: 1.1rem; }

/* Boot terminal */
.boot-terminal {
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-top: 2px solid var(--accent);
  border-radius: var(--radius);
  padding: 1.5rem;
  font-family: var(--font-mono);
  display: flex;
  flex-direction: column;
  gap: 0.4rem;
  overflow: hidden;
  position: relative;
}

.boot-terminal::after {
  content: '';
  position: absolute;
  inset: 0;
  background: repeating-linear-gradient(
    0deg, transparent, transparent 3px,
    rgba(124,58,237,0.012) 3px, rgba(124,58,237,0.012) 6px
  );
  pointer-events: none;
}

.boot-line {
  font-size: 0.96rem;
  overflow: hidden;
  white-space: nowrap;
  animation: typewriter 0.5s steps(60) both;
  opacity: 0;
  animation-fill-mode: forwards;
}

.bl-1 { font-size: 1.05rem; font-weight: 700; color: var(--accent-bright); text-shadow: 0 0 10px rgba(124,58,237,0.4); animation-delay: 0.1s; }
.bl-2 { color: var(--text-muted); font-size: 0.94rem; animation-delay: 0.45s; }
.bl-3 { color: var(--text-secondary); animation-delay: 0.75s; }
.bl-4 { animation-delay: 1.05s; }
.bl-5 { animation-delay: 1.35s; }

.bval { color: var(--text-primary); margin-left: 1rem; font-weight: 600; }

/* KPI row */
.kpi-row {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(160px, 1fr));
  gap: 0.6rem;
}

.kpi-card {
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-top: 2px solid rgba(124, 58, 237, 0.25);
  border-radius: var(--radius);
  padding: 0.9rem 1rem;
  transition: border-top-color var(--transition);
}
.kpi-card:hover { border-top-color: var(--accent); }

.kpi-label {
  font-size: 0.86rem;
  text-transform: uppercase;
  letter-spacing: 0.12em;
  color: var(--text-secondary);
  margin-bottom: 0.3rem;
}
.kpi-value {
  font-size: 1.25rem;
  font-weight: 700;
  font-family: var(--font-mono);
  color: var(--text-primary);
  line-height: 1.2;
}
.kpi-sub { font-size: 1.00rem; color: var(--text-muted); margin-top: 0.2rem; }
.val-neutral { color: var(--text-bright); }
.mono-sm { font-size: 0.86rem; word-break: break-all; color: var(--text-secondary); }

/* Mid row */
.mid-row {
  display: flex;
  gap: 0.6rem;
}
.flex-2 { flex: 2; min-width: 0; }
.flex-1 { flex: 1; min-width: 0; }

/* Panel */
.panel {
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-top: 1px solid var(--accent);
  border-radius: var(--radius);
  overflow: hidden;
}

.panel-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0.4rem 1rem;
  border-bottom: 1px solid var(--border);
  font-size: 1.00rem;
  text-transform: uppercase;
  letter-spacing: 0.10em;
  color: var(--accent);
  background: rgba(124, 58, 237, 0.03);
}

.ph-right { color: var(--text-secondary); font-size: 0.90rem; }

.chart-canvas { width: 100%; }

/* Subsystems */
.subsystem-list { display: flex; flex-direction: column; }

.sub-row {
  display: flex;
  align-items: center;
  gap: 0.65rem;
  padding: 0.55rem 1rem;
  border-bottom: 1px solid var(--border-subtle);
  transition: background var(--transition);
}
.sub-row:last-child { border-bottom: none; }
.sub-row:hover { background: var(--bg-hover); }

.sub-name { flex: 1; font-size: 0.96rem; color: var(--text-primary); }

/* Wallet summary */
.wallet-agg {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  flex-wrap: wrap;
  padding: 0.6rem 1rem;
  border-bottom: 1px solid var(--border);
  font-size: 0.94rem;
}

.agg-item { display: flex; align-items: center; gap: 0.4rem; }
.agg-label { font-size: 0.86rem; text-transform: uppercase; letter-spacing: 0.10em; color: var(--text-secondary); }
.agg-val { font-weight: 700; font-family: var(--font-mono); color: var(--text-primary); }
.agg-sep { color: var(--border); user-select: none; }

.status-text { font-size: 0.86rem; color: var(--text-secondary); margin-left: 0.3rem; }

.mono { font-family: var(--font-mono); font-size: 0.92rem; }
.empty-state { padding: 1.5rem; text-align: center; color: var(--text-muted); font-size: 0.92rem; }

/* Section header */
.section-header {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 1.00rem;
  text-transform: uppercase;
  letter-spacing: 0.12em;
  color: var(--accent);
  font-weight: 600;
}
.section-header::after {
  content: '';
  flex: 1;
  height: 1px;
  background: linear-gradient(90deg, var(--border) 0%, transparent 100%);
}

/* Badge */
.badge {
  display: inline-flex;
  align-items: center;
  font-size: 0.86rem;
  font-weight: 600;
  padding: 0.18rem 0.55rem;
  border-radius: 1px;
  text-transform: uppercase;
  letter-spacing: 0.06em;
}
.badge--ok  { background: var(--success-dim); color: var(--success); border: 1px solid rgba(16,217,148,0.20); }
.badge--off { background: var(--badge-bg);    color: var(--text-muted); border: 1px solid var(--badge-border); }

/* Status dot */
.status-dot { display: inline-block; width: 6px; height: 6px; border-radius: 50%; flex-shrink: 0; }
.status-dot--on  { background: var(--success); box-shadow: 0 0 6px var(--success); animation: pulse-dot 2.5s ease infinite; }
.status-dot--off { background: var(--text-muted); }

/* Glows */
.num-glow    { color: var(--price-bright); text-shadow: 0 0 10px rgba(251,191,36,0.35); }
.num-success { color: var(--success); text-shadow: 0 0 8px rgba(16,217,148,0.30); }
.num-danger  { color: var(--danger);  text-shadow: 0 0 8px rgba(255,77,106,0.30); }

/* Data table */
.data-table { width: 100%; border-collapse: collapse; font-size: 0.96rem; }
.data-table th {
  padding: 0.6rem 1.2rem; text-align: left; font-size: 1.00rem;
  text-transform: uppercase; letter-spacing: 0.10em;
  color: var(--text-secondary); border-bottom: 1px solid var(--border);
  background: rgba(124, 58, 237, 0.03);
}
.data-table td { padding: 0.6rem 1.2rem; border-bottom: 1px solid var(--border-subtle); }
.data-table tr:nth-child(even) td { background: rgba(124,58,237,0.02); }
.data-table tr:last-child td { border-bottom: none; }
.data-table tr:hover td { background: var(--bg-hover) !important; }

/* Health card */
.health-panel { padding: 0.75rem 1rem 0.5rem; }
.health-row { display: flex; align-items: center; gap: 8px; padding: 3px 0; font-size: 13px; font-family: var(--font-mono); }
.health-dot { font-size: 10px; }
.health-dot.ok { color: var(--success); }
.health-dot.degraded { color: var(--warning); }
.health-dot.down { color: var(--danger); }
.health-name { width: 90px; color: var(--text-muted); }
.health-lat.ok { color: var(--success); }
.health-lat.degraded { color: var(--warning); }
.health-lat.down { color: var(--danger); }
.health-err { color: var(--danger); font-size: 11px; flex: 1; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.geo-blocked { color: var(--danger); }
.geo-ok { color: var(--success); }
.health-updated { font-size: 11px; color: var(--text-muted); margin-top: 6px; }
.health-empty { font-size: 13px; }
.muted { color: var(--text-muted); }

@media (max-width: 800px) {
  .mid-row { flex-direction: column; }
}
</style>
