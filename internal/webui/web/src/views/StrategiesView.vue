<template>
  <div class="view">
    <div class="page-header anim-in">
      <h2 class="view-title">{{ $t('nav.strategies') }}</h2>
      <p class="view-sub">Configure and launch automated trading strategies</p>
    </div>

    <!-- Strategy Cards Grid -->
    <div class="strategies-grid">
      <div
        v-for="(s, i) in displayStrategies"
        :key="s.name"
        class="strategy-card anim-in"
        :style="{ animationDelay: i * 60 + 'ms' }"
      >
        <div class="sc-header">
          <div class="sc-title">
            <span class="sc-icon">{{ s.icon }}</span>
            <h3>{{ s.label }}</h3>
          </div>
          <span class="sc-badge" :class="s.status === 'active' ? 'sc-badge--on' : 'sc-badge--off'">
            {{ s.status === 'active' ? 'ON' : 'OFF' }}
          </span>
        </div>
        <div class="sc-body">
          <p class="sc-fields-count">{{ s.details }}</p>
          <div v-if="s.walletLabel" class="sc-wallet-info">
            <p class="sc-wallet">Wallet: {{ s.walletLabel }}</p>
            <p v-if="s.wallet_address && s.wallet_address !== '—'" class="sc-addr">{{ s.wallet_address }}</p>
          </div>
        </div>
        <div class="sc-footer">
          <button class="btn-configure" @click="openDrawer(s.name)">CONFIGURE</button>
        </div>
      </div>
    </div>

    <!-- Drawer Backdrop -->
    <Transition name="fade-backdrop">
      <div v-if="activeKey" class="drawer-backdrop" @click="closeDrawer" />
    </Transition>

    <!-- Right Drawer -->
    <Transition name="slide-drawer">
      <div v-if="activeKey && activeUIStrategy" class="drawer">
        <div class="drawer-header">
          <div class="drawer-title">
            <span class="drawer-icon">{{ activeUIStrategy.icon }}</span>
            <span>{{ $t(activeUIStrategy.nameKey) }}</span>
          </div>
          <button class="drawer-close" @click="closeDrawer">✕</button>
        </div>

        <div class="drawer-body">
          <!-- Enabled toggle -->
          <div class="drawer-section">
            <div class="drawer-row-toggle">
              <span class="drawer-label">{{ $t('settings.tradingEnabled') }}</span>
              <label class="toggle">
                <input type="checkbox"
                  :checked="form[activeUIStrategy.enabledField]"
                  @change="form[activeUIStrategy.enabledField] = $event.target.checked; save(activeUIStrategy.configKey, $event.target.checked)"
                />
                <span class="toggle-track"><span class="toggle-thumb" /></span>
              </label>
            </div>
          </div>

          <!-- Param fields -->
          <div class="drawer-section">
            <div class="drawer-section-title">PARAMETERS</div>
            <template v-for="f in activeUIStrategy.fields" :key="f.field">
              <!-- Number field -->
              <div v-if="f.type === 'number'" class="drawer-field">
                <label class="drawer-label">{{ $t(f.label) }}</label>
                <div class="input-row">
                  <input type="number" class="setting-input" v-model.number="form[f.field]" :step="f.step" />
                  <button class="btn-save" @click="save(f.configKey, form[f.field])">✓</button>
                </div>
              </div>
              <!-- Toggle field -->
              <div v-else-if="f.type === 'toggle'" class="drawer-row-toggle">
                <span class="drawer-label">{{ $t(f.label) }}</span>
                <label class="toggle toggle-sm">
                  <input type="checkbox" v-model="form[f.field]" @change="save(f.configKey, form[f.field])" />
                  <span class="toggle-track"><span class="toggle-thumb" /></span>
                </label>
              </div>
            </template>
          </div>

          <!-- Wallet selection -->
          <div class="drawer-section">
            <div class="drawer-section-title">WALLETS</div>
            <div v-if="wallets.length === 0" class="drawer-empty">No wallets configured</div>
            <div v-else class="wallet-list">
              <label
                v-for="w in wallets"
                :key="w.id"
                class="wallet-item"
                :class="{ 'wallet-item--selected': selectedWalletIds.includes(w.id) }"
              >
                <input
                  type="checkbox"
                  :value="w.id"
                  v-model="selectedWalletIds"
                  class="wallet-cb"
                />
                <span class="wallet-label">{{ w.label || (w.address ? w.address.slice(0,8) + '…' : w.id) }}</span>
                <span class="wallet-addr">{{ w.address ? w.address.slice(0,6) + '…' + w.address.slice(-4) : '' }}</span>
              </label>
            </div>
          </div>
        </div>

        <!-- Drawer footer actions -->
        <div class="drawer-footer">
          <button
            class="btn-launch"
            :disabled="launching"
            @click="launchStrategy"
          >
            <span :class="{ spin: launching }">{{ launching ? '⟳' : '▶ LAUNCH' }}</span>
          </button>
          <button class="btn-stop" @click="stopStrategy">■ STOP</button>
        </div>
      </div>
    </Transition>

    <!-- Saved toast -->
    <Transition name="fade">
      <div v-if="savedMsg" class="saved-toast">{{ $t('settings.saved') }} ✓</div>
    </Transition>
  </div>
