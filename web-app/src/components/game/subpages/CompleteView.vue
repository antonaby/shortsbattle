<script setup lang="ts">
import { useGameStore } from '@/stores/game.store';
import { useUIStore } from '@/stores/ui.store';

const uiStore = useUIStore();
const gameStore = useGameStore();
</script>

<template>
  <div class="page-container">
    <div class="text-2xl font-medium text-center">
      <span>Game Complete</span>
    </div>
    <ul class="space-y-2" v-if="gameStore.finalResult">
      <li v-for="result in gameStore.finalResult.videos">
        <div class="flex items-stretch justify-stretch rounded bg-gray-100">
          <div class="w-1/4 font-medium text-xl">
            Likes: {{ result.likes }}
          </div>
          <div class="w-1/4 font-medium text-xl">
            Dislikes: {{ result.dislikes }}
          </div>
        </div>
      </li>
    </ul>
    <div class="text-center">
      <button class="px-3 py-1 rounded bg-gray-100 hover:bg-gray-200" @click="uiStore.returnToHub()">
        <span class="text-xl font-medium">Return To Game List</span>
      </button>
    </div>
  </div>
</template>