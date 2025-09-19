<script lang="ts" setup>
import { ref, type Ref } from 'vue';

const url: Ref<string> = ref("")

function submit() {

}
</script>
<template>
  <form class="space-y-4 mt-2" @submit.prevent="submit">
    <p>
      Share a <span class="font-medium">YouTube Shorts</span> or 
      <span class="font-medium">TikTok</span> with the <span class="font-medium">Bot</span> 🤖
      <br/>Or paste a <span class="font-medium">Video URL</span> below 👇
    </p>
    <label class="block">
      <span class="block mb-1 text-sm">🔗 URL</span>
      <input v-model="url" type="url" name="url" placeholder="https://www.youtube.com/shorts/xyz" required class="w-full border rounded-lg p-2" />
    </label>
    <button type="submit" class="action-button w-full">⚡ Submit</button>
  </form>
</template>