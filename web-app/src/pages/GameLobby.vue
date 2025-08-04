<script setup lang="ts">
import { useGameStore } from "../stores/gameStore";

const gameStore = useGameStore();
const game = gameStore.games[gameStore.gameId || 0];

function goBack() {
  gameStore.clearGameId()
}

</script>

<template>
  <div class="flex flex-col">
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
        @click="goBack"
        class="w-35 py-3 text-center bg-gray-100 text-gray-700 rounded-xl hover:bg-gray-200 active:bg-gray-300 transition"
      >
        🔙 Back
      </button>
      <button class="w-35 py-3 text-center bg-gray-100 text-gray-700 rounded-xl hover:bg-gray-200 active:bg-gray-300 transition">
        🎮 Join
      </button>
    </div>
  </div>
</template>