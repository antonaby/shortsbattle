<template>
  <div class="flex flex-col gap-3">
    <!-- 9:16 viewport container -->
    <div class="video-wrapper">
      <div id="yt-shorts-vue" class="absolute inset-0"></div>
    </div>

    <div v-if="showUi" class="flex flex-wrap gap-2 items-center">
      <button class="px-3 py-2 rounded bg-gray-200 hover:bg-gray-300" @click="play">Play</button>
      <button class="px-3 py-2 rounded bg-gray-200 hover:bg-gray-300" @click="pause">Pause</button>
      <button class="px-3 py-2 rounded bg-gray-200 hover:bg-gray-300" @click="() => seek(-5)">⏪ 5s</button>
      <button class="px-3 py-2 rounded bg-gray-200 hover:bg-gray-300" @click="() => seek(5)">⏩ 5s</button>
      <button class="px-3 py-2 rounded bg-gray-200 hover:bg-gray-300" @click="toggleMute">Mute/Unmute</button>
      <button class="px-3 py-2 rounded bg-gray-200 hover:bg-gray-300" @click="() => setRate(1)">1×</button>
      <button class="px-3 py-2 rounded bg-gray-200 hover:bg-gray-300" @click="() => setRate(1.5)">1.5×</button>
      <button class="px-3 py-2 rounded bg-gray-200 hover:bg-gray-300" @click="() => setRate(2)">2×</button>
      <button class="px-3 py-2 rounded bg-gray-200 hover:bg-gray-300" @click="prev">Prev</button>
      <button class="px-3 py-2 rounded bg-gray-200 hover:bg-gray-300" @click="next">Next</button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, onBeforeUnmount, ref, watch, computed, defineExpose, withDefaults } from 'vue'

declare global {
  interface Window {
    onYouTubeIframeAPIReady?: () => void
  }
}

export interface Props {
  /** Array of Shorts/YouTube IDs or URLs */
  ids: string[]
  /** Player width in px (height computed to 9:16) */
  width?: number
  /** Show built-in UI controls below */
  showUi?: boolean
  /** Start playback muted (helps autoplay on mobile) */
  autoplayMuted?: boolean
  /** Loop the current video */
  loopSingle?: boolean
  /** Show YouTube controls bar */
  controlsBar?: 0 | 1
  /** Reduce YouTube branding */
  modestBranding?: 0 | 1
  /** Play inline on iOS */
  playsInline?: 0 | 1
  /** Origin (set to your site in production for security) */
  origin?: string
}

export type ExposedMethods = {
  play: () => void
  pause: () => void
  seek: (delta: number) => void
  toggleMute: () => void
  setRate: (r: number) => void
  next: () => void
  prev: () => void
}

const p = withDefaults(defineProps<Props>(), {
  width: 360,
  showUi: true,
  autoplayMuted: false,
  loopSingle: true,
  controlsBar: 1 as 0 | 1,
  modestBranding: 1 as 0 | 1,
  playsInline: 1 as 0 | 1,
  origin: ''
})

const width = p.width
const showUi = p.showUi

const normalizedIds = computed(() => p.ids.map(extractId).filter(Boolean))
const index = ref(0)
let player: YT.Player | null = null

onMounted(async () => {
  await ensureYouTubeIframeAPI()
  createPlayer()
  window.addEventListener('keydown', onKey)
})

onBeforeUnmount(() => {
  window.removeEventListener('keydown', onKey)
})

watch(index, (i) => {
  if (player && normalizedIds.value.length) {
    player.loadVideoById(normalizedIds.value[i])
  }
})

watch(normalizedIds, (arr) => {
  if (!arr.length) return
  if (player) {
    player.loadVideoById(arr[index.value])
  }
})

