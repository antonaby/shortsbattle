<script setup lang="ts">
import { onMounted } from 'vue';
import { useGameStore } from '../../../stores/game.store';
import VideoPlayerWrapper from '../players/VideoPlayerWrapper.vue';
import FullscreenLoaderView from '@/components/common/FullscreenLoaderView.vue';

const gameStore = useGameStore();

</script>

<template>
  <FullscreenLoaderView />
  
  <!-- <VideoPlayerWrapper 
    v-if="gameStore.gameVideos.length > 0" 
    :videos="gameStore.gameVideos" 
    @vote="gameStore.voteForVideo" 
    @error="gameStore.handleVideoError" 
    @finished="gameStore.handleFinished" /> -->
  
</template>