<template>
  <div class="view">
    <div class="view-header anim-in">
      <div class="header-left">
        <h2 class="view-title">{{ $t('nav.copytrading') }}</h2>
        <span class="status-badge" :class="ct.enabled ? 'badge--ok' : 'badge--off'">
          {{ ct.enabled ? $t('copytrading.enabled') : $t('copytrading.disabled') }}
        </span>
      </div>
      <button class="btn-add" @click="showAdd = true">+ {{ $t('copytrading.addTrader') }}</button>
    </div>

    <!-- Trader cards -->
    <div v-if="ct.traders?.length" class="traders-grid">
      <div v-for="t in ct.traders" :key="t.address" class="trader-card anim-in">
        <div class="card-top">
          <span class="status-dot" :class="t.enabled ? 'status-dot--on' : 'status-dot--off'" />
          <span class="card-label">{{ t.label || 'Unnamed' }}</span>
          <div class="card-badges">
            <span class="alloc-badge">{{ t.allocation_pct }}%</span>
            <span v-if="t.max_position_usd" class="max-badge">${{ t.max_position_usd }}</span>
          </div>
        </div>
        <div class="card-addr">{{ t.address?.slice(0, 12) }}…{{ t.address?.slice(-5) }}</div>
        <div class="card-actions">
          <button
            class="card-btn"
            :disabled="togglingAddr === t.address"
            @click="doToggle(t.address)"
          >
            <span :class="{ spin: togglingAddr === t.address }">
              {{ togglingAddr === t.address ? '⟳' : $t('copytrading.toggle') }}
            </span>
          </button>
          <button class="card-btn" @click="startEdit(t)">{{ $t('copytrading.edit') }}</button>
          <button class="card-btn card-btn--danger" @click="removeTarget = t">{{ $t('copytrading.remove') }}</button>
        </div>
      </div>
    </div>
    <div v-else class="empty-state anim-in">{{ $t('copytrading.noTraders') }}</div>

    <!-- Recent trades feed -->
    <div class="section-header anim-in">Recent Copy Trades</div>
    <div class="trades-panel anim-in">
      <div class="term-chrome">
        <div class="chrome-dots">
          <span class="cdot cdot--r" /><span class="cdot cdot--y" /><span class="cdot cdot--g" />
        </div>
        <span class="term-label">TRADE FEED</span>
        <span class="term-count">{{ copyTrades.length }} recent</span>
      </div>
      <div class="trades-body">
        <div v-if="!copyTrades.length" class="feed-empty">
          <span class="prompt-glyph">$ </span>no recent copy trades
        </div>
        <div v-for="(line, i) in copyTrades" :key="i" class="trade-line">
          <span class="trade-idx">{{ String(i+1).padStart(2, '0') }}</span>
          <span class="trade-sep">│</span>
          <span class="trade-text">{{ line }}</span>
        </div>
      </div>
    </div>

    <!-- Add trader dialog -->
    <div v-if="showAdd" class="overlay" @click.self="showAdd = false">
      <div class="dialog">
        <div class="dialog-title">ADD TRADER</div>
        <div class="form-fields">
          <div class="field">
            <label class="field-label">{{ $t('copytrading.address') }}</label>
            <input v-model="form.address" class="field-input" placeholder="0x..." />
          </div>
          <div class="field">
            <label class="field-label">{{ $t('copytrading.label') }}</label>
            <input v-model="form.label" class="field-input" />
          </div>
          <div class="field">
            <label class="field-label">{{ $t('copytrading.allocation') }} (%)</label>
            <input v-model.number="form.allocPct" type="number" min="0" max="100" class="field-input" />
          </div>
        </div>
        <div class="dialog-actions">
          <button class="btn-ghost" @click="showAdd = false">{{ $t('common.cancel') }}</button>
          <button class="btn-primary" :disabled="!form.address || adding" @click="doAdd">
            <span :class="{ spin: adding }">{{ adding ? '⟳' : $t('common.confirm') }}</span>
          </button>
        </div>
      </div>
    </div>

    <!-- Edit trader dialog -->
    <div v-if="editTarget" class="overlay" @click.self="editTarget = null">
      <div class="dialog">
        <div class="dialog-title">EDIT TRADER</div>
        <div class="form-fields">
          <div class="field">
            <label class="field-label">{{ $t('copytrading.label') }}</label>
            <input v-model="editForm.label" class="field-input" />
          </div>
          <div class="field">
            <label class="field-label">{{ $t('copytrading.allocation') }} (%)</label>
            <input v-model.number="editForm.allocPct" type="number" min="0" max="100" class="field-input" />
          </div>
          <div class="field">
            <label class="field-label">Max Position (USD)</label>
            <input v-model.number="editForm.maxPositionUsd" type="number" min="0" class="field-input" />
          </div>
        </div>
        <div class="dialog-actions">
          <button class="btn-ghost" @click="editTarget = null">{{ $t('common.cancel') }}</button>
          <button class="btn-primary" :disabled="editing" @click="doEdit">
            <span :class="{ spin: editing }">{{ editing ? '⟳' : 'Save' }}</span>
          </button>
        </div>
      </div>
    </div>

    <!-- Confirm remove dialog -->
    <div v-if="removeTarget" class="overlay" @click.self="removeTarget = null">
      <div class="dialog">
        <div class="dialog-title">REMOVE TRADER</div>
        <p class="dialog-body">Remove <span class="mono">{{ removeTarget.label || removeTarget.address?.slice(0, 14) }}…</span>?</p>
        <div class="dialog-actions">
          <button class="btn-ghost" @click="removeTarget = null">{{ $t('common.cancel') }}</button>
          <button class="btn-danger" :disabled="removingAddr === removeTarget?.address" @click="doRemove">
            <span :class="{ spin: removingAddr === removeTarget?.address }">
              {{ removingAddr === removeTarget?.address ? '⟳' : $t('common.confirm') }}
            </span>
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { onMounted, ref, reactive } from 'vue'
import { storeToRefs } from 'pinia'
import { useAppStore } from '@/stores/app'
import { useApi } from '@/composables/useApi'

