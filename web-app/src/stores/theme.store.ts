import { GamesAPI } from "@/api/games";
import type { Theme } from "@/types/game";
import { defineStore } from "pinia";
import { ref } from "vue";
import { useUIStore } from "./ui.store";

export const useThemeStore = defineStore("theme", () => {
  const theme = ref<Theme | undefined>(undefined);
  const joinGameLoader = ref<boolean>(false);
  const uiStore = useUIStore();

  async function joinGame(themeId: number): Promise<void> {
    joinGameLoader.value = true;
    // try {
    //   const game = await GamesAPI.joinGame(themeId);
    //   uiStore.openGame(game.game_id);
    // } catch (error) {
    //   uiStore.handleNetworkError(error);
    // } finally {
    //   joinGameLoader.value = false;
    // }
  }

  return {
    theme,
    joinGameLoader,
    joinGame
  };
});
