import { defineStore } from "pinia";
import type { Theme, GameUpdate, Video, GameVideo } from "../types/game";
import { computed, ref } from "vue";
import { useWSStore } from "./ws.store";
import { useUIStore } from "./ui.store";
import {
  type PublicationContext,
  type SubscribedContext,
  type Subscription,
} from "centrifuge";
import { GamesAPI } from "@/api/games";

const SUB_READY_TIMEOUT = 10000; // 10 sec

export const useGameStore = defineStore("game", () => {
  const wsStore = useWSStore();
  const uiStore = useUIStore();

  const theme = ref<Theme | undefined>(undefined);
  const lastUpdate = ref<GameUpdate | undefined>(undefined);
  const remainingTimeMs = ref<number>(-1);
  const playerVideo = ref<GameVideo | undefined>(undefined); 

  let gameSub: Subscription | null = null;
  let intervalId: number | null = null;

  const formattedTime = computed(() => {
    if (remainingTimeMs.value <= 0) {
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
    playerVideo.value = undefined;
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
      if (!upd.theme) {
        uiStore.handleNetworkError("no subscription data");
        return;
      }
      theme.value = upd.theme;
      handleGameUpdate(upd);
    } else {
      uiStore.handleNetworkError("no subscription data");
    }
  }

  function handlePublication(ctx: PublicationContext) {
    var upd: GameUpdate = ctx.data;
    handleGameUpdate(upd);
  }

  function handleGameUpdate(upd: GameUpdate) {
    lastUpdate.value = upd;
    switch (upd.stage) {
      case "lobby":
        startTimer(upd.remaining_ms);
        uiStore.openGameLobby(upd.id);
        break;
      case "lobby-full":
        remainingTimeMs.value = -1;
        uiStore.openGameLobby(upd.id);
        break;
      case "submit":
        startTimer(upd.remaining_ms);
        uiStore.openGameSubmit(upd.id);
        break;
      case "submit-complete":
        remainingTimeMs.value = -1;
        uiStore.openGameSubmit(upd.id);
        break;
    }
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

  async function submitVideo(video: Video) {
    if (!lastUpdate.value) {
      return;
    }

    try {
      playerVideo.value = await GamesAPI.submitExistingVideo(
        lastUpdate.value.id,
        video.id,
        lastUpdate.value.round
      );
    } catch (error) {
      uiStore.handleNetworkError(error);
    }
  }

  async function submitNewVideo(url: string) {
    if (!lastUpdate.value) {
      return;
    }

    try {
      playerVideo.value = await GamesAPI.submitNewVideo(
        lastUpdate.value.id,
        url,
        lastUpdate.value.round
      );
    } catch (error) {
      uiStore.handleNetworkError(error);
    }
  }

  return {
    theme,
    formattedTime,
    lastUpdate,
    playerVideo,
    joinGame,
    leaveGame,
    startTimer,
    submitVideo,
    submitNewVideo
  };

  // // TODO: show loading element

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
