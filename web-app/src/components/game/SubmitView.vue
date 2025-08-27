<script setup lang="ts">
import { ref, type Ref } from 'vue';
import { useGameStore } from '../../stores/game.store';
import TimerView from './TimerView.vue';
import HeaderView from './HeaderView.vue';

const gameStore = useGameStore();
const url: Ref<string> = ref("")
</script>

<template>
  <HeaderView />
  <form class="w-full mx-auto space-y-4" @submit.prevent="gameStore.submitVideo(url)">
    <TimerView caption="Submit your video" :remaning-time="gameStore.formattedTime" />
    <label class="block">
      <span class="block mb-1">Video URL</span>
      <input v-model="url" type="url" name="url" placeholder="https://…" required class="w-full border rounded p-2" />
    </label>
    <button type="submit" class="w-full py-2 bg-black text-white rounded">Submit</button>
  </form>
</template>