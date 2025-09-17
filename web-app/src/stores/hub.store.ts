import { defineStore } from "pinia";
import type { Theme } from "../types/game";
import { ref } from "vue";
import { useUIStore } from "./ui.store";
import { GamesAPI } from "../api/games";

export const useGameHubStore = defineStore("gamehub", () => {
  const themes = ref<Theme[]>([]);
  const uiStore = useUIStore();

  async function fetchThemes() {
    try {
      themes.value = await GamesAPI.fetchThemes();
    } catch (error) {
      uiStore.handleNetworkError(error);
    }
  }

  async function openGameDetails(theme: Theme) {
    uiStore.openTheme(theme.id);
  }

  return { themes, fetchThemes, openGameDetails };
});
