import axios from 'axios'

export function useApi() {
  async function getOverview()   { return (await axios.get('/api/v1/overview')).data }
  async function getOrders()     { return (await axios.get('/api/v1/orders')).data }
  async function getPositions()  { return (await axios.get('/api/v1/positions')).data }
  async function getLogs()       { return (await axios.get('/api/v1/logs')).data }
  async function getCopytrading(){ return (await axios.get('/api/v1/copytrading')).data }
  async function getSettings()   { return (await axios.get('/api/v1/settings')).data }

  async function cancelOrder(id) {
    return (await axios.delete(`/api/v1/orders/${id}`)).data
  }
  async function cancelAll() {
    return (await axios.delete('/api/v1/orders')).data
  }
  async function postSettings(key, value) {
    return (await axios.post('/api/v1/settings', { key, value })).data
  }
  async function addTrader(addr, label, allocPct) {
    return (await axios.post('/api/v1/copytrading', { address: addr, label, allocation_pct: allocPct })).data
  }
  async function removeTrader(addr) {
    return (await axios.delete(`/api/v1/copytrading/traders/${addr}`)).data
  }
  async function toggleTrader(addr) {
    return (await axios.patch(`/api/v1/copytrading/traders/${addr}`)).data
  }

  return {
    getOverview, getOrders, getPositions, getLogs,
    getCopytrading, getSettings, cancelOrder, cancelAll,
    postSettings, addTrader, removeTrader, toggleTrader
  }
}
