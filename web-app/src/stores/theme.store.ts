import { GamesAPI } from "@/api/games";
import type { PlayerMode, Theme } from "@/types/game";
import { defineStore } from "pinia";
import { ref } from "vue";
import { useUIStore } from "./ui.store";
import { useGameStore } from "./game.store";

export const useThemeStore = defineStore("theme", () => {
  const theme = ref<Theme | undefined>(undefined);
  const joinGameLoader = ref<boolean>(false);
  const gameStore = useGameStore();
  const uiStore = useUIStore();

  async function loadTheme(themeId: number) {
    try {
      theme.value = await GamesAPI.fetchTheme(themeId);
    } catch (error) {
      uiStore.handleNetworkError(error);
    }
  }

  async function joinGame(themeId: number, mode: PlayerMode): Promise<void> {
    joinGameLoader.value = true;
    try {
      const game = await GamesAPI.joinGame(themeId, mode);
      await gameStore.joinGame(game.game_id);
    } catch (error) {
      uiStore.handleNetworkError(error);
    } finally {
      joinGameLoader.value = false;
    }
  }

  return {
    theme,
    joinGameLoader,
    loadTheme,
    joinGame,
  };
});
