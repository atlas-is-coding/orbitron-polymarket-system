<template>
  <div class="wallets-view">
    <!-- KPI row -->
    <div class="kpi-row anim-in">
      <div class="kpi-card"><div class="kpi-lbl">TOTAL BALANCE</div><div class="kpi-val price-val">${{ fmt(totalBalance) }}</div></div>
      <div class="kpi-card"><div class="kpi-lbl">TOTAL P&L</div><div class="kpi-val" :class="totalPnl>=0?'pos-val':'neg-val'">${{ fmt(totalPnl) }}</div></div>
      <div class="kpi-card"><div class="kpi-lbl">ACTIVE WALLETS</div><div class="kpi-val">{{ activeCount }}/{{ wallets.length }}</div></div>
      <div class="kpi-card"><div class="kpi-lbl">OPEN ORDERS</div><div class="kpi-val">{{ totalOrders }}</div></div>
    </div>

    <!-- Action bar -->
    <div class="action-bar anim-in">
      <button class="btn btn-primary" @click="showAdd = true">+ ADD WALLET</button>
    </div>

    <!-- Wallet cards grid -->
    <div class="cards-grid anim-in">
      <div v-for="w in wallets" :key="w.id" class="wallet-card" :class="{ disabled: !w.enabled }">
        <!-- Card header -->
        <div class="card-header">
          <div class="wallet-avatar" :style="avatarStyle(w)">{{ (w.label || w.address || '?')[0].toUpperCase() }}</div>
          <div class="wallet-meta">
            <div class="wallet-name">
              {{ w.label || 'Wallet' }}
              <span v-if="w.primary" class="badge-primary">PRIMARY</span>
            </div>
            <div class="wallet-addr mono">{{ shortAddr(w.address) }}</div>
          </div>
          <span class="status-pill" :class="w.enabled ? 'pill-on' : 'pill-off'">{{ w.enabled ? 'ACTIVE' : 'DISABLED' }}</span>
        </div>

        <!-- Balance -->
        <div class="card-balance">
          <span class="balance-val price-val">${{ fmt(w.balance_usd) }}</span>
          <span class="pnl-val" :class="(w.pnl_usd||0)>=0?'pos-val':'neg-val'">{{ (w.pnl_usd||0)>=0?'+':'' }}${{ fmt(w.pnl_usd) }}</span>
        </div>

        <!-- Stats -->
        <div class="card-stats">
          <div class="stat-item"><span class="stat-lbl">ORDERS</span><span class="stat-val">{{ w.open_orders || 0 }}</span></div>
          <div class="stat-item"><span class="stat-lbl">TRADES</span><span class="stat-val">{{ w.total_trades || 0 }}</span></div>
        </div>

        <!-- Footer actions -->
        <div class="card-footer">
          <button class="btn btn-ghost sm" @click="router.push('/orders?wallet='+w.id)">ORDERS</button>
          <button
            class="btn sm"
            :class="w.enabled ? 'btn-danger' : 'btn-success'"
            :disabled="togglingId === w.id"
            @click="toggleWallet(w)"
          >{{ togglingId===w.id?'...': w.enabled?'DISABLE':'ENABLE' }}</button>
        </div>
      </div>

      <!-- Add wallet placeholder -->
      <div class="wallet-card add-card" @click="showAdd = true">
        <div class="add-icon">+</div>
        <div class="add-label">ADD WALLET</div>
      </div>
    </div>

    <!-- USDC Allowance section -->
    <div class="section-header anim-in">USDC ALLOWANCE</div>
    <div class="panel anim-in">
      <table class="data-table">
        <thead>
          <tr>
            <th>Wallet</th>
            <th>CTF Exchange</th>
            <th>NegRisk Adapter</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="w in wallets" :key="w.id">
            <td class="mono muted-txt">{{ shortAddr(w.address) }}</td>
            <td>
              <button class="btn btn-ghost sm" :disabled="approving[w.id+':ctf']" @click="approve(w.id, 'ctf')">
                {{ approving[w.id+':ctf'] ? '...' : 'APPROVE \u221e' }}
              </button>
            </td>
            <td>
              <button class="btn btn-ghost sm" :disabled="approving[w.id+':negrisk']" @click="approve(w.id, 'negrisk')">
                {{ approving[w.id+':negrisk'] ? '...' : 'APPROVE \u221e' }}
              </button>
            </td>
          </tr>
        </tbody>
      </table>
      <div v-if="!wallets.length" class="empty-state">No wallets</div>
    </div>

    <AddWalletDialog v-if="showAdd" @close="showAdd=false" @added="onWalletAdded" />
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { storeToRefs } from 'pinia'
import { useAppStore } from '@/stores/app'
import { useApi } from '@/composables/useApi'
import AddWalletDialog from '@/components/wallets/AddWalletDialog.vue'

