import { defineStore } from "pinia";
import { computed, ref } from "vue";
import { useRouter } from "vue-router";

export const useUIStore = defineStore("ui", () => {
  const wsConnected = ref<boolean>(false);
  const isReady = computed(() => wsConnected.value);
  const router = useRouter();

  function setWsConnected(value: boolean) {
    wsConnected.value = value;
  }

  function openGame(gameId: number) {
    router.push({ name: "game", params: { id: gameId } });
  }

  return { isReady, setWsConnected, openGame };
});
