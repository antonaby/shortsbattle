import { GamesAPI } from "@/api/games";
import type { Video } from "@/types/game";
import { defineStore } from "pinia";
import { ref } from "vue";
import { useUIStore } from "./ui.store";

export const useVideoStore = defineStore("video", () => {
  const uiStore = useUIStore();

  const playerVideos = ref<Video[]>([]);

  async function loadVideos(query: string) {
    try {
      playerVideos.value = await GamesAPI.getPayerVideos(query);
    } catch (error) {
      uiStore.handleNetworkError(error);
    }
  }

  return {
    playerVideos,
    loadVideos,
  };
});
