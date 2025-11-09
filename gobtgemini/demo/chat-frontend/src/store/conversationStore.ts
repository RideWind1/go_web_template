import { create } from 'zustand';

export interface Conversation {
  id: string;
  title: string;
  created_at: string;
  updated_at: string;
  message_count?: number;
  last_message?: string;
}

interface ConversationState {
  conversations: Conversation[];
  currentConversationId: string | null;
  isLoading: boolean;
  error: string | null;
}

interface ConversationActions {
  setConversations: (conversations: Conversation[]) => void;
  addConversation: (conversation: Conversation) => void;
  removeConversation: (conversationId: string) => void;
  setCurrentConversation: (conversationId: string | null) => void;
  updateConversation: (conversationId: string, updates: Partial<Conversation>) => void;
  setLoading: (loading: boolean) => void;
  setError: (error: string | null) => void;
  clearError: () => void;
}

export const useConversationStore = create<ConversationState & ConversationActions>()((set, get) => ({
  // State
  conversations: [],
  currentConversationId: null,
  isLoading: false,
  error: null,

  // Actions
  setConversations: (conversations) => {
    set({ conversations });
  },

  addConversation: (conversation) => {
    set((state) => ({
      conversations: [conversation, ...state.conversations],
    }));
  },

  removeConversation: (conversationId) => {
    set((state) => ({
      conversations: state.conversations.filter(conv => conv.id !== conversationId),
      currentConversationId: state.currentConversationId === conversationId ? null : state.currentConversationId,
    }));
  },

  setCurrentConversation: (conversationId) => {
    set({ currentConversationId: conversationId });
  },

  updateConversation: (conversationId, updates) => {
    set((state) => ({
      conversations: state.conversations.map((conv) =>
        conv.id === conversationId ? { ...conv, ...updates } : conv
      ),
    }));
  },

  setLoading: (loading) => {
    set({ isLoading: loading });
  },

  setError: (error) => {
    set({ error, isLoading: false });
  },

  clearError: () => {
    set({ error: null });
  },
}));
