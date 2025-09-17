import { defineStore } from "pinia";
import { useRouter } from "vue-router";
import { NetworkError } from "../types/errors";

export const useUIStore = defineStore("ui", () => {
  const router = useRouter();

  function openTheme(themeId: number) {
    router.push({ name: "theme", params: { id: themeId } });
  }

  function openGame(gameId: number) {
    router.push({ name: "game-open", params: { id: gameId } });
  }

  function returnToHub() {
    router.replace({ name: "hub" });
  }

  // TODO: show error modal
  function handleNetworkError(error: any) {
    if (error instanceof NetworkError) {
      console.log(error.message);
    } else {
      console.log(error);
    }
  }

  // TODO: show error modal
  function handleWsError(type: string, error: any) {
    console.log(error);
  }

  return { openTheme, openGame, returnToHub, handleNetworkError, handleWsError };
});
