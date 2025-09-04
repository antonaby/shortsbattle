<script setup lang="ts">
import VideoButton from './VideoButton.vue';
import type { Video } from '@/types/game';

defineProps<{
  video: Video
}>();

const emits = defineEmits<{
  (e: 'unselect'): void
}>();
</script>

<template>
  <VideoButton :video="video" />
  <button @click="emits('unselect')">
    <div class="text-sm text-blue-500">Pick another video</div>
  </button>
</template>