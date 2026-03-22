<template>
  <aside class="sidebar" :class="{ collapsed: ui.sidebarCollapsed }">
    <div class="sidebar-logo">
      <span class="logo-icon">◈</span>
      <span class="logo-text" v-if="!ui.sidebarCollapsed">PolyTrade</span>
    </div>

    <nav class="sidebar-nav">
      <template v-for="section in navSections" :key="section.label">
        <div class="nav-section" v-if="!ui.sidebarCollapsed">
          <span class="nav-section-label">{{ section.label }}</span>
        </div>
        <div v-else class="nav-section-divider"></div>
        <router-link
          v-for="item in section.items"
          :key="item.to"
          :to="item.to"
          class="nav-item"
          :class="{ 'nav-item--exact': item.exact }"
          :title="ui.sidebarCollapsed ? item.label : undefined"
          :exact-active-class="item.exact ? 'nav-item--active' : ''"
          :active-class="!item.exact ? 'nav-item--active' : ''"
        >
          <span class="nav-icon">{{ item.icon }}</span>
          <span class="nav-label" v-if="!ui.sidebarCollapsed">{{ item.label }}</span>
          <span
            v-if="item.badge != null && item.badge !== 0 && !ui.sidebarCollapsed"
            class="nav-badge"
            :class="item.badgeClass"
          >{{ item.badge }}</span>
        </router-link>
      </template>
    </nav>

    <div class="sidebar-footer">
      <div class="wallet-info" v-if="!ui.sidebarCollapsed">
        <span class="wallet-dot" :class="app.connected ? 'online' : 'offline'"></span>
        <div class="wallet-details">
          <span class="wallet-addr">{{ shortAddr }}</span>
          <span class="wallet-balance">{{ balance }}</span>
        </div>
      </div>
      <span v-else class="wallet-dot-solo" :class="app.connected ? 'online' : 'offline'"></span>
      <button class="collapse-btn" @click="ui.toggleSidebar()" :title="ui.sidebarCollapsed ? 'Expand' : 'Collapse'">
        {{ ui.sidebarCollapsed ? '▶' : '◀' }}
      </button>
    </div>
  </aside>
</template>

<script setup>
import { computed } from 'vue'
import { useUiStore } from '../../stores/ui.js'
import { useAppStore } from '../../stores/app.js'

const ui = useUiStore()
const app = useAppStore()

const navSections = computed(() => [
  {
    label: 'MAIN',
    items: [
      { to: '/', icon: '⊞', label: 'Overview', exact: true },
      { to: '/markets', icon: '◎', label: 'Markets' },
      {
        to: '/orders',
        icon: '≡',
        label: 'Orders',
        badge: app.orders?.filter(o => o.status === 'OPEN').length || null,
        badgeClass: 'badge-neutral',
      },
      {
        to: '/orders?tab=positions',
        icon: '▲',
        label: 'Positions',
        badge: app.positions?.length || null,
        badgeClass: 'badge-success',
      },
      { to: '/wallets', icon: '◈', label: 'Wallets' },
    ],
  },
  {
    label: 'AUTOMATION',
    items: [
      {
        to: '/strategies',
        icon: '⚡',
        label: 'Strategies',
        badge: app.strategies?.filter(s => s.enabled).length || null,
        badgeClass: 'badge-success',
      },
      { to: '/copytrading', icon: '⟳', label: 'Copytrading' },
    ],
  },
  {
    label: 'SYSTEM',
    items: [
      { to: '/logs', icon: '▤', label: 'Logs' },
      { to: '/settings', icon: '⚙', label: 'Settings' },
    ],
  },
])

const shortAddr = computed(() => {
  const addr = app.overview?.wallet?.address || ''
  return addr ? addr.slice(0, 6) + '...' + addr.slice(-4) : '—'
})

const balance = computed(() => {
  const b = app.overview?.wallet?.balance
  return b != null ? `$${Number(b).toFixed(2)}` : '—'
})
</script>

