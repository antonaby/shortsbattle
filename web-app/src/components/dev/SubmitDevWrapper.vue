<script lang="ts" setup>
import { useVideoStore } from '@/stores/video.store';
import SubmitView from '@/components/game/subpages/SubmitView.vue';
import data from '@/components/dev/videos.json';
import { useGameStore } from '@/stores/game.store';
import { onMounted } from 'vue';

const gameStore = useGameStore();
const videoStore = useVideoStore();
videoStore.playerVideos = data;

onMounted(() => {
	gameStore.startTimer(5000);
})
</script>

<template>
  <SubmitView />
</template>