<template>
  <div class="view">
    <div class="page-header anim-in">
      <div class="header-top">
        <h2 class="view-title">{{ $t('nav.strategies') }}</h2>
      </div>
    </div>

    <!-- Strategies Grid -->
    <div class="strategies-grid">
      <!-- Arbitrage -->
      <div class="strategy-card anim-in">
        <div class="sc-header">
          <div class="sc-title">
            <span class="sc-icon">⟿</span>
            <h3>{{ $t('settings.sectionArbitrage') }}</h3>
          </div>
          <label class="toggle">
            <input type="checkbox" v-model="form.arbitrageEnabled" @change="save('trading.strategies.arbitrage.enabled', form.arbitrageEnabled)" />
            <span class="toggle-track"><span class="toggle-thumb" /></span>
          </label>
        </div>
        <div class="sc-body">
          <div class="sc-field">
            <label>{{ $t('settings.minProfitUsd') }}</label>
            <div class="input-row">
              <input type="number" class="setting-input" v-model.number="form.arbitrageMinProfit" step="0.1" />
              <button class="btn-save" @click="save('trading.strategies.arbitrage.min_profit_usd', form.arbitrageMinProfit)">✓</button>
            </div>
          </div>
          <div class="sc-field">
            <label>{{ $t('settings.maxPositionUsd') }}</label>
            <div class="input-row">
              <input type="number" class="setting-input" v-model.number="form.arbitrageMaxPos" step="10" />
              <button class="btn-save" @click="save('trading.strategies.arbitrage.max_position_usd', form.arbitrageMaxPos)">✓</button>
            </div>
          </div>
          <div class="sc-field">
            <label>{{ $t('settings.monitorInterval') }}</label>
            <div class="input-row">
              <input type="number" class="setting-input" v-model.number="form.arbitragePoll" step="1000" />
              <button class="btn-save" @click="save('trading.strategies.arbitrage.poll_interval_ms', form.arbitragePoll)">✓</button>
            </div>
          </div>
          <div class="sc-field flex-row">
            <label>{{ $t('settings.executeOrders') }}</label>
            <label class="toggle toggle-sm">
              <input type="checkbox" v-model="form.arbitrageExec" @change="save('trading.strategies.arbitrage.execute_orders', form.arbitrageExec)" />
              <span class="toggle-track"><span class="toggle-thumb" /></span>
            </label>
          </div>
        </div>
      </div>

      <!-- Market Making -->
      <div class="strategy-card anim-in" style="animation-delay: 60ms;">
        <div class="sc-header">
          <div class="sc-title">
            <span class="sc-icon">⟷</span>
            <h3>{{ $t('settings.sectionMarketMaking') }}</h3>
          </div>
          <label class="toggle">
            <input type="checkbox" v-model="form.mmEnabled" @change="save('trading.strategies.market_making.enabled', form.mmEnabled)" />
            <span class="toggle-track"><span class="toggle-thumb" /></span>
          </label>
        </div>
        <div class="sc-body">
          <div class="sc-field">
            <label>{{ $t('settings.spreadPct') }}</label>
            <div class="input-row">
              <input type="number" class="setting-input" v-model.number="form.mmSpread" step="0.1" />
              <button class="btn-save" @click="save('trading.strategies.market_making.spread_pct', form.mmSpread)">✓</button>
            </div>
          </div>
          <div class="sc-field">
            <label>{{ $t('settings.maxPositionUsd') }}</label>
            <div class="input-row">
              <input type="number" class="setting-input" v-model.number="form.mmMaxPos" step="10" />
              <button class="btn-save" @click="save('trading.strategies.market_making.max_position_usd', form.mmMaxPos)">✓</button>
            </div>
          </div>
          <div class="sc-field">
            <label>{{ $t('settings.rebalanceIntervalSec') }}</label>
            <div class="input-row">
              <input type="number" class="setting-input" v-model.number="form.mmRebalance" step="5" />
              <button class="btn-save" @click="save('trading.strategies.market_making.rebalance_interval_sec', form.mmRebalance)">✓</button>
            </div>
          </div>
          <div class="sc-field">
            <label>{{ $t('settings.minLiquidityUsd') }}</label>
            <div class="input-row">
              <input type="number" class="setting-input" v-model.number="form.mmMinLiq" step="1000" />
              <button class="btn-save" @click="save('trading.strategies.market_making.min_liquidity_usd', form.mmMinLiq)">✓</button>
            </div>
          </div>
          <div class="sc-field flex-row">
            <label>{{ $t('settings.executeOrders') }}</label>
            <label class="toggle toggle-sm">
              <input type="checkbox" v-model="form.mmExec" @change="save('trading.strategies.market_making.execute_orders', form.mmExec)" />
              <span class="toggle-track"><span class="toggle-thumb" /></span>
            </label>
          </div>
        </div>
      </div>

      <!-- Positive EV -->
      <div class="strategy-card anim-in" style="animation-delay: 120ms;">
        <div class="sc-header">
          <div class="sc-title">
            <span class="sc-icon">📈</span>
            <h3>{{ $t('settings.sectionPositiveEv') }}</h3>
          </div>
          <label class="toggle">
            <input type="checkbox" v-model="form.pevEnabled" @change="save('trading.strategies.positive_ev.enabled', form.pevEnabled)" />
            <span class="toggle-track"><span class="toggle-thumb" /></span>
          </label>
        </div>
        <div class="sc-body">
          <div class="sc-field">
            <label>{{ $t('settings.minEdgePct') }}</label>
            <div class="input-row">
              <input type="number" class="setting-input" v-model.number="form.pevMinEdge" step="0.1" />
              <button class="btn-save" @click="save('trading.strategies.positive_ev.min_edge_pct', form.pevMinEdge)">✓</button>
            </div>
          </div>
          <div class="sc-field">
            <label>{{ $t('settings.minLiquidityUsd') }}</label>
            <div class="input-row">
              <input type="number" class="setting-input" v-model.number="form.pevMinLiq" step="1000" />
              <button class="btn-save" @click="save('trading.strategies.positive_ev.min_liquidity_usd', form.pevMinLiq)">✓</button>
            </div>
          </div>
          <div class="sc-field">
            <label>{{ $t('settings.maxPositionUsd') }}</label>
            <div class="input-row">
              <input type="number" class="setting-input" v-model.number="form.pevMaxPos" step="10" />
              <button class="btn-save" @click="save('trading.strategies.positive_ev.max_position_usd', form.pevMaxPos)">✓</button>
            </div>
          </div>
          <div class="sc-field">
            <label>{{ $t('settings.monitorInterval') }}</label>
            <div class="input-row">
              <input type="number" class="setting-input" v-model.number="form.pevPoll" step="5000" />
              <button class="btn-save" @click="save('trading.strategies.positive_ev.poll_interval_ms', form.pevPoll)">✓</button>
            </div>
          </div>
          <div class="sc-field flex-row">
            <label>{{ $t('settings.executeOrders') }}</label>
            <label class="toggle toggle-sm">
              <input type="checkbox" v-model="form.pevExec" @change="save('trading.strategies.positive_ev.execute_orders', form.pevExec)" />
              <span class="toggle-track"><span class="toggle-thumb" /></span>
            </label>
          </div>
        </div>
      </div>

      <!-- Riskless Rate -->
      <div class="strategy-card anim-in" style="animation-delay: 180ms;">
        <div class="sc-header">
          <div class="sc-title">
            <span class="sc-icon">🛡</span>
            <h3>{{ $t('settings.sectionRisklessRate') }}</h3>
          </div>
          <label class="toggle">
            <input type="checkbox" v-model="form.risklessEnabled" @change="save('trading.strategies.riskless_rate.enabled', form.risklessEnabled)" />
            <span class="toggle-track"><span class="toggle-thumb" /></span>
          </label>
        </div>
        <div class="sc-body">
          <div class="sc-field">
            <label>{{ $t('settings.minDurationDays') }}</label>
            <div class="input-row">
              <input type="number" class="setting-input" v-model.number="form.risklessMinDur" step="1" />
              <button class="btn-save" @click="save('trading.strategies.riskless_rate.min_duration_days', form.risklessMinDur)">✓</button>
            </div>
          </div>
          <div class="sc-field">
            <label>{{ $t('settings.maxNoPrice') }}</label>
            <div class="input-row">
              <input type="number" class="setting-input" v-model.number="form.risklessMaxNo" step="0.01" />
              <button class="btn-save" @click="save('trading.strategies.riskless_rate.max_no_price', form.risklessMaxNo)">✓</button>
            </div>
          </div>
          <div class="sc-field">
            <label>{{ $t('settings.maxPositionUsd') }}</label>
            <div class="input-row">
              <input type="number" class="setting-input" v-model.number="form.risklessMaxPos" step="10" />
              <button class="btn-save" @click="save('trading.strategies.riskless_rate.max_position_usd', form.risklessMaxPos)">✓</button>
            </div>
          </div>
          <div class="sc-field">
            <label>{{ $t('settings.monitorInterval') }}</label>
            <div class="input-row">
              <input type="number" class="setting-input" v-model.number="form.risklessPoll" step="5000" />
              <button class="btn-save" @click="save('trading.strategies.riskless_rate.poll_interval_ms', form.risklessPoll)">✓</button>
            </div>
          </div>
          <div class="sc-field flex-row">
            <label>{{ $t('settings.executeOrders') }}</label>
            <label class="toggle toggle-sm">
              <input type="checkbox" v-model="form.risklessExec" @change="save('trading.strategies.riskless_rate.execute_orders', form.risklessExec)" />
              <span class="toggle-track"><span class="toggle-thumb" /></span>
            </label>
          </div>
        </div>
      </div>

      <!-- Fade Chaos -->
      <div class="strategy-card anim-in" style="animation-delay: 240ms;">
        <div class="sc-header">
          <div class="sc-title">
            <span class="sc-icon">🌪</span>
            <h3>{{ $t('settings.sectionFadeChaos') }}</h3>
          </div>
          <label class="toggle">
            <input type="checkbox" v-model="form.fadeEnabled" @change="save('trading.strategies.fade_chaos.enabled', form.fadeEnabled)" />
            <span class="toggle-track"><span class="toggle-thumb" /></span>
          </label>
        </div>
        <div class="sc-body">
          <div class="sc-field">
            <label>{{ $t('settings.spikeThresholdPct') }}</label>
            <div class="input-row">
              <input type="number" class="setting-input" v-model.number="form.fadeSpike" step="1" />
              <button class="btn-save" @click="save('trading.strategies.fade_chaos.spike_threshold_pct', form.fadeSpike)">✓</button>
            </div>
          </div>
          <div class="sc-field">
            <label>{{ $t('settings.cooldownSec') }}</label>
            <div class="input-row">
              <input type="number" class="setting-input" v-model.number="form.fadeCooldown" step="10" />
              <button class="btn-save" @click="save('trading.strategies.fade_chaos.cooldown_sec', form.fadeCooldown)">✓</button>
            </div>
          </div>
          <div class="sc-field">
            <label>{{ $t('settings.maxPositionUsd') }}</label>
            <div class="input-row">
              <input type="number" class="setting-input" v-model.number="form.fadeMaxPos" step="10" />
              <button class="btn-save" @click="save('trading.strategies.fade_chaos.max_position_usd', form.fadeMaxPos)">✓</button>
            </div>
          </div>
          <div class="sc-field">
            <label>{{ $t('settings.monitorInterval') }}</label>
            <div class="input-row">
              <input type="number" class="setting-input" v-model.number="form.fadePoll" step="5000" />
              <button class="btn-save" @click="save('trading.strategies.fade_chaos.poll_interval_ms', form.fadePoll)">✓</button>
            </div>
          </div>
          <div class="sc-field flex-row">
            <label>{{ $t('settings.executeOrders') }}</label>
            <label class="toggle toggle-sm">
              <input type="checkbox" v-model="form.fadeExec" @change="save('trading.strategies.fade_chaos.execute_orders', form.fadeExec)" />
              <span class="toggle-track"><span class="toggle-thumb" /></span>
            </label>
          </div>
        </div>
      </div>

      <!-- Cross Market -->
      <div class="strategy-card anim-in" style="animation-delay: 300ms;">
        <div class="sc-header">
          <div class="sc-title">
            <span class="sc-icon">⎔</span>
            <h3>{{ $t('settings.sectionCrossMarket') }}</h3>
          </div>
          <label class="toggle">
            <input type="checkbox" v-model="form.crossEnabled" @change="save('trading.strategies.cross_market.enabled', form.crossEnabled)" />
            <span class="toggle-track"><span class="toggle-thumb" /></span>
          </label>
        </div>
        <div class="sc-body">
          <div class="sc-field">
            <label>{{ $t('settings.minDivergencePct') }}</label>
            <div class="input-row">
              <input type="number" class="setting-input" v-model.number="form.crossMinDiv" step="0.1" />
              <button class="btn-save" @click="save('trading.strategies.cross_market.min_divergence_pct', form.crossMinDiv)">✓</button>
            </div>
          </div>
          <div class="sc-field">
            <label>{{ $t('settings.maxPositionUsd') }}</label>
            <div class="input-row">
              <input type="number" class="setting-input" v-model.number="form.crossMaxPos" step="10" />
              <button class="btn-save" @click="save('trading.strategies.cross_market.max_position_usd', form.crossMaxPos)">✓</button>
            </div>
          </div>
          <div class="sc-field">
            <label>{{ $t('settings.monitorInterval') }}</label>
            <div class="input-row">
              <input type="number" class="setting-input" v-model.number="form.crossPoll" step="5000" />
              <button class="btn-save" @click="save('trading.strategies.cross_market.poll_interval_ms', form.crossPoll)">✓</button>
            </div>
          </div>
          <div class="sc-field flex-row">
            <label>{{ $t('settings.executeOrders') }}</label>
            <label class="toggle toggle-sm">
              <input type="checkbox" v-model="form.crossExec" @change="save('trading.strategies.cross_market.execute_orders', form.crossExec)" />
              <span class="toggle-track"><span class="toggle-thumb" /></span>
            </label>
          </div>
        </div>
      </div>
    </div>

    <!-- Saved toast -->
    <Transition name="fade">
      <div v-if="savedMsg" class="saved-toast">{{ $t('settings.saved') }} ✓</div>
    </Transition>
  </div>
