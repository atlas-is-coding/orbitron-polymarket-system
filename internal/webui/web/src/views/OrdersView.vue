<template>
  <div class="orders-view">
    <div class="tab-bar anim-in">
      <button v-for="t in tabs" :key="t.key" class="tab-btn" :class="{ active: activeTab === t.key }" @click="setTab(t.key)">
        {{ t.label }}
        <span v-if="t.count" class="tab-count">{{ t.count }}</span>
      </button>
    </div>

    <!-- ORDERS -->
    <template v-if="activeTab === 'orders'">
      <div class="kpi-row anim-in">
        <div class="kpi-card"><div class="kpi-lbl">OPEN</div><div class="kpi-val">{{ openCount }}</div></div>
        <div class="kpi-card"><div class="kpi-lbl">FILLED</div><div class="kpi-val">{{ filledCount }}</div></div>
        <div class="kpi-card"><div class="kpi-lbl">CANCELLED</div><div class="kpi-val">{{ cancelledCount }}</div></div>
        <div class="kpi-card"><div class="kpi-lbl">EXPOSURE</div><div class="kpi-val price-val">${{ totalExposure }}</div></div>
      </div>
      <div class="filter-row anim-in">
        <div class="ftabs">
          <button v-for="f in orderFilters" :key="f" class="ftab" :class="{on:orderFilter===f}" @click="orderFilter=f">{{ f }}</button>
        </div>
        <div class="ftabs">
          <button v-for="s in sides" :key="s" class="ftab" :class="{on:orderSide===s}" @click="orderSide=s">{{ s }}</button>
        </div>
        <input v-model="orderSearch" class="field-input srch" placeholder="Search..." />
        <button v-if="openCount > 0" class="btn btn-danger" :disabled="canceling" @click="confirmCancelAll=true">CANCEL ALL</button>
      </div>
      <div class="panel anim-in">
        <div v-if="loadingOrders" class="skels"><div v-for="i in 5" :key="i" class="skeleton skel-row" /></div>
        <template v-else>
          <table class="data-table">
            <thead><tr><th>Market</th><th>Side</th><th>Price</th><th>Size</th><th>Status</th><th>Created</th><th></th></tr></thead>
            <tbody>
              <tr v-for="o in filteredOrders" :key="o.id">
                <td class="mkt-cell">{{ o.market }}</td>
                <td><span class="side-pill" :class="sideClass(o.side)">{{ o.side }}</span></td>
                <td class="mono price-val">${{ fmt(o.price) }}</td>
                <td class="mono">{{ fmt(o.size) }}</td>
                <td><span class="st-badge" :class="stClass(o.status)">{{ o.status }}</span></td>
                <td class="mono muted-txt">{{ fmtTime(o.created_at) }}</td>
                <td class="act-cell">
                  <button v-if="o.status==='OPEN'" class="btn-x" :disabled="cancelingId===o.id" @click="doCancel(o.id)">{{ cancelingId===o.id?'...':'CANCEL' }}</button>
                </td>
              </tr>
            </tbody>
          </table>
          <div v-if="!filteredOrders.length" class="empty-state">No orders</div>
        </template>
      </div>
    </template>

    <!-- POSITIONS -->
    <template v-else-if="activeTab === 'positions'">
      <div class="kpi-row anim-in">
        <div class="kpi-card"><div class="kpi-lbl">POSITIONS</div><div class="kpi-val">{{ positions.length }}</div></div>
        <div class="kpi-card"><div class="kpi-lbl">UNREALISED P&L</div><div class="kpi-val" :class="totalPnl>=0?'pos-val':'neg-val'">${{ fmt(totalPnl) }}</div></div>
        <div class="kpi-card"><div class="kpi-lbl">TOTAL VALUE</div><div class="kpi-val price-val">${{ fmt(totalValue) }}</div></div>
        <div class="kpi-card"><div class="kpi-lbl">WIN RATE</div><div class="kpi-val">{{ winRate }}%</div></div>
      </div>
      <div class="filter-row anim-in">
        <div class="ftabs">
          <button v-for="f in posFilters" :key="f" class="ftab" :class="{on:posFilter===f}" @click="posFilter=f">{{ f }}</button>
        </div>
        <input v-model="posSearch" class="field-input srch" placeholder="Search..." />
      </div>
      <div class="panel anim-in">
        <table class="data-table">
          <thead><tr><th>Market</th><th>Side</th><th>Size</th><th>Avg Price</th><th>Current</th><th>P&L</th><th></th></tr></thead>
          <tbody>
            <tr v-for="p in filteredPositions" :key="p.id">
              <td class="mkt-cell">{{ p.market }}</td>
              <td><span class="side-pill" :class="sideClass(p.side)">{{ p.side }}</span></td>
              <td class="mono">{{ fmt(p.size) }}</td>
              <td class="mono price-val">${{ fmt(p.avg_price) }}</td>
              <td class="mono">${{ fmt(p.current_price) }}</td>
              <td class="mono" :class="(p.pnl_usd||0)>=0?'pos-val':'neg-val'">${{ fmt(p.pnl_usd) }} <small>({{ fmt(p.pnl_pct) }}%)</small></td>
              <td class="act-cell">
                <button class="btn-x close-x" @click="closingPosition=p">CLOSE</button>
              </td>
            </tr>
          </tbody>
        </table>
        <div v-if="!filteredPositions.length" class="empty-state">No positions</div>
      </div>
    </template>

    <!-- TRADE HISTORY -->
    <template v-else-if="activeTab === 'history'">
      <div class="panel anim-in">
        <div v-if="loadingTrades" class="skels"><div v-for="i in 5" :key="i" class="skeleton skel-row" /></div>
        <template v-else>
          <table class="data-table">
            <thead><tr><th>Time</th><th>Market</th><th>Side</th><th>Type</th><th>Price</th><th>Size</th><th>Fee</th><th>P&L</th></tr></thead>
            <tbody>
              <tr v-for="t in trades" :key="t.id">
                <td class="mono muted-txt">{{ fmtTime(t.created_at||t.timestamp) }}</td>
                <td class="mkt-cell">{{ t.market }}</td>
                <td><span class="side-pill" :class="sideClass(t.side)">{{ t.side }}</span></td>
                <td class="mono">{{ t.type }}</td>
                <td class="mono price-val">${{ fmt(t.price) }}</td>
                <td class="mono">{{ fmt(t.size) }}</td>
                <td class="mono muted-txt">{{ fmt(t.fee) }}</td>
                <td class="mono" :class="(t.pnl||0)>=0?'pos-val':'neg-val'">${{ fmt(t.pnl) }}</td>
              </tr>
            </tbody>
          </table>
          <div v-if="!trades.length" class="empty-state">No trade history</div>
        </template>
      </div>
    </template>

    <ClosePositionDialog v-if="closingPosition" :position="closingPosition" @close="closingPosition=null" @closed="onPosClosed" />

    <div v-if="confirmCancelAll" class="overlay" @click.self="confirmCancelAll=false">
      <div class="dialog">
        <div class="dialog-title">CANCEL ALL ORDERS</div>
        <p class="dialog-body">Cancel {{ openCount }} open orders?</p>
        <div class="dialog-actions">
          <button class="btn btn-ghost" @click="confirmCancelAll=false">No</button>
          <button class="btn btn-danger" @click="doCancelAll">Yes</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { storeToRefs } from 'pinia'
