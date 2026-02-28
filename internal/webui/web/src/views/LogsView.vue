<template>
  <div class="view">
    <div class="view-header">
      <h2 class="view-title">{{ $t('nav.logs') }}</h2>
      <button class="btn-ghost" @click="app.logs = []">{{ $t('logs.clear') }}</button>
    </div>

    <div class="log-wrap" ref="logWrap">
      <div v-if="!logs.length" class="empty">{{ $t('logs.noLogs') }}</div>
      <div v-for="(entry, i) in logs" :key="i" class="log-row" :class="`log-${entry.level}`">
        <span class="log-level">{{ entry.level?.toUpperCase().slice(0, 4) }}</span>
        <span class="log-msg">{{ entry.message }}</span>
      </div>
    </div>
  </div>
</template>

<script setup>
import { onMounted, watch, nextTick, ref } from 'vue'
import { storeToRefs } from 'pinia'
import { useAppStore } from '@/stores/app'
import { useApi } from '@/composables/useApi'

const app = useAppStore()
const { logs } = storeToRefs(app)
const api = useApi()
const logWrap = ref(null)

onMounted(async () => {
  try {
    const data = await api.getLogs()
    app.logs = data
  } catch {}
})

watch(logs, async () => {
  await nextTick()
  if (logWrap.value) logWrap.value.scrollTop = 0
})
</script>

<style scoped>
.view { display: flex; flex-direction: column; gap: 1rem; height: calc(100vh - 120px); }
.view-header { display: flex; align-items: center; justify-content: space-between; }
.view-title  { font-size: 1.4rem; font-weight: 700; }

.btn-ghost {
  background: none; border: 1px solid var(--border); color: var(--text-secondary);
  border-radius: var(--radius); padding: 0.3rem 0.8rem; font-size: 0.8rem; cursor: pointer;
}
.btn-ghost:hover { background: var(--bg-hover); }

.log-wrap {
  flex: 1;
  overflow-y: auto;
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  font-family: var(--font-mono);
  font-size: 0.8rem;
  padding: 0.5rem 0;
}

.empty { padding: 2rem; text-align: center; color: var(--text-muted); font-family: var(--font-ui); }

.log-row {
  display: flex;
  gap: 0.75rem;
  padding: 0.25rem 1rem;
  border-bottom: 1px solid transparent;
  line-height: 1.5;
  transition: background var(--transition);
}
.log-row:hover { background: var(--bg-hover); }

.log-level {
  width: 3.5rem;
  flex-shrink: 0;
  font-weight: 700;
  font-size: 0.72rem;
}

.log-msg {
  color: var(--text-primary);
  word-break: break-all;
  white-space: pre-wrap;
}

/* Level colors */
.log-trace .log-level { color: var(--text-muted); }
.log-debug .log-level { color: var(--text-secondary); }
.log-info  .log-level { color: var(--accent); }
.log-warn  .log-level { color: var(--warning); }
.log-error .log-level { color: var(--danger); }

.log-warn  { background: rgba(210,153,34,0.06); }
.log-error { background: rgba(248,81,73,0.08); }
</style>
