<template>
  <div class="overlay" @click.self="$emit('close')">
    <div class="dialog">
      <div class="dialog-title">CLOSE POSITION</div>
      <div class="dialog-body">
        <div class="field-group">
          <label class="field-label">Market</label>
          <div class="field-val">{{ position?.market || '—' }}</div>
        </div>
        <div class="field-group">
          <label class="field-label">Current Price</label>
          <div class="field-val price-val">${{ fmt(position?.current_price) }}</div>
        </div>
        <div class="field-group">
          <label class="field-label">Quantity</label>
          <input v-model.number="quantity" type="number" class="field-input" :max="position?.shares" min="0.01" step="0.01" />
        </div>
        <div class="field-group">
          <label class="field-label">Estimated Return</label>
          <div class="field-val" :class="estReturn >= 0 ? 'pos-val' : 'neg-val'">${{ fmt(estReturn) }}</div>
        </div>
      </div>
      <div class="dialog-actions">
        <button class="btn btn-ghost" @click="$emit('close')">Cancel</button>
        <button class="btn btn-primary" :disabled="closing || !quantity" @click="confirm">
          {{ closing ? '...' : 'CLOSE POSITION' }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { useApi } from '@/composables/useApi.js'
import { useAppStore } from '@/stores/app.js'

const props = defineProps({ position: Object })
const emit = defineEmits(['close', 'closed'])
const api = useApi()
const app = useAppStore()
const quantity = ref(props.position?.shares || 0)
const closing = ref(false)
const estReturn = computed(() => (quantity.value || 0) * (props.position?.current_price || 0))
function fmt(n) { return n != null ? Number(n || 0).toFixed(2) : '—' }
async function confirm() {
  if (!quantity.value || closing.value) return
  closing.value = true
  try {
    await api.closePosition(props.position.id, quantity.value)
    app.toast('Position closed', 'success')
    emit('closed', props.position.id)
    emit('close')
  } catch (e) {
    app.toast(e?.response?.data?.error || 'Failed to close position', 'error')
  } finally {
    closing.value = false
  }
}
</script>

<style scoped>
.overlay { position: fixed; inset: 0; background: var(--bg-overlay); display: flex; align-items: center; justify-content: center; z-index: 200; backdrop-filter: blur(4px); }
.dialog { background: var(--bg-card); border: 1px solid var(--border); border-top: 2px solid var(--accent); border-radius: var(--radius); padding: 1.5rem; min-width: 320px; box-shadow: var(--shadow-lg); }
.dialog-title { font-size: 11px; font-weight: 700; color: var(--accent); text-transform: uppercase; letter-spacing: 0.10em; margin-bottom: 1.25rem; }
.dialog-body { display: flex; flex-direction: column; gap: 12px; margin-bottom: 1.25rem; }
.field-group { display: flex; flex-direction: column; gap: 4px; }
.field-label { font-size: 10px; text-transform: uppercase; letter-spacing: 0.08em; color: var(--text-secondary); }
.field-val { font-size: 14px; color: var(--text-primary); font-family: var(--font-mono); }
.price-val { color: var(--price-bright); }
.pos-val { color: var(--success); }
.neg-val { color: var(--danger); }
.dialog-actions { display: flex; gap: 8px; justify-content: flex-end; }
.btn { display: inline-flex; align-items: center; padding: 6px 14px; border-radius: var(--radius); font-family: var(--font-mono); font-size: 13px; font-weight: 500; cursor: pointer; border: 1px solid transparent; transition: all var(--transition); }
.btn-ghost { background: none; border-color: var(--border); color: var(--text-secondary); }
.btn-primary { background: var(--accent); color: #fff; border-color: var(--accent); }
.btn:disabled { opacity: 0.4; cursor: not-allowed; }
</style>
