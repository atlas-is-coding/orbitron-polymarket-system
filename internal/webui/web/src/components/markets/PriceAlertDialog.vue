<template>
  <div class="dialog-overlay" @click.self="$emit('close')">
    <div class="dialog">
      <h3>{{ $t('markets.alert_title') }}</h3>
      <p class="market-q">{{ market.question }}</p>
      <div class="direction-row">
        <button :class="['dir-btn', { active: direction === 'above' }]" @click="direction = 'above'">
          📈 {{ $t('markets.alert_above') }}
        </button>
        <button :class="['dir-btn', { active: direction === 'below' }]" @click="direction = 'below'">
          📉 {{ $t('markets.alert_below') }}
        </button>
      </div>
      <label class="input-label">{{ $t('markets.alert_threshold') }}</label>
      <input v-model.number="threshold" type="number" min="0.01" max="0.99" step="0.01" class="input" />
      <div class="dialog-actions">
        <button class="btn-primary" @click="submit">{{ $t('markets.confirm') }}</button>
        <button class="btn-secondary" @click="$emit('close')">{{ $t('markets.cancel') }}</button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'

const props = defineProps({ market: { type: Object, required: true } })
const emit = defineEmits(['close', 'created'])

const direction = ref('above')
const threshold = ref(0.80)

function submit() {
  const tokenId = props.market.clobTokenIds?.[0] ?? ''
  emit('created', {
    conditionId: props.market.conditionId,
    tokenId,
    direction: direction.value,
    threshold: parseFloat(threshold.value),
  })
  emit('close')
}
</script>

<style scoped>
.dialog-overlay {
  position: fixed; inset: 0; background: rgba(0,0,0,0.65);
  display: flex; align-items: center; justify-content: center; z-index: 1000;
}
.dialog {
  background: var(--surface-1); border: 1px solid var(--border);
  border-radius: 12px; padding: 28px; width: 400px; max-width: 95vw;
}
h3 { margin: 0 0 10px; font-size: 1.05rem; }
.market-q { font-size: 0.88rem; color: var(--text-muted); margin-bottom: 18px; line-height: 1.4; }
.direction-row { display: flex; gap: 10px; margin-bottom: 18px; }
.dir-btn {
  flex: 1; padding: 9px; border: 1px solid var(--border); border-radius: 7px;
  background: transparent; cursor: pointer; font-size: 0.85rem;
  color: var(--text-muted); font-family: 'IBM Plex Mono', monospace;
  transition: all 0.15s;
}
.dir-btn.active { border-color: var(--accent); color: var(--accent); background: rgba(124,58,237,0.12); }
.input-label { font-size: 0.8rem; color: var(--text-muted); display: block; margin-bottom: 8px; }
.input {
  width: 100%; padding: 9px 14px; background: var(--surface-2);
  border: 1px solid var(--border); border-radius: 7px; color: var(--text);
  font-family: 'IBM Plex Mono', monospace; font-size: 0.95rem;
  margin-bottom: 20px; box-sizing: border-box;
}
.input:focus { outline: none; border-color: var(--accent); }
.dialog-actions { display: flex; gap: 10px; }
.btn-primary {
  flex: 1; padding: 10px; background: var(--accent); border: none;
  border-radius: 7px; color: #fff; cursor: pointer;
  font-family: 'IBM Plex Mono', monospace; font-weight: 600;
}
.btn-primary:hover { background: var(--accent-hover, #6d28d9); }
.btn-secondary {
  flex: 1; padding: 10px; background: transparent;
  border: 1px solid var(--border); border-radius: 7px;
  color: var(--text-muted); cursor: pointer; font-family: 'IBM Plex Mono', monospace;
}
</style>
