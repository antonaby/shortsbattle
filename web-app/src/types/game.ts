export type GameState = "lobby" | "submitting" | "watching" | "completed";

export type MsgType = "details" | "state_change";

export interface ThemeDetails {
  id: number;
  name: string;
  description: string;
}

export interface GameUpdate {
  id: number;
  msg_type: MsgType;
  state: GameState;
  state_changed_at: string;
  next_state_change_at: string;
  remaning_time_ms: number;
  theme?: ThemeDetails;
}

export interface Video {
  id: number;
  player_id: number;
  video_url: string;
  oembed: OEmbed;
  added_at: string;
}

export interface OEmbed {
  html?: string;
  type?: string;
  title?: string;
  width?: number;
  height?: number;
  version?: string;
  author_url?: string;
  author_name?: string;
  provider_url?: string;
  provider_name?: string;
  thumbnail_url?: string;
  thumbnail_width?: number;
  thumbnail_height?: number;

  [key: string]: unknown;
}

export type VoteValue = "like" | "dislike" | "skip";

export interface Vote {
  game_id: number;
  player_id: number;
  video_id: number;
  voted_at: string;
  value: VoteValue;
}
export interface ThemeVideoRequest {
  id: number;
  request: string;
  theme_id: number;
  created_at: string;
}

export interface Theme {
  id: number;
  name: string;
  description: string;
  created_at: string;
  requests: ThemeVideoRequest[];
}

export interface GameJoined {
  game_id: number;
  message: string;
}

export interface OkResponse {
  message: string;
}
