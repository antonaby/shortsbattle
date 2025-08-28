<script setup lang="ts">
import { onMounted } from 'vue';
import { useGameStore } from '../../stores/game.store';
import VideoButton from './VideoButton.vue';

const gameStore = useGameStore();

onMounted(() => {
  gameStore.loadPlayerVideos();
})
</script>
<template>
  <ul class="space-y-2">
    <li v-for="video in gameStore.playerVideos">
      <VideoButton :video="video" />
    </li>
  </ul>
</template>