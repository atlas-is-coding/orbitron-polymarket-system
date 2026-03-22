<template>
  <div class="settings-layout">
    <!-- Left nav -->
    <nav class="settings-nav">
      <a v-for="s in sections" :key="s.id" class="snav-item" :class="{ active: activeSection === s.id }"
        @click.prevent="scrollTo(s.id)">{{ s.label }}</a>
    </nav>

    <!-- Form area -->
    <div class="settings-body" ref="bodyEl" @scroll="onScroll">
      <!-- Save bar -->
      <Transition name="savebar">
        <div v-if="store.isDirty" class="save-bar">
          <span class="dirty-badge">Unsaved changes</span>
          <div class="save-actions">
            <button class="btn-ghost" @click="store.reset()">Reset</button>
            <button class="btn-primary" :disabled="saving" @click="doSave">
              {{ saving ? 'Saving…' : 'Save Config' }}
            </button>
          </div>
        </div>
      </Transition>

      <!-- Auth & Keys -->
      <section id="s-auth" class="s-section">
        <div class="s-heading">Auth &amp; Keys</div>
        <div class="s-row">
          <label>Private Key</label>
          <input class="s-input" type="password" :value="store.local.auth?.private_key || ''"
            @input="store.set('auth.private_key', $event.target.value)" placeholder="64-hex, no 0x prefix" />
        </div>
        <div class="s-row">
          <label>API Key</label>
          <input class="s-input" :value="store.local.auth?.api_key || ''"
            @input="store.set('auth.api_key', $event.target.value)" placeholder="CLOB API key" />
        </div>
        <div class="s-row">
          <label>API Secret</label>
          <input class="s-input" type="password" :value="store.local.auth?.api_secret || ''"
            @input="store.set('auth.api_secret', $event.target.value)" />
        </div>
        <div class="s-row">
          <label>API Passphrase</label>
          <input class="s-input" type="password" :value="store.local.auth?.passphrase || ''"
            @input="store.set('auth.passphrase', $event.target.value)" />
        </div>
        <div class="s-row">
          <label>Chain ID</label>
          <select class="s-input" :value="store.local.auth?.chain_id || 137"
            @change="store.set('auth.chain_id', parseInt($event.target.value))">
            <option :value="137">137 — Polygon Mainnet</option>
            <option :value="80002">80002 — Amoy Testnet</option>
          </select>
        </div>
      </section>

      <!-- Network -->
      <section id="s-network" class="s-section">
        <div class="s-heading">Network</div>
        <div v-for="field in networkFields" :key="field.key" class="s-row">
          <label>{{ field.label }}</label>
          <div class="s-input-row">
            <input class="s-input" :value="getNestedVal(field.key)"
              @input="store.set(field.key, $event.target.value)" :placeholder="field.placeholder" />
            <button class="btn-test" :class="testState[field.key]" @click="testUrl(field.key)">
              {{ testState[field.key] === 'ok' ? '✓' : testState[field.key] === 'err' ? '✗' : 'TEST' }}
            </button>
          </div>
        </div>
      </section>

      <!-- Trading Engine -->
      <section id="s-trading" class="s-section">
        <div class="s-heading">Trading Engine</div>
        <div class="s-row">
          <label>Enabled</label>
          <label class="toggle">
            <input type="checkbox" :checked="store.local.trading?.enabled"
              @change="store.set('trading.enabled', $event.target.checked)" />
            <span class="track"></span>
          </label>
        </div>
        <div class="s-row">
          <label>Max Position USD</label>
          <input class="s-input s-input-sm" type="number" :value="store.local.trading?.max_position_usd || 100"
            @input="store.set('trading.max_position_usd', parseFloat($event.target.value))" />
        </div>
        <div class="s-row">
          <label>Max Daily Trades</label>
          <input class="s-input s-input-sm" type="number" :value="store.local.trading?.max_daily_trades || 50"
            @input="store.set('trading.max_daily_trades', parseInt($event.target.value))" />
        </div>
        <div class="s-row">
          <label>Order Fill Timeout (s)</label>
          <input class="s-input s-input-sm" type="number" :value="store.local.trading?.order_fill_timeout_s || 60"
            @input="store.set('trading.order_fill_timeout_s', parseInt($event.target.value))" />
        </div>
      </section>

      <!-- Telegram -->
      <section id="s-telegram" class="s-section">
        <div class="s-heading">Telegram</div>
        <div class="s-row">
          <label>Bot Token</label>
          <input class="s-input" type="password" :value="store.local.telegram?.token || ''"
            @input="store.set('telegram.token', $event.target.value)" placeholder="123456:ABC-..." />
        </div>
        <div class="s-row">
          <label>Chat ID</label>
          <input class="s-input s-input-sm" :value="store.local.telegram?.chat_id || ''"
            @input="store.set('telegram.chat_id', $event.target.value)" placeholder="-100..." />
        </div>
        <div class="s-row">
          <label>Notifications</label>
          <label class="toggle">
            <input type="checkbox" :checked="store.local.telegram?.notify_fills"
              @change="store.set('telegram.notify_fills', $event.target.checked)" />
            <span class="track"></span>
          </label>
        </div>
      </section>

      <!-- Logging -->
      <section id="s-logging" class="s-section">
        <div class="s-heading">Logging</div>
        <div class="s-row">
          <label>Log Level</label>
          <select class="s-input" :value="store.local.log?.level || 'info'"
            @change="store.set('log.level', $event.target.value)">
            <option value="debug">debug</option>
            <option value="info">info</option>
            <option value="warn">warn</option>
            <option value="error">error</option>
          </select>
        </div>
        <div class="s-row">
          <label>Format</label>
          <select class="s-input" :value="store.local.log?.format || 'pretty'"
            @change="store.set('log.format', $event.target.value)">
            <option value="pretty">pretty</option>
            <option value="json">json</option>
          </select>
        </div>
        <div class="s-row">
          <label>Log to File</label>
          <label class="toggle">
            <input type="checkbox" :checked="store.local.log?.file_enabled"
              @change="store.set('log.file_enabled', $event.target.checked)" />
            <span class="track"></span>
          </label>
        </div>
        <div class="s-row" v-if="store.local.log?.file_enabled">
          <label>Log File Path</label>
          <input class="s-input" :value="store.local.log?.file_path || 'bot.log'"
            @input="store.set('log.file_path', $event.target.value)" />
        </div>
      </section>

      <!-- Interface -->
      <section id="s-interface" class="s-section">
        <div class="s-heading">Interface</div>
        <div class="s-row">
          <label>Language</label>
          <select class="s-input" :value="currentLang" @change="changeLang($event.target.value)">
            <option v-for="l in LANGS" :key="l" :value="l">{{ l.toUpperCase() }}</option>
          </select>
        </div>
        <div class="s-row">
          <label>Poll Interval (ms)</label>
          <input class="s-input s-input-sm" type="number" :value="store.local.webui?.poll_interval_ms || 5000"
            @input="store.set('webui.poll_interval_ms', parseInt($event.target.value))" />
        </div>
      </section>

      <!-- Danger Zone -->
      <section id="s-danger" class="s-section s-danger">
        <div class="s-heading danger-heading">Danger Zone</div>
        <div class="danger-actions">
          <div class="danger-row">
            <div>
              <div class="danger-title">Cancel All Orders</div>
              <div class="danger-desc">Immediately cancel every open order across all wallets.</div>
            </div>
            <button class="btn-danger" @click="confirmAction('cancelAllOrders')">Cancel All</button>
          </div>
          <div class="danger-row">
            <div>
              <div class="danger-title">Stop All Strategies</div>
              <div class="danger-desc">Halt all running strategies. Does not cancel open orders.</div>
            </div>
            <button class="btn-danger" @click="confirmAction('stopAllStrategies')">Stop All</button>
          </div>
          <div class="danger-row">
            <div>
              <div class="danger-title">Reset Config</div>
              <div class="danger-desc">Restore default configuration. This cannot be undone.</div>
            </div>
            <button class="btn-danger" @click="confirmAction('resetConfig')">Reset</button>
          </div>
        </div>
      </section>
    </div>

    <!-- Confirm dialog -->
    <div v-if="confirmTarget" class="modal-backdrop" @click.self="confirmTarget = null">
      <div class="modal">
        <div class="modal-title">Are you sure?</div>
        <div class="modal-body">{{ confirmMessages[confirmTarget] }}</div>
        <div class="modal-actions">
          <button class="btn-ghost" @click="confirmTarget = null">Cancel</button>
          <button class="btn-danger" @click="runConfirmed">Confirm</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useSettingsStore } from '@/stores/settings'
