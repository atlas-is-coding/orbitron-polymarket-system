import { defineStore } from 'pinia'
import { ref } from 'vue'
import { useMarketsStore } from '@/stores/markets'
import { useHealthStore } from '@/stores/health'

export const useAppStore = defineStore('app', () => {
  const overview = ref({ balance: 0, pnl: 0, wallet: '', subsystems: [], orders: [], positions: [], strategies: [] })
  const orders = ref([])
  const positions = ref([])
  const strategies = ref([])
  const logs = ref([])
  const copytrading = ref({ enabled: false, traders: [] })
  const settings = ref({})
  const connected = ref(false)
  const walletsMap = ref({})
  const settingsStale = ref(false)
  const copyTrades = ref([])   // recent copy trades feed (last 30)

  const toasts = ref([])
  let _toastId = 0

  function toast(msg, type = 'info', duration = 3000) {
    const id = ++_toastId
    toasts.value.push({ id, msg, type })
    setTimeout(() => { toasts.value = toasts.value.filter(t => t.id !== id) }, duration)
  }

  function applyEvent(event) {
    switch (event.type) {
      case 'initial_state':
        overview.value = {
          balance: event.data.balance,
          pnl: event.data.pnl,
          wallet: event.data.wallet,
          subsystems: event.data.subsystems,
        }
        orders.value = event.data.orders || []
        positions.value = event.data.positions || []
        strategies.value = event.data.strategies || []
        logs.value = (event.data.logs || []).reverse()
        if (event.data.wallets) {
          const m = {}
          event.data.wallets.forEach(w => { m[w.id] = w })
          walletsMap.value = m
        }
        break
      case 'overview':
        overview.value = event.data;
        if (event.data.orders) orders.value = event.data.orders;
        if (event.data.positions) positions.value = event.data.positions;
        if (event.data.strategies) strategies.value = event.data.strategies;
        break
      case 'balance':
        overview.value = { ...overview.value, balance: event.data.usdc }
        break
      case 'subsystem': {
        const subs = overview.value.subsystems || []
        const idx = subs.findIndex(s => s.name === event.data.name)
        if (idx >= 0) {
          const next = [...subs]
          next[idx] = { name: event.data.name, active: event.data.active }
          overview.value = { ...overview.value, subsystems: next }
        } else {
          overview.value = { ...overview.value, subsystems: [...subs, { name: event.data.name, active: event.data.active }] }
        }
        break
      }
      case 'orders':      orders.value = event.data; break
      case 'positions':   positions.value = event.data; break
      case 'strategies':  strategies.value = event.data; break
      case 'log':         logs.value = [event.data, ...logs.value].slice(0, 200); break
      case 'copytrading': copytrading.value = event.data; break
      case 'settings':    settings.value = event.data; break
      case 'wallet_added':
        walletsMap.value = { ...walletsMap.value, [event.data.id]: event.data }
        if (event.data.primary) {
          overview.value = { ...overview.value, wallet: event.data.id }
        }
        break
      case 'wallet_removed': {
        const m = { ...walletsMap.value }
        delete m[event.data.id]
        walletsMap.value = m
        break
      }
      case 'wallet_changed': {
        const existing = walletsMap.value[event.data.id]
        if (existing) {
          const updated = { ...existing, enabled: event.data.enabled }
          if ('primary' in event.data) updated.primary = event.data.primary
          // If this wallet became primary, clear primary from all others
          if (event.data.primary) {
            const next = {}
            for (const [id, w] of Object.entries(walletsMap.value)) {
              next[id] = id === event.data.id ? updated : { ...w, primary: false }
            }
            walletsMap.value = next
            overview.value = { ...overview.value, wallet: event.data.id }
          } else {
            walletsMap.value = { ...walletsMap.value, [event.data.id]: updated }
          }
        }
        break
      }
      case 'wallet_stats':
        walletsMap.value = { ...walletsMap.value, [event.data.id]: event.data }
        // Update aggregate balance/pnl in overview too
        {
          const all = Object.values(walletsMap.value)
          const totalBal = all.reduce((s, w) => s + (w.balance_usd || 0), 0)
          const totalPnL = all.reduce((s, w) => s + (w.pnl_usd || 0), 0)
          const primary = all.find(w => w.primary)
          overview.value = {
            ...overview.value,
            balance: totalBal,
            pnl: totalPnL,
            wallet: primary ? primary.id : overview.value.wallet
          }
        }
        break
      case 'market_alert':
        toast(
          `🔔 ${event.data.question || event.data.conditionId}: price went ${event.data.direction} ${Number(event.data.threshold).toFixed(3)} (now ${Number(event.data.currentPrice).toFixed(3)})`,
          'info',
          8000
        )
        break
      case 'markets_updated':
        // Server finished a poll cycle — refresh the markets list in the background.
        useMarketsStore().fetchMarkets()
        break
      case 'markets_loading':
        useMarketsStore().onMarketsLoading(event.data)
        break
      case 'markets_ready':
        useMarketsStore().onMarketsReady()
        break
      case 'config_reloaded':
        settingsStale.value = true
        break
      case 'copy_trade':
        copyTrades.value = [event.data.line, ...copyTrades.value].slice(0, 30)
        break
      case 'health_updated':
        useHealthStore().applyHealthUpdate(event.data)
        break
    }
  }

  return {
    overview, orders, positions, strategies, logs, copytrading, settings,
    connected, walletsMap, settingsStale, copyTrades,
    applyEvent, toasts, toast
  }
})
