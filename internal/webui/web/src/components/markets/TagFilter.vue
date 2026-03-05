<template>
  <div class="tag-filter">
    <button
      v-for="tag in allTags"
      :key="tag.value"
      class="tag-pill"
      :class="{ active: modelValue === tag.value }"
      @click="$emit('update:modelValue', tag.value)"
    >
      {{ tag.label }}
    </button>
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

const allTags = computed(() => [
  { value: '', label: t('markets.filter_all') },
  ...props.tags.map(tg => ({ value: tg.slug, label: tg.label }))
])
</script>

<style scoped>
.tag-filter { display: flex; flex-wrap: wrap; gap: 8px; margin-bottom: 20px; }
.tag-pill {
  padding: 5px 16px;
  border-radius: 20px;
  border: 1px solid var(--border);
  background: transparent;
  color: var(--text-muted);
  cursor: pointer;
  font-size: 0.82rem;
  font-family: 'IBM Plex Mono', monospace;
  transition: all 0.15s;
}
.tag-pill:hover { border-color: var(--accent); color: var(--accent); }
.tag-pill.active { background: var(--accent); color: #fff; border-color: var(--accent); }
</style>