</template>

<script setup>
import { onMounted, reactive, ref, watch } from 'vue'
import { storeToRefs } from 'pinia'
import { useAppStore } from '@/stores/app'
import { useApi } from '@/composables/useApi'

const app = useAppStore()
const api = useApi()
const savedMsg = ref(false)
const { settingsStale } = storeToRefs(app)

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
  try { const s = await api.getSettings(); app.settings = s; applySettings(s) } catch {}
})

function applySettings(s) {
  const st = s.trading?.strategies || {}
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

.page-header { display: flex; flex-direction: column; gap: 1rem; }
.header-top { display: flex; align-items: center; justify-content: space-between; }
.view-title { font-size: 1.25rem; font-weight: 800; letter-spacing: 0.1em; color: var(--text-bright); text-transform: uppercase; }

/* Strategies Grid */
.strategies-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(340px, 1fr));
  gap: 1.25rem;
}

/* Strategy Card */
.strategy-card {
  background: var(--bg-card);
  backdrop-filter: blur(12px);
  -webkit-backdrop-filter: blur(12px);
  border: 1px solid var(--border);
  border-top: 3px solid rgba(157,0,255,0.4);
  border-radius: var(--radius-lg);
  display: flex; flex-direction: column;
  transition: all var(--transition-slow);
  box-shadow: var(--shadow-sm);
  position: relative;
  overflow: hidden;
}

