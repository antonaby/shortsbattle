export type GameState =
  | "lobby"
  | "submitting"
  | "wathching"
  | "completed";

export interface GameUpdate {
  id: number;
  theme_id: number;
  state: GameState;
  created_at: string;
  state_changed_at: string;
  next_state_change_at: string;
}

export interface GameTheme {
  id: number;
  name: string;
  description: string;
  created_at: string;
}

export interface GameJoined {
  game_id: number;
  message: string
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
