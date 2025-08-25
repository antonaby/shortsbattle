import { defineStore } from "pinia";
import axios from "axios";
import type { GameJoined, GameUpdate } from "../models/common";

interface GameStoreState {
  lastGameUpdate: GameUpdate | null;
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
    lastGameUpdate: null,
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
    setGameInstance(upd: GameUpdate) {
      this.lastGameUpdate = upd;
      this.remainingTimeMs = diffInMilliseconds(
        upd.state_changed_at,
        upd.next_state_change_at
      );

      if (this.intervalId) {
        clearInterval(this.intervalId);
      }

      this.intervalId = setInterval(() => {
        this.remainingTimeMs -= 1000;
      }, 1000);
    },
    updateGameStatus(upd: GameUpdate) {
      if (this.lastGameUpdate) {
        if (this.lastGameUpdate.state !== upd.state) {
          this.remainingTimeMs = diffInMilliseconds(
            upd.state_changed_at,
            upd.next_state_change_at
          );
        }
        this.lastGameUpdate = upd;
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
      this.lastGameUpdate = null;
    },
  },
});
