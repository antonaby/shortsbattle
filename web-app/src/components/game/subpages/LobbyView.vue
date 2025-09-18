<script setup lang="ts">
import { useUIStore } from '@/stores/ui.store';
import { useGameStore } from '@/stores/game.store';
import HeaderView from '@/components/game/HeaderView.vue';
import TimerView from '@/components/game/TimerView.vue';

const gameStore = useGameStore();
const uiStore = useUIStore();
</script>

<template>
  <div class="page-container">
    <img src="https://cdn2.thecatapi.com/images/8q1.jpg" alt="Main image" class="w-full h-48 object-cover rounded-lg" />
    <h1 class="text-title text-center">{{ gameStore.theme?.title }}</h1>
    <p class="text-description text-center">{{ gameStore.theme?.description }}</p>   
    <div class="bg-gray-100 rounded-lg p-4 flex flex-col justify-center items-center gap-2">
      <span>Waiting Room</span>
      <span class="text-4xl font-mono font-semibold">{{ gameStore.formattedTime }}</span>
      <span class="text-base font-light">🎶😎 Hang tight, others are on their way!</span>
    </div>
    <div class="flex gap-2 mt-4">
      <button class="action-button flex-1" @click="uiStore.returnToHub()">
        <span>🙅</span>
        <span>Quit</span>
      </button>
    </div>
  </div>
</template>