</template>

<script setup>
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { storeToRefs } from 'pinia'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { useApi } from '@/composables/useApi'

const app = useAppStore()
const { strategies: storeStrategies, settingsStale } = storeToRefs(app)
const api = useApi()
const { t } = useI18n()
const savedMsg = ref(false)

// Drawer state
const activeKey = ref(null)
const wallets = ref([])
const selectedWalletIds = ref([])
const launching = ref(false)

// UI static metadata
const uiMetadata = [
  {
    key: 'arbitrage',
    icon: '⟿',
    nameKey: 'settings.sectionArbitrage',
    enabledField: 'arbitrageEnabled',
    configKey: 'trading.strategies.arbitrage.enabled',
    fields: [
      { label: 'settings.minProfitUsd',    field: 'arbitrageMinProfit', configKey: 'trading.strategies.arbitrage.min_profit_usd',   step: 0.1,    type: 'number' },
      { label: 'settings.maxPositionUsd',  field: 'arbitrageMaxPos',    configKey: 'trading.strategies.arbitrage.max_position_usd', step: 10,     type: 'number' },
      { label: 'settings.monitorInterval', field: 'arbitragePoll',      configKey: 'trading.strategies.arbitrage.poll_interval_ms', step: 1000,   type: 'number' },
      { label: 'settings.executeOrders',   field: 'arbitrageExec',      configKey: 'trading.strategies.arbitrage.execute_orders',   type: 'toggle' },
    ]
  },
  {
    key: 'market_making',
    icon: '⟷',
    nameKey: 'settings.sectionMarketMaking',
    enabledField: 'mmEnabled',
    configKey: 'trading.strategies.market_making.enabled',
    fields: [
      { label: 'settings.spreadPct',          field: 'mmSpread',    configKey: 'trading.strategies.market_making.spread_pct',            step: 0.1,  type: 'number' },
      { label: 'settings.maxPositionUsd',      field: 'mmMaxPos',    configKey: 'trading.strategies.market_making.max_position_usd',      step: 10,   type: 'number' },
      { label: 'settings.rebalanceIntervalSec',field: 'mmRebalance', configKey: 'trading.strategies.market_making.rebalance_interval_sec',step: 5,    type: 'number' },
      { label: 'settings.minLiquidityUsd',     field: 'mmMinLiq',    configKey: 'trading.strategies.market_making.min_liquidity_usd',     step: 1000, type: 'number' },
      { label: 'settings.executeOrders',       field: 'mmExec',      configKey: 'trading.strategies.market_making.execute_orders',        type: 'toggle' },
    ]
  },
  {
    key: 'positive_ev',
    icon: '📈',
    nameKey: 'settings.sectionPositiveEv',
    enabledField: 'pevEnabled',
    configKey: 'trading.strategies.positive_ev.enabled',
    fields: [
      { label: 'settings.minEdgePct',     field: 'pevMinEdge', configKey: 'trading.strategies.positive_ev.min_edge_pct',        step: 0.1,  type: 'number' },
      { label: 'settings.minLiquidityUsd',field: 'pevMinLiq',  configKey: 'trading.strategies.positive_ev.min_liquidity_usd',   step: 1000, type: 'number' },
      { label: 'settings.maxPositionUsd', field: 'pevMaxPos',  configKey: 'trading.strategies.positive_ev.max_position_usd',    step: 10,   type: 'number' },
      { label: 'settings.monitorInterval',field: 'pevPoll',    configKey: 'trading.strategies.positive_ev.poll_interval_ms',    step: 5000, type: 'number' },
      { label: 'settings.executeOrders',  field: 'pevExec',    configKey: 'trading.strategies.positive_ev.execute_orders',      type: 'toggle' },
    ]
  },
  {
    key: 'riskless_rate',
    icon: '🛡',
    nameKey: 'settings.sectionRisklessRate',
    enabledField: 'risklessEnabled',
    configKey: 'trading.strategies.riskless_rate.enabled',
    fields: [
      { label: 'settings.minDurationDays',field: 'risklessMinDur', configKey: 'trading.strategies.riskless_rate.min_duration_days', step: 1,    type: 'number' },
      { label: 'settings.maxNoPrice',     field: 'risklessMaxNo',  configKey: 'trading.strategies.riskless_rate.max_no_price',      step: 0.01, type: 'number' },
      { label: 'settings.maxPositionUsd', field: 'risklessMaxPos', configKey: 'trading.strategies.riskless_rate.max_position_usd',  step: 10,   type: 'number' },
      { label: 'settings.monitorInterval',field: 'risklessPoll',   configKey: 'trading.strategies.riskless_rate.poll_interval_ms',  step: 5000, type: 'number' },
      { label: 'settings.executeOrders',  field: 'risklessExec',   configKey: 'trading.strategies.riskless_rate.execute_orders',    type: 'toggle' },
    ]
  },
  {
    key: 'fade_chaos',
    icon: '🌪',
    nameKey: 'settings.sectionFadeChaos',
    enabledField: 'fadeEnabled',
    configKey: 'trading.strategies.fade_chaos.enabled',
    fields: [
      { label: 'settings.spikeThresholdPct',field: 'fadeSpike',    configKey: 'trading.strategies.fade_chaos.spike_threshold_pct', step: 1,    type: 'number' },
      { label: 'settings.cooldownSec',       field: 'fadeCooldown', configKey: 'trading.strategies.fade_chaos.cooldown_sec',        step: 10,   type: 'number' },
      { label: 'settings.maxPositionUsd',    field: 'fadeMaxPos',   configKey: 'trading.strategies.fade_chaos.max_position_usd',    step: 10,   type: 'number' },
      { label: 'settings.monitorInterval',   field: 'fadePoll',     configKey: 'trading.strategies.fade_chaos.poll_interval_ms',    step: 5000, type: 'number' },
      { label: 'settings.executeOrders',     field: 'fadeExec',     configKey: 'trading.strategies.fade_chaos.execute_orders',      type: 'toggle' },
    ]
  },
  {
    key: 'cross_market',
    icon: '⎔',
    nameKey: 'settings.sectionCrossMarket',
    enabledField: 'crossEnabled',
    configKey: 'trading.strategies.cross_market.enabled',
    fields: [
      { label: 'settings.minDivergencePct',field: 'crossMinDiv', configKey: 'trading.strategies.cross_market.min_divergence_pct', step: 0.1, type: 'number' },
      { label: 'settings.maxPositionUsd',  field: 'crossMaxPos', configKey: 'trading.strategies.cross_market.max_position_usd',  step: 10,  type: 'number' },
      { label: 'settings.monitorInterval', field: 'crossPoll',   configKey: 'trading.strategies.cross_market.poll_interval_ms',  step: 5000,type: 'number' },
      { label: 'settings.executeOrders',   field: 'crossExec',   configKey: 'trading.strategies.cross_market.execute_orders',    type: 'toggle' },
    ]
  },
]

