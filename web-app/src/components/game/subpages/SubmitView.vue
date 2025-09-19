<script setup lang="ts">
import AddVideoFormView from '@/components/common/AddVideoFormView.vue';
import TabView from '@/components/common/TabView.vue';
import TimerView from '@/components/common/TimerView.vue';
import VideoListView from '@/components/common/VideoListView.vue';
import { useGameStore } from '@/stores/game.store';
import { useUIStore } from '@/stores/ui.store';
import { useVideoStore } from '@/stores/video.store';
import type { Tab } from '@/types/components';
import { computed, ref } from 'vue';

const videoStore = useVideoStore();
const gameStore = useGameStore();
const uiStore = useUIStore();

const tabs: Tab[] = [
  {
    position: 0,
    label: "🔗 Paste New",
    key: "new"
  },
  {
    position: 1,
    label: "📂 My Videos",
    key: "library"
  },
];

const activeTab = ref<Tab>(tabs[0]);
function onTabSelect(tab: Tab) {
  activeTab.value = tab;
}
</script>
<template>
  <div class="page-container">
    <p class="text-2xl font-medium text-center">😎 Time to Drop a Video!</p>
    <div class="default-card">
      <p>✨ This Round`s Challenge</p>
      <div class="flex items-center gap-4">
        <img src="https://cdn2.thecatapi.com/images/bpc.jpg" alt="Round" class="w-16 h-16 object-cover rounded-lg" />
        <div class="flex-1 ">
          <h2 class="text-title-item">1. Playful Paws</h2>
          <p class="text-description">Capture cats playing with toys, chasing, or just being mischievous.</p>
        </div>
      </div>
      <hr class="border-t-2 border-gray-200 h-1 w-full" />
      <TimerView 
        :formatted-time="gameStore.formattedTime" 
        :stage="gameStore.lastUpdate?.stage"
        complete-stage="submit-complete" />
      <span class="text-base font-light">
        🔗 Paste URL or 📂 Pick a Video
      </span>
    </div>
    <TabView :tabs="tabs" :selected-key="activeTab.key" @select="onTabSelect" />
    <AddVideoFormView v-if="activeTab.key == 'new'" />
    <VideoListView v-if="activeTab.key == 'library'" :videos="videoStore.playerVideos" />
  </div>
</template>