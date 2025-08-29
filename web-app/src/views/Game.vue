<script setup lang="ts">
import { onMounted, onUnmounted } from "vue";
import { useRoute } from "vue-router";
import { useGameStore } from "../stores/game.store";

const route = useRoute();
const gameStore = useGameStore();

onMounted(async () => {
  let id = Number(route.params.id);
  gameStore.joinGame(id);
})

onUnmounted(() => {
  gameStore.leaveGame();
})
</script>

<template>
  <router-view />
</template>