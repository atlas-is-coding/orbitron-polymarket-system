<template>
  <div class="view">
    <div class="page-header anim-in">
      <h2 class="view-title">{{ $t('settings.title') }}</h2>
    </div>

    <!-- UI -->
    <section class="settings-section anim-in">
      <div class="section-header"><span>{{ $t('settings.sectionUi') }}</span></div>
      <div class="settings-grid">
        <div class="setting-card">
          <label class="setting-label">{{ $t('settings.language') }}</label>
          <select class="setting-input" v-model="form.language" @change="save('ui.language', form.language)">
            <option v-for="l in LANGS" :key="l" :value="l">{{ l.toUpperCase() }}</option>
          </select>
        </div>
      </div>
    </section>

    <!-- Auth -->
    <section class="settings-section anim-in">
      <div class="section-header"><span>{{ $t('settings.sectionAuth') }}</span></div>
      <div class="settings-grid">
        <div class="setting-card">
          <label class="setting-label">{{ $t('settings.chainId') }}</label>
          <div class="input-row">
            <select class="setting-input" v-model.number="form.chainId" @change="save('auth.chain_id', form.chainId)">
              <option :value="137">137 (Polygon Mainnet)</option>
              <option :value="80002">80002 (Amoy Testnet)</option>
            </select>
          </div>
        </div>
      </div>
    </section>

    <!-- API -->
    <section class="settings-section anim-in">
      <div class="section-header"><span>{{ $t('settings.sectionApi') }}</span></div>
      <div class="settings-grid">
        <div class="setting-card">
          <label class="setting-label">{{ $t('settings.apiTimeout') }}</label>
          <div class="input-row">
            <input type="number" class="setting-input" v-model.number="form.apiTimeout" min="1" step="1" />
            <button class="btn-save" @click="save('api.timeout_sec', form.apiTimeout)">{{ $t('settings.save') }}</button>
          </div>
        </div>
        <div class="setting-card">
          <label class="setting-label">{{ $t('settings.apiMaxRetries') }}</label>
          <div class="input-row">
            <input type="number" class="setting-input" v-model.number="form.apiMaxRetries" min="0" step="1" />
            <button class="btn-save" @click="save('api.max_retries', form.apiMaxRetries)">{{ $t('settings.save') }}</button>
          </div>
        </div>
      </div>
    </section>

    <!-- Logging -->
    <section class="settings-section anim-in">
      <div class="section-header"><span>{{ $t('settings.sectionLog') }}</span></div>
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
        <div class="setting-card">
          <label class="setting-label">{{ $t('settings.logFile') }}</label>
          <input class="setting-input" type="text" v-model="form.logFile" placeholder="./polytrade.log"
            @change="save('log.file', form.logFile)" />
        </div>
      </div>
    </section>

    <!-- Monitor -->
    <section class="settings-section anim-in">
      <div class="section-header"><span>{{ $t('settings.sectionMonitor') }}</span></div>
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

    <!-- Trades Monitor -->
    <section class="settings-section anim-in">
      <div class="section-header"><span>{{ $t('settings.sectionTradesMonitor') }}</span></div>
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
        <div class="setting-card">
          <label class="setting-label">{{ $t('settings.tradesLimit') }}</label>
          <div class="input-row">
            <input type="number" class="setting-input" v-model.number="form.tradesLimit" min="1" step="1" />
            <button class="btn-save" @click="save('monitor.trades.trades_limit', form.tradesLimit)">{{ $t('settings.save') }}</button>
          </div>
        </div>
      </div>
    </section>

    <!-- Trading -->
    <section class="settings-section anim-in">
      <div class="section-header"><span>{{ $t('settings.sectionTrading') }}</span></div>
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
        <div class="setting-card">
          <label class="setting-label">{{ $t('settings.negRisk') }}</label>
          <label class="toggle">
            <input type="checkbox" v-model="form.negRisk" @change="save('trading.neg_risk', form.negRisk)" />
            <span class="toggle-track"><span class="toggle-thumb" /></span>
          </label>
        </div>
      </div>
    </section>

    <!-- Copytrading -->
    <section class="settings-section anim-in">
      <div class="section-header"><span>{{ $t('settings.sectionCopytrading') }}</span></div>
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

    <!-- Telegram -->
    <section class="settings-section anim-in">
      <div class="section-header"><span>{{ $t('settings.sectionTelegram') }}</span></div>
      <div class="settings-grid">
        <div class="setting-card">
          <label class="setting-label">{{ $t('settings.telegramEnabled') }}</label>
          <label class="toggle">
            <input type="checkbox" v-model="form.telegramEnabled" @change="save('telegram.enabled', form.telegramEnabled)" />
            <span class="toggle-track"><span class="toggle-thumb" /></span>
          </label>
        </div>
        <div class="setting-card">
          <label class="setting-label">{{ $t('settings.telegramToken') }}</label>
          <div class="input-row">
            <input type="password" class="setting-input" v-model="form.telegramToken" placeholder="***" />
            <button class="btn-save" @click="save('telegram.bot_token', form.telegramToken)">{{ $t('settings.save') }}</button>
          </div>
        </div>
        <div class="setting-card">
          <label class="setting-label">{{ $t('settings.telegramAdminChatId') }}</label>
          <div class="input-row">
            <input type="text" class="setting-input" v-model="form.telegramAdminChatId" />
            <button class="btn-save" @click="save('telegram.admin_chat_id', form.telegramAdminChatId)">{{ $t('settings.save') }}</button>
          </div>
        </div>
      </div>
    </section>

    <!-- Database -->
    <section class="settings-section anim-in">
      <div class="section-header"><span>{{ $t('settings.sectionDatabase') }}</span></div>
      <div class="settings-grid">
        <div class="setting-card">
          <label class="setting-label">{{ $t('settings.databaseEnabled') }}</label>
          <label class="toggle">
            <input type="checkbox" v-model="form.databaseEnabled" @change="save('database.enabled', form.databaseEnabled)" />
            <span class="toggle-track"><span class="toggle-thumb" /></span>
          </label>
        </div>
        <div class="setting-card">
          <label class="setting-label">{{ $t('settings.databasePath') }}</label>
          <div class="input-row">
            <input type="text" class="setting-input" v-model="form.databasePath" placeholder="bot.db" />
            <button class="btn-save" @click="save('database.path', form.databasePath)">{{ $t('settings.save') }}</button>
          </div>
        </div>
      </div>
    </section>

    <!-- Web UI -->
    <section class="settings-section anim-in">
      <div class="section-header"><span>{{ $t('settings.sectionWebUi') }}</span></div>
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

    <!-- Proxy -->
    <section class="settings-section anim-in">
      <div class="section-header"><span>{{ $t('settings.sectionProxy') }}</span></div>
      <div class="settings-grid">
        <div class="setting-card">
          <label class="setting-label">{{ $t('settings.proxyEnabled') }}</label>
          <label class="toggle">
            <input type="checkbox" v-model="form.proxyEnabled" @change="save('proxy.enabled', form.proxyEnabled)" />
            <span class="toggle-track"><span class="toggle-thumb" /></span>
          </label>
        </div>
        <div class="setting-card">
          <label class="setting-label">{{ $t('settings.proxyType') }}</label>
          <select class="setting-input" v-model="form.proxyType" @change="save('proxy.type', form.proxyType)">
            <option v-for="t in proxyTypes" :key="t" :value="t">{{ t }}</option>
          </select>
        </div>
        <div class="setting-card">
          <label class="setting-label">{{ $t('settings.proxyAddr') }}</label>
          <div class="input-row">
            <input type="text" class="setting-input" v-model="form.proxyAddr" placeholder="host:port" />
            <button class="btn-save" @click="save('proxy.addr', form.proxyAddr)">{{ $t('settings.save') }}</button>
          </div>
        </div>
        <div class="setting-card">
          <label class="setting-label">{{ $t('settings.proxyUsername') }}</label>
          <div class="input-row">
            <input type="text" class="setting-input" v-model="form.proxyUsername" />
            <button class="btn-save" @click="save('proxy.username', form.proxyUsername)">{{ $t('settings.save') }}</button>
          </div>
        </div>
        <div class="setting-card">
          <label class="setting-label">{{ $t('settings.proxyPassword') }}</label>
          <div class="input-row">
            <input type="password" class="setting-input" v-model="form.proxyPassword" />
            <button class="btn-save" @click="save('proxy.password', form.proxyPassword)">{{ $t('settings.save') }}</button>
          </div>
        </div>
      </div>
    </section>

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
import { LANGS } from '@/i18n'

