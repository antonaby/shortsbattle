import {  defineStore } from 'pinia'
import router from '../router'

interface GameState {
  gameId: string | null
}

export const useGameStore = defineStore('game', {
  state: (): GameState => ({
    gameId: null
  }),
  actions: {
    setGameId(id: string) {
      this.gameId = id;
      router.push("/game")
    },
    clearGameId() {
      this.gameId = null;
      router.push("/")
    }
  }
});