<script setup lang="ts">
import { useGameStore } from '@/stores/game.store';
import { useUIStore } from '@/stores/ui.store';
import VHIcon from '@/components/common/VHIcon.vue';


const uiStore = useUIStore();
const gameStore = useGameStore();
</script>

<template>
  <div class="page-container">
    <div class="stage-header">
      <span>🎉 Game Complete</span>
    </div>
    <div class="p-4 flex flex-col justify-start items-center gap-2">
      <span class="text-lg font-medium w-full text-center">Your Stats</span>
      <div class="text-lg font-semibold flex flex-wrap items-center gap-1 text-center">
        <span class="min-w-4">👍</span>
        <span class="text-green-700 min-w-6">10</span>
        <span class="min-w-4">👎</span>
        <span class="text-red-700 min-w-6">2</span>
        <span  class="min-w-4">⚡</span>
        <span class="text-green-700 min-w-6">+200</span>
      </div>
      <hr class="border-t-2 border-gray-200 h-1 w-full" />
      <span class="text-lg font-medium">⚡Score⚡</span>
      <div class="grid grid-cols-[24px_24px_1fr_1fr] place-items-start gap-1">

        <span>&nbsp;</span>
        <span>🥇</span>
        <span class="text-start text-gray-700">Player</span>
        <div class="flex gap-1 text-xs place-self-center">
          <span>👍</span>
          <span class="text-green-700">10</span>
          <span>👎</span>
          <span class="text-red-700">2</span>
        </div>

        <span>&nbsp;</span>
        <span>🥈</span>
        <span class=" text-gray-700">Player 1</span>
        <div class="flex gap-1 text-xs place-self-center">
          <span>👍</span>
          <span class="text-green-700">10</span>
          <span>👎</span>
          <span class="text-red-700">2</span>
        </div>

        <span>👉</span>
        <span>🥉</span>
        <span class="text-gray-700 font-semibold">Player PPP</span>
        <div class="flex gap-1 text-xs place-self-center">
          <span>👍</span>
          <span class="text-green-700">10</span>
          <span>👎</span>
          <span class="text-red-700">2</span>
        </div>

        <span>&nbsp;</span>
        <span>&nbsp;</span>
        <span class="text-start">Player Three</span>
        <div class="flex gap-1 text-xs place-self-center">
          <span>👍</span>
          <span class="text-green-700">10</span>
          <span>👎</span>
          <span class="text-red-700">2</span>
        </div>

      </div>
    </div>
    <button
      class="text-white p-2 font-semibold rounded-lg flex items-center justify-center gap-2 h-9 bg-gradient-to-r from-green-400 to-green-600 w-full"
      @click="uiStore.returnToHub()">
      🎉 Finish
    </button>

    <span class="text-lg font-medium w-full text-center">⚡Rounds⚡</span>

    <div class="sub-container">
      <div class="flex items-center gap-4 bg-gray-100 px-2 py-2 rounded-lg my-2">
        <img src="https://cdn2.thecatapi.com/images/bpc.jpg" alt="Round" class="w-16 h-16 object-cover rounded-lg" />
        <div class="flex-1">
          <h2 class="text-title-item">1. Curious Cats</h2>
          <p class="text-description">Capture cats playing with toys, chasing, or just being mischievous.</p>
        </div>
      </div>
      <!-- Video Info -->
      <div class="flex gap-3 w-full items-start">
        <div class="flex justify-center items-center bg-gray-100 rounded-lg overflow-hidden w-1/3">
          <img src="https://i.ytimg.com/vi/fM9Xt04ieIM/hq2.jpg" alt="Video Info" class="object-cover max-h-28"
            loading="lazy" />
        </div>
        <div class="flex-1 flex flex-col items-start gap-1">
          <div class="flex flex-wrap items-center justify-start gap-1">
            <div class="text-sm px-2 py-1 rounded-lg bg-gradient-to-r from-cyan-500 to-blue-600 text-white">
              <span>👤</span>
              <span class="font-semibold max-w-[120px] truncate">PlayerOne</span>
            </div>
            <div class="flex items-center">
              <VHIcon platform="tiktok" />
              <span class="text-description">
                Leovannn
              </span>
            </div>
          </div>
          <span class="text-sm">that is not his first time #memes #funny #shorts"</span>
          <div class="w-full flex gap-2">
            <span
              class="flex-1 px-2 py-1 rounded-lg text-center font-medium border border-green-300 text-green-700 text-sm">👍
              10</span>
            <span
              class="flex-1 px-2 py-1 rounded-lg text-center font-medium border border-red-300 text-red-700 text-sm">👎
              5</span>
          </div>
        </div>
      </div>
      <!-- Video Info -->
      <div class="flex gap-3 w-full items-start">
        <div class="flex justify-center items-center bg-gray-100 rounded-xl overflow-hidden w-1/3">
          <img src="https://i.ytimg.com/vi/fM9Xt04ieIM/hq2.jpg" alt="Video Info" class="object-cover max-h-28"
            loading="lazy" />
        </div>
        <div class="flex-1 flex flex-col items-start gap-1">
          <div class="flex items-center gap-2 text-sm px-2 py-1 rounded-lg bg-gradient-to-r from-cyan-500 to-blue-600
