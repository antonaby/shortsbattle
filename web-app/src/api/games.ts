import type { GameJoined, PlayerMode, Theme, Video, Vote, VoteValue } from "../types/game";
import { api } from "./http";

export const GamesAPI = {
  async fetchThemes(): Promise<Theme[]> {
    const response = await api.get<Theme[]>("/api/v1/themes");
    return response.data;
  },
  async fetchTheme(themeId: number): Promise<Theme> {
    const response = await api.get<Theme>(`/api/v1/themes/${themeId}`);
    return response.data;
  },
  async joinGame(themeId: number, mode: PlayerMode): Promise<GameJoined> {
    const response = await api.put<GameJoined>("/api/v1/games/join", {
      theme_id: themeId,
      mode: mode
    });
    return response.data;
  },
  async getMyVideos(): Promise<Video[]> {
    const response = await api.get<Video[]>("/api/v1/me/videos");
    return response.data;
  },
  async submitExistingVideo(
    gameId: number,
    videoId: number,
    requestId: number
  ): Promise<Video> {
    const response = await api.put<Video>(`/api/v1/games/${gameId}/submit`, {
      video_id: videoId,
      video_request_id: requestId,
    });
    return response.data;
  },
  async submitNewVideo(
    gameId: number,
    videoUrl: string,
    requestId: number
  ): Promise<Video> {
    const response = await api.put<Video>(`/api/v1/games/${gameId}/submit`, {
      video_url: videoUrl,
      video_request_id: requestId,
    });
    return response.data;
  },
  async fetchVideosToWatch(gameId: number): Promise<Video[]> {
    const response = await api.get<Video[]>(
      `${import.meta.env.VITE_BASE_URL}/api/v1/games/${gameId}/videos`
    );
    return response.data;
  },
  async voteForVideo(
    gameId: number,
    videoId: number,
    value: VoteValue
  ): Promise<Vote> {
    const response = await api.put<Vote>(
      `${import.meta.env.VITE_BASE_URL}/api/v1/games/${gameId}/vote`,
      {
        video_id: videoId,
        value: value,
      }
    );
    return response.data;
  },
};
