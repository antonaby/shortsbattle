<script setup lang="ts">
import { useYouTubeIframeApi } from '@/composables/yt';
import { extractYTVideoId } from '@/utils/players';
import { nextTick, onMounted, onUnmounted, watch } from 'vue';

const props = defineProps<{
  videoUrl: string
}>();

const emits = defineEmits<{
  (e: 'stateChange', state: YT.PlayerState): void
  (e: 'error', code: YT.PlayerError): void
}>();

let player: YT.Player | null = null;

watch(() => props.videoUrl, (newVal) => {
  if (player) {
    player.loadVideoById(extractYTVideoId(newVal));
  }
});

onMounted(async () => {
  const YT = await useYouTubeIframeApi();
  await nextTick();

  player = new YT.Player('yt-shorts', {
    videoId: extractYTVideoId(props.videoUrl),
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
        e.target.mute();
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