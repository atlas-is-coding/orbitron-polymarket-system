<template>
  <div class="overlay" @click.self="$emit('close')">
    <div class="dialog">
      <div class="dialog-title">CONFIGURE — {{ strategy?.name }}</div>
      <div class="dialog-body">
        <template v-if="params && Object.keys(params).length">
          <div v-for="(val, key) in params" :key="key" class="field-group">
            <label class="field-label">{{ key.replace(/_/g,' ').toUpperCase() }}</label>
            <label v-if="typeof val === 'boolean'" class="toggle">
              <input type="checkbox" v-model="localParams[key]" />
              <span class="toggle-track"><span class="toggle-thumb" /></span>
            </label>
            <input v-else-if="typeof val === 'number'" v-model.number="localParams[key]" type="number" class="field-input" />
            <input v-else v-model="localParams[key]" type="text" class="field-input" />
          </div>
        </template>
        <div v-else class="empty-msg">No configurable parameters</div>
      </div>
      <div class="dialog-actions">
        <button class="btn btn-ghost" @click="$emit('close')">Cancel</button>
        <button class="btn btn-primary" :disabled="saving" @click="save">{{ saving ? '...' : 'SAVE' }}</button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { useApi } from '@/composables/useApi.js'
import { useAppStore } from '@/stores/app.js'

const props = defineProps({ strategy: Object })
const emit = defineEmits(['close'])
const api = useApi()
const app = useAppStore()
const saving = ref(false)

const params = computed(() => props.strategy?.params || {})
const localParams = ref(JSON.parse(JSON.stringify(params.value)))

async function save() {
  saving.value = true
  try {
    await api.saveStrategyConfig(props.strategy.name, localParams.value)
    app.toast('Strategy config saved', 'success')
    emit('close')
  } catch (e) { app.toast(e?.response?.data?.error || 'Save failed', 'error') }
  saving.value = false
}
</script>

<style scoped>
.overlay { position: fixed; inset: 0; background: var(--bg-overlay); display: flex; align-items: center; justify-content: center; z-index: 200; backdrop-filter: blur(4px); }
.dialog { background: var(--bg-card); border: 1px solid var(--border); border-top: 2px solid var(--accent); border-radius: var(--radius); padding: 1.5rem; min-width: 360px; max-width: 500px; max-height: 80vh; overflow-y: auto; box-shadow: var(--shadow-lg); }
.dialog-title { font-size: 11px; font-weight: 700; color: var(--accent); text-transform: uppercase; letter-spacing: 0.10em; margin-bottom: 1.25rem; }
.dialog-body { display: flex; flex-direction: column; gap: 14px; margin-bottom: 1.25rem; }
.field-group { display: flex; flex-direction: column; gap: 5px; }
.field-label { font-size: 10px; text-transform: uppercase; letter-spacing: 0.08em; color: var(--text-secondary); }
.field-input { background: var(--bg-input, rgba(0,0,0,0.3)); border: 1px solid var(--border); border-radius: var(--radius); color: var(--text-primary); padding: 6px 10px; font-family: var(--font-mono); font-size: 13px; outline: none; transition: border-color var(--transition); }
.field-input:focus { border-color: var(--accent); }
.empty-msg { font-size: 12px; color: var(--text-muted); }
.dialog-actions { display: flex; gap: 8px; justify-content: flex-end; }
.btn { display: inline-flex; align-items: center; padding: 6px 14px; border-radius: var(--radius); font-family: var(--font-mono); font-size: 13px; font-weight: 500; cursor: pointer; border: 1px solid transparent; transition: all var(--transition); }
.btn-ghost { background: none; border-color: var(--border); color: var(--text-secondary); }
.btn-primary { background: var(--accent); color: #fff; }
.btn:disabled { opacity: 0.4; cursor: not-allowed; }
.toggle { display: inline-flex; align-items: center; cursor: pointer; }
.toggle input { display: none; }
.toggle-track { width: 36px; height: 18px; background: var(--border); border-radius: 1px; position: relative; transition: background var(--transition); }
.toggle input:checked ~ .toggle-track { background: var(--accent); }
.toggle-thumb { position: absolute; width: 12px; height: 12px; background: var(--text-muted); border-radius: 1px; top: 3px; left: 3px; transition: left var(--transition), background var(--transition); }
.toggle input:checked ~ .toggle-track .toggle-thumb { left: 21px; background: #fff; }
</style>
