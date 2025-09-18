import { defineStore } from "pinia";
import type { Theme, GameUpdate } from "../types/game";
import { computed, ref } from "vue";
import { useWSStore } from "./ws.store";
import { useUIStore } from "./ui.store";
import {
  type PublicationContext,
  type SubscribedContext,
  type Subscription,
} from "centrifuge";

const SUB_READY_TIMEOUT = 10000; // 10 sec

export const useGameStore = defineStore("game", () => {
  const wsStore = useWSStore();
  const uiStore = useUIStore();

  const theme = ref<Theme | undefined>(undefined);
  const lastUpdate = ref<GameUpdate | undefined>(undefined);
  const remainingTimeMs = ref<number>(-1);

  let gameSub: Subscription | null = null;
  let intervalId: number | null = null;

  const formattedTime = computed(() => {
    if (remainingTimeMs.value < 0) {
      return "00:00";
    } else if (remainingTimeMs.value == 0) {
      return "";
    }

    const totalSeconds = Math.max(0, Math.floor(remainingTimeMs.value / 1000));

    const minutes = Math.floor(totalSeconds / 60);
    const seconds = totalSeconds % 60;
    return `${minutes.toString().padStart(2, "0")}:${seconds
      .toString()
      .padStart(2, "0")}`;
  });

  async function joinGame(gameId: number): Promise<void> {
    gameSub = wsStore.subscribe(`game_${gameId}`, {
      subscribed: handleSubscribed,
      publication: handlePublication,
      unsubscribed: (ctx) => {
        if (ctx.reason == "permission denied") {
          leaveGame();
          uiStore.returnToHub();
        }
      },
      error: (ctx) => {
        leaveGame();
        uiStore.handleWsError(ctx.type, ctx.error);
      },
    });

    return gameSub.ready(SUB_READY_TIMEOUT);
  }

  function leaveGame() {
    theme.value = undefined;
    lastUpdate.value = undefined;
    remainingTimeMs.value = -1;

    if (gameSub) {
      gameSub.unsubscribe();
    }
    if (intervalId) {
      clearInterval(intervalId);
      intervalId = null;
    }
  }

  function handleSubscribed(ctx: SubscribedContext) {
    if (ctx.data) {
      var upd: GameUpdate = ctx.data;
      theme.value = upd.theme;
      lastUpdate.value = upd;
      switch (upd.stage) {
        case "lobby":
          startTimer(upd.remaining_ms);
          uiStore.openGameLobby(upd.id);
          break;
      }
    } else {
      uiStore.handleNetworkError("no subscription data");
    }
  }

  function handlePublication(ctx: PublicationContext) {
    var upd: GameUpdate = ctx.data;
    lastUpdate.value = upd;
  }

  function startTimer(newRemainingTimeMs: number) {
    if (intervalId) {
      clearInterval(intervalId);
    }

    remainingTimeMs.value = newRemainingTimeMs;

    intervalId = setInterval(() => {
      if (remainingTimeMs.value - 1000 > 0) {
        remainingTimeMs.value -= 1000;
      } else {
        remainingTimeMs.value = 0;
      }
    }, 1000);
  }

  return {
    theme,
    formattedTime,
    lastUpdate,
    joinGame,
    leaveGame,
    startTimer,
  };

  // const lastGameUpdate = ref<GameUpdate | null>(null);
  // const remainingTimeMs = ref<number>(0);
  // const theme = ref<Theme | null>(null);
  // const selectedVideo = ref<Video | null>(null);
  // const playerVideos = ref<Video[]>([]);
  // const loadingData = ref<boolean>(false);
  // const sendingRequest = ref<boolean>(false);
  // const gameVideos = ref<Video[]>([]);
  // const finalResult = ref<GameResult | null>(null);

  // let gameId: number | null = null;
  //

  // function updateRemainingTime(upd: GameUpdate) {
  //   remainingTimeMs.value = upd.remaining_ms;
  // }

  // function navigateToGameState(upd: GameUpdate) {
  //   switch (upd.stage) {
  //     case "lobby":
  //       router.push({ name: "game-lobby", params: { id: gameId } });
  //       break;
  //     case "submit":
  //       router.push({ name: "game-submit", params: { id: gameId } });
  //       break;
  //     case "watch":
  //       router.push({ name: "game-watch", params: { id: gameId } });
  //       break;
  //     case "complete":
  //       router.push({ name: "game-complete", params: { id: gameId } });
  //       break;
  //   }
  // }

  // function setLastUpdate(upd: GameUpdate) {
  //   lastGameUpdate.value = upd;

  //   if (upd.theme && upd.msg_type == "details") {
  //     theme.value = upd.theme;
  //   }
  // }

  // async function reloadPlayerVideos() {
  //   loadingData.value = true;
  //   try {
  //     playerVideos.value = await GamesAPI.getMyVideos();
  //   } catch (error) {
  //     uiStore.handleNetworkError(error);
  //   } finally {
  //     loadingData.value = false;
  //   }
  // }

  // // TODO: show loading element
  // async function selectVideo(videoId: number) {
  //   if (!gameId) {
  //     return;
  //   }

  //   sendingRequest.value = true;
  //   try {
  //     const video = await GamesAPI.submitExistingVideo(gameId, videoId, 1);
  //     selectedVideo.value = video;
  //   } catch (error) {
  //     uiStore.handleNetworkError(error);
  //   } finally {
  //     sendingRequest.value = false;
  //   }
  // }

  // // TODO: show loading element
  // async function newVideo(url: string) {
  //   if (!gameId) {
  //     return;
  //   }

  //   sendingRequest.value = true;
  //   try {
  //     const video = await GamesAPI.submitNewVideo(gameId, url, 1);
  //     selectedVideo.value = video;
  //   } catch (error) {
  //     uiStore.handleNetworkError(error);
  //   } finally {
  //     sendingRequest.value = false;
  //   }
  // }

  // function unselectVideo() {
  //   selectedVideo.value = null;
  // }

  // async function loadVideosToWatch() {
  //   if (!gameId) {
  //     return;
  //   }

  //   try {
  //     const videos = await GamesAPI.fetchVideosToWatch(gameId);
  //     gameVideos.value = videos;
  //   } catch (error) {
  //     uiStore.handleNetworkError(error);
  //   }
  // }

  // async function voteForVideo(video: Video, value: VoteValue) {
  //   if (!gameId) {
  //     return;
  //   }

  //   try {
  //     await GamesAPI.voteForVideo(gameId, video.id, value);
  //   } catch (error) {
  //     uiStore.handleNetworkError(error);
  //   }
  // }

  // // TODO: handle error
  // function handleVideoError(error: any) {
  //   console.log(`Video error: ${error}`)
  // }

  // // TODO: handle finished
  // function handleFinished() {
  //   console.log("Player finished")
  // }
});
