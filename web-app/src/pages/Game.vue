<script setup lang="ts">
import { onMounted, onUnmounted } from "vue";
import { useGameStore } from "../stores/gameStore";
import { useCentrifugeStore } from "../stores/centrifugeStore";

const gameStore = useGameStore();
const game = gameStore.currentGame;
const gameInstance = gameStore.currentGameInstance;

const centrifugeStore = useCentrifugeStore()

onMounted(() => {
  if (game && gameInstance) {
    centrifugeStore.subscribe(gameInstance.id)
  }
})

onUnmounted(() => {
  centrifugeStore.unsubsribe();
})
</script>

<template>
  <div class="flex flex-col" v-if="game">
    <div v-if="!centrifugeStore.subscribed" class="flex-1 flex items-center justify-center">
      <div class="flex items-center gap-3">
        <!-- simple spinner -->
        <svg class="animate-spin h-6 w-6" viewBox="0 0 24 24" fill="none">
          <circle cx="12" cy="12" r="10" stroke="currentColor" opacity="0.2" stroke-width="4" />
          <path d="M22 12a10 10 0 0 1-10 10" stroke="currentColor" stroke-width="4" />
        </svg>
        <span>Connecting…</span>
      </div>
    </div>

    <div v-else>
      <!-- Top Content -->
      <div class="flex-1">
        <h1 class="text-2xl font-bold text-center mb-6">🎮 Game Lobby</h1>
        <div class="text-center">
          <h2 class="text-xl font-bold text-gray-900">{{ game.name }}</h2>
          <p class="text-sm text-gray-500 mt-1">{{ game.description }}</p>
          <p class="text-lg text-gray-500 mt-1" v-if="gameInstance">{{ gameInstance.status }}</p>
        </div>
      </div>

      <!-- Bottom Button -->
      <div class="pt-4 flex justify-center gap-3">
        <button @click="gameStore.leaveGame()"
          class="w-35 py-3 text-center bg-gray-100 text-gray-700 rounded-xl hover:bg-gray-200 active:bg-gray-300 transition">
          🔙 Back
        </button>
      </div>

      <div class="pt-4 flex justify-center gap-3">
        <button @click="centrifugeStore.publish({ type: 'message', text: 'Hello world', id: 123 })"
          class="w-35 py-3 text-center bg-gray-100 text-gray-700 rounded-xl hover:bg-gray-200 active:bg-gray-300 transition">
          Send Something
        </button>
      </div>
    </div>
  </div>

  <!-- Fallback when game is null -->
  <div v-else class="text-center mt-6">
    <p class="text-gray-500">No game selected.</p>
    <button @click="gameStore.leaveGame()" class="mt-3 py-2 px-4 bg-gray-200 rounded-lg hover:bg-gray-300">
      Go Back
    </button>
  </div>
</template>