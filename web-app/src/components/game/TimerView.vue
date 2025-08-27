<script setup lang="ts">

defineProps<{
  remaningTime: string
  caption: string
}>();
</script>

<template>
  <div class="flex items-center space-x-4">
    <p class="font-mono font-medium text-3xl text-gray-950">{{ remaningTime }}</p>
    <h1 class="flex-1 text-lg">{{ caption }}</h1>
  </div>
</template>