const app = useAppStore()
const api = useApi()
const savedMsg = ref(false)
const { settingsStale } = storeToRefs(app)

watch(settingsStale, async (stale) => {
  if (!stale) return
  try { const s = await api.getSettings(); applySettings(s) } catch {}
  app.settingsStale = false
})

const logLevels = ['trace', 'debug', 'info', 'warn', 'error']
const logFormats = ['pretty', 'json']
const orderTypes = ['GTC', 'GTD', 'FOK', 'FAK']
const sizeModes = ['proportional', 'fixed_pct']
const proxyTypes = ['socks5', 'http']

const form = reactive({
  language: 'en', logLevel: 'info', logFormat: 'pretty', logFile: '',
  chainId: 137, apiTimeout: 10, apiMaxRetries: 3,
  monitorEnabled: true, monitorInterval: 1000,
  tradesEnabled: false, tradesInterval: 5000,
  alertOnFill: true, alertOnCancel: true, trackPositions: true, tradesLimit: 50,
  tradingEnabled: false, maxPositionUsd: 100.0, slippagePct: 0.5, defaultOrderType: 'GTC', negRisk: false,
  copytradingEnabled: false, copytradingInterval: 10000, sizeMode: 'proportional',
  telegramEnabled: false, telegramToken: '', telegramAdminChatId: '',
  databaseEnabled: false, databasePath: 'bot.db',
  webUiEnabled: true, webUiListen: '127.0.0.1:8080', webUiJwtSecret: '',
  proxyEnabled: false, proxyType: 'socks5', proxyAddr: '', proxyUsername: '', proxyPassword: '',
})

