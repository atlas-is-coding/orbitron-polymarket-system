<template>
  <Transition name="bar">
    <div v-if="count > 0" class="action-bar">
      <button class="bar-clear" @click="$emit('clear')" title="Clear selection">
        <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5">
          <line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/>
        </svg>
      </button>

      <span class="bar-count">{{ count }} selected</span>

      <div class="bar-divider" />

      <button
        class="bar-side"
        :class="{ 'bar-side--yes': side === 'YES' }"
        @click="$emit('update:side', 'YES')"
      >YES</button>
      <button
        class="bar-side"
        :class="{ 'bar-side--no': side === 'NO' }"
        @click="$emit('update:side', 'NO')"
      >NO</button>

      <div class="bar-divider" />

      <div class="bar-size-wrap">
        <span class="bar-label">$</span>
        <input
          v-model.number="localSize"
          type="number" min="1" step="1"
          class="bar-size-input"
          placeholder="Size"
        />
      </div>

      <div class="bar-divider" />

      <select class="bar-wallet" v-model="localWallet">
        <option value="" disabled>Wallet</option>
        <option v-for="w in enabledWallets" :key="w.id" :value="w.id">
          {{ w.primary ? '★ ' : '' }}{{ w.label || w.id }}
        </option>
      </select>

      <button
        class="bar-execute"
        :disabled="!canExecute || executing"
        @click="doExecute"
      >
        <span v-if="!executing">EXECUTE ALL ({{ count }})</span>
        <span v-else>{{ progress }}/{{ count }}</span>
      </button>
    </div>
  </Transition>
</template>

<script setup>
import { ref, computed, watch } from 'vue'
import { storeToRefs } from 'pinia'
import { useAppStore } from '@/stores/app'
import { useApi } from '@/composables/useApi'

const props = defineProps({
  count:   { type: Number, default: 0 },
  side:    { type: String, default: 'YES' },
  markets: { type: Array, default: () => [] },
})
const emit = defineEmits(['clear', 'update:side'])

const app = useAppStore()
const api = useApi()
const { walletsMap } = storeToRefs(app)

const localSize   = ref(parseFloat(localStorage.getItem('qb_last_size') || '10'))
const localWallet = ref(localStorage.getItem('qb_last_wallet') || '')
const executing   = ref(false)
const progress    = ref(0)

const enabledWallets = computed(() => Object.values(walletsMap.value).filter(w => w.enabled))

watch(enabledWallets, (ws) => {
  if (!localWallet.value || !walletsMap.value[localWallet.value]?.enabled) {
    const primary = ws.find(w => w.primary) ?? ws[0]
    if (primary) localWallet.value = primary.id
  }
}, { immediate: true })

const canExecute = computed(() =>
  localSize.value > 0 && localWallet.value && props.markets.length > 0
)

async function doExecute() {
  if (!canExecute.value) return
  executing.value = true
  progress.value = 0
  localStorage.setItem('qb_last_size', String(localSize.value))
  localStorage.setItem('qb_last_wallet', localWallet.value)

  const errors = []
  for (const market of props.markets) {
    const prices = market.outcomePrices ?? []
    const ids    = market.clobTokenIds  ?? []
    const idx    = props.side === 'YES' ? 0 : 1
    const tokenId = ids[idx]
    const price   = parseFloat(prices[idx] ?? 0)

    if (!tokenId || price <= 0) {
      errors.push(`${(market.question ?? '').slice(0, 30)}: no token/price`)
      progress.value++
      continue
    }
    try {
      await api.placeOrder(tokenId, props.side, 'GTC', price, localSize.value, localWallet.value)
    } catch (e) {
      errors.push(`${(market.question ?? '').slice(0, 30)}: ${e?.response?.data?.error || e.message}`)
    }
    progress.value++
  }

  executing.value = false
  const placed = props.markets.length - errors.length
  if (errors.length) {
    app.toast(
      `${placed}/${props.markets.length} placed. Errors: ${errors.slice(0, 2).join('; ')}${errors.length > 2 ? '…' : ''}`,
      'error', 10000
    )
  } else {
    app.toast(`${placed}/${props.markets.length} ${props.side} orders placed`, 'success', 5000)
  }
  emit('clear')
}
</script>

<style scoped>
.action-bar {
  position: fixed; bottom: 1.5rem; left: 50%; transform: translateX(-50%);
  display: flex; align-items: center; gap: 0.65rem;
  background: var(--bg-secondary);
  border: 1px solid var(--accent); border-radius: var(--radius);
  padding: 0.6rem 1rem;
  box-shadow: var(--shadow-lg), 0 0 24px rgba(0,200,255,0.15);
  z-index: 200;
  white-space: nowrap;
  min-width: 560px;
}

.bar-clear {
  background: none; border: none; color: var(--text-muted); cursor: pointer;
  padding: 0.2rem; display: flex; align-items: center;
  transition: color var(--transition);
}
.bar-clear:hover { color: var(--danger); }

.bar-count {
  font-size: 0.90rem; font-weight: 700; letter-spacing: 0.06em;
  color: var(--text-bright); font-family: var(--font-mono);
}

.bar-divider { width: 1px; height: 20px; background: var(--border); flex-shrink: 0; }

.bar-side {
  padding: 0.3rem 0.75rem; border: 1px solid var(--border);
  border-radius: var(--radius); background: transparent;
  color: var(--text-muted); font-size: 0.86rem; font-weight: 700;
  font-family: var(--font-mono); cursor: pointer; letter-spacing: 0.08em;
  transition: all var(--transition);
}
.bar-side--yes { background: rgba(16,217,148,0.10); border-color: var(--success); color: var(--success); }
.bar-side--no  { background: rgba(248,113,113,0.10); border-color: var(--danger);  color: var(--danger); }

.bar-size-wrap { display: flex; align-items: center; gap: 0.25rem; }
.bar-label { font-size: 0.90rem; color: var(--text-muted); font-family: var(--font-mono); }
.bar-size-input {
  width: 70px; padding: 0.3rem 0.5rem;
  background: var(--bg-input); border: 1px solid var(--border);
  border-radius: var(--radius); color: var(--text-primary);
  font-family: var(--font-mono); font-size: 0.92rem; outline: none;
  transition: border-color var(--transition);
}
.bar-size-input:focus { border-color: var(--accent); }

.bar-wallet {
  padding: 0.3rem 0.5rem;
  background: var(--bg-input); border: 1px solid var(--border);
  border-radius: var(--radius); color: var(--text-primary);
  font-family: var(--font-mono); font-size: 0.90rem; cursor: pointer; outline: none;
  max-width: 130px;
}

.bar-execute {
  padding: 0.35rem 1rem;
  background: var(--accent); border: none;
  border-radius: var(--radius); color: #000;
  font-size: 0.86rem; font-weight: 700; font-family: var(--font-mono);
  letter-spacing: 0.06em; cursor: pointer;
  transition: all var(--transition);
}
.bar-execute:hover:not(:disabled) { box-shadow: var(--accent-glow); }
.bar-execute:disabled { opacity: 0.40; cursor: not-allowed; }

.bar-enter-active, .bar-leave-active { transition: transform 0.22s ease, opacity 0.22s ease; }
.bar-enter-from, .bar-leave-to { transform: translateX(-50%) translateY(20px); opacity: 0; }
</style>
