import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { useApi } from '../composables/useApi.js'

export const useSettingsStore = defineStore('settings', () => {
  const api = useApi()
  const server = ref({})
  const local = ref({})
  const loading = ref(false)
  const saving = ref(false)

  const isDirty = computed(() => JSON.stringify(server.value) !== JSON.stringify(local.value))

  async function load() {
    loading.value = true
    try {
      const data = await api.getSettings()
      server.value = data
      local.value = JSON.parse(JSON.stringify(data))
    } finally {
      loading.value = false
    }
  }

  async function save() {
    saving.value = true
    try {
      // Flatten both objects and POST only changed non-secret keys
      const flat = _flatten(local.value)
      const serverFlat = _flatten(server.value)
      for (const [k, v] of Object.entries(flat)) {
        const strVal = String(v ?? '')
        if (strVal === '***') continue          // masked secret — skip
        if (String(serverFlat[k] ?? '') === strVal) continue // unchanged — skip
        await api.postSettings(k, strVal)
      }
      server.value = JSON.parse(JSON.stringify(local.value))
    } finally {
      saving.value = false
    }
  }

  function _flatten(obj, prefix = '', out = {}) {
    for (const [k, v] of Object.entries(obj || {})) {
      const key = prefix ? `${prefix}.${k}` : k
      if (v !== null && typeof v === 'object' && !Array.isArray(v)) {
        _flatten(v, key, out)
      } else {
        out[key] = v
      }
    }
    return out
  }

  function reset() {
    local.value = JSON.parse(JSON.stringify(server.value))
  }

  // set navigates dot-notation path to update nested value in local config
  function set(dotKey, value) {
    const parts = dotKey.split('.')
    let cur = local.value
    for (let i = 0; i < parts.length - 1; i++) {
      if (cur[parts[i]] === undefined || cur[parts[i]] === null || typeof cur[parts[i]] !== 'object') {
        cur[parts[i]] = {}
      }
      cur = cur[parts[i]]
    }
    cur[parts[parts.length - 1]] = value
  }

  return { server, local, loading, saving, isDirty, load, save, reset, set }
})
