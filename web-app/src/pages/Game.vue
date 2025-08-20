<script setup lang="ts">
import { computed, onMounted, onUnmounted, type Component } from "vue";
import { useWSStore } from "../stores/wsStore";
import { useRoute } from "vue-router";
import { useGameStore } from "../stores/gameStore";

import CreatedView from "../components/game/CreatedView.vue";
import LobbyView from "../components/game/LobbyView.vue";
import type { GameStatus } from "../models/common";
import HeaderView from "../components/game/HeaderView.vue";
import SubmittingView from "../components/game/SubmittingView.vue";

const { useView, viewName } = defineProps<{
  useView: boolean,
  viewName: GameStatus
}>();

const componentMap: Record<GameStatus, Component> = {
  created: CreatedView,
  lobby: LobbyView,
  submitting: SubmittingView,
  voting: CreatedView,
  winner: CreatedView,
  complete: CreatedView
}

const route = useRoute();
const wsStore = useWSStore();
const gameStore = useGameStore();

const currentComponent = computed(() => {
  if (useView) {
    return componentMap[viewName];
  }

  if (!gameStore.gameInstance) {
    return componentMap['created'];
  }

  return componentMap[gameStore.gameInstance.status];
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
  <div v-if="gameStore.gameInstance && wsStore.subscribed" class="flex flex-col justify-start min-h-screen">
    <HeaderView v-if="gameStore.gameInstance.status != 'voting'" />
    <div class="flex-1">
      <component :is="currentComponent" />
    </div>
  </div>

  <div v-else class="flex items-center justify-center min-h-screen">
    <div class="animate-spin rounded-full h-12 w-12 border-4 border-gray-300 border-t-gray-600"></div>
  </div>
</template>