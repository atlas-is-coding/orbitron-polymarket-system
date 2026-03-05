<template>
  <div class="panel-overlay" @click.self="$emit('close')">
    <div class="detail-panel">
      <button class="close-btn" @click="$emit('close')">✕</button>
      <h2 class="panel-title">{{ market.question }}</h2>

      <div class="panel-meta">
        <span v-for="tag in market.tags" :key="tag.slug ?? tag.id" class="tag-chip">
          {{ tag.label }}
        </span>
        <span v-if="market.endDateIso" class="meta-item">⏰ {{ formatDate(market.endDateIso) }}</span>
      </div>

      <div class="probs-row">
        <div class="prob-block yes">
          <span class="prob-lbl">YES</span>
          <span class="prob-big">{{ yesPct }}¢</span>
        </div>
        <div class="prob-block no">
          <span class="prob-lbl">NO</span>
          <span class="prob-big">{{ noPct }}¢</span>
        </div>
      </div>

      <div class="stats-row">
        <div class="stat">
          <span class="stat-lbl">{{ $t('markets.volume') }}</span>
          <span class="stat-val">{{ fmtVol(market.volume) }}</span>
        </div>
        <div class="stat">
          <span class="stat-lbl">{{ $t('markets.liquidity') }}</span>
          <span class="stat-val">{{ fmtVol(market.liquidity) }}</span>
        </div>
      </div>

      <div class="panel-actions">
        <button class="btn-yes" @click="$emit('buy', { market, side: 'YES' })">
          💚 Buy YES
        </button>
        <button class="btn-no" @click="$emit('buy', { market, side: 'NO' })">
          ❤️ Buy NO
        </button>
        <button class="btn-alert" @click="$emit('alert', market)">
          🔔 {{ $t('markets.set_alert') }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({ market: { type: Object, required: true } })
defineEmits(['close', 'buy', 'alert'])

const yesProb = computed(() => {
  const prices = props.market.outcomePrices
  if (!prices || prices.length === 0) return 0.5
  const p = parseFloat(Array.isArray(prices) ? prices[0] : prices)
  return isNaN(p) ? 0.5 : p
})
const yesPct = computed(() => (yesProb.value * 100).toFixed(1))
const noPct  = computed(() => ((1 - yesProb.value) * 100).toFixed(1))

function formatDate(iso) {
  try { return new Date(iso).toLocaleDateString() } catch { return iso }
}
function fmtVol(v) {
  const n = parseFloat(v || 0)
  if (isNaN(n)) return '$0'
  if (n >= 1e6) return `$${(n/1e6).toFixed(1)}M`
  if (n >= 1e3) return `$${(n/1e3).toFixed(1)}K`
  return `$${n.toFixed(0)}`
}
</script>

<style scoped>
.panel-overlay {
  position: fixed; inset: 0; background: rgba(0,0,0,0.5);
  display: flex; justify-content: flex-end; z-index: 900;
}
.detail-panel {
  width: 440px; max-width: 95vw;
  background: var(--surface-1); border-left: 1px solid var(--border);
  padding: 28px; overflow-y: auto;
  animation: slideInRight 0.2s ease;
}
.close-btn {
  background: none; border: none; color: var(--text-muted);
  font-size: 1.1rem; cursor: pointer; float: right; padding: 0;
  line-height: 1;
}
.close-btn:hover { color: var(--text); }
.panel-title { font-size: 1.2rem; font-weight: 700; margin: 8px 0 14px; line-height: 1.4; }
.panel-meta { display: flex; flex-wrap: wrap; gap: 6px; margin-bottom: 22px; }
.tag-chip {
  padding: 3px 12px; background: rgba(124,58,237,0.15);
  border-radius: 12px; font-size: 0.78rem; color: var(--accent);
}
.meta-item { font-size: 0.82rem; color: var(--text-muted); align-self: center; }
.probs-row { display: flex; gap: 16px; margin-bottom: 20px; }
.prob-block {
  flex: 1; padding: 16px; background: var(--surface-2);
  border-radius: 8px; text-align: center;
}
.prob-lbl { display: block; font-size: 0.75rem; color: var(--text-muted); margin-bottom: 6px; letter-spacing: 0.06em; }
.prob-big { font-size: 2.2rem; font-weight: 800; }
.prob-block.yes .prob-big { color: var(--success); }
.prob-block.no  .prob-big { color: var(--danger); }
.stats-row { display: flex; gap: 20px; margin-bottom: 24px; }
.stat { flex: 1; }
.stat-lbl { display: block; font-size: 0.75rem; color: var(--text-muted); margin-bottom: 4px; }
.stat-val { font-size: 1.05rem; font-weight: 600; }
.panel-actions { display: flex; gap: 10px; flex-wrap: wrap; }
.btn-yes, .btn-no, .btn-alert {
  padding: 10px 18px; border-radius: 8px; border: none;
  cursor: pointer; font-family: 'IBM Plex Mono', monospace;
  font-size: 0.9rem; font-weight: 600; transition: opacity 0.15s;
}
.btn-yes { background: var(--success); color: #fff; }
.btn-yes:hover { opacity: 0.88; }
.btn-no  { background: var(--danger); color: #fff; }
.btn-no:hover  { opacity: 0.88; }
.btn-alert { background: var(--surface-2); border: 1px solid var(--border); color: var(--text); }
.btn-alert:hover { border-color: var(--accent); }
</style>
