<script setup lang="ts">
import { extractTikTokVideoId } from '@/utils/players';
import { computed, onMounted, onUnmounted } from 'vue'

const props = defineProps<{
  videoUrl: string
}>()

const src = computed(() => {
  const videoUrl = props.videoUrl.trim();
  const videoId = extractTikTokVideoId(videoUrl);
  return `https://www.tiktok.com/player/v1/${videoId}?music_info=0&description=0&fullscreen_button=0&autoplay=1&rel=0`
})

const emits = defineEmits<{
  (e: 'stateChange', state: number): void
  (e: 'error', code: number): void
}>();

function videoMessageHandler(event: MessageEvent<any>) {
  if (event.data["x-tiktok-player"]) {
    let type: string = event.data["type"];
    if (type == "onStateChange") {
      emits("stateChange", event.data.value);
    } else if (type == "onError") {
      emits("error", event.data.value);
    }
  }
}

onMounted(() => {
  window.addEventListener('message', videoMessageHandler);
})

onUnmounted(() => {
  window.removeEventListener('message', videoMessageHandler);
})
</script>

<template>
  <div class="video-wrapper">
    <iframe :src="src" frameborder="0" allowfullscreen loading="lazy"
      allow="autoplay; clipboard-write; encrypted-media;" referrerpolicy="strict-origin-when-cross-origin"
      title="TikTok video"></iframe>
  </div>
</template>