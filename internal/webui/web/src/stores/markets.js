import { defineStore } from 'pinia'
import axios from 'axios'

export const useMarketsStore = defineStore('markets', {
  state: () => ({
    markets: [],
    tags: [],
    activeTag: '',
    search: '',
    sortBy: 'volume',
    selectedMarket: null,
    loading: false,
    error: null,
    // multi-select batch buy
    selectedMarkets: {},   // conditionId → true
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
    async fetchTags() {
      try {
        const { data } = await axios.get('/api/v1/markets/tags')
        this.tags = data
      } catch { /* ignore */ }
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
  }
})
