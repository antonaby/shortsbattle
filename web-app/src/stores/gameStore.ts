import { defineStore } from "pinia";
import axios from "axios";
import router from "../router";
import type { GameUpdate } from "../models/common";

interface Game {
  id: number;
  name: string;
  description: string;
  created_at: string;
}

interface GameInstance {
  id: number;
  theme_id: number;
  status: string;
  created_at: string;
}

interface GameStoreState {
  gameId: number | null;
  gameInstance: GameInstance | null;
  remainingTimeMs: number;
  intervalId: number | null;
  games: Game[];
}

export const useGameStore = defineStore("game", {
  state: (): GameStoreState => ({
    gameId: null,
    gameInstance: null,
    remainingTimeMs: 0,
    intervalId: null,
    games: [],
  }),
  getters: {
    currentGame: (state) => {
      return state.gameId
        ? state.games.find((game) => game.id === state.gameId) || null
        : null;
    },
    currentGameInstance: (state) => {
      return state.gameInstance;
    },
    formattedTime: (state) => {
      const totalSeconds = Math.max(0, Math.floor(state.remainingTimeMs / 1000));
      const minutes = Math.floor(totalSeconds / 60);
      const seconds = totalSeconds % 60;
      return `${minutes}:${seconds.toString().padStart(2, "0")}`;
    },
  },
  actions: {
    async fetchGames() {
      try {
        const response = await axios.get<Game[]>(
          `${import.meta.env.VITE_BASE_URL}/api/v1/themes`
        );
        this.games = response.data;
      } catch (error) {
        console.error("Failed to fetch games:", error);
      } finally {
      }
    },
    async joinGame(id: number) {
      this.gameId = id;
      try {
        const response = await axios.put<GameInstance>(
          `${import.meta.env.VITE_BASE_URL}/api/v1/games/join`,
          {
            theme_id: id,
            player_id: 1,
          }
        );
        this.gameInstance = response.data;
      } catch (error) {
        console.error("Failed to fetch the game:", error);
      }

      router.push({ name: "game", params: { id: this.gameInstance?.id } });
    },
    startTimer() {
      if (this.intervalId) {
        clearInterval(this.intervalId);
      }

      this.intervalId = setInterval(() => {
        this.remainingTimeMs -= 1000;
      }, 1000);
    },
    stopTimer() {
      if (this.intervalId) {
        clearInterval(this.intervalId);
        this.intervalId = null;
      }
    },
    updateGameStatus(data: GameUpdate) {
      if (this.gameInstance) {
        if (this.gameInstance.status !== data.status) {
          this.remainingTimeMs = data.stage_time_remaining;
        }
        this.gameInstance.status = data.status;
      }
    },
    leaveGame() {
      this.gameId = null;
      this.gameInstance = null;
      router.replace({ name: "home" });
    },
  },
});
