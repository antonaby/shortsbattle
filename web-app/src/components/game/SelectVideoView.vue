<script setup lang="ts">
import { ref } from 'vue';
import NewVideoView from './NewVideoView.vue';
import ExistingVideoView from './ExistingVideoView.vue';
import type { Tab } from '../../types/components';
import TabView from '../common/TabView.vue';

const tabs: Tab[] = [
  {
    position: 0,
    label: "Your Videos",
    key: "library"
  },
  {
    position: 1,
    label: "New Video",
    key: "new"
  }
];

const activeTab = ref<Tab>(tabs[0]);

function onTabSelect(tab: Tab) {
  activeTab.value = tab;
}
</script>

<template>
  <TabView :tabs="tabs" :selected-key="activeTab.key" @select="onTabSelect" />
  <ExistingVideoView v-if="activeTab.key == 'library'" />
  <NewVideoView v-if="activeTab.key == 'new'" />
</template>