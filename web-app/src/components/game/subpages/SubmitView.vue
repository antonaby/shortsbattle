<script setup lang="ts">
import AddVideoFormView from '@/components/common/AddVideoFormView.vue';
import TabView from '@/components/common/TabView.vue';
import TimerView from '@/components/common/TimerView.vue';
import VideoButton from '@/components/common/VideoButton.vue';
import VideoListView from '@/components/common/VideoListView.vue';
import { useGameStore } from '@/stores/game.store';
import { useVideoStore } from '@/stores/video.store';
import type { Tab } from '@/types/components';
import type { Round, Video } from '@/types/game';
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

const loadPlayerVideos = ref<boolean>(false);
async function onSearchVideo(query: string) {
  loadPlayerVideos.value = true;
  await videoStore.loadVideos(query);
  loadPlayerVideos.value = false;
}

const showTimer = ref<boolean>(true);
function onShowTimer(value: boolean) {
  showTimer.value = value;
}

const uploadingVideo = ref<boolean>(false);

async function onSubmitUrl(url: string) {
  uploadingVideo.value = true;
  await gameStore.submitNewVideo(url);
  uploadingVideo.value = false;
}

async function onSubmitVideo(video: Video) {
  uploadingVideo.value = true;
  await gameStore.submitVideo(video);
  uploadingVideo.value = false;
}
</script>
<template>
  <div class="page-container">
    <p class="stage-header">😎 Time to Drop a Video!</p>
    <div class="default-card">
      <p>✨ This Round`s Challenge</p>
      <div class="flex items-center gap-4 w-full" v-if="round">
        <img src="https://cdn2.thecatapi.com/images/bpc.jpg" alt="Round" class="w-16 h-16 object-cover rounded-lg" />
        <div class="flex-1">
          <h2 class="text-title-item">{{ round.round_n }}. {{ round.title }}</h2>
          <p class="text-description">{{ round.description }}</p>
        </div>
      </div>
      <div class="loader-container h-16" v-else>
        <div class="loader-big"></div>
      </div>
      <hr class="border-t-2 border-gray-200 h-1 w-full" />
      <TimerView :formatted-time="gameStore.formattedTime" :stage="gameStore.lastUpdate?.stage"
        complete-stage="submit-complete" @show-timer="onShowTimer" />
      <span class="text-base font-light" v-if="showTimer">
        🔗 Paste URL or 📂 Pick a Video
      </span>
      <Transition enter-from-class="opacity-100" enter-active-class="transition ease-out duration-400"
        enter-to-class="opacity-100">
        <span class="text-base font-light" v-if="!showTimer">
          ⚡ Almost done — get ready!
        </span>
      </Transition>
    </div>
    <!--TODO: handle the case when a user hasn't submitted any video -->
    <div class="sub-container" v-if="gameStore.playerVideo">
      <p class="w-full text-center">✨ Your Video</p>
      <VideoButton :video="gameStore.playerVideo.video" />
      <Transition 
        leave-active-class="transition ease-out duration-400" 
        leave-from-class="opacity-100"
        leave-to-class="opacity-0">
        <button class="text-blue-600" @click="gameStore.playerVideo = undefined"
          v-if="gameStore.lastUpdate?.stage == 'submit'">
          <span>🔁 Use Antother</span>
        </button>
      </Transition>
    </div>
    <div class="sub-container" v-if="!uploadingVideo && !gameStore.playerVideo">
      <TabView :tabs="tabs" :selected-key="activeTab.key" @select="onTabSelect" />
      <AddVideoFormView v-if="activeTab.key == 'new'" @submit="onSubmitUrl" />
      <VideoListView 
        v-if="activeTab.key == 'library'" 
        :videos="videoStore.playerVideos" 
        @select="onSubmitVideo" 
        @search="onSearchVideo" 
        :loading="loadPlayerVideos" />
    </div>
    <div class="loader-container gap-4 py-4" v-if="uploadingVideo && !gameStore.playerVideo">
      <p class="text-2xl font-medium text-center">🚀 Submit & Go!</p>
      <div class="loader-big"></div>
    </div>
  </div>
</template>