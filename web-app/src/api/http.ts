import axios, { AxiosError } from "axios";
import { useUserStore } from "../stores/user.store";
import { NetworkError } from "../types/errors";

export const api = axios.create({
  baseURL: import.meta.env.VITE_BASE_URL,
});

api.interceptors.request.use(async (config) => {
  if (config.url?.includes("/api/v1/auth")) {
    return config;
  }

  const auth = useUserStore();

  const token = await auth.getToken();
  if (token) {
    config.headers = config.headers || {};
    config.headers.Authorization = `Bearer ${token}`;
  }

  return config;
});

api.interceptors.response.use(
  (res) => res,
  (error: AxiosError) => {
    let data = error.response?.data || {};
    let status = error.response?.status || 0;

    let wrapper = new NetworkError(
      `request failed, status: ${status}, response: ${JSON.stringify(data)}`,
      status,
      error
    );

    return Promise.reject(wrapper);
  }
);
