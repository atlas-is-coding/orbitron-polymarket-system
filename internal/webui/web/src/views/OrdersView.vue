<template>
  <div class="view">
    <div class="view-header">
      <h2 class="view-title">{{ $t('nav.orders') }}</h2>
      <button
        v-if="orders.length"
        class="btn-danger"
        @click="confirmCancelAll = true"
      >{{ $t('orders.cancelAll') }}</button>
    </div>

    <div class="table-wrap">
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
          <tr v-for="o in orders" :key="o.id">
            <td class="mono id-cell">{{ o.id?.slice(0, 8) }}…</td>
            <td>{{ o.market }}</td>
            <td :class="o.side === 'BUY' ? 'text-success' : 'text-danger'">{{ o.side }}</td>
            <td class="mono">{{ o.price }}</td>
            <td class="mono">{{ o.size }}</td>
            <td><span class="badge-status">{{ o.status }}</span></td>
            <td>
              <button class="btn-xs-danger" @click="doCancel(o.id)">
                {{ $t('orders.cancel') }}
              </button>
            </td>
          </tr>
        </tbody>
      </table>
      <div v-if="!orders.length" class="empty">{{ $t('orders.noOrders') }}</div>
    </div>

    <!-- Confirm Cancel All -->
    <div v-if="confirmCancelAll" class="overlay" @click.self="confirmCancelAll = false">
      <div class="dialog">
        <p>{{ $t('orders.cancelAll') }}?</p>
        <div class="dialog-actions">
          <button class="btn-ghost" @click="confirmCancelAll = false">{{ $t('common.cancel') }}</button>
          <button class="btn-danger" @click="doCancelAll">{{ $t('common.yes') }}</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { onMounted } from 'vue'
import { ref } from 'vue'
import { storeToRefs } from 'pinia'
import { useAppStore } from '@/stores/app'
import { useApi } from '@/composables/useApi'

const app = useAppStore()
const { orders } = storeToRefs(app)
const api = useApi()
const confirmCancelAll = ref(false)

onMounted(async () => {
  try { app.orders = await api.getOrders() } catch {}
})

async function doCancel(id) {
  try {
    await api.cancelOrder(id)
    app.orders = orders.value.filter(o => o.id !== id)
  } catch {}
}

async function doCancelAll() {
  confirmCancelAll.value = false
  try {
    await api.cancelAll()
    app.orders = []
  } catch {}
}
</script>

<style scoped>
.view { display: flex; flex-direction: column; gap: 1rem; }
.view-header { display: flex; align-items: center; justify-content: space-between; }
.view-title  { font-size: 1.4rem; font-weight: 700; }

.table-wrap {
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  overflow-x: auto;
}

.data-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 0.875rem;
}

.data-table th {
  padding: 0.6rem 1rem;
  text-align: left;
  font-size: 0.72rem;
  text-transform: uppercase;
  letter-spacing: 0.06em;
  color: var(--text-secondary);
  border-bottom: 1px solid var(--border);
  white-space: nowrap;
}

.data-table td {
  padding: 0.6rem 1rem;
  border-bottom: 1px solid var(--border);
  color: var(--text-primary);
  white-space: nowrap;
}
.data-table tr:last-child td { border-bottom: none; }
.data-table tr:hover td { background: var(--bg-hover); }

.mono     { font-family: var(--font-mono); font-size: 0.82rem; }
.id-cell  { color: var(--text-secondary); }
.text-success { color: var(--success); font-weight: 600; }
.text-danger  { color: var(--danger);  font-weight: 600; }

.badge-status {
  font-size: 0.72rem;
  padding: 0.18rem 0.5rem;
  border-radius: 999px;
  background: var(--badge-bg);
  color: var(--text-secondary);
}

.btn-xs-danger {
  background: rgba(248,81,73,0.12);
  border: 1px solid var(--danger);
  color: var(--danger);
  border-radius: var(--radius);
  padding: 0.2rem 0.5rem;
  font-size: 0.75rem;
  cursor: pointer;
  transition: background var(--transition);
}
.btn-xs-danger:hover { background: var(--danger); color: #fff; }

.btn-danger {
  background: rgba(248,81,73,0.12);
  border: 1px solid var(--danger);
  color: var(--danger);
  border-radius: var(--radius);
  padding: 0.35rem 0.9rem;
  font-size: 0.85rem;
  cursor: pointer;
  transition: background var(--transition);
}
.btn-danger:hover { background: var(--danger); color: #fff; }

.empty { padding: 2rem; text-align: center; color: var(--text-muted); font-size: 0.875rem; }

.overlay {
  position: fixed; inset: 0;
  background: rgba(0,0,0,0.6);
  display: flex; align-items: center; justify-content: center;
  z-index: 200;
}
.dialog {
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  padding: 1.5rem;
  min-width: 280px;
  box-shadow: var(--shadow);
}
.dialog p { margin-bottom: 1.25rem; font-size: 0.95rem; }
.dialog-actions { display: flex; gap: 0.75rem; justify-content: flex-end; }
.btn-ghost {
  background: none;
  border: 1px solid var(--border);
  color: var(--text-secondary);
  border-radius: var(--radius);
  padding: 0.35rem 0.9rem;
  font-size: 0.85rem;
  cursor: pointer;
}
.btn-ghost:hover { background: var(--bg-hover); }
</style>
