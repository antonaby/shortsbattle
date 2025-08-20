<script setup lang="ts">
import { onUnmounted } from 'vue';
import { useWSStore } from './stores/wsStore';

const centrifugeStore = useWSStore();
centrifugeStore.connect();

onUnmounted(() => {
  centrifugeStore.disconnect()
})
</script>

<template>
  <div class="p-3">
    <!-- Loading Spinner -->
    <div v-if="!centrifugeStore.connected" class="flex items-center justify-center">
      <div class="flex items-center justify-center min-h-screen">
        <div class="animate-spin rounded-full h-12 w-12 border-4 border-gray-300 border-t-gray-600"></div>
      </div>
    </div>

    <!-- Content -->
    <main v-else>
      <router-view />
    </main>
  </div>
</template>

<style scoped></style>
