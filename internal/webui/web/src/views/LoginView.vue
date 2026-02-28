<template>
  <div class="login-wrap">
    <div class="login-card">
      <div class="login-logo">◈</div>
      <h1 class="login-title">{{ $t('login.title') }}</h1>
      <p class="login-subtitle">{{ $t('login.subtitle') }}</p>

      <form class="login-form" @submit.prevent="submit">
        <div class="field">
          <label class="field-label">{{ $t('login.password') }}</label>
          <input
            v-model="password"
            type="password"
            class="field-input"
            autocomplete="current-password"
            autofocus
            :disabled="loading"
          />
        </div>

        <div v-if="error" class="login-error">{{ $t('login.error') }}</div>

        <button type="submit" class="btn-submit" :disabled="loading || !password">
          <span v-if="loading" class="spinner" />
          <span v-else>{{ $t('login.submit') }}</span>
        </button>
      </form>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useWebSocket } from '@/composables/useWebSocket'

const router = useRouter()
const auth = useAuthStore()
const { connect } = useWebSocket()

const password = ref('')
const loading = ref(false)
const error = ref(false)

async function submit() {
  if (!password.value) return
  loading.value = true
  error.value = false
  try {
    await auth.login(password.value)
    connect()
    router.push('/overview')
  } catch {
    error.value = true
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.login-wrap {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 100vh;
  background: var(--bg-primary);
  font-family: var(--font-ui);
}

.login-card {
  width: 100%;
  max-width: 380px;
  padding: 2.5rem;
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: calc(var(--radius) * 1.5);
  box-shadow: var(--shadow);
  animation: fadeUp 0.3s ease;
}

@keyframes fadeUp {
  from { opacity: 0; transform: translateY(12px); }
  to   { opacity: 1; transform: translateY(0); }
}

.login-logo {
  font-size: 2.5rem;
  color: var(--accent);
  text-align: center;
  margin-bottom: 0.75rem;
  animation: pulse 2s ease-in-out infinite;
}
@keyframes pulse {
  0%, 100% { opacity: 1; }
  50%       { opacity: 0.6; }
}

.login-title {
  text-align: center;
  font-size: 1.4rem;
  font-weight: 700;
  color: var(--text-primary);
  margin-bottom: 0.25rem;
}

.login-subtitle {
  text-align: center;
  font-size: 0.85rem;
  color: var(--text-secondary);
  margin-bottom: 2rem;
}

.login-form { display: flex; flex-direction: column; gap: 1rem; }

.field { display: flex; flex-direction: column; gap: 0.4rem; }

.field-label {
  font-size: 0.8rem;
  font-weight: 600;
  color: var(--text-secondary);
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.field-input {
  width: 100%;
  padding: 0.6rem 0.75rem;
  background: var(--bg-primary);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  color: var(--text-primary);
  font-size: 0.95rem;
  font-family: var(--font-mono);
  outline: none;
  transition: border-color var(--transition);
}
.field-input:focus { border-color: var(--accent); }

.login-error {
  background: rgba(248, 81, 73, 0.12);
  border: 1px solid var(--danger);
  color: var(--danger);
  border-radius: var(--radius);
  padding: 0.5rem 0.75rem;
  font-size: 0.85rem;
}

.btn-submit {
  width: 100%;
  padding: 0.65rem;
  background: var(--accent);
  color: #fff;
  border: none;
  border-radius: var(--radius);
  font-size: 0.95rem;
  font-weight: 600;
  cursor: pointer;
  transition: background var(--transition), opacity var(--transition);
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
}
.btn-submit:hover:not(:disabled) { background: var(--accent-hover); }
.btn-submit:disabled { opacity: 0.5; cursor: not-allowed; }

.spinner {
  width: 16px; height: 16px;
  border: 2px solid rgba(255,255,255,0.3);
  border-top-color: #fff;
  border-radius: 50%;
  animation: spin 0.6s linear infinite;
}
@keyframes spin { to { transform: rotate(360deg); } }
</style>
