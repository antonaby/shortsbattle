<script setup lang="ts">

</script>

<template>
  <div class="flex flex-col min-h-screen bg-white text-gray-900">
    <!-- Header -->
    <header class="p-4 bg-gray-100 border-b border-gray-300 text-center text-lg font-semibold">
      Shorts Battle 🎮
    </header>

    <!-- Main content area (natural height) -->
    <main class="flex-1 px-4 py-6">
      <router-view />
    </main>

    <!-- Footer -->
    <footer class="p-4 text-center text-sm text-gray-500 bg-gray-50 border-t border-gray-200">
      Telegram Web App
    </footer>
  </div>
</template>

<style scoped>

</style>
