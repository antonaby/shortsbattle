<script setup lang="ts">
import { onMounted } from 'vue'
import { useGameHubStore } from '../../stores/hub.store';
import ThemeButtonView from './ThemeButtonView.vue';
import type { Theme } from '@/types/game';
import { useUIStore } from '@/stores/ui.store';

const uiStore = useUIStore();
const gameHubStore = useGameHubStore();

function openTheme(theme: Theme) {
  if (!theme) {
    return;
  }

  uiStore.openTheme(theme.id);
}

onMounted(() => {
  gameHubStore.fetchThemes();
})
</script>

<template>
  <ul v-if="gameHubStore.themes.length > 0">
    <li v-for="theme in gameHubStore.themes">
      <ThemeButtonView :theme="theme" @open="openTheme" />
    </li>
  </ul>
  <div v-else class="loader-container gap-2 my-4">
    <div class="loader-big"></div>
    <span class="text-lg font-medium">Loading...</span>
  </div>
</template>