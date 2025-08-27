import { defineStore } from "pinia";
import { computed, ref } from "vue";

export const useUIStore = defineStore("ui", () => {
  const wsConnected = ref<boolean>(false);

  const isReady = computed(() => wsConnected.value);

  function setWsConnected(value: boolean) {
    wsConnected.value = value;
  }

  return { isReady, setWsConnected };
});
