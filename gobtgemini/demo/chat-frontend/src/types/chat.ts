export interface ChatMessage {
  id: string;
  user_id: string;
  content: string;
  type: MessageType;
  timestamp: string;
  conversation_id: string;
  metadata?: any;
}

export type MessageType = 'user' | 'assistant' | 'system';

export interface SendMessageRequest {
  conversation_id:any
  content: string;
}

export interface SendMessageResponse {
  user_message: ChatMessage;
  assistant_message: ChatMessage;
  processing_time: string;
}

export interface WebSocketMessage {
  type: 'system' | 'pong' | 'chat_response' | 'ping';
  content: string;
  user_id?: string;
  username?: string;
  timestamp: string;
  data?: any;
}

export interface ChatContextType {
  messages: ChatMessage[];
  sendMessage: (content: string) => Promise<void>;
  loadHistory: () => Promise<void>;
  clearHistory: () => Promise<void>;
  deleteMessage: (messageId: string) => Promise<void>;
  isLoading: boolean;
  isSending: boolean;
  error: string | null;
  isConnected: boolean;
}
