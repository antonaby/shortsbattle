<script setup lang="ts">
import { computed, onMounted } from 'vue';
import { useGameStore } from '../../../stores/game.store';
import VideoPlayerWrapper from '../players/VideoPlayerWrapper.vue';
import FullscreenLoaderView from '@/components/common/FullscreenLoaderView.vue';

const gameStore = useGameStore();

const isComplete = computed<boolean>(() => {
  return gameStore.lastUpdate?.stage == 'watch-complete';
})

const isReady = computed<boolean>(() => {
  return gameStore.videosToWatch.length > 0
})

onMounted(() => {
  if (gameStore.lastUpdate?.stage == 'watch') {
    gameStore.loadVideosToWatch()
  }
})
</script>

<template>
  <FullscreenLoaderView v-if="!isReady && !isComplete" :gray-background="false" msg="⚡ Loading videos..." />
  <VideoPlayerWrapper 
    v-if="isReady || isComplete" 
    :videos="gameStore.videosToWatch"
    :completed="isComplete" 
    @vote="gameStore.voteForVideo" 
    @error="gameStore.videoError" />
  <!-- TODO: add view for Ad -->
</template>