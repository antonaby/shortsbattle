<script setup lang="ts">
import { useGameStore } from '@/stores/game.store';
import { useUIStore } from '@/stores/ui.store';
import TabView from '@/components/common/TabView.vue';
import type { Tab } from '@/types/components';
import { computed, onMounted, ref } from 'vue';
import VideoScoreView from '@/components/common/VideoScoreView.vue';
import type { PlayerResult, RoundResult } from '@/types/game';
import RoundView from '@/components/common/RoundView.vue';
import BottomView from '@/components/common/BottomView.vue';
import { useUserStore } from '@/stores/user.store';

const tabs: Tab[] = [
  {
    position: 0,
    label: "🎥 All Videos",
    key: "all"
  },
  {
    position: 1,
    label: "👤 My Videos",
    key: "my"
  },
];

const activeTab = ref<Tab>(tabs[0]);
function onTabSelect(tab: Tab) {
  activeTab.value = tab;
}

const uiStore = useUIStore();
const gameStore = useGameStore();
const userStore = useUserStore();

const rounds = computed<RoundResult[]>(() => {
  if (!gameStore.gameResult) {
    return [];
  }

  if (activeTab.value.key == 'all') {
    return gameStore.gameResult.rounds;
  }

  return gameStore.gameResult.rounds
    .map(round => ({
      ...round,
      videos: round.videos.filter(v => v.author.tgId === userStore.userId)
    }))
    .filter(round => round.videos.length > 0)
})

const players = computed<PlayerResult[]>(() => {
  if (!gameStore.gameResult) {
    return [];
  }

  return [...gameStore.gameResult.players].sort((a, b) => a.place - b.place);
});

const player = computed<PlayerResult | undefined>(() => {
  if (!gameStore.gameResult) {
    return undefined;
  }

  let player = gameStore.gameResult.players.find(p => p.tg_id === userStore.userId)

  return player
})

function getPlaceMedal(place: number): string {
  switch (place) {
    case 1:
      return "🥇";
    case 2:
      return "🥈";
    case 3:
      return "🥉";
  }

  return "\u00A0"
}

onMounted(() => {
  gameStore.loadGameResult();
});
</script>

<template>
  <div class="page-container">
    <div class="stage-header">
      <span>🎉 Game Complete</span>
    </div>
    <div class="flex flex-col items-center" v-if="players.length > 0">
      <div class="grid grid-cols-[25px_1fr_50px_50px] place-items-start gap-1">
        <template v-for="player in players" :key="player.tg_id">
          <span>{{ getPlaceMedal(player.place) }}</span>
          <span class="text-gray-700" :class="{ 'font-medium': player.tg_id == userStore.userId }">
            {{ player.username }}
          </span>
          <div class="ld-table">
            <span>👍</span>
            <span class="text-emerald-700">{{ player.likes }}</span>
          </div>
          <div class="ld-table">
            <span>👎</span>
            <span class="text-rose-700">{{ player.dislikes }}</span>
          </div>
        </template>
      </div>
    </div>
    <div v-if="player"
      class="sub-container text-lg p-2 rounded-lg bg-gradient-to-br from-slate-50 to-slate-200">
      <div class="font-medium text-center text-gray-900">Your Stats</div>
      <div class="font-semibold flex flex-wrap items-center justify-center gap-1 text-center">
        <span class="min-w-4">👍</span>
        <span class="text-emerald-700 min-w-6">{{ player.likes }}</span>
        <span class="min-w-4">👎</span>
        <span class="text-rose-700 min-w-6">{{ player.dislikes }}</span>
        <span class="min-w-4">⚡</span>
        <span class="text-emerald-700 min-w-6">+{{ gameStore.gameResult?.outcome.plusEnergy }}</span>
      </div>
    </div>
    <button
      class="text-white p-2 font-semibold rounded-lg flex items-center justify-center gap-2 h-9 bg-gradient-to-r from-green-400 to-green-600 w-full"
      @click="uiStore.returnToHub()">
      🎉 Finish
    </button>
    <div class="sub-container" v-if="rounds.length > 0">
      <TabView :tabs="tabs" :selected-key="activeTab.key" @select="onTabSelect" />
      <template v-for="round in rounds" :key="round.round.round_n">
        <RoundView :round="round.round" />
        <template v-for="video in round.videos">
          <VideoScoreView :videoResult="video" :player-id="userStore.userId" />
        </template>
      </template>
    </div>
    <BottomView />
  </div>
</template>

<style>
@import "tailwindcss";

.ld-table {
  @apply flex gap-1 text-xs place-self-center w-full;
}
</style>