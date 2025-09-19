<script lang="ts" setup>
import type { Video } from '@/types/game';
import VideoButton from './VideoButton.vue';

const props = defineProps<{
  videos: Video[]
}>();

const emits = defineEmits<{
  (e: 'select', video: Video): void
}>();

function onClick(video: Video) {
  emits('select', video)
}
</script>

<template>
  <ul class="space-y-4 mt-4">
    <li v-for="video in videos">
      <VideoButton :video="video" @click="onClick" />
    </li>
  </ul>
</template>