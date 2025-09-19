<script setup lang="ts">
import { onMounted } from 'vue';
import { useGameStore } from '@/stores/game.store';
import VideoButton from '../common/VideoButton.vue';
import type { Video } from '@/types/game';

const gameStore = useGameStore();

function onSubmitVideo(video?: Video | null) {
  if (video) {
    gameStore.selectVideo(video.id);
  }
}

const emit = defineEmits<{
  (e: 'newVideo'): void
}>();

onMounted(() => {
  gameStore.reloadPlayerVideos();
})
</script>
<template>
  <button 
  v-if="!gameStore.loadingData && gameStore.playerVideos.length > 0"
    class="text-sm text-blue-500" 
    @click="gameStore.reloadPlayerVideos()">
    Reload
  </button>
  <div 
  v-if="!gameStore.loadingData && gameStore.playerVideos.length == 0" 
    class="flex flex-col items-center justify-start space-y-2">
    <p class="text-2xl font-medium text-center">
      You have no videos.
    </p>
    <button 
      @click="emit('newVideo')" 
      class="p-2 bg-black text-white rounded">
      Upload video
    </button>
  </div>

  <ul class="space-y-2">
    <li v-for="video in gameStore.playerVideos" v-if="!gameStore.loadingData">
      <VideoButton :video="video" @submit="onSubmitVideo" />
    </li>
    <li v-for="n in 3" :key="n" v-else>
      <VideoButton />
    </li>
  </ul>
</template>