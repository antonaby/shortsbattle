<script setup lang="ts">
import type { Video } from '../../types/game';

const props = defineProps<{
  video?: Video | null
}>();

const emit = defineEmits<{
  (e: 'submit', video?: Video | null): void
}>();
</script>

<template>
  <button class="transition w-full" @click="emit('submit', props.video)">
    <div class="flex" v-if="video">
      <div class="flex justify-center items-center bg-gray-100 rounded-xl overflow-hidden w-2/5">
        <img v-if="video?.oembed?.thumbnail_url" :src="video.oembed.thumbnail_url"
          :alt="video.oembed?.title ?? 'Video thumbnail'" class="object-cover" loading="lazy" />
        <div v-else class="flex items-center justify-center text-gray-500">
          No thumbnail
        </div>
      </div>
      <div class="flex flex-col flex-wrap justify-center w-3/5 px-2">
        <h3 class="text-base font-semibold text-gray-900">
          {{ video?.oembed?.title ?? 'Untitled video' }}
        </h3>
        <p class="mt-1 text-sm text-gray-600">
          {{ video?.oembed?.author_name ?? 'Unknown author' }}
        </p>
      </div>
    </div>
    <div class="flex animate-pulse w-full transition active:bg-gray-100 mb-1" v-else>
      <div class="bg-gray-200 rounded-xl overflow-hidden w-2/5 h-28">&nbsp;</div>
      <div class="flex flex-col items-center justify-center w-3/5 space-y-1">
        <div class="bg-gray-200 rounded-xl text-base w-4/5">
          &nbsp;
        </div>
        <div class="bg-gray-200 rounded-xl w-1/5 text-sm">
          &nbsp;
        </div>
      </div>
    </div>
  </button>
</template>