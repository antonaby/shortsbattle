import { defineStore } from "pinia";
import { ref } from "vue";
import { useUIStore } from "./ui.store";
import { AuthAPI } from "../api/auth";
import { jwtDecode, type JwtPayload } from "jwt-decode";

export const useUserStore = defineStore("auth", () => {
  const token = ref<string | null>(null);
  const uiStore = useUIStore();

  let decodedToken: JwtPayload | null = null;

  async function getToken(): Promise<string> {
    if (token.value && !maybeRefreshToken()) {
      return token.value;
    }

    try {
      let newToken = await AuthAPI.getToken("userId=1"); // TODO: get user from TG
      decodedToken = jwtDecode<JwtPayload>(newToken);
      token.value = newToken;
    } catch (error) {
      uiStore.handleNetworkError(error);
    }

    return token.value || "";
  }

  function maybeRefreshToken(cooldown = 5): boolean {
    if (!decodedToken || !decodedToken.exp) {
      return true;
    }

    const exp = decodedToken.exp;
    const now = Math.floor(Date.now() / 1000);

    return exp <= now + cooldown
  }

  return { token, getToken };
});
