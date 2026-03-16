<template>
  <div class="view">
    <div class="view-header anim-in">
      <h2 class="view-title">{{ $t('nav.positions') }}</h2>
      <div v-if="positions.length" class="header-meta">
        <span class="meta-item">{{ positions.length }} positions</span>
        <span class="meta-sep">·</span>
        <span class="meta-item pnl-label" :class="totalPnL >= 0 ? 'pnl-pos' : 'pnl-neg'">
          P&amp;L {{ totalPnL >= 0 ? '+' : '' }}{{ fmt2(totalPnL) }}
        </span>
      </div>
    </div>

    <div class="panel anim-in">
      <table class="data-table">
        <thead>
          <tr>
            <th>{{ $t('positions.market') }}</th>
            <th>SIDE</th>
            <th>{{ $t('positions.size') }}</th>
            <th>{{ $t('positions.avgPrice') }}</th>
            <th class="sortable" @click="toggleSort">
              P&amp;L
              <span class="sort-icon">{{ sortDir === 'desc' ? '↓' : '↑' }}</span>
            </th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="p in sorted" :key="(p.market || '') + (p.side || '')">
            <td class="market-cell">{{ p.market }}</td>
            <td>
              <span class="side-badge" :class="p.side === 'BUY' ? 'side--buy' : 'side--sell'">
                {{ p.side }}
              </span>
            </td>
            <td class="mono">{{ p.size }}</td>
            <td class="mono price-val">{{ p.avg_price }}</td>
            <td>
              <div class="pnl-cell">
                <span class="mono pnl-number" :class="(p.pnl_usd || 0) >= 0 ? 'pnl-pos' : 'pnl-neg'">
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
      <div v-if="!positions.length" class="empty-state">{{ $t('positions.noPositions') }}</div>

      <!-- Summary footer -->
      <div v-if="positions.length" class="summary-footer">
        <div class="sf-item">
          <span class="sf-label">TOTAL P&amp;L</span>
          <span class="sf-val mono" :class="totalPnL >= 0 ? 'pnl-pos' : 'pnl-neg'">
            {{ totalPnL >= 0 ? '+' : '' }}{{ fmt2(totalPnL) }}
          </span>
        </div>
        <div class="sf-sep">│</div>
        <div class="sf-item">
          <span class="sf-label">POSITIONS</span>
          <span class="sf-val mono">{{ positions.length }}</span>
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
const { positions } = storeToRefs(app)
const api = useApi()
const sortDir = ref('desc')

const maxAbsPnl = computed(() => Math.max(...positions.value.map(p => Math.abs(p.pnl_usd || 0)), 1))
const totalPnL = computed(() => positions.value.reduce((s, p) => s + (p.pnl_usd || 0), 0))
const sorted = computed(() =>
  [...positions.value].sort((a, b) =>
    sortDir.value === 'desc'
      ? (b.pnl_usd || 0) - (a.pnl_usd || 0)
      : (a.pnl_usd || 0) - (b.pnl_usd || 0)
  )
)

function fmt2(n) { return (+(n || 0)).toFixed(2) }
function toggleSort() { sortDir.value = sortDir.value === 'desc' ? 'asc' : 'desc' }
function pnlBarWidth(pnl) { return Math.round(Math.abs(pnl || 0) / maxAbsPnl.value * 100) + '%' }

onMounted(async () => {
  try { app.positions = await api.getPositions() } catch {}
})
</script>

<style scoped>
.view { display: flex; flex-direction: column; gap: 0.9rem; }

.view-header { display: flex; align-items: center; justify-content: space-between; flex-wrap: wrap; gap: 0.5rem; }
.view-title { font-size: 1rem; font-weight: 700; letter-spacing: 0.04em; color: var(--text-bright); }
.header-meta { display: flex; align-items: center; gap: 0.4rem; font-size: 0.86rem; font-family: var(--font-mono); }
.meta-item { color: var(--text-secondary); }
.meta-sep { color: var(--text-muted); }
.pnl-label { font-weight: 600; }

/* Panel */
.panel {
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-top: 1px solid var(--accent);
  border-radius: var(--radius);
  overflow-x: auto;
}

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

.market-cell { max-width: 240px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.mono { font-family: var(--font-mono); font-size: 0.92rem; }
.price-val { color: var(--price-bright); }

.sortable { cursor: pointer; user-select: none; }
.sortable:hover { color: var(--accent); }
.sort-icon { margin-left: 0.25rem; font-size: 0.92rem; }

/* Side badge */
.side-badge {
  display: inline-block; font-size: 1.00rem; font-weight: 700;
  padding: 0.18rem 0.55rem; border-radius: 1px; letter-spacing: 0.06em;
}
.side--buy  { background: rgba(16,217,148,0.10); color: var(--success); border: 1px solid rgba(16,217,148,0.22); }
.side--sell { background: rgba(255,77,106,0.10);  color: var(--danger);  border: 1px solid rgba(255,77,106,0.22); }

/* P&L cell */
.pnl-cell { display: flex; flex-direction: column; gap: 0.18rem; }
.pnl-number { font-weight: 600; }
.pnl-bar-track { height: 2px; background: var(--bg-hover); border-radius: 1px; width: 80px; }
.pnl-bar { height: 2px; border-radius: 1px; transition: width var(--transition-slow); min-width: 2px; }
.pnl-bar--pos { background: var(--success); box-shadow: 0 0 4px rgba(16,217,148,0.40); }
.pnl-bar--neg { background: var(--danger); box-shadow: 0 0 4px rgba(255,77,106,0.40); }

/* P&L colors */
.pnl-pos { color: var(--success); text-shadow: 0 0 8px rgba(16,217,148,0.25); }
.pnl-neg { color: var(--danger);  text-shadow: 0 0 8px rgba(255,77,106,0.25); }

/* Summary footer */
.summary-footer {
  display: flex; align-items: center; gap: 0.75rem;
  padding: 0.55rem 1rem;
  border-top: 1px solid var(--border);
  background: rgba(124, 58, 237, 0.02);
}
.sf-item { display: flex; align-items: center; gap: 0.45rem; }
.sf-label { font-size: 0.86rem; text-transform: uppercase; letter-spacing: 0.10em; color: var(--text-secondary); }
.sf-val { font-weight: 700; }
.sf-sep { color: var(--border); user-select: none; }

.empty-state { padding: 2.5rem; text-align: center; color: var(--text-muted); font-size: 0.96rem; }
</style>
