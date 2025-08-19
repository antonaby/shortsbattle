<script setup lang="ts">
import { onMounted } from 'vue'
import { useGameListStore } from '../stores/gameListStore';
import { useGameStore } from '../stores/gameStore';

const gameListStore = useGameListStore();
const gameStore = useGameStore()

onMounted(() => {
  gameListStore.fetchGames();
})
</script>

<template>
  <div class="min-h-screen bg-white mx-auto">
    <h1 class="text-2xl font-bold text-center mb-6">🎮 Choose a Game</h1>
    <ul class="divide-y divide-gray-100">
      <li
        v-for="(game, index) in gameListStore.games"
        :key="index"
      >
        <button
          @click="gameStore.joinGame(game.id)"
          class="w-full text-left py-4 px-2 hover:bg-gray-50 active:bg-gray-100 rounded-lg transition"
        >
          <div class="text-base font-medium text-gray-900">
            {{ game.name }}
          </div>
          <div class="text-sm text-gray-500 mt-1">
            {{ game.description }}
          </div>
        </button>
      </li>
    </ul>
  </div>
</template>