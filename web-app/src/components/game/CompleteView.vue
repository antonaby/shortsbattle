<script setup lang="ts">
import { useUIStore } from '../../stores/ui.store';

const uiStore = useUIStore();
</script>

<template>
  <div class="page-container">
    <div class="text-2xl font-medium text-center">
      <span>Game Complete</span>
    </div>
    <div class="text-center">
      <button class="px-3 py-1 rounded bg-gray-100 hover:bg-gray-200" @click="uiStore.returnToHub()">
        <span class="text-xl font-medium">Return To Game List</span>
      </button>
    </div>
  </div>
</template>