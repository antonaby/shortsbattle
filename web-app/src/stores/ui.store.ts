import { defineStore } from "pinia";
import { useRouter } from "vue-router";

export const useUIStore = defineStore("ui", () => {
  const router = useRouter();

  function openGame(gameId: number) {
    router.push({ name: "game-open", params: { id: gameId } });
  }

  function returnToHub() {
    router.replace({ name: "hub" });
  }

  return { openGame, returnToHub };
});
