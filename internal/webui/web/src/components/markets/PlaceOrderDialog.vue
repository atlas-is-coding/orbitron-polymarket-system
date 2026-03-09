<template>
  <div class="overlay" @click.self="$emit('close')">
    <div class="dialog">
      <!-- Topbar -->
      <div class="dialog-topbar">
        <span class="dialog-title">PLACE ORDER</span>
        <button class="close-btn" @click="$emit('close')">
          <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5">
            <line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/>
          </svg>
        </button>
      </div>

      <div class="dialog-body">
        <!-- Market -->
        <div class="market-q">{{ market.question }}</div>

        <!-- Side -->
        <div class="form-group">
          <div class="field-label">SIDE</div>
          <div class="btn-group">
            <button
              class="btn-side"
              :class="{ 'btn-side--yes': form.side === 'YES' }"
              @click="form.side = 'YES'"
            >YES</button>
            <button
              class="btn-side"
              :class="{ 'btn-side--no': form.side === 'NO' }"
              @click="form.side = 'NO'"
            >NO</button>
          </div>
        </div>

        <!-- Order type -->
        <div class="form-group">
          <div class="field-label">ORDER TYPE</div>
          <select class="field-input" v-model="form.orderType">
            <option value="GTC">GTC — Good Till Cancel</option>
            <option value="FOK">FOK — Fill or Kill</option>
          </select>
        </div>

        <!-- Price -->
        <div class="form-group">
          <div class="field-label">
            PRICE
            <span class="field-hint">current: {{ currentPrice }}</span>
          </div>
          <input v-model.number="form.price" type="number" step="0.01" min="0.01" max="0.99" class="field-input mono" placeholder="0.01 – 0.99" />
        </div>

        <!-- Size -->
        <div class="form-group">
          <div class="field-label">SIZE (USD)</div>
          <input v-model.number="form.sizeUsd" type="number" step="1" min="1" class="field-input mono" placeholder="Min $1" />
        </div>

        <!-- Wallets (multi-select) -->
        <div class="form-group">
          <div class="field-label">WALLETS</div>
          <div class="wallet-list">
            <label
              v-for="w in enabledWallets"
              :key="w.id"
              class="wallet-row"
              :class="{ 'wallet-row--checked': selectedWallets.has(w.id) }"
            >
              <input
                type="checkbox"
                class="wallet-cb"
                :checked="selectedWallets.has(w.id)"
                @change="toggleWallet(w.id)"
              />
              <span class="wallet-name">
                <span v-if="w.primary" class="primary-star">★</span>
                {{ w.label || w.id }}
              </span>
              <span class="wallet-bal">${{ fmt2(w.balance_usd) }}</span>
            </label>
            <div v-if="!enabledWallets.length" class="wallet-empty">No enabled wallets</div>
          </div>
        </div>

        <!-- Preview -->
        <div v-if="form.price && form.sizeUsd" class="order-preview">
          <span class="op-key">BUYING</span>
          <span class="op-val">{{ fmt2(form.sizeUsd / form.price) }} tokens</span>
          <span class="op-at">@</span>
          <span class="op-price">{{ form.price }}</span>
          <span class="op-eq">=</span>
          <span class="op-total num-glow">${{ fmt2(form.sizeUsd) }}</span>
        </div>
      </div>

      <!-- Actions -->
      <div class="dialog-actions">
        <button class="btn-ghost" @click="$emit('close')">CANCEL</button>
        <button class="btn-primary" :disabled="!canSubmit || placing" @click="doPlace">
          <span :class="{ spin: placing }">{{ placing ? '⟳' : 'EXECUTE ORDER' }}</span>
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, computed } from 'vue'
import { storeToRefs } from 'pinia'
import { useAppStore } from '@/stores/app'
import { useApi } from '@/composables/useApi'

const props = defineProps({ market: Object })
const emit = defineEmits(['close', 'placed'])

const app = useAppStore()
const api = useApi()
const { walletsMap } = storeToRefs(app)

const placing = ref(false)
const form = reactive({ side: 'YES', orderType: 'GTC', price: null, sizeUsd: null })
const selectedWallets = ref(new Set())

const enabledWallets = computed(() => Object.values(walletsMap.value).filter(w => w.enabled))

