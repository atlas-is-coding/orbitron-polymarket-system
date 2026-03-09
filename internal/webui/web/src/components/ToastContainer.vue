<template>
  <div class="toast-container">
    <TransitionGroup name="toast-list">
      <div
        v-for="t in toasts"
        :key="t.id"
        class="toast"
        :class="`toast--${t.type}`"
      >
        <span class="toast-icon">{{ icons[t.type] || '◈' }}</span>
        <span class="toast-msg">{{ t.msg }}</span>
      </div>
    </TransitionGroup>
  </div>
</template>

<script setup>
import { storeToRefs } from 'pinia'
import { useAppStore } from '@/stores/app'

const { toasts } = storeToRefs(useAppStore())

const icons = {
  success: '✓',
  error:   '✗',
  info:    '◈',
  warning: '⚠',
}
</script>

<style scoped>
.toast-container {
  position: fixed;
  bottom: 1.25rem;
  right: 1.25rem;
  display: flex;
  flex-direction: column;
  gap: 0.4rem;
  z-index: 999;
  pointer-events: none;
}

.toast {
  display: flex;
  align-items: flex-start;
  gap: 0.55rem;
  padding: 0.55rem 0.9rem;
  border-radius: var(--radius);
  font-size: 0.92rem;
  font-family: var(--font-mono);
  border: 1px solid var(--border);
  background: var(--bg-card);
  color: var(--text-primary);
  box-shadow: var(--shadow), 0 0 20px rgba(0,0,0,0.3);
  max-width: 380px;
  pointer-events: auto;
  line-height: 1.4;
  border-top: 2px solid transparent;
}

.toast-icon {
  font-size: 0.90rem;
  flex-shrink: 0;
  margin-top: 0.05rem;
  font-weight: 700;
}

.toast-msg {
  flex: 1;
  word-break: break-word;
}

.toast--success { border-top-color: var(--success); }
.toast--success .toast-icon { color: var(--success); }

.toast--error   { border-top-color: var(--danger); }
.toast--error .toast-icon { color: var(--danger); }

.toast--info    { border-top-color: var(--accent); }
.toast--info .toast-icon { color: var(--accent); }

.toast--warning { border-top-color: var(--warning); }
.toast--warning .toast-icon { color: var(--warning); }

/* TransitionGroup animations */
.toast-list-enter-active { animation: slideInRight 0.20s ease both; }
.toast-list-leave-active { animation: slideInRight 0.18s ease reverse both; }
.toast-list-move { transition: transform 0.20s ease; }
</style>
