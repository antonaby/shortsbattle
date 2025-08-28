<script setup lang="ts">
import { ref } from 'vue';
import { useGameStore } from '../../stores/game.store';
import TimerView from './TimerView.vue';
import HeaderView from './HeaderView.vue';
import NewVideoView from './NewVideoView.vue';
import PlayerVideoLibraryView from './PlayerVideoLibraryView.vue';
import type { Tab } from '../../types/components';
import TabView from '../common/TabView.vue';

const tabs: Tab[] = [
  {
    position: 0,
    label: "Your videos",
    key: "library"
  },
  {
    position: 1,
    label: "Use New Video",
    key: "new"
  }
];

const gameStore = useGameStore();
const activeTab = ref<Tab>(tabs[0]);

function onTabSelect(tab: Tab) {
  activeTab.value = tab;
}
</script>
<template>
  <HeaderView />
  <div class="w-full">
    <TimerView caption="Submit your video" :remaning-time="gameStore.formattedTime" />
    <TabView :tabs="tabs" :selected-key="activeTab.key" @select="onTabSelect" />
    <div class="pt-4">
      <NewVideoView v-if="activeTab.key == 'new'"/>
      <PlayerVideoLibraryView v-if="activeTab.key == 'library'"/>
    </div>
  </div>
</template>