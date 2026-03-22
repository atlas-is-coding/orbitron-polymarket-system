<template>
  <Transition name="batch-bar">
    <div v-if="selectedMarkets.length > 0" class="batch-bar">
      <span class="batch-count">{{ selectedMarkets.length }} selected</span>
      <div class="batch-side">
        <button :class="['side-btn', { active: side === 'YES' }]" @click="side = 'YES'">YES</button>
        <button :class="['side-btn', { active: side === 'NO' }]" @click="side = 'NO'">NO</button>
      </div>
      <input
        v-model.number="sizeUsd"
        type="number"
        class="field-input batch-size"
        placeholder="Size USD"
        min="1"
      />
      <button class="btn btn-primary" :disabled="!sizeUsd || placing" @click="placeOrders">
        {{ placing ? '...' : 'PLACE BATCH ORDER' }}
      </button>
      <button class="btn btn-ghost" @click="$emit('clear')">CLEAR</button>
    </div>
  </Transition>
</template>

<script setup>
import { ref } from 'vue'
import { useMarketsStore } from '../../stores/markets.js'
import { useAppStore } from '../../stores/app.js'

const props = defineProps({ selectedMarkets: { type: Array, default: () => [] } })
const emit = defineEmits(['clear'])

const store = useMarketsStore()
const app = useAppStore()

const side = ref('YES')
const sizeUsd = ref('')
const placing = ref(false)

async function placeOrders() {
  if (!sizeUsd.value || placing.value) return
  placing.value = true
  try {
    await store.placeBatchOrder(props.selectedMarkets, side.value, sizeUsd.value)
    app.toast(`Batch order placed for ${props.selectedMarkets.length} markets`, 'success')
    emit('clear')
  } catch (e) {
    app.toast(e?.response?.data?.error || 'Batch order failed', 'error')
  } finally {
    placing.value = false
  }
}
</script>

<style scoped>
.batch-bar {
  position: fixed;
  bottom: 20px;
  left: 50%;
  transform: translateX(-50%);
  display: flex;
  align-items: center;
  gap: 10px;
  background: var(--bg-card);
  border: 1px solid var(--accent);
  border-radius: 6px;
  padding: 10px 16px;
  box-shadow: var(--shadow-lg);
  z-index: 100;
  white-space: nowrap;
}
.batch-count { font-size: var(--font-size-sm, 12px); color: var(--fg-muted); }
.batch-side { display: flex; gap: 4px; }
.side-btn {
  padding: 3px 10px; border-radius: 3px; border: 1px solid var(--border);
  background: transparent; color: var(--fg-muted); cursor: pointer; font-size: 12px;
  font-family: var(--font-mono);
}
.side-btn.active { background: var(--accent); color: #fff; border-color: var(--accent); }
.batch-size { width: 90px; }

.batch-bar-enter-active, .batch-bar-leave-active { transition: all 0.2s ease; }
.batch-bar-enter-from, .batch-bar-leave-to { opacity: 0; transform: translateX(-50%) translateY(20px); }
</style>
