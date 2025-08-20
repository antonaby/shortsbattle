export type GameStatus =
  | "created"
  | "lobby"
  | "submitting"
  | "voting"
  | "winner"
  | "complete";

export interface GameUpdate {
  id: number;
  status: GameStatus;
  stage_time_remaining: number;
}

export interface GameTheme {
  id: number;
  name: string;
  description: string;
  created_at: string;
}

export interface Game {
  id: number;
  theme_id: number;
  status: GameStatus;
  created_at: string;
}

export interface Player {
  id: number;
  username: string;
  created_at: string;
}

export interface Video {
  id: number;
  game_id: number;
  player_id: number;
  video_url: string;
  is_actual: boolean;
  submitted_at: string;
}

export interface Vote {
  id: number;
  game_id: number;
  video_id: number;
  voter_id: number;
  value: string;
  is_actual: boolean;
  voted_at: string;
}

export interface GameInstance {
  id: number;
  theme: GameTheme;
  status: GameStatus;
  created_at: string;
  stage_time_remaining: number;
  players: Player[];
  videos: Video[];
  votes: Vote[];
}
