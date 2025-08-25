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
