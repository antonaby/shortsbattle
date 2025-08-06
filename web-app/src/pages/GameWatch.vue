<script setup lang="ts">

</script>

<template>
  <div class="flex flex-col">
    <div class="shorts-container">
      <iframe
        width="315" height="560"
        src="https://www.youtube.com/embed/PdN_UbAnMj8"
        frameborder="0"
        allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture"
        allowfullscreen>
      </iframe>
    </div>
  </div>
</template>

<style>
  .shorts-container {
    position: relative;
    width: 100%;
    padding-bottom: 177%; /* 9:16 ratio */
    height: 0;
  }
  .shorts-container iframe {
    position: absolute;
    top: 0; left: 0;
    width: 100%;
    height: 100%;
  }
</style>