onMounted(async () => {
  try { const s = await api.getSettings(); app.settings = s; applySettings(s) } catch {}
})

function applySettings(s) {
  if (s.ui?.language)                           form.language = s.ui.language
  if (s.auth?.chain_id)                         form.chainId = s.auth.chain_id
  if (s.api?.timeout_sec)                       form.apiTimeout = s.api.timeout_sec
  if (s.api?.max_retries)                       form.apiMaxRetries = s.api.max_retries
  if (s.log?.level)                             form.logLevel = s.log.level
  if (s.log?.format)                            form.logFormat = s.log.format
  if (s.log?.file !== undefined)                form.logFile = s.log.file ?? ''
  form.monitorEnabled = !!s.monitor?.enabled
  if (s.monitor?.poll_interval_ms)              form.monitorInterval = s.monitor.poll_interval_ms
  form.tradesEnabled = !!s.monitor?.trades?.enabled
  if (s.monitor?.trades?.poll_interval_ms)      form.tradesInterval = s.monitor.trades.poll_interval_ms
  form.alertOnFill = !!s.monitor?.trades?.alert_on_fill
  form.alertOnCancel = !!s.monitor?.trades?.alert_on_cancel
  form.trackPositions = !!s.monitor?.trades?.track_positions
  if (s.monitor?.trades?.trades_limit)          form.tradesLimit = s.monitor.trades.trades_limit
  form.tradingEnabled = !!s.trading?.enabled
  if (s.trading?.max_position_usd)              form.maxPositionUsd = s.trading.max_position_usd
  if (s.trading?.slippage_pct)                  form.slippagePct = s.trading.slippage_pct
  if (s.trading?.default_order_type)            form.defaultOrderType = s.trading.default_order_type
  form.negRisk = !!s.trading?.neg_risk
  form.copytradingEnabled = !!s.copytrading?.enabled
  if (s.copytrading?.poll_interval_ms)          form.copytradingInterval = s.copytrading.poll_interval_ms
  if (s.copytrading?.size_mode)                 form.sizeMode = s.copytrading.size_mode
  form.telegramEnabled = !!s.telegram?.enabled
  if (s.telegram?.bot_token && s.telegram.bot_token !== '***') form.telegramToken = s.telegram.bot_token
  if (s.telegram?.admin_chat_id)                form.telegramAdminChatId = s.telegram.admin_chat_id
  form.databaseEnabled = !!s.database?.enabled
  if (s.database?.path)                         form.databasePath = s.database.path
  form.webUiEnabled = !!s.webui?.enabled
  if (s.webui?.listen)                          form.webUiListen = s.webui.listen
  // Don't populate password fields with masked "***" value — leave blank so user must re-enter to change
  if (s.webui?.jwt_secret && s.webui.jwt_secret !== '***') form.webUiJwtSecret = s.webui.jwt_secret
  form.proxyEnabled = !!s.proxy?.enabled
  if (s.proxy?.type)     form.proxyType = s.proxy.type
  if (s.proxy?.addr)     form.proxyAddr = s.proxy.addr
  if (s.proxy?.username) form.proxyUsername = s.proxy.username
  if (s.proxy?.password && s.proxy.password !== '***') form.proxyPassword = s.proxy.password
}

