<script setup lang="ts">
import { useGameStore } from "../stores/gameStore";

const gameStore = useGameStore();

</script>

<template>
  <div>
    <h1>Game Select</h1>
    <div>
      <buttom @click="gameStore.setGameId('game1')">Select Game 1</buttom>
    </div>
  </div>
</template>