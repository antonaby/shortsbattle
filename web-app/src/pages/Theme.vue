<script lang="ts" setup>
import { useThemeStore } from '@/stores/theme.store';
import GameModeModal from '@/components/theme/GameModeModal.vue';
import { onMounted, ref } from 'vue';
import { useRoute } from 'vue-router';
import ThemeLoadingScreen from '@/components/theme/ThemeLoadingScreen.vue';

const route = useRoute();
const themeStore = useThemeStore()
const selectedMode = ref<string | undefined>(undefined);

let themeId = Number(route.params.id);

function selectJoin() {
  selectedMode.value = 'join';
}

function selectWatch() {
  selectedMode.value = 'watch';
}

function unselectMode() {
  selectedMode.value = undefined;
}

async function joinGame() {
  await themeStore.joinGame(themeId);
}

onMounted(() => {
  themeStore.loadTheme(themeId);
})
</script>
<template>
  <GameModeModal :mode="selectedMode" v-if="selectedMode" @ok="joinGame" @cancel="unselectMode"
    :loading="themeStore.joinGameLoader" />
  <ThemeLoadingScreen v-if="!themeStore.theme" />
  <div class="page-container" v-else>
    <!-- 1) Main image -->
    <img src="https://cdn2.thecatapi.com/images/8q1.jpg" alt="Main image" class="w-full h-48 object-cover rounded-lg" />
    <!-- 2) Title -->
    <h1 class="text-title text-center">Cat Clash</h1>
    <!-- 3) Description -->
    <p class="text-description text-center">Share your funniest, cutest, or silliest cat videos and compete for votes.
    </p>
    <!-- 4) Controls -->
    <div class="flex gap-2">
      <button class="action-button flex-1" @click="selectJoin">
        <span>🚀</span>
        <span>Join Game</span>
      </button>
      <button class="action-button flex-1" @click="selectWatch">
        <span>👀</span>
        <span>Watch</span>
      </button>
    </div>
    <!-- 5) List of rounds -->
    <div class="flex flex-col gap-2 p-2 bg-gray-100 rounded-xl">
      <h1 class="text-title-2 text-center">Rounds</h1>
      <ul class="space-y-2">
        <!-- Round 1 -->
        <li class="flex items-center gap-4">
          <img src="https://cdn2.thecatapi.com/images/bpc.jpg" alt="Round 1"
            class="w-16 h-16 object-cover rounded-lg" />
          <div class="flex-1">
            <h2 class="text-title-item">1. Curious Cats</h2>
            <p class="text-description">Show cats exploring, investigating, or getting into unexpected places.</p>
          </div>
        </li>
        <!-- Round 2 -->
        <li class="flex items-center gap-4">
          <img src="https://cdn2.thecatapi.com/images/MTY3ODIyMQ.jpg" alt="Round 2"
            class="w-16 h-16 object-cover rounded-lg" />
          <div class="flex-1">
            <h2 class="text-title-item">2. Sleepy Whiskers</h2>
            <p class="text-description">Share adorable moments of cats napping in funny or unusual spots.</p>
          </div>
        </li>
        <!-- Round 3 -->
        <li class="flex items-center gap-4">
          <img src="https://cdn2.thecatapi.com/images/384.jpg" alt="Round 2"
            class="w-16 h-16 object-cover rounded-lg" />
          <div class="flex-1">
            <h2 class="text-title-item">3. Playful Paws</h2>
            <p class="text-description">Capture cats playing with toys, chasing, or just being mischievous.</p>
          </div>
        </li>
      </ul>
    </div>
  </div>
</template>
