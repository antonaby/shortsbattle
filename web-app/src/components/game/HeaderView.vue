<script setup lang="ts">
import { useRouter } from 'vue-router';
import { useGameStore } from '../../stores/game.store';

const router = useRouter();

function returnHome() {
  router.replace({ name: "home" });
}

const gameStore = useGameStore();
</script>

<template>
  <div class="flex mb-4">
    <button class="px-3 py-1 rounded bg-gray-100 hover:bg-gray-200 mr-4" @click="returnHome()">&larr; Back</button>
    <div class="flex-1">
      <h1 class="text-lg font-bold">{{ gameStore.theme.name }}</h1>
      <p class="text-sm text-gray-600">{{ gameStore.theme.description }}</p>
    </div>
  </div>
</template>