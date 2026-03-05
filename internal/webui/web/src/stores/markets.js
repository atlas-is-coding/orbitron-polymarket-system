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
  }),
  getters: {
    filteredMarkets(state) {
      let list = state.markets
      if (state.search) {
        const q = state.search.toLowerCase()
        list = list.filter(m => m.question?.toLowerCase().includes(q))
      }
      return list
    }
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
  }
})
