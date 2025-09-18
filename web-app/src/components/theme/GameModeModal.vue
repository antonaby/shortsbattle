<script lang="ts" setup>
import { computed } from 'vue';


const emits = defineEmits<{
  (e: 'ok'): void
  (e: 'cancel'): void
}>();

interface ModalText {
  titleEmoji: string
  title: string
  description: string
}

interface ModalTextOptions {
  [key: string]: ModalText;
}

const options: ModalTextOptions = {
  "join": {
    titleEmoji: "🚀",
    title: "Player Mode",
    description: "Submit your own videos, watch others, and vote to decide the winners."
  },
  "watch": {
    titleEmoji: "👀",
    title: "Watch Mode",
    description: "Sit back, watch the submissions, and cast your vote for your favorite videos."
  }
}

const props = defineProps<{
  mode: string,
  loading: boolean
}>();

const modalText = computed(() => {
  return options[props.mode];
})
</script>
<template>
  <div class="modal-fullscreen">
    <div class="modal-fullscreen-content">
      <h1 class="text-4xl">{{ modalText.titleEmoji }}</h1>
      <h1 class="text-title-2 text-center">
        {{ modalText.title }}
      </h1>
      <p class="text-description text-center">
        {{ modalText.description }}
      </p>
      <div class="flex gap-2 w-full">
        <button class="load-button flex-1" v-if="loading">
          <div class="loader"></div>
        </button>
        <button class="action-button flex-1" @click="emits('cancel')" v-if="!loading">
          <span>🙅</span>
          <span>Back</span>
        </button>
        <button class="action-button flex-1" @click="emits('ok')" v-if="!loading">
          <span>👍</span>
          <span>OK</span>
        </button>
      </div>
    </div>
  </div>
</template>