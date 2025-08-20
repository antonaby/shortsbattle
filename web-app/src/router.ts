import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'
import GameList from './pages/GameList.vue'
import Game from './pages/Game.vue'


const routes: RouteRecordRaw[] = [
  {
    path: '/',
    name: 'home',
    component: GameList,
  },{
    path: '/game/:id',
    name: 'game',
    component: Game,
    props: { useView: true, viewName: 'submitting' }
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

export default router