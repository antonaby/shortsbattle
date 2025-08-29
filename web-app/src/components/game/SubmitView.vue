<script setup lang="ts">
import TimerView from './TimerView.vue';
import HeaderView from './HeaderView.vue';
import SelectVideoView from './SelectVideoView.vue';
import { useGameStore } from '../../stores/game.store';
import SelectedVideoView from './SelectedVideoView.vue';

const gameStore = useGameStore();
</script>
<template>
  <HeaderView />
  <TimerView caption="Submit your video" :remaning-time="gameStore.formattedTime" />
  <SelectVideoView v-if="!gameStore.selectedVideo" />
  <SelectedVideoView v-else />
</template>