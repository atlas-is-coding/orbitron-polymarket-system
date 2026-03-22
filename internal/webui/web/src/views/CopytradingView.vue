<template>
  <div class="ct-view">
    <!-- Left column: trader list -->
    <div class="ct-left">
      <!-- Add trader -->
      <div class="add-trader-bar">
        <input v-model="newAddr" class="field-input" placeholder="Polymarket address (0x...)" @keyup.enter="addTrader" />
        <button class="btn btn-primary" :disabled="adding || !newAddr" @click="addTrader">
          {{ adding ? '...' : 'ADD' }}
        </button>
      </div>

      <!-- Trader cards -->
      <div v-if="traders.length" class="trader-list">
        <div
          v-for="t in traders"
          :key="t.address"
          class="trader-card"
          :class="{ selected: selectedTrader?.address === t.address }"
          @click="selectedTrader = t"
        >
          <div class="tc-header">
            <div class="tc-avatar">{{ (t.label || t.address || '?')[0].toUpperCase() }}</div>
            <div class="tc-meta">
              <div class="tc-name">{{ t.label || shortAddr(t.address) }}</div>
              <div class="tc-addr mono muted-txt">{{ shortAddr(t.address) }}</div>
            </div>
            <span class="status-pill" :class="ctStatusClass(t.status)">{{ t.status || 'STOPPED' }}</span>
          </div>
          <div class="tc-metrics">
            <div class="tc-m"><span class="tc-ml">30D ROI</span><span class="tc-mv" :class="(t.roi_30d||0)>=0?'pos-val':'neg-val'">{{ fmt(t.roi_30d) }}%</span></div>
            <div class="tc-m"><span class="tc-ml">WIN RATE</span><span class="tc-mv">{{ t.win_rate ? fmt(t.win_rate)+'%' : '—' }}</span></div>
            <div class="tc-m"><span class="tc-ml">COPY P&L</span><span class="tc-mv" :class="(t.copy_pnl||0)>=0?'pos-val':'neg-val'">${{ fmt(t.copy_pnl) }}</span></div>
          </div>
          <div class="tc-footer">
            <button class="btn btn-ghost sm" :disabled="togglingAddr===t.address" @click.stop="doToggle(t)">
              {{ togglingAddr===t.address ? '...' : t.status==='COPYING' ? 'PAUSE' : 'RESUME' }}
            </button>
            <button class="btn btn-ghost sm danger-x" @click.stop="doRemove(t.address)">REMOVE</button>
          </div>
        </div>
      </div>

      <div v-else class="empty-state">No traders configured. Add one above.</div>
    </div>

    <!-- Right column: detail + feed -->
    <div class="ct-right">
      <!-- Detail / settings panel -->
      <div class="detail-panel" v-if="selectedTrader">
        <div class="panel-hdr">COPY SETTINGS — {{ selectedTrader.label || shortAddr(selectedTrader.address) }}</div>
        <div class="settings-form">
          <div class="field-group">
            <label class="field-label">Size Mode</label>
            <select v-model="cfg.size_mode" class="field-input">
              <option value="FIXED">FIXED</option>
              <option value="RATIO">RATIO</option>
              <option value="SCALE">SCALE</option>
            </select>
          </div>
          <div class="field-group">
            <label class="field-label">Fixed Size (USD)</label>
            <input v-model.number="cfg.fixed_size" type="number" class="field-input" />
          </div>
          <div class="field-group">
            <label class="field-label">Max Daily Exposure (USD)</label>
            <input v-model.number="cfg.max_daily_exposure" type="number" class="field-input" />
          </div>
          <div class="field-group">
            <label class="field-label">Min Size (USD)</label>
            <input v-model.number="cfg.min_size" type="number" class="field-input" />
          </div>
          <div class="field-group">
            <label class="field-label">Max Size (USD)</label>
            <input v-model.number="cfg.max_size" type="number" class="field-input" />
          </div>
          <div class="field-group toggle-row">
            <label class="field-label">YES only</label>
            <label class="toggle">
              <input type="checkbox" v-model="cfg.yes_only" />
              <span class="toggle-track"><span class="toggle-thumb" /></span>
            </label>
          </div>
          <button class="btn btn-primary" :disabled="saving" @click="saveConfig">{{ saving ? '...' : 'SAVE CONFIG' }}</button>
        </div>
      </div>
      <div v-else class="detail-panel empty-detail">
        <div class="empty-state">Select a trader to configure</div>
      </div>

      <!-- Live feed -->
      <div class="feed-panel">
        <div class="panel-hdr">LIVE FEED</div>
        <div class="feed-list" ref="feedEl">
          <div v-for="(ev, i) in copyTrades" :key="i" class="feed-item">
            <span class="feed-dot" :class="feedDotClass(ev.type)"></span>
            <span class="feed-time muted-txt">{{ fmtTime(ev.timestamp) }}</span>
            <span class="feed-type" :class="feedTypeClass(ev.type)">{{ ev.type }}</span>
            <span class="feed-market muted-txt">{{ ev.market?.slice(0,30) }}</span>
            <span v-if="ev.side" class="feed-side side-pill" :class="ev.side==='YES'?'side--yes':'side--no'">{{ ev.side }}</span>
          </div>
          <div v-if="!copyTrades.length" class="empty-state">No events yet</div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch, nextTick } from 'vue'
