<template>
  <div class="panel-overlay" @click.self="$emit('close')">
    <div class="detail-panel">

      <!-- Topbar -->
      <div class="panel-topbar">
        <span class="panel-title">MARKET DETAIL</span>
        <button class="close-btn" @click="$emit('close')">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5">
            <line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/>
          </svg>
        </button>
      </div>

      <!-- Content -->
      <div class="panel-body">
        <!-- Badges -->
        <div class="badges-row">
          <span v-if="market.category" class="badge badge--cat">{{ market.category }}</span>
          <span v-if="market.negRisk" class="badge badge--risk">NEG RISK</span>
          <span v-if="isCategorical" class="badge badge--multi">{{ outcomes.length }} outcomes</span>
          <span v-if="endLabel" class="badge badge--date">{{ endLabel }}</span>
        </div>

        <!-- Question -->
        <h2 class="panel-question">{{ market.question }}</h2>

        <!-- Tags -->
        <div v-if="market.tags?.length" class="tags-row">
          <span v-for="tag in market.tags" :key="tag.slug ?? tag.id" class="tag-chip">
            {{ tag.label }}
          </span>
        </div>

        <div class="divider" />

        <!-- Outcomes -->
        <div class="section-label">OUTCOMES</div>

        <!-- Binary -->
        <div v-if="!isCategorical" class="binary-grid">
          <div class="prob-card prob-card--yes">
            <span class="pc-tag">YES</span>
            <span class="pc-value yes-val">{{ fmtPct(outcomes[0]?.prob) }}</span>
            <div class="pc-bar">
              <div class="pc-fill yes-fill" :style="{ width: fmtPct(outcomes[0]?.prob) }" />
            </div>
          </div>
          <div class="prob-card prob-card--no">
            <span class="pc-tag">NO</span>
            <span class="pc-value no-val">{{ fmtPct(outcomes[1]?.prob) }}</span>
            <div class="pc-bar">
              <div class="pc-fill no-fill" :style="{ width: fmtPct(outcomes[1]?.prob) }" />
            </div>
          </div>
        </div>

        <!-- Categorical -->
        <div v-else class="outcome-table">
          <div v-for="(o, i) in outcomes" :key="i" class="ot-row">
            <span class="ot-rank">{{ String(i+1).padStart(2,'0') }}</span>
            <span class="ot-label">{{ o.label }}</span>
            <div class="ot-bar-wrap"><div class="ot-bar" :style="{ width: fmtPct(o.prob) }" /></div>
            <span class="ot-pct">{{ fmtPct(o.prob) }}</span>
          </div>
        </div>

        <div class="divider" />

        <!-- Stats -->
        <div class="section-label">STATS</div>
        <div class="stats-grid">
          <div class="stat-cell">
            <span class="stat-k">VOLUME</span>
            <span class="stat-v num-glow">{{ fmtMoney(market.volume) }}</span>
          </div>
          <div class="stat-cell">
            <span class="stat-k">LIQUIDITY</span>
            <span class="stat-v">{{ fmtMoney(market.liquidity) }}</span>
          </div>
          <div v-if="market.endDateIso" class="stat-cell">
            <span class="stat-k">CLOSES</span>
            <span class="stat-v">{{ fmtDate(market.endDateIso) }}</span>
          </div>
          <div v-if="market.resolutionSource" class="stat-cell">
            <span class="stat-k">RESOLUTION</span>
            <span class="stat-v stat-v--link">{{ shortUrl(market.resolutionSource) }}</span>
          </div>
        </div>

        <!-- Condition ID -->
        <div class="condition-id">
          <span class="ci-label">ID:</span>
          {{ market.conditionId }}
        </div>

        <div class="divider" />

        <!-- Actions -->
        <div class="panel-actions">
          <button class="btn-place" @click="showPlaceOrder = true">PLACE ORDER</button>
          <button class="btn-alert" @click="$emit('alert', market)">
            <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M18 8A6 6 0 0 0 6 8c0 7-3 9-3 9h18s-3-2-3-9"/>
              <path d="M13.73 21a2 2 0 0 1-3.46 0"/>
            </svg>
            ALERT
          </button>
        </div>
      </div>

      <PlaceOrderDialog v-if="showPlaceOrder" :market="market" @close="showPlaceOrder = false" />
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import PlaceOrderDialog from './PlaceOrderDialog.vue'

const props = defineProps({ market: { type: Object, required: true } })
defineEmits(['close', 'buy', 'alert'])

const showPlaceOrder = ref(false)

