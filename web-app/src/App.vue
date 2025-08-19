<script setup lang="ts">
import { onMounted, onUnmounted } from 'vue';
import { useWSStore } from './stores/wsStore';

const centrifugeStore = useWSStore();

onMounted(() => {
  centrifugeStore.connect()
})

onUnmounted(() => {
  centrifugeStore.disconnect()
})
</script>

<template>
  <div class="flex flex-col min-h-screen bg-white text-gray-900">
    <!-- Loading screen -->
    <div v-if="!centrifugeStore.connected" class="flex-1 flex items-center justify-center">
      <div class="flex items-center gap-3">
        <!-- simple spinner -->
        <svg class="animate-spin h-6 w-6" viewBox="0 0 24 24" fill="none">
          <circle cx="12" cy="12" r="10" stroke="currentColor" opacity="0.2" stroke-width="4" />
          <path d="M22 12a10 10 0 0 1-10 10" stroke="currentColor" stroke-width="4" />
        </svg>
        <span>Connecting…</span>
      </div>
    </div>

    <!-- App content once connected -->
    <main v-else class="flex-1 px-4 py-6">
      <router-view />
    </main>
  </div>
</template>

<style scoped></style>
