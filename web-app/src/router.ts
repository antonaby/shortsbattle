import { createRouter, createWebHistory } from 'vue-router'
import GameSelect from './pages/GameSelect.vue'
import GameLobby from './pages/GameLobby.vue'


const routes = [
  {
    path: '/',
    name: 'select',
    component: GameSelect,
  },
  {
    path: '/game',
    name: 'game',
    component: GameLobby
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

export default router