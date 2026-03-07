<template>
  <div class="markets-view">

    <!-- Page header -->
    <div class="page-header anim-in">
      <div class="header-top">
        <h2 class="view-title">{{ $t('markets.title') }}</h2>
        <span v-if="!store.loading && sortedMarkets.length" class="market-count">
          {{ sortedMarkets.length }} markets
        </span>
      </div>
      <div class="search-row">
        <div class="search-wrap">
          <svg class="search-icon" width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <circle cx="11" cy="11" r="8"/><line x1="21" y1="21" x2="16.65" y2="16.65"/>
          </svg>
          <input
            v-model="store.search"
            class="search-input"
            :placeholder="$t('markets.search_placeholder')"
          />
          <button v-if="store.search" class="search-clear" @click="store.search = ''">
            <svg width="11" height="11" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5">
              <line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/>
            </svg>
          </button>
        </div>
        <select v-model="sortBy" class="sort-select">
          <option value="volume">{{ $t('markets.sort_volume') }}</option>
          <option value="liquidity">{{ $t('markets.sort_liquidity') }}</option>
          <option value="newest">{{ $t('markets.sort_newest') }}</option>
        </select>
      </div>
    </div>

    <!-- Tag filter -->
    <TagFilter
      :tags="store.tags"
      :model-value="store.activeTag"
      @update:model-value="store.setTag($event)"
    />

    <!-- Skeleton loading -->
    <div v-if="store.loading" class="cards-grid">
      <div v-for="i in 6" :key="i" class="skeleton-card">
        <div class="sk sk-hdr" />
        <div class="sk sk-title" />
        <div class="sk sk-title sk-short" />
        <div class="sk sk-probs" />
        <div class="sk sk-meta" />
      </div>
    </div>

    <!-- Empty state -->
    <div v-else-if="sortedMarkets.length === 0" class="empty-state anim-in">
      <svg width="36" height="36" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" class="empty-icon">
        <circle cx="11" cy="11" r="8"/><line x1="21" y1="21" x2="16.65" y2="16.65"/>
      </svg>
      <div>{{ store.search ? $t('markets.no_results') : $t('markets.no_markets') }}</div>
      <button v-if="store.activeTag || store.search" class="btn-reset" @click="resetFilters">
        RESET FILTERS
      </button>
    </div>

    <!-- Markets grid -->
    <div v-else class="cards-grid">
      <MarketCard
        v-for="m in sortedMarkets"
        :key="m.conditionId"
        :market="m"
        :selected="!!store.selectedMarkets[m.conditionId]"
        @select="store.selectMarket(m)"
        @buy="store.selectMarket(m)"
        @alert="openAlert(m)"
        @quickbuy="onQuickBuy($event)"
        @toggle-select="store.toggleSelect(m)"
      />
    </div>

    <!-- Detail panel -->
    <MarketDetailPanel
      v-if="store.selectedMarket"
      :market="store.selectedMarket"
      @close="store.closeDetail()"
      @alert="openAlert($event)"
    />

    <!-- Alert dialog -->
    <PriceAlertDialog
      v-if="alertMarket"
      :market="alertMarket"
      @close="alertMarket = null"
      @created="onAlertCreated($event)"
    />

    <!-- Quick buy dialog -->
    <QuickBuyDialog
      v-if="quickBuy.market"
      :market="quickBuy.market"
      :side="quickBuy.side"
      @close="quickBuy.market = null"
      @placed="quickBuy.market = null"
    />

    <!-- Batch buy action bar -->
    <ActionBar
      :count="store.selectedCount"
      :side="store.batchSide"
      :markets="store.selectedMarketsArray"
      @clear="store.clearSelection()"
      @update:side="store.batchSide = $event"
    />
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { useMarketsStore } from '@/stores/markets'
import TagFilter from '@/components/markets/TagFilter.vue'
import MarketCard from '@/components/markets/MarketCard.vue'
import MarketDetailPanel from '@/components/markets/MarketDetailPanel.vue'
import PriceAlertDialog from '@/components/markets/PriceAlertDialog.vue'
import QuickBuyDialog from '@/components/markets/QuickBuyDialog.vue'
import ActionBar from '@/components/markets/ActionBar.vue'

const store = useMarketsStore()
const sortBy = ref('volume')
const alertMarket = ref(null)
const quickBuy = reactive({ market: null, side: 'YES' })

onMounted(async () => {
  await Promise.all([store.fetchTags(), store.fetchMarkets()])
})

