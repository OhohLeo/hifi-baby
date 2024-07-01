import './assets/main.css'

import { createApp } from 'vue'
import { createPinia } from 'pinia'
import PrimeVue from 'primevue/config'
import Aura from '@/presets/aura';

import App from './App.vue'

const pinia = createPinia()
const app = createApp(App)

app.use(pinia)
app.use(PrimeVue, {
  unstyled: true,
  pt: Aura  
})

app.mount('#app')
