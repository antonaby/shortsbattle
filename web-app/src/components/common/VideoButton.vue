<script setup lang="ts">
import { computed } from 'vue';
import type { Video } from '../../types/game';
import { detectVideoPlatform, type VideoPlatform } from '@/utils/players';

const props = defineProps<{
  video?: Video | null
}>();

const emit = defineEmits<{
  (e: 'submit', video?: Video | null): void
}>();

const platform = computed<VideoPlatform>(() => {
  if (!props.video) {
    return "unknown";
  }

  return detectVideoPlatform(props.video?.video_url);
})
</script>

<template>
  <button v-if="video" class="flex gap-4 w-full items-start" @click="emit('submit', props.video)">
    <div class="flex justify-center items-center bg-gray-100 rounded-xl overflow-hidden w-1/3">
      <img v-if="video.oembed.thumbnail_url" :src="video.oembed.thumbnail_url"
        :alt="video.oembed.title ?? 'Video thumbnail'" class="object-cover max-h-28" loading="lazy" />
      <div v-else class="flex items-center justify-center text-gray-500 h-24">
        🙈 No Preview
      </div>
    </div>
    <div class="flex-1 flex flex-col items-start justify-center gap-2">
      <div class="flex items-center gap-2">
        <span v-if="platform != 'unknown'" class="w-16 py-1 text-xs text-teal-800 bg-emerald-200 rounded-lg">
          <span v-if="platform == 'tiktok'">TikTok</span>
          <span v-if="platform == 'youtube'">YouTube</span>
        </span>
        <span class="text-description">
          {{ video.oembed.author_name ?? 'Unknown author' }}
        </span>
      </div>
      <p class="text-left w-full">
        {{ video.oembed.title ?? 'Untitled video' }}
      </p>
    </div>
  </button>
</template>