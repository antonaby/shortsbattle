import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'
import GameList from './views/Hub.vue'
import Game from './views/Game.vue'


const routes: RouteRecordRaw[] = [
  {
    path: '/',
    name: 'hub',
    component: GameList,
  },{
    path: '/game/:id',
    name: 'game',
    component: Game,
    props: { useView: false, viewName: 'submitting' }
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

export default router