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
      // Use saveConfig if available, otherwise fall back to postSettings
      if (api.saveConfig) {
        await api.saveConfig(local.value)
      } else {
        await api.postSettings(local.value)
      }
      server.value = JSON.parse(JSON.stringify(local.value))
    } finally {
      saving.value = false
    }
  }

  function reset() {
    local.value = JSON.parse(JSON.stringify(server.value))
  }

  function set(key, value) {
    local.value[key] = value
  }

  return { server, local, loading, saving, isDirty, load, save, reset, set }
})
