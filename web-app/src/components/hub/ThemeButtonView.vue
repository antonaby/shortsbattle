<script setup lang="ts">
import type { Theme } from '../../types/game';

const props = defineProps<{
  theme?: Theme
}>()

const emit = defineEmits<{
  (e: 'join', theme?: Theme): void
}>();
</script>

<template>
  <button @click="emit('join', props.theme)" class="w-full text-left hover:bg-gray-50 active:bg-gray-100 transition space-y-1 p-2 rounded-xl"
    :class="{ 'animate-pulse': !props.theme }">
    <div class="text-base font-medium text-gray-900" :class="{ 'w-2/3 bg-gray-200 rounded-lg': !props.theme }">
      {{ props.theme ? props.theme.title : "&nbsp;" }}
    </div>
    <div class="text-sm text-gray-500 mt-1" :class="{ 'w-5/6 bg-gray-200 rounded-lg': !props.theme }">
      {{ props.theme ? props.theme.description : "&nbsp;" }}
    </div>
  </button>
</template>