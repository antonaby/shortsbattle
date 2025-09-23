<script lang="ts" setup>
import tkIcon from '@/assets/tiktok.png';
import ytIcon from '@/assets/youtube.png';
import type { VideoPlatform } from '@/utils/players';

defineProps<{
  platform: VideoPlatform
}>();
</script>

<template>
  <img :src="tkIcon" alt="TikTok Icon" v-if="platform == 'tiktok'" class="vh-icon"/>
  <img :src="ytIcon" alt="YouTube Icon" v-else class="vh-icon"/>
</template>

<style>
@import "tailwindcss";

.vh-icon {
  @apply object-cover w-7 h-7;
}
</style>