<style scoped>
.sidebar {
  width: var(--sidebar-w, 220px);
  min-height: 100vh;
  background: var(--bg-panel);
  border-right: 1px solid var(--border);
  display: flex;
  flex-direction: column;
  transition: width 0.2s ease;
  flex-shrink: 0;
  overflow: hidden;
}
.sidebar.collapsed { width: var(--sidebar-w-collapsed, 56px); }

/* Responsive: auto-collapse on small screens */
@media (max-width: 768px) {
  .sidebar { width: var(--sidebar-w-collapsed, 56px) !important; }
  .nav-label,
  .logo-text,
  .nav-section,
  .wallet-info { display: none !important; }
  .wallet-dot-solo { display: inline-block !important; }
}

.sidebar-logo {
  height: var(--header-h-new, 54px);
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 0 14px;
  border-bottom: 1px solid var(--border);
  flex-shrink: 0;
}
.logo-icon { font-size: 20px; color: var(--accent); }
.logo-text { font-size: 15px; font-weight: 700; color: var(--fg); letter-spacing: 0.04em; white-space: nowrap; }

.sidebar-nav { flex: 1; padding: 8px 0; overflow-y: auto; overflow-x: hidden; }

.nav-section { padding: 12px 14px 4px; }
.nav-section-label { font-size: 9px; letter-spacing: 0.12em; color: var(--fg-muted); text-transform: uppercase; white-space: nowrap; }
.nav-section-divider { height: 8px; }

.nav-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 0 12px;
  height: 38px;
  color: var(--fg-muted);
  text-decoration: none;
  border-radius: 6px;
  margin: 1px 6px;
  font-size: var(--font-size-nav, 13px);
  transition: background 0.15s, color 0.15s;
  white-space: nowrap;
  overflow: hidden;
}
.nav-item:hover { background: var(--bg-hover); color: var(--fg); }
.nav-item--active {
  background: color-mix(in srgb, var(--accent) 15%, transparent);
  color: var(--accent-bright);
}

.nav-icon {
  width: var(--nav-icon-box, 32px);
  height: var(--nav-icon-box, 32px);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: var(--font-size-nav-icon, 16px);
  flex-shrink: 0;
}
.nav-label { flex: 1; overflow: hidden; text-overflow: ellipsis; }
.nav-badge {
  font-size: 10px;
  padding: 1px 6px;
  border-radius: 10px;
  flex-shrink: 0;
}
.nav-badge.badge-success { background: var(--success); color: #000; }
.nav-badge.badge-neutral { background: var(--accent); color: #fff; }
.nav-badge.badge-danger  { background: var(--danger); color: #fff; }

.sidebar-footer {
  border-top: 1px solid var(--border);
  padding: 10px 12px;
  display: flex;
  align-items: center;
  gap: 8px;
  flex-shrink: 0;
}
.wallet-info { display: flex; align-items: center; gap: 8px; flex: 1; min-width: 0; }
.wallet-dot,
.wallet-dot-solo {
  width: 8px; height: 8px; border-radius: 50%; flex-shrink: 0;
}
.wallet-dot-solo { display: none; }
.wallet-dot.online,
.wallet-dot-solo.online { background: var(--success); animation: ws-pulse 2s infinite; }
.wallet-dot.offline,
.wallet-dot-solo.offline { background: var(--fg-muted); }
@keyframes ws-pulse { 0%,100%{ opacity:1 } 50%{ opacity:0.4 } }

.wallet-details { display: flex; flex-direction: column; min-width: 0; }
.wallet-addr { font-size: 11px; color: var(--fg); font-family: monospace; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.wallet-balance { font-size: 10px; color: var(--fg-muted); }

.collapse-btn {
  background: none; border: 1px solid var(--border); color: var(--fg-muted);
  border-radius: 4px; width: 26px; height: 26px; cursor: pointer; font-size: 10px;
  display: flex; align-items: center; justify-content: center; flex-shrink: 0;
  margin-left: auto;
}
.collapse-btn:hover { background: var(--bg-hover); color: var(--fg); }
</style>