async function save(key, value) {
  // Never send back a masked sentinel value — user must explicitly type a new value to change secrets
  if (String(value) === '***') return
  try {
    await api.postSettings(key, String(value))
    savedMsg.value = true
    setTimeout(() => { savedMsg.value = false }, 2000)
  } catch {}
}
</script>

<style scoped>
.view { display: flex; flex-direction: column; gap: 1.75rem; position: relative; }

.page-header { }
.view-title { font-size: 1rem; font-weight: 700; letter-spacing: 0.04em; color: var(--text-bright); }

/* Section */
.settings-section { display: flex; flex-direction: column; gap: 0.75rem; }

/* Section header */
.section-header {
  display: flex; align-items: center; gap: 0.5rem;
  font-size: 1.00rem; text-transform: uppercase; letter-spacing: 0.12em;
  color: var(--accent); font-weight: 700;
  padding-bottom: 0.5rem;
  border-bottom: 1px solid var(--border);
}

/* Grid */
.settings-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(250px, 1fr));
  gap: 0.65rem;
}

/* Card */
.setting-card {
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  padding: 0.85rem 1rem;
  display: flex; flex-direction: column; gap: 0.5rem;
  transition: border-color var(--transition);
}
.setting-card:hover { border-color: rgba(0, 200, 255, 0.30); }

/* Label */
.setting-label {
  font-size: 0.90rem; font-weight: 600;
  text-transform: uppercase; letter-spacing: 0.08em;
  color: var(--text-secondary);
}

/* Input */
.setting-input {
  background: var(--bg-input);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  color: var(--text-primary);
  padding: 0.4rem 0.65rem;
  font-size: 0.96rem; font-family: var(--font-mono);
  outline: none; width: 100%;
  transition: border-color var(--transition);
}
.setting-input:focus { border-color: var(--accent); box-shadow: 0 0 0 1px rgba(0,200,255,0.12); }

select.setting-input { cursor: pointer; }

.input-row { display: flex; gap: 0.4rem; }
.input-row .setting-input { flex: 1; }

/* Save button */
.btn-save {
  background: transparent;
  border: 1px solid var(--accent);
  color: var(--accent);
  border-radius: var(--radius);
  padding: 0.4rem 0.7rem;
  font-size: 0.86rem; font-weight: 600;
  cursor: pointer; white-space: nowrap;
  font-family: var(--font-mono);
  transition: all var(--transition);
}
.btn-save:hover { background: var(--accent); color: #000; box-shadow: var(--accent-glow); }

/* Toggle */
.toggle { display: flex; align-items: center; cursor: pointer; }
.toggle input { display: none; }
.toggle-track {
  width: 36px; height: 18px;
  background: var(--border);
  border-radius: 1px; position: relative;
  transition: background var(--transition);
  border: 1px solid var(--border);
}
.toggle input:checked ~ .toggle-track { background: var(--accent); border-color: var(--accent); }
.toggle-thumb {
  position: absolute;
  width: 12px; height: 12px;
  background: var(--text-muted);
  border-radius: 1px;
  top: 2px; left: 2px;
  transition: left var(--transition), background var(--transition);
}
.toggle input:checked ~ .toggle-track .toggle-thumb { left: 20px; background: #000; }

/* Saved toast */
.saved-toast {
  position: fixed;
  bottom: 1.5rem; right: 1.5rem;
  background: var(--success);
  color: #000;
  padding: 0.5rem 1.1rem;
  border-radius: var(--radius);
  font-size: 0.92rem; font-weight: 700;
  box-shadow: 0 2px 16px rgba(16, 217, 148, 0.35);
  font-family: var(--font-mono); letter-spacing: 0.04em;
}
.fade-enter-active { animation: fadeSlideUp 0.18s ease both; }
.fade-leave-active { animation: fadeSlideUp 0.15s ease reverse both; }
</style>
