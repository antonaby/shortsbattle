<script setup lang="ts">
import { useRouter } from 'vue-router';
import { useGameHubStore } from '../../stores/hub.store';
import type { Theme } from '../../types/game';

const props = defineProps<{
  theme?: Theme
}>()

const router = useRouter()
const gameHubStore = useGameHubStore();

async function joinAndOpenGame() {
  if (!props.theme) {
    return;
  }

  let game = await gameHubStore.joinGame(props.theme.id);
  router.push({ name: "game", params: { id: game.game_id } });
}
</script>

<template>
  <button @click="joinAndOpenGame()" class="w-full text-left py-2 hover:bg-gray-50 active:bg-gray-100 transition"
    :class="{ 'animate-pulse': !props.theme }">
    <div class="text-base font-medium text-gray-900" :class="{ 'w-2/3 bg-gray-200 rounded-lg': !props.theme }">
      {{ props.theme ? props.theme.name : "&nbsp;" }}
    </div>
    <div class="text-sm text-gray-500 mt-1" :class="{ 'w-5/6 bg-gray-200 rounded-lg': !props.theme }">
      {{ props.theme ? props.theme.description : "&nbsp;" }}
    </div>
  </button>
</template>