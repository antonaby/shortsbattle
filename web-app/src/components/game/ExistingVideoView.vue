<script setup lang="ts">
import { onMounted } from 'vue';
import { useGameStore } from '../../stores/game.store';
import VideoButton from './VideoButton.vue';
import type { Video } from '../../types/game';

const gameStore = useGameStore();

function onSubmitVideo(video?: Video | null) {
  if (video) {
    gameStore.selectVideo(video.id);
  }
}

onMounted(() => {
  gameStore.reloadPlayerVideos();
})
</script>
<template>
  <ul class="space-y-2">
    <li v-for="video in gameStore.playerVideos" v-if="gameStore.playerVideos.length > 0">
      <VideoButton :video="video" @submit="onSubmitVideo" />
    </li>
    <li v-for="n in 3" :key="n" v-else>
      <VideoButton />
    </li>
  </ul>
</template>