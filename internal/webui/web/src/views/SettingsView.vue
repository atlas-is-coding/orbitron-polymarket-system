<template>
  <div class="view">
    <h2 class="view-title">{{ $t('settings.title') }}</h2>

    <div class="settings-grid">
      <!-- Language -->
      <div class="setting-card">
        <label class="setting-label">{{ $t('settings.language') }}</label>
        <select class="setting-input" v-model="form.language" @change="save('ui.language', form.language)">
          <option v-for="l in LANGS" :key="l" :value="l">{{ l.toUpperCase() }}</option>
        </select>
      </div>

      <!-- Log level -->
      <div class="setting-card">
        <label class="setting-label">{{ $t('settings.logLevel') }}</label>
        <select class="setting-input" v-model="form.logLevel" @change="save('log.level', form.logLevel)">
          <option v-for="l in logLevels" :key="l" :value="l">{{ l }}</option>
        </select>
      </div>

      <!-- Monitor interval -->
      <div class="setting-card">
        <label class="setting-label">{{ $t('settings.monitorInterval') }}</label>
        <div class="input-row">
          <input type="number" class="setting-input" v-model.number="form.monitorInterval" min="100" step="100" />
          <button class="btn-save" @click="save('monitor.poll_interval_ms', form.monitorInterval)">{{ $t('settings.save') }}</button>
        </div>
      </div>

      <!-- Trades interval -->
      <div class="setting-card">
        <label class="setting-label">{{ $t('settings.tradesInterval') }}</label>
        <div class="input-row">
          <input type="number" class="setting-input" v-model.number="form.tradesInterval" min="1000" step="1000" />
          <button class="btn-save" @click="save('monitor.trades.poll_interval_ms', form.tradesInterval)">{{ $t('settings.save') }}</button>
        </div>
      </div>

      <!-- Copytrading interval -->
      <div class="setting-card">
        <label class="setting-label">{{ $t('settings.copytradingInterval') }}</label>
        <div class="input-row">
          <input type="number" class="setting-input" v-model.number="form.copytradingInterval" min="1000" step="1000" />
          <button class="btn-save" @click="save('copytrading.poll_interval_ms', form.copytradingInterval)">{{ $t('settings.save') }}</button>
        </div>
      </div>

      <!-- Trading enabled -->
      <div class="setting-card">
        <label class="setting-label">{{ $t('settings.tradingEnabled') }}</label>
        <label class="toggle">
          <input type="checkbox" v-model="form.tradingEnabled" @change="save('trading.enabled', form.tradingEnabled)" />
          <span class="toggle-track"><span class="toggle-thumb" /></span>
        </label>
      </div>

      <!-- Copytrading enabled -->
      <div class="setting-card">
        <label class="setting-label">{{ $t('settings.copytradingEnabled') }}</label>
        <label class="toggle">
          <input type="checkbox" v-model="form.copytradingEnabled" @change="save('copytrading.enabled', form.copytradingEnabled)" />
          <span class="toggle-track"><span class="toggle-thumb" /></span>
        </label>
      </div>
    </div>

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

const form = reactive({
  language: 'en',
  logLevel: 'info',
  monitorInterval: 1000,
  tradesInterval: 5000,
  copytradingInterval: 10000,
  tradingEnabled: false,
  copytradingEnabled: false,
})

onMounted(async () => {
  try {
    const s = await api.getSettings()
    app.settings = s
    applySettings(s)
  } catch {}
})

function applySettings(s) {
  if (s.ui?.language)                    form.language = s.ui.language
  if (s.log?.level)                      form.logLevel = s.log.level
  if (s.monitor?.poll_interval_ms)       form.monitorInterval = s.monitor.poll_interval_ms
  if (s.monitor?.trades?.poll_interval_ms) form.tradesInterval = s.monitor.trades.poll_interval_ms
  if (s.copytrading?.poll_interval_ms)   form.copytradingInterval = s.copytrading.poll_interval_ms
  form.tradingEnabled = !!s.trading?.enabled
  form.copytradingEnabled = !!s.copytrading?.enabled
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
.view { display: flex; flex-direction: column; gap: 1.5rem; position: relative; }
.view-title { font-size: 1.4rem; font-weight: 700; }

.settings-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 1rem;
}

.setting-card {
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  padding: 1rem 1.25rem;
  display: flex;
  flex-direction: column;
  gap: 0.6rem;
}

.setting-label {
  font-size: 0.75rem;
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
  padding: 0.45rem 0.7rem;
  font-size: 0.9rem;
  font-family: var(--font-mono);
  outline: none;
  width: 100%;
  transition: border-color var(--transition);
}
.setting-input:focus { border-color: var(--accent); }

.input-row { display: flex; gap: 0.5rem; }
.input-row .setting-input { flex: 1; }

.btn-save {
  background: var(--accent);
  color: #fff;
  border: none;
  border-radius: var(--radius);
  padding: 0.45rem 0.8rem;
  font-size: 0.82rem;
  cursor: pointer;
  white-space: nowrap;
  transition: background var(--transition);
}
.btn-save:hover { background: var(--accent-hover); }

/* Toggle switch */
.toggle { display: flex; align-items: center; cursor: pointer; }
.toggle input { display: none; }
.toggle-track {
  width: 40px; height: 22px;
  background: var(--text-muted);
  border-radius: 11px;
  position: relative;
  transition: background var(--transition);
}
.toggle input:checked ~ .toggle-track { background: var(--accent); }
.toggle-thumb {
  position: absolute;
  width: 16px; height: 16px;
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
