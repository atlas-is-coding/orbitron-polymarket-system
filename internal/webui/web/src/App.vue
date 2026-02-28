<template>
  <div class="app-root">
    <AppHeader v-if="auth.isAuthenticated" />
    <main class="app-main">
      <RouterView />
    </main>
  </div>
</template>

<script setup>
import { onMounted } from 'vue'
import AppHeader from '@/components/AppHeader.vue'
import { useAuthStore } from '@/stores/auth'
import { useWebSocket } from '@/composables/useWebSocket'

const auth = useAuthStore()
const { connect } = useWebSocket()

onMounted(() => {
  auth.restore()
  if (auth.isAuthenticated) connect()
})
</script>

<style>
@import '@/assets/theme.css';

*, *::before, *::after { box-sizing: border-box; margin: 0; padding: 0; }

html { font-family: var(--font-ui); }

body {
  background: var(--bg-primary);
  color: var(--text-primary);
  min-height: 100vh;
  transition: background var(--transition), color var(--transition);
}

.app-root { display: flex; flex-direction: column; min-height: 100vh; }
.app-main  { flex: 1; padding: 1.5rem; max-width: 1400px; width: 100%; margin: 0 auto; }
</style>
