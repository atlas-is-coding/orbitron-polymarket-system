<template>
  <div class="markets-view">
    <div class="page-header">
      <h1 class="page-title">{{ $t('markets.title') }}</h1>
    </div>

    <TagFilter
      :tags="store.tags"
      :model-value="store.activeTag"
      @update:model-value="store.setTag($event)"
    />

    <div class="search-sort-bar">
      <input
        v-model="store.search"
        class="search-input"
        :placeholder="$t('markets.search_placeholder')"
      />
      <select v-model="sortBy" class="sort-select">
        <option value="volume">{{ $t('markets.sort_volume') }}</option>
        <option value="liquidity">{{ $t('markets.sort_liquidity') }}</option>
        <option value="newest">{{ $t('markets.sort_newest') }}</option>
      </select>
    </div>

    <div v-if="store.loading" class="state-msg">{{ $t('markets.loading') }}</div>
    <div v-else-if="store.filteredMarkets.length === 0" class="state-msg">{{ $t('markets.no_markets') }}</div>
    <div v-else class="cards-grid">
      <MarketCard
        v-for="m in sortedMarkets"
        :key="m.conditionId"
        :market="m"
        @select="store.selectMarket(m)"
        @buy="store.selectMarket(m)"
        @alert="openAlert(m)"
      />
    </div>

    <MarketDetailPanel
      v-if="store.selectedMarket"
      :market="store.selectedMarket"
      @close="store.closeDetail()"
      @alert="openAlert($event)"
    />

    <PriceAlertDialog
      v-if="alertMarket"
      :market="alertMarket"
      @close="alertMarket = null"
      @created="onAlertCreated($event)"
    />
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useMarketsStore } from '@/stores/markets'
import TagFilter from '@/components/markets/TagFilter.vue'
import MarketCard from '@/components/markets/MarketCard.vue'
import MarketDetailPanel from '@/components/markets/MarketDetailPanel.vue'
import PriceAlertDialog from '@/components/markets/PriceAlertDialog.vue'

const store = useMarketsStore()
const sortBy = ref('volume')
const alertMarket = ref(null)

onMounted(async () => {
  await Promise.all([store.fetchTags(), store.fetchMarkets()])
})

const sortedMarkets = computed(() => {
  const list = [...store.filteredMarkets]
  if (sortBy.value === 'volume') {
    list.sort((a, b) => parseFloat(b.volume || 0) - parseFloat(a.volume || 0))
  } else if (sortBy.value === 'liquidity') {
    list.sort((a, b) => parseFloat(b.liquidity || 0) - parseFloat(a.liquidity || 0))
  }
  return list
})

function openAlert(market) {
  alertMarket.value = market
}

async function onAlertCreated({ conditionId, tokenId, direction, threshold }) {
  try {
    await store.createAlert(conditionId, tokenId, direction, threshold)
  } catch (e) {
    console.error('Alert creation failed:', e)
  }
}
</script>

<style scoped>
.markets-view { padding: 28px; max-width: 1300px; }
.page-header { display: flex; align-items: baseline; gap: 16px; margin-bottom: 24px; }
.page-title { font-size: 1.5rem; font-weight: 700; margin: 0; }
.search-sort-bar { display: flex; gap: 12px; margin-bottom: 24px; }
.search-input {
  flex: 1; padding: 10px 16px;
  background: var(--surface-2); border: 1px solid var(--border);
  border-radius: 8px; color: var(--text);
  font-family: 'IBM Plex Mono', monospace; font-size: 0.92rem;
}
.search-input:focus { outline: none; border-color: var(--accent); }
.sort-select {
  padding: 10px 16px;
  background: var(--surface-2); border: 1px solid var(--border);
  border-radius: 8px; color: var(--text);
  font-family: 'IBM Plex Mono', monospace; font-size: 0.9rem; cursor: pointer;
}
.cards-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(460px, 1fr));
  gap: 18px;
}
.state-msg { padding: 60px 0; text-align: center; color: var(--text-muted); font-size: 1rem; }
</style>
