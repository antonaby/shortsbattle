<script setup lang="ts">
import { computed, ref } from 'vue';
import YTPlayer from './YTPlayer.vue'
import TikTokPlayer from './TikTokPlayer.vue';
import { detectVideoPlatform, type VideoPlatform } from '@/utils/players';
import type { GameVideo, VoteValue } from '@/types/game';

interface CurrentVideo {
  gameVideoId: number
  videoUrl: string
  platform: VideoPlatform
}

const props = defineProps<{
  videos: GameVideo[]
}>();

const emits = defineEmits<{
  (e: 'vote', video: GameVideo, value: VoteValue): void
  (e: 'error', video: GameVideo, error: any): void
}>();

const showMenu = ref<boolean>(false);
const pinMenu = ref<boolean>(false);
const currentVideoIndex = ref<number>(0);

let currentVideo = computed<CurrentVideo>(() => {
  let video = props.videos[currentVideoIndex.value];

  return {
    gameVideoId: video.id,
    videoUrl: video.video.video_url,
    platform: detectVideoPlatform(video.video.video_url),
  };
})

function nextVideo() {
  if (currentVideoIndex.value < props.videos.length - 1) {
    currentVideoIndex.value += 1;
  } else {
    currentVideoIndex.value = -1;
  }
}

function vote(value: VoteValue) {
  let video = props.videos[currentVideoIndex.value];
  emits('vote', video, value);
  nextVideo();
}

function handleStateChange(state: YT.PlayerState) {
  if (state == 0 || state == 2) {
    showMenu.value = true;
  } else {
    showMenu.value = false;
  }
}

function handleError(error: any) {
  let video = props.videos[currentVideoIndex.value];
  emits('error', video, error);
  nextVideo();
}
</script>

<template>
  <div class="flex flex-col items-center bg-gray-800 w-screen h-screen space-y-2"
    :class="[currentVideoIndex >= 0 ? 'justify-start' : 'justify-center']">
    <div v-if="currentVideoIndex >= 0" class="transition-[width] duration-500 ease-in-out"
      :class="[showMenu || pinMenu ? 'w-10/11' : 'w-full']">
      <YTPlayer v-if="currentVideo.platform == 'youtube'" :video-url="currentVideo.videoUrl"
        @state-change="handleStateChange" @error="handleError" />
      <TikTokPlayer v-if="currentVideo.platform == 'tiktok'" :video-url="currentVideo.videoUrl"
        @state-change="handleStateChange" @error="handleError" />
    </div>
    <Transition v-if="currentVideoIndex >= 0" enter-active-class="transition duration-300 ease-out"
      enter-from-class="opacity-0 translate-y-2" enter-to-class="opacity-100 translate-y-0"
      leave-active-class="transition duration-200 ease-in" leave-from-class="opacity-100 translate-y-0"
      leave-to-class="opacity-0 translate-y-2">
      <div v-if="showMenu || pinMenu"
        class="bg-white/90 rounded-xl p-2 flex justify-center items-center space-x-2 shadow mb-2">
        <button class="w-16 rounded-md p-2 font-medium" @click="pinMenu = !pinMenu">
          <span v-if="pinMenu">📍</span>
          <span v-else>📌</span>
        </button>
        <button class="w-16 rounded-md p-2" @click="vote('like')">
          <span>👍</span>
        </button>
        <button class="w-16 rounded-md p-2" @click="vote('dislike')">
          <span>👎</span>
        </button>
      </div>
    </Transition>
    <div v-else class="bg-white/90 rounded-xl p-4 shadow flex flex-col items-center">
      <span class="text-2xl font-medium">
        You have watched all videos 🥳
      </span>
    </div>
  </div>
</template>