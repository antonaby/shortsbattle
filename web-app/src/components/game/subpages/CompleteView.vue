<script setup lang="ts">
import { useGameStore } from '@/stores/game.store';
import { useUIStore } from '@/stores/ui.store';

const uiStore = useUIStore();
const gameStore = useGameStore();
</script>

<template>
  <div class="page-container">
    <div class="text-2xl font-medium text-center">
      <span>Game Complete</span>
    </div>

    <button class="action-button w-full" @click="uiStore.returnToHub()">
      🎉 Finish
    </button>
  </div>
</template>