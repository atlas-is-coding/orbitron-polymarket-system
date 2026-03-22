<template>
  <div class="strategies-view">
    <!-- KPI row -->
    <div class="kpi-row anim-in">
      <div class="kpi-card"><div class="kpi-lbl">ACTIVE / TOTAL</div><div class="kpi-val">{{ activeCount }}/{{ strategies.length }}</div></div>
      <div class="kpi-card"><div class="kpi-lbl">TOTAL P&L</div><div class="kpi-val" :class="totalPnl>=0?'pos-val':'neg-val'">${{ fmt(totalPnl) }}</div></div>
      <div class="kpi-card"><div class="kpi-lbl">TOTAL TRADES</div><div class="kpi-val">{{ totalTrades }}</div></div>
      <div class="kpi-card"><div class="kpi-lbl">WIN RATE</div><div class="kpi-val">{{ avgWinRate }}%</div></div>
    </div>

    <!-- Cards grid -->
    <div class="cards-grid anim-in">
      <div v-for="s in strategies" :key="s.name" class="strategy-card">
        <!-- Header -->
        <div class="card-header">
          <div class="strat-icon" :style="iconStyle(s)">{{ (s.name||'?')[0].toUpperCase() }}</div>
          <div class="strat-meta">
            <div class="strat-name">{{ s.name }}</div>
            <div class="strat-type muted-txt">{{ s.type || s.wallet_label || 'Strategy' }}</div>
          </div>
          <span class="status-pill" :class="statusPillClass(s)">{{ statusLabel(s) }}</span>
          <label class="toggle" @click.prevent="toggleStrategy(s)">
            <input type="checkbox" :checked="isRunning(s)" @change.prevent />
            <span class="toggle-track"><span class="toggle-thumb" /></span>
          </label>
        </div>

        <!-- Description -->
        <div v-if="s.description || s.details" class="strat-desc muted-txt">{{ s.description || s.details }}</div>

        <!-- Metrics -->
        <div class="metrics-row">
          <div class="metric"><div class="metric-lbl">P&L</div><div class="metric-val" :class="(s.pnl_usd||0)>=0?'pos-val':'neg-val'">${{ fmt(s.pnl_usd) }}</div></div>
          <div class="metric"><div class="metric-lbl">TRADES</div><div class="metric-val">{{ s.trades_count || 0 }}</div></div>
          <div class="metric"><div class="metric-lbl">WIN RATE</div><div class="metric-val">{{ s.win_rate ? fmt(s.win_rate)+'%' : '—' }}</div></div>
        </div>

        <!-- Daily limit bar -->
        <div v-if="s.max_daily_trades" class="daily-bar">
          <div class="daily-label"><span class="muted-txt">DAILY</span> {{ s.daily_trade_count || 0 }}/{{ s.max_daily_trades }}</div>
          <div class="bar-track"><div class="bar-fill" :style="{ width: dailyPct(s)+'%' }"></div></div>
        </div>

        <!-- Footer -->
        <div class="card-footer">
          <button class="btn btn-ghost sm" @click="configuring=s">CONFIGURE</button>
          <button class="btn btn-ghost sm" @click="router.push('/orders?strategy='+s.name)">TRADES</button>
          <button
            class="btn sm"
            :class="isRunning(s) ? 'btn-danger' : 'btn-success'"
            :disabled="togglingId===s.name"
            @click="toggleStrategy(s)"
          >{{ togglingId===s.name ? '...' : isRunning(s) ? 'STOP' : 'START' }}</button>
        </div>
      </div>
    </div>

    <div v-if="!strategies.length" class="empty-state">No strategies configured</div>

    <StrategyConfigDialog v-if="configuring" :strategy="configuring" @close="configuring=null" />
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { storeToRefs } from 'pinia'
import { useAppStore } from '@/stores/app'
import { useApi } from '@/composables/useApi'
import StrategyConfigDialog from '@/components/strategies/StrategyConfigDialog.vue'

const router = useRouter()
const app = useAppStore()
const { strategies } = storeToRefs(app)
const api = useApi()

const togglingId = ref(null)
const configuring = ref(null)

function isRunning(s) { return (s.status||'').toLowerCase() === 'running' || s.enabled === true }
function statusLabel(s) { const st = (s.status||'').toLowerCase(); if(st==='running') return 'RUNNING'; if(st==='error') return 'ERROR'; return 'STOPPED' }
function statusPillClass(s) { const st=(s.status||'').toLowerCase(); if(st==='running') return 'pill-on'; if(st==='error') return 'pill-warn'; return 'pill-off' }
function fmt(n) { return n!=null ? Number(n||0).toFixed(2) : '—' }
function dailyPct(s) { return s.max_daily_trades ? Math.min(100, (s.daily_trade_count||0)/s.max_daily_trades*100) : 0 }

const ICON_COLORS = ['#7c3aed','#2563eb','#0891b2','#059669']
function iconStyle(s) {
  const idx = (s.name||'').charCodeAt(0) % ICON_COLORS.length
  return { background: ICON_COLORS[idx] }
}

const activeCount = computed(() => strategies.value.filter(isRunning).length)
const totalPnl    = computed(() => strategies.value.reduce((s,x)=>s+(+x.pnl_usd||0),0))
const totalTrades = computed(() => strategies.value.reduce((s,x)=>s+(+x.trades_count||0),0))
const avgWinRate  = computed(() => {
  const ws = strategies.value.filter(x=>x.win_rate!=null)
  return ws.length ? Math.round(ws.reduce((s,x)=>s+(+x.win_rate),0)/ws.length) : 0
})

