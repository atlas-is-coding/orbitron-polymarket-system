<template>
  <div class="view">
    <div class="view-header anim-in">
      <div class="header-left">
        <h2 class="view-title">{{ $t('nav.orders') }}</h2>
        <span class="counter">{{ countLabel }}</span>
      </div>
      <div class="header-right">
        <div class="filter-tabs">
          <button
            v-for="f in filters"
            :key="f"
            class="filter-tab"
            :class="{ 'filter-tab--active': filter === f }"
            @click="filter = f"
          >{{ f }}</button>
        </div>
        <button
          v-if="orders.length"
          class="btn btn-danger"
          :disabled="canceling"
          @click="confirmCancelAll = true"
        >
          <span :class="{ spin: canceling }">{{ canceling ? '⟳' : $t('orders.cancelAll') }}</span>
        </button>
      </div>
    </div>

    <div class="panel anim-in">
      <!-- Skeleton -->
      <div v-if="loading" class="skeleton-wrap">
        <div v-for="i in 5" :key="i" class="skeleton skeleton-row" />
      </div>

      <template v-else>
        <table class="data-table">
          <thead>
            <tr>
              <th>{{ $t('orders.id') }}</th>
              <th>{{ $t('orders.market') }}</th>
              <th>SIDE</th>
              <th>{{ $t('orders.price') }}</th>
              <th>{{ $t('orders.size') }}</th>
              <th>{{ $t('orders.status') }}</th>
              <th></th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="o in filtered" :key="o.id">
              <td class="mono muted id-cell">{{ o.id?.slice(0, 8) }}…</td>
              <td class="market-cell">{{ o.market }}</td>
              <td>
                <span class="side-badge" :class="o.side === 'BUY' ? 'side--buy' : 'side--sell'">
                  {{ o.side }}
                </span>
              </td>
              <td class="mono price-cell">{{ o.price }}</td>
              <td class="mono">{{ o.size }}</td>
              <td>
                <span class="status-badge" :class="statusClass(o.status)">{{ o.status }}</span>
              </td>
              <td class="action-cell">
                <button
                  class="btn-cancel"
                  :disabled="cancelingId === o.id"
                  @click="doCancel(o.id)"
                >
                  <span :class="{ spin: cancelingId === o.id }">{{ cancelingId === o.id ? '⟳' : $t('orders.cancel') }}</span>
                </button>
              </td>
            </tr>
          </tbody>
        </table>
        <div v-if="!filtered.length" class="empty-state">{{ $t('orders.noOrders') }}</div>
      </template>
    </div>

    <!-- Confirm Cancel All dialog -->
    <div v-if="confirmCancelAll" class="overlay" @click.self="confirmCancelAll = false">
      <div class="dialog">
        <div class="dialog-title">CANCEL ALL ORDERS</div>
        <p class="dialog-body">Cancel all {{ orders.length }} open orders? This action cannot be undone.</p>
        <div class="dialog-actions">
          <button class="btn btn-ghost" @click="confirmCancelAll = false">{{ $t('common.cancel') }}</button>
          <button class="btn btn-danger-solid" @click="doCancelAll">{{ $t('common.yes') }}</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed, onMounted, ref } from 'vue'
import { storeToRefs } from 'pinia'
import { useAppStore } from '@/stores/app'
import { useApi } from '@/composables/useApi'

const app = useAppStore()
const { orders } = storeToRefs(app)
const api = useApi()

const loading = ref(true)
const filter = ref('ALL')
const filters = ['ALL', 'BUY', 'SELL']
const confirmCancelAll = ref(false)
const canceling = ref(false)
const cancelingId = ref(null)

const filtered = computed(() =>
  filter.value === 'ALL' ? orders.value : orders.value.filter(o => o.side === filter.value)
)

const countLabel = computed(() => {
  const b = orders.value.filter(o => o.side === 'BUY').length
  const s = orders.value.filter(o => o.side === 'SELL').length
  return `${orders.value.length} total · ${b} buy · ${s} sell`
})

function statusClass(s) {
  if (!s) return ''
  const u = s.toUpperCase()
  if (u === 'OPEN' || u === 'LIVE') return 'st--accent'
  if (u === 'FILLED' || u === 'MATCHED') return 'st--ok'
  return 'st--off'
}

onMounted(async () => {
  try { app.orders = await api.getOrders() } catch {}
  loading.value = false
})

async function doCancel(id) {
  cancelingId.value = id
  try { await api.cancelOrder(id); app.orders = orders.value.filter(o => o.id !== id) } catch {}
  cancelingId.value = null
}

async function doCancelAll() {
  confirmCancelAll.value = false; canceling.value = true
  try { await api.cancelAll(); app.orders = [] } catch {}
  canceling.value = false
}
</script>

<style scoped>
.view { display: flex; flex-direction: column; gap: 0.9rem; }

/* Header */
.view-header { display: flex; align-items: center; justify-content: space-between; flex-wrap: wrap; gap: 0.5rem; }
.header-left { display: flex; align-items: baseline; gap: 0.75rem; }
.header-right { display: flex; align-items: center; gap: 0.5rem; }
.view-title { font-size: 1rem; font-weight: 700; letter-spacing: 0.04em; color: var(--text-bright); }
.counter { font-size: 0.92rem; color: var(--text-secondary); font-family: var(--font-mono); }

