import { defineStore } from "pinia";
import type { GameJoined, Theme } from "../types/game";
import axios from "axios";
import { ref } from "vue";
import { NetworkError } from "../types/errors";
import { useUIStore } from "./ui.store";

export const useGameHubStore = defineStore("gamehub", () => {
  const themes = ref<Theme[]>([]);
  const loadingGame = ref<boolean>(false);
  const uiStore = useUIStore();

  async function fetchThemes() {
    try {
      const response = await axios.get<Theme[]>(
        `${import.meta.env.VITE_BASE_URL}/api/v1/themes`
      );
      themes.value = response.data;
    } catch (error) {
      // TODO: handle with UI Store
      throw new NetworkError("failed to fetch game themes", error);
    }
  }

  async function joinGame(themeId: number): Promise<void> {
    loadingGame.value = true;
    try {
      const response = await axios.put<GameJoined>(
        `${import.meta.env.VITE_BASE_URL}/api/v1/games/join`,
        {
          theme_id: themeId,
          player_id: 1,
        }
      );
      const game = response.data;
      uiStore.openGame(game.game_id);
    } catch (error) {
      // TODO: handle with UI Store
      throw new NetworkError("failed to join game", error);
    } finally {
      loadingGame.value = false;
    }
  }

  return { themes, loadingGame, fetchThemes, joinGame };
});