import { useApi } from '@/composables/useApi'
import { LANGS } from '@/i18n'

const { locale } = useI18n()
const store = useSettingsStore()
const api = useApi()

const saving = ref(false)
const bodyEl = ref(null)
const activeSection = ref('s-auth')
const testState = ref({})
const confirmTarget = ref(null)
const currentLang = ref(locale.value)

const sections = [
  { id: 's-auth',      label: 'Auth & Keys' },
  { id: 's-network',   label: 'Network' },
  { id: 's-trading',   label: 'Trading Engine' },
  { id: 's-telegram',  label: 'Telegram' },
  { id: 's-logging',   label: 'Logging' },
  { id: 's-interface', label: 'Interface' },
  { id: 's-danger',    label: 'Danger Zone' },
]

const networkFields = [
  { key: 'network.clob_url',  label: 'CLOB URL',  placeholder: 'https://clob.polymarket.com' },
  { key: 'network.gamma_url', label: 'Gamma URL', placeholder: 'https://gamma-api.polymarket.com' },
  { key: 'network.data_url',  label: 'Data URL',  placeholder: 'https://data-api.polymarket.com' },
  { key: 'network.ws_url',    label: 'WS URL',    placeholder: 'wss://ws-subscriptions-clob.polymarket.com/ws/' },
]

