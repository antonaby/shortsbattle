<script setup lang="ts">
import { useUIStore } from '../../stores/ui.store';

const uiStore = useUIStore();
</script>

<template>
  <div class="flex-1 flex items-center justify-center text-2xl font-medium flex-col space-y-2">
    <span>Game Complete</span>
    <button class="px-3 py-1 rounded bg-gray-100 hover:bg-gray-200" @click="uiStore.returnToHub()">Back To Hub</button>
  </div>
</template>