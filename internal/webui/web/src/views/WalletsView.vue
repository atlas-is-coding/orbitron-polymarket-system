<template>
  <div class="view">
    <div class="view-header anim-in">
      <div class="header-left">
        <h2 class="view-title">{{ $t('nav.wallets') }}</h2>
        <span v-if="wallets.length" class="counter mono">{{ activeCount }}/{{ wallets.length }} {{ $t('wallets.active') }}</span>
      </div>
      <button class="btn-add" @click="showAdd = true">+ {{ $t('wallets.addWallet') }}</button>
    </div>

    <!-- Summary bar -->
    <div v-if="wallets.length" class="summary-bar anim-in">
      <div class="summary-item">
        <span class="summary-label">{{ $t('wallets.totalBalance') }}</span>
        <span class="summary-value">${{ totalBalance }}</span>
      </div>
      <div class="summary-item">
        <span class="summary-label">{{ $t('wallets.totalPnL') }}</span>
        <span class="summary-value" :class="totalPnL >= 0 ? 'pnl-pos' : 'pnl-neg'">
          {{ totalPnL >= 0 ? '+' : '' }}{{ fmt2(totalPnL) }}
        </span>
      </div>
    </div>

    <div class="table-wrap anim-in">
      <table class="data-table">
        <thead>
          <tr>
            <th>{{ $t('wallets.label') }}</th>
            <th>{{ $t('wallets.address') }}</th>
            <th>{{ $t('wallets.balance') }}</th>
            <th>{{ $t('wallets.pnl') }}</th>
            <th>{{ $t('wallets.openOrders') }}</th>
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
            <td class="mono addr">{{ w.address ? w.address.slice(0, 8) + '…' + w.address.slice(-4) : '—' }}</td>
            <td class="mono">${{ fmt2(w.balance_usd) }}</td>
            <td class="mono" :class="w.pnl_usd >= 0 ? 'pnl-pos' : 'pnl-neg'">
              {{ w.pnl_usd >= 0 ? '+' : '' }}{{ fmt2(w.pnl_usd) }}
            </td>
            <td class="mono">{{ w.open_orders ?? 0 }}</td>
            <td>
              <div class="status-cell">
                <span class="sub-dot" :class="w.enabled ? 'sub-dot--on' : 'sub-dot--off'" />
                <span class="status-text">{{ w.enabled ? $t('wallets.on') : $t('wallets.off') }}</span>
              </div>
            </td>
            <td class="actions">
              <button
                class="btn-xs"
                :disabled="togglingId === w.id"
                @click="doToggle(w)"
              >
                <span :class="{ spin: togglingId === w.id }">
                  {{ togglingId === w.id ? '⟳' : (w.enabled ? $t('wallets.disable') : $t('wallets.enable')) }}
                </span>
              </button>
              <button class="btn-xs" v-if="editingId !== w.id" @click="startRename(w)">{{ $t('wallets.rename') }}</button>
              <button class="btn-xs btn-xs--save" v-else @click="saveRename(w.id)">{{ $t('common.confirm') }}</button>
              <button class="btn-xs-danger" @click="removeTarget = w">{{ $t('wallets.remove') }}</button>
            </td>
          </tr>
        </tbody>
      </table>
      <div v-if="!wallets.length" class="empty">{{ $t('wallets.noWallets') }}</div>
    </div>

    <!-- Add wallet dialog -->
    <div v-if="showAdd" class="overlay" @click.self="showAdd = false">
      <div class="dialog">
        <h3 class="dialog-title">{{ $t('wallets.addWallet') }}</h3>
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
        <h3 class="dialog-title">{{ $t('wallets.confirmRemove') }}</h3>
        <p class="dialog-body">{{ $t('wallets.confirmRemoveMsg', { label: removeTarget.label || removeTarget.id }) }}</p>
        <div class="dialog-actions">
          <button class="btn-ghost" @click="removeTarget = null">{{ $t('common.cancel') }}</button>
          <button class="btn-danger-solid" :disabled="removingId === removeTarget?.id" @click="doRemove">
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

onMounted(async () => {
  try { await refreshWallets() } catch {}
})

async function doToggle(w) {
  togglingId.value = w.id
  try { await api.toggleWallet(w.id, !w.enabled); await refreshWallets() } catch {}
  togglingId.value = null
}

function startRename(w) { editingId.value = w.id; editLabel.value = w.label || '' }

async function saveRename(id) {
  if (!editLabel.value.trim()) { editingId.value = null; return }
  try { await api.renameWallet(id, editLabel.value.trim()) } catch {}
  editingId.value = null
}

async function doRemove() {
  if (!removeTarget.value) return
  const id = removeTarget.value.id
  removingId.value = id
  try { await api.removeWallet(id); await refreshWallets() } catch {}
  removingId.value = null
  removeTarget.value = null
}

async function doAddWallet() {
  adding.value = true
  try {
    await api.addWallet(addForm.privateKey)
    await refreshWallets()
    app.toast('Wallet added', 'success')
    showAdd.value = false
    addForm.privateKey = ''
    addForm.label = ''
  } catch (e) {
    app.toast(e?.response?.data?.error || 'Failed to add wallet', 'error')
  }
  adding.value = false
}
</script>

<style scoped>
.view { display: flex; flex-direction: column; gap: 1rem; }
.view-header { display: flex; align-items: center; justify-content: space-between; flex-wrap: wrap; gap: 0.5rem; }
.header-left { display: flex; align-items: center; gap: 0.75rem; }
.view-title { font-size: 1.1rem; font-weight: 700; }
.counter { font-size: 0.7rem; color: var(--text-muted); }

