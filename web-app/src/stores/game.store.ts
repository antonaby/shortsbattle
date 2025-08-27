import { defineStore } from "pinia";
import axios from "axios";
import type { GameJoined, GameUpdate } from "../types/game";
import { useUserStore } from "./user.store";

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
    theme: () => {
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
    initGame(upd: GameUpdate) {
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
