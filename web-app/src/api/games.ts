import type { GameJoined, Theme } from "../types/game";
import { api } from "./http";

export const GamesAPI = {
  async fetchThemes(): Promise<Theme[]> {
    const response = await api.get<Theme[]>("/api/v1/themes");
    return response.data;
  },
  async joinGame(themeId: number): Promise<GameJoined> {
    const response = await api.put<GameJoined>("/api/v1/games/join", {
      theme_id: themeId,
    });
    return response.data;
  },
};
