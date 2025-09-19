<script setup lang="ts">
import TimerView from '@/components/common/TimerView.vue';
import { useUIStore } from '@/stores/ui.store';
import { useGameStore } from '@/stores/game.store';
import { ref } from 'vue';

const gameStore = useGameStore();
const uiStore = useUIStore();

const showTimer = ref<boolean>(true);
function onShowTimer(value: boolean) {
  showTimer.value = value;
}
</script>

<template>
  <div class="page-container">
    <p class="text-2xl font-medium text-center">⏳ Waiting Room</p>
    <div class="bg-gray-100 rounded-lg p-4 flex flex-col justify-center items-center gap-2">
      <TimerView :formatted-time="gameStore.formattedTime" :stage="gameStore.lastUpdate?.stage"
        complete-stage="lobby-full" @show-timer="onShowTimer" />
      <span class="text-base font-light" v-if="showTimer">
        🎶😎 Hang tight, others are on their way!
      </span>
      <Transition enter-from-class="opacity-0" enter-active-class="transition ease-out duration-400"
        enter-to-class="opacity-100">
        <span class="text-base font-light" v-if="!showTimer">
          🎲⏳ Just a moment, the game begins shortly!
        </span>
      </Transition>
      <hr class="border-t-2 border-gray-200 h-1 w-full" />
      <p>⚔️ Players Joined: 5</p> <!-- TODO: add real number of players joined -->
    </div>
    <p class="text-title-2 text-center">🌟 Game Theme</p>
    <img src="https://cdn2.thecatapi.com/images/8q1.jpg" alt="Main image" class="w-full h-48 object-cover rounded-lg" />
    <h1 class="text-title text-center">{{ gameStore.theme?.title }}</h1>
    <p class="text-description text-center">{{ gameStore.theme?.description }}</p>
    <div class="flex gap-2 mt-4">
      <button class="action-button flex-1" @click="uiStore.returnToHub()">
        <span>🙅</span>
        <span>Quit</span>
      </button>
    </div>
  </div>
</template>