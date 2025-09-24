<script lang="ts" setup>
import { useGameStore } from '@/stores/game.store';
import CompleteView from '@/components/game/subpages/CompleteView.vue';
import result from './result.json'

const gameStore = useGameStore();
gameStore.gameResult = result;
</script>

<template>
  <CompleteView />
</template>