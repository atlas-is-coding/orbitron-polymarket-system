<template>
  <div class="view">
    <div class="view-header anim-in">
      <div class="header-left">
        <h2 class="view-title">{{ $t('nav.wallets') }}</h2>
        <span v-if="wallets.length" class="counter">{{ activeCount }}/{{ wallets.length }} {{ $t('wallets.active') }}</span>
      </div>
      <button class="btn-add" @click="showAdd = true">+ {{ $t('wallets.addWallet') }}</button>
    </div>

    <!-- Summary bar -->
    <div v-if="wallets.length" class="summary-bar anim-in">
      <div class="summary-item">
        <span class="summary-label">{{ $t('wallets.totalBalance') }}</span>
        <span class="summary-value num-glow">${{ totalBalance }}</span>
      </div>
      <div class="summary-sep">│</div>
      <div class="summary-item">
        <span class="summary-label">{{ $t('wallets.totalPnL') }}</span>
        <span class="summary-value" :class="totalPnL >= 0 ? 'pnl-pos' : 'pnl-neg'">
          {{ totalPnL >= 0 ? '+' : '' }}{{ fmt2(totalPnL) }}
        </span>
      </div>
      <div class="summary-sep">│</div>
      <div class="summary-item">
        <span class="summary-label">WALLETS</span>
        <span class="summary-value">{{ activeCount }}/{{ wallets.length }}</span>
      </div>
    </div>

    <!-- Table -->
    <div class="panel anim-in">
      <table class="data-table">
        <thead>
          <tr>
            <th>{{ $t('wallets.label') }}</th>
            <th>{{ $t('wallets.address') }}</th>
            <th>{{ $t('wallets.balance') }}</th>
            <th>{{ $t('wallets.pnl') }}</th>
            <th>ORDERS</th>
            <th>{{ $t('wallets.status') }}</th>
            <th></th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="w in wallets" :key="w.id">
            <td class="label-cell">
              <span v-if="editingId !== w.id">{{ w.label || w.id }}</span>
              <input
                v-else
                v-model="editLabel"
                class="inline-input"
                @keyup.enter="saveRename(w.id)"
                @keyup.esc="editingId = null"
              />
            </td>
            <td class="mono addr-cell">{{ w.address ? w.address.slice(0, 8) + '…' + w.address.slice(-4) : '—' }}</td>
            <td class="mono price-val">${{ fmt2(w.balance_usd) }}</td>
            <td class="mono" :class="w.pnl_usd >= 0 ? 'pnl-pos' : 'pnl-neg'">
              {{ w.pnl_usd >= 0 ? '+' : '' }}{{ fmt2(w.pnl_usd) }}
            </td>
            <td class="mono">{{ w.open_orders ?? 0 }}</td>
            <td>
              <div class="status-cell">
                <span class="status-dot" :class="w.enabled ? 'status-dot--on' : 'status-dot--off'" />
                <span class="st-text">{{ w.enabled ? 'ON' : 'OFF' }}</span>
              </div>
            </td>
            <td class="actions-cell">
              <button class="act-btn" :disabled="togglingId === w.id" @click="doToggle(w)">
                <span :class="{ spin: togglingId === w.id }">
                  {{ togglingId === w.id ? '⟳' : (w.enabled ? $t('wallets.disable') : $t('wallets.enable')) }}
                </span>
              </button>
              <button class="act-btn" v-if="editingId !== w.id" @click="startRename(w)">{{ $t('wallets.rename') }}</button>
              <button class="act-btn act-btn--save" v-else @click="saveRename(w.id)">Save</button>
              <button class="act-btn act-btn--danger" @click="removeTarget = w">{{ $t('wallets.remove') }}</button>
            </td>
          </tr>
        </tbody>
      </table>
      <div v-if="!wallets.length" class="empty-state">{{ $t('wallets.noWallets') }}</div>
    </div>

    <!-- Add wallet dialog -->
    <div v-if="showAdd" class="overlay" @click.self="showAdd = false">
      <div class="dialog">
        <div class="dialog-title">ADD WALLET</div>
        <div class="form-fields">
          <div class="field">
            <label class="field-label">{{ $t('wallets.privateKey') }}</label>
            <input
              v-model="addForm.privateKey"
              type="password"
              class="field-input mono"
              placeholder="64-char hex, no 0x prefix"
              autocomplete="off"
            />
            <span class="field-hint">{{ $t('wallets.privateKeyHint') }}</span>
          </div>
          <div class="field">
            <label class="field-label">{{ $t('wallets.label') }} <span class="field-opt">(optional)</span></label>
            <input v-model="addForm.label" class="field-input" />
          </div>
        </div>
        <div class="dialog-actions">
          <button class="btn-ghost" @click="showAdd = false">{{ $t('common.cancel') }}</button>
          <button class="btn-primary" :disabled="!addForm.privateKey || adding" @click="doAddWallet">
            <span :class="{ spin: adding }">{{ adding ? '⟳' : $t('wallets.addWallet') }}</span>
          </button>
        </div>
      </div>
    </div>

    <!-- Confirm remove dialog -->
    <div v-if="removeTarget" class="overlay" @click.self="removeTarget = null">
      <div class="dialog">
        <div class="dialog-title">REMOVE WALLET</div>
        <p class="dialog-body">{{ $t('wallets.confirmRemoveMsg', { label: removeTarget.label || removeTarget.id }) }}</p>
        <div class="dialog-actions">
          <button class="btn-ghost" @click="removeTarget = null">{{ $t('common.cancel') }}</button>
          <button class="btn-danger" :disabled="removingId === removeTarget?.id" @click="doRemove">
            <span :class="{ spin: removingId === removeTarget?.id }">
              {{ removingId === removeTarget?.id ? '⟳' : $t('common.confirm') }}
            </span>
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed, onMounted, ref, reactive } from 'vue'
import { storeToRefs } from 'pinia'
import { useAppStore } from '@/stores/app'
import { useApi } from '@/composables/useApi'

