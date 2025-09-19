<script lang="ts" setup>
import { useVideoStore } from '@/stores/video.store';
import SubmitView from '@/components/game/subpages/SubmitView.vue';
import data from '@/components/dev/videos.json';

const videoStore = useVideoStore();
videoStore.playerVideos = data;
</script>

<template>
  <SubmitView />
</template>