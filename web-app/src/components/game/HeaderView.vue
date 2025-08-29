<script setup lang="ts">
import { useGameStore } from '../../stores/game.store';
import { useUIStore } from '../../stores/ui.store';

const uiStore = useUIStore();
const gameStore = useGameStore();
</script>

<template>
  <div class="flex">
    <button class="px-3 py-1 rounded bg-gray-100 hover:bg-gray-200 mr-4" @click="uiStore.returnToHub()">&larr; Back</button>
    <div class="flex-1">
      <h1 class="text-lg font-bold">{{ gameStore.theme?.name }}</h1>
      <p class="text-sm text-gray-600">{{ gameStore.theme?.description }}</p>
    </div>
  </div>
</template>