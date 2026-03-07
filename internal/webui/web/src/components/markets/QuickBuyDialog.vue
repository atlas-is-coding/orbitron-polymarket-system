<template>
  <div class="overlay" @click.self="$emit('close')">
    <div class="dialog">
      <div class="dialog-topbar">
        <span class="dialog-title" :class="side === 'YES' ? 'title-yes' : 'title-no'">
          QUICK BUY {{ side }}
        </span>
        <button class="close-btn" @click="$emit('close')">
          <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5">
            <line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/>
          </svg>
        </button>
      </div>

      <div class="dialog-body">
        <div class="market-q">{{ market.question }}</div>

        <!-- Price -->
        <div class="form-group">
          <div class="field-label">
            PRICE
            <span class="field-hint">market: {{ marketPrice }}</span>
          </div>
          <input
            v-model.number="form.price"
            type="number" step="0.01" min="0.01" max="0.99"
            class="field-input"
            :placeholder="String(marketPrice ?? '0.50')"
          />
        </div>

        <!-- Size -->
        <div class="form-group">
          <div class="field-label">SIZE (USD)</div>
          <input
            v-model.number="form.sizeUsd"
            type="number" step="1" min="1"
            class="field-input"
            placeholder="Min $1"
          />
        </div>

        <!-- Wallets -->
        <div class="form-group">
          <div class="field-label">WALLETS</div>
          <div class="wallet-list">
            <label
              v-for="w in enabledWallets"
              :key="w.id"
              class="wallet-row"
              :class="{ 'wallet-row--checked': selectedWallets.has(w.id) }"
            >
              <input type="checkbox" class="wallet-cb"
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
        <div v-if="form.price && form.sizeUsd" class="order-preview" :class="side === 'YES' ? 'preview-yes' : 'preview-no'">
          <span class="op-key">{{ side }}</span>
          <span class="op-val">{{ fmt2(form.sizeUsd / form.price) }} tokens</span>
          <span class="op-at">@</span>
          <span class="op-price">{{ form.price }}</span>
          <span class="op-eq">=</span>
          <span class="op-total">${{ fmt2(form.sizeUsd) }}</span>
          <span v-if="selectedWallets.size > 1" class="op-multi">
            × {{ selectedWallets.size }} wallets = ${{ fmt2(form.sizeUsd * selectedWallets.size) }} total
          </span>
        </div>
      </div>

      <div class="dialog-actions">
        <button class="btn-ghost" @click="$emit('close')">CANCEL</button>
        <button
          class="btn-execute"
          :class="side === 'YES' ? 'btn-execute--yes' : 'btn-execute--no'"
          :disabled="!canSubmit || placing"
          @click="doPlace"
        >
          <span :class="{ spin: placing }">{{ placing ? '⟳' : `EXECUTE ${side}` }}</span>
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { storeToRefs } from 'pinia'
import { useAppStore } from '@/stores/app'
import { useApi } from '@/composables/useApi'

const props = defineProps({
  market: { type: Object, required: true },
  side:   { type: String, default: 'YES' },
})
const emit = defineEmits(['close', 'placed'])

const app = useAppStore()
const api = useApi()
const { walletsMap } = storeToRefs(app)

const placing = ref(false)
const form = reactive({ price: null, sizeUsd: null })
const selectedWallets = ref(new Set())

const enabledWallets = computed(() => Object.values(walletsMap.value).filter(w => w.enabled))

const marketPrice = computed(() => {
  const prices = props.market?.outcomePrices ?? []
  const idx = props.side === 'YES' ? 0 : 1
  const p = prices[idx]
  return p != null ? parseFloat(p) : null
})

const tokenId = computed(() => {
  const ids = props.market?.clobTokenIds ?? []
  return props.side === 'YES' ? (ids[0] ?? '') : (ids[1] ?? '')
})

const canSubmit = computed(() =>
  form.price > 0 && form.price < 1 &&
  form.sizeUsd > 0 &&
  selectedWallets.value.size > 0 &&
  !!tokenId.value
)

function fmt2(n) { return (+(n || 0)).toFixed(2) }

function toggleWallet(id) {
  const next = new Set(selectedWallets.value)
  next.has(id) ? next.delete(id) : next.add(id)
  selectedWallets.value = next
}

onMounted(() => {
  if (marketPrice.value != null) form.price = marketPrice.value
  const lastSize = localStorage.getItem('qb_last_size')
  if (lastSize) form.sizeUsd = parseFloat(lastSize)
  const lastWallet = localStorage.getItem('qb_last_wallet')
  if (lastWallet && walletsMap.value[lastWallet]?.enabled) {
    selectedWallets.value = new Set([lastWallet])
  } else {
    const primary = enabledWallets.value.find(w => w.primary) ?? enabledWallets.value[0]
    if (primary) selectedWallets.value = new Set([primary.id])
  }
})

