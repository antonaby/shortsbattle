<script setup lang="ts">
import { useRouter } from 'vue-router';

const router = useRouter();

function returnHome() {
  router.replace({ name: "hub" });
}
</script>

<template>
  <div class="flex mb-4">
    <button class="px-3 py-1 rounded bg-gray-100 hover:bg-gray-200 mr-4" @click="returnHome()">&larr; Back</button>
    <div class="flex-1">
      <h1 class="text-lg font-bold">TBD</h1>
      <p class="text-sm text-gray-600">TBD</p>
    </div>
  </div>
</template>