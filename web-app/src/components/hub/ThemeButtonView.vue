<script setup lang="ts">
import type { Theme } from '../../types/game';

const props = defineProps<{
  theme: Theme
}>()

const emit = defineEmits<{
  (e: 'open', theme: Theme): void
}>();
</script>

<template>
  <button 
    @click="emit('open', props.theme)" 
    class="w-full text-left hover:bg-gray-50 active:bg-gray-100 transition space-y-1 p-2 rounded-xl">
    <div class="text-base font-medium text-gray-900">
      {{ props.theme ? props.theme.title : "&nbsp;" }}
    </div>
    <div class="text-sm text-gray-500 mt-1">
      {{ props.theme ? props.theme.description : "&nbsp;" }}
    </div>
  </button>
</template>