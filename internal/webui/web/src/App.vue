<template>
  <div v-if="auth.isAuthenticated" class="app-layout">
    <AppTopbar />
    <div class="app-body">
      <AppSidebar />
      <main class="app-content">
        <RouterView />
      </main>
    </div>
    <ToastContainer />
  </div>
  <RouterView v-else />
</template>

<script setup>
import { onMounted } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { useWebSocket } from '@/composables/useWebSocket'
import AppTopbar from '@/components/AppHeader.vue'
import AppSidebar from '@/components/AppSidebar.vue'
import ToastContainer from '@/components/ToastContainer.vue'

const auth = useAuthStore()
const { connect } = useWebSocket()

// Restore token synchronously so child onMounted hooks (e.g. MarketsView.fetchMarkets)
// already have the JWT header set before their first API call.
auth.restore()

onMounted(() => {
  if (auth.isAuthenticated) connect()
})
</script>

<style>
@import '@/assets/theme.css';

html, body { height: 100%; }
#app { height: 100%; }

.app-layout {
  display: flex;
  flex-direction: column;
  height: 100%;
  overflow: hidden;
}
.app-body {
  display: flex;
  flex: 1;
  overflow: hidden;
  min-height: 0;
}
.app-content {
  flex: 1;
  overflow-y: auto;
  padding: 1.5rem;
  background: var(--bg-primary);
}
</style>