import { useAppStore } from '@/stores/app'
import { useApi } from '@/composables/useApi'
import ClosePositionDialog from '@/components/orders/ClosePositionDialog.vue'

const route = useRoute()
const router = useRouter()
const app = useAppStore()
const { orders, positions } = storeToRefs(app)
const api = useApi()

const activeTab = ref(route.query.tab || 'orders')
watch(() => route.query.tab, v => { if (v) activeTab.value = v })
function setTab(k) { activeTab.value = k; router.replace({ query: k === 'orders' ? {} : { tab: k } }) }

const tabs = computed(() => [
  { key: 'orders',    label: 'ORDERS',        count: openCount.value || null },
  { key: 'positions', label: 'POSITIONS',     count: positions.value.length || null },
  { key: 'history',  label: 'TRADE HISTORY',  count: null },
])

// Orders
const loadingOrders = ref(true)
const orderFilter = ref('ALL'); const orderFilters = ['ALL','OPEN','FILLED','CANCELLED']
const orderSide = ref('ALL');   const sides = ['ALL','YES','NO']
const orderSearch = ref('')
const canceling = ref(false); const cancelingId = ref(null); const confirmCancelAll = ref(false)

const openCount      = computed(() => orders.value.filter(o=>o.status==='OPEN').length)
const filledCount    = computed(() => orders.value.filter(o=>o.status==='FILLED').length)
const cancelledCount = computed(() => orders.value.filter(o=>o.status==='CANCELLED').length)
const totalExposure  = computed(() => fmt(orders.value.filter(o=>o.status==='OPEN').reduce((s,o)=>s+(+o.size||0),0)))

