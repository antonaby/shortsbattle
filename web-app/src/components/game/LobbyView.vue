<script setup lang="ts">
import { useGameStore } from '../../stores/game.store';
import HeaderView from './HeaderView.vue';
import TimerView from './TimerView.vue';

const gameStore = useGameStore();
</script>

<template>
  <div class="page-container">
    <HeaderView />
    <TimerView caption="Waiting for other players" :remaning-time="gameStore.formattedTime" />
  </div>
</template>