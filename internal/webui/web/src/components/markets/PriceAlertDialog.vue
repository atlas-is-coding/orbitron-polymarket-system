<template>
  <div class="overlay" @click.self="$emit('close')">
    <div class="dialog">
      <!-- Header -->
      <div class="dialog-topbar">
        <span class="dialog-title">SET PRICE ALERT</span>
        <button class="close-btn" @click="$emit('close')">
          <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5">
            <line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/>
          </svg>
        </button>
      </div>

      <div class="dialog-body">
        <!-- Market question -->
        <div class="market-q">{{ market.question }}</div>

        <!-- Outcome selector (categorical) -->
        <template v-if="isCategorical">
          <div class="field-label">OUTCOME</div>
          <div class="outcome-list">
            <button
              v-for="(o, i) in outcomes"
              :key="i"
              class="outcome-btn"
              :class="{ active: selectedOutcome === i }"
              @click="selectedOutcome = i"
            >
              <span class="ob-name">{{ o.label }}</span>
              <span class="ob-pct">{{ fmtPct(o.prob) }}</span>
            </button>
          </div>
        </template>

        <!-- Direction -->
        <div class="field-label">TRIGGER WHEN PRICE</div>
        <div class="direction-row">
          <button class="dir-btn" :class="{ active: direction === 'above' }" @click="direction = 'above'">
            <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5">
              <polyline points="18 15 12 9 6 15"/>
            </svg>
            GOES ABOVE
          </button>
          <button class="dir-btn" :class="{ active: direction === 'below' }" @click="direction = 'below'">
            <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5">
              <polyline points="6 9 12 15 18 9"/>
            </svg>
            FALLS BELOW
          </button>
        </div>

        <!-- Threshold -->
        <div class="field-label">
          THRESHOLD
          <span class="field-hint">0.01 – 0.99</span>
        </div>
        <div class="threshold-wrap">
          <input
            v-model.number="threshold"
            type="number" min="0.01" max="0.99" step="0.01"
            class="threshold-input"
            placeholder="0.75"
          />
          <span class="threshold-suffix">{{ fmtPct(threshold) }}</span>
        </div>

        <!-- Current price preview -->
        <div v-if="currentPrice != null" class="price-preview">
          <span class="pp-key">CURRENT</span>
          <span class="pp-val">{{ fmtPct(currentPrice) }}</span>
          <span class="pp-arrow">→</span>
          <span class="pp-dir" :class="direction === 'above' ? 'pp--up' : 'pp--down'">
            {{ direction === 'above' ? '≥ ' : '≤ ' }}{{ fmtPct(threshold) }}
          </span>
        </div>
      </div>

      <!-- Actions -->
      <div class="dialog-actions">
        <button class="btn-secondary" @click="$emit('close')">CANCEL</button>
        <button class="btn-primary" :disabled="!isValid" @click="submit">CREATE ALERT</button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'

const props = defineProps({ market: { type: Object, required: true } })
const emit = defineEmits(['close', 'created'])

const direction = ref('above')
const threshold = ref(0.75)
const selectedOutcome = ref(0)

const outcomes = computed(() => {
  const prices = props.market.outcomePrices ?? []
  const labels = props.market.outcomes ?? []
  const arr = Array.isArray(prices) ? prices : [prices]
  return arr.map((p, i) => ({
    label: labels[i] ?? (i === 0 ? 'YES' : i === 1 ? 'NO' : `Option ${i + 1}`),
    prob: parseFloat(p) || 0,
  }))
})

const isCategorical = computed(() => outcomes.value.length > 2)
const currentPrice = computed(() => outcomes.value[selectedOutcome.value]?.prob ?? null)
const isValid = computed(() => {
  const t = parseFloat(threshold.value)
  return !isNaN(t) && t >= 0.01 && t <= 0.99
})

function fmtPct(p) {
  if (p == null || isNaN(p)) return '—'
  return (parseFloat(p) * 100).toFixed(1) + '%'
}

