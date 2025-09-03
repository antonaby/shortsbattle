<script setup lang="ts">
import { ref } from 'vue';
import YTPlayer from './YTPlayer.vue'

const showMenu = ref<boolean>(false);
const pinMenu = ref<boolean>(false);
const currentVideoIndex = ref<number>(0);

let videos = [
  "https://www.youtube.com/shorts/wve9udh3qM4",
  "https://www.youtube.com/shorts/Z5Jlxowr4IM",
  "https://www.youtube.com/shorts/XUretomgAAA"
];

function stateChange(state: YT.PlayerState) {
  if (state == 0 || state == 2) {
    showMenu.value = true;
  } else if (!pinMenu.value) {
    showMenu.value = false;
  }
}

function handleError(error: YT.PlayerError) {
  // TODO: handle error when a vido can't be played as embedded
  console.log(error);
}

function nextVideo() {
  if (currentVideoIndex.value < videos.length - 1) {
    currentVideoIndex.value += 1;
  } else {
    currentVideoIndex.value = 0;
  }
}
</script>

<template>
  <div class="flex flex-col items-center justify-center bg-gray-200 w-screen h-screen space-y-2">
    <div 
      class="transition-[width] duration-500 ease-in-out" 
      :class="[showMenu ? 'w-7/8' : 'w-full']">
      <YTPlayer :video-id="videos[currentVideoIndex]" @state-change="stateChange" @error="handleError" />
    </div>
    <Transition 
      enter-active-class="transition duration-300 ease-out" 
      enter-from-class="opacity-0 translate-y-2"
      enter-to-class="opacity-100 translate-y-0" 
      leave-active-class="transition duration-200 ease-in"
      leave-from-class="opacity-100 translate-y-0" 
      leave-to-class="opacity-0 translate-y-2">
      <div v-if="showMenu" class="bg-white rounded-xl p-2 flex justify-center items-center space-x-2">
        <button class="w-16 rounded-md  p-2 font-medium" @click="pinMenu = !pinMenu">
          <span v-if="!pinMenu">📌</span>
          <span v-else>📍</span>
        </button>
        <button class="w-16 rounded-md  p-2 font-medium" @click="nextVideo">
          <span>🙈</span>
        </button>
        <button class="w-16 rounded-md  p-2 font-medium" @click="nextVideo">
          <span>👍</span>
        </button>
        <button class="w-16 rounded-md  p-2 font-medium" @click="nextVideo">
          <span>👎</span>
        </button>
      </div>
    </Transition>
  </div>
</template>