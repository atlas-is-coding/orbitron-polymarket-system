<template>
  <div class="market-card" :class="{ 'is-categorical': isCategorical, 'is-selected': selected }" @click="$emit('select', market)">

    <!-- Checkbox for multi-select -->
    <div class="card-cb-wrap" @click.stop="$emit('toggleSelect', market)">
      <input
        type="checkbox"
        class="card-cb"
        :checked="selected"
        @click.stop
        @change="$emit('toggleSelect', market)"
      />
    </div>

    <!-- Header -->
    <div class="card-header">
      <span v-if="market.category" class="tag-badge">{{ market.category }}</span>
      <span v-if="market.negRisk" class="risk-badge">NEG RISK</span>
      <span v-if="isCategorical" class="multi-badge">{{ outcomes.length }}OUT</span>
      <span class="hdr-spacer" />
      <span v-if="endLabel" class="end-label" :class="endLabel === 'Ended' ? 'ended' : ''">{{ endLabel }}</span>
    </div>

    <!-- Question -->
    <div class="market-question">{{ market.question }}</div>

    <!-- Binary YES / NO -->
    <template v-if="!isCategorical">
      <div class="binary-probs">
        <div class="binary-side yes-side">
          <span class="side-tag">YES</span>
          <span class="side-pct yes-pct">{{ fmtPct(outcomes[0]?.prob) }}</span>
          <div class="prob-bar">
            <div class="prob-fill yes-fill" :style="{ width: fmtPct(outcomes[0]?.prob) }" />
          </div>
        </div>
        <div class="binary-divider" />
        <div class="binary-side no-side">
          <span class="side-tag">NO</span>
          <span class="side-pct no-pct">{{ fmtPct(outcomes[1]?.prob) }}</span>
          <div class="prob-bar">
            <div class="prob-fill no-fill" :style="{ width: fmtPct(outcomes[1]?.prob) }" />
          </div>
        </div>
      </div>
    </template>

    <!-- Categorical multi-outcome -->
    <template v-else>
      <div class="outcome-list">
        <div v-for="(o, i) in visibleOutcomes" :key="i" class="outcome-row">
          <span class="outcome-rank">{{ String(i+1).padStart(2,'0') }}</span>
          <span class="outcome-label" :title="o.label">{{ o.label }}</span>
          <div class="outcome-bar-wrap">
            <div class="outcome-bar" :style="{ width: fmtPct(o.prob) }" />
          </div>
          <span class="outcome-pct">{{ fmtPct(o.prob) }}</span>
        </div>
        <div v-if="outcomes.length > maxVisible" class="more-label">+{{ outcomes.length - maxVisible }} more</div>
      </div>
    </template>

    <!-- Meta row -->
    <div class="meta-row">
      <div class="meta-item">
        <span class="meta-key">VOL</span>
        <span class="meta-val">{{ fmtMoney(market.volume) }}</span>
      </div>
      <div class="meta-dot" />
      <div class="meta-item">
        <span class="meta-key">LIQ</span>
        <span class="meta-val">{{ fmtMoney(market.liquidity) }}</span>
      </div>
    </div>

    <!-- Actions: YES / NO quick buy + alert -->
    <div class="card-actions" @click.stop>
      <button
        v-if="!isCategorical && yesPrice !== null"
        class="btn-yes"
        @click="$emit('quickbuy', { market, side: 'YES' })"
      >YES {{ fmtPct(outcomes[0]?.prob) }}</button>
      <button
        v-if="!isCategorical && noPrice !== null"
        class="btn-no"
        @click="$emit('quickbuy', { market, side: 'NO' })"
      >NO {{ fmtPct(outcomes[1]?.prob) }}</button>
      <button v-if="isCategorical" class="btn-trade" @click="$emit('buy', market)">TRADE</button>
      <button class="btn-alert" @click="$emit('alert', market)" title="Set price alert">
        <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M18 8A6 6 0 0 0 6 8c0 7-3 9-3 9h18s-3-2-3-9"/>
          <path d="M13.73 21a2 2 0 0 1-3.46 0"/>
        </svg>
      </button>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  market: { type: Object, required: true },
  selected: { type: Boolean, default: false },
})
defineEmits(['select', 'buy', 'quickbuy', 'alert', 'toggleSelect'])

const maxVisible = 4

const outcomes = computed(() => {
  const prices = props.market.outcomePrices ?? []
  const labels = props.market.outcomes ?? []
  const priceArr = Array.isArray(prices) ? prices : [prices]
  return priceArr.map((p, i) => ({
    label: labels[i] ?? (i === 0 ? 'YES' : i === 1 ? 'NO' : `Option ${i + 1}`),
    prob: parseFloat(p) || 0,
  }))
})

