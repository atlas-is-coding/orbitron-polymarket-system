<template>
  <div class="logs-view">
    <!-- Toolbar -->
    <div class="toolbar anim-in">
      <div class="level-btns">
        <button v-for="lvl in ['DEBUG','INFO','WARN','ERROR']" :key="lvl"
          class="lvl-btn" :class="[`lvl-${lvl.toLowerCase()}`, { off: !store.levels[lvl] }]"
          @click="store.levels[lvl] = !store.levels[lvl]">
          {{ lvl }}
        </button>
      </div>
      <input v-model="store.searchQuery" class="field-input srch" placeholder="Regex search..." />
      <select v-model="store.sourceFilter" class="field-input src-sel">
        <option value="">All sources</option>
        <option v-for="s in sources" :key="s" :value="s">{{ s }}</option>
      </select>
      <button class="btn" :class="store.paused ? 'btn-success' : 'btn-ghost'" @click="store.paused = !store.paused">
        {{ store.paused ? 'RESUME' : 'PAUSE' }}
      </button>
      <button class="btn btn-ghost" @click="store.exportLogs()">EXPORT</button>
      <button class="btn btn-ghost danger-x" @click="store.clear()">CLEAR</button>
    </div>

    <!-- Main area -->
    <div class="logs-main">
      <!-- Log area -->
      <div class="log-area" ref="logEl">
        <div v-if="store.filtered.length === 0" class="empty-state">No log lines</div>
        <div
          v-for="(line, i) in store.filtered"
          :key="i"
          class="log-line"
          :class="{ selected: selectedLine === line }"
          @click="selectedLine = line"
        >
          <span class="log-ts mono">{{ fmtTs(line.ts || line.timestamp) }}</span>
          <span class="log-lvl" :class="`ll-${(line.level||'INFO').toLowerCase()}`">{{ (line.level||'INFO').padEnd(5) }}</span>
          <span class="log-src muted-txt">{{ line.source }}</span>
          <span class="log-msg" v-html="highlight(line.message || line.msg || '')"></span>
        </div>
      </div>

      <!-- Detail sidebar -->
      <div v-if="selectedLine" class="log-sidebar">
        <div class="sidebar-hdr">
          <span>LOG DETAIL</span>
          <button class="close-btn" @click="selectedLine = null">✕</button>
        </div>
        <div class="sidebar-body">
          <div class="detail-row"><span class="dl">Time</span><span class="dv mono">{{ fmtTs(selectedLine.ts || selectedLine.timestamp) }}</span></div>
          <div class="detail-row"><span class="dl">Level</span><span class="dv" :class="`ll-${(selectedLine.level||'info').toLowerCase()}`">{{ selectedLine.level }}</span></div>
          <div class="detail-row"><span class="dl">Source</span><span class="dv mono">{{ selectedLine.source }}</span></div>
          <div class="detail-row col">
            <span class="dl">Message</span>
            <pre class="detail-raw">{{ selectedLine.message || selectedLine.msg }}</pre>
          </div>
          <div v-if="selectedLine.raw" class="detail-row col">
            <span class="dl">Raw</span>
            <pre class="detail-raw">{{ selectedLine.raw }}</pre>
          </div>
        </div>
      </div>
    </div>

    <!-- Stats bar -->
    <div class="stats-bar anim-in">
      <span v-for="lvl in ['DEBUG','INFO','WARN','ERROR']" :key="lvl" class="stat-count" :class="`ll-${lvl.toLowerCase()}`">
        {{ lvl }}: {{ lvlCount(lvl) }}
      </span>
      <span class="stat-total">Total: {{ store.filtered.length }}</span>
      <label class="follow-toggle">
        <input type="checkbox" v-model="store.follow" />
        <span>FOLLOW</span>
      </label>
      <span v-if="store.paused" class="paused-badge">PAUSED</span>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch, nextTick, onMounted, onUnmounted } from 'vue'
