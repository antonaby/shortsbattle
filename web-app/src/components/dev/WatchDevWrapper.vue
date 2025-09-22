<script lang="ts" setup>
import { useGameStore } from '@/stores/game.store';
import WatchView from '../game/subpages/WatchView.vue';

const gameStore = useGameStore();
gameStore.videosToWatch = [
  // {
	// 	"id": 22,
	// 	"game_id": 57,
	// 	"player_id": 2,
	// 	"round_n": 1,
	// 	"submitted_at": "2025-09-22T14:34:34.545481+04:00",
	// 	"video": {
	// 		"id": 20,
	// 		"video_url": "https://www.youtube.com/shorts/IcLbrNT2IrE",
	// 		"oembed": {
	// 			"html": "<iframe width=\"113\" height=\"200\" src=\"https://www.youtube.com/embed/IcLbrNT2IrE?feature=oembed\" frameborder=\"0\" allow=\"accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share\" referrerpolicy=\"strict-origin-when-cross-origin\" allowfullscreen title=\"This Guy Had So Much Potential\"></iframe>",
	// 			"type": "video",
	// 			"title": "This Guy Had So Much Potential",
	// 			"width": 113,
	// 			"height": 200,
	// 			"version": "1.0",
	// 			"author_url": "https://www.youtube.com/@Knightfalledits",
	// 			"author_name": "Knightfall Edits",
	// 			"provider_url": "https://www.youtube.com/",
	// 			"provider_name": "YouTube",
	// 			"thumbnail_url": "https://i.ytimg.com/vi/IcLbrNT2IrE/hq2.jpg",
	// 			"thumbnail_width": 480,
	// 			"thumbnail_height": 360
	// 		},
	// 		"added_at": "2025-09-22T14:34:34.545481+04:00"
	// 	}
	// }
];
</script>
<template>
  <WatchView />
</template>