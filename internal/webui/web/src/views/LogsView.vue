<template>
  <div class="view">
    <div class="view-header anim-in">
      <div class="header-left">
        <h2 class="view-title">{{ $t('nav.logs') }}</h2>
        <span class="counter mono">{{ filteredLogs.length }}</span>
      </div>
      <div class="header-actions">
        <div class="filter-tabs">
          <button
            v-for="lv in levels"
            :key="lv"
            class="filter-tab"
            :class="{ 'filter-tab--active': filterLevel === lv, [`lvl-${lv.toLowerCase()}`]: lv !== 'ALL' }"
            @click="filterLevel = lv"
          >{{ lv }}</button>
        </div>
        <button class="btn-ghost" :class="{ 'btn-ghost--active': autoScroll }" @click="toggleAutoScroll" :title="$t('logs.autoScroll')">
          ⇩
        </button>
        <button class="btn-ghost" @click="app.logs = []; newCount = 0">{{ $t('logs.clear') }}</button>
      </div>
    </div>

    <div class="log-shell">
      <div class="log-wrap" ref="logWrap" @scroll="onScroll">
        <div v-if="!filteredLogs.length" class="empty">{{ $t('logs.noLogs') }}</div>
        <div
          v-for="(entry, i) in filteredLogs"
          :key="i"
          class="log-row"
          :class="`log-${entry.level?.toLowerCase()}`"
        >
          <span class="log-ts mono">{{ fmtTime(entry.time) }}</span>
          <span class="log-level" :class="`lvl-${entry.level?.toLowerCase()}`">{{ entry.level?.toUpperCase().slice(0, 4) }}</span>
          <span class="log-msg">{{ entry.message }}</span>
        </div>
      </div>

      <!-- New lines badge -->
      <div v-if="newCount > 0 && !atBottom" class="new-badge" @click="scrollToBottom">
        ↓ {{ newCount }} new
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed, nextTick, onMounted, ref, watch } from 'vue'
import { storeToRefs } from 'pinia'
import { useAppStore } from '@/stores/app'
import { useApi } from '@/composables/useApi'

const app = useAppStore()
const { logs } = storeToRefs(app)
const api = useApi()

const logWrap = ref(null)
const autoScroll = ref(true)
const atBottom = ref(true)
const newCount = ref(0)
const filterLevel = ref('ALL')

const levels = ['ALL', 'TRACE', 'DEBUG', 'INFO', 'WARN', 'ERROR']

const filteredLogs = computed(() => {
  if (filterLevel.value === 'ALL') return logs.value
  return logs.value.filter(e => e.level?.toUpperCase() === filterLevel.value)
})

function fmtTime(ts) {
  if (!ts) return ''
  try {
    const d = new Date(ts)
    return d.toTimeString().slice(0, 8)
  } catch { return '' }
}

function isAtBottom() {
  const el = logWrap.value
  if (!el) return true
  return el.scrollHeight - el.scrollTop - el.clientHeight < 36
}

function onScroll() {
  atBottom.value = isAtBottom()
  if (atBottom.value) newCount.value = 0
}

function scrollToBottom() {
  const el = logWrap.value
  if (el) el.scrollTop = el.scrollHeight
  newCount.value = 0
  atBottom.value = true
}

function toggleAutoScroll() {
  autoScroll.value = !autoScroll.value
  if (autoScroll.value) scrollToBottom()
}

watch(logs, async () => {
  await nextTick()
  if (autoScroll.value) {
    scrollToBottom()
  } else if (!isAtBottom()) {
    newCount.value++
  }
})

onMounted(async () => {
  try {
    const data = await api.getLogs()
    app.logs = data
    await nextTick()
    scrollToBottom()
  } catch {}
})
</script>

<style scoped>
.view { display: flex; flex-direction: column; gap: 0.75rem; height: calc(100vh - 100px); }
.view-header { display: flex; align-items: center; justify-content: space-between; flex-wrap: wrap; gap: 0.5rem; }
.header-left { display: flex; align-items: center; gap: 0.75rem; }
.view-title { font-size: 1.1rem; font-weight: 700; }
.counter { font-size: 0.7rem; color: var(--text-muted); }

.header-actions { display: flex; align-items: center; gap: 0.5rem; flex-wrap: wrap; }

.filter-tabs { display: flex; gap: 0.2rem; background: var(--bg-card); border: 1px solid var(--border); border-radius: var(--radius); padding: 0.2rem; }
.filter-tab { background: none; border: none; color: var(--text-muted); border-radius: calc(var(--radius) - 2px); padding: 0.15rem 0.55rem; font-size: 0.7rem; font-family: var(--font-mono); font-weight: 700; cursor: pointer; transition: all var(--transition); }
.filter-tab:hover { color: var(--text-primary); }
.filter-tab--active { background: var(--bg-hover); color: var(--text-primary); }
.filter-tab.lvl-trace.filter-tab--active { color: var(--text-muted); }
.filter-tab.lvl-debug.filter-tab--active { color: var(--text-secondary); }
.filter-tab.lvl-info.filter-tab--active  { color: var(--accent-bright); }
.filter-tab.lvl-warn.filter-tab--active  { color: var(--warning); }
.filter-tab.lvl-error.filter-tab--active { color: var(--danger); }

.btn-ghost { background: none; border: 1px solid var(--border); color: var(--text-secondary); border-radius: var(--radius); padding: 0.28rem 0.7rem; font-size: 0.8rem; cursor: pointer; transition: all var(--transition); }
.btn-ghost:hover { background: var(--bg-hover); }
.btn-ghost--active { border-color: var(--accent); color: var(--accent-bright); }

.log-shell { position: relative; flex: 1; min-height: 0; display: flex; flex-direction: column; }

.log-wrap {
  flex: 1;
  min-height: 0;
  overflow-y: auto;
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  font-family: var(--font-mono);
  font-size: 0.78rem;
  padding: 0.5rem 0;
  scroll-behavior: smooth;
}

.empty { padding: 2rem; text-align: center; color: var(--text-muted); font-family: var(--font-ui); }

.log-row {
  display: flex;
  gap: 0.75rem;
  padding: 0.22rem 1rem;
  line-height: 1.5;
  transition: background var(--transition);
  align-items: baseline;
}
.log-row:hover { background: var(--bg-hover); }

.log-ts { width: 5.5rem; flex-shrink: 0; color: var(--text-muted); font-size: 0.7rem; }
.log-level { width: 3rem; flex-shrink: 0; font-weight: 700; font-size: 0.7rem; }
.log-msg { color: var(--text-primary); word-break: break-all; white-space: pre-wrap; }

/* Level badge colors */
.lvl-trace { color: var(--text-muted); }
.lvl-debug { color: var(--text-secondary); }
.lvl-info  { color: var(--accent-bright); }
.lvl-warn  { color: var(--warning); }
.lvl-error { color: var(--danger); }

/* Row background tints */
.log-warn  { background: rgba(251,191,36,0.04); }
.log-error { background: rgba(248,113,113,0.06); }

/* New badge */
.new-badge {
  position: absolute;
  bottom: 0.75rem;
  left: 50%;
  transform: translateX(-50%);
  background: var(--accent);
  color: #fff;
  border-radius: 999px;
  padding: 0.2rem 0.85rem;
  font-size: 0.7rem;
  font-family: var(--font-mono);
  cursor: pointer;
  box-shadow: 0 2px 8px rgba(124,58,237,0.4);
  animation: fadeSlideUp 0.2s ease both;
  user-select: none;
}
.new-badge:hover { background: var(--accent-hover); }
</style>
