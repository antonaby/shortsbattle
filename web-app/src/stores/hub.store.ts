import { defineStore } from "pinia";
import type { Theme } from "../types/game";
import { ref } from "vue";
import { useUIStore } from "./ui.store";
import { GamesAPI } from "../api/games";

export const useGameHubStore = defineStore("gamehub", () => {
  const themes = ref<Theme[]>([]);
  const loadingGame = ref<boolean>(false);
  const uiStore = useUIStore();

  async function fetchThemes() {
    try {
      themes.value = await GamesAPI.fetchThemes();
    } catch (error) {
      uiStore.handleNetworkError(error);
    }
  }

  async function joinGame(themeId: number): Promise<void> {
    loadingGame.value = true;
    try {
      const game = await GamesAPI.joinGame(themeId);
      uiStore.openGame(game.game_id);
    } catch (error) {
      uiStore.handleNetworkError(error);
    } finally {
      loadingGame.value = false;
    }
  }

  return { themes, loadingGame, fetchThemes, joinGame };
});
