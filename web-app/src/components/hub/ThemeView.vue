<script setup lang="ts">
import { useRouter } from 'vue-router';
import { useGameHubStore } from '../../stores/hub.store';
import type { Theme } from '../../types/game';


const router = useRouter()
const gameHubStore = useGameHubStore();

async function joinAndOpenGame(themeId: number) {
  let game = await gameHubStore.joinGame(themeId)
  if (game) {
    router.push({ name: "game", params: { id: game?.game_id } });
  }
}

const props = defineProps<{
  theme: Theme,
}>()
</script>

<template>
  <button @click="joinAndOpenGame(props.theme.id)"
    class="w-full text-left py-2 hover:bg-gray-50 active:bg-gray-100 transition">
    <div class="text-base font-medium text-gray-900">
      {{ props.theme.name }}
    </div>
    <div class="text-sm text-gray-500 mt-1">
      {{ props.theme.description }}
    </div>
  </button>
</template>