const yesPrice = computed(() => outcomes.value[0]?.prob ?? null)
const noPrice  = computed(() => outcomes.value[1]?.prob ?? null)

const isCategorical = computed(() => outcomes.value.length > 2)
const visibleOutcomes = computed(() => outcomes.value.slice(0, maxVisible))

const endLabel = computed(() => {
  const iso = props.market.endDateIso
  if (!iso) return ''
  try {
    const d = new Date(iso)
    const diff = d - new Date()
    if (diff < 0) return 'Ended'
    const days = Math.ceil(diff / 86400000)
    if (days <= 3) return `${days}d left`
    return d.toLocaleDateString('en', { month: 'short', day: 'numeric' })
  } catch { return '' }
})

function fmtPct(p) {
  if (p == null || isNaN(p)) return '—'
  return (p * 100).toFixed(1) + '%'
}

function fmtMoney(v) {
  const n = parseFloat(v ?? 0)
  if (isNaN(n) || n === 0) return '$0'
  if (n >= 1e6) return '$' + (n / 1e6).toFixed(1) + 'M'
  if (n >= 1e3) return '$' + (n / 1e3).toFixed(0) + 'K'
  return '$' + n.toFixed(0)
}
</script>

<style scoped>
.market-card {
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-top: 2px solid rgba(0, 200, 255, 0.20);
  border-radius: var(--radius);
  padding: 1rem;
  cursor: pointer;
  transition: border-top-color var(--transition), box-shadow var(--transition);
  display: flex;
  flex-direction: column;
  gap: 0.7rem;
  animation: fadeSlideUp 0.25s ease both;
  position: relative;
}
.market-card:hover {
  border-top-color: var(--accent);
  box-shadow: 0 4px 20px rgba(0, 200, 255, 0.10);
}

/* Header */
.card-header {
  display: flex;
  align-items: center;
  gap: 0.35rem;
  min-height: 16px;
}
.hdr-spacer { flex: 1; }

.tag-badge {
  font-size: 0.86rem; font-weight: 700; letter-spacing: 0.08em; text-transform: uppercase;
  padding: 0.1rem 0.4rem; border-radius: 1px;
  background: var(--accent-dim); color: var(--accent); border: 1px solid rgba(0,200,255,0.20);
}
.risk-badge {
  font-size: 0.86rem; font-weight: 700; letter-spacing: 0.08em; text-transform: uppercase;
  padding: 0.1rem 0.4rem; border-radius: 1px;
  background: var(--danger-dim); color: var(--danger); border: 1px solid rgba(255,77,106,0.22);
}
.multi-badge {
  font-size: 0.86rem; font-weight: 700; letter-spacing: 0.08em; text-transform: uppercase;
  padding: 0.1rem 0.4rem; border-radius: 1px;
  background: var(--price-dim); color: var(--price); border: 1px solid rgba(245,158,11,0.22);
}
.end-label {
  font-size: 1.00rem; color: var(--text-muted); font-variant-numeric: tabular-nums;
  font-weight: 600;
}
.end-label.ended { color: var(--danger); }

/* Question */
.market-question {
  font-size: 1.00rem;
  font-weight: 600;
  color: var(--text-primary);
  line-height: 1.45;
  display: -webkit-box;
  -webkit-line-clamp: 3;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

/* Binary probs */
.binary-probs {
  display: flex;
  align-items: stretch;
  background: var(--bg-secondary);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius);
  padding: 0.7rem 0.85rem;
  gap: 0;
}
.binary-side { flex: 1; display: flex; flex-direction: column; gap: 0.3rem; }
.yes-side { padding-right: 0.85rem; }
.no-side  { padding-left: 0.85rem; }
.binary-divider { width: 1px; background: var(--border-subtle); flex-shrink: 0; }

.side-tag {
  font-size: 0.86rem; font-weight: 700; letter-spacing: 0.10em;
  color: var(--text-secondary); text-transform: uppercase;
}
.side-pct {
  font-size: 1.5rem; font-weight: 800; line-height: 1;
  font-variant-numeric: tabular-nums;
}
.yes-pct { color: var(--success); text-shadow: 0 0 10px rgba(16,217,148,0.25); }
.no-pct  { color: var(--danger);  text-shadow: 0 0 10px rgba(255,77,106,0.25); }

