<template>
  <div class="view">
    <h2 class="view-title">{{ $t('settings.title') }}</h2>

    <!-- UI Section -->
    <section class="settings-section">
      <h3 class="section-title">{{ $t('settings.sectionUi') }}</h3>
      <div class="settings-grid">
        <div class="setting-card">
          <label class="setting-label">{{ $t('settings.language') }}</label>
          <select class="setting-input" v-model="form.language" @change="save('ui.language', form.language)">
            <option v-for="l in LANGS" :key="l" :value="l">{{ l.toUpperCase() }}</option>
          </select>
        </div>
      </div>
    </section>

    <!-- Log Section -->
    <section class="settings-section">
      <h3 class="section-title">{{ $t('settings.sectionLog') }}</h3>
      <div class="settings-grid">
        <div class="setting-card">
          <label class="setting-label">{{ $t('settings.logLevel') }}</label>
          <select class="setting-input" v-model="form.logLevel" @change="save('log.level', form.logLevel)">
            <option v-for="l in logLevels" :key="l" :value="l">{{ l }}</option>
          </select>
        </div>
        <div class="setting-card">
          <label class="setting-label">{{ $t('settings.logFormat') }}</label>
          <select class="setting-input" v-model="form.logFormat" @change="save('log.format', form.logFormat)">
            <option v-for="f in logFormats" :key="f" :value="f">{{ f }}</option>
          </select>
        </div>
      </div>
    </section>

    <!-- Monitor Section -->
    <section class="settings-section">
      <h3 class="section-title">{{ $t('settings.sectionMonitor') }}</h3>
      <div class="settings-grid">
        <div class="setting-card">
          <label class="setting-label">{{ $t('settings.monitorEnabled') }}</label>
          <label class="toggle">
            <input type="checkbox" v-model="form.monitorEnabled" @change="save('monitor.enabled', form.monitorEnabled)" />
            <span class="toggle-track"><span class="toggle-thumb" /></span>
          </label>
        </div>
        <div class="setting-card">
          <label class="setting-label">{{ $t('settings.monitorInterval') }}</label>
          <div class="input-row">
            <input type="number" class="setting-input" v-model.number="form.monitorInterval" min="100" step="100" />
            <button class="btn-save" @click="save('monitor.poll_interval_ms', form.monitorInterval)">{{ $t('settings.save') }}</button>
          </div>
        </div>
      </div>
    </section>

    <!-- Trades Monitor Section -->
    <section class="settings-section">
      <h3 class="section-title">{{ $t('settings.sectionTradesMonitor') }}</h3>
      <div class="settings-grid">
        <div class="setting-card">
          <label class="setting-label">{{ $t('settings.tradesEnabled') }}</label>
          <label class="toggle">
            <input type="checkbox" v-model="form.tradesEnabled" @change="save('monitor.trades.enabled', form.tradesEnabled)" />
            <span class="toggle-track"><span class="toggle-thumb" /></span>
          </label>
        </div>
        <div class="setting-card">
          <label class="setting-label">{{ $t('settings.tradesInterval') }}</label>
          <div class="input-row">
            <input type="number" class="setting-input" v-model.number="form.tradesInterval" min="1000" step="1000" />
            <button class="btn-save" @click="save('monitor.trades.poll_interval_ms', form.tradesInterval)">{{ $t('settings.save') }}</button>
          </div>
        </div>
        <div class="setting-card">
          <label class="setting-label">{{ $t('settings.alertOnFill') }}</label>
          <label class="toggle">
            <input type="checkbox" v-model="form.alertOnFill" @change="save('monitor.trades.alert_on_fill', form.alertOnFill)" />
            <span class="toggle-track"><span class="toggle-thumb" /></span>
          </label>
        </div>
        <div class="setting-card">
          <label class="setting-label">{{ $t('settings.alertOnCancel') }}</label>
          <label class="toggle">
            <input type="checkbox" v-model="form.alertOnCancel" @change="save('monitor.trades.alert_on_cancel', form.alertOnCancel)" />
            <span class="toggle-track"><span class="toggle-thumb" /></span>
          </label>
        </div>
        <div class="setting-card">
          <label class="setting-label">{{ $t('settings.trackPositions') }}</label>
          <label class="toggle">
            <input type="checkbox" v-model="form.trackPositions" @change="save('monitor.trades.track_positions', form.trackPositions)" />
            <span class="toggle-track"><span class="toggle-thumb" /></span>
          </label>
        </div>
      </div>
    </section>

    <!-- Trading Section -->
    <section class="settings-section">
      <h3 class="section-title">{{ $t('settings.sectionTrading') }}</h3>
      <div class="settings-grid">
        <div class="setting-card">
          <label class="setting-label">{{ $t('settings.tradingEnabled') }}</label>
          <label class="toggle">
            <input type="checkbox" v-model="form.tradingEnabled" @change="save('trading.enabled', form.tradingEnabled)" />
            <span class="toggle-track"><span class="toggle-thumb" /></span>
          </label>
        </div>
        <div class="setting-card">
          <label class="setting-label">{{ $t('settings.maxPositionUsd') }}</label>
          <div class="input-row">
            <input type="number" class="setting-input" v-model.number="form.maxPositionUsd" min="0" step="10" />
            <button class="btn-save" @click="save('trading.max_position_usd', form.maxPositionUsd)">{{ $t('settings.save') }}</button>
          </div>
        </div>
        <div class="setting-card">
          <label class="setting-label">{{ $t('settings.slippagePct') }}</label>
          <div class="input-row">
            <input type="number" class="setting-input" v-model.number="form.slippagePct" min="0" step="0.1" />
            <button class="btn-save" @click="save('trading.slippage_pct', form.slippagePct)">{{ $t('settings.save') }}</button>
          </div>
        </div>
        <div class="setting-card">
          <label class="setting-label">{{ $t('settings.defaultOrderType') }}</label>
          <select class="setting-input" v-model="form.defaultOrderType" @change="save('trading.default_order_type', form.defaultOrderType)">
            <option v-for="t in orderTypes" :key="t" :value="t">{{ t }}</option>
          </select>
        </div>
      </div>
    </section>

    <!-- Copytrading Section -->
    <section class="settings-section">
      <h3 class="section-title">{{ $t('settings.sectionCopytrading') }}</h3>
      <div class="settings-grid">
        <div class="setting-card">
          <label class="setting-label">{{ $t('settings.copytradingEnabled') }}</label>
          <label class="toggle">
            <input type="checkbox" v-model="form.copytradingEnabled" @change="save('copytrading.enabled', form.copytradingEnabled)" />
            <span class="toggle-track"><span class="toggle-thumb" /></span>
          </label>
        </div>
        <div class="setting-card">
          <label class="setting-label">{{ $t('settings.copytradingInterval') }}</label>
          <div class="input-row">
            <input type="number" class="setting-input" v-model.number="form.copytradingInterval" min="1000" step="1000" />
            <button class="btn-save" @click="save('copytrading.poll_interval_ms', form.copytradingInterval)">{{ $t('settings.save') }}</button>
          </div>
        </div>
        <div class="setting-card">
          <label class="setting-label">{{ $t('settings.sizeMode') }}</label>
          <select class="setting-input" v-model="form.sizeMode" @change="save('copytrading.size_mode', form.sizeMode)">
            <option v-for="m in sizeModes" :key="m" :value="m">{{ m }}</option>
          </select>
        </div>
      </div>
    </section>

    <!-- Web UI Section -->
    <section class="settings-section">
      <h3 class="section-title">{{ $t('settings.sectionWebUi') }}</h3>
      <div class="settings-grid">
        <div class="setting-card">
          <label class="setting-label">{{ $t('settings.webUiEnabled') }}</label>
          <label class="toggle">
            <input type="checkbox" v-model="form.webUiEnabled" @change="save('webui.enabled', form.webUiEnabled)" />
            <span class="toggle-track"><span class="toggle-thumb" /></span>
          </label>
        </div>
        <div class="setting-card">
          <label class="setting-label">{{ $t('settings.webUiListen') }}</label>
          <div class="input-row">
            <input type="text" class="setting-input" v-model="form.webUiListen" placeholder="127.0.0.1:8080" />
            <button class="btn-save" @click="save('webui.listen', form.webUiListen)">{{ $t('settings.save') }}</button>
          </div>
        </div>
        <div class="setting-card">
          <label class="setting-label">{{ $t('settings.webUiJwtSecret') }}</label>
          <div class="input-row">
            <input type="password" class="setting-input" v-model="form.webUiJwtSecret" />
            <button class="btn-save" @click="save('webui.jwt_secret', form.webUiJwtSecret)">{{ $t('settings.save') }}</button>
          </div>
        </div>
      </div>
    </section>

    <div v-if="savedMsg" class="saved-toast">{{ $t('settings.saved') }} ✓</div>
  </div>
