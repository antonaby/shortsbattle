<script setup lang="ts">
import { computed, ref } from 'vue';
import YTPlayer from './YTPlayer.vue'
import TikTokPlayer from './TikTokPlayer.vue';
import { detectVideoPlatform } from '@/utils/players';

const props = defineProps<{
  videoUrls: string[]
}>();

const showMenu = ref<boolean>(false);
const pinMenu = ref<boolean>(false);
const currentVideoIndex = ref<number>(0);

let platform = computed(() => {
  return detectVideoPlatform(props.videoUrls[currentVideoIndex.value])
})

function handleStateChange(state: YT.PlayerState) {
  if (state == 0 || state == 2) {
    showMenu.value = true;
  } else {
    showMenu.value = false;
  }
}

function handleErrorYT(error: YT.PlayerError) {
  // TODO: handle error when a vido can't be played as embedded
  console.log(error);
}

function handleErrorTT(error: number) {
  // TODO: handle error when a vido can't be played as embedded
  console.log(error);
}

function nextVideo() {
  if (currentVideoIndex.value < props.videoUrls.length - 1) {
    currentVideoIndex.value += 1;
  } else {
    currentVideoIndex.value = -1;
  }
}
</script>

<template>
  <div class="flex flex-col items-center bg-gray-800 w-screen h-screen space-y-2"
    :class="[currentVideoIndex >= 0 ? 'justify-start' : 'justify-center']"
    >
    <div v-if="currentVideoIndex >= 0"
      class="transition-[width] duration-500 ease-in-out" 
      :class="[showMenu || pinMenu ? 'w-10/11' : 'w-full']">
      <YTPlayer v-if="platform == 'youtube'"
        :video-url="props.videoUrls[currentVideoIndex]" 
        @state-change="handleStateChange" 
        @error="handleErrorYT" />
      <TikTokPlayer v-if="platform == 'tiktok'"
        :video-url="props.videoUrls[currentVideoIndex]" 
        @state-change="handleStateChange" 
        @error="handleErrorTT" />
    </div>
    <Transition v-if="currentVideoIndex >= 0"
      enter-active-class="transition duration-300 ease-out" 
      enter-from-class="opacity-0 translate-y-2"
      enter-to-class="opacity-100 translate-y-0" 
      leave-active-class="transition duration-200 ease-in"
      leave-from-class="opacity-100 translate-y-0" 
      leave-to-class="opacity-0 translate-y-2">
      <div v-if="showMenu || pinMenu" class="bg-white/90 rounded-xl p-2 flex justify-center items-center space-x-2 shadow">
        <button class="w-16 rounded-md p-2 font-medium" @click="pinMenu = !pinMenu">
          <span v-if="pinMenu">📍</span>
          <span v-else>📌</span>
        </button>
        <button class="w-16 rounded-md p-2" @click="nextVideo">
          <span>🙈</span>
        </button>
        <button class="w-16 rounded-md p-2" @click="nextVideo">
          <span>👍</span>
        </button>
        <button class="w-16 rounded-md p-2" @click="nextVideo">
          <span>👎</span>
        </button>
      </div>
    </Transition>
    <div v-else class="bg-white/90 rounded-xl p-4 shadow flex flex-col items-center">
      <span class="text-2xl font-medium">
        You have watched all videos 🥳
      </span>
      <button class="w-16 rounded-md p-2 text-2xl">
        <span>👍</span>
      </button>
    </div>
  </div>
</template>