const filteredOrders = computed(() => {
  let l = orders.value
  if (orderFilter.value !== 'ALL') l = l.filter(o=>o.status===orderFilter.value)
  if (orderSide.value !== 'ALL')   l = l.filter(o=>o.side===orderSide.value)
  if (orderSearch.value)           l = l.filter(o=>o.market?.toLowerCase().includes(orderSearch.value.toLowerCase()))
  return l
})

// Positions
const posFilter = ref('ALL'); const posFilters = ['ALL','WINNING','LOSING']
const posSearch = ref('')
const closingPosition = ref(null)

const filteredPositions = computed(() => {
  let l = positions.value
  if (posFilter.value === 'WINNING') l = l.filter(p=>(p.pnl_usd||0)>0)
  if (posFilter.value === 'LOSING')  l = l.filter(p=>(p.pnl_usd||0)<0)
  if (posSearch.value) l = l.filter(p=>p.market?.toLowerCase().includes(posSearch.value.toLowerCase()))
  return l
})
const totalPnl   = computed(() => positions.value.reduce((s,p)=>s+(+p.pnl_usd||0),0))
const totalValue = computed(() => positions.value.reduce((s,p)=>s+(+p.size||0)*(+p.current_price||0),0))
const winRate    = computed(() => !positions.value.length ? 0 : Math.round(positions.value.filter(p=>(p.pnl_usd||0)>0).length/positions.value.length*100))
function onPosClosed(id) { app.positions = positions.value.filter(p=>p.id!==id) }

// Trades
const trades = ref([]); const loadingTrades = ref(false)
watch(activeTab, async tab => {
  if (tab==='history' && !trades.value.length) {
    loadingTrades.value = true
    try { trades.value = await api.getTrades() } catch {}
    loadingTrades.value = false
  }
})

// Helpers
function fmt(n) { return n!=null ? Number(n||0).toFixed(2) : '—' }
function sideClass(s) { const u=(s||'').toUpperCase(); if(u==='YES'||u==='BUY') return 'side--yes'; if(u==='NO'||u==='SELL') return 'side--no'; return '' }
function stClass(s)   { const u=(s||'').toUpperCase(); if(u==='OPEN') return 'st--on'; if(u==='FILLED') return 'st--ok'; return 'st--off' }
function fmtTime(t)   { if(!t) return '—'; try { return new Date(t).toLocaleString(undefined,{month:'short',day:'numeric',hour:'2-digit',minute:'2-digit'}) } catch { return t } }

onMounted(async () => {
  try { app.orders = await api.getOrders() } catch {}
  loadingOrders.value = false
  if (route.query.tab === 'history') {
    loadingTrades.value = true
    try { trades.value = await api.getTrades() } catch {}
    loadingTrades.value = false
  }
})

async function doCancel(id) {
  cancelingId.value = id
  try { await api.cancelOrder(id); app.orders = orders.value.filter(o=>o.id!==id); app.toast('Order cancelled','success') } catch { app.toast('Cancel failed','error') }
  cancelingId.value = null
}
async function doCancelAll() {
  confirmCancelAll.value=false; canceling.value=true
  try { await api.cancelAll(); app.orders=[]; app.toast('All orders cancelled','success') } catch { app.toast('Failed','error') }
  canceling.value=false
}
</script>

<style scoped>
.orders-view { display:flex; flex-direction:column; gap:12px; }

