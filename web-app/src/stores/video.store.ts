import type { Video } from "@/types/game";
import { defineStore } from "pinia";
import { ref } from "vue";


export const useVideoStore = defineStore("video", () => {
  const playerVideos = ref<Video[]>([]);


  return {
    playerVideos
  }
})