<script setup lang="ts">
import { onMounted } from 'vue';
import { useGameStore } from '../../stores/game.store';

const gameStore = useGameStore();
onMounted(async () => {
  await gameStore.loadVideosToWatch();
});
</script>

<template>
  <div class="flex-1 flex items-center justify-center text-2xl font-medium">
    <span>Watch videos from other players</span>
  </div>
</template>