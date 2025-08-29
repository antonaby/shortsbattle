<script setup lang="ts">
import { onMounted } from 'vue';
import { useGameStore } from '../../stores/game.store';

// TODO: ensure it's connected and subscribed with composables

const gameStore = useGameStore();
onMounted(async () => {
  await gameStore.loadVideosToWatch();
});
</script>

<template>
  <div class="w-screen h-screen bg-gray-200">
    <div class="fixed top-0 left-0 z-50 text-center">
      Total Videos: {{ gameStore.gameVideos.length }}
    </div>
    <div v-html="gameStore.currentVideoOEmbed" v-if="gameStore.currentVideoOEmbed" class="video-wrapper">
    </div>
    <div v-else>
      No video
    </div>
  </div>
</template>