const confirmMessages = {
  cancelAllOrders:   'This will cancel all open orders across all wallets immediately.',
  stopAllStrategies: 'This will stop all running strategies immediately.',
  resetConfig:       'This will reset your configuration to defaults. All custom settings will be lost.',
}

function getNestedVal(dotKey) {
  const parts = dotKey.split('.')
  let cur = store.local
  for (const p of parts) cur = cur?.[p]
  return cur || ''
}

function changeLang(lang) {
  currentLang.value = lang
  locale.value = lang
  localStorage.setItem('lang', lang)
}

function scrollTo(id) {
  const el = document.getElementById(id)
  if (el) el.scrollIntoView({ behavior: 'smooth', block: 'start' })
}

function onScroll() {
  for (const s of [...sections].reverse()) {
    const el = document.getElementById(s.id)
    if (el && el.getBoundingClientRect().top <= 120) {
      activeSection.value = s.id
      return
    }
  }
}

async function doSave() {
  saving.value = true
  try { await store.save() } finally { saving.value = false }
}

async function testUrl(key) {
  const url = getNestedVal(key)
  if (!url) return
  testState.value = { ...testState.value, [key]: 'loading' }
  try {
    await api.testEndpoint(url)
    testState.value = { ...testState.value, [key]: 'ok' }
  } catch {
    testState.value = { ...testState.value, [key]: 'err' }
  }
  setTimeout(() => {
    testState.value = { ...testState.value, [key]: null }
  }, 3000)
}

function confirmAction(action) {
  confirmTarget.value = action
}

async function runConfirmed() {
  const action = confirmTarget.value
  confirmTarget.value = null
  if (action === 'cancelAllOrders') {
    await api.cancelAll()
  } else if (action === 'stopAllStrategies') {
    const strats = await api.getStrategies().catch(() => ({ strategies: [] }))
    const list = strats.strategies || []
    await Promise.allSettled(list.filter(s => s.running).map(s => api.stopStrategy(s.key)))
  } else if (action === 'resetConfig') {
    store.reset()
  }
}

onMounted(() => store.load())
</script>

<style scoped>
.settings-layout {
  display: flex;
  height: 100%;
  overflow: hidden;
}

/* Left nav */
.settings-nav {
  width: 180px;
  flex-shrink: 0;
  border-right: 1px solid var(--border);
  padding: 24px 0;
  display: flex;
  flex-direction: column;
  gap: 2px;
  overflow-y: auto;
}
.snav-item {
  padding: 8px 20px;
  font-size: var(--font-size-sm, 12px);
  color: var(--fg-muted);
  cursor: pointer;
  border-left: 2px solid transparent;
  transition: all var(--transition);
  white-space: nowrap;
  user-select: none;
}
.snav-item:hover { color: var(--fg); background: rgba(255,255,255,0.03); }
.snav-item.active { color: var(--accent-bright); border-left-color: var(--accent-bright); background: rgba(124,58,237,0.06); }

/* Body */
.settings-body {
  flex: 1;
  overflow-y: auto;
  padding: 0 32px 60px;
  position: relative;
}

/* Save bar */
.save-bar {
  position: sticky;
  top: 0;
  z-index: 20;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 16px;
  background: var(--bg-panel);
  border-bottom: 1px solid var(--border);
  margin: 0 -32px 0;
  padding-left: 32px;
  padding-right: 32px;
}
.dirty-badge {
  font-size: var(--font-size-sm, 12px);
  color: var(--warning);
  font-weight: 600;
}
.save-actions { display: flex; gap: 8px; }

.savebar-enter-active, .savebar-leave-active { transition: opacity 0.2s, transform 0.2s; }
.savebar-enter-from, .savebar-leave-to { opacity: 0; transform: translateY(-8px); }

/* Sections */
.s-section {
  padding: 28px 0 16px;
  border-bottom: 1px solid var(--border);
}
.s-section:last-child { border-bottom: none; }
.s-heading {
  font-size: var(--font-size-xs, 10px);
  font-weight: 700;
  letter-spacing: 0.1em;
  text-transform: uppercase;
  color: var(--fg-dim);
  margin-bottom: 16px;
}

