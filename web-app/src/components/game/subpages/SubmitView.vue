<script setup lang="ts">
import TimerView from '@/components/game/TimerView.vue';
import HeaderView from '@/components/game/HeaderView.vue';
import SelectVideoView from '@/components/game/SelectVideoView.vue';
import { useGameStore } from '@/stores/game.store';
import SelectedVideoView from '@/components/game/SelectedVideoView.vue';
import { useUIStore } from '@/stores/ui.store';

const gameStore = useGameStore();
const uiStore = useUIStore();
</script>
<template>
  <div class="page-container">
    <HeaderView :theme="gameStore.theme" @return="uiStore.returnToHub()" />
    <TimerView :caption="gameStore.selectedVideo ? 'Awaiting other players' : 'Submit your video'" :remaning-time="gameStore.formattedTime" />
    <SelectVideoView v-if="!gameStore.selectedVideo" />
    <SelectedVideoView :video="gameStore.selectedVideo" @unselect="gameStore.unselectVideo" v-else />
  </div>
</template>