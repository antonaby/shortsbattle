<script setup lang="ts">
import { useUIStore } from '@/stores/ui.store';
import { useGameStore } from '@/stores/game.store';
import { computed } from 'vue';

const gameStore = useGameStore();
const uiStore = useUIStore();

const showTimer = computed<boolean>(() => {
  if (gameStore.lastUpdate?.stage == 'lobby-full') {
    return false;
  }

  return !!gameStore.formattedTime;
})

</script>

<template>
  <div class="page-container">
    <img src="https://cdn2.thecatapi.com/images/8q1.jpg" alt="Main image" class="w-full h-48 object-cover rounded-lg" />
    <h1 class="text-title text-center">{{ gameStore.theme?.title }}</h1>
    <p class="text-description text-center">{{ gameStore.theme?.description }}</p>
    <div class="bg-gray-100 rounded-lg p-4 flex flex-col justify-center items-center gap-2">
      <span>Waiting Room</span>
      <span class="text-4xl font-mono font-semibold" v-if="showTimer">
        {{ gameStore.formattedTime }}
      </span>
      <div class="flex flex-col items-center justify-center h-10" v-else>
        <div class="loader-big"></div>
      </div>
      <span class="text-base font-light" v-if="showTimer">
        🎶😎 Hang tight, others are on their way!
      </span>
      <Transition 
        enter-from-class="opacity-0" 
        enter-active-class="transition ease-out duration-400"
        enter-to-class="opacity-100" 
        leave-active-class="duration-0">
        <span class="text-base font-light" v-if="!showTimer">
          🎲⏳ Just a moment, the game begins shortly!
        </span>
      </Transition>
    </div>
    <div class="flex gap-2 mt-4">
      <button class="action-button flex-1" @click="uiStore.returnToHub()">
        <span>🙅</span>
        <span>Quit</span>
      </button>
    </div>
  </div>
</template>