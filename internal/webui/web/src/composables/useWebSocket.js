import { onUnmounted } from 'vue'
import { useAppStore } from '@/stores/app'
import { useAuthStore } from '@/stores/auth'

let ws = null
let reconnectTimer = null

export function useWebSocket() {
  const app = useAppStore()
  const auth = useAuthStore()

  function connect() {
    if (ws && ws.readyState < 2) return

    const proto = location.protocol === 'https:' ? 'wss' : 'ws'
    const url = `${proto}://${location.host}/ws?token=${auth.token}`
    ws = new WebSocket(url)

    ws.onopen = () => { app.connected = true }
    ws.onclose = () => {
      app.connected = false
      reconnectTimer = setTimeout(connect, 3000)
    }
    ws.onerror = () => ws.close()
    ws.onmessage = (e) => {
      try { app.applyEvent(JSON.parse(e.data)) } catch {}
    }
  }

  function disconnect() {
    clearTimeout(reconnectTimer)
    if (ws) { ws.close(); ws = null }
  }

  onUnmounted(disconnect)

  return { connect, disconnect }
}
