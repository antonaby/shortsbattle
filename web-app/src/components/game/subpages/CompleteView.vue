<script setup lang="ts">
import { useGameStore } from '@/stores/game.store';
import { useUIStore } from '@/stores/ui.store';
import tkIcon from '@/assets/tiktok.png';
import VHIcon from '@/components/common/VHIcon.vue';


const uiStore = useUIStore();
const gameStore = useGameStore();
</script>

<template>
  <div class="page-container">
    <div class="stage-header">
      <span>🎉 Game Complete</span>
    </div>

    <button class="action-button w-full" @click="uiStore.returnToHub()">
      🎉 Finish
    </button>

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
          <div class="flex items-center gap-2 text-sm px-2 py-1 rounded-lg bg-gradient-to-r from-cyan-500 to-blue-600
text-white">
            <span>👤</span>
            <span class="font-semibold max-w-[120px] truncate">PlayerOne</span>
          </div>
          <div class="flex justify-start items-center gap-1">
            <VHIcon platform="tiktok" />
            <span class="text-description">
              Leovannn
            </span>
          </div>
          <span class="text-sm">that is not his first time #memes #funny #shorts"</span>
          <div class="w-full flex gap-2">
            <span
              class="flex-1 px-2 py-1 rounded-lg text-center font-medium border border-green-300 text-green-600 text-sm">👍
              10</span>
            <span
              class="flex-1 px-2 py-1 rounded-lg text-center font-medium border border-red-300 text-red-600 text-sm">👎
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