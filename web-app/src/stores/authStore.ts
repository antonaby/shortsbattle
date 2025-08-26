import { defineStore } from "pinia";
import axios from "axios";

interface AuthStoreState {
  token: string | null;
}

export const useAuthStore = defineStore("auth", {
  state: (): AuthStoreState => ({
    token: null
  }),
  actions: {
    async getToken() {
      
    }
  },
});