const router = useRouter()
const app = useAppStore()
const { walletsMap } = storeToRefs(app)
const api = useApi()

const showAdd = ref(false)
const togglingId = ref(null)
const approving = ref({})

const wallets = computed(() => Object.values(walletsMap.value))
const totalBalance = computed(() => wallets.value.reduce((s,w) => s+(+w.balance_usd||0), 0))
const totalPnl     = computed(() => wallets.value.reduce((s,w) => s+(+w.pnl_usd||0), 0))
const activeCount  = computed(() => wallets.value.filter(w => w.enabled).length)
const totalOrders  = computed(() => wallets.value.reduce((s,w) => s+(+w.open_orders||0), 0))

function fmt(n) { return n != null ? Number(n||0).toFixed(2) : '—' }
function shortAddr(a) { return a ? a.slice(0,6)+'...'+a.slice(-4) : '—' }

const AVATAR_COLORS = ['#7c3aed','#2563eb','#0891b2','#059669','#d97706']
function avatarStyle(w) {
  const idx = w.address ? w.address.charCodeAt(2) % AVATAR_COLORS.length : 0
  return { background: AVATAR_COLORS[idx] }
}

async function toggleWallet(w) {
  togglingId.value = w.id
  try {
    await api.toggleWallet(w.id, !w.enabled)
    if (walletsMap.value[w.id]) walletsMap.value[w.id].enabled = !w.enabled
    app.toast(`Wallet ${w.enabled ? 'disabled' : 'enabled'}`, 'success')
  } catch { app.toast('Failed to toggle wallet', 'error') }
  togglingId.value = null
}

async function approve(walletId, contract) {
  const key = walletId + ':' + contract
  approving.value[key] = true
  try {
    await api.approveAllowance(walletId, contract)
    app.toast('Approval submitted', 'success')
  } catch (e) { app.toast(e?.response?.data?.error || 'Approval failed', 'error') }
  delete approving.value[key]
}

function onWalletAdded(wallet) {
  if (wallet?.id) walletsMap.value[wallet.id] = wallet
}
</script>

<style scoped>
.wallets-view { display: flex; flex-direction: column; gap: 16px; }

.kpi-row { display: grid; grid-template-columns: repeat(auto-fit, minmax(150px, 1fr)); gap: 10px; }
.kpi-card { background: var(--bg-card); border: 1px solid var(--border); border-radius: var(--radius); padding: 12px 16px; }
.kpi-lbl { font-size: 10px; text-transform: uppercase; letter-spacing: 0.10em; color: var(--text-secondary); margin-bottom: 6px; }
.kpi-val { font-size: 20px; font-weight: 700; color: var(--text-primary); font-family: var(--font-mono); }

.action-bar { display: flex; gap: 8px; }

.cards-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(320px, 1fr)); gap: 14px; }

.wallet-card {
  background: var(--bg-card); border: 1px solid var(--border); border-radius: var(--radius-xl);
  padding: 16px; display: flex; flex-direction: column; gap: 12px;
  transition: border-color 0.15s;
}
.wallet-card:hover { border-color: var(--accent); }
.wallet-card.disabled { opacity: 0.6; }

