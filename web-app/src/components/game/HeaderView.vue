<script setup lang="ts">
import type { ThemeDetails } from '@/types/game';
defineProps<{
  theme: ThemeDetails | null
}>()

const emit = defineEmits<{
  (e: 'return'): void
}>();
</script>

<template>
  <div class="flex">
    <button class="px-3 py-1 rounded bg-gray-100 hover:bg-gray-200 mr-4" @click="emit('return')">&larr; Back</button>
    <div class="flex-1">
      <h1 class="text-lg font-bold">{{ theme?.name || "&nbsp;" }}</h1>
      <p class="text-sm text-gray-600">{{ theme?.description || "&nbsp;" }}</p>
    </div>
  </div>
</template>