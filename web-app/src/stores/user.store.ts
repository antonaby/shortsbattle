import { defineStore } from "pinia";

interface UserStoreState {
  tgUserId: number | null;
  token: string | null;
}

export const useUserStore = defineStore("auth", {
  state: (): UserStoreState => ({
    tgUserId: 1,
    token: null
  }),
  getters: {
    playerId: (store) => {
      return store.tgUserId
    }
  },
  actions: {
    
  },
});