</template>

<script setup>
import { onMounted, reactive, ref } from 'vue'
import { useAppStore } from '@/stores/app'
import { useApi } from '@/composables/useApi'
import { LANGS } from '@/i18n'

const app = useAppStore()
const api = useApi()
const savedMsg = ref(false)

const logLevels = ['trace', 'debug', 'info', 'warn', 'error']
const logFormats = ['pretty', 'json']
const orderTypes = ['GTC', 'GTD', 'FOK', 'FAK']
const sizeModes = ['proportional', 'fixed_pct']

const form = reactive({
  language: 'en',
  logLevel: 'info',
  logFormat: 'pretty',
  monitorEnabled: true,
  monitorInterval: 1000,
  tradesEnabled: false,
  tradesInterval: 5000,
  alertOnFill: true,
  alertOnCancel: true,
  trackPositions: true,
  tradingEnabled: false,
  maxPositionUsd: 100.0,
  slippagePct: 0.5,
  defaultOrderType: 'GTC',
  copytradingEnabled: false,
  copytradingInterval: 10000,
  sizeMode: 'proportional',
  webUiEnabled: true,
  webUiListen: '127.0.0.1:8080',
  webUiJwtSecret: '',
})

onMounted(async () => {
  try {
    const s = await api.getSettings()
    app.settings = s
    applySettings(s)
  } catch {}
})

