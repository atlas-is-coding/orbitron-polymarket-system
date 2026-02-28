<template>
  <div class="view">
    <h2 class="view-title">{{ $t('nav.overview') }}</h2>

    <div class="stat-grid">
      <div class="stat-card">
        <div class="stat-label">{{ $t('overview.balance') }}</div>
        <div class="stat-value mono">${{ overview.balance?.toFixed(2) ?? '—' }}</div>
      </div>
      <div class="stat-card">
        <div class="stat-label">{{ $t('overview.wallet') }}</div>
        <div class="stat-value mono addr">{{ overview.wallet || '—' }}</div>
      </div>
    </div>

    <h3 class="section-title">{{ $t('overview.subsystems') }}</h3>
    <div class="subsystem-list">
      <div
        v-for="s in overview.subsystems"
        :key="s.name"
        class="subsystem-row"
      >
        <span class="sub-dot" :class="s.active ? 'sub-dot--on' : 'sub-dot--off'" />
        <span class="sub-name">{{ s.name }}</span>
        <span class="sub-badge" :class="s.active ? 'badge--ok' : 'badge--off'">
          {{ s.active ? $t('overview.active') : $t('overview.inactive') }}
        </span>
      </div>
      <div v-if="!overview.subsystems?.length" class="empty">{{ $t('common.loading') }}</div>
    </div>
  </div>
</template>

<script setup>
import { onMounted } from 'vue'
import { storeToRefs } from 'pinia'
import { useAppStore } from '@/stores/app'
import { useApi } from '@/composables/useApi'

const app = useAppStore()
const { overview } = storeToRefs(app)
const api = useApi()

onMounted(async () => {
  try { app.overview = await api.getOverview() } catch {}
})
</script>

<style scoped>
.view { display: flex; flex-direction: column; gap: 1.5rem; }
.view-title { font-size: 1.4rem; font-weight: 700; color: var(--text-primary); }

.stat-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
  gap: 1rem;
}

.stat-card {
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  padding: 1.25rem 1.5rem;
}

.stat-label {
  font-size: 0.75rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.06em;
  color: var(--text-secondary);
  margin-bottom: 0.5rem;
}

.stat-value {
  font-size: 1.6rem;
  font-weight: 700;
  color: var(--accent);
}
.stat-value.mono { font-family: var(--font-mono); }
.stat-value.addr { font-size: 0.85rem; word-break: break-all; }

.section-title {
  font-size: 0.9rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.06em;
  color: var(--text-secondary);
}

.subsystem-list {
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  overflow: hidden;
}

.subsystem-row {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.75rem 1rem;
  border-bottom: 1px solid var(--border);
  transition: background var(--transition);
}
.subsystem-row:last-child { border-bottom: none; }
.subsystem-row:hover { background: var(--bg-hover); }

.sub-dot {
  width: 8px; height: 8px; border-radius: 50%; flex-shrink: 0;
}
.sub-dot--on  { background: var(--success); box-shadow: 0 0 6px var(--success); }
.sub-dot--off { background: var(--text-muted); }

.sub-name { flex: 1; font-size: 0.9rem; color: var(--text-primary); }

.sub-badge {
  font-size: 0.72rem;
  font-weight: 600;
  padding: 0.2rem 0.5rem;
  border-radius: 999px;
}
.badge--ok  { background: rgba(63,185,80,0.15); color: var(--success); }
.badge--off { background: var(--badge-bg); color: var(--text-muted); }

.empty { padding: 1rem; color: var(--text-muted); font-size: 0.875rem; text-align: center; }
</style>
