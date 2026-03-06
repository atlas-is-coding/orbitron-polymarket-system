import { defineStore } from 'pinia'
import { ref } from 'vue'

export const useAppStore = defineStore('app', () => {
  const overview = ref({ balance: 0, wallet: '', subsystems: [] })
  const orders = ref([])
  const positions = ref([])
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
      case 'overview':    overview.value = event.data; break
      case 'orders':      orders.value = event.data; break
      case 'positions':   positions.value = event.data; break
      case 'log':         logs.value = [event.data, ...logs.value].slice(0, 200); break
      case 'copytrading': copytrading.value = event.data; break
      case 'settings':    settings.value = event.data; break
      case 'wallet_added':
        walletsMap.value = { ...walletsMap.value, [event.data.id]: event.data }
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
          walletsMap.value = { ...walletsMap.value, [event.data.id]: { ...existing, enabled: event.data.enabled } }
        }
        break
      }
      case 'wallet_stats':
        walletsMap.value = { ...walletsMap.value, [event.data.id]: event.data }
        break
      case 'market_alert':
        toast(
          `🔔 ${event.data.question || event.data.conditionId}: price went ${event.data.direction} ${Number(event.data.threshold).toFixed(3)} (now ${Number(event.data.currentPrice).toFixed(3)})`,
          'info',
          8000
        )
        break
      case 'config_reloaded':
        settingsStale.value = true
        break
      case 'copy_trade':
        copyTrades.value = [event.data.line, ...copyTrades.value].slice(0, 30)
        break
    }
  }

  return {
    overview, orders, positions, logs, copytrading, settings,
    connected, walletsMap, settingsStale, copyTrades,
    applyEvent, toasts, toast
  }
})
