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

  function applyEvent(event) {
    switch (event.type) {
      case 'overview':    overview.value = event.data; break
      case 'orders':      orders.value = event.data; break
      case 'positions':   positions.value = event.data; break
      case 'log':         logs.value = [event.data, ...logs.value].slice(0, 200); break
      case 'copytrading': copytrading.value = event.data; break
      case 'settings':    settings.value = event.data; break
    }
  }

  return { overview, orders, positions, logs, copytrading, settings, connected, applyEvent }
})
