<script setup lang="ts">
import { useGameStore } from '../../stores/game.store';

const gameStore = useGameStore();
</script>

<template>
  <div class="text-center">
    <p class="text-sm text-gray-600">Waiting for outher players...</p>
    <p class="font-medium text-3xl text-gray-950">{{ gameStore.formattedTime }}</p>
  </div>
</template>