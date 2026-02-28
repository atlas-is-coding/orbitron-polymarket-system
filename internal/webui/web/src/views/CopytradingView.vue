<template>
  <div class="view">
    <div class="view-header">
      <h2 class="view-title">{{ $t('nav.copytrading') }}</h2>
      <div class="header-right">
        <span class="status-badge" :class="ct.enabled ? 'badge--ok' : 'badge--off'">
          {{ ct.enabled ? $t('copytrading.enabled') : $t('copytrading.disabled') }}
        </span>
        <button class="btn-add" @click="showAdd = true">+ {{ $t('copytrading.addTrader') }}</button>
      </div>
    </div>

    <div class="table-wrap">
      <table class="data-table">
        <thead>
          <tr>
            <th>{{ $t('copytrading.label') }}</th>
            <th>{{ $t('copytrading.address') }}</th>
            <th>{{ $t('copytrading.allocation') }}</th>
            <th>{{ $t('common.connected') }}</th>
            <th></th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="t in ct.traders" :key="t.address">
            <td>{{ t.label || '—' }}</td>
            <td class="mono addr">{{ t.address?.slice(0, 10) }}…</td>
            <td class="mono">{{ t.allocation_pct }}%</td>
            <td>
              <span class="sub-dot" :class="t.enabled ? 'sub-dot--on' : 'sub-dot--off'" />
            </td>
            <td class="actions">
              <button class="btn-xs" @click="doToggle(t.address)">{{ $t('copytrading.toggle') }}</button>
              <button class="btn-xs-danger" @click="doRemove(t.address)">{{ $t('copytrading.remove') }}</button>
            </td>
          </tr>
        </tbody>
      </table>
      <div v-if="!ct.traders?.length" class="empty">{{ $t('copytrading.noTraders') }}</div>
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
            <label class="field-label">{{ $t('copytrading.allocation') }}</label>
            <input v-model.number="form.allocPct" type="number" min="0" max="100" class="field-input" />
          </div>
        </div>
        <div class="dialog-actions">
          <button class="btn-ghost" @click="showAdd = false">{{ $t('common.cancel') }}</button>
          <button class="btn-primary" @click="doAdd" :disabled="!form.address">{{ $t('common.confirm') }}</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { onMounted, ref, reactive } from 'vue'
import { storeToRefs } from 'pinia'
import { useAppStore } from '@/stores/app'
import { useApi } from '@/composables/useApi'
RouterView
const app = useAppStore()
const { copytrading: ct } = storeToRefs(app)
const api = useApi()

const showAdd = ref(false)
const form = reactive({ address: '', label: '', allocPct: 5 })

onMounted(async () => {
  try { app.copytrading = await api.getCopytrading() } catch {}
})

async function doToggle(addr) {
  try {
    await api.toggleTrader(addr)
    app.copytrading = await api.getCopytrading()
  } catch {}
}

async function doRemove(addr) {
  try {
    await api.removeTrader(addr)
    app.copytrading = await api.getCopytrading()
  } catch {}
}

async function doAdd() {
  try {
    await api.addTrader(form.address, form.label, form.allocPct)
    app.copytrading = await api.getCopytrading()
    showAdd.value = false
    form.address = ''; form.label = ''; form.allocPct = 5
  } catch {}
}
</script>

<style scoped>
.view { display: flex; flex-direction: column; gap: 1rem; }
.view-header { display: flex; align-items: center; justify-content: space-between; flex-wrap: wrap; gap: 0.5rem; }
.view-title  { font-size: 1.4rem; font-weight: 700; }
.header-right { display: flex; align-items: center; gap: 0.75rem; }

.status-badge {
  font-size: 0.75rem; font-weight: 600;
  padding: 0.2rem 0.6rem; border-radius: 999px;
}
.badge--ok  { background: rgba(63,185,80,0.15); color: var(--success); }
.badge--off { background: var(--badge-bg); color: var(--text-muted); }

