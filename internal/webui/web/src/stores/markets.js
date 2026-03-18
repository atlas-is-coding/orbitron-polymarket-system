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
    // pagination
    page: 1,
    pageSize: 24,
    totalFiltered: 0,
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
    totalPages(state) {
      return Math.ceil(state.totalFiltered / state.pageSize)
    }
  },
  actions: {
    async fetchMarkets(options = { background: false }) {
      if (!options.background) this.loading = true
      this.error = null
      try {
        const params = new URLSearchParams()
        if (this.activeTag) params.set('tag', this.activeTag)
        
        const offset = (this.page - 1) * this.pageSize
        params.set('limit', this.pageSize)
        params.set('offset', offset)

        const { data } = await axios.get(`/api/v1/markets?${params}`)
        this.markets = data.markets
        this.totalFiltered = data.total
      } catch (e) {
        this.error = e.message
      } finally {
        if (!options.background) this.loading = false
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
      this.page = 1
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
      this.page = 1
      this.fetchMarkets()
    },
    setPage(p) {
      this.page = p
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
