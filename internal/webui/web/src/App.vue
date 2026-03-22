<template>
  <div v-if="auth.isAuthenticated" class="app-shell">
    <AppSidebar />
    <div class="app-main">
      <AppHeader />
      <main class="app-content">
        <RouterView />
      </main>
    </div>
  </div>
  <RouterView v-else />
  <ToastContainer />
</template>

<script setup>
import { onMounted } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { useWebSocket } from '@/composables/useWebSocket'
import AppSidebar from '@/components/layout/AppSidebar.vue'
import AppHeader from '@/components/AppHeader.vue'
import ToastContainer from '@/components/ToastContainer.vue'

const auth = useAuthStore()
const { connect } = useWebSocket()

// Restore token synchronously so child onMounted hooks already have JWT set.
auth.restore()

onMounted(() => {
  if (auth.isAuthenticated) connect()
})
</script>

<style>
@import '@/assets/theme.css';

html, body { height: 100%; }
#app { height: 100%; }

.app-shell {
  display: flex;
  height: 100%;
  overflow: hidden;
}

.app-main {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  min-width: 0;
}

.app-content {
  flex: 1;
  overflow-y: auto;
  padding: 20px 24px;
  background: var(--bg-primary);
  min-height: 0;
}
</style>
