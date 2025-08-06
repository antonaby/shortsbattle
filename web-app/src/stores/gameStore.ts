import {  defineStore } from 'pinia'
import axios from 'axios';
import router from '../router'

interface GameDetails {
  name: string
  description: string
}

interface GameState {
  loadingGames: boolean
  gameId: number | null
  games: GameDetails[]
}

export const useGameStore = defineStore('game', {
  state: (): GameState => ({
    loadingGames: true,
    gameId: null,
    games: [],
  }),
  actions: {
    async fetchGames() {
      try {
        const response = await axios.get('http://localhost:8080/v1/games');
        this.games = response.data;
      } catch (error) {
        console.error('Failed to fetch games:', error);
      } finally {
        this.loadingGames = false;
      }
    },
    setGameId(id: number) {
      this.gameId = id;
      router.push("/game")
    },
    clearGameId() {
      this.gameId = null;
      router.push("/")
    },
    watchWideo() {
      router.push('/watch');
    },
    backToSelect() {
      router.push("/");
    }
  }
});