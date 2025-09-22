<script setup lang="ts">
import { onMounted } from 'vue';
import { useGameStore } from '../../../stores/game.store';
import VideoPlayerWrapper from '../players/VideoPlayerWrapper.vue';
import FullscreenLoaderView from '@/components/common/FullscreenLoaderView.vue';

const gameStore = useGameStore();

onMounted(() => {
  gameStore.loadVideosToWatch()
})
</script>

<template>
  <FullscreenLoaderView v-if="gameStore.videosToWatch.length == 0" :gray-background="false" msg="⚡ Loading videos..." />
  <!-- TODO: handle video errors! -->
  <!-- TODO: check player errors -->
  <VideoPlayerWrapper 
    v-if="gameStore.videosToWatch.length > 0" 
    :videos="gameStore.videosToWatch" 
    :completed="gameStore.lastUpdate?.stage == 'submit-complete'"
    @vote="gameStore.voteForVideo"
   />
</template>