.card-header { display: flex; align-items: center; gap: 10px; }
.wallet-avatar { width: 36px; height: 36px; border-radius: 50%; display: flex; align-items: center; justify-content: center; font-size: 16px; font-weight: 700; color: #fff; flex-shrink: 0; }
.wallet-meta { flex: 1; min-width: 0; }
.wallet-name { font-size: 13px; font-weight: 600; color: var(--text-primary); display: flex; align-items: center; gap: 6px; }
.wallet-addr { font-size: 11px; color: var(--text-secondary); overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.badge-primary { font-size: 9px; padding: 1px 6px; border-radius: 2px; background: var(--accent-dim); color: var(--accent); border: 1px solid rgba(124,58,237,0.30); text-transform: uppercase; letter-spacing: 0.06em; }
.status-pill { font-size: 10px; font-weight: 700; padding: 2px 8px; border-radius: 3px; border: 1px solid; letter-spacing: 0.06em; flex-shrink: 0; }
.pill-on  { color: var(--success); border-color: rgba(16,217,148,0.40); background: rgba(16,217,148,0.08); }
.pill-off { color: var(--text-secondary); border-color: var(--border); background: transparent; }

.card-balance { display: flex; align-items: baseline; gap: 10px; }
.balance-val { font-size: 22px; font-weight: 700; font-family: var(--font-mono); }
.pnl-val { font-size: 13px; font-family: var(--font-mono); }

.card-stats { display: flex; gap: 16px; }
.stat-item { display: flex; flex-direction: column; gap: 2px; }
.stat-lbl { font-size: 9px; text-transform: uppercase; letter-spacing: 0.10em; color: var(--text-secondary); }
.stat-val { font-size: 14px; font-weight: 600; color: var(--text-primary); font-family: var(--font-mono); }

.card-footer { display: flex; gap: 6px; flex-wrap: wrap; margin-top: auto; }

.add-card { border-style: dashed; border-color: var(--border); cursor: pointer; align-items: center; justify-content: center; min-height: 120px; }
.add-card:hover { border-color: var(--accent); background: rgba(124,58,237,0.04); }
.add-icon { font-size: 24px; color: var(--text-muted); }
.add-label { font-size: 12px; color: var(--text-muted); letter-spacing: 0.08em; }

.section-header { font-size: 10px; text-transform: uppercase; letter-spacing: 0.12em; color: var(--accent); font-weight: 600; padding: 4px 0; border-bottom: 1px solid var(--border); margin-top: 4px; }

.panel { background: var(--bg-card); border: 1px solid var(--border); border-top: 1px solid var(--accent); border-radius: var(--radius); overflow-x: auto; }
.data-table { width: 100%; border-collapse: collapse; font-size: 12px; }
.data-table th { padding: 9px 14px; text-align: left; font-size: 10px; text-transform: uppercase; letter-spacing: 0.10em; color: var(--text-secondary); border-bottom: 1px solid var(--border); background: rgba(124,58,237,0.03); }
.data-table td { padding: 8px 14px; border-bottom: 1px solid rgba(255,255,255,0.03); vertical-align: middle; }
.data-table tr:last-child td { border-bottom: none; }
.mono { font-family: var(--font-mono); }
.muted-txt { color: var(--text-secondary); }
.empty-state { padding: 2rem; text-align: center; color: var(--text-muted); font-size: 12px; }

.price-val { color: var(--price-bright); }
.pos-val { color: var(--success); }
.neg-val { color: var(--danger); }

.btn { display: inline-flex; align-items: center; padding: 5px 12px; border-radius: var(--radius); font-family: var(--font-mono); font-size: 12px; font-weight: 500; cursor: pointer; border: 1px solid transparent; transition: all var(--transition); white-space: nowrap; }
.btn.sm { padding: 3px 9px; font-size: 11px; }
.btn-ghost  { background: none; border-color: var(--border); color: var(--text-secondary); }
.btn-ghost:hover:not(:disabled) { background: var(--bg-hover); color: var(--text-primary); }
.btn-primary { background: var(--accent); color: #fff; border-color: var(--accent); }
.btn-danger  { background: var(--danger-dim); color: var(--danger); border-color: var(--danger); }
.btn-success { background: var(--success-dim); color: var(--success); border-color: rgba(16,217,148,0.40); }
.btn:disabled { opacity: 0.4; cursor: not-allowed; }
</style>