import { storeToRefs } from 'pinia'
import { useAppStore } from '@/stores/app'
import { useApi } from '@/composables/useApi'

const app = useAppStore()
const { copytrading, copyTrades } = storeToRefs(app)
const api = useApi()

const traders = computed(() => copytrading.value?.traders || [])
const selectedTrader = ref(null)
const togglingAddr = ref(null)
const adding = ref(false)
const newAddr = ref('')
const saving = ref(false)
const feedEl = ref(null)

const cfg = ref({ size_mode: 'FIXED', fixed_size: 10, max_daily_exposure: 100, min_size: 1, max_size: 50, yes_only: false })

watch(selectedTrader, t => {
  if (t) cfg.value = {
    size_mode: t.size_mode || 'FIXED',
    fixed_size: t.fixed_size || 10,
    max_daily_exposure: t.max_daily_exposure || 100,
    min_size: t.min_size || 1,
    max_size: t.max_size || 50,
    yes_only: t.yes_only || false,
  }
})

watch(copyTrades, async () => {
  await nextTick()
  if (feedEl.value) feedEl.value.scrollTop = feedEl.value.scrollHeight
})

function fmt(n) { return n != null ? Number(n||0).toFixed(2) : '—' }
function shortAddr(a) { return a ? a.slice(0,6)+'...'+a.slice(-4) : '—' }
function fmtTime(t) { if (!t) return ''; try { return new Date(t).toLocaleTimeString(undefined,{hour:'2-digit',minute:'2-digit',second:'2-digit'}) } catch { return '' } }
function ctStatusClass(s) { const u=(s||'').toUpperCase(); if(u==='COPYING') return 'pill-on'; if(u==='PAUSED') return 'pill-warn'; return 'pill-off' }
function feedDotClass(t) { if(t==='COPIED') return 'dot-green'; if(t==='SKIPPED') return 'dot-yellow'; return 'dot-grey' }
function feedTypeClass(t) { if(t==='COPIED') return 'pos-val'; if(t==='SKIPPED') return 'warn-val'; return 'muted-txt' }

async function addTrader() {
  if (!newAddr.value || adding.value) return
  adding.value = true
  try {
    await api.addTrader(newAddr.value, '', 0)
    app.toast('Trader added', 'success')
    newAddr.value = ''
  } catch (e) { app.toast(e?.response?.data?.error || 'Failed to add trader', 'error') }
  adding.value = false
}

async function doToggle(t) {
  togglingAddr.value = t.address
  try { await api.toggleTrader(t.address); app.toast('Trader toggled', 'success') } catch { app.toast('Failed', 'error') }
  togglingAddr.value = null
}

async function doRemove(addr) {
  if (!confirm('Remove this trader?')) return
  try { await api.removeTrader(addr); app.toast('Trader removed', 'success') } catch { app.toast('Failed', 'error') }
}

async function saveConfig() {
  if (!selectedTrader.value || saving.value) return
  saving.value = true
  try {
    await api.updateTrader(selectedTrader.value.address, cfg.value)
    app.toast('Config saved', 'success')
  } catch (e) { app.toast(e?.response?.data?.error || 'Save failed', 'error') }
  saving.value = false
}
</script>

<style scoped>
.ct-view { display: flex; gap: 16px; height: calc(100vh - 120px); }

.ct-left { flex: 1; display: flex; flex-direction: column; gap: 12px; overflow-y: auto; min-width: 0; }
.ct-right { width: 380px; flex-shrink: 0; display: flex; flex-direction: column; gap: 12px; overflow-y: auto; }

@media (max-width: 768px) {
  .ct-view { flex-direction: column; height: auto; }
  .ct-right { width: 100%; }
}

.add-trader-bar { display: flex; gap: 8px; }

.trader-list { display: flex; flex-direction: column; gap: 10px; }

.trader-card { background: var(--bg-card); border: 1px solid var(--border); border-radius: var(--radius-xl); padding: 14px; cursor: pointer; display: flex; flex-direction: column; gap: 10px; transition: border-color 0.15s; }
.trader-card:hover, .trader-card.selected { border-color: var(--accent); }