const outcomes = computed(() => {
  const prices = props.market.outcomePrices ?? []
  const labels = props.market.outcomes ?? []
  const priceArr = Array.isArray(prices) ? prices : [prices]
  return priceArr.map((p, i) => ({
    label: labels[i] ?? (i === 0 ? 'YES' : i === 1 ? 'NO' : `Option ${i + 1}`),
    prob: parseFloat(p) || 0,
  }))
})

const isCategorical = computed(() => outcomes.value.length > 2)

const endLabel = computed(() => {
  const iso = props.market.endDateIso
  if (!iso) return ''
  try {
    const d = new Date(iso)
    const diff = d - new Date()
    if (diff < 0) return 'Ended'
    return Math.ceil(diff / 86400000) + 'd left'
  } catch { return '' }
})

function fmtPct(p) { return (p == null || isNaN(p)) ? '—' : (p * 100).toFixed(1) + '%' }

function fmtMoney(v) {
  const n = parseFloat(v ?? 0)
  if (isNaN(n) || n === 0) return '$0'
  if (n >= 1e6) return '$' + (n / 1e6).toFixed(1) + 'M'
  if (n >= 1e3) return '$' + (n / 1e3).toFixed(0) + 'K'
  return '$' + n.toFixed(0)
}

function fmtDate(iso) {
  try { return new Date(iso).toLocaleDateString('en', { year: 'numeric', month: 'short', day: 'numeric' }) }
  catch { return iso }
}

function shortUrl(url) {
  try { return new URL(url).hostname.replace('www.', '') }
  catch { return url?.slice(0, 32) ?? '' }
}
</script>

<style scoped>
.panel-overlay {
  position: fixed; inset: 0;
  background: rgba(3, 5, 12, 0.75);
  display: flex; justify-content: flex-end;
  z-index: 900;
  backdrop-filter: blur(3px);
}

.detail-panel {
  width: 480px; max-width: 96vw;
  background: var(--bg-secondary);
  border-left: 1px solid var(--border);
  border-left-color: var(--accent);
  overflow-y: auto;
  animation: slideInRight 0.22s ease;
  display: flex;
  flex-direction: column;
  position: relative;
}

/* Scanline */
.detail-panel::after {
  content: '';
  position: absolute;
  inset: 0;
  background: repeating-linear-gradient(
    0deg, transparent, transparent 2px,
    rgba(124,58,237,0.008) 2px, rgba(124,58,237,0.008) 4px
  );
  pointer-events: none;
  z-index: 0;
}

/* Topbar */
.panel-topbar {
  display: flex; align-items: center; justify-content: space-between;
  padding: 0.7rem 1.25rem;
  background: rgba(124, 58, 237, 0.04);
  border-bottom: 1px solid var(--border);
  position: sticky; top: 0; z-index: 5;
  backdrop-filter: blur(8px);
}

.panel-title {
  font-size: 1.00rem; font-weight: 700; letter-spacing: 0.12em;
  color: var(--accent); text-transform: uppercase;
}

.close-btn {
  background: var(--bg-hover); border: 1px solid var(--border);
  border-radius: var(--radius); color: var(--text-secondary);
  cursor: pointer; padding: 0.3rem;
  display: flex; align-items: center; justify-content: center;
  transition: all var(--transition);
}
.close-btn:hover { color: var(--danger); border-color: var(--danger); }

/* Body */
.panel-body {
  padding: 1.25rem;
  display: flex; flex-direction: column; gap: 0.75rem;
  position: relative; z-index: 1;
}

/* Badges */
.badges-row { display: flex; flex-wrap: wrap; gap: 0.35rem; }
.badge {
  font-size: 0.86rem; font-weight: 700; letter-spacing: 0.06em;
  text-transform: uppercase; padding: 0.20rem 0.55rem; border-radius: 1px;
}
.badge--cat  { background: var(--accent-dim); color: var(--accent); border: 1px solid rgba(124,58,237,0.22); }
.badge--risk { background: var(--danger-dim); color: var(--danger); border: 1px solid rgba(255,77,106,0.22); }
.badge--multi { background: var(--price-dim); color: var(--price); border: 1px solid rgba(245,158,11,0.22); }
.badge--date { background: var(--bg-hover); color: var(--text-muted); border: 1px solid var(--border); }

/* Question */
.panel-question {
  font-size: 1rem; font-weight: 700; color: var(--text-bright); line-height: 1.5;
  margin: 0;
}

/* Tags */
.tags-row { display: flex; flex-wrap: wrap; gap: 0.35rem; }
.tag-chip {
  padding: 0.15rem 0.6rem;
  background: var(--bg-hover); border: 1px solid var(--border);
  border-radius: 1px; font-size: 0.94rem; color: var(--text-secondary);
  transition: border-color var(--transition);
}
.tag-chip:hover { border-color: var(--accent); color: var(--accent); }

