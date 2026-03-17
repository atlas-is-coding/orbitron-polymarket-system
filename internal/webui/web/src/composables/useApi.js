import axios from 'axios'

export function useApi() {
  async function getOverview()   { return (await axios.get('/api/v1/overview')).data }
  async function getOrders()     { return (await axios.get('/api/v1/orders')).data }
  async function getPositions()  { return (await axios.get('/api/v1/positions')).data }
  async function getLogs()       { return (await axios.get('/api/v1/logs')).data }
  async function getStrategies() { return (await axios.get('/api/v1/strategies')).data }
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
    return (await axios.patch(`/api/v1/copytrading/traders/${addr}/toggle`)).data
  }
  async function editTrader(addr, label, allocPct, maxPositionUsd) {
    return (await axios.patch(`/api/v1/copytrading/traders/${addr}/edit`, {
      label, alloc_pct: allocPct, max_position_usd: maxPositionUsd
    })).data
  }

  async function getWallets() {
    return (await axios.get('/api/v1/wallets')).data
  }
  async function toggleWallet(id, enabled) {
    return (await axios.post(`/api/v1/wallets/${id}/toggle`, { enabled })).data
  }
  async function renameWallet(id, label) {
    return (await axios.patch(`/api/v1/wallets/${id}`, { label })).data
  }
  async function removeWallet(id) {
    return (await axios.delete(`/api/v1/wallets/${id}`)).data
  }
  async function addWallet(privateKey) {
    return (await axios.post('/api/v1/wallets', { private_key: privateKey })).data
  }

  async function placeOrder(tokenId, side, orderType, price, sizeUsd, walletId) {
    return (await axios.post('/api/v1/orders', {
      token_id: tokenId, side, order_type: orderType,
      price, size_usd: sizeUsd, wallet_id: walletId
    })).data
  }

  async function getMarkets(tag) {
    const params = tag ? { tag } : {}
    return (await axios.get('/api/v1/markets', { params })).data
  }
  async function getMarketTags() {
    return (await axios.get('/api/v1/markets/tags')).data
  }

  async function startStrategy(key, walletIds = []) {
    return (await axios.post(`/api/v1/strategies/${key}/start`, { wallet_ids: walletIds })).data
  }
  async function stopStrategy(key) {
    return (await axios.post(`/api/v1/strategies/${key}/stop`)).data
  }

  return {
    getOverview, getOrders, getPositions, getLogs,
    getCopytrading, getSettings, cancelOrder, cancelAll,
    postSettings, addTrader, removeTrader, toggleTrader,
    getWallets, toggleWallet, renameWallet, removeWallet, addWallet,
    editTrader, placeOrder, getMarkets, getMarketTags,
    startStrategy, stopStrategy
  }
}
