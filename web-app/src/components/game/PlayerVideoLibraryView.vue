<script setup lang="ts">
import { onMounted } from 'vue';
import { useGameStore } from '../../stores/game.store';
import VideoButton from './VideoButton.vue';

const gameStore = useGameStore();

onMounted(() => {
  gameStore.reloadPlayerVideos();
})
</script>
<template>
  <ul class="space-y-2 divide-y divide-gray-200">
    <li v-for="video in gameStore.playerVideos" v-if="gameStore.playerVideos.length > 0">
      <VideoButton :video="video" />
    </li>
    <li v-for="n in 3" :key="n" v-else>
      <VideoButton />
    </li>
  </ul>
</template>