const app = useAppStore()
const { copytrading: ct, copyTrades } = storeToRefs(app)
const api = useApi()

const showAdd = ref(false)
const form = reactive({ address: '', label: '', allocPct: 5 })
const adding = ref(false)
const togglingAddr = ref(null)
const removingAddr = ref(null)
const removeTarget = ref(null)
const editTarget = ref(null)
const editForm = reactive({ label: '', allocPct: 5, maxPositionUsd: 50 })
const editing = ref(false)

function startEdit(t) {
  editTarget.value = t
  editForm.label = t.label || ''
  editForm.allocPct = t.allocation_pct || 5
  editForm.maxPositionUsd = t.max_position_usd || 50
}

async function doEdit() {
  if (!editTarget.value) return
  editing.value = true
  try {
    await api.editTrader(editTarget.value.address, editForm.label, editForm.allocPct, editForm.maxPositionUsd)
    app.copytrading = await api.getCopytrading()
    editTarget.value = null
  } catch (e) {
    app.toast(e?.response?.data?.error || 'Edit failed', 'error')
  }
  editing.value = false
}

onMounted(async () => {
  try { app.copytrading = await api.getCopytrading() } catch {}
})

async function doToggle(addr) {
  togglingAddr.value = addr
  try { await api.toggleTrader(addr); app.copytrading = await api.getCopytrading() } catch {}
  togglingAddr.value = null
}

async function doRemove() {
  if (!removeTarget.value) return
  const addr = removeTarget.value.address
  removingAddr.value = addr
  try { await api.removeTrader(addr); app.copytrading = await api.getCopytrading() } catch {}
  removingAddr.value = null
  removeTarget.value = null
}

async function doAdd() {
  adding.value = true
  try {
    await api.addTrader(form.address, form.label, form.allocPct)
    app.copytrading = await api.getCopytrading()
    showAdd.value = false
    form.address = ''; form.label = ''; form.allocPct = 5
  } catch {}
  adding.value = false
}
</script>

<style scoped>
.view { display: flex; flex-direction: column; gap: 1rem; }

/* Header */
.view-header { display: flex; align-items: center; justify-content: space-between; flex-wrap: wrap; gap: 0.5rem; }
.header-left { display: flex; align-items: center; gap: 0.65rem; }
.view-title { font-size: 1rem; font-weight: 700; letter-spacing: 0.04em; color: var(--text-bright); }

