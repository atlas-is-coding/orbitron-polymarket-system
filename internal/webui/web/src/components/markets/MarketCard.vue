<template>
  <div class="market-card" @click="$emit('select', market)">
    <div class="market-question">{{ market.question }}</div>
    <div class="market-probs">
      <div class="prob yes">
        <span class="prob-label">{{ $t('markets.yes') }}</span>
        <span class="prob-value">{{ yesPct }}</span>
        <div class="prob-bar">
          <div class="prob-fill yes-fill" :style="{ width: yesPct }"></div>
        </div>
      </div>
      <div class="prob no">
        <span class="prob-label">{{ $t('markets.no') }}</span>
        <span class="prob-value">{{ noPct }}</span>
        <div class="prob-bar">
          <div class="prob-fill no-fill" :style="{ width: noPct }"></div>
        </div>
      </div>
    </div>
    <div class="market-meta">
      <span>{{ $t('markets.volume') }} {{ formatVolume(market.volume) }}</span>
      <span>{{ $t('markets.liquidity') }} {{ formatVolume(market.liquidity) }}</span>
      <span v-if="market.endDateIso">{{ $t('markets.ends') }} {{ formatDate(market.endDateIso) }}</span>
    </div>
    <div class="market-actions" @click.stop>
      <button class="btn-buy" @click="$emit('buy', market)">{{ $t('markets.buy') }}</button>
      <button class="btn-alert" @click="$emit('alert', market)">🔔</button>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({ market: { type: Object, required: true } })
defineEmits(['select', 'buy', 'alert'])

const yesProb = computed(() => {
  const prices = props.market.outcomePrices
  if (!prices || prices.length === 0) return 0.5
  const p = parseFloat(Array.isArray(prices) ? prices[0] : prices)
  return isNaN(p) ? 0.5 : p
})

const yesPct = computed(() => (yesProb.value * 100).toFixed(1) + '%')
const noPct  = computed(() => ((1 - yesProb.value) * 100).toFixed(1) + '%')

function formatVolume(v) {
  if (!v && v !== 0) return '$0'
  const n = parseFloat(v)
  if (isNaN(n)) return '$0'
  if (n >= 1e6) return `$${(n/1e6).toFixed(1)}M`
  if (n >= 1e3) return `$${(n/1e3).toFixed(1)}K`
  return `$${n.toFixed(0)}`
}

function formatDate(iso) {
  if (!iso) return ''
  try { return new Date(iso).toLocaleDateString() } catch { return iso }
}
</script>

<style scoped>
.market-card {
  background: var(--surface-2);
  border: 1px solid var(--border);
  border-radius: 10px;
  padding: 22px;
  cursor: pointer;
  transition: border-color 0.15s, transform 0.1s;
}
.market-card:hover { border-color: var(--accent); transform: translateY(-1px); }
.market-question {
  font-size: 1.1rem;
  font-weight: 600;
  color: var(--text);
  margin-bottom: 16px;
  line-height: 1.4;
}
.market-probs { display: flex; gap: 20px; margin-bottom: 16px; }
.prob { flex: 1; }
.prob-label { font-size: 0.75rem; color: var(--text-muted); letter-spacing: 0.06em; text-transform: uppercase; }
.prob-value { font-size: 1.6rem; font-weight: 800; display: block; margin: 4px 0 8px; }
.prob.yes .prob-value { color: var(--success); }
.prob.no  .prob-value { color: var(--danger); }
.prob-bar { height: 5px; background: var(--surface-3); border-radius: 3px; overflow: hidden; }
.prob-fill { height: 100%; border-radius: 3px; }
.yes-fill { background: var(--success); }
.no-fill  { background: var(--danger); }
.market-meta { display: flex; gap: 20px; font-size: 0.8rem; color: var(--text-muted); margin-bottom: 16px; }
.market-actions { display: flex; gap: 10px; }
.btn-buy {
  padding: 8px 20px;
  background: var(--accent);
  border: none;
  border-radius: 7px;
  color: #fff;
  font-size: 0.9rem;
  font-family: 'IBM Plex Mono', monospace;
  font-weight: 600;
  cursor: pointer;
  transition: background 0.15s;
}
.btn-buy:hover { background: var(--accent-hover, #6d28d9); }
.btn-alert {
  padding: 8px 14px;
  background: transparent;
  border: 1px solid var(--border);
  border-radius: 7px;
  cursor: pointer;
  font-size: 1rem;
  transition: border-color 0.15s;
}
.btn-alert:hover { border-color: var(--warning, #fbbf24); }
</style>