.btn-add { background: var(--accent); color: #fff; border: none; border-radius: var(--radius); padding: 0.3rem 0.85rem; font-size: 0.78rem; cursor: pointer; font-family: var(--font-mono); transition: background var(--transition); }
.btn-add:hover { background: var(--accent-hover); }

.summary-bar { display: flex; gap: 2rem; flex-wrap: wrap; padding: 0.75rem 1.25rem; background: var(--bg-card); border: 1px solid var(--border); border-radius: var(--radius); }
.summary-item { display: flex; flex-direction: column; gap: 0.15rem; }
.summary-label { font-size: 0.65rem; text-transform: uppercase; letter-spacing: 0.08em; color: var(--text-secondary); }
.summary-value { font-size: 1.05rem; font-weight: 700; font-family: var(--font-mono); color: var(--text-primary); }

.pnl-pos { color: var(--success); }
.pnl-neg { color: var(--danger); }

.table-wrap { background: var(--bg-card); border: 1px solid var(--border); border-radius: var(--radius); overflow-x: auto; }
.data-table { width: 100%; border-collapse: collapse; font-size: 0.83rem; }
.data-table th { padding: 0.5rem 1rem; text-align: left; font-size: 0.65rem; text-transform: uppercase; letter-spacing: 0.08em; color: var(--text-secondary); border-bottom: 1px solid var(--border); white-space: nowrap; }
.data-table td { padding: 0.5rem 1rem; border-bottom: 1px solid var(--border-subtle); }
.data-table tr:last-child td { border-bottom: none; }
.data-table tr:hover td { background: var(--bg-hover); }

.mono { font-family: var(--font-mono); font-size: 0.78rem; }
.addr { color: var(--text-secondary); }
.label-cell { min-width: 120px; }

.inline-input { padding: 0.2rem 0.4rem; background: var(--bg-input); border: 1px solid var(--accent); border-radius: var(--radius); color: var(--text-primary); font-size: 0.83rem; outline: none; width: 100%; font-family: var(--font-mono); }

.status-cell { display: flex; align-items: center; gap: 0.35rem; }
.sub-dot { width: 7px; height: 7px; border-radius: 50%; flex-shrink: 0; }
.sub-dot--on  { background: var(--success); box-shadow: 0 0 5px var(--success); animation: pulse-dot 2.5s ease infinite; }
.sub-dot--off { background: var(--text-muted); }
.status-text { font-size: 0.75rem; color: var(--text-secondary); }

.actions { display: flex; gap: 0.4rem; flex-wrap: nowrap; }

.btn-xs { background: var(--bg-hover); border: 1px solid var(--border); color: var(--text-secondary); border-radius: var(--radius); padding: 0.2rem 0.55rem; font-size: 0.72rem; cursor: pointer; font-family: var(--font-mono); transition: all var(--transition); white-space: nowrap; }
.btn-xs:hover:not(:disabled) { color: var(--text-primary); border-color: var(--accent); }
.btn-xs:disabled { opacity: 0.5; cursor: not-allowed; }
.btn-xs--save { border-color: var(--accent); color: var(--accent-bright); }

.btn-xs-danger { background: var(--danger-dim); border: 1px solid var(--danger); color: var(--danger); border-radius: var(--radius); padding: 0.2rem 0.55rem; font-size: 0.72rem; cursor: pointer; font-family: var(--font-mono); white-space: nowrap; }
.btn-xs-danger:hover { background: var(--danger); color: #fff; }

.empty { padding: 2rem; text-align: center; color: var(--text-muted); }

.overlay { position: fixed; inset: 0; background: var(--bg-overlay); display: flex; align-items: center; justify-content: center; z-index: 200; }
.dialog { background: var(--bg-card); border: 1px solid var(--border); border-radius: var(--radius); padding: 1.5rem; min-width: 320px; box-shadow: var(--shadow); }
.dialog-title { font-size: 1rem; font-weight: 700; margin-bottom: 1rem; }
.dialog-body { color: var(--text-secondary); font-size: 0.85rem; margin-bottom: 1.25rem; }

.form-fields { display: flex; flex-direction: column; gap: 0.75rem; margin-bottom: 1.25rem; }
.field { display: flex; flex-direction: column; gap: 0.3rem; }
.field-label { font-size: 0.65rem; font-weight: 600; text-transform: uppercase; letter-spacing: 0.06em; color: var(--text-secondary); }
.field-opt { font-weight: 400; text-transform: none; letter-spacing: 0; color: var(--text-muted); }
.field-input { padding: 0.45rem 0.65rem; background: var(--bg-input); border: 1px solid var(--border); border-radius: var(--radius); color: var(--text-primary); font-size: 0.85rem; font-family: var(--font-mono); outline: none; }
.field-input:focus { border-color: var(--accent); }
.field-hint { font-size: 0.65rem; color: var(--text-muted); }

.dialog-actions { display: flex; gap: 0.75rem; justify-content: flex-end; }
.btn-ghost { background: none; border: 1px solid var(--border); color: var(--text-secondary); border-radius: var(--radius); padding: 0.28rem 0.75rem; font-size: 0.78rem; cursor: pointer; }
.btn-ghost:hover { background: var(--bg-hover); }
.btn-primary { background: var(--accent); color: #fff; border: none; border-radius: var(--radius); padding: 0.28rem 0.75rem; font-size: 0.78rem; cursor: pointer; font-family: var(--font-mono); }
.btn-primary:hover:not(:disabled) { background: var(--accent-hover); }
.btn-primary:disabled { opacity: 0.5; cursor: not-allowed; }
.btn-danger-solid { background: var(--danger); color: #fff; border: none; border-radius: var(--radius); padding: 0.28rem 0.75rem; font-size: 0.78rem; cursor: pointer; }
.btn-danger-solid:disabled { opacity: 0.5; cursor: not-allowed; }
</style>
