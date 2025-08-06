import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'
import GameSelect from './pages/GameSelect.vue'
import GameLobby from './pages/GameLobby.vue'
import GameWatch from './pages/GameWatch.vue'


const routes: RouteRecordRaw[] = [
  {
    path: '/',
    name: 'select',
    component: GameSelect,
  },{
    path: '/game',
    name: 'game',
    component: GameLobby
  },{
    path: '/watch',
    name: 'watch',
    component: GameWatch
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

export default router