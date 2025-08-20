<script setup lang="ts">
import { ref, type Ref } from 'vue';
import { useGameStore } from '../../stores/gameStore';

const gameStore = useGameStore();
const url: Ref<string> = ref("")
</script>

<template>
  <form class="max-w-md mx-auto space-y-4" @submit.prevent="gameStore.submitVideo(url)">
    <div class="flex items-center space-x-4">
      <p class="font-mono font-medium text-3xl text-gray-950">{{ gameStore.formattedTime }}</p>
      <h1 class="flex-1 text-lg">Submit your video</h1>
    </div>
    <label class="block">
      <span class="block mb-1">Video URL</span>
      <input v-model="url" type="url" name="url" placeholder="https://…" required class="w-full border rounded p-2" />
    </label>
    <button type="submit" class="w-full py-2 bg-black text-white rounded">Submit</button>
  </form>
</template>