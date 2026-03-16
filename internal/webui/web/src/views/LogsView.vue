<template>
  <div class="view">
    <div class="view-header anim-in">
      <div class="header-left">
        <h2 class="view-title">{{ $t('nav.logs') }}</h2>
        <span class="counter">{{ filteredLogs.length }}</span>
      </div>
      <div class="header-actions">
        <div class="filter-tabs">
          <button
            v-for="lv in levels"
            :key="lv"
            class="filter-tab"
            :class="[
              { 'filter-tab--active': filterLevel === lv },
              lv !== 'ALL' ? `lvl-${lv.toLowerCase()}` : ''
            ]"
            @click="filterLevel = lv"
          >{{ lv }}</button>
        </div>
        <button class="ctrl-btn" :class="{ 'ctrl-btn--active': autoScroll }" @click="toggleAutoScroll" :title="$t('logs.autoScroll')">
          ⇩
        </button>
        <button class="ctrl-btn" @click="app.logs = []; newCount = 0">{{ $t('logs.clear') }}</button>
      </div>
    </div>

    <!-- Terminal shell -->
    <div class="terminal-shell anim-in">
      <!-- Terminal chrome -->
      <div class="term-chrome">
        <div class="chrome-dots">
          <span class="cdot cdot--r" />
          <span class="cdot cdot--y" />
          <span class="cdot cdot--g" />
        </div>
        <span class="term-title">POLYTRADE // SYSTEM LOG</span>
        <span class="term-count">{{ filteredLogs.length }} entries</span>
      </div>

      <!-- Log output -->
      <div class="log-wrap" ref="logWrap" @scroll="onScroll">
        <div v-if="!filteredLogs.length" class="log-empty">
          <span class="prompt-glyph">$ </span>no log entries match filter
        </div>
        <div
          v-for="(entry, i) in filteredLogs"
          :key="i"
          class="log-row"
          :class="`log-${entry.level?.toLowerCase()}`"
        >
          <span class="log-ts">{{ fmtTime(entry.time) }}</span>
          <span class="log-level" :class="`lvl-${entry.level?.toLowerCase()}`">{{ entry.level?.toUpperCase().slice(0, 4) }}</span>
          <span class="log-sep">│</span>
          <span class="log-msg">{{ entry.message }}</span>
        </div>
        <!-- Blinking cursor at end -->
        <div class="log-cursor" v-if="filteredLogs.length">
          <span class="prompt-glyph">$ </span><span class="cursor-blink">█</span>
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
  if (!ts) return '--------'
  try { return new Date(ts).toTimeString().slice(0, 8) } catch { return '--------' }
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
  if (autoScroll.value) scrollToBottom()
  else if (!isAtBottom()) newCount.value++
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
.view {
  display: flex; flex-direction: column; gap: 0.9rem;
  height: calc(100vh - 100px);
}

/* Header */
.view-header { display: flex; align-items: center; justify-content: space-between; flex-wrap: wrap; gap: 0.5rem; flex-shrink: 0; }
.header-left { display: flex; align-items: center; gap: 0.75rem; }
.view-title { font-size: 1rem; font-weight: 700; letter-spacing: 0.04em; color: var(--text-bright); }
.counter { font-size: 0.92rem; color: var(--text-secondary); font-family: var(--font-mono); }
.header-actions { display: flex; align-items: center; gap: 0.4rem; flex-wrap: wrap; }

/* Filter tabs */
.filter-tabs {
  display: flex; gap: 1px;
  background: var(--bg-hover);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  padding: 2px;
}
.filter-tab {
  background: none; border: none;
  color: var(--text-muted);
  border-radius: calc(var(--radius) - 1px);
  padding: 0.25rem 0.65rem;
  font-size: 0.92rem; font-family: var(--font-mono); font-weight: 700;
  cursor: pointer; transition: all var(--transition);
  letter-spacing: 0.04em;
}
.filter-tab:hover { color: var(--text-primary); }
.filter-tab--active { background: var(--bg-card); }
.filter-tab.lvl-trace.filter-tab--active { color: var(--text-secondary); }
.filter-tab.lvl-debug.filter-tab--active { color: var(--text-secondary); }
.filter-tab.lvl-info.filter-tab--active  { color: var(--accent); }
.filter-tab.lvl-warn.filter-tab--active  { color: var(--warning); }
.filter-tab.lvl-error.filter-tab--active { color: var(--danger); }

