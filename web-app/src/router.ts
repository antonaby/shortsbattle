import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'
import GameList from './pages/GameList.vue'
import Game from './pages/Game.vue'
import GameWatch from './pages/GameWatch.vue'


const routes: RouteRecordRaw[] = [
  {
    path: '/',
    name: 'list',
    component: GameList,
  },{
    path: '/game',
    name: 'game',
    component: Game
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