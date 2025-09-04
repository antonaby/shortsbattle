import { defineStore } from "pinia";
import type {
  ThemeDetails,
  GameUpdate,
  Video,
  VoteValue,
  GameResult,
} from "../types/game";
import { computed, ref } from "vue";
import { useWSStore } from "./ws.store";
import { useRouter } from "vue-router";
import { useUIStore } from "./ui.store";
import type {
  ErrorContext,
  PublicationContext,
  SubscribedContext,
  UnsubscribedContext,
} from "centrifuge";
import { GamesAPI } from "../api/games";

export const useGameStore = defineStore("game", () => {
  const wsStore = useWSStore();
  const uiStore = useUIStore();
  const router = useRouter();

  const lastGameUpdate = ref<GameUpdate | null>(null);
  const remainingTimeMs = ref<number>(0);
  const theme = ref<ThemeDetails | null>(null);
  const selectedVideo = ref<Video | null>(null);
  const playerVideos = ref<Video[]>([]);
  const loadingPlayerVideos = ref<boolean>(false);
  const gameVideos = ref<Video[]>([]);
  const finalResult = ref<GameResult | null>(null);

  let gameId: number | null = null;
  let intervalId: number | null = null;

  const formattedTime = computed(() => {
    const totalSeconds = Math.max(0, Math.floor(remainingTimeMs.value / 1000));
    const minutes = Math.floor(totalSeconds / 60);
    const seconds = totalSeconds % 60;
    return `${minutes}:${seconds.toString().padStart(2, "0")}`;
  });

  function updateRemainingTime(upd: GameUpdate) {
    remainingTimeMs.value = upd.remaning_time_ms;
  }

  function navigateToGameState(upd: GameUpdate) {
    switch (upd.state) {
      case "lobby":
        router.push({ name: "game-lobby", params: { id: gameId } });
        break;
      case "submitting":
        router.push({ name: "game-submit", params: { id: gameId } });
        break;
      case "watching":
        router.push({ name: "game-watch", params: { id: gameId } });
        break;
      case "completed":
        router.push({ name: "game-complete", params: { id: gameId } });
        break;
    }
  }

  function setLastUpdate(upd: GameUpdate) {
    lastGameUpdate.value = upd;

    if (upd.theme && upd.msg_type == "details") {
      theme.value = upd.theme;
    }
    if (upd.result && upd.msg_type == "complete") {
      finalResult.value = upd.result;
    }
  }

  function handleSubscribed(ctx: SubscribedContext) {
    if (ctx.data) {
      var upd: GameUpdate = ctx.data;
      setLastUpdate(upd);

      if (intervalId) {
        clearInterval(intervalId);
      }

      updateRemainingTime(upd);
      intervalId = setInterval(() => {
        remainingTimeMs.value -= 1000;
      }, 1000);

      navigateToGameState(upd);
    } else {
      console.log("no subscription data");
    }
  }

  function handlePublication(ctx: PublicationContext) {
    var upd: GameUpdate = ctx.data;
    if (lastGameUpdate.value) {
      if (lastGameUpdate.value.state !== upd.state) {
        updateRemainingTime(upd);
        navigateToGameState(upd);
      }
      setLastUpdate(upd);
    }
  }

  function handleUnsubscribed(ctx: UnsubscribedContext) {
    if (ctx.reason == "permission denied") {
      leaveGame();
      uiStore.returnToHub();
    }
  }

  function hanldeError(ctx: ErrorContext) {
    uiStore.handleWsError(ctx.type, ctx.error);
  }

  function joinGame(openGameId: number) {
    gameId = openGameId;

    // TODO: handle sub
    wsStore.subscribe(`game_${gameId}`, {
      subscribed: handleSubscribed,
      publication: handlePublication,
      unsubscribed: handleUnsubscribed,
      error: hanldeError,
    });
  }

  async function reloadPlayerVideos() {
    loadingPlayerVideos.value = true;
    try {
      playerVideos.value = await GamesAPI.getMyVideos();
    } catch (error) {
      uiStore.handleNetworkError(error);
    } finally {
      loadingPlayerVideos.value = false;
    }
  }

  // TODO: show loading element
  async function selectVideo(videoId: number) {
    if (!gameId) {
      return;
    }

    try {
      const video = await GamesAPI.submitExistingVideo(gameId, videoId);
      selectedVideo.value = video;
    } catch (error) {
      uiStore.handleNetworkError(error);
    }
  }

  // TODO: show loading element
  async function newVideo(url: string) {
    if (!gameId) {
      return;
    }

    try {
      const video = await GamesAPI.submitNewVideo(gameId, url);
      selectedVideo.value = video;
    } catch (error) {
      uiStore.handleNetworkError(error);
    }
  }

  function unselectVideo() {
    selectedVideo.value = null;
  }

  async function loadVideosToWatch() {
    if (!gameId) {
      return;
    }

    try {
      const videos = await GamesAPI.fetchVideosToWatch(gameId);
      gameVideos.value = videos;
    } catch (error) {
      uiStore.handleNetworkError(error);
    }
  }

  async function voteForVideo(video: Video, value: VoteValue) {
    if (!gameId) {
      return;
    }

    try {
      await GamesAPI.voteForVideo(gameId, video.id, value);
    } catch (error) {
      uiStore.handleNetworkError(error);
    }
  }

  // TODO: handle error
  function handleVideoError(error: any) {
    console.log(`Video error: ${error}`)
  }

  // TODO: handle finished
  function handleFinished() {
    console.log("Player finished")
  }

  function leaveGame() {
    if (intervalId) {
      clearInterval(intervalId);
      intervalId = null;
    }

    if (gameId) {
      wsStore.unsubscribe(`game_${gameId}`);
    }

    gameId = null;
    lastGameUpdate.value = null;
    remainingTimeMs.value = 0;
    theme.value = null;
    playerVideos.value = [];
    selectedVideo.value = null;
    gameVideos.value = [];
    finalResult.value = null;
  }

  return {
    theme,
    lastGameUpdate,
    formattedTime,
    playerVideos,
    selectedVideo,
    gameVideos,
    finalResult,
    loadingPlayerVideos,
    joinGame,
    leaveGame,
    reloadPlayerVideos,
    selectVideo,
    newVideo,
    unselectVideo,
    loadVideosToWatch,
    voteForVideo,
    handleVideoError,
    handleFinished
  };
});
