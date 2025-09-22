import type {
  GameJoined,
  GameVideo,
  PlayerMode,
  Theme,
  Video,
  Vote,
  VoteValue,
} from "../types/game";
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
      mode: mode,
    });
    return response.data;
  },
  async getPayerVideos(query: string): Promise<Video[]> {
    const response = await api.get<Video[]>("/api/v1/me/videos", {
      params: {
        query: query,
      },
    });
    return response.data;
  },
  async submitExistingVideo(
    gameId: number,
    videoId: number,
    round: number
  ): Promise<GameVideo> {
    const response = await api.put<GameVideo>(
      `/api/v1/games/${gameId}/submit`,
      {
        video_id: videoId,
      },
      {
        params: {
          round: round,
        },
      }
    );
    return response.data;
  },
  async submitNewVideo(
    gameId: number,
    videoUrl: string,
    round: number
  ): Promise<GameVideo> {
    const response = await api.put<GameVideo>(
      `/api/v1/games/${gameId}/submit`,
      {
        video_url: videoUrl,
      },
      {
        params: {
          round: round,
        },
      }
    );
    return response.data;
  },
  async fetchVideosToWatch(gameId: number, round: number): Promise<GameVideo[]> {
    const response = await api.get<GameVideo[]>(`/api/v1/games/${gameId}/videos`, {
      params: {
        round,
      },
    });
    return response.data;
  },

  // TODO: review
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