import { useLogsStore } from '@/stores/logs'
import { useAppStore } from '@/stores/app'
import { useApi } from '@/composables/useApi'

const store = useLogsStore()
const app = useAppStore()
const api = useApi()
const logEl = ref(null)
const selectedLine = ref(null)

// Load initial logs from API
onMounted(async () => {
  if (!store.buffer.length) {
    try {
      const data = await api.getLogs()
      const lines = Array.isArray(data) ? data : (data.logs || [])
      lines.forEach(l => store.push({
        level: l.level || 'INFO',
        source: l.source || l.component || '',
        message: l.message || l.msg || '',
        ts: l.timestamp || l.ts || Date.now(),
        raw: JSON.stringify(l),
      }))
    } catch {}
  }
})

// Pipe app store log events to logsStore
const stopWatch = watch(() => app.logs, (logs) => {
  if (Array.isArray(logs) && logs.length) {
    const last = logs[logs.length - 1]
    if (last) store.push({
      level: last.level || 'INFO',
      source: last.source || last.component || '',
      message: last.message || last.msg || '',
      ts: last.timestamp || last.ts || Date.now(),
      raw: JSON.stringify(last),
    })
  }
}, { deep: false })
onUnmounted(() => stopWatch())

// Auto-scroll when following
watch(() => store.filtered.length, async () => {
  if (store.follow && logEl.value) {
    await nextTick()
    logEl.value.scrollTop = logEl.value.scrollHeight
  }
})

const sources = computed(() => {
  const s = new Set(store.buffer.map(l => l.source).filter(Boolean))
  return [...s].sort()
})

function lvlCount(lvl) { return store.buffer.filter(l => l.level === lvl).length }
function fmtTs(t) {
  if (!t) return ''
  try {
    const d = typeof t === 'number' ? new Date(t) : new Date(t)
    return d.toLocaleTimeString(undefined, { hour: '2-digit', minute: '2-digit', second: '2-digit' })
  } catch { return String(t) }
}

