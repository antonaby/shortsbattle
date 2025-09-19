<script lang="ts" setup>
import type { GameStage } from '@/types/game';
import { computed, onMounted, watch } from 'vue';

const props = defineProps<{
  formattedTime: string,
  stage?: GameStage,
  completeStage: GameStage
}>();

const emits = defineEmits<{
  (e: 'showTimer', value: boolean): void
}>();

const showTimer = computed<boolean>(() => {
  if (props.stage == props.completeStage) {
    return false;
  }

  return !!props.formattedTime;
})

watch(showTimer, (newValue) => {
  emits('showTimer', newValue);
});

onMounted(() => {
  emits('showTimer', showTimer.value);
})
</script>

<template>
  <span class="text-4xl font-mono font-semibold" v-if="showTimer">
    {{ formattedTime }}
  </span>
  <div class="loader-container h-10" v-else>
    <div class="loader-big"></div>
  </div>
</template>