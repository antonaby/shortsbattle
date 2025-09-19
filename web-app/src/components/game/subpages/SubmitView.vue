<script setup lang="ts">
import AddVideoFormView from '@/components/common/AddVideoFormView.vue';
import TabView from '@/components/common/TabView.vue';
import TimerView from '@/components/common/TimerView.vue';
import VideoListView from '@/components/common/VideoListView.vue';
import { useGameStore } from '@/stores/game.store';
import { useVideoStore } from '@/stores/video.store';
import type { Tab } from '@/types/components';
import type { Round } from '@/types/game';
import { computed, ref } from 'vue';

const videoStore = useVideoStore();
const gameStore = useGameStore();

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

const round = computed<Round | null>(() => {
  if (!gameStore.theme || !gameStore.theme.rounds || !gameStore.lastUpdate) {
    return null;
  }

  let found = gameStore.theme.rounds.find(v => v.round_n == gameStore.lastUpdate?.round);
  return found || null;
});

const showTimer = ref<boolean>(true);
function onShowTimer(value: boolean) {
  showTimer.value = value;
}
</script>
<template>
  <div class="page-container">
    <p class="text-2xl font-medium text-center">😎 Time to Drop a Video!</p>
    <div class="default-card">
      <p>✨ This Round`s Challenge</p>
      <div class="flex items-center gap-4 w-full" v-if="round">
        <img src="https://cdn2.thecatapi.com/images/bpc.jpg" alt="Round" class="w-16 h-16 object-cover rounded-lg" />
        <div class="flex-1">
          <h2 class="text-title-item">{{ round.round_n }}. {{ round.title }}</h2>
          <p class="text-description">{{ round.description }}</p>
        </div>
      </div>
      <div class="flex flex-col items-center justify-center h-16" v-else>
        <div class="loader-big"></div>
      </div>
      <hr class="border-t-2 border-gray-200 h-1 w-full" />
      <TimerView :formatted-time="gameStore.formattedTime" :stage="gameStore.lastUpdate?.stage"
        complete-stage="submit-complete" @show-timer="onShowTimer" />
      <span class="text-base font-light" v-if="showTimer">
        🔗 Paste URL or 📂 Pick a Video
      </span>
      <Transition enter-from-class="opacity-0" enter-active-class="transition ease-out duration-400"
        enter-to-class="opacity-100">
        <span class="text-base font-light" v-if="!showTimer">
          ⚡ Almost time — get ready!
        </span>
      </Transition>
    </div>
    <TabView :tabs="tabs" :selected-key="activeTab.key" @select="onTabSelect" />
    <AddVideoFormView v-if="activeTab.key == 'new'" />
    <VideoListView v-if="activeTab.key == 'library'" :videos="videoStore.playerVideos" />
  </div>
</template>