/* Divider */
.divider { height: 1px; background: var(--border); }

/* Section label */
.section-label {
  font-size: 0.86rem; font-weight: 700; letter-spacing: 0.12em;
  color: var(--accent); text-transform: uppercase;
}

/* Binary grid */
.binary-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 0.5rem; }
.prob-card {
  background: var(--bg-card); border: 1px solid var(--border);
  border-radius: var(--radius); padding: 0.85rem;
  display: flex; flex-direction: column; gap: 0.4rem;
  transition: border-color var(--transition);
}
.prob-card--yes:hover { border-color: var(--success); }
.prob-card--no:hover  { border-color: var(--danger); }

.pc-tag { font-size: 0.86rem; font-weight: 700; letter-spacing: 0.10em; color: var(--text-muted); }
.pc-value { font-size: 1.8rem; font-weight: 800; font-variant-numeric: tabular-nums; line-height: 1; }
.yes-val { color: var(--success); text-shadow: 0 0 12px rgba(16,217,148,0.30); }
.no-val  { color: var(--danger);  text-shadow: 0 0 12px rgba(255,77,106,0.30); }
.pc-bar { height: 3px; background: var(--border-subtle); border-radius: 1px; overflow: hidden; }
.pc-fill { height: 100%; border-radius: 1px; min-width: 2px; }
.yes-fill { background: var(--success); }
.no-fill  { background: var(--danger); }

/* Categorical outcome table */
.outcome-table { display: flex; flex-direction: column; gap: 0.3rem; }
.ot-row {
  display: flex; align-items: center; gap: 0.6rem;
  padding: 0.5rem 0.75rem;
  background: var(--bg-card); border: 1px solid var(--border-subtle);
  border-radius: var(--radius); transition: border-color var(--transition);
}
.ot-row:hover { border-color: var(--accent); }
.ot-rank { font-size: 1.00rem; color: var(--text-muted); width: 1.5rem; flex-shrink: 0; text-align: right; font-family: var(--font-mono); }
.ot-label { width: 130px; flex-shrink: 0; font-size: 0.94rem; color: var(--text-secondary); overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.ot-bar-wrap { flex: 1; height: 4px; background: var(--border-subtle); border-radius: 1px; overflow: hidden; }
.ot-bar { height: 100%; background: var(--accent); border-radius: 1px; min-width: 2px; }
.ot-pct { width: 48px; text-align: right; font-size: 0.94rem; font-weight: 700; color: var(--accent); font-variant-numeric: tabular-nums; flex-shrink: 0; }

/* Stats */
.stats-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 0.5rem 1rem; }
.stat-cell { display: flex; flex-direction: column; gap: 0.2rem; }
.stat-k { font-size: 0.86rem; text-transform: uppercase; letter-spacing: 0.10em; color: var(--text-secondary); }
.stat-v { font-size: 1.00rem; font-weight: 700; color: var(--text-primary); font-variant-numeric: tabular-nums; }
.stat-v--link { color: var(--accent); font-size: 0.92rem; }
.num-glow { color: var(--price-bright); text-shadow: 0 0 10px rgba(251,191,36,0.30); }

/* Condition ID */
.condition-id {
  font-size: 1.00rem; font-family: var(--font-mono);
  color: var(--text-muted); word-break: break-all; line-height: 1.6;
  background: var(--bg-input); padding: 0.4rem 0.65rem;
  border-radius: var(--radius); border: 1px solid var(--border-subtle);
}
.ci-label { color: var(--text-secondary); font-weight: 600; }

/* Actions */
.panel-actions { display: flex; gap: 0.5rem; }

.btn-place {
  flex: 1; padding: 0.55rem;
  background: transparent; border: 1px solid var(--accent); color: var(--accent);
  border-radius: var(--radius); font-family: var(--font-mono);
  font-size: 0.86rem; font-weight: 700; letter-spacing: 0.08em;
  cursor: pointer; transition: all var(--transition);
}
.btn-place:hover { background: var(--accent); color: #000; box-shadow: var(--accent-glow); }

.btn-alert {
  padding: 0.55rem 1rem;
  background: transparent; border: 1px solid var(--border); color: var(--text-muted);
  border-radius: var(--radius); font-family: var(--font-mono);
  font-size: 0.94rem; font-weight: 700; letter-spacing: 0.06em;
  cursor: pointer; display: flex; align-items: center; gap: 0.35rem;
  transition: all var(--transition);
}
.btn-alert:hover { border-color: var(--warning); color: var(--warning); }
</style>
