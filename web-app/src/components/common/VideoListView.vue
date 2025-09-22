<script lang="ts" setup>
import type { Video } from '@/types/game';
import VideoButton from './VideoButton.vue';
import debounce from 'lodash.debounce'
import { ref, watch } from 'vue';

defineProps<{
  videos: Video[],
  loading: boolean
}>();

const emits = defineEmits<{
  (e: 'select', video: Video): void
  (e: 'search', query: string): void
}>();

function onClick(video: Video) {
  emits('select', video)
}

const query = ref<string>("");
const onInput = debounce((value) => {
  emits('search', value);
}, 400);
watch(query, value => onInput(value));

function onClear() {
  query.value = "";
}
</script>

<template>
  <div>
    <span class="block text-sm mb-1">🔍 Search</span>
    <label class="block relative">
      <button 
        class="absolute right-0 top-1/2 -translate-y-1/2 text-xs text-blue-600 p-2" 
        v-if="query" 
        @click="onClear">
        🧹clear
      </button>
      <input 
        v-model="query"
        placeholder="That funny cat video..." 
        class="w-full border rounded-lg px-2 py-1 outline-none text-sm" />
    </label>
  </div>
  <div class="loader-container h-16" v-if="loading">
    <div class="loader-big"></div>
  </div>
  <ul class="space-y-4 mt-4" v-if="!loading && videos.length > 0">
    <li v-for="video in videos">
      <VideoButton :video="video" @click="onClick" />
    </li>
  </ul>
  <div v-if="!loading && videos.length == 0" class="text-center text-lg text-gray-600">
    <span>🤷‍♂️ No videos found</span>
  </div>
</template>