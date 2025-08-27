<script setup lang="ts">
import { onMounted } from 'vue'
import { useGameHubStore } from '../../stores/hub.store';
import ThemeButtonView from './ThemeButtonView.vue';

const gameHubStore = useGameHubStore();

onMounted(() => {
  gameHubStore.fetchThemes();
})
</script>

<template>
  <ul  class="divide-y divide-gray-100">
    <li v-for="theme in gameHubStore.themes" v-if="gameHubStore.themes.length > 0">
      <ThemeButtonView :theme="theme" />
    </li>
    <li v-for="n in 3" :key="n" v-else>
      <ThemeButtonView />
    </li>
  </ul>
</template>