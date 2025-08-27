<script setup lang="ts">
import { useUIStore } from '../../stores/ui.store';

const uiStore = useUIStore();
</script>

<template>
  <div class="flex mb-4">
    <button class="px-3 py-1 rounded bg-gray-100 hover:bg-gray-200 mr-4" @click="uiStore.returnToHub()">&larr; Back</button>
    <div class="flex-1">
      <h1 class="text-lg font-bold">TBD</h1>
      <p class="text-sm text-gray-600">TBD</p>
    </div>
  </div>
</template>