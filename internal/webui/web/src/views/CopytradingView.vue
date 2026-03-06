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
          <span class="card-status-dot" :class="t.enabled ? 'dot--on' : 'dot--off'" />
          <span class="card-label">{{ t.label || 'Unnamed' }}</span>
          <span class="card-alloc">{{ t.allocation_pct }}%</span>
        </div>
        <div class="card-addr mono">{{ t.address?.slice(0, 10) }}…{{ t.address?.slice(-4) }}</div>
        <div class="card-actions">
          <button
            class="btn-xs"
            :disabled="togglingAddr === t.address"
            @click="doToggle(t.address)"
          >
            <span :class="{ spin: togglingAddr === t.address }">
              {{ togglingAddr === t.address ? '⟳' : $t('copytrading.toggle') }}
            </span>
          </button>
          <button class="btn-xs" @click="startEdit(t)">✏️ {{ $t('copytrading.edit') }}</button>
          <button class="btn-xs-danger" @click="removeTarget = t">{{ $t('copytrading.remove') }}</button>
        </div>
      </div>
    </div>
    <div v-else class="empty anim-in">{{ $t('copytrading.noTraders') }}</div>

    <!-- Recent Copy Trades -->
    <div class="section-header anim-in">Recent Trades</div>
    <div class="trades-feed anim-in">
      <div v-if="!copyTrades.length" class="empty">No recent copy trades.</div>
      <div v-for="(line, i) in copyTrades" :key="i" class="trade-line mono">{{ line }}</div>
    </div>

    <!-- Add trader dialog -->
    <div v-if="showAdd" class="overlay" @click.self="showAdd = false">
      <div class="dialog">
        <h3 class="dialog-title">{{ $t('copytrading.addTrader') }}</h3>
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
        <h3 class="dialog-title">{{ $t('copytrading.edit') }} Trader</h3>
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
        <h3 class="dialog-title">{{ $t('copytrading.remove') }}</h3>
        <p class="dialog-body">Remove trader <span class="mono">{{ removeTarget.label || removeTarget.address?.slice(0, 10) }}…</span>?</p>
        <div class="dialog-actions">
          <button class="btn-ghost" @click="removeTarget = null">{{ $t('common.cancel') }}</button>
          <button class="btn-danger-solid" :disabled="removingAddr === removeTarget?.address" @click="doRemove">
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
import { onMounted, ref, reactive, computed } from 'vue'
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
  try {
    await api.removeTrader(addr)
    app.copytrading = await api.getCopytrading()
  } catch {}
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
.view-header { display: flex; align-items: center; justify-content: space-between; flex-wrap: wrap; gap: 0.5rem; }
.header-left { display: flex; align-items: center; gap: 0.75rem; }
.view-title { font-size: 1.1rem; font-weight: 700; }

.status-badge { font-size: 0.65rem; font-weight: 600; padding: 0.18rem 0.5rem; border-radius: 999px; }
.badge--ok  { background: var(--success-dim); color: var(--success); }
.badge--off { background: var(--badge-bg);    color: var(--text-muted); }

.btn-add { background: var(--accent); color: #fff; border: none; border-radius: var(--radius); padding: 0.3rem 0.85rem; font-size: 0.78rem; cursor: pointer; font-family: var(--font-mono); transition: background var(--transition); }
.btn-add:hover { background: var(--accent-hover); }

/* Trader cards grid */
.traders-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(260px, 1fr)); gap: 0.75rem; }

.trader-card {
  background: var(--bg-card); border: 1px solid var(--border);
  border-radius: var(--radius); padding: 1rem;
  display: flex; flex-direction: column; gap: 0.6rem;
  transition: border-color var(--transition);
}
.trader-card:hover { border-color: var(--accent); }

