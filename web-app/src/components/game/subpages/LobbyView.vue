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
    <!-- 3) Description -->
    <p class="text-description text-center">{{ gameStore.theme?.description }}</p>
    <!-- <HeaderView :theme="gameStore.theme" @return="uiStore.returnToHub()" /> -->
    <!-- <TimerView caption="Waiting for other players" :remaning-time="gameStore.formattedTime" /> -->
  </div>
</template>