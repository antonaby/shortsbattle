<script lang="ts" setup>
import type { Round } from '@/types/game';

defineProps<{
  round: Round
}>();
</script>

<template>
  <div class="flex flex-wrap items-start gap-2 bg-slate-100 p-2 rounded-lg">
    <img src="https://cdn2.thecatapi.com/images/bpc.jpg" alt="Round" class="w-20 h-20 object-cover rounded-lg" />
    <div class="flex-1">
      <p class="text-title-item">{{ round.round_n }}. {{ round.title }}</p>
      <p class="text-description">{{ round.description }}</p>
    </div>
  </div>
</template>