const sortedMarkets = computed(() => {
  const list = [...store.filteredMarkets]
  if (sortBy.value === 'volume') {
    list.sort((a, b) => parseFloat(b.volume || 0) - parseFloat(a.volume || 0))
  } else if (sortBy.value === 'liquidity') {
    list.sort((a, b) => parseFloat(b.liquidity || 0) - parseFloat(a.liquidity || 0))
  } else if (sortBy.value === 'newest') {
    list.sort((a, b) => new Date(b.createdAt || 0) - new Date(a.createdAt || 0))
  }
  return list
})

function openAlert(market) { alertMarket.value = market }

function resetFilters() { store.setTag(''); store.search = '' }

function onQuickBuy({ market, side }) {
  quickBuy.market = market
  quickBuy.side = side
}

async function onAlertCreated({ conditionId, tokenId, direction, threshold }) {
  try { await store.createAlert(conditionId, tokenId, direction, threshold) }
  catch (e) { console.error('Alert creation failed:', e) }
}
</script>

<style scoped>
.markets-view {
  display: flex;
  flex-direction: column;
  gap: 0.9rem;
}

/* Header */
.page-header { display: flex; flex-direction: column; gap: 0.65rem; }
.header-top { display: flex; align-items: baseline; gap: 0.75rem; }
.view-title { font-size: 1rem; font-weight: 700; letter-spacing: 0.04em; color: var(--text-bright); }
.market-count { font-size: 0.90rem; font-weight: 600; letter-spacing: 0.08em; color: var(--text-secondary); font-family: var(--font-mono); }

.search-row { display: flex; gap: 0.5rem; }
.search-wrap { flex: 1; position: relative; display: flex; align-items: center; }
.search-icon { position: absolute; left: 0.65rem; color: var(--text-muted); pointer-events: none; }
.search-input {
  width: 100%; padding: 0.45rem 2.2rem 0.45rem 2rem;
  background: var(--bg-card); border: 1px solid var(--border);
  border-radius: var(--radius); color: var(--text-primary);
  font-family: var(--font-mono); font-size: 0.96rem;
  outline: none; transition: border-color var(--transition);
}
.search-input::placeholder { color: var(--text-muted); }
.search-input:focus { border-color: var(--accent); box-shadow: 0 0 0 1px rgba(0,200,255,0.12); }
.search-clear {
  position: absolute; right: 0.6rem;
  background: none; border: none; color: var(--text-muted);
  cursor: pointer; padding: 2px; display: flex; align-items: center;
  transition: color var(--transition);
}
.search-clear:hover { color: var(--text-primary); }

.sort-select {
  padding: 0.45rem 0.75rem;
  background: var(--bg-card); border: 1px solid var(--border);
  border-radius: var(--radius); color: var(--text-primary);
  font-family: var(--font-mono); font-size: 0.92rem; cursor: pointer; outline: none;
  transition: border-color var(--transition);
}
.sort-select:focus { border-color: var(--accent); }

/* Cards grid */
.cards-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(340px, 1fr));
  gap: 0.6rem;
}

/* Skeleton */
.skeleton-card {
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-top: 2px solid rgba(0,200,255,0.15);
  border-radius: var(--radius);
  padding: 1rem;
  display: flex; flex-direction: column; gap: 0.65rem;
}
.sk {
  background: linear-gradient(90deg, var(--bg-card) 25%, var(--bg-hover) 50%, var(--bg-card) 75%);
  background-size: 600px 100%;
  animation: shimmer 1.6s infinite linear;
  border-radius: var(--radius);
}
.sk-hdr   { height: 12px; width: 35%; }
.sk-title { height: 13px; width: 92%; }
.sk-short { width: 65%; }
.sk-probs { height: 72px; }
.sk-meta  { height: 11px; width: 48%; }

/* Empty state */
.empty-state {
  display: flex; flex-direction: column; align-items: center;
  justify-content: center; gap: 0.65rem;
  padding: 5rem 2rem; color: var(--text-muted); font-size: 0.96rem;
  text-align: center;
}
.empty-icon { opacity: 0.35; color: var(--text-secondary); }

.btn-reset {
  padding: 0.35rem 1.1rem;
  background: transparent; border: 1px solid var(--border);
  border-radius: var(--radius); color: var(--text-secondary);
  font-family: var(--font-mono); font-size: 0.94rem; font-weight: 700;
  letter-spacing: 0.08em; cursor: pointer; transition: all var(--transition);
}
.btn-reset:hover { border-color: var(--accent); color: var(--accent); }
</style>
