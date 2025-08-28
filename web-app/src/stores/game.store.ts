import { defineStore } from "pinia";
import type { ThemeDetails, GameUpdate, Video } from "../types/game";
import { computed, ref } from "vue";
import { useWSStore } from "./ws.store";
import { useRouter } from "vue-router";
import { useUIStore } from "./ui.store";
import type {
  PublicationContext,
  SubscribedContext,
  UnsubscribedContext,
} from "centrifuge";
import axios from "axios";
import { NetworkError } from "../types/errors";

export const useGameStore = defineStore("game", () => {
  const wsStore = useWSStore();
  const uiStore = useUIStore();
  const router = useRouter();

  const lastGameUpdate = ref<GameUpdate | null>(null);
  const remainingTimeMs = ref<number>(0);
  const theme = ref<ThemeDetails | null>(null);
  const playerVideos = ref<Video[]>([]);

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
    if (upd.theme && upd.msg_type == "details") {
      theme.value = upd.theme;
    }
    lastGameUpdate.value = upd;
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

  function joinGame(openGameId: number) {
    gameId = openGameId;

    let sub = wsStore.subscribe(`game_${gameId}`, {
      subscribed: handleSubscribed,
      publication: handlePublication,
      unsubscribed: handleUnsubscribed,
    });
  }

  async function loadPlayerVideos() {
    try {
      // TODO: set proper player id
      const response = await axios.get<Video[]>(
        `${import.meta.env.VITE_BASE_URL}/api/v1/players/1/videos`
      );
      playerVideos.value = response.data;
    } catch (error) {
      // TODO: handle with UI Store
      throw new NetworkError("failed to join game", error);
    }
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
  }

  function selectVideo(videoId: number) {
    console.log(videoId);
  }

  function submitVideo(url: string) {
    console.log(url);
  }

  return {
    theme,
    lastGameUpdate,
    formattedTime,
    playerVideos,
    joinGame,
    leaveGame,
    loadPlayerVideos,
    selectVideo,
    submitVideo,
  };
});