/* Filter tabs */
.filter-tabs {
  display: flex; gap: 1px;
  background: var(--bg-hover);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  padding: 2px;
}
.filter-tab {
  padding: 0.2rem 0.7rem;
  border-radius: calc(var(--radius) - 1px);
  border: none;
  background: none;
  color: var(--text-secondary);
  font-size: 0.94rem;
  font-family: var(--font-mono);
  font-weight: 600;
  cursor: pointer;
  transition: all var(--transition);
  letter-spacing: 0.04em;
}
.filter-tab:hover { color: var(--text-primary); }
.filter-tab--active { background: var(--bg-card); color: var(--accent); }

/* Buttons */
.btn {
  display: inline-flex; align-items: center; gap: 0.3rem;
  padding: 0.38rem 0.90rem;
  border-radius: var(--radius);
  font-family: var(--font-mono);
  font-size: 0.86rem;
  font-weight: 500;
  cursor: pointer;
  transition: all var(--transition);
  border: 1px solid transparent;
  white-space: nowrap;
}
.btn-danger { background: var(--danger-dim); color: var(--danger); border-color: var(--danger); }
.btn-danger:hover:not(:disabled) { background: var(--danger); color: #fff; }
.btn-ghost { background: none; border-color: var(--border); color: var(--text-secondary); }
.btn-ghost:hover { background: var(--bg-hover); }
.btn-danger-solid { background: var(--danger); color: #fff; border-color: var(--danger); }
.btn:disabled { opacity: 0.4; cursor: not-allowed; }

/* Panel */
.panel {
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-top: 1px solid var(--accent);
  border-radius: var(--radius);
  overflow-x: auto;
}

/* Skeleton */
.skeleton-wrap { padding: 0.75rem; display: flex; flex-direction: column; gap: 0.45rem; }
.skeleton-row { height: 36px; }

/* Table */
.data-table { width: 100%; border-collapse: collapse; font-size: 0.96rem; }
.data-table th {
  padding: 0.6rem 1.2rem; text-align: left; font-size: 1.00rem;
  text-transform: uppercase; letter-spacing: 0.10em;
  color: var(--text-secondary); border-bottom: 1px solid var(--border);
  background: rgba(124, 58, 237, 0.03); white-space: nowrap;
}
.data-table td { padding: 0.6rem 1.2rem; border-bottom: 1px solid var(--border-subtle); vertical-align: middle; }
.data-table tr:nth-child(even) td { background: rgba(124,58,237,0.018); }
.data-table tr:last-child td { border-bottom: none; }
.data-table tr:hover td { background: var(--bg-hover) !important; }

.mono { font-family: var(--font-mono); font-size: 0.92rem; }
.muted { color: var(--text-secondary); }
.id-cell { font-size: 0.94rem; }
.market-cell { max-width: 200px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.price-cell { color: var(--price-bright); }
.action-cell { text-align: right; }

/* Side badge */
.side-badge {
  display: inline-block;
  font-size: 0.90rem;
  font-weight: 700;
  padding: 0.20rem 0.55rem;
  border-radius: 1px;
  letter-spacing: 0.06em;
}
.side--buy  { background: rgba(16,217,148,0.12); color: var(--success); border: 1px solid rgba(16,217,148,0.25); }
.side--sell { background: rgba(255,77,106,0.12);  color: var(--danger);  border: 1px solid rgba(255,77,106,0.25); }

/* Status badge */
.status-badge {
  font-size: 1.00rem; font-weight: 600;
  padding: 0.20rem 0.55rem; border-radius: 1px;
  text-transform: uppercase; letter-spacing: 0.06em;
}
.st--accent { background: var(--accent-dim); color: var(--accent); border: 1px solid rgba(124,58,237,0.20); }
.st--ok     { background: var(--success-dim); color: var(--success); border: 1px solid rgba(16,217,148,0.20); }
.st--off    { background: var(--badge-bg); color: var(--text-muted); border: 1px solid var(--badge-border); }

/* Cancel button */
.btn-cancel {
  background: none; border: 1px solid rgba(255,77,106,0.30); color: var(--danger);
  border-radius: var(--radius); padding: 0.25rem 0.65rem; font-size: 0.94rem;
  cursor: pointer; font-family: var(--font-mono); transition: all var(--transition);
}
.btn-cancel:hover:not(:disabled) { background: var(--danger); color: #fff; border-color: var(--danger); }
.btn-cancel:disabled { opacity: 0.4; cursor: not-allowed; }

.empty-state { padding: 2.5rem; text-align: center; color: var(--text-muted); font-size: 0.96rem; }

/* Dialog */
.overlay { position: fixed; inset: 0; background: var(--bg-overlay); display: flex; align-items: center; justify-content: center; z-index: 200; backdrop-filter: blur(4px); }
.dialog { background: var(--bg-card); border: 1px solid var(--border); border-top: 2px solid var(--accent); border-radius: var(--radius); padding: 1.5rem; min-width: 300px; box-shadow: var(--shadow-lg); }
.dialog-title { font-size: 0.94rem; font-weight: 700; color: var(--accent); text-transform: uppercase; letter-spacing: 0.08em; margin-bottom: 0.75rem; }
.dialog-body { color: var(--text-secondary); font-size: 0.96rem; margin-bottom: 1.25rem; line-height: 1.6; }
.dialog-actions { display: flex; gap: 0.5rem; justify-content: flex-end; }
</style>
