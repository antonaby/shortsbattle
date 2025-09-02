import type { GameJoined, Theme } from "../types/game";
import { api, checkStatusCode } from "./http";

export const GamesAPI = {
  async fetchThemes(): Promise<Theme[]> {
    const response = await api.get<Theme[]>("/api/v1/themes");
    checkStatusCode(response, 200);
    return response.data;
  },
  async joinGame(themeId: number): Promise<GameJoined> {
    const response = await api.put<GameJoined>("/api/v1/games/join", {
      theme_id: themeId,
    });
    checkStatusCode(response, 200);
    return response.data;
  },
};
