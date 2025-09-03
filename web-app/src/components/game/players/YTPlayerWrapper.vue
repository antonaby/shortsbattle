<script setup lang="ts">
import YTPlayer from './YTPlayer.vue'

function stateChange(data: YT.PlayerState) {
  console.log(data);
}

</script>

<template>
  <div class="flex flex-col items-center justify-center bg-gray-200 w-screen h-screen">
    <YTPlayer video-id="https://www.youtube.com/shorts/Z5Jlxowr4IM" @state-change="stateChange" />
  </div>
</template>