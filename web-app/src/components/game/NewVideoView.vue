<script setup lang="ts">
import { ref, type Ref } from 'vue';
import { useGameStore } from '../../stores/game.store';


const gameStore = useGameStore();
const url: Ref<string> = ref("")
</script>

<template>
  <form class="space-y-4" @submit.prevent="gameStore.submitVideo(url)">
    <label class="block">
      <span class="block mb-1">Video URL</span>
      <input v-model="url" type="url" name="url" placeholder="https://…" required class="w-full border rounded p-2" />
    </label>
    <button type="submit" class="w-full py-2 bg-black text-white rounded">Ok</button>
  </form>
</template>