import { defineStore } from "pinia";
import type { GameJoined, Theme } from "../types/game";
import axios from "axios";
import { ref } from "vue";

export const useGameHubStore = defineStore("gamehub", () => {
  const themes = ref<Theme[]>([]);

  async function fetchThemes() {
    try {
      const response = await axios.get<Theme[]>(
        `${import.meta.env.VITE_BASE_URL}/api/v1/themes`
      );
      themes.value = response.data;
    } catch (error) {
      // TODO: handle with UI Store
      console.error("Failed to fetch themes:", error);
    }
  }

  async function joinGame(themeId: number): Promise<GameJoined | null> {
    try {
      const response = await axios.put<GameJoined>(
        `${import.meta.env.VITE_BASE_URL}/api/v1/games/join`,
        {
          theme_id: themeId,
          player_id: 1,
        }
      );
      return response.data;
    } catch (error) {
      // TODO: handle with UI Store
      console.error("Failed to fetch the game:", error);
      return null;
    }
  }

  return { themes, fetchThemes, joinGame };
});
