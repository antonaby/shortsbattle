<script setup lang="ts">
import { onMounted, onUnmounted } from "vue";
import { useWSStore } from "../stores/wsStore";
import { useRoute, useRouter } from "vue-router";
import { useGameStore } from "../stores/gameStore";

const route = useRoute();
const router = useRouter();
const wsStore = useWSStore();
const gameStore = useGameStore();

function returnHome() {
  router.replace({ name: "home" });
}

onMounted(() => {
  let id = Number(route.params.id);
  wsStore.subscribe(id);
})

onUnmounted(() => {
  gameStore.leaveGame();
  wsStore.unsubsribe();
})
</script>

<template>
  <div class="flex flex-col" v-if="gameStore.gameInstance">
    <div v-if="!wsStore.subscribed" class="flex-1 flex items-center justify-center">
      <div class="flex items-center gap-3">
        <svg class="animate-spin h-6 w-6" viewBox="0 0 24 24" fill="none">
          <circle cx="12" cy="12" r="10" stroke="currentColor" opacity="0.2" stroke-width="4" />
          <path d="M22 12a10 10 0 0 1-10 10" stroke="currentColor" stroke-width="4" />
        </svg>
        <span>Connecting…</span>
      </div>
    </div>

    <div v-else>
      <div class="flex-1">
        <h1 class="text-2xl font-bold text-center mb-6">🎮 Game Lobby</h1>
        <div class="text-center">
          <h2 class="text-xl font-bold text-gray-900">{{ gameStore.theme.name }}</h2>
          <p class="text-sm text-gray-500 mt-1">{{ gameStore.theme.description }}</p>
          <p class="text-lg text-gray-500 mt-1">{{ gameStore.gameInstance.status }}</p>
          <p class="text-lg text-gray-500 mt-1">{{ gameStore.formattedTime }}</p>
        </div>
      </div>

      <div class="pt-4 flex justify-center gap-3">
        <button @click="returnHome()"
          class="w-35 py-3 text-center bg-gray-100 text-gray-700 rounded-xl hover:bg-gray-200 active:bg-gray-300 transition">
          🔙 Back
        </button>
      </div>
    </div>
  </div>

  <div v-else class="text-center mt-6">
    <p class="text-gray-500">Loading</p>
  </div>
</template>