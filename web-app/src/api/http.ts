import axios from "axios";
import { useUserStore } from "../stores/user.store";

export const api = axios.create({
  baseURL: import.meta.env.VITE_BASE_URL,
});

api.interceptors.request.use(async (config) => {
  const auth = useUserStore();

  if (auth.token) {
    config.headers = config.headers || {};
    config.headers.Authorization = `Bearer ${auth.token}`
  }

  return config;
});
