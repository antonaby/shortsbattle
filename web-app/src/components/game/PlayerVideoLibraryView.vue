<script setup lang="ts">
import { onMounted } from 'vue';
import { useGameStore } from '../../stores/game.store';

const gameStore = useGameStore();

onMounted(() => {
  gameStore.loadPlayerVideos();
})
</script>
<template>
  <div>
    <span>List video</span>
  </div>
</template>