const activeUIStrategy = computed(() => uiMetadata.find(s => s.key === activeKey.value))

const displayStrategies = computed(() => {
  return uiMetadata.map(meta => {
    // Find runtime state from store (if any)
    const s = (storeStrategies.value || []).find(rs => rs.name === meta.key)
    return {
      name: meta.key,
      status: s ? s.status : 'off',
      label: t(meta.nameKey),
      icon: meta.icon,
      details: s ? s.details : t('settings.notInitialized'),
      walletLabel: s ? (s.wallet_label || '—') : '—',
      wallet_address: s ? (s.wallet_address || '—') : '—'
    }
  })
})

function openDrawer(key) {
  activeKey.value = key
  
  // 1. Initialise from settings (if available)
  const st = app.settings?.trading?.strategies?.[key]
  if (st && Array.isArray(st.wallet_ids)) {
    selectedWalletIds.value = [...st.wallet_ids]
  } else {
    // 2. Fallback to runtime state
    const storeVer = (storeStrategies.value || []).find(s => s.name === key)
    if (storeVer && storeVer.wallet_id) {
      selectedWalletIds.value = [storeVer.wallet_id]
    } else {
      selectedWalletIds.value = []
    }
  }
}
function closeDrawer() { activeKey.value = null }

async function loadWallets() {
  try { wallets.value = await api.getWallets() } catch {}
}

