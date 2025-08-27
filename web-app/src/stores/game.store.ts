import { defineStore } from "pinia";
import type { GameUpdate } from "../types/game";
import { computed, ref } from "vue";
import { useWSStore } from "./ws.store";

export const useGameStore = defineStore("game", () => {
  const wsStore = useWSStore();

  const lastGameUpdate = ref<GameUpdate | null>(null);
  const remainingTimeMs = ref<number>(0);

  let channelId: string | null = null;
  let intervalId: number | null = null;

  const formattedTime = computed(() => {
    const totalSeconds = Math.max(0, Math.floor(remainingTimeMs.value / 1000));
    const minutes = Math.floor(totalSeconds / 60);
    const seconds = totalSeconds % 60;
    return `${minutes}:${seconds.toString().padStart(2, "0")}`;
  });

  function updateRemainingTime(upd: GameUpdate) {
    const start = new Date(upd.state_changed_at).getTime();
    const end = new Date(upd.next_state_change_at).getTime();

    remainingTimeMs.value = end - start;
  }

  function joinGame(gameId: number) {
    channelId = `game_${gameId}`;

    wsStore.subscribe(channelId, {
      subscribed: (ctx) => {
        if (ctx.data) {
          var upd: GameUpdate = ctx.data;

          lastGameUpdate.value = upd;
          updateRemainingTime(upd);

          if (intervalId) {
            clearInterval(intervalId);
          }

          intervalId = setInterval(() => {
            remainingTimeMs.value -= 1000;
          }, 1000);
        } else {
          console.log("no subscription data");
        }
      },
      publication: (ctx) => {
        var upd: GameUpdate = ctx.data;
        if (lastGameUpdate.value) {
          if (lastGameUpdate.value.state !== upd.state) {
            updateRemainingTime(upd);
          }
          lastGameUpdate.value = upd;
        }
      },
    });
  }

  function leaveGame() {
    if (intervalId) {
      clearInterval(intervalId);
      intervalId = null;
    }

    if (channelId) {
      wsStore.unsubscribe(channelId);
    }

    lastGameUpdate.value = null;
  }

  return { lastGameUpdate, formattedTime, joinGame, leaveGame };
});
