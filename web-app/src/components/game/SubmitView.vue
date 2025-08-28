<script setup lang="ts">
import { onMounted, ref, type Ref } from 'vue';
import { useGameStore } from '../../stores/game.store';
import TimerView from './TimerView.vue';
import HeaderView from './HeaderView.vue';

type TabType = "new" | "library"
interface Tab {
  position: number
  label: string
  type: TabType
}

const tabs: Tab[] = [
  {
    position: 0,
    label: "Your videos",
    type: "library"
  },
  {
    position: 1,
    label: "Use New Video",
    type: "new"
  }
];
const activeTab = ref<Tab>(tabs[0]);

const gameStore = useGameStore();
const url: Ref<string> = ref("")

onMounted(() => {
  gameStore.loadPlayerVideos();
})
</script>

<template>
  <HeaderView />
  <div class="w-full">
    <TimerView caption="Submit your video" :remaning-time="gameStore.formattedTime" />
    <div class="flex justify-around">
      <button
        v-for="tab in tabs"
        :key="tab.position"
        @click="activeTab = tab"
        class="px-4 py-2 focus:outline-none"
        :class="activeTab.type === tab.type
          ? 'text-blue-600 border-b-2 border-blue-600 -mb-px' 
          : 'text-gray-600 hover:text-blue-600'"
      >
        {{ tab.label }}
      </button>
    </div>
    <div class="pt-4">
      <form class="space-y-4" @submit.prevent="gameStore.submitVideo(url)" v-if="activeTab.type == 'new'">
        <label class="block">
          <span class="block mb-1">Video URL</span>
          <input v-model="url" type="url" name="url" placeholder="https://…" required
            class="w-full border rounded p-2" />
        </label>
        <button type="submit" class="w-full py-2 bg-black text-white rounded">Ok</button>
      </form>
      <div v-if="activeTab.type == 'library'">
        <span>List video</span>
      </div>
    </div>
  </div>
</template>