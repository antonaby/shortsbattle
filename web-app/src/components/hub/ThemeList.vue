<script setup lang="ts">
import { onMounted } from 'vue'
import { useGameHubStore } from '../../stores/hub.store';
import ThemeButtonView from './ThemeButtonView.vue';
import type { Theme } from '@/types/game';

const gameHubStore = useGameHubStore();

async function joinAndOpenGame(theme?: Theme) {
  if (!theme) {
    return;
  }
  
  await gameHubStore.openGameDetails(theme);
}

onMounted(() => {
  gameHubStore.fetchThemes();
})
</script>

<template>
  <ul>
    <li v-for="theme in gameHubStore.themes" v-if="gameHubStore.themes.length > 0">
      <ThemeButtonView :theme="theme" @join="joinAndOpenGame" />
    </li>
    <li v-for="n in 3" :key="n" v-else>
      <ThemeButtonView />
    </li>
  </ul>
</template>