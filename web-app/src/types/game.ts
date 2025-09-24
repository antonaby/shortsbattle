export type GameStage =
  | "lobby"
  | "lobby-full"
  | "submit"
  | "submit-complete"
  | "watch"
  | "watch-complete"
  | "complete";

export type MsgType = "details" | "stage_updated";

export interface GameUpdate {
  id: number;
  msg_type: MsgType;
  stage: GameStage;
  state_change_reason?: string;
  round: number;
  state_changed_at: string;
  remaining_ms: number;
  theme?: Theme;
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

export interface GameVideo {
  id: number;
  game_id: number;
  player_id: number;
  round_n: number;
  submitted_at: string;
  video: Video;
}

export type VoteValue = "like" | "dislike";

export interface LikeDislikeVote {
  value: VoteValue;
}

export interface Vote {
  game_video_id: number;
  player_id: number;
  value: LikeDislikeVote;
  voted_at: string;
}

export interface VideoAuthor {
  tgId: number;
  username: string;
}

export interface VideoResult {
  video: Video;
  author: VideoAuthor;
  likes: number;
  dislikes: number;
}

export interface RoundResult {
  round: Round;
  videos: VideoResult[];
}

export interface PlayerResult {
  tgId: number;
  username: string;
  place: number;
  likes: number;
  dislikes: number;
}

export interface PlayerOutcome {
  plusEnergy: number;
}

export interface GameResult {
  players: PlayerResult[];
  rounds: RoundResult[];
  outcome: PlayerOutcome;
}
