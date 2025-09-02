<script setup lang="ts">

defineProps<{
  oembedHtml: string
}>();
</script>

<template>
  <div v-html="oembedHtml" class="video-wrapper"></div>
</template>