function submit() {
  if (!isValid.value) return
  const tokenIds = props.market.clobTokenIds ?? []
  const tokenId = tokenIds[selectedOutcome.value] ?? tokenIds[0] ?? ''
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
.overlay {
  position: fixed; inset: 0;
  background: var(--bg-overlay);
  display: flex; align-items: center; justify-content: center;
  z-index: 1000; backdrop-filter: blur(4px);
}

.dialog {
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-top: 2px solid var(--accent);
  border-radius: var(--radius);
  width: 420px; max-width: 95vw;
  box-shadow: var(--shadow-lg), var(--shadow-cyan);
  animation: fadeSlideUp 0.18s ease both;
  overflow: hidden;
  display: flex; flex-direction: column;
}

/* Topbar */
.dialog-topbar {
  display: flex; align-items: center; justify-content: space-between;
  padding: 0.65rem 1rem;
  background: rgba(0, 200, 255, 0.04);
  border-bottom: 1px solid var(--border);
}
.dialog-title {
  font-size: 1.00rem; font-weight: 700; letter-spacing: 0.12em;
  color: var(--accent); text-transform: uppercase;
}
.close-btn {
  background: none; border: none; color: var(--text-muted); cursor: pointer;
  padding: 0.2rem; display: flex; align-items: center; border-radius: var(--radius-sm);
  transition: color var(--transition);
}
.close-btn:hover { color: var(--danger); }

/* Body */
.dialog-body {
  padding: 1rem; display: flex; flex-direction: column; gap: 0.65rem;
}

.market-q {
  font-size: 0.92rem; color: var(--text-secondary); line-height: 1.5;
  display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical; overflow: hidden;
  background: var(--bg-hover); padding: 0.5rem 0.65rem;
  border-radius: var(--radius); border: 1px solid var(--border-subtle);
}

/* Field label */
.field-label {
  font-size: 0.86rem; font-weight: 700; letter-spacing: 0.10em;
  color: var(--text-secondary); text-transform: uppercase;
  display: flex; align-items: baseline; gap: 0.5rem;
}
.field-hint { font-size: 1.00rem; font-weight: 400; letter-spacing: 0; text-transform: none; color: var(--text-muted); }

/* Outcome list */
.outcome-list { display: flex; flex-direction: column; gap: 0.25rem; max-height: 150px; overflow-y: auto; }
.outcome-btn {
  display: flex; align-items: center; justify-content: space-between;
  padding: 0.4rem 0.65rem;
  background: var(--bg-hover); border: 1px solid var(--border-subtle);
  border-radius: var(--radius); cursor: pointer;
  font-family: var(--font-mono); font-size: 0.90rem;
  color: var(--text-secondary); text-align: left;
  transition: all var(--transition);
}
.outcome-btn.active { border-color: var(--accent); background: var(--accent-dim); color: var(--accent); }
.outcome-btn:hover:not(.active) { background: var(--bg-card); color: var(--text-primary); }
.ob-name { flex: 1; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.ob-pct { font-weight: 700; color: var(--accent); margin-left: 0.5rem; font-variant-numeric: tabular-nums; }
.outcome-btn.active .ob-pct { color: var(--accent-bright); }

/* Direction */
.direction-row { display: flex; gap: 0.4rem; }
.dir-btn {
  flex: 1; padding: 0.45rem 0.5rem;
  border: 1px solid var(--border); border-radius: var(--radius);
  background: var(--bg-hover); cursor: pointer;
  font-size: 0.94rem; font-weight: 700; letter-spacing: 0.06em;
  color: var(--text-muted); font-family: var(--font-mono);
  display: flex; align-items: center; justify-content: center; gap: 0.35rem;
  transition: all var(--transition);
}
.dir-btn.active { border-color: var(--accent); color: var(--accent); background: var(--accent-dim); }
.dir-btn:hover:not(.active) { border-color: var(--text-secondary); color: var(--text-primary); }

/* Threshold */
.threshold-wrap { display: flex; align-items: center; position: relative; }
.threshold-input {
  flex: 1; padding: 0.5rem 3.5rem 0.5rem 0.65rem;
  background: var(--bg-input); border: 1px solid var(--border);
  border-radius: var(--radius); color: var(--text-primary);
  font-family: var(--font-mono); font-size: 1.00rem;
  transition: border-color var(--transition);
}
.threshold-input:focus { outline: none; border-color: var(--accent); box-shadow: 0 0 0 1px rgba(0,200,255,0.15); }
.threshold-suffix {
  position: absolute; right: 0.65rem;
  font-size: 0.86rem; color: var(--text-secondary); pointer-events: none;
  font-variant-numeric: tabular-nums; font-family: var(--font-mono);
}

/* Price preview */
.price-preview {
  display: flex; align-items: center; gap: 0.4rem;
  font-size: 0.86rem; font-family: var(--font-mono);
  background: var(--bg-hover); padding: 0.4rem 0.65rem;
  border-radius: var(--radius); border: 1px solid var(--border-subtle);
}
.pp-key { color: var(--text-secondary); font-weight: 600; letter-spacing: 0.06em; }
.pp-val { color: var(--text-primary); font-weight: 700; }
.pp-arrow { color: var(--text-muted); }
.pp-dir { font-weight: 700; }
.pp--up   { color: var(--success); }
.pp--down { color: var(--danger); }

/* Actions */
.dialog-actions {
  display: flex; gap: 0.5rem;
  padding: 0.75rem 1rem;
  border-top: 1px solid var(--border);
  background: rgba(0, 200, 255, 0.02);
}
.btn-primary {
  flex: 1; padding: 0.5rem;
  background: var(--accent); border: none;
  border-radius: var(--radius); color: #000;
  cursor: pointer; font-family: var(--font-mono);
  font-size: 0.86rem; font-weight: 700; letter-spacing: 0.06em;
  transition: all var(--transition);
}
.btn-primary:hover:not(:disabled) { background: var(--accent-hover); box-shadow: var(--accent-glow); }
.btn-primary:disabled { opacity: 0.35; cursor: not-allowed; }
.btn-secondary {
  padding: 0.5rem 1rem;
  background: transparent; border: 1px solid var(--border);
  border-radius: var(--radius); color: var(--text-secondary);
  cursor: pointer; font-family: var(--font-mono); font-size: 0.86rem; font-weight: 600;
  letter-spacing: 0.04em; transition: all var(--transition);
}
.btn-secondary:hover { border-color: var(--text-secondary); color: var(--text-primary); }
</style>
