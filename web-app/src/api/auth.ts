import type { AuthResult } from "../types/common";
import { NetworkError } from "../types/errors";
import { api } from "./http";

export const AuthAPI = {
  async getToken(): Promise<string> {
    const response = await api.post<AuthResult>(
      `${import.meta.env.VITE_BASE_URL}/api/v1/auth`,
      {
        init_data: "userId=1", // TODO: get user from TG
      }
    );

    if (response.status != 200) {
      throw new NetworkError(
        `request failed, status: ${response.status}, response: ${JSON.stringify(
          response.data
        )}`,
        response.status
      );
    }

    if (response.data.token) {
      return response.data.token;
    }

    throw new NetworkError("no token in the response", 401);
  },
};
