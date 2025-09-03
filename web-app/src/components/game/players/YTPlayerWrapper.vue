<script setup lang="ts">
import { ref } from 'vue';
import YTPlayer from './YTPlayer.vue'

const showMenu = ref<boolean>(false);

function stateChange(state: YT.PlayerState) {
  if (state == 0 || state == 2) {
    showMenu.value = true;
  } else {
    showMenu.value = false;
  }
}
</script>

<template>
  <div class="flex flex-col items-center justify-center bg-gray-200 w-screen h-screen space-y-2">
    <Transition
      enter-active-class="transition duration-300 ease-out"
      enter-from-class="-translate-x-full opacity-0"
      enter-to-class="translate-x-0 opacity-100"
      leave-active-class="transition duration-300 ease-in"
      leave-from-class="translate-x-0 opacity-100"
      leave-to-class="-translate-x-full opacity-0">
    <div v-if="!showMenu" class="fixed top-1/2 -translate-y-1/2 left-0 z-50 pl-0 pt-3 pr-3 pb-3">
      <button class="rounded-r-lg bg-white shadow p-2" @click="showMenu = true">
        <!-- TODO: add image instead of emoji -->
        <span>🎮</span>
      </button>
    </div>
    </Transition>
    <div 
      class="transition-[width] duration-500 ease-in-out" 
      :class="[showMenu ? 'w-7/8' : 'w-full']">
      <YTPlayer video-id="https://www.youtube.com/shorts/Z5Jlxowr4IM" @state-change="stateChange" />
    </div>
    <Transition 
      enter-active-class="transition duration-300 ease-out" 
      enter-from-class="opacity-0 translate-y-2"
      enter-to-class="opacity-100 translate-y-0" 
      leave-active-class="transition duration-200 ease-in"
      leave-from-class="opacity-100 translate-y-0" 
      leave-to-class="opacity-0 translate-y-2">
      <div v-if="showMenu" class="bg-white rounded-xl p-2 flex justify-center items-center space-x-2">
        <button class="w-16 rounded-md  p-2 font-medium">
        <span>⏮️</span>
      </button>
      <button class="w-16 rounded-md  p-2 font-medium">
        <span>⏭️</span>
      </button>
      <button class="w-16 rounded-md  p-2 font-medium">
        <span>👍</span>
      </button>
      <button class="w-16 rounded-md  p-2 font-medium">
        <span>👎</span>
      </button>
      </div>
    </Transition>
  </div>
</template>