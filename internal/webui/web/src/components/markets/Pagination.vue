<template>
  <div v-if="totalPages > 1" class="pagination anim-in">
    <button 
      class="pag-btn" 
      :disabled="page === 1" 
      @click="$emit('change', 1)"
      title="First Page"
    >
      <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <path d="M11 17l-5-5 5-5M18 17l-5-5 5-5"/>
      </svg>
    </button>
    
    <button 
      class="pag-btn" 
      :disabled="page === 1" 
      @click="$emit('change', page - 1)"
    >
      <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <path d="M15 18l-6-6 6-6"/>
      </svg>
    </button>

    <div class="pag-numbers">
      <template v-for="p in visiblePages" :key="p">
        <span v-if="p === '...'" class="pag-dots">...</span>
        <button 
          v-else
          :class="['pag-num', { active: p === page }]" 
          @click="$emit('change', p)"
        >{{ p }}</button>
      </template>
    </div>

    <button 
      class="pag-btn" 
      :disabled="page === totalPages" 
      @click="$emit('change', page + 1)"
    >
      <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <path d="M9 18l6-6-6-6"/>
      </svg>
    </button>
    
    <button 
      class="pag-btn" 
      :disabled="page === totalPages" 
      @click="$emit('change', totalPages)"
      title="Last Page"
    >
      <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <path d="M13 17l5-5-5-5M6 17l5-5-5-5"/>
      </svg>
    </button>

    <span class="pag-info">
      PAGE {{ page }} OF {{ totalPages }}
    </span>
  </div>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  page: { type: Number, required: true },
  totalPages: { type: Number, required: true }
})

defineEmits(['change'])

const visiblePages = computed(() => {
  const total = props.totalPages
  const current = props.page
  const delta = 2
  const pages = []

  for (let i = 1; i <= total; i++) {
    if (i === 1 || i === total || (i >= current - delta && i <= current + delta)) {
      pages.push(i)
    } else if (pages[pages.length - 1] !== '...') {
      pages.push('...')
    }
  }
  return pages
})
</script>

<style scoped>
.pagination {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.4rem;
  margin-top: 1.5rem;
  padding: 1rem 0;
  border-top: 1px solid var(--border-subtle);
  font-family: var(--font-mono);
}

.pag-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: var(--radius-sm);
  color: var(--text-secondary);
  cursor: pointer;
  transition: all var(--transition);
}

.pag-btn:hover:not(:disabled) {
  border-color: var(--accent);
  color: var(--accent);
  background: var(--bg-hover);
}

.pag-btn:disabled {
  opacity: 0.3;
  cursor: not-allowed;
}

.pag-numbers {
  display: flex;
  align-items: center;
  gap: 0.3rem;
}

.pag-num {
  min-width: 32px;
  height: 32px;
  padding: 0 0.5rem;
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: var(--radius-sm);
  color: var(--text-secondary);
  font-size: 0.85rem;
  font-weight: 600;
  cursor: pointer;
  transition: all var(--transition);
  display: flex;
  align-items: center;
  justify-content: center;
}

.pag-num:hover:not(.active) {
  border-color: var(--accent);
  color: var(--accent);
}

.pag-num.active {
  background: var(--accent);
  border-color: var(--accent);
  color: var(--text-inverse);
  box-shadow: var(--accent-glow);
}

.pag-dots {
  color: var(--text-muted);
  padding: 0 0.2rem;
  font-size: 0.8rem;
}

.pag-info {
  margin-left: 0.8rem;
  font-size: 0.72rem;
  font-weight: 600;
  letter-spacing: 0.08em;
  color: var(--text-muted);
  text-transform: uppercase;
}
</style>
