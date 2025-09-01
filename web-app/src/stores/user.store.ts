import { defineStore } from "pinia";
import type { AuthResult } from "../types/common";
import axios from "axios";
import { ref } from "vue";
import { NetworkError } from "../types/errors";

export const useUserStore = defineStore("auth", () => {
  const token = ref<string | null>(null);

  // TODO: handle http statuses
  async function getToken(): Promise<string> {
    if (token.value) {
      return token.value;
    }

    try {
      const response = await axios.post<AuthResult>(
        `${import.meta.env.VITE_BASE_URL}/api/v1/auth`,
        {
          "init_data": "userId=1" // TODO: get user from TG
        }
      );

      if (response.data.token) {
        token.value = response.data.token;
        return token.value;
      }

      if (response.data.error) {
        throw new NetworkError("authorization failed");
      }
    } catch (error) {
      // TODO: handle with UI Store
      throw new NetworkError("authorization failed", error);
    }

    return "";
  }

  return { token, getToken };
});
