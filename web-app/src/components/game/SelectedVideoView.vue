<script setup lang="ts">
import VideoButton from './VideoButton.vue';
import { useGameStore } from '../../stores/game.store';

const gameStore = useGameStore();
</script>

<template>
  <span class="text-xl font-medium">Selected</span>
  <VideoButton :video="gameStore.selectedVideo" />
  <button @click="gameStore.unselectVideo()">
    <div class="text-sm text-blue-500">Pick another video</div>
  </button>
</template>