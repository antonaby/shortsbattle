export type GameState = "lobby" | "submitting" | "watching" | "completed";

export type MsgType = "details" | "state_change";

export interface GameUpdate {
  id: number;
  msg_type: MsgType;
  state: GameState;
  state_changed_at: string;
  next_state_change_at: string;
  remaning_time_ms: number;
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
