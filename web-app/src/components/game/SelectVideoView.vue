<script setup lang="ts">
import { ref } from 'vue';
import { useGameStore } from '../../stores/game.store';
import NewVideoView from './NewVideoView.vue';
import ExistingVideoView from './ExistingVideoView.vue';
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
  <TabView :tabs="tabs" :selected-key="activeTab.key" @select="onTabSelect" />
  <div class="pt-2" v-if="!gameStore.selectedVideo">
    <ExistingVideoView v-if="activeTab.key == 'library'" />
    <NewVideoView v-if="activeTab.key == 'new'" />
  </div>
</template>