const app = useAppStore()
const { walletsMap } = storeToRefs(app)
const api = useApi()

const editingId = ref(null)
const editLabel = ref('')
const removeTarget = ref(null)
const removingId = ref(null)
const togglingId = ref(null)
const showAdd = ref(false)
const adding = ref(false)
const addForm = reactive({ privateKey: '', label: '' })

const wallets = computed(() => Object.values(walletsMap.value))
const totalBalance = computed(() => fmt2(wallets.value.reduce((s, w) => s + (w.balance_usd || 0), 0)))
const totalPnL = computed(() => wallets.value.reduce((s, w) => s + (w.pnl_usd || 0), 0))
const activeCount = computed(() => wallets.value.filter(w => w.enabled).length)

function fmt2(n) { return (+(n || 0)).toFixed(2) }

async function refreshWallets() {
  const list = await api.getWallets()
  if (Array.isArray(list)) {
    const m = {}
    for (const w of list) m[w.id] = w
    app.walletsMap = m
  }
}

onMounted(async () => { try { await refreshWallets() } catch {} })

async function doToggle(w) {
  togglingId.value = w.id
  try { await api.toggleWallet(w.id, !w.enabled); await refreshWallets() } catch {}
  togglingId.value = null
}

function startRename(w) { editingId.value = w.id; editLabel.value = w.label || '' }

async function saveRename(id) {
  if (!editLabel.value.trim()) { editingId.value = null; return }
  try { await api.renameWallet(id, editLabel.value.trim()); await refreshWallets() } catch {}
  editingId.value = null
}

async function doRemove() {
  if (!removeTarget.value) return
  const id = removeTarget.value.id
  removingId.value = id
  try { await api.removeWallet(id); await refreshWallets() } catch {}
  removingId.value = null; removeTarget.value = null
}

