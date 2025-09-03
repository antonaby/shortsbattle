<script setup lang="ts">
import { useYouTubeIframeApi } from '@/composables/yt';
import { extractVideoId } from '@/utils/yt';
import { nextTick, onMounted, onUnmounted } from 'vue';

const props = defineProps<{
  videoId: string
}>();

const emits = defineEmits<{
  (e: 'stateChange', state: YT.PlayerState): void
  (e: 'error', state: YT.PlayerError): void
}>();

let player: YT.Player | null = null;

onMounted(async () => {
  const YT = await useYouTubeIframeApi();
  await nextTick();

  player = new YT.Player('yt-shorts', {
    videoId: extractVideoId(props.videoId),
    width: '100%',
    height: '100%',
    playerVars: {
      autoplay: 0,
      controls: 1,
      playsinline: 1,
      modestbranding: 0,
      fs: 0,
      rel: 0,
      loop: 0,
      enablejsapi: 1
    },
    events: {
      onReady: (e: YT.PlayerEvent) => {
        e.target.playVideo();
      },
      onStateChange: (e: YT.OnStateChangeEvent) => {
        emits("stateChange", e.data);
      },
      onError: (e: YT.OnErrorEvent) => {
        emits("error", e.data);
      }
    }
  });
})

onUnmounted(async () => {
  if (player) {
    player.destroy();
  }
})
</script>

<template>
  <div class="video-wrapper">
    <div id="yt-shorts" class="absolute inset-0"></div>
  </div>
</template>