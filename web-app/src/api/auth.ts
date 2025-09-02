import type { AuthResult } from "../types/common";
import { NetworkError } from "../types/errors";
import { api } from "./http";

export const AuthAPI = {
  async getToken(initData: string): Promise<string> {
    const response = await api.post<AuthResult>(
      '/api/v1/auth',
      {
        init_data: initData, 
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