.btn-add {
  background: var(--accent);
  color: #fff;
  border: none;
  border-radius: var(--radius);
  padding: 0.35rem 0.9rem;
  font-size: 0.85rem;
  cursor: pointer;
  transition: background var(--transition);
}
.btn-add:hover { background: var(--accent-hover); }

.table-wrap {
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  overflow-x: auto;
}
.data-table { width: 100%; border-collapse: collapse; font-size: 0.875rem; }
.data-table th {
  padding: 0.6rem 1rem; text-align: left;
  font-size: 0.72rem; text-transform: uppercase; letter-spacing: 0.06em;
  color: var(--text-secondary); border-bottom: 1px solid var(--border);
}
.data-table td {
  padding: 0.6rem 1rem;
  border-bottom: 1px solid var(--border);
  color: var(--text-primary);
}
.data-table tr:last-child td { border-bottom: none; }
.data-table tr:hover td { background: var(--bg-hover); }

.mono { font-family: var(--font-mono); font-size: 0.82rem; }
.addr { color: var(--text-secondary); }

.sub-dot { display: inline-block; width: 8px; height: 8px; border-radius: 50%; }
.sub-dot--on  { background: var(--success); box-shadow: 0 0 6px var(--success); }
.sub-dot--off { background: var(--text-muted); }

.actions { display: flex; gap: 0.4rem; }

.btn-xs {
  background: var(--bg-hover);
  border: 1px solid var(--border);
  color: var(--text-secondary);
  border-radius: var(--radius);
  padding: 0.2rem 0.5rem;
  font-size: 0.75rem;
  cursor: pointer;
}
.btn-xs:hover { color: var(--text-primary); }

.btn-xs-danger {
  background: rgba(248,81,73,0.1);
  border: 1px solid var(--danger);
  color: var(--danger);
  border-radius: var(--radius);
  padding: 0.2rem 0.5rem;
  font-size: 0.75rem;
  cursor: pointer;
}
.btn-xs-danger:hover { background: var(--danger); color: #fff; }

.empty { padding: 2rem; text-align: center; color: var(--text-muted); }

.overlay {
  position: fixed; inset: 0; background: rgba(0,0,0,0.6);
  display: flex; align-items: center; justify-content: center; z-index: 200;
}
.dialog {
  background: var(--bg-card); border: 1px solid var(--border);
  border-radius: var(--radius); padding: 1.5rem; min-width: 320px;
  box-shadow: var(--shadow);
}
.dialog-title { font-size: 1.05rem; font-weight: 700; margin-bottom: 1.25rem; }
.form-fields { display: flex; flex-direction: column; gap: 0.75rem; margin-bottom: 1.25rem; }
.field { display: flex; flex-direction: column; gap: 0.3rem; }
.field-label {
  font-size: 0.75rem; font-weight: 600; text-transform: uppercase;
  letter-spacing: 0.05em; color: var(--text-secondary);
}
.field-input {
  padding: 0.5rem 0.7rem;
  background: var(--bg-primary);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  color: var(--text-primary);
  font-size: 0.9rem;
  font-family: var(--font-mono);
  outline: none;
}
.field-input:focus { border-color: var(--accent); }

.dialog-actions { display: flex; gap: 0.75rem; justify-content: flex-end; }
.btn-ghost {
  background: none; border: 1px solid var(--border); color: var(--text-secondary);
  border-radius: var(--radius); padding: 0.35rem 0.9rem; font-size: 0.85rem; cursor: pointer;
}
.btn-primary {
  background: var(--accent); color: #fff; border: none;
  border-radius: var(--radius); padding: 0.35rem 0.9rem; font-size: 0.85rem; cursor: pointer;
}
.btn-primary:hover:not(:disabled) { background: var(--accent-hover); }
.btn-primary:disabled { opacity: 0.5; cursor: not-allowed; }
</style>