async function launchStrategy() {
  if (!activeKey.value) return
  launching.value = true
  try {
    await api.startStrategy(activeKey.value, selectedWalletIds.value)
    savedMsg.value = true; setTimeout(() => { savedMsg.value = false }, 2000)
  } catch {}
  launching.value = false
}

async function stopStrategy() {
  if (!activeKey.value) return
  try {
    await api.stopStrategy(activeKey.value)
    savedMsg.value = true; setTimeout(() => { savedMsg.value = false }, 2000)
  } catch {}
}

watch(settingsStale, async (stale) => {
  if (!stale) return
  try { const s = await api.getSettings(); applySettings(s) } catch {}
  app.settingsStale = false
})

const form = reactive({
  arbitrageEnabled: false, arbitrageMinProfit: 0.5, arbitrageMaxPos: 100, arbitragePoll: 5000, arbitrageExec: false,
  mmEnabled: false, mmSpread: 2.0, mmMaxPos: 200, mmRebalance: 30, mmMinLiq: 10000, mmExec: false,
  pevEnabled: false, pevMinEdge: 5.0, pevMinLiq: 5000, pevMaxPos: 50, pevPoll: 30000, pevExec: false,
  risklessEnabled: false, risklessMinDur: 30, risklessMaxNo: 0.05, risklessMaxPos: 50, risklessPoll: 60000, risklessExec: false,
  fadeEnabled: false, fadeSpike: 10.0, fadeCooldown: 300, fadeMaxPos: 50, fadePoll: 10000, fadeExec: false,
  crossEnabled: false, crossMinDiv: 5.0, crossMaxPos: 75, crossPoll: 30000, crossExec: false,
})

onMounted(async () => {
  try {
    let s = await api.getSettings()
    if (Array.isArray(s)) s = s[0] // handle [ { ... } ]
    app.settings = s
    applySettings(s)
  } catch {}
  
  // Only load via REST if store is empty to avoid overwriting WS initial_state
  if (!app.strategies || app.strategies.length === 0) {
    try {
      const s = await api.getStrategies()
      if (s && s.length > 0) app.strategies = s
    } catch {}
  }
  await loadWallets()
})

