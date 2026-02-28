<template>
  <div class="view">
    <h2 class="view-title">{{ $t('nav.positions') }}</h2>

    <div class="table-wrap">
      <table class="data-table">
        <thead>
          <tr>
            <th>{{ $t('positions.market') }}</th>
            <th>{{ $t('positions.side') }}</th>
            <th>{{ $t('positions.size') }}</th>
            <th>{{ $t('positions.avgPrice') }}</th>
            <th>{{ $t('positions.pnl') }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="p in positions" :key="p.market + p.side">
            <td>{{ p.market }}</td>
            <td :class="p.side === 'BUY' ? 'text-success' : 'text-danger'">{{ p.side }}</td>
            <td class="mono">{{ p.size }}</td>
            <td class="mono">{{ p.avg_price }}</td>
            <td class="mono" :class="(p.pnl ?? 0) >= 0 ? 'text-success' : 'text-danger'">
              {{ p.pnl != null ? (p.pnl >= 0 ? '+' : '') + p.pnl.toFixed(2) : '—' }}
            </td>
          </tr>
        </tbody>
      </table>
      <div v-if="!positions.length" class="empty">{{ $t('positions.noPositions') }}</div>
    </div>
  </div>
</template>

<script setup>
import { onMounted } from 'vue'
import { storeToRefs } from 'pinia'
import { useAppStore } from '@/stores/app'
import { useApi } from '@/composables/useApi'

const app = useAppStore()
const { positions } = storeToRefs(app)
const api = useApi()

onMounted(async () => {
  try { app.positions = await api.getPositions() } catch {}
})
</script>

<style scoped>
.view { display: flex; flex-direction: column; gap: 1rem; }
.view-title { font-size: 1.4rem; font-weight: 700; }

.table-wrap {
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  overflow-x: auto;
}
.data-table { width: 100%; border-collapse: collapse; font-size: 0.875rem; }
.data-table th {
  padding: 0.6rem 1rem;
  text-align: left;
  font-size: 0.72rem;
  text-transform: uppercase;
  letter-spacing: 0.06em;
  color: var(--text-secondary);
  border-bottom: 1px solid var(--border);
}
.data-table td {
  padding: 0.6rem 1rem;
  border-bottom: 1px solid var(--border);
  color: var(--text-primary);
}
.data-table tr:last-child td { border-bottom: none; }
.data-table tr:hover td { background: var(--bg-hover); }

.mono         { font-family: var(--font-mono); font-size: 0.82rem; }
.text-success { color: var(--success); font-weight: 600; }
.text-danger  { color: var(--danger);  font-weight: 600; }
.empty { padding: 2rem; text-align: center; color: var(--text-muted); }
</style>
