<script setup lang="ts">
import type { Tab } from '@/types/components';

const props = defineProps<{
  selectedKey: string
  tabs: Tab[];
}>();

const emit = defineEmits<{
  (e: 'select', tab: Tab): void
}>();
</script>

<template>
  <div class="flex justify-around">
    <button v-for="tab in props.tabs" :key="tab.position" @click="emit('select', tab)" class="px-4 py-2 focus:outline-none"
      :class="tab.key === selectedKey
        ? 'text-blue-600 border-b-2 border-blue-600'
        : 'text-gray-600 border-b-2 hover:text-blue-600 border-transparent'">
      {{ tab.label }}
    </button>
  </div>
</template>