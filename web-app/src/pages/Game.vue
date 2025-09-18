<script setup lang="ts">
import { onMounted, onUnmounted } from "vue";
import { useGameStore } from "../stores/game.store";
import { useRoute } from "vue-router";
import FullscreenLoaderView from "@/components/common/FullscreenLoaderView.vue";

const route = useRoute();
const gameStore = useGameStore();

onMounted(async () => {
  if (!gameStore.theme) {
    let id = Number(route.params.id);
    gameStore.joinGame(id);
  }
})

onUnmounted(() => {
  gameStore.leaveGame();
})
</script>

<template>
  <FullscreenLoaderView v-if="!gameStore.theme" />
  <router-view v-else />
</template>