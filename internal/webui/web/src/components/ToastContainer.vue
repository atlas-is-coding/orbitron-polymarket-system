<template>
  <div class="toast-container">
    <div
      v-for="t in toasts"
      :key="t.id"
      class="toast"
      :class="`toast--${t.type}`"
    >{{ t.msg }}</div>
  </div>
</template>

<script setup>
import { storeToRefs } from 'pinia'
import { useAppStore } from '@/stores/app'
const { toasts } = storeToRefs(useAppStore())
</script>

<style scoped>
.toast-container {
  position: fixed;
  bottom: 1.25rem;
  right: 1.25rem;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  z-index: 999;
  pointer-events: none;
}
.toast {
  padding: 0.55rem 1rem;
  border-radius: var(--radius);
  font-size: 0.82rem;
  border: 1px solid var(--border);
  background: var(--bg-card);
  color: var(--text-primary);
  animation: slideInRight 0.22s ease both;
  box-shadow: var(--shadow);
  max-width: 320px;
  pointer-events: auto;
}
.toast--success { border-color: var(--success); color: var(--success); }
.toast--error   { border-color: var(--danger);  color: var(--danger); }
.toast--info    { border-color: var(--accent);  color: var(--accent-bright); }
</style>