.strategy-card:hover {
  border-top-color: var(--accent-bright);
  transform: translateY(-4px);
  box-shadow: 0 12px 32px rgba(0,0,0,0.5), 0 0 24px rgba(157,0,255,0.15);
}

.strategy-card::before {
  content: ''; position: absolute; top: 0; left: 0; right: 0; height: 100px;
  background: radial-gradient(circle at 50% 0%, rgba(157,0,255,0.1) 0%, transparent 70%);
  pointer-events: none;
}

.sc-header {
  display: flex; align-items: center; justify-content: space-between;
  padding: 1.25rem 1.25rem 1rem;
  border-bottom: 1px solid var(--border-subtle);
  position: relative; z-index: 1;
}

.sc-title {
  display: flex; align-items: center; gap: 0.75rem;
}

.sc-icon {
  font-size: 1.4rem; color: var(--accent-bright);
  text-shadow: var(--accent-glow);
}

.sc-title h3 {
  margin: 0; font-size: 1.05rem; font-weight: 700; letter-spacing: 0.1em;
  text-transform: uppercase; color: var(--text-bright);
}

.sc-body {
  padding: 1.25rem;
  display: flex; flex-direction: column; gap: 1rem;
  position: relative; z-index: 1;
}

.sc-field {
  display: flex; flex-direction: column; gap: 0.4rem;
}