function createPlayer() {
  const videoId = normalizedIds.value[0]
  // Guard for SSR/no IDs
  if (!videoId || typeof window === 'undefined') return
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  player = new (window as any).YT.Player('yt-shorts-vue', {
    videoId,
    playerVars: {
      autoplay: p.autoplayMuted ? 1 : 0,
      controls: p.controlsBar,
      playsinline: p.playsInline,
      modestbranding: p.modestBranding,
      rel: 0,
      loop: p.loopSingle ? 1 : 0,
      playlist: p.loopSingle ? videoId : undefined,
      origin: p.origin || undefined,
      enablejsapi: 1
    },
    events: {
      onReady: (e: { target: YT.Player }) => {
        if (p.autoplayMuted) {
          try { e.target.mute() } catch { }
          try { e.target.playVideo() } catch { }
        }
      }
    }
  })
}

function ensureYouTubeIframeAPI(): Promise<void> {
  if (typeof window === 'undefined') { 
    return Promise.resolve(); 
  }

  if (window.YT?.Player) { 
    return Promise.resolve() 
  }

  return new Promise((resolve) => {
    const prev = window.onYouTubeIframeAPIReady
    window.onYouTubeIframeAPIReady = () => {
      prev && prev()
      resolve()
    }
    const tag = document.createElement('script')
    tag.src = 'https://www.youtube.com/iframe_api'
    document.head.appendChild(tag)
  })
}

// Public controls
function play(): void { player?.playVideo?.() }
function pause(): void { player?.pauseVideo?.() }
function seek(delta: number): void {
  if (!player?.getCurrentTime) return
  const t = Math.max(0, player.getCurrentTime() + delta)
  player.seekTo(t, true)
}
function toggleMute(): void { if (!player) return; player.isMuted() ? player.unMute() : player.mute() }
function setRate(r: number): void { player?.setPlaybackRate?.(r) }
function next(): void {
  if (!normalizedIds.value.length) return
  index.value = (index.value + 1) % normalizedIds.value.length
}
function prev(): void {
  if (!normalizedIds.value.length) return
  index.value = (index.value - 1 + normalizedIds.value.length) % normalizedIds.value.length
}

function onKey(e: KeyboardEvent): void {
  // Space: play/pause, ArrowLeft/Right: seek, ,/.: rate down/up
  const active = document.activeElement as HTMLElement | null
  if (active && ['INPUT', 'TEXTAREA'].includes(active.tagName)) return
  if (e.code === 'Space') {
    e.preventDefault();
    const state = (window as any).YT?.PlayerState?.PLAYING
    // Fallback if PlayerState not available yet
    const playing = player?.getPlayerState?.() === (state ?? 1)
    playing ? pause() : play()
  }
  if (e.code === 'ArrowLeft') { seek(-5) }
  if (e.code === 'ArrowRight') { seek(5) }
  if (e.key === ',') { setRate(Math.max(0.25, (player?.getPlaybackRate?.() || 1) - 0.25)) }
  if (e.key === '.') { setRate(Math.min(2, (player?.getPlaybackRate?.() || 1) + 0.25)) }
}

function extractId(input: string): string {
  if (!input) return ''
  try {
    // If raw ID is given, keep it
    if (!/^[a-zA-Z]+:\/\//.test(input) && !input.includes('youtube.com') && !input.includes('youtu.be')) return input
    const u = new URL(input)
    if (u.hostname.includes('youtu.be')) return u.pathname.slice(1)
    if (u.pathname.startsWith('/shorts/')) return u.pathname.split('/')[2]
    const v = u.searchParams.get('v'); if (v) return v
    return input
  } catch {
    return input
  }
}

// Expose methods to parent components
const exposed: ExposedMethods = { play, pause, seek, toggleMute, setRate, next, prev }
defineExpose(exposed)
</script>

<style scoped>
/* Make the iframe "cover" the 9:16 area */
#yt-shorts-vue iframe {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  width: 100vh;
  /* oversize to ensure cover */
  height: 100vw;
  min-width: 100%;
  min-height: 100%;
  pointer-events: auto;
}
</style>