.card-top { display: flex; align-items: center; gap: 0.5rem; }
.card-status-dot { width: 7px; height: 7px; border-radius: 50%; flex-shrink: 0; }
.dot--on  { background: var(--success); box-shadow: 0 0 5px var(--success); animation: pulse-dot 2.5s ease infinite; }
.dot--off { background: var(--text-muted); }
.card-label { flex: 1; font-size: 0.85rem; font-weight: 600; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.card-alloc { font-size: 0.72rem; color: var(--accent-bright); font-family: var(--font-mono); font-weight: 600; }

.card-addr { font-size: 0.72rem; color: var(--text-secondary); }

.card-actions { display: flex; gap: 0.4rem; }

.btn-xs { background: var(--bg-hover); border: 1px solid var(--border); color: var(--text-secondary); border-radius: var(--radius); padding: 0.2rem 0.55rem; font-size: 0.72rem; cursor: pointer; font-family: var(--font-mono); transition: all var(--transition); }
.btn-xs:hover:not(:disabled) { color: var(--text-primary); border-color: var(--accent); }
.btn-xs:disabled { opacity: 0.5; cursor: not-allowed; }

.btn-xs-danger { background: var(--danger-dim); border: 1px solid var(--danger); color: var(--danger); border-radius: var(--radius); padding: 0.2rem 0.55rem; font-size: 0.72rem; cursor: pointer; font-family: var(--font-mono); }
.btn-xs-danger:hover { background: var(--danger); color: #fff; }

.empty { padding: 2rem; text-align: center; color: var(--text-muted); }

.mono { font-family: var(--font-mono); }

.overlay { position: fixed; inset: 0; background: var(--bg-overlay); display: flex; align-items: center; justify-content: center; z-index: 200; }
.dialog { background: var(--bg-card); border: 1px solid var(--border); border-radius: var(--radius); padding: 1.5rem; min-width: 320px; box-shadow: var(--shadow); }
.dialog-title { font-size: 1rem; font-weight: 700; margin-bottom: 1rem; }
.dialog-body { color: var(--text-secondary); font-size: 0.85rem; margin-bottom: 1.25rem; }

.form-fields { display: flex; flex-direction: column; gap: 0.75rem; margin-bottom: 1.25rem; }
.field { display: flex; flex-direction: column; gap: 0.3rem; }
.field-label { font-size: 0.65rem; font-weight: 600; text-transform: uppercase; letter-spacing: 0.06em; color: var(--text-secondary); }
.field-input { padding: 0.45rem 0.65rem; background: var(--bg-input); border: 1px solid var(--border); border-radius: var(--radius); color: var(--text-primary); font-size: 0.85rem; font-family: var(--font-mono); outline: none; }
.field-input:focus { border-color: var(--accent); }

.dialog-actions { display: flex; gap: 0.75rem; justify-content: flex-end; }
.btn-ghost { background: none; border: 1px solid var(--border); color: var(--text-secondary); border-radius: var(--radius); padding: 0.28rem 0.75rem; font-size: 0.78rem; cursor: pointer; }
.btn-ghost:hover { background: var(--bg-hover); }
.btn-primary { background: var(--accent); color: #fff; border: none; border-radius: var(--radius); padding: 0.28rem 0.75rem; font-size: 0.78rem; cursor: pointer; font-family: var(--font-mono); }
.btn-primary:hover:not(:disabled) { background: var(--accent-hover); }
.btn-primary:disabled { opacity: 0.5; cursor: not-allowed; }
.btn-danger-solid { background: var(--danger); color: #fff; border: none; border-radius: var(--radius); padding: 0.28rem 0.75rem; font-size: 0.78rem; cursor: pointer; }
.btn-danger-solid:disabled { opacity: 0.5; cursor: not-allowed; }

.section-header { font-size: 0.65rem; text-transform: uppercase; letter-spacing: 0.08em; color: var(--text-secondary); }
.trades-feed {
  background: var(--bg-card); border: 1px solid var(--border);
  border-radius: var(--radius); padding: 0.5rem 1rem;
  max-height: 220px; overflow-y: auto;
}
.trade-line { font-size: 0.78rem; padding: 0.2rem 0; color: var(--text-secondary); border-bottom: 1px solid var(--border-subtle); }
.trade-line:last-child { border-bottom: none; }
</style>