.sc-field.flex-row {
  flex-direction: row; align-items: center; justify-content: space-between;
  background: rgba(0,0,0,0.2); padding: 0.75rem 1rem; border-radius: var(--radius);
  border: 1px dashed var(--border-subtle);
  margin-top: 0.5rem;
}

.sc-field label {
  font-size: 0.85rem; font-weight: 600; text-transform: uppercase;
  letter-spacing: 0.1em; color: var(--text-secondary);
}

.input-row { display: flex; gap: 0.5rem; }

.setting-input {
  flex: 1;
  background: rgba(0,0,0,0.4);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  color: var(--text-primary);
  padding: 0.5rem 0.75rem;
  font-size: 0.95rem; font-family: var(--font-mono);
  outline: none; transition: all var(--transition);
  box-shadow: inset 0 2px 6px rgba(0,0,0,0.3);
}
.setting-input:focus { border-color: var(--accent); box-shadow: 0 0 0 1px var(--accent), inset 0 2px 6px rgba(0,0,0,0.3); }

.btn-save {
  background: rgba(157,0,255,0.1);
  border: 1px solid rgba(157,0,255,0.3);
  color: var(--accent-bright);
  border-radius: var(--radius);
  width: 40px;
  font-size: 1.1rem; font-weight: 700;
  cursor: pointer; transition: all var(--transition);
  display: flex; align-items: center; justify-content: center;
}
.btn-save:hover { background: var(--accent); color: #fff; border-color: var(--accent); box-shadow: var(--accent-glow); }

/* Toggle */
.toggle { display: flex; align-items: center; cursor: pointer; }
.toggle input { display: none; }
.toggle-track {
  width: 44px; height: 24px;
  background: var(--bg-input);
  border-radius: 12px; position: relative;
  transition: all var(--transition);
  border: 1px solid var(--border);
}
.toggle input:checked ~ .toggle-track { background: var(--accent); border-color: var(--accent-bright); box-shadow: inset 0 0 8px rgba(0,0,0,0.3); }
.toggle-thumb {
  position: absolute; width: 18px; height: 18px;
  background: var(--text-muted); border-radius: 50%;
  top: 2px; left: 2px; transition: all var(--transition);
}
.toggle input:checked ~ .toggle-track .toggle-thumb { left: 22px; background: #fff; box-shadow: 0 2px 4px rgba(0,0,0,0.4); }

.toggle-sm .toggle-track { width: 36px; height: 20px; }
.toggle-sm .toggle-thumb { width: 14px; height: 14px; top: 2px; left: 2px; }
.toggle-sm input:checked ~ .toggle-track .toggle-thumb { left: 18px; }

/* Saved toast */
.saved-toast {
  position: fixed;
  bottom: 2rem; right: 2rem;
  background: var(--success);
  color: #000;
  padding: 0.6rem 1.2rem;
  border-radius: var(--radius-lg);
  font-size: 0.95rem; font-weight: 800;
  box-shadow: 0 4px 20px rgba(0, 255, 157, 0.4);
  font-family: var(--font-mono); letter-spacing: 0.1em;
  z-index: 1000;
}
.fade-enter-active { animation: fadeSlideUp 0.2s ease both; }
.fade-leave-active { animation: fadeSlideUp 0.2s ease reverse both; }
</style>