// Pre-select primary wallet (or first enabled if no primary)
const primary = computed(() => enabledWallets.value.find(w => w.primary) ?? enabledWallets.value[0])
if (primary.value) selectedWallets.value = new Set([primary.value.id])

function toggleWallet(id) {
  const next = new Set(selectedWallets.value)
  next.has(id) ? next.delete(id) : next.add(id)
  selectedWallets.value = next
}

const currentPrice = computed(() => {
  if (!props.market?.outcomePrices?.length) return '—'
  return form.side === 'YES' ? props.market.outcomePrices[0] : (props.market.outcomePrices[1] || '—')
})

const tokenId = computed(() => {
  if (!props.market?.clobTokenIds?.length) return ''
  return form.side === 'YES' ? props.market.clobTokenIds[0] : (props.market.clobTokenIds[1] || '')
})

const canSubmit = computed(() =>
  form.price > 0 && form.price < 1 && form.sizeUsd > 0 && selectedWallets.value.size > 0 && tokenId.value
)

function fmt2(n) { return (+(n || 0)).toFixed(2) }

async function doPlace() {
  if (!canSubmit.value) return
  placing.value = true
  const ids = [...selectedWallets.value]
  let lastOrderId = null
  let errors = []
  for (const walletId of ids) {
    try {
      const res = await api.placeOrder(tokenId.value, form.side, form.orderType, form.price, form.sizeUsd, walletId)
      lastOrderId = res.order_id
    } catch (e) {
      const label = walletsMap.value[walletId]?.label || walletId
      errors.push(`${label}: ${e?.response?.data?.error || e.message}`)
    }
  }
  if (errors.length) {
    app.toast(`Order errors: ${errors.join('; ')}`, 'error', 8000)
  }
  if (lastOrderId) {
    app.toast(`Order(s) placed (${ids.length - errors.length}/${ids.length})`, 'success', 6000)
    emit('placed', lastOrderId)
    emit('close')
  }
  placing.value = false
}
</script>

<style scoped>
.overlay {
  position: fixed; inset: 0;
  background: var(--bg-overlay); display: flex; align-items: center; justify-content: center;
  z-index: 300; backdrop-filter: blur(4px);
}

.dialog {
  background: var(--bg-card);
  border: 1px solid var(--border); border-top: 2px solid var(--accent);
  border-radius: var(--radius);
  width: 460px; max-width: 96vw;
  box-shadow: var(--shadow-lg), var(--shadow-cyan);
  animation: fadeSlideUp 0.18s ease both;
  overflow: hidden; display: flex; flex-direction: column;
}

/* Topbar */
.dialog-topbar {
  display: flex; align-items: center; justify-content: space-between;
  padding: 0.65rem 1rem;
  background: rgba(0, 200, 255, 0.04);
  border-bottom: 1px solid var(--border);
}
.dialog-title { font-size: 1.00rem; font-weight: 700; letter-spacing: 0.12em; color: var(--accent); }
.close-btn {
  background: none; border: none; color: var(--text-muted); cursor: pointer;
  padding: 0.2rem; display: flex; align-items: center; border-radius: var(--radius-sm);
  transition: color var(--transition);
}
.close-btn:hover { color: var(--danger); }

/* Body */
.dialog-body { padding: 1rem; display: flex; flex-direction: column; gap: 0.65rem; }

.market-q {
  font-size: 0.90rem; color: var(--text-secondary); line-height: 1.5;
  white-space: nowrap; overflow: hidden; text-overflow: ellipsis;
  font-family: var(--font-mono);
  background: var(--bg-hover); padding: 0.4rem 0.65rem;
  border-radius: var(--radius); border: 1px solid var(--border-subtle);
}

.form-group { display: flex; flex-direction: column; gap: 0.3rem; }

.field-label {
  font-size: 0.86rem; font-weight: 700; letter-spacing: 0.10em;
  color: var(--text-secondary); text-transform: uppercase;
  display: flex; align-items: baseline; gap: 0.5rem;
}
.field-hint { font-size: 1.00rem; font-weight: 400; letter-spacing: 0; text-transform: none; color: var(--text-muted); }

