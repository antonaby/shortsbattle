import { defineStore } from "pinia";
import axios from "axios";
import type { GameJoined, GameInstance, GameUpdate } from "../models/common";

interface GameStoreState {
  gameInstance: GameInstance | null;
  remainingTimeMs: number;
  intervalId: number | null;
}

function diffInMilliseconds(
  state_changed_at: string,
  next_state_change_at: string
): number {
  const start = new Date(state_changed_at).getTime();
  const end = new Date(next_state_change_at).getTime();

  return end - start;
}

export const useGameStore = defineStore("game", {
  state: (): GameStoreState => ({
    gameInstance: null,
    remainingTimeMs: 0,
    intervalId: null,
  }),
  getters: {
    theme: (state) => {
      return {
        name: "TDB",
        description: "TBD",
      };
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
    async joinGame(id: number): Promise<GameJoined | null> {
      try {
        const response = await axios.put<GameJoined>(
          `${import.meta.env.VITE_BASE_URL}/api/v1/games/join`,
          {
            theme_id: id,
            player_id: 1,
          }
        );
        return response.data;
      } catch (error) {
        // TODO: show error and get back
        console.error("Failed to fetch the game:", error);
        return null;
      }
    },
    setGameInstance(gi: GameInstance) {
      this.gameInstance = gi;
      this.remainingTimeMs = diffInMilliseconds(gi.state_changed_at, gi.next_state_change_at);

      if (this.intervalId) {
        clearInterval(this.intervalId);
      }

      this.intervalId = setInterval(() => {
        this.remainingTimeMs -= 1000;
      }, 1000);
    },
    updateGameStatus(data: GameUpdate) {
      if (this.gameInstance) {
        if (this.gameInstance.state !== data.status) {
          this.remainingTimeMs = data.stage_time_remaining;
        }
        this.gameInstance.state = data.status;
      }
    },
    submitVideo(url: string) {
      console.log(url);
    },
    leaveGame() {
      if (this.intervalId) {
        clearInterval(this.intervalId);
        this.intervalId = null;
      }
      this.gameInstance = null;
    },
  },
});