/* Status badge */
.status-badge {
  font-size: 0.86rem; font-weight: 700; padding: 0.20rem 0.55rem;
  border-radius: 1px; text-transform: uppercase; letter-spacing: 0.06em;
}
.badge--ok  { background: var(--success-dim); color: var(--success); border: 1px solid rgba(16,217,148,0.22); }
.badge--off { background: var(--badge-bg); color: var(--text-muted); border: 1px solid var(--badge-border); }

/* Add button */
.btn-add {
  background: transparent; border: 1px solid var(--accent); color: var(--accent);
  border-radius: var(--radius); padding: 0.38rem 1.00rem; font-size: 0.86rem; font-weight: 600;
  cursor: pointer; font-family: var(--font-mono); letter-spacing: 0.04em;
  transition: all var(--transition);
}
.btn-add:hover { background: var(--accent); color: #000; box-shadow: var(--accent-glow); }

/* Trader grid */
.traders-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(270px, 1fr)); gap: 0.75rem; }

.trader-card {
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-top: 2px solid rgba(0, 200, 255, 0.30);
  border-radius: var(--radius);
  padding: 1rem;
  display: flex; flex-direction: column; gap: 0.65rem;
  transition: border-top-color var(--transition);
}
.trader-card:hover { border-top-color: var(--accent); }

.card-top { display: flex; align-items: center; gap: 0.5rem; }
.status-dot { display: inline-block; width: 6px; height: 6px; border-radius: 50%; flex-shrink: 0; }
.status-dot--on  { background: var(--success); box-shadow: 0 0 6px var(--success); animation: pulse-dot 2.5s ease infinite; }
.status-dot--off { background: var(--text-muted); }
.card-label { flex: 1; font-size: 0.96rem; font-weight: 600; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }

.card-badges { display: flex; align-items: center; gap: 0.3rem; }
.alloc-badge {
  font-size: 1.00rem; font-weight: 700; padding: 0.10rem 0.35rem; border-radius: 1px;
  background: var(--accent-dim); color: var(--accent); border: 1px solid rgba(0,200,255,0.20);
}
.max-badge {
  font-size: 1.00rem; font-weight: 700; padding: 0.10rem 0.35rem; border-radius: 1px;
  background: var(--price-dim); color: var(--price); border: 1px solid rgba(245,158,11,0.20);
}

.card-addr {
  font-family: var(--font-mono); font-size: 0.86rem; color: var(--text-secondary);
  background: var(--bg-hover); padding: 0.2rem 0.5rem; border-radius: var(--radius);
  border: 1px solid var(--border-subtle);
}