.field-input {
  padding: 0.45rem 0.65rem;
  background: var(--bg-input); border: 1px solid var(--border);
  border-radius: var(--radius); color: var(--text-primary);
  font-family: var(--font-mono); font-size: 0.96rem; outline: none;
  transition: border-color var(--transition); width: 100%;
}
.field-input:focus { border-color: var(--accent); box-shadow: 0 0 0 1px rgba(0,200,255,0.12); }
.field-input::placeholder { color: var(--text-muted); }
select.field-input { cursor: pointer; }

/* Wallet multi-select */
.wallet-list {
  display: flex; flex-direction: column; gap: 0.25rem;
  max-height: 160px; overflow-y: auto;
  border: 1px solid var(--border); border-radius: var(--radius);
  padding: 0.35rem;
}
.wallet-row {
  display: flex; align-items: center; gap: 0.5rem;
  padding: 0.35rem 0.5rem; border-radius: var(--radius-sm);
  cursor: pointer; transition: background var(--transition);
  border: 1px solid transparent;
}
.wallet-row:hover { background: var(--bg-hover); }
.wallet-row--checked { background: var(--accent-dim); border-color: rgba(0,200,255,0.22); }
.wallet-cb { accent-color: var(--accent); cursor: pointer; flex-shrink: 0; }
.wallet-name { flex: 1; font-size: 0.92rem; color: var(--text-primary); font-family: var(--font-mono); }
.wallet-bal { font-size: 0.90rem; color: var(--text-muted); font-family: var(--font-mono); flex-shrink: 0; }
.primary-star { color: var(--warning); margin-right: 0.25rem; }
.wallet-empty { font-size: 0.90rem; color: var(--text-muted); padding: 0.5rem; text-align: center; }

/* Side buttons */
.btn-group { display: flex; gap: 0.4rem; }
.btn-side {
  flex: 1; padding: 0.4rem 0; border: 1px solid var(--border);
  border-radius: var(--radius); background: var(--bg-hover);
  color: var(--text-secondary); font-size: 0.92rem; font-weight: 700;
  cursor: pointer; font-family: var(--font-mono); letter-spacing: 0.08em;
  transition: all var(--transition);
}
.btn-side--yes { background: rgba(16,217,148,0.10); border-color: var(--success); color: var(--success); }
.btn-side--no  { background: rgba(255,77,106,0.10); border-color: var(--danger);  color: var(--danger); }

/* Order preview */
.order-preview {
  display: flex; align-items: center; gap: 0.4rem; flex-wrap: wrap;
  font-size: 0.86rem; font-family: var(--font-mono);
  background: var(--bg-hover); padding: 0.5rem 0.75rem;
  border-radius: var(--radius); border: 1px solid var(--border-subtle);
}
.op-key { font-size: 0.86rem; font-weight: 700; letter-spacing: 0.10em; color: var(--text-secondary); }
.op-val { font-weight: 600; color: var(--text-primary); }
.op-at, .op-eq { color: var(--text-muted); }
.op-price { color: var(--text-secondary); }
.op-total { font-weight: 800; }
.num-glow { color: var(--price-bright); text-shadow: 0 0 10px rgba(251,191,36,0.30); }

/* Actions */
.dialog-actions {
  display: flex; gap: 0.5rem;
  padding: 0.75rem 1rem;
  border-top: 1px solid var(--border);
  background: rgba(0, 200, 255, 0.02);
}
.btn-ghost {
  background: none; border: 1px solid var(--border); color: var(--text-secondary);
  border-radius: var(--radius); padding: 0.6rem 1.2rem; font-size: 0.86rem; font-weight: 600;
  cursor: pointer; font-family: var(--font-mono); letter-spacing: 0.04em; transition: all var(--transition);
}
.btn-ghost:hover { background: var(--bg-hover); color: var(--text-primary); }
.btn-primary {
  flex: 1; background: var(--accent); border: none;
  border-radius: var(--radius); color: #000;
  padding: 0.45rem; font-size: 0.86rem; font-weight: 700;
  cursor: pointer; font-family: var(--font-mono); letter-spacing: 0.06em;
  transition: all var(--transition);
}
.btn-primary:hover:not(:disabled) { background: var(--accent-hover); box-shadow: var(--accent-glow); }
.btn-primary:disabled { opacity: 0.35; cursor: not-allowed; }
</style>