async function doAddWallet() {
  adding.value = true
  try {
    await api.addWallet(addForm.privateKey)
    await refreshWallets()
    app.toast('Wallet added', 'success')
    showAdd.value = false
    addForm.privateKey = ''; addForm.label = ''
  } catch (e) {
    app.toast(e?.response?.data?.error || 'Failed to add wallet', 'error')
  }
  adding.value = false
}
</script>

<style scoped>
.view { display: flex; flex-direction: column; gap: 0.9rem; }

.view-header { display: flex; align-items: center; justify-content: space-between; flex-wrap: wrap; gap: 0.5rem; }
.header-left { display: flex; align-items: center; gap: 0.65rem; }
.view-title { font-size: 1rem; font-weight: 700; letter-spacing: 0.04em; color: var(--text-bright); }
.counter { font-size: 0.92rem; color: var(--text-secondary); font-family: var(--font-mono); }

.btn-add {
  background: transparent; border: 1px solid var(--accent); color: var(--accent);
  border-radius: var(--radius); padding: 0.38rem 1.00rem; font-size: 0.86rem; font-weight: 600;
  cursor: pointer; font-family: var(--font-mono); letter-spacing: 0.04em; transition: all var(--transition);
}
.btn-add:hover { background: var(--accent); color: #000; box-shadow: var(--accent-glow); }

/* Summary bar */
.summary-bar {
  display: flex; align-items: center; gap: 0.75rem; flex-wrap: wrap;
  padding: 0.75rem 1.25rem;
  background: var(--bg-card); border: 1px solid var(--border);
  border-top: 1px solid var(--accent); border-radius: var(--radius);
}
.summary-item { display: flex; align-items: center; gap: 0.45rem; }
.summary-label { font-size: 0.86rem; text-transform: uppercase; letter-spacing: 0.10em; color: var(--text-secondary); }
.summary-value { font-size: 1rem; font-weight: 700; font-family: var(--font-mono); color: var(--text-primary); }
.summary-sep { color: var(--border); user-select: none; }

.num-glow { color: var(--price-bright); text-shadow: 0 0 10px rgba(251,191,36,0.35); }
.pnl-pos  { color: var(--success); text-shadow: 0 0 8px rgba(16,217,148,0.25); }
.pnl-neg  { color: var(--danger);  text-shadow: 0 0 8px rgba(255,77,106,0.25); }

/* Panel */
.panel {
  background: var(--bg-card); border: 1px solid var(--border);
  border-top: 1px solid var(--accent); border-radius: var(--radius); overflow-x: auto;
}

/* Table */
.data-table { width: 100%; border-collapse: collapse; font-size: 0.96rem; }
.data-table th {
  padding: 0.6rem 1.2rem; text-align: left; font-size: 1.00rem;
  text-transform: uppercase; letter-spacing: 0.10em; color: var(--text-secondary);
  border-bottom: 1px solid var(--border); white-space: nowrap;
  background: rgba(0, 200, 255, 0.03);
}
.data-table td { padding: 0.6rem 1.2rem; border-bottom: 1px solid var(--border-subtle); vertical-align: middle; }
.data-table tr:nth-child(even) td { background: rgba(0,200,255,0.018); }
.data-table tr:last-child td { border-bottom: none; }
.data-table tr:hover td { background: var(--bg-hover) !important; }

.mono { font-family: var(--font-mono); font-size: 0.92rem; }
.addr-cell { color: var(--text-secondary); font-size: 0.86rem; }
.price-val { color: var(--price-bright); }
.label-cell { min-width: 120px; }

.inline-input {
  padding: 0.2rem 0.45rem; background: var(--bg-input);
  border: 1px solid var(--accent); border-radius: var(--radius);
  color: var(--text-primary); font-size: 0.96rem; outline: none;
  width: 100%; font-family: var(--font-mono);
}

/* Status cell */
.status-cell { display: flex; align-items: center; gap: 0.35rem; }
.status-dot { display: inline-block; width: 6px; height: 6px; border-radius: 50%; flex-shrink: 0; }
.status-dot--on  { background: var(--success); box-shadow: 0 0 6px var(--success); animation: pulse-dot 2.5s ease infinite; }
.status-dot--off { background: var(--text-muted); }
.st-text { font-size: 0.86rem; color: var(--text-secondary); }

/* Action buttons */
.actions-cell { display: flex; gap: 0.3rem; flex-wrap: nowrap; }
.act-btn {
  background: var(--bg-hover); border: 1px solid var(--border); color: var(--text-secondary);
  border-radius: var(--radius); padding: 0.25rem 0.65rem; font-size: 0.94rem;
  cursor: pointer; font-family: var(--font-mono); transition: all var(--transition); white-space: nowrap;
}
.act-btn:hover:not(:disabled) { color: var(--accent); border-color: var(--accent); }
.act-btn:disabled { opacity: 0.4; cursor: not-allowed; }
.act-btn--save { border-color: var(--accent); color: var(--accent); }
.act-btn--danger { color: var(--danger); border-color: rgba(255,77,106,0.30); }
.act-btn--danger:hover:not(:disabled) { background: var(--danger); color: #fff; border-color: var(--danger); }

.empty-state { padding: 2.5rem; text-align: center; color: var(--text-muted); font-size: 0.96rem; }

/* Dialogs */
.overlay { position: fixed; inset: 0; background: var(--bg-overlay); display: flex; align-items: center; justify-content: center; z-index: 200; backdrop-filter: blur(4px); }
.dialog { background: var(--bg-card); border: 1px solid var(--border); border-top: 2px solid var(--accent); border-radius: var(--radius); padding: 1.5rem; min-width: 320px; box-shadow: var(--shadow-lg); animation: fadeSlideUp 0.18s ease both; }
.dialog-title { font-size: 0.92rem; font-weight: 700; color: var(--accent); text-transform: uppercase; letter-spacing: 0.08em; margin-bottom: 1.25rem; }
.dialog-body { color: var(--text-secondary); font-size: 0.96rem; margin-bottom: 1.25rem; line-height: 1.6; }

.form-fields { display: flex; flex-direction: column; gap: 0.75rem; margin-bottom: 1.25rem; }
.field { display: flex; flex-direction: column; gap: 0.3rem; }
.field-label { font-size: 1.00rem; font-weight: 600; text-transform: uppercase; letter-spacing: 0.08em; color: var(--text-secondary); }
.field-opt { font-weight: 400; text-transform: none; letter-spacing: 0; color: var(--text-muted); }
.field-input {
  padding: 0.45rem 0.65rem; background: var(--bg-input); border: 1px solid var(--border);
  border-radius: var(--radius); color: var(--text-primary); font-family: var(--font-mono);
  font-size: 0.96rem; outline: none; transition: border-color var(--transition);
}
.field-input:focus { border-color: var(--accent); box-shadow: 0 0 0 1px rgba(0,200,255,0.15); }
.field-hint { font-size: 0.90rem; color: var(--text-muted); }

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
.btn-primary:hover:not(:disabled) { background: var(--accent-hover); box-shadow: var(--accent-glow); }
.btn-primary:disabled { opacity: 0.4; cursor: not-allowed; }
.btn-danger {
  background: var(--danger-dim); border: 1px solid var(--danger); color: var(--danger);
  border-radius: var(--radius); padding: 0.38rem 1.00rem; font-size: 0.90rem;
  cursor: pointer; font-family: var(--font-mono); transition: all var(--transition);
}
.btn-danger:hover:not(:disabled) { background: var(--danger); color: #fff; }
.btn-danger:disabled { opacity: 0.4; cursor: not-allowed; }
</style>
