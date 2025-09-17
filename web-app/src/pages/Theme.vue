<script lang="ts" setup>
import { useThemeStore } from '@/stores/theme.store';
import FullscreenLoaderView from '@/components/common/FullscreenLoaderView.vue'
import GameModeModal from '@/components/theme/GameModeModal.vue';
import { ref } from 'vue';

const themeStore = useThemeStore()
const selectedMode = ref<string | undefined>(undefined);
</script>
<template>
  <FullscreenLoaderView msg="Joining..." v-if="themeStore.joinGameLoader" />
  <GameModeModal :mode="selectedMode" v-if="selectedMode" @cancel="selectedMode = undefined" />
  <div class="page-container">
    <!-- 1) Main image -->
    <img src="https://cdn2.thecatapi.com/images/8q1.jpg" alt="Main image" class="w-full h-48 object-cover rounded-lg" />
    <!-- 2) Title -->
    <h1 class="text-title text-center">Cat Clash</h1>
    <!-- 3) Description -->
    <p class="text-description text-center">Share your funniest, cutest, or silliest cat videos and compete for votes.
    </p>
    <!-- 4) Controls -->
    <div class="flex gap-2">
      <button class="action-button flex-1" @click="selectedMode = 'join'">
        <span>🚀</span>
        <span>Join Game</span>
      </button>
      <button class="action-button flex-1" @click="selectedMode = 'watch'">
        <span>👀</span>
        <span>Watch</span>
      </button>
    </div>
    <!-- 
    <p class="text-help">
      <span class="font-medium">
        Watch:
      </span>
      
    </p> -->
    <!-- 5) List of rounds -->
    <h1 class="text-title-2 text-center">Rounds</h1>
    <ul class="space-y-2">
      <!-- Round 1 -->
      <li class="flex items-center gap-4">
        <img src="https://cdn2.thecatapi.com/images/bpc.jpg" alt="Round 1" class="w-16 h-16 object-cover rounded-lg" />
        <div class="flex-1">
          <h2 class="text-title-item">Curious Cats</h2>
          <p class="text-description">Show cats exploring, investigating, or getting into unexpected places.</p>
        </div>
      </li>
      <!-- Round 2 -->
      <li class="flex items-center gap-4">
        <img src="https://cdn2.thecatapi.com/images/MTY3ODIyMQ.jpg" alt="Round 2"
          class="w-16 h-16 object-cover rounded-lg" />
        <div class="flex-1">
          <h2 class="text-title-item">Sleepy Whiskers</h2>
          <p class="text-description">Share adorable moments of cats napping in funny or unusual spots.</p>
        </div>
      </li>
      <!-- Round 3 -->
      <li class="flex items-center gap-4">
        <img src="https://cdn2.thecatapi.com/images/384.jpg" alt="Round 2" class="w-16 h-16 object-cover rounded-lg" />
        <div class="flex-1">
          <h2 class="text-title-item">Playful Paws</h2>
          <p class="text-description">Capture cats playing with toys, chasing, or just being mischievous.</p>
        </div>
      </li>
    </ul>
  </div>
</template>
