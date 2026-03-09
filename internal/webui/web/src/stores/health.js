import { defineStore } from 'pinia'
import { ref } from 'vue'
import axios from 'axios'

export const useHealthStore = defineStore('health', () => {
  const snapshot = ref(null) // null = not yet loaded

  async function fetchHealth() {
    try {
      const { data } = await axios.get('/api/v1/health')
      snapshot.value = data
    } catch (_) {
      // silently ignore — show stale data or loading state
    }
  }

  function applyHealthUpdate(data) {
    snapshot.value = data
  }

  return { snapshot, fetchHealth, applyHealthUpdate }
})
