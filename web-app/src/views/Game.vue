<script setup lang="ts">
import { onMounted, onUnmounted } from "vue";
import { useRoute } from "vue-router";
import { useGameStore } from "../stores/game.store";
import FullscreenLoaderView from "../components/common/FullscreenLoaderView.vue";

const route = useRoute();
const gameStore = useGameStore();

onMounted(() => {
  let id = Number(route.params.id);
  gameStore.joinGame(id);
})

onUnmounted(() => {
  gameStore.leaveGame();
})
</script>

<template>
  <div v-if="gameStore.lastGameUpdate" class="flex flex-col justify-start min-h-screen">
    <div class="flex-1">
      <router-view />
    </div>
  </div>
  <FullscreenLoaderView v-else />
</template>