function applySettings(s) {
  if (s.ui?.language)                            form.language = s.ui.language
  if (s.log?.level)                              form.logLevel = s.log.level
  if (s.log?.format)                             form.logFormat = s.log.format
  form.monitorEnabled = !!s.monitor?.enabled
  if (s.monitor?.poll_interval_ms)               form.monitorInterval = s.monitor.poll_interval_ms
  form.tradesEnabled = !!s.monitor?.trades?.enabled
  if (s.monitor?.trades?.poll_interval_ms)       form.tradesInterval = s.monitor.trades.poll_interval_ms
  form.alertOnFill = !!s.monitor?.trades?.alert_on_fill
  form.alertOnCancel = !!s.monitor?.trades?.alert_on_cancel
  form.trackPositions = !!s.monitor?.trades?.track_positions
  form.tradingEnabled = !!s.trading?.enabled
  if (s.trading?.max_position_usd)               form.maxPositionUsd = s.trading.max_position_usd
  if (s.trading?.slippage_pct)                   form.slippagePct = s.trading.slippage_pct
  if (s.trading?.default_order_type)             form.defaultOrderType = s.trading.default_order_type
  form.copytradingEnabled = !!s.copytrading?.enabled
  if (s.copytrading?.poll_interval_ms)           form.copytradingInterval = s.copytrading.poll_interval_ms
  if (s.copytrading?.size_mode)                  form.sizeMode = s.copytrading.size_mode
  form.webUiEnabled = !!s.webui?.enabled
  if (s.webui?.listen)                           form.webUiListen = s.webui.listen
  if (s.webui?.jwt_secret)                       form.webUiJwtSecret = s.webui.jwt_secret
}

async function save(key, value) {
  try {
    await api.postSettings(key, String(value))
    savedMsg.value = true
    setTimeout(() => { savedMsg.value = false }, 2000)
  } catch {}
}
</script>

<style scoped>
.view { display: flex; flex-direction: column; gap: 2rem; position: relative; }
.view-title { font-size: 1.4rem; font-weight: 700; }

.settings-section { display: flex; flex-direction: column; gap: 0.75rem; }

.section-title {
  font-size: 0.7rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.1em;
  color: var(--accent);
  padding-bottom: 0.5rem;
  border-bottom: 1px solid var(--border);
}

.settings-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(260px, 1fr));
  gap: 0.75rem;
}

.setting-card {
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  padding: 0.875rem 1rem;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.setting-label {
  font-size: 0.72rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.06em;
  color: var(--text-secondary);
}

.setting-input {
  background: var(--bg-primary);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  color: var(--text-primary);
  padding: 0.4rem 0.65rem;
  font-size: 0.875rem;
  font-family: var(--font-mono);
  outline: none;
  width: 100%;
  transition: border-color var(--transition);
}
.setting-input:focus { border-color: var(--accent); }

.input-row { display: flex; gap: 0.4rem; }
.input-row .setting-input { flex: 1; }

.btn-save {
  background: var(--accent);
  color: #fff;
  border: none;
  border-radius: var(--radius);
  padding: 0.4rem 0.7rem;
  font-size: 0.8rem;
  cursor: pointer;
  white-space: nowrap;
  transition: background var(--transition);
}
.btn-save:hover { background: var(--accent-hover); }

/* Toggle switch */
.toggle { display: flex; align-items: center; cursor: pointer; }
.toggle input { display: none; }
.toggle-track {
  width: 38px; height: 20px;
  background: var(--text-muted);
  border-radius: 10px;
  position: relative;
  transition: background var(--transition);
}
.toggle input:checked ~ .toggle-track { background: var(--accent); }
.toggle-thumb {
  position: absolute;
  width: 14px; height: 14px;
  background: #fff;
  border-radius: 50%;
  top: 3px; left: 3px;
  transition: left var(--transition);
}
.toggle input:checked ~ .toggle-track .toggle-thumb { left: 21px; }

.saved-toast {
  position: fixed;
  bottom: 1.5rem; right: 1.5rem;
  background: var(--success);
  color: #fff;
  padding: 0.6rem 1.2rem;
  border-radius: var(--radius);
  font-size: 0.875rem;
  font-weight: 600;
  box-shadow: var(--shadow);
  animation: fadeUp 0.2s ease;
}
@keyframes fadeUp {
  from { opacity: 0; transform: translateY(8px); }
  to   { opacity: 1; transform: translateY(0); }
}
</style>