async function doPlace() {
  if (!canSubmit.value) return
  placing.value = true
  localStorage.setItem('qb_last_size', String(form.sizeUsd))
  const ids = [...selectedWallets.value]
  let lastOrderId = null
  const errors = []
  for (const walletId of ids) {
    try {
      const res = await api.placeOrder(tokenId.value, props.side, 'GTC', form.price, form.sizeUsd, walletId)
      lastOrderId = res.order_id
      localStorage.setItem('qb_last_wallet', walletId)
    } catch (e) {
      const label = walletsMap.value[walletId]?.label || walletId
      errors.push(`${label}: ${e?.response?.data?.error || e.message}`)
    }
  }
  if (errors.length) {
    app.toast(`Errors: ${errors.join('; ')}`, 'error', 8000)
  }
  if (lastOrderId) {
    app.toast(`${props.side} order placed (${ids.length - errors.length}/${ids.length})`, 'success', 5000)
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
  z-index: 400; backdrop-filter: blur(4px);
}
.dialog {
  background: var(--bg-card); border: 1px solid var(--border);
  border-radius: var(--radius); width: 420px; max-width: 96vw;
  box-shadow: var(--shadow-lg); animation: fadeSlideUp 0.18s ease both;
  overflow: hidden; display: flex; flex-direction: column;
}
.dialog-topbar {
  display: flex; align-items: center; justify-content: space-between;
  padding: 0.65rem 1rem; border-bottom: 1px solid var(--border);
}
.dialog-title { font-size: 1.00rem; font-weight: 700; letter-spacing: 0.12em; }
.title-yes { color: var(--success); }
.title-no  { color: var(--danger); }
.close-btn {
  background: none; border: none; color: var(--text-muted); cursor: pointer;
  padding: 0.2rem; display: flex; align-items: center;
  transition: color var(--transition);
}
.close-btn:hover { color: var(--danger); }

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
  color: var(--text-secondary); display: flex; align-items: baseline; gap: 0.5rem;
}
.field-hint { font-size: 1.00rem; font-weight: 400; letter-spacing: 0; color: var(--text-muted); }
.field-input {
  padding: 0.45rem 0.65rem; background: var(--bg-input); border: 1px solid var(--border);
  border-radius: var(--radius); color: var(--text-primary);
  font-family: var(--font-mono); font-size: 0.96rem; outline: none; width: 100%;
  transition: border-color var(--transition);
}
.field-input:focus { border-color: var(--accent); }

.wallet-list {
  display: flex; flex-direction: column; gap: 0.2rem;
  max-height: 140px; overflow-y: auto;
  border: 1px solid var(--border); border-radius: var(--radius); padding: 0.3rem;
}
.wallet-row {
  display: flex; align-items: center; gap: 0.5rem;
  padding: 0.3rem 0.5rem; border-radius: var(--radius-sm);
  cursor: pointer; transition: background var(--transition);
  border: 1px solid transparent;
}
.wallet-row:hover { background: var(--bg-hover); }
.wallet-row--checked { background: var(--accent-dim); border-color: rgba(0,200,255,0.22); }
.wallet-cb { accent-color: var(--accent); cursor: pointer; flex-shrink: 0; }
.wallet-name { flex: 1; font-size: 0.90rem; color: var(--text-primary); font-family: var(--font-mono); }
.wallet-bal { font-size: 0.88rem; color: var(--text-muted); font-family: var(--font-mono); }
.primary-star { color: var(--warning); margin-right: 0.2rem; }
.wallet-empty { font-size: 0.90rem; color: var(--text-muted); padding: 0.4rem; text-align: center; }

.order-preview {
  display: flex; align-items: center; gap: 0.4rem; flex-wrap: wrap;
  font-size: 0.86rem; font-family: var(--font-mono);
  padding: 0.5rem 0.75rem; border-radius: var(--radius);
  border: 1px solid var(--border-subtle);
}
.preview-yes { background: rgba(16,217,148,0.06); }
.preview-no  { background: rgba(248,113,113,0.06); }
.op-key { font-weight: 700; letter-spacing: 0.08em; color: var(--text-secondary); }
.op-val { font-weight: 600; color: var(--text-primary); }
.op-at, .op-eq { color: var(--text-muted); }
.op-price { color: var(--text-secondary); }
.op-total { font-weight: 800; color: var(--price-bright); }
.op-multi { color: var(--text-muted); font-size: 0.82rem; width: 100%; }

.dialog-actions {
  display: flex; gap: 0.5rem; padding: 0.75rem 1rem;
  border-top: 1px solid var(--border);
}
.btn-ghost {
  background: none; border: 1px solid var(--border); color: var(--text-secondary);
  border-radius: var(--radius); padding: 0.6rem 1.2rem;
  font-size: 0.86rem; font-weight: 600; cursor: pointer;
  font-family: var(--font-mono); transition: all var(--transition);
}
.btn-ghost:hover { background: var(--bg-hover); color: var(--text-primary); }
.btn-execute {
  flex: 1; border: none; border-radius: var(--radius);
  padding: 0.45rem; font-size: 0.86rem; font-weight: 700;
  cursor: pointer; font-family: var(--font-mono); letter-spacing: 0.06em;
  transition: all var(--transition);
}
.btn-execute--yes { background: var(--success); color: #000; }
.btn-execute--yes:hover:not(:disabled) { box-shadow: 0 0 16px rgba(16,217,148,0.40); }
.btn-execute--no { background: var(--danger); color: #fff; }
.btn-execute--no:hover:not(:disabled) { box-shadow: 0 0 16px rgba(248,113,113,0.35); }
.btn-execute:disabled { opacity: 0.35; cursor: not-allowed; }
</style>