async function toggleStrategy(s) {
  if (togglingId.value) return
  togglingId.value = s.name
  try {
    if (isRunning(s)) {
      await api.stopStrategy(s.name)
      s.status = 'stopped'; s.enabled = false
    } else {
      await api.startStrategy(s.name, s.wallet_id ? [s.wallet_id] : [])
      s.status = 'running'; s.enabled = true
    }
    app.toast(`Strategy ${isRunning(s)?'stopped':'started'}`, 'success')
  } catch (e) { app.toast(e?.response?.data?.error || 'Toggle failed', 'error') }
  togglingId.value = null
}
</script>

<style scoped>
.strategies-view { display: flex; flex-direction: column; gap: 16px; }

.kpi-row { display: grid; grid-template-columns: repeat(auto-fit, minmax(140px, 1fr)); gap: 10px; }
.kpi-card { background: var(--bg-card); border: 1px solid var(--border); border-radius: var(--radius); padding: 12px 16px; }
.kpi-lbl { font-size: 10px; text-transform: uppercase; letter-spacing: 0.10em; color: var(--text-secondary); margin-bottom: 6px; }
.kpi-val { font-size: 20px; font-weight: 700; color: var(--text-primary); font-family: var(--font-mono); }

.cards-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(340px, 1fr)); gap: 14px; }

.strategy-card { background: var(--bg-card); border: 1px solid var(--border); border-radius: var(--radius-xl); padding: 16px; display: flex; flex-direction: column; gap: 12px; transition: border-color 0.15s; }
.strategy-card:hover { border-color: var(--accent); }

.card-header { display: flex; align-items: center; gap: 10px; }
.strat-icon { width: 36px; height: 36px; border-radius: 6px; display: flex; align-items: center; justify-content: center; font-size: 16px; font-weight: 700; color: #fff; flex-shrink: 0; }
.strat-meta { flex: 1; min-width: 0; }
.strat-name { font-size: 13px; font-weight: 600; color: var(--text-primary); }
.strat-type { font-size: 11px; }
.muted-txt { color: var(--text-secondary); }
.status-pill { font-size: 10px; font-weight: 700; padding: 2px 8px; border-radius: 3px; border: 1px solid; letter-spacing: 0.05em; flex-shrink: 0; }
.pill-on   { color: var(--success); border-color: rgba(16,217,148,0.40); background: rgba(16,217,148,0.08); }
.pill-off  { color: var(--text-secondary); border-color: var(--border); background: transparent; }
.pill-warn { color: var(--warning); border-color: rgba(245,158,11,0.40); background: rgba(245,158,11,0.08); }

.strat-desc { font-size: 12px; line-height: 1.5; }

.metrics-row { display: flex; gap: 20px; }
.metric { display: flex; flex-direction: column; gap: 2px; }
.metric-lbl { font-size: 9px; text-transform: uppercase; letter-spacing: 0.10em; color: var(--text-secondary); }
.metric-val { font-size: 14px; font-weight: 600; font-family: var(--font-mono); color: var(--text-primary); }

.daily-bar { display: flex; flex-direction: column; gap: 4px; }
.daily-label { font-size: 10px; color: var(--text-secondary); }
.bar-track { height: 4px; background: var(--bg-hover); border-radius: 2px; overflow: hidden; }
.bar-fill { height: 100%; background: var(--accent); border-radius: 2px; transition: width 0.3s ease; }

.card-footer { display: flex; gap: 6px; flex-wrap: wrap; margin-top: auto; }

.empty-state { padding: 2.5rem; text-align: center; color: var(--text-muted); font-size: 12px; }

.pos-val { color: var(--success); }
.neg-val { color: var(--danger); }

.btn { display: inline-flex; align-items: center; padding: 5px 12px; border-radius: var(--radius); font-family: var(--font-mono); font-size: 12px; font-weight: 500; cursor: pointer; border: 1px solid transparent; transition: all var(--transition); white-space: nowrap; }
.btn.sm { padding: 3px 9px; font-size: 11px; }
.btn-ghost   { background: none; border-color: var(--border); color: var(--text-secondary); }
.btn-ghost:hover:not(:disabled) { background: var(--bg-hover); color: var(--text-primary); }
.btn-danger  { background: var(--danger-dim); color: var(--danger); border-color: var(--danger); }
.btn-danger:hover:not(:disabled) { background: var(--danger); color: #fff; }
.btn-success { background: var(--success-dim); color: var(--success); border-color: rgba(16,217,148,0.40); }
.btn-success:hover:not(:disabled) { background: var(--success); color: #000; }
.btn:disabled { opacity: 0.4; cursor: not-allowed; }

.toggle { display: inline-flex; align-items: center; cursor: pointer; }
.toggle input { display: none; }
.toggle-track { width: 36px; height: 18px; background: var(--border); border-radius: 1px; position: relative; transition: background var(--transition); border: 1px solid var(--border); }
.toggle input:checked ~ .toggle-track { background: var(--accent); border-color: var(--accent); }
.toggle-thumb { position: absolute; width: 12px; height: 12px; background: var(--text-muted); border-radius: 1px; top: 2px; left: 2px; transition: left var(--transition), background var(--transition); }
.toggle input:checked ~ .toggle-track .toggle-thumb { left: 20px; background: #fff; }
</style>
