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
import VideoPlayerWrapper from "@/components/game/players/VideoPlayerWrapper.vue";
import Theme from "@/pages/Theme.vue";
import OpenDevWrapper from "@/components/dev/OpenDevWrapper.vue";
import LobbyDevWrapper from "@/components/dev/LobbyDevWrapper.vue";
import SubmitDevWrapper from "@/components/dev/SubmitDevWrapper.vue";
import WatchDevWrapper from "@/components/dev/WatchDevWrapper.vue";

let routes: RouteRecordRaw[] = [
  {
    path: "/",
    name: "hub",
    component: Hub,
  },
  {
    path: "/theme/:id",
    name: "theme",
    component: Theme,
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
    {
      path: "/dev/game/open",
      component: OpenDevWrapper,
    },
    {
      path: "/dev/game/lobby",
      component: LobbyDevWrapper,
    },
    {
      path: "/dev/game/submit",
      component: SubmitDevWrapper,
    },
    {
      path: "/dev/game/watch",
      component: WatchDevWrapper,
    },
    {
      path: "/dev/player",
      component: VideoPlayerWrapper,
      props: {
        videos: [
          {
            id: 0,
            video: {
              video_url: "https://www.youtube.com/shorts/wve9udh3qM4",
            },
          },
          {
            id: 1,
            video: {
              video_url: "https://www.youtube.com/shorts/Z5Jlxowr4IM",
            },
          },
          {
            id: 2,
            video: {
              video_url:
                "https://www.tiktok.com/@br0pics/video/7540625283531967776",
            },
          },
          {
            id: 3,
            video: {
              video_url:
                "https://www.tiktok.com/@br0pics/video/7540986521684380960",
            },
          },
          {
            id: 4,
            video: {
              video_url: "https://www.youtube.com/shorts/XUretomgAAA",
            },
          },
        ],
      },
    },
  ]);
}

const router = createRouter({
  history: createWebHistory(),
  routes,
});

export default router;
