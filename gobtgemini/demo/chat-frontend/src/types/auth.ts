export interface User {
  id: string;
  username: string;
  email: string;
  nickname?: string;
  avatar?: string;
  is_active: boolean;
  created_at: string;
  updated_at: string;
  preferences?: UserPreferences;
}

export interface UserPreferences {
  llm_model: string;
  temperature: number;
  max_tokens: number;
  system_prompt?: string;
  context_window: number;
  memory_enabled: boolean;
}

export interface AuthResponse {
  token: string;
  user: User;
  expires_at: string;
}

export interface LoginRequest {
  username_or_email: string;
  password: string;
}

export interface RegisterRequest {
  username: string;
  email: string;
  password: string;
}

export interface AuthContextType {
  user: User | null;
  token: string | null;
  login: (credentials: LoginRequest) => Promise<void>;
  register: (userData: RegisterRequest) => Promise<void>;
  logout: () => void;
  refreshToken: () => Promise<void>;
  isLoading: boolean;
  error: string | null;
}
