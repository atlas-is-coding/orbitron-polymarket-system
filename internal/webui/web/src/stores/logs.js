import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

const MAX_LINES = 5000

export const useLogsStore = defineStore('logs', () => {
  const buffer = ref([])
  const paused = ref(false)
  const follow = ref(true)
  const levels = ref({ DEBUG: true, INFO: true, WARN: true, ERROR: true })
  const sourceFilter = ref('')
  const searchQuery = ref('')

  const filtered = computed(() => {
    let lines = buffer.value.filter(l => levels.value[l.level])
    if (sourceFilter.value) lines = lines.filter(l => l.source === sourceFilter.value)
    if (searchQuery.value) {
      try {
        const re = new RegExp(searchQuery.value, 'i')
        lines = lines.filter(l => re.test(l.raw))
      } catch {}
    }
    return lines
  })

  function push(line) {
    if (paused.value) return
    buffer.value.push(line)
    if (buffer.value.length > MAX_LINES) buffer.value.shift()
  }

  function clear() { buffer.value = [] }

  function exportLogs() {
    const text = filtered.value.map(l => l.raw).join('\n')
    const a = document.createElement('a')
    a.href = URL.createObjectURL(new Blob([text], { type: 'text/plain' }))
    a.download = `logs-${Date.now()}.log`
    a.click()
  }

  return { buffer, paused, follow, levels, sourceFilter, searchQuery, filtered, push, clear, exportLogs }
})
