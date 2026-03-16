<template>
  <div class="tag-filter-wrap">
    <div class="tag-filter">
      <button
        v-for="tag in allTags"
        :key="tag.value"
        class="tag-pill"
        :class="{ active: modelValue === tag.value }"
        @click="$emit('update:modelValue', tag.value)"
      >
  <span v-if="tag.icon" class="tag-icon">{{ tag.icon }}</span>{{ tag.label }}
      </button>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'

const props = defineProps({
  tags: { type: Array, default: () => [] },
  modelValue: { type: String, default: '' },
})
defineEmits(['update:modelValue'])
const { t } = useI18n()

const CATEGORY_ICONS = {
  politics: '🏛',
  'us-politics': '🏛',
  sports: '⚽',
  crypto: '🔮',
  science: '🔬',
  business: '💼',
  culture: '🎭',
  tech: '💻',
  weather: '🌦',
  entertainment: '🎬',
  economics: '📈',
  world: '🌍',
  nba: '🏀',
  nfl: '🏈',
  soccer: '⚽',
}

function tagIcon(slug) {
  return CATEGORY_ICONS[slug] ?? ''
}

const allTags = computed(() => [
  { value: '', label: t('markets.filter_all'), icon: '' },
  ...props.tags.map(tg => ({ value: tg.slug, label: tg.label, icon: tagIcon(tg.slug) })),
])
</script>

<style scoped>
.tag-filter-wrap {
  overflow-x: auto;
  scrollbar-width: none;
  -ms-overflow-style: none;
}
.tag-filter-wrap::-webkit-scrollbar { display: none; }

.tag-filter {
  display: flex;
  flex-wrap: nowrap;
  gap: 0.35rem;
  padding-bottom: 2px;
}

.tag-pill {
  padding: 0.38rem 1.00rem;
  border-radius: 2px;
  border: 1px solid var(--border);
  background: transparent;
  color: var(--text-muted);
  cursor: pointer;
  font-size: 0.86rem;
  font-family: var(--font-mono);
  font-weight: 600;
  letter-spacing: 0.04em;
  white-space: nowrap;
  flex-shrink: 0;
  transition: all var(--transition);
}

.tag-pill:hover:not(.active) {
  border-color: rgba(124, 58, 237, 0.40);
  color: var(--accent);
}

.tag-pill.active {
  background: var(--accent-dim);
  color: var(--accent);
  border-color: var(--accent);
}

.tag-icon { margin-right: 0.3rem; font-style: normal; }
</style>
