import { defineStore } from "pinia";
import type { Theme, GameUpdate } from "../types/game";
import { computed, ref } from "vue";
import { useWSStore } from "./ws.store";
import { useUIStore } from "./ui.store";
import {
  SubscriptionState,
  type ErrorContext,
  type PublicationContext,
  type SubscribedContext,
  type Subscription,
  type UnsubscribedContext,
} from "centrifuge";
import { GamesAPI } from "../api/games";
import { NetworkError } from "@/types/errors";

const SUB_READY_TIMEOUT = 10000; // 10 sec

export const useGameStore = defineStore("game", () => {
  const wsStore = useWSStore();
  const uiStore = useUIStore();

  const theme = ref<Theme | undefined>(undefined);
  let gameSub: Subscription | null = null;

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
        uiStore.handleWsError(ctx.type, ctx.error);
      },
    });

    return gameSub.ready(SUB_READY_TIMEOUT);
  }

  function leaveGame() {
    theme.value = undefined;
    if (gameSub && gameSub.state == SubscriptionState.Subscribed) {
      gameSub.unsubscribe();
    }

    // if (intervalId) {
    //   clearInterval(intervalId);
    //   intervalId = null;
    // }
    // if (gameId) {
    //   wsStore.unsubscribe(`game_${gameId}`);
    // }
    // gameId = null;
    // lastGameUpdate.value = null;
    // remainingTimeMs.value = 0;
    // theme.value = null;
    // playerVideos.value = [];
    // selectedVideo.value = null;
    // gameVideos.value = [];
    // finalResult.value = null;
  }

  function handleSubscribed(ctx: SubscribedContext) {
    if (ctx.data) {
      var upd: GameUpdate = ctx.data;
      theme.value = upd.theme;

      uiStore.openGameLobby(upd.id);

      // setLastUpdate(upd);

      // if (intervalId) {
      //   clearInterval(intervalId);
      // }

      // updateRemainingTime(upd);
      // intervalId = setInterval(() => {
      //   remainingTimeMs.value -= 1000;
      // }, 1000);

      // navigateToGameState(upd);
    } else {
      uiStore.handleNetworkError("no subscription data");
    }
  }

  function handlePublication(ctx: PublicationContext) {
    var upd: GameUpdate = ctx.data;
    // if (lastGameUpdate.value) {
    //   if (lastGameUpdate.value.stage !== upd.stage) {
    //     updateRemainingTime(upd);
    //     navigateToGameState(upd);
    //   }
    //   setLastUpdate(upd);
    // }
  }

  return {
    theme,
    joinGame,
    leaveGame,
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
  // let intervalId: number | null = null;

  // const formattedTime = computed(() => {
  //   const totalSeconds = Math.max(0, Math.floor(remainingTimeMs.value / 1000));
  //   const minutes = Math.floor(totalSeconds / 60);
  //   const seconds = totalSeconds % 60;
  //   return `${minutes}:${seconds.toString().padStart(2, "0")}`;
  // });

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
