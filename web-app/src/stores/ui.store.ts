import { defineStore } from "pinia";
import { useRouter } from "vue-router";
import { NetworkError } from "../types/errors";

export const useUIStore = defineStore("ui", () => {
  const router = useRouter();

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

  return { openGame, returnToHub, handleNetworkError };
});