.tc-header { display: flex; align-items: center; gap: 10px; }
.tc-avatar { width: 32px; height: 32px; border-radius: 50%; background: var(--accent); display: flex; align-items: center; justify-content: center; font-size: 14px; font-weight: 700; color: #fff; flex-shrink: 0; }
.tc-meta { flex: 1; min-width: 0; }
.tc-name { font-size: 13px; font-weight: 600; color: var(--text-primary); }
.tc-addr { font-size: 11px; }
.mono { font-family: var(--font-mono); }
.muted-txt { color: var(--text-secondary); }

.status-pill { font-size: 10px; font-weight: 700; padding: 2px 8px; border-radius: 3px; border: 1px solid; letter-spacing: 0.05em; flex-shrink: 0; }
.pill-on   { color: var(--success); border-color: rgba(16,217,148,0.40); background: rgba(16,217,148,0.08); }
.pill-off  { color: var(--text-secondary); border-color: var(--border); background: transparent; }
.pill-warn { color: var(--warning); border-color: rgba(245,158,11,0.40); background: rgba(245,158,11,0.08); }

.tc-metrics { display: flex; gap: 16px; }
.tc-m { display: flex; flex-direction: column; gap: 2px; }
.tc-ml { font-size: 9px; text-transform: uppercase; letter-spacing: 0.10em; color: var(--text-secondary); }
.tc-mv { font-size: 13px; font-weight: 600; font-family: var(--font-mono); color: var(--text-primary); }

.tc-footer { display: flex; gap: 6px; }
.danger-x { color: var(--danger) !important; border-color: rgba(255,77,106,0.30) !important; }

.detail-panel, .feed-panel { background: var(--bg-card); border: 1px solid var(--border); border-radius: var(--radius-xl); overflow: hidden; }
.detail-panel { flex-shrink: 0; }
.feed-panel { flex: 1; min-height: 200px; display: flex; flex-direction: column; }
.empty-detail { min-height: 100px; display: flex; align-items: center; justify-content: center; }

.panel-hdr { padding: 10px 14px; font-size: 10px; text-transform: uppercase; letter-spacing: 0.10em; color: var(--accent); border-bottom: 1px solid var(--border); background: rgba(124,58,237,0.03); }
.settings-form { padding: 14px; display: flex; flex-direction: column; gap: 12px; }
.field-group { display: flex; flex-direction: column; gap: 5px; }
.field-label { font-size: 10px; text-transform: uppercase; letter-spacing: 0.08em; color: var(--text-secondary); }
.toggle-row { flex-direction: row; align-items: center; justify-content: space-between; }

.feed-list { flex: 1; overflow-y: auto; padding: 8px; display: flex; flex-direction: column; gap: 4px; }
.feed-item { display: flex; align-items: center; gap: 8px; font-size: 11px; padding: 3px 6px; border-radius: 3px; }
.feed-item:hover { background: var(--bg-hover); }
.feed-dot { width: 6px; height: 6px; border-radius: 50%; flex-shrink: 0; }
.dot-green  { background: var(--success); }
.dot-yellow { background: var(--warning); }
.dot-grey   { background: var(--text-muted); }
.feed-time { color: var(--text-secondary); flex-shrink: 0; font-family: var(--font-mono); }
.feed-type { font-weight: 700; flex-shrink: 0; font-size: 10px; letter-spacing: 0.05em; }
.feed-market { flex: 1; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }

.side-pill { display: inline-block; font-size: 10px; font-weight: 700; padding: 1px 6px; border-radius: 2px; }
.side--yes { background: rgba(16,217,148,0.12); color: var(--success); border: 1px solid rgba(16,217,148,0.25); }
.side--no  { background: rgba(255,77,106,0.12); color: var(--danger);  border: 1px solid rgba(255,77,106,0.25); }

.empty-state { padding: 2rem; text-align: center; color: var(--text-muted); font-size: 12px; }
.pos-val { color: var(--success); }
.neg-val { color: var(--danger); }
.warn-val { color: var(--warning); }

.btn { display: inline-flex; align-items: center; padding: 5px 12px; border-radius: var(--radius); font-family: var(--font-mono); font-size: 12px; font-weight: 500; cursor: pointer; border: 1px solid transparent; transition: all var(--transition); white-space: nowrap; }
.btn.sm { padding: 3px 9px; font-size: 11px; }
.btn-ghost   { background: none; border-color: var(--border); color: var(--text-secondary); }
.btn-ghost:hover:not(:disabled) { background: var(--bg-hover); color: var(--text-primary); }
.btn-primary { background: var(--accent); color: #fff; border-color: var(--accent); }
.btn:disabled { opacity: 0.4; cursor: not-allowed; }

.field-input {
  padding: 0.45rem 0.65rem; background: var(--bg-input); border: 1px solid var(--border);
  border-radius: var(--radius); color: var(--text-primary); font-family: var(--font-mono);
  font-size: 0.96rem; outline: none; transition: border-color var(--transition);
}
.field-input:focus { border-color: var(--accent); box-shadow: 0 0 0 1px rgba(124,58,237,0.15); }

.toggle { display: inline-flex; align-items: center; cursor: pointer; }
.toggle input { display: none; }
.toggle-track { width: 36px; height: 18px; background: var(--border); border-radius: 1px; position: relative; transition: background var(--transition); }
.toggle input:checked ~ .toggle-track { background: var(--accent); }
.toggle-thumb { position: absolute; width: 12px; height: 12px; background: var(--text-muted); border-radius: 1px; top: 3px; left: 3px; transition: left var(--transition), background var(--transition); }
.toggle input:checked ~ .toggle-track .toggle-thumb { left: 21px; background: #fff; }
</style>
