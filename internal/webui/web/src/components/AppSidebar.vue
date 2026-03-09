<template>
  <nav class="sidebar">
    <!-- Brand -->
    <div class="sidebar-brand">
      <span class="brand-glyph">◈</span>
      <div class="brand-text">
        <span class="brand-name">POLYTRADE</span>
        <span class="brand-sub">NEXUS TERMINAL</span>
      </div>
    </div>

    <!-- Nav items -->
    <div class="nav-section">
      <RouterLink
        v-for="item in navItems"
        :key="item.to"
        :to="item.to"
        class="nav-item"
        active-class="nav-item--active"
      >
        <span class="nav-icon">{{ item.icon }}</span>
        <span class="nav-label">{{ $t(item.label) }}</span>
        <span v-if="item.badge" class="nav-badge">{{ item.badge }}</span>
      </RouterLink>
    </div>

    <!-- System status at bottom -->
    <div class="sidebar-footer">
      <div class="sys-row">
        <span class="sys-dot status-dot" :class="connected ? 'status-dot--on' : 'status-dot--off'" />
        <span class="sys-text">{{ connected ? 'CONNECTED' : 'OFFLINE' }}</span>
      </div>
    </div>
  </nav>
</template>

<script setup>
import { computed } from 'vue'
import { storeToRefs } from 'pinia'
import { useAppStore } from '@/stores/app'

const { connected } = storeToRefs(useAppStore())

const navItems = [
  { to: '/overview',    label: 'nav.overview',    icon: '⊞' },
  { to: '/orders',      label: 'nav.orders',      icon: '≡' },
  { to: '/positions',   label: 'nav.positions',   icon: '▣' },
  { to: '/copytrading', label: 'nav.copytrading', icon: '⇌' },
  { to: '/wallets',     label: 'nav.wallets',     icon: '◎' },
  { to: '/markets',     label: 'nav.markets',     icon: '⊛' },
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
  overflow-y: auto;
  position: relative;
}

/* Subtle scanline effect on sidebar */
.sidebar::before {
  content: '';
  position: absolute;
  inset: 0;
  background: repeating-linear-gradient(
    0deg,
    transparent,
    transparent 3px,
    rgba(0, 200, 255, 0.012) 3px,
    rgba(0, 200, 255, 0.012) 6px
  );
  pointer-events: none;
  z-index: 0;
}

/* Brand */
.sidebar-brand {
  display: flex;
  align-items: center;
  gap: 0.65rem;
  padding: 1rem 1rem 0.85rem;
  border-bottom: 1px solid var(--border);
  flex-shrink: 0;
  position: relative;
  z-index: 1;
}

.brand-glyph {
  font-size: 1.1rem;
  color: var(--accent);
  text-shadow: 0 0 14px rgba(0, 200, 255, 0.60);
  flex-shrink: 0;
}

.brand-text {
  display: flex;
  flex-direction: column;
  gap: 0.05rem;
}

.brand-name {
  font-size: 0.86rem;
  font-weight: 700;
  letter-spacing: 0.18em;
  color: var(--accent-bright);
  line-height: 1;
}

.brand-sub {
  font-size: 0.92rem;
  letter-spacing: 0.12em;
  color: var(--text-secondary);
  line-height: 1;
}

/* Nav section */
.nav-section {
  flex: 1;
  padding: 0.5rem 0;
  display: flex;
  flex-direction: column;
  position: relative;
  z-index: 1;
}

.nav-item {
  display: flex;
  align-items: center;
  gap: 0.6rem;
  padding: 0.55rem 1rem;
  color: var(--text-secondary);
  text-decoration: none;
  font-size: 0.90rem;
  font-weight: 500;
  letter-spacing: 0.03em;
  border-left: 2px solid transparent;
  transition: color var(--transition), background var(--transition), border-color var(--transition);
  white-space: nowrap;
  position: relative;
}

.nav-item:hover {
  color: var(--text-primary);
  background: rgba(0, 200, 255, 0.04);
  border-left-color: rgba(0, 200, 255, 0.40);
}

.nav-item--active {
  color: var(--accent-bright);
  background: rgba(0, 200, 255, 0.08);
  border-left-color: var(--accent);
}

.nav-item--active::after {
  content: '';
  position: absolute;
  right: 0;
  top: 50%;
  transform: translateY(-50%);
  width: 3px;
  height: 60%;
  background: var(--accent);
  border-radius: 2px 0 0 2px;
  opacity: 0.4;
}

.nav-icon {
  font-size: 0.96rem;
  width: 1rem;
  text-align: center;
  flex-shrink: 0;
  opacity: 0.8;
}

.nav-item--active .nav-icon { opacity: 1; }

.nav-badge {
  margin-left: auto;
  font-size: 0.86rem;
  background: var(--accent-dim);
  color: var(--accent);
  padding: 0.08rem 0.35rem;
  border-radius: 1px;
  border: 1px solid rgba(0,200,255,0.20);
  font-weight: 600;
}

/* Footer */
.sidebar-footer {
  padding: 0.65rem 1rem;
  border-top: 1px solid var(--border);
  flex-shrink: 0;
  position: relative;
  z-index: 1;
}

.sys-row {
  display: flex;
  align-items: center;
  gap: 0.45rem;
}

.sys-dot {
  width: 6px;
  height: 6px;
}

.sys-text {
  font-size: 0.86rem;
  letter-spacing: 0.12em;
  color: var(--text-muted);
  font-weight: 600;
}
</style>
