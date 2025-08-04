<script setup lang="ts">
import { useBaseStore } from '../stores/base';



defineProps<{ msg: string }>()

const store = useBaseStore()
</script>

<template>
  <div class="flex items-center justify-center min-h-screen bg-gray-100">
    <div class="bg-white p-6 rounded-xl shadow-lg w-80">
      <h2 v-if="store.loading || store.msg == ''" class="text-xl font-bold mb-2">Loading...</h2>
      <h2 v-else class="text-xl font-bold mb-2">{{ store.msg }}</h2>
      <button class="mt-4 px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600" @click="store.getMessage()">
        Load Message
      </button>
      <button class="mt-4 px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600" @click="store.initCentriguge()">
        Connect to Centrifuge
      </button>
    </div>
  </div>
</template>

<style scoped>

</style>
