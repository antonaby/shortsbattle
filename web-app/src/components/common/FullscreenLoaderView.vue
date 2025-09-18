<script setup lang="ts">
defineProps<{
  msg: string
}>();
</script>

<template>
  <div class="modal-fullscreen">
    <div class="rounded-xl bg-gray-100 flex flex-col items-center justify-center p-4 w-2/3 gap-3 max-w-sm">
      <span class="animate-spin rounded-full h-16 w-16 border-4 border-gray-300 border-t-gray-600"></span>
      <span>{{ msg }}</span>
    </div>
  </div>
</template>