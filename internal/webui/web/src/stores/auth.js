import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import axios from 'axios'

export const useAuthStore = defineStore('auth', () => {
  const token = ref(localStorage.getItem('jwt') || '')
  const isAuthenticated = computed(() => !!token.value)

  async function login(password) {
    const { data } = await axios.post('/api/v1/login', { password })
    token.value = data.token
    localStorage.setItem('jwt', data.token)
    axios.defaults.headers.common['Authorization'] = `Bearer ${data.token}`
  }

  function logout() {
    token.value = ''
    localStorage.removeItem('jwt')
    delete axios.defaults.headers.common['Authorization']
  }

  function restore() {
    if (token.value) {
      axios.defaults.headers.common['Authorization'] = `Bearer ${token.value}`
    }
  }

  return { token, isAuthenticated, login, logout, restore }
})