/* Control buttons */
.ctrl-btn {
  background: none; border: 1px solid var(--border); color: var(--text-secondary);
  border-radius: var(--radius); padding: 0.32rem 0.80rem; font-size: 0.90rem;
  cursor: pointer; transition: all var(--transition); font-family: var(--font-mono);
}
.ctrl-btn:hover { background: var(--bg-hover); color: var(--text-primary); }
.ctrl-btn--active { border-color: var(--accent); color: var(--accent); }

/* Terminal shell */
.terminal-shell {
  flex: 1; min-height: 0;
  display: flex; flex-direction: column;
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-top: 2px solid var(--accent);
  border-radius: var(--radius);
  overflow: hidden;
  position: relative;
}

/* Scanline overlay */
.terminal-shell::after {
  content: '';
  position: absolute;
  inset: 0;
  background: repeating-linear-gradient(
    0deg, transparent, transparent 2px,
    rgba(124,58,237,0.01) 2px, rgba(124,58,237,0.01) 4px
  );
  pointer-events: none;
  z-index: 10;
}

/* Chrome bar */
.term-chrome {
  display: flex; align-items: center; gap: 0.75rem;
  padding: 0.6rem 1.2rem;
  background: rgba(124, 58, 237, 0.04);
  border-bottom: 1px solid var(--border);
  flex-shrink: 0;
}

.chrome-dots { display: flex; gap: 0.28rem; flex-shrink: 0; }
.cdot { width: 9px; height: 9px; border-radius: 50%; }
.cdot--r { background: #ff5f57; }
.cdot--y { background: #ffbd2e; }
.cdot--g { background: #28ca41; }

.term-title { font-size: 0.86rem; letter-spacing: 0.10em; color: var(--text-secondary); flex: 1; text-transform: uppercase; }
.term-count  { font-size: 0.86rem; color: var(--text-muted); letter-spacing: 0.06em; }

/* Log output */
.log-wrap {
  flex: 1; min-height: 0;
  overflow-y: auto;
  font-family: var(--font-mono);
  font-size: 0.90rem;
  padding: 0.4rem 0;
  scroll-behavior: smooth;
  position: relative;
  z-index: 1;
}

.log-empty {
  padding: 1.5rem 1rem;
  color: var(--text-muted);
  font-size: 0.90rem;
}

.prompt-glyph { color: var(--accent); }

.log-row {
  display: flex;
  gap: 0.6rem;
  padding: 0.18rem 1rem;
  line-height: 1.5;
  align-items: baseline;
  transition: background var(--transition);
}
.log-row:hover { background: rgba(124, 58, 237, 0.03); }

.log-ts    { width: 5rem; flex-shrink: 0; color: var(--text-muted); font-size: 0.94rem; }
.log-level { width: 3rem; flex-shrink: 0; font-weight: 700; font-size: 0.94rem; }
.log-sep   { color: var(--border); user-select: none; }
.log-msg   { color: var(--text-primary); word-break: break-all; white-space: pre-wrap; flex: 1; }

/* Level colors */
.lvl-trace { color: var(--text-muted); }
.lvl-debug { color: var(--text-secondary); }
.lvl-info  { color: var(--accent); }
.lvl-warn  { color: var(--warning); }
.lvl-error { color: var(--danger); }

/* Row background tints */
.log-warn  { background: rgba(245,158,11,0.04); }
.log-error { background: rgba(255,77,106,0.05); }

/* Cursor */
.log-cursor {
  padding: 0.18rem 1rem;
  font-size: 0.90rem;
  color: var(--text-muted);
}
.cursor-blink {
  color: var(--accent);
  animation: blink 1s ease infinite;
}

/* New badge */
.new-badge {
  position: absolute;
  bottom: 0.75rem; left: 50%;
  transform: translateX(-50%);
  background: var(--accent);
  color: #000;
  border-radius: 1px;
  padding: 0.2rem 0.85rem;
  font-size: 0.94rem;
  font-family: var(--font-mono);
  font-weight: 700;
  cursor: pointer;
  box-shadow: 0 2px 12px rgba(124,58,237,0.40);
  animation: fadeSlideUp 0.15s ease both;
  user-select: none;
  z-index: 20;
  letter-spacing: 0.06em;
}
.new-badge:hover { background: var(--accent-hover); }
</style>