function applySettings(s) {
  if (!s || !s.trading) return
  const st = s.trading.strategies || {}
  form.arbitrageEnabled = !!st.arbitrage?.enabled
  form.arbitrageMinProfit = st.arbitrage?.min_profit_usd || 0.5
  form.arbitrageMaxPos = st.arbitrage?.max_position_usd || 100
  form.arbitragePoll = st.arbitrage?.poll_interval_ms || 5000
  form.arbitrageExec = !!st.arbitrage?.execute_orders

  form.mmEnabled = !!st.market_making?.enabled
  form.mmSpread = st.market_making?.spread_pct || 2.0
  form.mmMaxPos = st.market_making?.max_position_usd || 200
  form.mmRebalance = st.market_making?.rebalance_interval_sec || 30
  form.mmMinLiq = st.market_making?.min_liquidity_usd || 10000
  form.mmExec = !!st.market_making?.execute_orders

  form.pevEnabled = !!st.positive_ev?.enabled
  form.pevMinEdge = st.positive_ev?.min_edge_pct || 5.0
  form.pevMinLiq = st.positive_ev?.min_liquidity_usd || 5000
  form.pevMaxPos = st.positive_ev?.max_position_usd || 50
  form.pevPoll = st.positive_ev?.poll_interval_ms || 30000
  form.pevExec = !!st.positive_ev?.execute_orders

  form.risklessEnabled = !!st.riskless_rate?.enabled
  form.risklessMinDur = st.riskless_rate?.min_duration_days || 30
  form.risklessMaxNo = st.riskless_rate?.max_no_price || 0.05
  form.risklessMaxPos = st.riskless_rate?.max_position_usd || 50
  form.risklessPoll = st.riskless_rate?.poll_interval_ms || 60000
  form.risklessExec = !!st.riskless_rate?.execute_orders

  form.fadeEnabled = !!st.fade_chaos?.enabled
  form.fadeSpike = st.fade_chaos?.spike_threshold_pct || 10.0
  form.fadeCooldown = st.fade_chaos?.cooldown_sec || 300
  form.fadeMaxPos = st.fade_chaos?.max_position_usd || 50
  form.fadePoll = st.fade_chaos?.poll_interval_ms || 10000
  form.fadeExec = !!st.fade_chaos?.execute_orders

  form.crossEnabled = !!st.cross_market?.enabled
  form.crossMinDiv = st.cross_market?.min_divergence_pct || 5.0
  form.crossMaxPos = st.cross_market?.max_position_usd || 75
  form.crossPoll = st.cross_market?.poll_interval_ms || 30000
  form.crossExec = !!st.cross_market?.execute_orders
}

async function save(key, value) {
  if (String(value) === '***') return
  try {
    await api.postSettings(key, String(value))
    savedMsg.value = true
    setTimeout(() => { savedMsg.value = false }, 2000)
  } catch {}
}
</script>

<style scoped>
.view { display: flex; flex-direction: column; gap: 1.5rem; position: relative; padding-bottom: 2rem; }

.page-header { }
.view-title { font-size: 1rem; font-weight: 700; letter-spacing: 0.04em; color: var(--text-bright); }
.view-sub { font-size: 0.82rem; color: var(--text-secondary); margin-top: 0.2rem; }

/* Grid */
.strategies-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 1rem;
}

/* Strategy Card */
.strategy-card {
  background: var(--bg-card);
  backdrop-filter: blur(12px);
  -webkit-backdrop-filter: blur(12px);
  border: 1px solid var(--border);
  border-top: 2px solid rgba(124,58,237,0.30);
  border-radius: var(--radius-lg);
  display: flex; flex-direction: column;
  transition: all var(--transition-slow);
  overflow: hidden;
  position: relative;
}
.strategy-card::before {
  content: '';
  position: absolute; top: 0; left: 0; right: 0; height: 80px;
  background: radial-gradient(circle at 50% 0%, rgba(124,58,237,0.08) 0%, transparent 70%);
  pointer-events: none;
}
.strategy-card:hover {
  border-top-color: var(--accent);
  transform: translateY(-2px);
  box-shadow: 0 8px 32px rgba(0,0,0,0.5), 0 0 20px rgba(124,58,237,0.12);
}

.sc-header {
  display: flex; align-items: center; justify-content: space-between;
  padding: 1rem 1.25rem 0.75rem;
  position: relative; z-index: 1;
}
.sc-title { display: flex; align-items: center; gap: 0.6rem; }
.sc-icon { font-size: 1.3rem; color: var(--accent-bright); }
.sc-title h3 { margin: 0; font-size: 0.90rem; font-weight: 700; letter-spacing: 0.08em; text-transform: uppercase; color: var(--text-bright); }

.sc-badge {
  font-size: 0.76rem; font-weight: 700; letter-spacing: 0.08em;
  padding: 0.15rem 0.5rem; border-radius: 2px;
}
.sc-badge--on  { background: rgba(52,211,153,0.12); color: var(--success); border: 1px solid rgba(52,211,153,0.25); }
.sc-badge--off { background: rgba(255,255,255,0.04); color: var(--text-muted); border: 1px solid var(--border); }

.sc-body { padding: 0 1.25rem 0.75rem; position: relative; z-index: 1; }
.sc-fields-count { font-size: 0.80rem; color: var(--text-secondary); }
.sc-wallet-info { margin-top: 0.4rem; display: flex; flex-direction: column; gap: 0.1rem; }
.sc-wallet { font-size: 0.78rem; font-family: var(--font-mono); color: var(--accent-bright); margin: 0; }
.sc-addr { font-size: 0.70rem; font-family: var(--font-mono); color: var(--text-muted); word-break: break-all; margin: 0; }

