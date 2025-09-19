<script lang="ts" setup>
import type { Video } from '@/types/game';
import VideoButton from './VideoButton.vue';

const props = defineProps<{
  videos: Video[]
}>();
</script>

<template>
  <ul class="space-y-4 mt-4">
    <li v-for="video in videos">
      <VideoButton :video="video" />
    </li>
  </ul>
</template>