import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const routes = [
  {
    path: '/login',
    name: 'login',
    component: () => import('@/views/LoginView.vue'),
    meta: { public: true }
  },
  {
    path: '/',
    name: 'overview',
    component: () => import('@/views/OverviewView.vue'),
  },
  {
    path: '/overview',
    redirect: '/'
  },
  {
    path: '/orders',
    name: 'orders',
    component: () => import('@/views/OrdersView.vue')
  },
  {
    path: '/positions',
    redirect: '/orders?tab=positions'
  },
  {
    path: '/copytrading',
    name: 'copytrading',
    component: () => import('@/views/CopytradingView.vue')
  },
  {
    path: '/markets',
    name: 'markets',
    component: () => import('@/views/MarketsView.vue')
  },
  {
    path: '/logs',
    name: 'logs',
    component: () => import('@/views/LogsView.vue')
  },
  {
    path: '/wallets',
    name: 'wallets',
    component: () => import('@/views/WalletsView.vue')
  },
  {
    path: '/settings',
    name: 'settings',
    component: () => import('@/views/SettingsView.vue')
  },
  {
    path: '/strategies',
    name: 'strategies',
    component: () => import('@/views/StrategiesView.vue')
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

router.beforeEach((to) => {
  const auth = useAuthStore()
  if (!to.meta.public && !auth.isAuthenticated) {
    return { name: 'login' }
  }
})

export default router
