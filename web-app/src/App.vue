<script setup lang="ts">
import { useWSStore } from './stores/ws.store';
import { useUIStore } from './stores/ui.store';

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
    <div v-if="!uiStore.isReady" class="fixed">
      <div class="flex items-center justify-center min-h-screen min-w-screen flex-col">
        <span class="animate-spin rounded-full h-12 w-12 border-4 border-gray-300 border-t-gray-600"></span>
        <span class="text-sm">Loading…</span>
      </div>
    </div>

    <!-- Content -->
    <main v-else>
      <router-view />
    </main>
  </div>
</template>

<style scoped></style>
