<script setup lang="ts">
import { useWSStore } from './stores/ws.store';
import { useUIStore } from './stores/ui.store';
import FullscreenLoaderView from './components/common/FullscreenLoaderView.vue';

const uiStore = useUIStore();

const wsStore = useWSStore();
wsStore.connect();

// onUnmounted(() => {
//   centrifugeStore.disconnect()
// })
</script>

<template>
  <div class="p-3">
    <!-- Loading Spinner -->
    <FullscreenLoaderView v-if="!uiStore.isReady" />

    <!-- Content -->
    <main v-else>
      <router-view />
    </main>
  </div>
</template>

<style scoped></style>
