import { defineStore } from "pinia";
import axios from "axios";
import router from "../router";
import type { Game, GameInstance, GameUpdate } from "../models/common";

interface GameStoreState {
  game: Game | null;
  gameInstance: GameInstance | null;
  remainingTimeMs: number;
  intervalId: number | null;
}

export const useGameStore = defineStore("game", {
  state: (): GameStoreState => ({
    game: null,
    gameInstance: null,
    remainingTimeMs: 0,
    intervalId: null,
  }),
  getters: {
    theme: (state) => {
      return (
        state.gameInstance?.theme || {
          name: "?",
          description: "?",
        }
      );
    },
    formattedTime: (state) => {
      const totalSeconds = Math.max(
        0,
        Math.floor(state.remainingTimeMs / 1000)
      );
      const minutes = Math.floor(totalSeconds / 60);
      const seconds = totalSeconds % 60;
      return `${minutes}:${seconds.toString().padStart(2, "0")}`;
    },
  },
  actions: {
    async joinGame(id: number) {
      try {
        const response = await axios.put<Game>(
          `${import.meta.env.VITE_BASE_URL}/api/v1/games/join`,
          {
            theme_id: id,
            player_id: 1,
          }
        );
        this.game = response.data;
      } catch (error) {
        console.error("Failed to fetch the game:", error);
        return; // TODO: show error and get back
      }

      router.push({ name: "game", params: { id: this.game.id } });
    },
    setGameInstance(gi: GameInstance) {
      this.gameInstance = gi;
      this.remainingTimeMs = gi.stage_time_remaining;
      if (this.intervalId) {
        clearInterval(this.intervalId);
      }
      this.intervalId = setInterval(() => {
        this.remainingTimeMs -= 1000;
      }, 1000);
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
      if (this.intervalId) {
        clearInterval(this.intervalId);
        this.intervalId = null;
      }
      this.game = null;
      this.gameInstance = null;
    },
    returnHome() {
      router.replace({ name: "home" });
    },
  },
});
