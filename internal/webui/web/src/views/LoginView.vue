<template>
  <div class="login-wrap">
    <!-- Grid background overlay -->
    <div class="login-bg" />

    <!-- Terminal window -->
    <div class="login-terminal" :class="{ 'terminal--shake': shaking }">
      <!-- Terminal chrome -->
      <div class="terminal-chrome">
        <div class="chrome-dots">
          <span class="cdot cdot--r" />
          <span class="cdot cdot--y" />
          <span class="cdot cdot--g" />
        </div>
        <span class="chrome-title">POLYTRADE // NEXUS TERMINAL v2.0</span>
        <span class="chrome-status" :class="connected ? 'st--on' : 'st--warn'">
          {{ connected ? '● SYS.ONLINE' : '● SYS.INIT' }}
        </span>
      </div>

      <!-- Boot sequence -->
      <div class="terminal-body">
        <div class="boot-lines">
          <div class="boot-line boot-1">◈ POLYTRADE NEXUS TERMINAL</div>
          <div class="boot-line boot-2">━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━</div>
          <div class="boot-line boot-3">SECURE AUTHENTICATION REQUIRED</div>
          <div class="boot-line boot-4">Establish connection to continue...</div>
        </div>

        <form class="auth-form" @submit.prevent="submit" autocomplete="off">
          <div class="prompt-row">
            <span class="prompt-glyph">$&gt;&nbsp;</span>
            <div class="input-wrap">
              <input
                ref="inputEl"
                v-model="password"
                type="password"
                class="auth-input"
                placeholder="enter access key"
                autocomplete="current-password"
                :disabled="loading"
                autofocus
              />
              <span class="cursor-blink" v-if="!password && !loading">|</span>
            </div>
          </div>

          <div v-if="error" class="auth-error">
            <span class="err-prefix">ERR:</span> AUTHENTICATION FAILED — invalid credentials
          </div>

          <div class="auth-actions">
            <button type="submit" class="auth-btn" :disabled="loading || !password">
              <span v-if="loading" class="spin">⟳</span>
              <span v-else>CONNECT &rarr;</span>
            </button>
          </div>
        </form>
      </div>
    </div>

    <!-- Corner decorations -->
    <div class="corner corner-tl" />
    <div class="corner corner-tr" />
    <div class="corner corner-bl" />
    <div class="corner corner-br" />
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
const shaking = ref(false)
const connected = ref(false)
const inputEl = ref(null)

