<script setup lang="ts">
import { onMounted } from 'vue'
import { useGameHubStore } from '../stores/hub.store';
import ThemeView from '../components/hub/ThemeView.vue';

const gameHubStore = useGameHubStore();

onMounted(() => {
  gameHubStore.fetchThemes();
})
</script>

<template>
  <ul class="divide-y divide-gray-100">
    <li v-for="(theme, index) in gameHubStore.themes" :key="index">
      <ThemeView :theme="theme"></ThemeView>
    </li>
  </ul>
</template>