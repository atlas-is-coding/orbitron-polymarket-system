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
        console.log('[WS] initial_state received:', event.data)
        if (event.data.wallets) {
          const m = {}
          event.data.wallets.forEach(w => {
            const id = w.id || w.ID
            m[id] = {
              id: id,
              address: w.address || w.Address || '',
              label: w.label || w.Label || '',
              enabled: w.enabled !== undefined ? w.enabled : w.Enabled,
              primary: w.primary !== undefined ? w.primary : w.Primary,
              balance_usd: w.balance_usd !== undefined ? w.balance_usd : w.BalanceUSD,
              pnl_usd: w.pnl_usd !== undefined ? w.pnl_usd : w.PnLUSD,
              open_orders: w.open_orders !== undefined ? w.open_orders : w.OpenOrders,
              total_trades: w.total_trades !== undefined ? w.total_trades : w.TotalTrades,
            }
          })
          walletsMap.value = m
        }

        orders.value = event.data.orders || []
        positions.value = event.data.positions || []
        
        // Map strategies for consistency
        {
          const mapped = (event.data.strategies || []).map(s => ({
            name: s.name || s.Name,
            status: s.status || s.Status,
            wallet_id: s.wallet_id || s.WalletID,
            wallet_label: s.wallet_label || s.WalletLabel || s.Label || '',
            wallet_address: s.wallet_address || s.WalletAddress || '',
            details: s.details || s.Details || '',
          }))
          strategies.value = mapped
        }

        {
          const all = Object.values(walletsMap.value)
          const primary = all.find(w => w.primary) || all[0]
          overview.value = {
            ...overview.value,
            balance: event.data.balance,
            pnl: event.data.pnl,
            wallet: primary ? primary.id : (event.data.wallet || ''),
            wallet_address: primary ? primary.address : '',
            subsystems: event.data.subsystems || [],
            orders: orders.value,
            positions: positions.value,
            strategies: strategies.value
          }
        }
        logs.value = (event.data.logs || []).reverse()
        break

      case 'overview':
        {
          const data = event.data;
          const mappedStrategies = (data.strategies || []).map(s => ({
            name: s.name || s.Name,
            status: s.status || s.Status,
            wallet_id: s.wallet_id || s.WalletID,
            wallet_label: s.wallet_label || s.WalletLabel || s.Label || '',
            wallet_address: s.wallet_address || s.WalletAddress || '',
            details: s.details || s.Details || '',
          }));
          strategies.value = mappedStrategies
          
          overview.value = {
            ...overview.value,
            balance: data.balance ?? overview.value.balance,
            pnl: data.pnl ?? overview.value.pnl,
            subsystems: data.subsystems || overview.value.subsystems,
            wallet: data.wallet || overview.value.wallet,
            wallet_address: data.wallet_address || overview.value.wallet_address,
            orders: data.orders || overview.value.orders,
            positions: data.positions || overview.value.positions,
            strategies: mappedStrategies
          };
          
          if (data.orders) orders.value = data.orders;
          if (data.positions) positions.value = data.positions;
        }
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

      case 'strategies': {
        const mapped = (event.data || []).map(s => ({
          name: s.name || s.Name,
          status: s.status || s.Status,
          wallet_id: s.wallet_id || s.WalletID,
          wallet_label: s.wallet_label || s.WalletLabel || s.Label || '',
          wallet_address: s.wallet_address || s.WalletAddress || '',
          details: s.details || s.Details || '',
        }))
        strategies.value = mapped
        overview.value = { ...overview.value, strategies: mapped }
        break
      }

      case 'log':         logs.value = [event.data, ...logs.value].slice(0, 200); break
      case 'copytrading': copytrading.value = event.data; break
      case 'settings':    settings.value = event.data; break

      case 'wallet_added': {
        const w = event.data
        const id = w.id || w.ID
        const mapped = {
          id: id,
          address: w.address || w.Address || '',
          label: w.label || w.Label || '',
          enabled: w.enabled !== undefined ? w.enabled : w.Enabled,
          primary: w.primary !== undefined ? w.primary : w.Primary,
          balance_usd: 0,
          pnl_usd: 0,
          open_orders: 0,
          total_trades: 0,
        }
        walletsMap.value = { ...walletsMap.value, [id]: mapped }
        if (mapped.primary) {
          overview.value = { ...overview.value, wallet: id, wallet_address: mapped.address }
        }
        break
      }

      case 'wallet_removed': {
        const id = event.data.id || event.data.ID
        const m = { ...walletsMap.value }
        delete m[id]
        walletsMap.value = m
        break
      }

      case 'wallet_changed': {
        const id = event.data.id || event.data.ID
        const existing = walletsMap.value[id]
        if (existing) {
          const enabled = event.data.enabled !== undefined ? event.data.enabled : event.data.Enabled
          const primary = event.data.primary !== undefined ? event.data.primary : event.data.Primary
          const updated = { ...existing, enabled: enabled }
          if (primary !== undefined) updated.primary = primary
          // If this wallet became primary, clear primary from all others
          if (primary) {
            const next = {}
            for (const [wid, w] of Object.entries(walletsMap.value)) {
              next[wid] = wid === id ? updated : { ...w, primary: false }
            }
            walletsMap.value = next
            overview.value = { ...overview.value, wallet: id, wallet_address: updated.address }
          } else {
            walletsMap.value = { ...walletsMap.value, [id]: updated }
          }
        }
        break
      }

      case 'wallet_stats': {
        const w = event.data
        const id = w.id || w.ID
        const mapped = {
          id: id,
          address: w.address || w.Address || '',
          label: w.label || w.Label || '',
          enabled: w.enabled !== undefined ? w.enabled : w.Enabled,
          primary: w.primary !== undefined ? w.primary : w.Primary,
          balance_usd: w.balance_usd !== undefined ? w.balance_usd : w.BalanceUSD,
          pnl_usd: w.pnl_usd !== undefined ? w.pnl_usd : w.PnLUSD,
          open_orders: w.open_orders !== undefined ? w.open_orders : w.OpenOrders,
          total_trades: w.total_trades !== undefined ? w.total_trades : w.TotalTrades,
        }
        walletsMap.value = { ...walletsMap.value, [id]: mapped }
        // Update aggregate balance/pnl in overview too
        {
          const all = Object.values(walletsMap.value)
          const totalBal = all.reduce((s, wl) => s + (wl.balance_usd || 0), 0)
          const totalPnL = all.reduce((s, wl) => s + (wl.pnl_usd || 0), 0)
          const primary = all.find(wl => wl.primary) || all[0]
          overview.value = {
            ...overview.value,
            balance: totalBal,
            pnl: totalPnL,
            wallet: primary ? primary.id : overview.value.wallet,
            wallet_address: primary ? primary.address : (overview.value.wallet_address || '')
          }
        }
        break
      }

      case 'market_alert':
        toast(
          `🔔 ${event.data.question || event.data.conditionId}: price went ${event.data.direction} ${Number(event.data.threshold).toFixed(3)} (now ${Number(event.data.currentPrice).toFixed(3)})`,
          'info',
          8000
        )
        break

      case 'markets_updated':
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