text-white">
            <span>👤</span>
            <span class="font-semibold max-w-[120px] truncate">PlayerTwo</span>
          </div>
          <div class="flex justify-start items-center gap-2">
            <VHIcon platform="youtube" />
            <span class="text-description">
              ХЛЕБЕЦ
            </span>
          </div>
          <span class="text-sm">that is not his first time #memes #funny #shorts"</span>
          <div class="w-full flex gap-2">
            <span
              class="flex-1 p-1 rounded-lg text-center font-medium border border-green-300 text-green-600 text-sm">👍
              10</span>
            <span class="flex-1 p-1 rounded-lg text-center font-medium border border-red-300 text-red-600 text-sm">👎
              5</span>
          </div>
        </div>
      </div>
      <!-- Video Info -->
      <div class="flex gap-3 w-full items-start">
        <div class="flex justify-center items-center bg-gray-100 rounded-xl overflow-hidden w-1/3">
          <img src="https://i.ytimg.com/vi/fM9Xt04ieIM/hq2.jpg" alt="Video Info" class="object-cover max-h-28"
            loading="lazy" />
        </div>
        <div class="flex-1 flex flex-col items-start gap-1">
          <div class="flex items-center gap-2 text-sm px-2 py-1 rounded-lg bg-gradient-to-r from-orange-400 via-pink-500 to-purple-600
text-white">
            <span>👤</span>
            <span class="font-semibold max-w-[120px] truncate">Yours</span>
          </div>
          <div class="flex justify-start items-center gap-2">
            <VHIcon platform="youtube" />
            <span class="text-description">
              ХЛЕБЕЦ
            </span>
          </div>
          <span class="text-sm">that is not his first time #memes #funny #shorts"</span>
          <div class="w-full flex gap-2">
            <span
              class="flex-1 p-1 rounded-lg text-center font-medium border border-green-300 text-green-600 text-sm">👍
              10</span>
            <span class="flex-1 p-1 rounded-lg text-center font-medium border border-red-300 text-red-600 text-sm">👎
              5</span>
          </div>
        </div>
      </div>
    </div>
    <div class="sub-container">
      <div class="flex items-center gap-4 bg-gray-100 px-2 py-2 rounded-lg my-2">
        <img src="https://cdn2.thecatapi.com/images/bpc.jpg" alt="Round" class="w-16 h-16 object-cover rounded-lg" />
        <div class="flex-1">
          <h2 class="text-title-item">2. Sleepy Whiskers</h2>
          <p class="text-description">Capture cats playing with toys, chasing, or just being mischievous.</p>
        </div>
      </div>
      <!-- Video Info -->
      <div class="flex gap-4 w-full items-start">
        <div class="flex justify-center items-center bg-gray-100 rounded-xl overflow-hidden w-1/3">
          <img src="https://i.ytimg.com/vi/fM9Xt04ieIM/hq2.jpg" alt="Video Info" class="object-cover max-h-28"
            loading="lazy" />
        </div>
        <div class="flex-1 flex flex-col items-start">
          <div class="flex justify-start items-center gap-2">
            <VHIcon platform="youtube" />
            <span class="text-description">
              ХЛЕБЕЦ
            </span>
          </div>
          <span class="text-sm">that is not his first time #memes #funny #shorts"</span>
          <div class="w-full flex gap-2 mt-2">
            <span
              class="flex-1 p-1 rounded-lg text-center font-medium border border-green-300 text-green-600 text-sm">👍
              10</span>
            <span class="flex-1 p-1 rounded-lg text-center font-medium border border-red-300 text-red-600 text-sm">👎
              5</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>