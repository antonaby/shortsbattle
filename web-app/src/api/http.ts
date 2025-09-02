import axios, { type AxiosResponse } from "axios";
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

export function checkStatusCode(
  response: AxiosResponse,
  expectedStatusCode: number
) {
  if (response.status != expectedStatusCode) {
    throw new NetworkError(
      `request failed, status: ${response.status}, response: ${JSON.stringify(
        response.data
      )}`,
      response.status
    );
  }
}
