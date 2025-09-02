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
    <HeaderView :theme="gameStore.theme" @return="uiStore.returnToHub()" />
    <TimerView caption="Waiting for other players" :remaning-time="gameStore.formattedTime" />
  </div>
</template>