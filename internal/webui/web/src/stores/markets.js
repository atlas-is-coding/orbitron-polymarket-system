import { defineStore } from 'pinia'
import axios from 'axios'

export const useMarketsStore = defineStore('markets', {
  state: () => ({
    markets: [],
    trending: [],
    tags: [],
    activeTag: '',
    search: '',
    viewMode: 'trending',         // 'trending' | 'categories'
    selectedMarket: null,
    loading: false,
    error: null,
    totalCount: 0,
    syncing: false,
    loadProgress: { loaded: 0, total: 500 },
    // multi-select batch buy
    selectedMarkets: {},
    batchSide: 'YES',
  }),
  getters: {
    filteredMarkets(state) {
      let list = state.markets
      if (state.search) {
        const q = state.search.toLowerCase()
        list = list.filter(m => m.question?.toLowerCase().includes(q))
      }
      return list
    },
    selectedMarketsArray(state) {
      return state.markets.filter(m => state.selectedMarkets[m.conditionId])
    },
    selectedCount(state) {
      return Object.keys(state.selectedMarkets).length
    },
  },
  actions: {
    async fetchMarkets() {
      this.loading = true
      this.error = null
      try {
        const params = new URLSearchParams()
        if (this.activeTag) params.set('tag', this.activeTag)
        const { data } = await axios.get(`/api/v1/markets?${params}`)
        this.markets = data
      } catch (e) {
        this.error = e.message
      } finally {
        this.loading = false
      }
    },
    async fetchTrending(limit = 50) {
      this.loading = true
      this.error = null
      try {
        const { data } = await axios.get(`/api/v1/markets/trending?limit=${limit}`)
        this.trending = data
      } catch (e) {
        this.error = e.message
      } finally {
        this.loading = false
      }
    },
    async fetchStats() {
      try {
        const { data } = await axios.get('/api/v1/markets/stats')
        this.totalCount = data.total ?? 0
      } catch { /* non-fatal */ }
    },
    async fetchTags() {
      try {
        const { data } = await axios.get('/api/v1/markets/tags')
        this.tags = data
      } catch { /* ignore */ }
    },
    setViewMode(mode) {
      this.viewMode = mode
      if (mode === 'categories') this.fetchMarkets()
      else this.fetchTrending()
    },
    selectMarket(market) {
      this.selectedMarket = market
    },
    closeDetail() {
      this.selectedMarket = null
    },
    setTag(slug) {
      this.activeTag = slug
      this.fetchMarkets()
    },
    async createAlert(conditionId, tokenId, direction, threshold) {
      await axios.post('/api/v1/alerts', { conditionId, tokenId, direction, threshold })
    },
    toggleSelect(market) {
      const id = market.conditionId
      if (this.selectedMarkets[id]) {
        const next = { ...this.selectedMarkets }
        delete next[id]
        this.selectedMarkets = next
      } else {
        this.selectedMarkets = { ...this.selectedMarkets, [id]: true }
      }
    },
    clearSelection() {
      this.selectedMarkets = {}
    },
    // Called from WebSocket handler
    onMarketsLoading(data) {
      this.syncing = true
      this.loadProgress = data
    },
    onMarketsReady() {
      this.syncing = false
      this.fetchStats()
    },
  }
})
