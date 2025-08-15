import { defineStore } from "pinia";
import axios from "axios";
import router from "../router";

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
  loading: boolean;
  gameId: number | null;
  gameInstance: GameInstance | null;
  games: Game[];
}

export const useGameStore = defineStore("game", {
  state: (): GameStoreState => ({
    loading: false,
    gameId: null,
    gameInstance: null,
    games: [],
  }),
  getters: {
    currentGame: (state) => {
      return state.gameId
        ? state.games.find((game) => game.id === state.gameId) || null
        : null;
    },
  },
  actions: {
    async fetchGames() {
      this.loading = true;
      try {
        const response = await axios.get(
          `${import.meta.env.VITE_BASE_URL}/api/v1/themes`
        );
        this.games = response.data;
      } catch (error) {
        console.error("Failed to fetch games:", error);
      } finally {
        this.loading = false;
      }
    },
    async joinGame(id: number) {
      this.gameId = id;
      this.loading = true;
      try {
        const response = await axios.post(
          `${import.meta.env.VITE_BASE_URL}/api/v1/games`,
          {
            theme_id: id,
          }
        );
        this.gameInstance = response.data;
      } catch (error) {
        console.error("Failed to fetch games:", error);
      } finally {
        this.loading = false;
      }

      router.push("/game");
    },
    clearGameId() {
      this.gameId = null;
      this.gameInstance = null;
      router.push("/");
    },
  },
});
