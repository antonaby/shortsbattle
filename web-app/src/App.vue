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
  <div class="p-3">
    <FullscreenLoaderView v-if="!uiStore.isReady" />
    <main v-else>
      <router-view />
    </main>
  </div>
</template>

<style scoped></style>
