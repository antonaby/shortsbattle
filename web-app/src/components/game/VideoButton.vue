<script setup lang="ts">
import { useGameStore } from '../../stores/game.store';
import type { Video } from '../../types/game';

defineProps<{
  video: Video
}>();

const gameStore = useGameStore();
</script>

<template>
  <button class="flex w-full border rounded-lg overflow-hidden" @click="gameStore.selectVideo(video.id)">
    <div class="flex justify-center items-center bg-gray-100 w-2/5">
      <img v-if="video.oembed?.thumbnail_url" :src="video.oembed.thumbnail_url"
        :alt="video.oembed?.title ?? 'Video thumbnail'" class="object-cover" loading="lazy" />
      <div v-else class="flex items-center justify-center text-gray-500">
        No thumbnail
      </div>
    </div>
    <div class="flex flex-col flex-wrap w-3/5">
      <h3 class="text-base font-semibold text-gray-900">
        {{ video.oembed?.title ?? 'Untitled video' }}
      </h3>
      <p class="mt-1 text-sm text-gray-600">
        {{ video.oembed?.author_name ?? 'Unknown author' }}
      </p>
    </div>
  </button>
</template>