import { defineStore } from "pinia";
import type {
  Theme,
  GameUpdate,
  Video,
  GameVideo,
  VoteValue,
  GameResult,
} from "../types/game";
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

  const remainingTimeMs = ref<number>(-1);

  const theme = ref<Theme | undefined>(undefined);
  const lastUpdate = ref<GameUpdate | undefined>(undefined);
  const playerVideo = ref<GameVideo | undefined>(undefined);
  const videosToWatch = ref<GameVideo[]>([]);
  const gameResult = ref<GameResult | undefined>(undefined);

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
    clearGameState();

    if (gameSub) {
      gameSub.unsubscribe();
    }
  }

  function clearGameState() {
    stopTimer();

    theme.value = undefined;
    lastUpdate.value = undefined;
    playerVideo.value = undefined;
    videosToWatch.value = [];
    gameResult.value = undefined;
  }

  function handleSubscribed(ctx: SubscribedContext) {
    clearGameState();

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
        stopTimer();
        uiStore.openGameLobby(upd.id);
        break;
      case "submit":
        playerVideo.value = undefined;
        startTimer(upd.remaining_ms);
        uiStore.openGameSubmit(upd.id);
        break;
      case "submit-complete":
        stopTimer();
        uiStore.openGameSubmit(upd.id);
        break;
      case "watch":
        videosToWatch.value = [];
        stopTimer();
        uiStore.openGameWatch(upd.id);
        break;
      case "watch-complete":
        videosToWatch.value = [];
        stopTimer();
        uiStore.openGameWatch(upd.id);
        break;
      case "complete":
        gameResult.value = undefined;
        stopTimer();
        uiStore.openGameComplete(upd.id);
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

  function stopTimer() {
    remainingTimeMs.value = -1;

    if (intervalId) {
      clearInterval(intervalId);
      intervalId = null;
    }
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

  async function loadVideosToWatch() {
    if (!lastUpdate.value) {
      return;
    }

    try {
      videosToWatch.value = await GamesAPI.fetchVideosToWatch(
        lastUpdate.value.id,
        lastUpdate.value.round
      );
    } catch (error) {
      uiStore.handleNetworkError(error);
    }
  }

  async function voteForVideo(gameVideo: GameVideo, value: VoteValue) {
    if (!lastUpdate.value) {
      return;
    }

    try {
      await GamesAPI.voteForVideo(gameVideo.id, value);
    } catch (error) {
      uiStore.handleNetworkError(error);
    }
  }

  async function videoError(gameVideo: GameVideo, error: any) {
    console.log(gameVideo, error);
    // TODO: send to the backend
  }

  async function loadGameResult() {
    if (!lastUpdate.value) {
      return;
    }

    try {
      gameResult.value = await GamesAPI.getGameResult(lastUpdate.value.id);
    } catch (error) {
      uiStore.handleNetworkError(error);
    }
  }

  return {
    theme,
    formattedTime,
    lastUpdate,
    playerVideo,
    videosToWatch,
    gameResult,
    joinGame,
    leaveGame,
    startTimer,
    submitVideo,
    submitNewVideo,
    loadVideosToWatch,
    voteForVideo,
    videoError,
    loadGameResult,
  };
});
