<script setup lang="ts">
import { useGameStore } from "../stores/gameStore";

const gameStore = useGameStore();
const game = gameStore.currentGame;

</script>

<template>
  <div class="flex flex-col" v-if="game">
    <!-- Top Content -->
    <div class="flex-1">
      <h1 class="text-2xl font-bold text-center mb-6">🎮 Game Lobby</h1>
      <div class="text-center">
        <h2 class="text-xl font-bold text-gray-900">{{ game.name }}</h2>
        <p class="text-sm text-gray-500 mt-1">{{ game.description }}</p>
      </div>
    </div>

    <!-- Bottom Button -->
    <div class="pt-4 flex justify-center gap-3">
      <button
        @click="gameStore.clearGameId()"
        class="w-35 py-3 text-center bg-gray-100 text-gray-700 rounded-xl hover:bg-gray-200 active:bg-gray-300 transition"
      >
        🔙 Back
      </button>
    </div>
  </div>

  <!-- Fallback when game is null -->
  <div v-else class="text-center mt-6">
    <p class="text-gray-500">No game selected.</p>
    <button
      @click="gameStore.clearGameId()"
      class="mt-3 py-2 px-4 bg-gray-200 rounded-lg hover:bg-gray-300"
    >
      Go Back
    </button>
  </div>
</template>