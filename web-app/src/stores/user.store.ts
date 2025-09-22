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
      // TODO: get init data from TG
      let initData = "userId=1&username=test&languageCode=ru";

      if (import.meta.env.DEV && import.meta.env.VITE_DEBUG_USER) {
        const params = new URLSearchParams(window.location.search);
        let userId = params.get("userId") || "1";
        let username = params.get("username") || "test";
        let languageCode = params.get("languageCode") || "ru";

        initData = `userId=${userId}&username=${username}&languageCode=${languageCode}`
      }

      let newToken = await AuthAPI.getToken(initData);
      decodedToken = jwtDecode<JwtPayload>(newToken);
      token.value = newToken;
    } catch (error) {
      uiStore.handleAuthError(error);
    }

    return token.value || "";
  }

  function maybeRefreshToken(cooldown = 5): boolean {
    if (!decodedToken || !decodedToken.exp) {
      return true;
    }

    const exp = decodedToken.exp;
    const now = Math.floor(Date.now() / 1000);

    return exp <= now + cooldown;
  }

  return { token, getToken };
});
