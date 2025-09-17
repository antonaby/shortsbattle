<script setup lang="ts">
defineProps<{
  msg: string
}>();
</script>

<template>
  <div class="modal-fullscreen">
    <div class="flex items-center justify-center flex-col rounded-xl bg-gray-100 p-6 gap-3">
      <span class="animate-spin rounded-full h-16 w-16 border-4 border-gray-300 border-t-gray-600"></span>
      <span>{{ msg }}</span>
    </div>
  </div>
</template>