.tab-bar { display:flex; gap:2px; border-bottom:1px solid var(--border); }
.tab-btn { padding:8px 18px; border:none; background:none; color:var(--text-secondary); font-family:var(--font-mono); font-size:12px; font-weight:600; letter-spacing:0.06em; cursor:pointer; border-bottom:2px solid transparent; margin-bottom:-1px; display:flex; align-items:center; gap:6px; }
.tab-btn:hover { color:var(--text-primary); }
.tab-btn.active { color:var(--accent-bright); border-bottom-color:var(--accent); }
.tab-count { font-size:10px; padding:1px 6px; border-radius:10px; background:var(--accent); color:#fff; }

.kpi-row { display:grid; grid-template-columns:repeat(auto-fit,minmax(140px,1fr)); gap:10px; }
.kpi-card { background:var(--bg-card); border:1px solid var(--border); border-radius:var(--radius); padding:12px 16px; }
.kpi-lbl { font-size:10px; text-transform:uppercase; letter-spacing:0.10em; color:var(--text-secondary); margin-bottom:6px; }
.kpi-val { font-size:20px; font-weight:700; color:var(--text-primary); font-family:var(--font-mono); }

.filter-row { display:flex; align-items:center; gap:8px; flex-wrap:wrap; }
.ftabs { display:flex; gap:2px; }
.ftab { padding:4px 10px; border-radius:var(--radius); border:1px solid var(--border); background:none; color:var(--text-secondary); font-size:11px; font-family:var(--font-mono); cursor:pointer; }
.ftab.on { background:var(--accent); color:#fff; border-color:var(--accent); }
.srch { flex:1; min-width:150px; max-width:240px; }

.panel { background:var(--bg-card); border:1px solid var(--border); border-top:1px solid var(--accent); border-radius:var(--radius); overflow-x:auto; }
.skels { padding:12px; display:flex; flex-direction:column; gap:8px; }
.skel-row { height:34px; }

.data-table { width:100%; border-collapse:collapse; font-size:12px; }
.data-table th { padding:9px 14px; text-align:left; font-size:10px; text-transform:uppercase; letter-spacing:0.10em; color:var(--text-secondary); border-bottom:1px solid var(--border); background:rgba(124,58,237,0.03); white-space:nowrap; }
.data-table td { padding:8px 14px; border-bottom:1px solid rgba(255,255,255,0.03); vertical-align:middle; }
.data-table tr:hover td { background:var(--bg-hover) !important; }
.data-table tr:last-child td { border-bottom:none; }

.mkt-cell { max-width:220px; overflow:hidden; text-overflow:ellipsis; white-space:nowrap; }
.mono { font-family:var(--font-mono); }
.muted-txt { color:var(--text-secondary); }
.price-val { color:var(--price-bright); }
.pos-val { color:var(--success); }
.neg-val { color:var(--danger); }
.act-cell { text-align:right; white-space:nowrap; }

.side-pill { display:inline-block; font-size:10px; font-weight:700; padding:2px 8px; border-radius:2px; }
.side--yes { background:rgba(16,217,148,0.12); color:var(--success); border:1px solid rgba(16,217,148,0.25); }
.side--no  { background:rgba(255,77,106,0.12); color:var(--danger);  border:1px solid rgba(255,77,106,0.25); }

.st-badge { font-size:10px; font-weight:600; padding:2px 8px; border-radius:2px; text-transform:uppercase; }
.st--on  { background:var(--accent-dim); color:var(--accent); border:1px solid rgba(124,58,237,0.20); }
.st--ok  { background:var(--success-dim); color:var(--success); border:1px solid rgba(16,217,148,0.20); }
.st--off { background:var(--badge-bg); color:var(--text-muted); border:1px solid var(--badge-border); }

.btn-x { background:none; border:1px solid rgba(255,77,106,0.30); color:var(--danger); border-radius:var(--radius); padding:3px 10px; font-size:10px; cursor:pointer; font-family:var(--font-mono); }
.btn-x:hover:not(:disabled) { background:var(--danger); color:#fff; }
.btn-x:disabled { opacity:0.4; cursor:not-allowed; }
.close-x { border-color:rgba(16,217,148,0.35); color:var(--success); }
.close-x:hover { background:var(--success); color:#000; }
.empty-state { padding:2.5rem; text-align:center; color:var(--text-muted); font-size:12px; }

.btn { display:inline-flex; align-items:center; padding:5px 12px; border-radius:var(--radius); font-family:var(--font-mono); font-size:12px; font-weight:500; cursor:pointer; border:1px solid transparent; }
.btn-ghost  { background:none; border-color:var(--border); color:var(--text-secondary); }
.btn-danger { background:var(--danger-dim); color:var(--danger); border-color:var(--danger); }
.btn-danger:hover:not(:disabled) { background:var(--danger); color:#fff; }
.btn:disabled { opacity:0.4; cursor:not-allowed; }

.overlay { position:fixed; inset:0; background:var(--bg-overlay); display:flex; align-items:center; justify-content:center; z-index:200; backdrop-filter:blur(4px); }
.dialog { background:var(--bg-card); border:1px solid var(--border); border-top:2px solid var(--accent); border-radius:var(--radius); padding:1.5rem; min-width:300px; }
.dialog-title { font-size:11px; font-weight:700; color:var(--accent); text-transform:uppercase; letter-spacing:0.08em; margin-bottom:0.75rem; }
.dialog-body { color:var(--text-secondary); font-size:13px; margin-bottom:1.25rem; }
.dialog-actions { display:flex; gap:8px; justify-content:flex-end; }
</style>
