<script lang="ts" setup>
import { useThemeStore } from '@/stores/theme.store';
import GameModeModal from '@/components/theme/GameModeModal.vue';
import { onMounted, ref } from 'vue';
import { useRoute } from 'vue-router';
import ThemeLoadingScreen from '@/components/theme/ThemeLoadingScreen.vue';
import type { PlayerMode } from '@/types/game';

const route = useRoute();
const themeStore = useThemeStore()
const selectedMode = ref<PlayerMode | undefined>(undefined);

let themeId = Number(route.params.id);

function selectJoin() {
  selectedMode.value = 'submit_and_vote';
}

function selectWatch() {
  selectedMode.value = 'only_vote';
}

function unselectMode() {
  selectedMode.value = undefined;
}

async function joinGame() {
  if (selectedMode.value) {
    await themeStore.joinGame(themeId, selectedMode.value);
  }
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
    <h1 class="text-title text-center">{{ themeStore.theme.title }}</h1>
    <!-- 3) Description -->
    <p class="text-description text-center">{{ themeStore.theme.description }}</p>
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
    <div class="flex flex-col gap-2 p-2 bg-gray-100 rounded-xl"
      v-if="themeStore.theme.rounds && themeStore.theme.rounds.length > 0">
      <h1 class="text-title-2 text-center">Rounds</h1>
      <ul class="space-y-2">
        <li v-for="round in themeStore.theme.rounds" :key="round.round_n" class="flex items-center gap-4">
          <img src="https://cdn2.thecatapi.com/images/bpc.jpg" alt="Round" class="w-16 h-16 object-cover rounded-lg" />
          <div class="flex-1">
            <h2 class="text-title-item">{{ round.round_n }}. {{ round.title }}</h2>
            <p class="text-description">{{ round.description }}</p>
          </div>
        </li>
      </ul>
    </div>
  </div>
</template>
