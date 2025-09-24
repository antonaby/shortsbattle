<script lang="ts" setup>
import VHIcon from '@/components/common/VHIconView.vue';
import type { VideoResult } from '@/types/game';
import { detectVideoPlatform, type VideoPlatform } from '@/utils/players';
import { computed } from 'vue';

const props = defineProps<{
  videoResult: VideoResult
}>();

const platform = computed<VideoPlatform>(() => {
  return detectVideoPlatform(props.videoResult.video.video_url);
});

const videoFrom = computed<string>(() => {
  return props.videoResult.video.oembed.author_name || "unknown";
});

</script>

<template>
  <div class="flex flex-wrap gap-2 items-start">
    <div v-if="videoResult.video.oembed.thumbnail_url" 
      class="thumbnail-container">
      <img 
        :src="videoResult.video.oembed.thumbnail_url" 
        :alt="videoResult.video.oembed.title || 'Video Cover'" 
        class="object-cover max-h-28" 
        loading="lazy" />
    </div>
    <div v-else class="thumbnail-container min-h-26 text-gray-600 font-light">
      <span>No Image</span>
    </div>
    <div class="flex-1 flex flex-col gap-2">
      <div class="flex flex-wrap items-center justify-start gap-1">
        <div class="text-sm px-2 py-1 rounded-lg bg-gradient-to-r from-cyan-500 to-blue-600 text-white">
          <span>👤</span>
          <span class="font-semibold max-w-[120px] truncate">{{ videoResult.author.username }}</span>
        </div>
        <VHIcon :platform="platform" :author="videoFrom" />
      </div>
      <span class="text-sm">{{ videoResult.video.oembed.title || "no title" }}</span>
      <div class="flex flex-wrap gap-2">
        <div class="ld-container border-emerald-300 text-emerald-700">
          <span>👍</span>
          <span>{{ videoResult.likes }}</span>
        </div>
        <div class="ld-container border-rose-300 text-rose-700">
          <span>👎</span>
          <span>{{ videoResult.dislikes }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<style>
@import "tailwindcss";
.thumbnail-container {
  @apply flex justify-center items-center bg-gray-200 rounded-lg overflow-hidden w-1/3 min-w-28;
}

.ld-container {
  @apply flex-1 text-sm px-2 py-1 rounded-lg text-center font-medium border;
}
</style>