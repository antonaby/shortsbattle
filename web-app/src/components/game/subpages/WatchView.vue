<script setup lang="ts">
import { onMounted, ref } from 'vue';
import { useGameStore } from '../../../stores/game.store';


const showControls = ref<boolean>(false);
function openControls() {
  showControls.value = true;
}
function closeControls() {
  showControls.value = false;
}

// TODO: ensure it's connected and subscribed with composables
const gameStore = useGameStore();
onMounted(async () => {
  await gameStore.loadVideosToWatch();
});
</script>

<template>
  <div class="w-screen h-screen bg-gray-200">
    <!-- TODO: add proper control panel with trabsition -->
    <div v-if="showControls" class="fixed top-1/2 -translate-y-1/2 right-0 z-50 p-5">
      <div class="flex flex-col items-right space-y-2">
        <button class="rounded-md bg-gray-100 p-2  font-medium" @click="gameStore.previousVideo">
          <span>Previous</span>
        </button>
        <button class="rounded-md bg-gray-100 p-2 font-medium" @click="gameStore.nextVideo">
          <span>Next</span>
        </button>
        <button class="rounded-md bg-gray-100 p-2 font-medium" @click="gameStore.likeVideo"
          v-if="gameStore.currentVideoOEmbed">
          <span>Like</span>
        </button>
        <button class="rounded-md bg-gray-100 p-2 font-medium" @click="gameStore.dislikeVideo"
          v-if="gameStore.currentVideoOEmbed">
          <span>Dislike</span>
        </button>
        <button class="rounded-md bg-gray-100 p-2 font-medium" @click="closeControls"
          v-if="gameStore.currentVideoOEmbed">
          <span>Minimize</span>
        </button>
      </div>
    </div>
    <div v-else class="fixed top-1/2 -translate-y-1/2 right-0 z-50 p-5">
      <button class="rounded-md bg-gray-100 p-2 font-medium" @click="openControls">
        <span>Open</span>
      </button>
    </div>
    <!-- <YTPlayer :oembed-html="gameStore.currentVideoOEmbed" v-if="gameStore.currentVideoOEmbed" /> -->
    <div v-else class="flex flex-col min-h-full min-w-full justify-center items-center text-2xl font-medium">
      <span>You have watched all videos!</span>
    </div>
  </div>
</template>