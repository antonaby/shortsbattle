<script setup lang="ts">
import YTPlayer, { type ExposedMethods } from './YTPlayer.vue'
import { ref } from 'vue'

const playerRef = ref<ExposedMethods>()
</script>

<template>
  <YTPlayer
    ref="playerRef"
    :ids="[
      'https://www.youtube.com/shorts/ktzQ2hQ_0Sg',
      'https://www.youtube.com/shorts/wQQu81CA8Ys'
    ]"
    :width="360"
    :autoplayMuted="true"
    :loopSingle="true"
    :controlsBar="1"
    :modestBranding="1"
    :playsInline="1"
    origin="https://yourdomain.com"
  />
</template>