.prob-bar { height: 3px; background: var(--border-subtle); border-radius: 1px; overflow: hidden; }
.prob-fill { height: 100%; border-radius: 1px; min-width: 2px; }
.yes-fill { background: var(--success); }
.no-fill  { background: var(--danger); }

/* Categorical */
.outcome-list {
  display: flex; flex-direction: column; gap: 0.3rem;
  background: var(--bg-secondary); border: 1px solid var(--border-subtle);
  border-radius: var(--radius); padding: 0.5rem 0.65rem;
}
.outcome-row { display: flex; align-items: center; gap: 0.5rem; }
.outcome-rank { font-size: 1.00rem; color: var(--text-muted); width: 1.5rem; flex-shrink: 0; text-align: right; font-family: var(--font-mono); }
.outcome-label {
  width: 110px; flex-shrink: 0;
  font-size: 0.86rem; color: var(--text-secondary);
  overflow: hidden; text-overflow: ellipsis; white-space: nowrap;
}
.outcome-bar-wrap { flex: 1; height: 4px; background: var(--border-subtle); border-radius: 1px; overflow: hidden; }
.outcome-bar { height: 100%; background: var(--accent); border-radius: 1px; min-width: 2px; transition: width 0.4s ease; }
.outcome-pct { width: 40px; text-align: right; font-size: 0.86rem; font-weight: 700; color: var(--accent); font-variant-numeric: tabular-nums; flex-shrink: 0; }
.more-label { font-size: 0.90rem; color: var(--text-muted); text-align: right; padding-top: 0.2rem; }

/* Meta row */
.meta-row { display: flex; align-items: center; gap: 0.5rem; }
.meta-item { display: flex; align-items: center; gap: 0.3rem; }
.meta-key { font-size: 0.86rem; font-weight: 700; letter-spacing: 0.10em; color: var(--text-muted); }
.meta-val { font-size: 0.90rem; font-weight: 600; color: var(--text-secondary); font-variant-numeric: tabular-nums; }
.meta-dot { width: 3px; height: 3px; border-radius: 50%; background: var(--border); }

/* Checkbox overlay */
.card-cb-wrap {
  position: absolute; top: 0.55rem; right: 0.55rem;
  opacity: 0; transition: opacity var(--transition);
  z-index: 2;
}
.market-card:hover .card-cb-wrap,
.market-card.is-selected .card-cb-wrap {
  opacity: 1;
}
.card-cb { width: 14px; height: 14px; accent-color: var(--accent); cursor: pointer; }

/* Selected state */
.market-card.is-selected {
  border-top-color: var(--accent);
  box-shadow: 0 0 0 1px rgba(0,200,255,0.18);
}

/* Actions */
.card-actions { display: flex; gap: 0.4rem; margin-top: 0.1rem; }

.btn-yes {
  flex: 1; padding: 0.35rem 0;
  background: rgba(16,217,148,0.08); border: 1px solid var(--success);
  border-radius: var(--radius); color: var(--success);
  font-size: 0.90rem; font-weight: 700; font-family: var(--font-mono);
  letter-spacing: 0.06em; cursor: pointer;
  transition: all var(--transition);
}
.btn-yes:hover { background: var(--success); color: #000; box-shadow: 0 0 12px rgba(16,217,148,0.30); }

.btn-no {
  flex: 1; padding: 0.35rem 0;
  background: rgba(248,113,113,0.08); border: 1px solid var(--danger);
  border-radius: var(--radius); color: var(--danger);
  font-size: 0.90rem; font-weight: 700; font-family: var(--font-mono);
  letter-spacing: 0.06em; cursor: pointer;
  transition: all var(--transition);
}
.btn-no:hover { background: var(--danger); color: #fff; box-shadow: 0 0 12px rgba(248,113,113,0.30); }

.btn-trade {
  flex: 1; padding: 0.35rem 0;
  background: transparent; border: 1px solid var(--accent);
  border-radius: var(--radius); color: var(--accent);
  font-size: 0.94rem; font-weight: 700; font-family: var(--font-mono);
  letter-spacing: 0.08em; cursor: pointer;
  transition: all var(--transition);
}
.btn-trade:hover { background: var(--accent); color: #000; box-shadow: var(--accent-glow); }

.btn-alert {
  padding: 0.35rem 0.6rem;
  background: transparent; border: 1px solid var(--border);
  border-radius: var(--radius); color: var(--text-muted);
  cursor: pointer; display: flex; align-items: center;
  transition: all var(--transition);
}
.btn-alert:hover { border-color: var(--warning); color: var(--warning); }
</style>
