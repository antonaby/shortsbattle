<script lang="ts" setup>
import { useVideoStore } from '@/stores/video.store';
import SubmitView from '@/components/game/subpages/SubmitView.vue';
import data from '@/components/dev/videos.json';
import { useGameStore } from '@/stores/game.store';
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

gameStore.lastUpdate = {
	"id": 27,
	"msg_type": "stage_updated",
	"stage": "submit",
	"round": 2,
	"state_changed_at": "2025-09-18T15:36:44.804063+04:00",
	"remaining_ms": 5000
}

const videoStore = useVideoStore();
videoStore.playerVideos = data;

onMounted(() => {
	gameStore.startTimer(5000);
})
</script>

<template>
  <SubmitView />
</template>