.card-actions { display: flex; gap: 0.35rem; }
.card-btn {
  background: var(--bg-hover); border: 1px solid var(--border); color: var(--text-secondary);
  border-radius: var(--radius); padding: 0.30rem 0.70rem; font-size: 0.94rem;
  cursor: pointer; font-family: var(--font-mono); transition: all var(--transition);
}
.card-btn:hover:not(:disabled) { color: var(--accent); border-color: var(--accent); }
.card-btn:disabled { opacity: 0.4; cursor: not-allowed; }
.card-btn--danger { color: var(--danger); border-color: rgba(255,77,106,0.30); }
.card-btn--danger:hover:not(:disabled) { background: var(--danger); color: #fff; border-color: var(--danger); }

/* Section header */
.section-header {
  display: flex; align-items: center; gap: 0.5rem;
  font-size: 1.00rem; text-transform: uppercase; letter-spacing: 0.12em;
  color: var(--accent); font-weight: 600;
}
.section-header::after {
  content: ''; flex: 1; height: 1px;
  background: linear-gradient(90deg, var(--border) 0%, transparent 100%);
}

/* Trades panel */
.trades-panel {
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-top: 1px solid var(--accent);
  border-radius: var(--radius);
  overflow: hidden;
  max-height: 260px;
  display: flex; flex-direction: column;
}

.term-chrome {
  display: flex; align-items: center; gap: 0.6rem;
  padding: 0.4rem 1rem;
  background: rgba(0, 200, 255, 0.04);
  border-bottom: 1px solid var(--border);
  flex-shrink: 0;
}
.chrome-dots { display: flex; gap: 0.25rem; }
.cdot { width: 8px; height: 8px; border-radius: 50%; }
.cdot--r { background: #ff5f57; } .cdot--y { background: #ffbd2e; } .cdot--g { background: #28ca41; }
.term-label { font-size: 0.86rem; letter-spacing: 0.10em; color: var(--text-secondary); flex: 1; text-transform: uppercase; }
.term-count  { font-size: 0.86rem; color: var(--text-muted); }

.trades-body { flex: 1; overflow-y: auto; }
.feed-empty { padding: 1rem; font-size: 0.90rem; color: var(--text-muted); font-family: var(--font-mono); }
.prompt-glyph { color: var(--accent); }

.trade-line {
  display: flex; align-items: baseline; gap: 0.6rem;
  padding: 0.22rem 1rem; border-bottom: 1px solid var(--border-subtle);
  font-family: var(--font-mono); font-size: 0.86rem;
  transition: background var(--transition);
}
.trade-line:last-child { border-bottom: none; }
.trade-line:hover { background: rgba(0, 200, 255, 0.03); }
.trade-idx { color: var(--text-muted); font-size: 0.94rem; flex-shrink: 0; width: 1.5rem; text-align: right; }
.trade-sep { color: var(--border); user-select: none; }
.trade-text { color: var(--text-secondary); }

/* Dialogs */
.overlay {
  position: fixed; inset: 0; background: var(--bg-overlay);
  display: flex; align-items: center; justify-content: center;
  z-index: 200; backdrop-filter: blur(4px);
}
.dialog {
  background: var(--bg-card); border: 1px solid var(--border);
  border-top: 2px solid var(--accent); border-radius: var(--radius);
  padding: 1.5rem; min-width: 320px; box-shadow: var(--shadow-lg);
  animation: fadeSlideUp 0.18s ease both;
}
.dialog-title { font-size: 0.92rem; font-weight: 700; color: var(--accent); text-transform: uppercase; letter-spacing: 0.08em; margin-bottom: 1.25rem; }
.dialog-body { color: var(--text-secondary); font-size: 0.96rem; margin-bottom: 1.25rem; }
.mono { font-family: var(--font-mono); color: var(--text-primary); }

.form-fields { display: flex; flex-direction: column; gap: 0.75rem; margin-bottom: 1.25rem; }
.field { display: flex; flex-direction: column; gap: 0.3rem; }
.field-label { font-size: 1.00rem; font-weight: 600; text-transform: uppercase; letter-spacing: 0.08em; color: var(--text-secondary); }
.field-input {
  padding: 0.45rem 0.65rem; background: var(--bg-input); border: 1px solid var(--border);
  border-radius: var(--radius); color: var(--text-primary); font-family: var(--font-mono);
  font-size: 0.96rem; outline: none; transition: border-color var(--transition);
}
.field-input:focus { border-color: var(--accent); box-shadow: 0 0 0 1px rgba(0,200,255,0.15); }

.dialog-actions { display: flex; gap: 0.5rem; justify-content: flex-end; }
.btn-ghost {
  background: none; border: 1px solid var(--border); color: var(--text-secondary);
  border-radius: var(--radius); padding: 0.38rem 0.90rem; font-size: 0.90rem;
  cursor: pointer; font-family: var(--font-mono); transition: all var(--transition);
}
.btn-ghost:hover { background: var(--bg-hover); }
.btn-primary {
  background: var(--accent); color: #000; border: none; border-radius: var(--radius);
  padding: 0.38rem 1.00rem; font-size: 0.90rem; font-weight: 700;
  cursor: pointer; font-family: var(--font-mono); transition: all var(--transition);
}
.btn-primary:hover:not(:disabled) { background: var(--accent-hover); }
.btn-primary:disabled { opacity: 0.4; cursor: not-allowed; }
.btn-danger {
  background: var(--danger-dim); border: 1px solid var(--danger); color: var(--danger);
  border-radius: var(--radius); padding: 0.38rem 1.00rem; font-size: 0.90rem;
  cursor: pointer; font-family: var(--font-mono); transition: all var(--transition);
}
.btn-danger:hover:not(:disabled) { background: var(--danger); color: #fff; }
.btn-danger:disabled { opacity: 0.4; cursor: not-allowed; }

.empty-state { padding: 2.5rem; text-align: center; color: var(--text-muted); font-size: 0.96rem; }
</style>
