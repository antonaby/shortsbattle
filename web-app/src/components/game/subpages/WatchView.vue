<script setup lang="ts">
import { useGameStore } from '../../../stores/game.store';
import VideoPlayerWrapper from '../players/VideoPlayerWrapper.vue';
import FullscreenLoaderView from '@/components/common/FullscreenLoaderView.vue';

const gameStore = useGameStore();

</script>

<template>
  <FullscreenLoaderView v-if="gameStore.videosToWatch.length == 0" />
  
  <VideoPlayerWrapper 
    v-if="gameStore.videosToWatch.length > 0" 
    :videos="gameStore.videosToWatch" 
   />
  
</template>