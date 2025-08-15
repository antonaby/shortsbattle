<script setup lang="ts">
import { onMounted } from 'vue'
import { useGameStore } from "../stores/gameStore";

const gameStore = useGameStore();

onMounted(() => {
  gameStore.fetchGames();
})

</script>

<template>
  <div class="min-h-screen bg-white mx-auto">
    <h1 class="text-2xl font-bold text-center mb-6">🎮 Choose a Game</h1>

    <!-- Spinner -->
    <div v-if="gameStore.loading" class="flex justify-center items-center py-10">
      <svg
        class="animate-spin h-8 w-8 text-gray-500"
        xmlns="http://www.w3.org/2000/svg"
        fill="none"
        viewBox="0 0 24 24"
      >
        <circle
          class="opacity-25"
          cx="12" cy="12" r="10"
          stroke="currentColor"
          stroke-width="4"
        ></circle>
        <path
          class="opacity-75"
          fill="currentColor"
          d="M4 12a8 8 0 018-8v4a4 4 0 00-4 4H4z"
        ></path>
      </svg>
    </div>

    <!-- Games list -->
    <ul v-else class="divide-y divide-gray-100">
      <li
        v-for="(game, index) in gameStore.games"
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