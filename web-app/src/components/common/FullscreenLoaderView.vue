<script setup lang="ts">

withDefaults(defineProps<{
  grayBackground?: boolean
  msg?: string
}>(), {
  grayBackground: true,
  msg: ""
});
</script>

<template>
  <div class="modal-fullscreen" :class="[
    grayBackground ? 'bg-black/10' : 'bg-white'
  ]">
    <div class="loader-big"></div>
    <div v-if="msg" class="text-lg text-gray-600">
      {{ msg }}
    </div>
  </div>
</template>