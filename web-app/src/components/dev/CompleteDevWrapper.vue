<script lang="ts" setup>
import { useGameStore } from '@/stores/game.store';
import CompleteView from '@/components/game/subpages/CompleteView.vue';
import result from './result.json'
import { useUserStore } from '@/stores/user.store';

const gameStore = useGameStore();
gameStore.gameResult = result;

const userStore = useUserStore();
userStore.userId = 202;

</script>

<template>
  <CompleteView />
</template>