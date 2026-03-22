<template>
  <div class="overlay" @click.self="$emit('close')">
    <div class="dialog">
      <div class="dialog-title">ADD WALLET</div>
      <div class="dialog-body">
        <div class="field-group">
          <label class="field-label">Private Key (64 hex chars)</label>
          <input v-model="privateKey" type="password" class="field-input" placeholder="Enter private key..." @input="error=''" />
          <span v-if="error" class="field-error">{{ error }}</span>
        </div>
      </div>
      <div class="dialog-actions">
        <button class="btn btn-ghost" @click="$emit('close')">Cancel</button>
        <button class="btn btn-primary" :disabled="importing || !privateKey" @click="doImport">
          {{ importing ? '...' : 'IMPORT' }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useApi } from '@/composables/useApi.js'
import { useAppStore } from '@/stores/app.js'

const emit = defineEmits(['close', 'added'])
const api = useApi()
const app = useAppStore()
const privateKey = ref('')
const error = ref('')
const importing = ref(false)

async function doImport() {
  if (!/^[0-9a-fA-F]{64}$/.test(privateKey.value)) {
    error.value = 'Private key must be 64 hex characters'
    return
  }
  importing.value = true
  try {
    const wallet = await api.addWallet(privateKey.value)
    app.toast('Wallet added', 'success')
    emit('added', wallet)
    emit('close')
  } catch (e) {
    error.value = e?.response?.data?.error || 'Failed to add wallet'
  } finally {
    importing.value = false
  }
}
</script>

<style scoped>
.overlay { position: fixed; inset: 0; background: var(--bg-overlay); display: flex; align-items: center; justify-content: center; z-index: 200; backdrop-filter: blur(4px); }
.dialog { background: var(--bg-card); border: 1px solid var(--border); border-top: 2px solid var(--accent); border-radius: var(--radius); padding: 1.5rem; min-width: 360px; box-shadow: var(--shadow-lg); }
.dialog-title { font-size: 11px; font-weight: 700; color: var(--accent); text-transform: uppercase; letter-spacing: 0.10em; margin-bottom: 1.25rem; }
.dialog-body { margin-bottom: 1.25rem; }
.field-group { display: flex; flex-direction: column; gap: 6px; }
.field-label { font-size: 10px; text-transform: uppercase; letter-spacing: 0.08em; color: var(--text-secondary); }
.field-error { font-size: 11px; color: var(--danger); }
.dialog-actions { display: flex; gap: 8px; justify-content: flex-end; }
.btn { display: inline-flex; align-items: center; padding: 6px 14px; border-radius: var(--radius); font-family: var(--font-mono); font-size: 13px; font-weight: 500; cursor: pointer; border: 1px solid transparent; transition: all var(--transition); }
.btn-ghost { background: none; border-color: var(--border); color: var(--text-secondary); }
.btn-primary { background: var(--accent); color: #fff; }
.btn:disabled { opacity: 0.4; cursor: not-allowed; }
</style>
