<script setup lang="ts">
import { useWSStore } from './stores/ws.store';
import { useUIStore } from './stores/ui.store';
import FullscreenLoaderView from './components/common/FullscreenLoaderView.vue';
import { onUnmounted } from 'vue';

const uiStore = useUIStore();

const wsStore = useWSStore();
wsStore.connect();

onUnmounted(() => {
  wsStore.disconnect();
})
</script>

<template>
  <FullscreenLoaderView v-if="!uiStore.isReady" />
  <main v-else>
    <router-view />
  </main>
</template>

<style scoped></style>
