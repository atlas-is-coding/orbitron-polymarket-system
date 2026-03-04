<template>
  <div class="view">
    <div class="view-header anim-in">
      <div class="header-left">
        <h2 class="view-title">{{ $t('nav.orders') }}</h2>
        <span class="counter mono">{{ countLabel }}</span>
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
          class="btn-danger"
          :disabled="canceling"
          @click="confirmCancelAll = true"
        >
          <span :class="{ spin: canceling }">{{ canceling ? '⟳' : $t('orders.cancelAll') }}</span>
        </button>
      </div>
    </div>

    <div class="table-wrap anim-in">
      <!-- Skeleton rows while loading -->
      <div v-if="loading" class="skeleton-wrap">
        <div v-for="i in 5" :key="i" class="skeleton skeleton-row" />
      </div>

      <template v-else>
        <table class="data-table">
          <thead>
            <tr>
              <th>{{ $t('orders.id') }}</th>
              <th>{{ $t('orders.market') }}</th>
              <th>{{ $t('orders.side') }}</th>
              <th>{{ $t('orders.price') }}</th>
              <th>{{ $t('orders.size') }}</th>
              <th>{{ $t('orders.status') }}</th>
              <th></th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="o in filtered" :key="o.id">
              <td class="mono muted">{{ o.id?.slice(0, 8) }}…</td>
              <td class="market-cell">{{ o.market }}</td>
              <td :class="o.side === 'BUY' ? 'text-success' : 'text-danger'" class="fw6">{{ o.side }}</td>
              <td class="mono">{{ o.price }}</td>
              <td class="mono">{{ o.size }}</td>
              <td><span class="badge-status" :class="statusClass(o.status)">{{ o.status }}</span></td>
              <td>
                <button
                  class="btn-xs-danger"
                  :disabled="cancelingId === o.id"
                  @click="doCancel(o.id)"
                >
                  <span :class="{ spin: cancelingId === o.id }">{{ cancelingId === o.id ? '⟳' : $t('orders.cancel') }}</span>
                </button>
              </td>
            </tr>
          </tbody>
        </table>
        <div v-if="!filtered.length" class="empty">{{ $t('orders.noOrders') }}</div>
      </template>
    </div>

    <!-- Confirm Cancel All -->
    <div v-if="confirmCancelAll" class="overlay" @click.self="confirmCancelAll = false">
      <div class="dialog">
        <p>{{ $t('orders.cancelAll') }}?</p>
        <div class="dialog-actions">
          <button class="btn-ghost" @click="confirmCancelAll = false">{{ $t('common.cancel') }}</button>
          <button class="btn-danger-solid" @click="doCancelAll">{{ $t('common.yes') }}</button>
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
  return `${orders.value.length} · ${b} BUY · ${s} SELL`
})

function statusClass(s) {
  if (!s) return ''
  const u = s.toUpperCase()
  if (u === 'OPEN' || u === 'LIVE') return 'badge--accent'
  if (u === 'FILLED' || u === 'MATCHED') return 'badge--ok'
  return 'badge--off'
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
.view { display: flex; flex-direction: column; gap: 1rem; }
.view-header { display: flex; align-items: center; justify-content: space-between; flex-wrap: wrap; gap: 0.5rem; }
.header-left { display: flex; align-items: baseline; gap: 0.75rem; }
.header-right { display: flex; align-items: center; gap: 0.5rem; }
.view-title { font-size: 1.1rem; font-weight: 700; }
.counter { font-size: 0.7rem; color: var(--text-muted); }

.filter-tabs { display: flex; gap: 2px; background: var(--bg-card); border: 1px solid var(--border); border-radius: var(--radius); padding: 2px; }
.filter-tab { padding: 0.18rem 0.6rem; border-radius: calc(var(--radius) - 2px); border: none; background: none; color: var(--text-secondary); font-size: 0.72rem; cursor: pointer; font-family: var(--font-mono); transition: all var(--transition); }
.filter-tab--active { background: var(--accent); color: white; }

.table-wrap { background: var(--bg-card); border: 1px solid var(--border); border-radius: var(--radius); overflow-x: auto; }
.skeleton-wrap { padding: 0.5rem; display: flex; flex-direction: column; gap: 0.4rem; }
.skeleton-row { height: 38px; }

.data-table { width: 100%; border-collapse: collapse; font-size: 0.83rem; }
.data-table th { padding: 0.5rem 1rem; text-align: left; font-size: 0.65rem; text-transform: uppercase; letter-spacing: 0.08em; color: var(--text-secondary); border-bottom: 1px solid var(--border); white-space: nowrap; }
.data-table td { padding: 0.5rem 1rem; border-bottom: 1px solid var(--border-subtle); }
.data-table tr:last-child td { border-bottom: none; }
.data-table tr:hover td { background: var(--bg-hover); }

.mono { font-family: var(--font-mono); font-size: 0.78rem; }
.muted { color: var(--text-secondary); }
.market-cell { max-width: 200px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.text-success { color: var(--success); }
.text-danger  { color: var(--danger); }
.fw6 { font-weight: 600; }

.badge-status { font-size: 0.65rem; font-weight: 600; padding: 0.15rem 0.45rem; border-radius: 999px; }
.badge--accent { background: var(--accent-dim);   color: var(--accent-bright); }
.badge--ok     { background: var(--success-dim);  color: var(--success); }
.badge--off    { background: var(--badge-bg);      color: var(--text-muted); }

.btn-xs-danger { background: var(--danger-dim); border: 1px solid var(--danger); color: var(--danger); border-radius: var(--radius); padding: 0.18rem 0.5rem; font-size: 0.7rem; cursor: pointer; transition: background var(--transition); font-family: var(--font-mono); }
.btn-xs-danger:hover:not(:disabled) { background: var(--danger); color: #fff; }
.btn-xs-danger:disabled { opacity: 0.5; cursor: not-allowed; }

.btn-danger { background: var(--danger-dim); border: 1px solid var(--danger); color: var(--danger); border-radius: var(--radius); padding: 0.28rem 0.75rem; font-size: 0.78rem; cursor: pointer; font-family: var(--font-mono); }
.btn-danger:hover:not(:disabled) { background: var(--danger); color: #fff; }
.btn-danger:disabled { opacity: 0.5; cursor: not-allowed; }

.empty { padding: 2rem; text-align: center; color: var(--text-muted); }

.overlay { position: fixed; inset: 0; background: var(--bg-overlay); display: flex; align-items: center; justify-content: center; z-index: 200; }
.dialog { background: var(--bg-card); border: 1px solid var(--border); border-radius: var(--radius); padding: 1.5rem; min-width: 280px; box-shadow: var(--shadow); }
.dialog p { margin-bottom: 1.25rem; font-size: 0.9rem; }
.dialog-actions { display: flex; gap: 0.75rem; justify-content: flex-end; }
.btn-ghost { background: none; border: 1px solid var(--border); color: var(--text-secondary); border-radius: var(--radius); padding: 0.28rem 0.75rem; font-size: 0.78rem; cursor: pointer; }
.btn-ghost:hover { background: var(--bg-hover); }
.btn-danger-solid { background: var(--danger); color: #fff; border: none; border-radius: var(--radius); padding: 0.28rem 0.75rem; font-size: 0.78rem; cursor: pointer; }
</style>
