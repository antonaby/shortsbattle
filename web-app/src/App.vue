<script setup lang="ts">
import { useWSStore } from './stores/ws.store';
import FullscreenLoaderView from './components/common/FullscreenLoaderView.vue';
import { computed, onMounted, onUnmounted } from 'vue';
import { useUserStore } from './stores/user.store';

const userStore = useUserStore();
const wsStore = useWSStore();

const isReady = computed<boolean>(() => {
  return wsStore.connected
})

onMounted(async () => {
  await userStore.getToken();

  wsStore.connect(async () => {
    let token = await userStore.getToken();
    return token;
  });
})

onUnmounted(() => {
  wsStore.disconnect();
})
</script>

<template>
  <FullscreenLoaderView msg="Connecting..." v-if="!isReady" />
  <main v-else>
    <router-view />
  </main>
</template>

<style scoped></style>