async function submit() {
  if (!password.value) return
  loading.value = true
  error.value = false
  try {
    await auth.login(password.value)
    connected.value = true
    connect()
    setTimeout(() => router.push('/overview'), 200)
  } catch {
    error.value = true
    shaking.value = true
    password.value = ''
    setTimeout(() => { shaking.value = false }, 400)
    setTimeout(() => inputEl.value?.focus(), 100)
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.login-wrap {
  position: relative;
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 100vh;
  background: var(--bg-primary);
  font-family: var(--font-mono);
  overflow: hidden;
}

/* Background grid with radial vignette */
.login-bg {
  position: absolute;
  inset: 0;
  background:
    radial-gradient(ellipse 60% 60% at 50% 50%, rgba(124, 58, 237, 0.05) 0%, transparent 70%),
    linear-gradient(rgba(124, 58, 237, 0.03) 1px, transparent 1px),
    linear-gradient(90deg, rgba(124, 58, 237, 0.03) 1px, transparent 1px);
  background-size: 100% 100%, 32px 32px, 32px 32px;
}

/* Corner decorations */
.corner {
  position: absolute;
  width: 28px;
  height: 28px;
  pointer-events: none;
}
.corner-tl { top: 1.5rem; left: 1.5rem;  border-top: 1px solid var(--accent); border-left: 1px solid var(--accent); }
.corner-tr { top: 1.5rem; right: 1.5rem; border-top: 1px solid var(--accent); border-right: 1px solid var(--accent); }
.corner-bl { bottom: 1.5rem; left: 1.5rem;  border-bottom: 1px solid var(--accent); border-left: 1px solid var(--accent); }
.corner-br { bottom: 1.5rem; right: 1.5rem; border-bottom: 1px solid var(--accent); border-right: 1px solid var(--accent); }

/* Terminal window */
.login-terminal {
  position: relative;
  width: 100%;
  max-width: 520px;
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-top: 2px solid var(--accent);
  border-radius: var(--radius);
  background: rgba(255,255,255,0.04);
  backdrop-filter: blur(16px);
  -webkit-backdrop-filter: blur(16px);
  box-shadow: var(--shadow-lg), 0 0 60px rgba(124, 58, 237, 0.10);
  animation: fadeSlideUp 0.35s ease both;
  overflow: hidden;
}

.terminal--shake {
  animation: shake 0.35s ease both;
}

@keyframes shake {
  0%, 100% { transform: translateX(0); }
  20% { transform: translateX(-6px); }
  40% { transform: translateX(6px); }
  60% { transform: translateX(-4px); }
  80% { transform: translateX(4px); }
}

/* Scanline overlay on terminal */
.login-terminal::after {
  content: '';
  position: absolute;
  inset: 0;
  background: repeating-linear-gradient(
    0deg,
    transparent,
    transparent 2px,
    rgba(124, 58, 237, 0.012) 2px,
    rgba(124, 58, 237, 0.012) 4px
  );
  pointer-events: none;
  z-index: 1;
}

/* Terminal chrome bar */
.terminal-chrome {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.55rem 1rem;
  background: rgba(124, 58, 237, 0.04);
  border-bottom: 1px solid var(--border);
  position: relative;
  z-index: 2;
}

.chrome-dots {
  display: flex;
  gap: 0.3rem;
  flex-shrink: 0;
}
.cdot {
  width: 10px; height: 10px;
  border-radius: 50%;
}
.cdot--r { background: #ff5f57; }
.cdot--y { background: #ffbd2e; }
.cdot--g { background: #28ca41; }

.chrome-title {
  font-size: 1.00rem;
  letter-spacing: 0.10em;
  color: var(--text-secondary);
  flex: 1;
  text-transform: uppercase;
}

.chrome-status {
  font-size: 0.86rem;
  font-weight: 700;
  letter-spacing: 0.08em;
}
.st--on   { color: var(--success); }
.st--warn { color: var(--warning); }

/* Terminal body */
.terminal-body {
  padding: 1.5rem 1.5rem 1.75rem;
  position: relative;
  z-index: 2;
}

/* Boot lines with typewriter */
.boot-lines {
  display: flex;
  flex-direction: column;
  gap: 0.3rem;
  margin-bottom: 1.75rem;
}

.boot-line {
  font-size: 0.94rem;
  overflow: hidden;
  white-space: nowrap;
  opacity: 0;
  animation: typewriter 0.5s steps(40) both, fadeIn 0.1s ease both;
}

@keyframes fadeIn {
  from { opacity: 0; }
  to   { opacity: 1; }
}

.boot-1 {
  font-size: 1.08rem;
  font-weight: 700;
  color: var(--accent-bright);
  text-shadow: 0 0 12px rgba(124, 58, 237, 0.50);
  animation-delay: 0.05s;
}
.boot-2 {
  color: var(--text-muted);
  font-size: 0.94rem;
  animation-delay: 0.35s;
}
.boot-3 {
  color: var(--warning);
  font-weight: 600;
  font-size: 0.86rem;
  letter-spacing: 0.06em;
  animation-delay: 0.65s;
}
.boot-4 {
  color: var(--text-secondary);
  font-size: 0.86rem;
  animation-delay: 0.95s;
}

/* Auth form */
.auth-form {
  display: flex;
  flex-direction: column;
  gap: 1rem;
  animation: fadeSlideUp 0.3s ease 1.2s both;
}

.prompt-row {
  display: flex;
  align-items: center;
  background: var(--bg-input);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  padding: 0.55rem 0.75rem;
  transition: border-color var(--transition);
}

.prompt-row:focus-within {
  border-color: var(--accent);
  box-shadow: 0 0 0 1px rgba(124, 58, 237, 0.15);
}

.prompt-glyph {
  color: var(--accent);
  font-size: 1.00rem;
  flex-shrink: 0;
  user-select: none;
}

.input-wrap {
  flex: 1;
  position: relative;
  display: flex;
  align-items: center;
}

.auth-input {
  flex: 1;
  background: none;
  border: none;
  outline: none;
  color: var(--text-primary);
  font-family: var(--font-mono);
  font-size: 1.00rem;
  caret-color: var(--accent);
}

.auth-input::placeholder { color: var(--text-muted); }
.auth-input:disabled { opacity: 0.5; }

.cursor-blink {
  color: var(--accent);
  font-size: 1.05rem;
  animation: blink 1s ease infinite;
  pointer-events: none;
  margin-left: 1px;
}

/* Error */
.auth-error {
  font-size: 0.86rem;
  color: var(--danger);
  background: var(--danger-dim);
  border: 1px solid rgba(255, 77, 106, 0.25);
  border-radius: var(--radius);
  padding: 0.4rem 0.75rem;
  letter-spacing: 0.02em;
}

.err-prefix {
  font-weight: 700;
  color: var(--danger);
}

/* Actions */
.auth-actions {
  display: flex;
  justify-content: flex-end;
}

.auth-btn {
  background: transparent;
  border: 1px solid var(--accent);
  color: var(--accent);
  font-family: var(--font-mono);
  font-size: 0.90rem;
  font-weight: 700;
  letter-spacing: 0.08em;
  padding: 0.45rem 1.25rem;
  border-radius: var(--radius);
  cursor: pointer;
  transition: all var(--transition);
  display: flex;
  align-items: center;
  gap: 0.4rem;
}

.auth-btn:hover:not(:disabled) {
  background: var(--accent);
  color: #000;
  box-shadow: var(--accent-glow);
}

.auth-btn:disabled {
  opacity: 0.35;
  cursor: not-allowed;
}
</style>