/* Rows */
.s-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 9px 0;
  border-bottom: 1px solid rgba(255,255,255,0.04);
}
.s-row:last-child { border-bottom: none; }
.s-row > label:first-child {
  font-size: var(--font-size-sm, 12px);
  color: var(--fg-muted);
  min-width: 180px;
}

.s-input {
  background: var(--bg-input, rgba(255,255,255,0.05));
  border: 1px solid var(--border);
  border-radius: var(--radius);
  color: var(--fg);
  font-size: var(--font-size-sm, 12px);
  font-family: var(--font-mono);
  padding: 5px 10px;
  width: 280px;
  outline: none;
  transition: border-color var(--transition);
}
.s-input:focus { border-color: var(--accent); }
.s-input-sm { width: 120px; }

.s-input-row { display: flex; gap: 8px; align-items: center; }

/* Test button */
.btn-test {
  padding: 5px 10px;
  font-size: 10px;
  font-weight: 700;
  letter-spacing: 0.05em;
  border-radius: var(--radius);
  border: 1px solid var(--border);
  background: transparent;
  color: var(--fg-muted);
  cursor: pointer;
  transition: all var(--transition);
  width: 50px;
  text-align: center;
}
.btn-test:hover { border-color: var(--accent); color: var(--accent-bright); }
.btn-test.ok { color: var(--success); border-color: var(--success); }
.btn-test.err { color: var(--danger); border-color: var(--danger); }

/* Toggle */
.toggle { position: relative; display: inline-block; width: 36px; height: 20px; cursor: pointer; }
.toggle input { display: none; }
.track {
  position: absolute; inset: 0;
  background: rgba(255,255,255,0.1);
  border-radius: 10px;
  border: 1px solid var(--border);
  transition: background var(--transition);
}
.track::after {
  content: '';
  position: absolute;
  width: 14px; height: 14px;
  border-radius: 50%;
  background: var(--fg-dim);
  top: 2px; left: 2px;
  transition: transform var(--transition), background var(--transition);
}
.toggle input:checked + .track { background: rgba(124,58,237,0.3); border-color: var(--accent); }
.toggle input:checked + .track::after { transform: translateX(16px); background: var(--accent-bright); }

/* Danger zone */
.s-danger { }
.danger-heading { color: var(--danger) !important; }
.danger-actions { display: flex; flex-direction: column; gap: 0; }
.danger-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 14px 0;
  border-bottom: 1px solid rgba(255,255,255,0.04);
}
.danger-row:last-child { border-bottom: none; }
.danger-title { font-size: var(--font-size-base, 14px); color: var(--fg); margin-bottom: 3px; }
.danger-desc { font-size: var(--font-size-sm, 12px); color: var(--fg-dim); }

.btn-danger {
  padding: 7px 16px;
  background: var(--danger-dim, rgba(239,68,68,0.12));
  border: 1px solid rgba(239,68,68,0.4);
  color: var(--danger);
  border-radius: var(--radius);
  font-size: var(--font-size-sm, 12px);
  font-weight: 600;
  cursor: pointer;
  transition: all var(--transition);
  white-space: nowrap;
}
.btn-danger:hover { background: var(--danger); color: #fff; }

/* Buttons */
.btn-ghost {
  padding: 7px 16px;
  background: transparent;
  border: 1px solid var(--border);
  color: var(--fg-muted);
  border-radius: var(--radius);
  font-size: var(--font-size-sm, 12px);
  cursor: pointer;
  transition: all var(--transition);
}
.btn-ghost:hover { color: var(--fg); border-color: var(--fg-muted); }
.btn-primary {
  padding: 7px 18px;
  background: var(--accent);
  border: 1px solid var(--accent);
  color: #fff;
  border-radius: var(--radius);
  font-size: var(--font-size-sm, 12px);
  font-weight: 600;
  cursor: pointer;
  transition: all var(--transition);
}
.btn-primary:disabled { opacity: 0.5; cursor: not-allowed; }
.btn-primary:not(:disabled):hover { background: var(--accent-bright); }

/* Modal */
.modal-backdrop {
  position: fixed; inset: 0;
  background: rgba(0,0,0,0.6);
  display: flex; align-items: center; justify-content: center;
  z-index: 100;
}
.modal {
  background: var(--bg-panel);
  border: 1px solid var(--border);
  border-radius: var(--radius-lg, 8px);
  padding: 28px;
  width: 400px;
  max-width: 90vw;
}
.modal-title { font-size: var(--font-size-md, 15px); font-weight: 700; color: var(--fg); margin-bottom: 10px; }
.modal-body { font-size: var(--font-size-sm, 12px); color: var(--fg-muted); margin-bottom: 24px; }
.modal-actions { display: flex; gap: 10px; justify-content: flex-end; }
</style>
