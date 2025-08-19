import { defineStore } from "pinia";
import type { GameTheme } from "../models/common";
import axios from "axios";

interface GameListStore {
  games: GameTheme[]
}

export const useGameListStore = defineStore("game-list", {
  state: (): GameListStore => ({
    games: [] 
  }),
  actions: {
    async fetchGames() {
      try {
        const response = await axios.get<GameTheme[]>(
          `${import.meta.env.VITE_BASE_URL}/api/v1/themes`
        );
        this.games = response.data;
      } catch (error) {
        console.error("Failed to fetch games:", error);
      } finally {
        // TODO: add spinner
      }
    },
  }  
});