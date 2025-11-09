import { create } from 'zustand';
import { ChatMessage } from '@/types/chat';

interface ChatState {
  messages: ChatMessage[];
  currentConversationId: string | null;
  isTyping: boolean;
  isConnected: boolean;
  error: string | null;
}

interface ChatActions {
  addMessage: (message: ChatMessage) => void;
  setMessages: (messages: ChatMessage[]) => void;
  clearMessages: () => void;
  setCurrentConversationId: (id: string | null) => void;
  setTyping: (typing: boolean) => void;
  setConnected: (connected: boolean) => void;
  setError: (error: string | null) => void;
  updateMessage: (messageId: string, updates: Partial<ChatMessage>) => void;
  removeMessage: (messageId: string) => void;
  clearError: () => void;
}

export const useChatStore = create<ChatState & ChatActions>()((set, get) => ({
  // State
  messages: [],
  currentConversationId: null,
  isTyping: false,
  isConnected: false,
  error: null,

  // Actions
  addMessage: (message) => {
    set((state) => ({
      messages: [...state.messages, message],
    }));
  },

  setMessages: (messages) => {
    set({ messages });
  },

  clearMessages: () => {
    set({ messages: [] });
  },

  setCurrentConversationId: (id) => {
    set({ currentConversationId: id });
  },

  setTyping: (typing) => {
    set({ isTyping: typing });
  },

  setConnected: (connected) => {
    set({ isConnected: connected });
  },

  setError: (error) => {
    set({ error });
  },

  updateMessage: (messageId, updates) => {
    set((state) => ({
      messages: state.messages.map((message) =>
        message.id === messageId ? { ...message, ...updates } : message
      ),
    }));
  },

  removeMessage: (messageId) => {
    set((state) => ({
      messages: state.messages.filter((message) => message.id !== messageId),
    }));
  },

  clearError: () => {
    set({ error: null });
  },
}));
