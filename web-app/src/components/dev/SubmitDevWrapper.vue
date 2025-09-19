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
	"stage": "submit-complete",
	"round": 2,
	"state_changed_at": "2025-09-18T15:36:44.804063+04:00",
	"remaining_ms": 5000
}

gameStore.playerVideo = {
	"id": 6,
	"game_id": 41,
	"player_id": 1,
	"round_n": 1,
	"submitted_at": "2025-09-19T15:04:53.092615+04:00",
	"video": {
		"id": 8,
		"video_url": "https://www.youtube.com/shorts/abx9X4nIoSo",
		"oembed": {
			"html": "\u003ciframe width=\"113\" height=\"200\" src=\"https://www.youtube.com/embed/abx9X4nIoSo?feature=oembed\" frameborder=\"0\" allow=\"accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share\" referrerpolicy=\"strict-origin-when-cross-origin\" allowfullscreen title=\"#веселыекоты\"\u003e\u003c/iframe\u003e",
			"type": "video",
			"title": "#веселыекоты",
			"width": 113,
			"height": 200,
			"version": "1.0",
			"author_url": "https://www.youtube.com/@MaysonandTayson",
			"author_name": "Tayson vs Mayson",
			"provider_url": "https://www.youtube.com/",
			"provider_name": "YouTube",
			"thumbnail_url": "https://i.ytimg.com/vi/abx9X4nIoSo/hq2.jpg",
			"thumbnail_width": 480,
			"thumbnail_height": 360
		},
		"added_at": "2025-09-19T14:27:33.326685+04:00",
		"updated_at": "2025-09-19T15:04:53.092615+04:00"
	}
};

const videoStore = useVideoStore();
videoStore.playerVideos = data;

onMounted(() => {
	gameStore.startTimer(5000);
})
</script>

<template>
	<SubmitView />
</template>