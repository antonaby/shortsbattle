import {  defineStore } from 'pinia'
import router from '../router'

interface GameDetails {
  name: string
  description: string
}

interface GameState {
  gameId: number | null
  games: GameDetails[]
}

export const useGameStore = defineStore('game', {
  state: (): GameState => ({
    gameId: null,
    games: [
      {
        name: "The most cute cat 🐈",
        description: "A game about the cutest cat in the world. 🐱"
      },
      {
        name: "Funniest fail video 😂",
        description: "Submit a hilarious fail that makes everyone laugh!"
      },
      {
        name: "Best dance move 💃",
        description: "Show off your craziest or smoothest dance step."
      },
      {
        name: "Unexpected twist 🎭",
        description: "Videos that take a surprising turn. Shock us!"
      },
      {
        name: "Cutest baby animal 🐾",
        description: "Puppies, kittens, ducklings... bring the awws!"
      },
      {
        name: "Most epic moment ⚡",
        description: "Highlight something legendary, heroic, or just cool."
      },
      {
        name: "Mind-blowing magic trick 🎩✨",
        description: "Is it real? Is it edited? Blow our minds!"
      },
      {
        name: "Satisfying video 🍰",
        description: "Soap cutting, symmetry, pouring — we want chill."
      },
      {
        name: "Cringe overload 😬",
        description: "Bring the secondhand embarrassment in a fun way."
      },
      {
        name: "Best pet reaction 🐶😲",
        description: "Pets doing something wild, unexpected, or smart!"
      }
    ],
  }),
  actions: {
    setGameId(id: number) {
      this.gameId = id;
      router.push("/game")
    },
    clearGameId() {
      this.gameId = null;
      router.push("/")
    }
  }
});