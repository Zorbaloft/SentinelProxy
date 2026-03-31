import { createApp } from 'vue'
import { createRouter, createWebHistory } from 'vue-router'
import App from './App.vue'
import './style.css'

import LogsPage from './pages/LogsPage.vue'
import RulesPage from './pages/RulesPage.vue'
import IncidentsPage from './pages/IncidentsPage.vue'
import IPActionsPage from './pages/IPActionsPage.vue'

const routes = [
  { path: '/', redirect: '/logs' },
  { path: '/logs', component: LogsPage },
  { path: '/rules', component: RulesPage },
  { path: '/incidents', component: IncidentsPage },
  { path: '/ip-actions', component: IPActionsPage },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

const app = createApp(App)
app.use(router)
app.mount('#app')
