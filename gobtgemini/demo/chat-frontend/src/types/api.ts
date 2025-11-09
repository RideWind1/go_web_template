import { ChatMessage } from './chat';

export interface ApiResponse<T = any> {
  data?: T;
  message?: string;
  code?: string;
}

export interface ApiError {
  error: string;
  code?: string;
  message?: string;
}

export interface PaginationParams {
  limit?: number;
  offset?: number;
}

export interface ChatHistoryResponse {
  messages: ChatMessage[];
  limit: number;
  offset: number;
  count: number;
}

export type Theme = 'light' | 'dark';

export interface AppSettings {
  theme: Theme;
  language: string;
  notifications: boolean;
}