.sc-footer {
  padding: 0.75rem 1.25rem;
  border-top: 1px solid var(--border);
  position: relative; z-index: 1;
}
.btn-configure {
  width: 100%;
  padding: 0.45rem 0;
  background: rgba(124,58,237,0.08);
  border: 1px solid rgba(124,58,237,0.30);
  border-radius: var(--radius);
  color: var(--accent-bright);
  font-size: 0.84rem; font-weight: 700; letter-spacing: 0.10em;
  font-family: var(--font-mono);
  cursor: pointer;
  transition: all var(--transition);
}
.btn-configure:hover {
  background: var(--accent);
  border-color: var(--accent);
  color: #fff;
  box-shadow: var(--accent-glow);
}

/* Drawer Backdrop */
.drawer-backdrop {
  position: fixed; inset: 0;
  background: rgba(0,0,0,0.55);
  backdrop-filter: blur(2px);
  z-index: 49;
}

/* Drawer */
.drawer {
  position: fixed;
  top: var(--topbar-h);
  right: 0;
  bottom: 0;
  width: 400px;
  background: #0d0b1a;
  border-left: 1px solid rgba(124,58,237,0.25);
  display: flex; flex-direction: column;
  z-index: 50;
  overflow: hidden;
}

.drawer-header {
  display: flex; align-items: center; justify-content: space-between;
  padding: 1rem 1.25rem;
  border-bottom: 1px solid var(--border);
  background: rgba(124,58,237,0.06);
  flex-shrink: 0;
}
.drawer-title {
  display: flex; align-items: center; gap: 0.6rem;
  font-size: 0.90rem; font-weight: 700; letter-spacing: 0.10em;
  text-transform: uppercase; color: var(--text-bright);
}
.drawer-icon { font-size: 1.2rem; color: var(--accent-bright); }
.drawer-close {
  background: none; border: 1px solid var(--border);
  color: var(--text-secondary); border-radius: var(--radius);
  width: 28px; height: 28px; font-size: 0.80rem;
  cursor: pointer; transition: all var(--transition);
  display: flex; align-items: center; justify-content: center;
}
.drawer-close:hover { border-color: var(--danger); color: var(--danger); }

.drawer-body {
  flex: 1; overflow-y: auto;
  padding: 1rem 1.25rem;
  display: flex; flex-direction: column; gap: 1.25rem;
}

.drawer-section { display: flex; flex-direction: column; gap: 0.75rem; }
.drawer-section-title {
  font-size: 0.76rem; font-weight: 700; letter-spacing: 0.12em;
  color: var(--accent); text-transform: uppercase;
  padding-bottom: 0.4rem;
  border-bottom: 1px solid var(--border);
}

.drawer-field { display: flex; flex-direction: column; gap: 0.35rem; }
.drawer-label { font-size: 0.82rem; font-weight: 600; text-transform: uppercase; letter-spacing: 0.08em; color: var(--text-secondary); }

.drawer-row-toggle {
  display: flex; align-items: center; justify-content: space-between;
  background: rgba(0,0,0,0.2);
  padding: 0.6rem 0.85rem; border-radius: var(--radius);
  border: 1px dashed var(--border);
}

.input-row { display: flex; gap: 0.4rem; }
.setting-input {
  flex: 1; background: rgba(0,0,0,0.35);
  border: 1px solid var(--border); border-radius: var(--radius);
  color: var(--text-primary);
  padding: 0.4rem 0.65rem; font-size: 0.90rem; font-family: var(--font-mono);
  outline: none; transition: border-color var(--transition);
}
.setting-input:focus { border-color: var(--accent); box-shadow: 0 0 0 1px rgba(124,58,237,0.15); }

