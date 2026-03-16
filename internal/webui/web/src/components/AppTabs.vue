<template>
  <nav class="app-tabs">
    <RouterLink
      v-for="tab in tabs"
      :key="tab.to"
      :to="tab.to"
      class="tab"
      active-class="tab--active"
    >
      {{ $t(tab.label) }}
    </RouterLink>
  </nav>
</template>

<script setup>
const tabs = [
  { to: '/overview',    label: 'nav.overview' },
  { to: '/orders',      label: 'nav.orders' },
  { to: '/positions',   label: 'nav.positions' },
  { to: '/copytrading', label: 'nav.copytrading' },
  { to: '/wallets',     label: 'nav.wallets' },
  { to: '/markets',     label: 'nav.markets' },
  { to: '/strategies',  label: 'nav.strategies' },
  { to: '/logs',        label: 'nav.logs' },
  { to: '/settings',    label: 'nav.settings' },
]
</script>

<style scoped>
.app-tabs {
  display: flex;
  align-items: center;
  height: var(--tabs-h);
  background: rgba(255,255,255,0.02);
  border-bottom: 1px solid var(--border);
  padding: 0 1rem;
  gap: 0;
  overflow-x: auto;
  scrollbar-width: none;
  flex-shrink: 0;
}
.app-tabs::-webkit-scrollbar { display: none; }

.tab {
  display: flex;
  align-items: center;
  height: 100%;
  padding: 0 0.85rem;
  color: var(--text-secondary);
  text-decoration: none;
  font-size: 0.82rem;
  font-weight: 600;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  white-space: nowrap;
  border-bottom: 2px solid transparent;
  transition: color var(--transition), border-color var(--transition);
  position: relative;
  top: 1px;
}

.tab:hover {
  color: var(--text-primary);
  border-bottom-color: rgba(124,58,237,0.35);
}

.tab--active {
  color: var(--accent-bright);
  border-bottom-color: var(--accent);
}
</style>