function highlight(msg) {
  if (!msg) return ''
  return msg
    .replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;')
    .replace(/("[\w_]+"):/g, '<span class="hl-key">$1</span>:')
    .replace(/:(\s*"[^"]*")/g, ':<span class="hl-val">$1</span>')
    .replace(/\b(\d+\.?\d*)\b/g, '<span class="hl-num">$1</span>')
    .replace(/\b(error|failed|fatal)\b/gi, '<span class="hl-err">$1</span>')
}
</script>

<style scoped>
.logs-view { display: flex; flex-direction: column; gap: 10px; height: calc(100vh - 120px); }

.toolbar { display: flex; align-items: center; gap: 8px; flex-wrap: wrap; }
.level-btns { display: flex; gap: 4px; }
.lvl-btn { padding: 3px 10px; border-radius: var(--radius); border: 1px solid; background: none; cursor: pointer; font-family: var(--font-mono); font-size: 11px; font-weight: 700; letter-spacing: 0.05em; transition: opacity 0.15s; }
.lvl-btn.off { opacity: 0.3; }
.lvl-debug { color: var(--text-secondary); border-color: var(--border); }
.lvl-info  { color: var(--success); border-color: rgba(16,217,148,0.40); }
.lvl-warn  { color: var(--warning); border-color: rgba(245,158,11,0.40); }
.lvl-error { color: var(--danger); border-color: rgba(255,77,106,0.40); }
.srch { flex: 1; min-width: 150px; max-width: 240px; }
.src-sel { width: 130px; }

.logs-main { flex: 1; display: flex; gap: 10px; overflow: hidden; }

.log-area {
  flex: 1; overflow-y: auto; background: var(--bg-card); border: 1px solid var(--border);
  border-radius: var(--radius); font-family: var(--font-mono); font-size: 12px;
}
.log-line {
  display: flex; gap: 8px; align-items: baseline;
  padding: 2px 10px; border-bottom: 1px solid rgba(255,255,255,0.02);
  cursor: pointer; line-height: 1.6;
}
.log-line:hover { background: var(--bg-hover); }
.log-line.selected { background: rgba(124,58,237,0.10); }

.log-ts { font-size: 10px; color: var(--text-muted); flex-shrink: 0; }
.log-lvl { font-size: 10px; font-weight: 700; flex-shrink: 0; width: 40px; }
.log-src { font-size: 10px; flex-shrink: 0; max-width: 100px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.log-msg { font-size: 11px; flex: 1; word-break: break-word; }

.ll-debug { color: var(--text-secondary); }
.ll-info  { color: var(--success); }
.ll-warn  { color: var(--warning); }
.ll-error { color: var(--danger); }

:deep(.hl-key) { color: var(--accent-bright); }
:deep(.hl-val) { color: var(--price-bright); }
:deep(.hl-num) { color: var(--success); }
:deep(.hl-err) { color: var(--danger); font-weight: 700; }

.muted-txt { color: var(--text-secondary); }
.mono { font-family: var(--font-mono); }

.log-sidebar { width: 320px; flex-shrink: 0; background: var(--bg-card); border: 1px solid var(--border); border-radius: var(--radius); display: flex; flex-direction: column; overflow: hidden; }
.sidebar-hdr { padding: 8px 12px; font-size: 10px; text-transform: uppercase; letter-spacing: 0.10em; color: var(--accent); border-bottom: 1px solid var(--border); background: rgba(124,58,237,0.03); display: flex; align-items: center; justify-content: space-between; }
.close-btn { background: none; border: none; color: var(--text-secondary); cursor: pointer; font-size: 12px; }
.sidebar-body { flex: 1; overflow-y: auto; padding: 12px; display: flex; flex-direction: column; gap: 10px; }
.detail-row { display: flex; gap: 8px; align-items: flex-start; font-size: 12px; }
.detail-row.col { flex-direction: column; gap: 4px; }
.dl { font-size: 10px; text-transform: uppercase; letter-spacing: 0.08em; color: var(--text-secondary); flex-shrink: 0; width: 60px; }
.dv { color: var(--text-primary); font-family: var(--font-mono); font-size: 11px; word-break: break-all; }
.detail-raw { font-size: 11px; background: var(--bg-primary); border: 1px solid var(--border); border-radius: var(--radius); padding: 8px; white-space: pre-wrap; word-break: break-all; color: var(--text-primary); font-family: var(--font-mono); max-height: 200px; overflow-y: auto; margin: 0; }

.stats-bar { display: flex; align-items: center; gap: 16px; padding: 6px 12px; background: var(--bg-card); border: 1px solid var(--border); border-radius: var(--radius); font-size: 11px; }
.stat-count { font-weight: 700; font-family: var(--font-mono); }
.stat-total { color: var(--text-secondary); font-family: var(--font-mono); margin-left: auto; }
.follow-toggle { display: flex; align-items: center; gap: 6px; cursor: pointer; font-size: 11px; color: var(--text-secondary); }
.follow-toggle input { accent-color: var(--accent); }
.paused-badge { font-size: 10px; font-weight: 700; padding: 2px 8px; border-radius: 3px; background: rgba(245,158,11,0.12); color: var(--warning); border: 1px solid rgba(245,158,11,0.30); }

.empty-state { padding: 2rem; text-align: center; color: var(--text-muted); font-size: 12px; }

.btn { display: inline-flex; align-items: center; padding: 4px 10px; border-radius: var(--radius); font-family: var(--font-mono); font-size: 11px; font-weight: 500; cursor: pointer; border: 1px solid transparent; transition: all var(--transition); }
.btn-ghost   { background: none; border-color: var(--border); color: var(--text-secondary); }
.btn-ghost:hover { background: var(--bg-hover); color: var(--text-primary); }
.btn-success { background: var(--success-dim); color: var(--success); border-color: rgba(16,217,148,0.40); }
.danger-x { color: var(--danger) !important; border-color: rgba(255,77,106,0.30) !important; }
</style>
