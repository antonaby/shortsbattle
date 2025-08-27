<script setup lang="ts">
import { computed, onMounted, onUnmounted, type Component } from "vue";
import { useWSStore } from "../stores/ws.store";
import { useRoute } from "vue-router";
import { useGameStore } from "../stores/game.store";

import LobbyView from "../components/game/LobbyView.vue";
import type { GameState } from "../types/game";
import HeaderView from "../components/game/HeaderView.vue";
import SubmittingView from "../components/game/SubmittingView.vue";

const { useView, viewName } = defineProps<{
  useView: boolean,
  viewName: GameState
}>();

const componentMap: Record<GameState, Component> = {
  lobby: LobbyView,
  submitting: SubmittingView,
  watching: LobbyView,
  completed: LobbyView
}

const route = useRoute();
const wsStore = useWSStore();
const gameStore = useGameStore();

const currentComponent = computed(() => {
  if (useView) {
    return componentMap[viewName];
  }

  if (!gameStore.lastGameUpdate) {
    return componentMap['lobby'];
  }

  return componentMap[gameStore.lastGameUpdate.state];
})

onMounted(() => {
  let id = Number(route.params.id);
  wsStore.subscribe(id);
})

onUnmounted(() => {
  gameStore.leaveGame();
  wsStore.unsubsribe();
})
</script>

<template>
  <div v-if="gameStore.lastGameUpdate && wsStore.subscribed" class="flex flex-col justify-start min-h-screen">
    <HeaderView v-if="gameStore.lastGameUpdate.state != 'watching'" />
    <div class="flex-1">
      <component :is="currentComponent" />
    </div>
  </div>

  <div v-else class="flex items-center justify-center min-h-screen">
    <div class="animate-spin rounded-full h-12 w-12 border-4 border-gray-300 border-t-gray-600"></div>
  </div>
</template>