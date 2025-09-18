<script lang="ts" setup>
import { useGameStore } from '@/stores/game.store';
import LobbyView from '../game/subpages/LobbyView.vue';
import { onMounted } from 'vue';

const gameStore = useGameStore();
gameStore.theme = {
	"id": 3,
	"title": "Cat Clash",
	"description": "Share your funniest, cutest, or silliest cat videos and compete for votes.",
	"mode": "likedislike",
	"rounds": [
		{
			"round_n": 1,
			"title": "Curious Cats",
			"description": "Show cats exploring, investigating, or getting into unexpected places."
		},
		{
			"round_n": 2,
			"title": "Sleepy Whiskers",
			"description": "Share adorable moments of cats napping in funny or unusual spots."
		},
		{
			"round_n": 3,
			"title": "Playful Paws",
			"description": "Capture cats playing with toys, chasing, or just being mischievous."
		}
	]
};

onMounted(() => {
	gameStore.startTimer(60000);
})
</script>
<template>
  <LobbyView />
</template>