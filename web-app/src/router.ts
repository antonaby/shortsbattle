import {
  createRouter,
  createWebHistory,
  type RouteRecordRaw,
} from "vue-router";
import Hub from "@/pages/Hub.vue";
import Game from "@/pages/Game.vue";
import LobbyView from "@/components/game/subpages/LobbyView.vue";
import SubmitView from "@/components/game/subpages/SubmitView.vue";
import OpenView from "@/components/game/subpages/OpenView.vue";
import WatchView from "@/components/game/subpages/WatchView.vue";
import CompleteView from "@/components/game/subpages/CompleteView.vue";

let routes: RouteRecordRaw[] = [
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
      { path: "watch", name: "game-watch", component: WatchView },
      { path: "complete", name: "game-complete", component: CompleteView },
    ],
  },
];

if (import.meta.env.DEV) {
  routes = routes.concat([
    
  ]);
}

const router = createRouter({
  history: createWebHistory(),
  routes,
});

export default router;
