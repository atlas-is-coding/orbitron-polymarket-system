<template>
  <div class="view">
    <div class="view-header anim-in">
      <h2 class="view-title">{{ $t('nav.positions') }}</h2>
    </div>

    <div class="table-wrap anim-in">
      <table class="data-table">
        <thead>
          <tr>
            <th>{{ $t('positions.market') }}</th>
            <th>{{ $t('positions.side') }}</th>
            <th>{{ $t('positions.size') }}</th>
            <th>{{ $t('positions.avgPrice') }}</th>
            <th class="sortable" @click="toggleSort">
              {{ $t('positions.pnl') }}
              <span class="sort-icon">{{ sortDir === 'desc' ? '↓' : '↑' }}</span>
            </th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="p in sorted" :key="(p.market || '') + (p.side || '')">
            <td class="market-cell">{{ p.market }}</td>
            <td :class="p.side === 'BUY' ? 'text-success' : 'text-danger'" class="fw6">{{ p.side }}</td>
            <td class="mono">{{ p.size }}</td>
            <td class="mono">{{ p.avg_price }}</td>
            <td>
              <div class="pnl-cell">
                <span class="mono" :class="(p.pnl_usd || 0) >= 0 ? 'text-success' : 'text-danger'">
                  {{ (p.pnl_usd || 0) >= 0 ? '+' : '' }}{{ fmt2(p.pnl_usd) }}
                </span>
                <div class="pnl-bar-track">
                  <div
                    class="pnl-bar"
                    :class="(p.pnl_usd || 0) >= 0 ? 'pnl-bar--pos' : 'pnl-bar--neg'"
                    :style="{ width: pnlBarWidth(p.pnl_usd) }"
                  />
                </div>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
      <div v-if="!positions.length" class="empty">{{ $t('positions.noPositions') }}</div>

      <!-- Summary footer -->
      <div v-if="positions.length" class="summary-footer">
        <span>Total P&L:
          <span class="mono" :class="totalPnL >= 0 ? 'text-success' : 'text-danger'">
            {{ totalPnL >= 0 ? '+' : '' }}{{ fmt2(totalPnL) }}
          </span>
        </span>
        <span>Positions: <span class="mono">{{ positions.length }}</span></span>
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
const { positions } = storeToRefs(app)
const api = useApi()
const sortDir = ref('desc')

const maxAbsPnl = computed(() =>
  Math.max(...positions.value.map(p => Math.abs(p.pnl_usd || 0)), 1)
)
const totalPnL = computed(() =>
  positions.value.reduce((s, p) => s + (p.pnl_usd || 0), 0)
)
const sorted = computed(() =>
  [...positions.value].sort((a, b) =>
    sortDir.value === 'desc'
      ? (b.pnl_usd || 0) - (a.pnl_usd || 0)
      : (a.pnl_usd || 0) - (b.pnl_usd || 0)
  )
)

function fmt2(n) { return (+(n || 0)).toFixed(2) }
function toggleSort() { sortDir.value = sortDir.value === 'desc' ? 'asc' : 'desc' }
function pnlBarWidth(pnl) {
  return Math.round(Math.abs(pnl || 0) / maxAbsPnl.value * 100) + '%'
}

onMounted(async () => {
  try { app.positions = await api.getPositions() } catch {}
})
</script>

<style scoped>
.view { display: flex; flex-direction: column; gap: 1rem; }
.view-header { display: flex; align-items: center; justify-content: space-between; }
.view-title { font-size: 1.1rem; font-weight: 700; }

.table-wrap { background: var(--bg-card); border: 1px solid var(--border); border-radius: var(--radius); overflow-x: auto; }
.data-table { width: 100%; border-collapse: collapse; font-size: 0.83rem; }
.data-table th { padding: 0.5rem 1rem; text-align: left; font-size: 0.65rem; text-transform: uppercase; letter-spacing: 0.08em; color: var(--text-secondary); border-bottom: 1px solid var(--border); }
.data-table td { padding: 0.5rem 1rem; border-bottom: 1px solid var(--border-subtle); }
.data-table tr:last-child td { border-bottom: none; }
.data-table tr:hover td { background: var(--bg-hover); }

.market-cell { max-width: 220px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.mono { font-family: var(--font-mono); font-size: 0.78rem; }
.text-success { color: var(--success); }
.text-danger  { color: var(--danger); }
.fw6 { font-weight: 600; }

.sortable { cursor: pointer; user-select: none; }
.sortable:hover { color: var(--accent-bright); }
.sort-icon { margin-left: 0.25rem; font-size: 0.6rem; }

.pnl-cell { display: flex; flex-direction: column; gap: 0.22rem; }
.pnl-bar-track { height: 2px; background: var(--bg-hover); border-radius: 1px; width: 72px; }
.pnl-bar { height: 2px; border-radius: 1px; transition: width var(--transition-slow); min-width: 2px; }
.pnl-bar--pos { background: var(--success); }
.pnl-bar--neg { background: var(--danger); }

.summary-footer {
  display: flex; gap: 1.5rem; padding: 0.6rem 1rem;
  border-top: 1px solid var(--border); font-size: 0.75rem; color: var(--text-secondary);
}
.empty { padding: 2rem; text-align: center; color: var(--text-muted); }
</style>
