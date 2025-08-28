export type GameState = "lobby" | "submitting" | "watching" | "completed";

export interface GameUpdate {
  id: number;
  theme_id: number;
  state: GameState;
  created_at: string;
  state_changed_at: string;
  next_state_change_at: string;
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
