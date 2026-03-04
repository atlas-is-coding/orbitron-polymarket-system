<template>
  <nav class="sidebar">
    <RouterLink
      v-for="item in navItems"
      :key="item.to"
      :to="item.to"
      class="nav-item"
      active-class="nav-item--active"
    >
      <span class="nav-icon">{{ item.icon }}</span>
      <span class="nav-label">{{ $t(item.label) }}</span>
    </RouterLink>
  </nav>
</template>

<script setup>
const navItems = [
  { to: '/overview',    label: 'nav.overview',    icon: '⊞' },
  { to: '/orders',      label: 'nav.orders',      icon: '≡' },
  { to: '/positions',   label: 'nav.positions',   icon: '▣' },
  { to: '/copytrading', label: 'nav.copytrading', icon: '⇌' },
  { to: '/wallets',     label: 'nav.wallets',     icon: '◎' },
  { to: '/logs',        label: 'nav.logs',        icon: '▦' },
  { to: '/settings',    label: 'nav.settings',    icon: '✦' },
]
</script>

<style scoped>
.sidebar {
  width: var(--sidebar-w);
  flex-shrink: 0;
  background: var(--bg-sidebar);
  border-right: 1px solid var(--border);
  display: flex;
  flex-direction: column;
  padding: 0.5rem 0;
  overflow-y: auto;
}

.nav-item {
  display: flex;
  align-items: center;
  gap: 0.6rem;
  padding: 0.55rem 1rem;
  color: var(--text-secondary);
  text-decoration: none;
  font-size: 0.8rem;
  border-left: 2px solid transparent;
  transition: color var(--transition), background var(--transition), border-color var(--transition);
  white-space: nowrap;
  letter-spacing: 0.02em;
}
.nav-item:hover {
  color: var(--text-primary);
  background: var(--bg-hover);
}
.nav-item--active {
  color: var(--accent-bright);
  background: var(--accent-dim);
  border-left-color: var(--accent);
}

.nav-icon  { font-size: 0.85rem; width: 1rem; text-align: center; flex-shrink: 0; }
.nav-label { }
</style>
