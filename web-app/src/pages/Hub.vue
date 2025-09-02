<script setup lang="ts">
import FullscreenLoaderView from '../components/common/FullscreenLoaderView.vue';
import ThemeList from '../components/hub/ThemeList.vue';
import { useGameHubStore } from '../stores/hub.store';

const hubStore = useGameHubStore();
</script>

<template>
  <FullscreenLoaderView v-if="hubStore.loadingGame" />
  <ThemeList />
</template>