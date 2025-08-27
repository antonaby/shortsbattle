import {
  createRouter,
  createWebHistory,
  type RouteRecordRaw,
} from "vue-router";
import Hub from "./views/Hub.vue";
import Game from "./views/Game.vue";
import LobbyView from "./components/game/LobbyView.vue";
import SubmitView from "./components/game/SubmitView.vue";
import OpenView from "./components/game/OpenView.vue";

const routes: RouteRecordRaw[] = [
  {
    path: "/",
    name: "hub",
    component: Hub,
  },
  {
    path: "/game/:id",
    name: "game",
    component: Game,
    children: [
      { path: "open", name: "game-open", component: OpenView },
      { path: "lobby", name: "game-lobby", component: LobbyView },
      { path: "submit", name: "game-submit", component: SubmitView },
    ],
  },
];

const router = createRouter({
  history: createWebHistory(),
  routes,
});

export default router;
