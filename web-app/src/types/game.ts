export type GameState = "lobby" | "lobby-full" | "submit" |  "submit-complete" | "watch" | "watch-complete" | "complete";

export type MsgType = "details" | "stage_updated";

export interface GameUpdate {
  id: number;
  msg_type: MsgType;
  stage: GameState;
  state_change_reason?: string;
  round: number;
  state_changed_at: string;
  remaining_ms: number;
  theme?: Theme;
  result?: any;
}

export type PlayerMode = "submit_and_vote" | "only_vote";

export type GameMode = "likedislike";

export interface Round {
  round_n: number;
  title: string;
  description: string;
}

export interface Theme {
  id: number;
  title: string;
  description: string;
  mode: GameMode;
  rounds?: Round[];
}

export interface GameJoined {
  game_id: number;
  message: string;
}

export interface Video {
  id: number;
  video_url: string;
  oembed: OEmbed;
  added_at: string;
  updated_at: string;
}

export interface OEmbed {
  html?: string;
  type?: string;
  title?: string;
  width?: number | string;
  height?: number | string;
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





// TODO: review
export interface GameResult {
  videos: VideoResult[];
}

export interface VideoResult {
  id: number;
  likes: number;
  dislikes: number;
}




export type VoteValue = "like" | "dislike" | "skip";

export interface Vote {
  game_id: number;
  player_id: number;
  video_id: number;
  voted_at: string;
  value: VoteValue;
}