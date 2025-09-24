import { defineStore } from "pinia";
import { useRouter } from "vue-router";
import { NetworkError } from "../types/errors";

export const useUIStore = defineStore("ui", () => {
  const router = useRouter();

  function openTheme(themeId: number) {
    router.push({ name: "theme", params: { id: themeId } });
  }

  function returnToHub() {
    router.replace({ name: "hub" });
  }

  // TODO: prohibit moving back for all game stages
  function openGameLobby(gameId: number) {
    router.replace({ name: "game-lobby", params: { id: gameId } });
  }

  function openGameSubmit(gameId: number) {
    router.replace({ name: "game-submit", params: { id: gameId } });
  }

  function openGameWatch(gameId: number) {
    router.replace({ name: "game-watch", params: { id: gameId } });
  }

  function openGameComplete(gameId: number) {
    router.replace({ name: "game-complete", params: { id: gameId } });
  }

  // TODO: show error modal
  function handleNetworkError(error: any) {
    if (error instanceof NetworkError) {
      console.log(error.message);
    } else {
      console.log(error);
    }
  }

  function handleAuthError(error: any) {
    console.log(error);
  }

  // TODO: show error modal
  function handleWsError(type: string, error: any) {
    console.log(error);
  }

  return {
    openTheme,
    returnToHub,
    openGameLobby,
    openGameSubmit,
    openGameWatch,
    openGameComplete,
    handleNetworkError,
    handleWsError,
    handleAuthError,
  };
});