.btn-save {
  background: rgba(124,58,237,0.10); border: 1px solid rgba(124,58,237,0.30);
  color: var(--accent-bright); border-radius: var(--radius);
  width: 36px; font-size: 1rem; font-weight: 700;
  cursor: pointer; transition: all var(--transition);
  display: flex; align-items: center; justify-content: center;
}
.btn-save:hover { background: var(--accent); color: #fff; border-color: var(--accent); }

/* Wallet list */
.drawer-empty { font-size: 0.84rem; color: var(--text-muted); text-align: center; padding: 0.75rem; }
.wallet-list { display: flex; flex-direction: column; gap: 0.35rem; }
.wallet-item {
  display: flex; align-items: center; gap: 0.6rem;
  padding: 0.5rem 0.75rem; border-radius: var(--radius);
  border: 1px solid var(--border);
  background: rgba(255,255,255,0.02);
  cursor: pointer; transition: all var(--transition);
}
.wallet-item:hover { border-color: rgba(124,58,237,0.30); background: rgba(124,58,237,0.05); }
.wallet-item--selected { border-color: var(--accent); background: rgba(124,58,237,0.10); }
.wallet-cb { accent-color: var(--accent); cursor: pointer; }
.wallet-label { flex: 1; font-size: 0.86rem; font-weight: 600; color: var(--text-primary); }
.wallet-addr { font-size: 0.80rem; color: var(--text-muted); font-family: var(--font-mono); }

/* Toggle */
.toggle { display: flex; align-items: center; cursor: pointer; }
.toggle input { display: none; }
.toggle-track {
  width: 44px; height: 24px; background: rgba(255,255,255,0.08);
  border-radius: 12px; position: relative;
  transition: all var(--transition);
  border: 1px solid var(--border);
}
.toggle input:checked ~ .toggle-track { background: var(--accent); border-color: var(--accent-bright); }
.toggle-thumb {
  position: absolute; width: 18px; height: 18px;
  background: var(--text-muted); border-radius: 50%;
  top: 2px; left: 2px; transition: all var(--transition);
}
.toggle input:checked ~ .toggle-track .toggle-thumb { left: 22px; background: #fff; }
.toggle-sm .toggle-track { width: 36px; height: 20px; }
.toggle-sm .toggle-thumb { width: 14px; height: 14px; top: 2px; left: 2px; }
.toggle-sm input:checked ~ .toggle-track .toggle-thumb { left: 18px; }

/* Drawer footer */
.drawer-footer {
  display: flex; gap: 0.75rem;
  padding: 1rem 1.25rem;
  border-top: 1px solid var(--border);
  flex-shrink: 0;
  background: rgba(0,0,0,0.2);
}
.btn-launch {
  flex: 1; padding: 0.55rem;
  background: var(--accent); border: none;
  border-radius: var(--radius); color: #fff;
  font-size: 0.88rem; font-weight: 700; letter-spacing: 0.10em;
  font-family: var(--font-mono); cursor: pointer;
  transition: all var(--transition);
}
.btn-launch:hover:not(:disabled) { background: var(--accent-hover); box-shadow: var(--accent-glow); }
.btn-launch:disabled { opacity: 0.4; cursor: not-allowed; }

.btn-stop {
  padding: 0.55rem 1rem;
  background: var(--danger-dim); border: 1px solid var(--danger);
  border-radius: var(--radius); color: var(--danger);
  font-size: 0.88rem; font-weight: 700; letter-spacing: 0.10em;
  font-family: var(--font-mono); cursor: pointer;
  transition: all var(--transition);
}
.btn-stop:hover { background: var(--danger); color: #fff; }

/* Saved toast */
.saved-toast {
  position: fixed; bottom: 2rem; right: 2rem;
  background: var(--success); color: #000;
  padding: 0.6rem 1.2rem; border-radius: var(--radius-lg);
  font-size: 0.90rem; font-weight: 800;
  box-shadow: 0 4px 20px rgba(52,211,153,0.35);
  font-family: var(--font-mono); letter-spacing: 0.08em;
  z-index: 1000;
}
.fade-enter-active { animation: fadeSlideUp 0.2s ease both; }
.fade-leave-active { animation: fadeSlideUp 0.2s ease reverse both; }

/* Transitions */
.fade-backdrop-enter-active, .fade-backdrop-leave-active { transition: opacity 0.25s ease; }
.fade-backdrop-enter-from, .fade-backdrop-leave-to { opacity: 0; }

.slide-drawer-enter-active, .slide-drawer-leave-active { transition: transform 0.3s ease; }
.slide-drawer-enter-from, .slide-drawer-leave-to { transform: translateX(100%); }
</style>
