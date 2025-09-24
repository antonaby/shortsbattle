<script lang="ts" setup>
import VHIcon from '@/components/common/VHIconView.vue';
import type { VideoResult } from '@/types/game';
import { detectVideoPlatform, type VideoPlatform } from '@/utils/players';
import { computed } from 'vue';

const props = defineProps<{
  playerId: number,
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
    <div class="flex flex-wrap flex-col gap-1 w-10">
      <div class="ld-container border-emerald-300 text-emerald-700">
        <p>👍</p>
        <p>{{ videoResult.likes }}</p>
      </div>
      <div class="ld-container border-rose-300 text-rose-700">
        <p>👎</p>
        <p>{{ videoResult.dislikes }}</p>
      </div>
    </div>
    <div v-if="videoResult.video.oembed.thumbnail_url" class="thumbnail-container">
      <img :src="videoResult.video.oembed.thumbnail_url" :alt="videoResult.video.oembed.title || 'Video Cover'"
        class="object-cover min-h-74" loading="lazy" />
    </div>
    <div v-else class="thumbnail-container min-h-26 text-gray-600 font-light">
      <span>No Image</span>
    </div>
    <div class="flex-1 flex flex-col gap-1">
      <div class="flex flex-wrap items-center justify-start gap-1">
        <div class="text-sm px-2 py-1 rounded-lg  text-white"
          :class="videoResult.author.tgId == playerId 
            ? 'bg-gradient-to-r from-orange-400 to-pink-600' 
            : 'bg-gradient-to-r from-cyan-500 to-blue-600'">
          <span>👤</span>
          <span class="font-semibold max-w-[120px] truncate">{{ videoResult.author.username }}</span>
        </div>
      </div>
      <VHIcon :platform="platform" :author="videoFrom" />
      <span class="text-sm">{{ videoResult.video.oembed.title || "no title" }}</span>
    </div>
  </div>
</template>

<style>
@import "tailwindcss";

.thumbnail-container {
  @apply flex justify-center items-center bg-gradient-to-br from-slate-50 to-slate-200 rounded-lg overflow-hidden w-2/5 min-w-28;
}

.ld-container {
  @apply flex-1 text-sm px-1 rounded-lg text-center font-medium border;
}
</style>