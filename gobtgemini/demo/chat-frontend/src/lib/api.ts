import { ApiResponse, ApiError } from '@/types/api';
import { AuthResponse, LoginRequest, RegisterRequest, User } from '@/types/auth';
import { ChatMessage, SendMessageRequest, SendMessageResponse } from '@/types/chat';

const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://192.168.31.205:8080/api/v1';

class ApiClient {
  private token: string | null = null;

  constructor() {
    this.token = localStorage.getItem('auth_token');
  }

  setToken(token: string | null) {
    this.token = token;
    if (token) {
      localStorage.setItem('auth_token', token);
    } else {
      localStorage.removeItem('auth_token');
    }
  }

  private async request<T>(
    endpoint: string,
    options: RequestInit = {}
  ): Promise<T> {
    const url = `${API_BASE_URL}${endpoint}`;
    const headers: Record<string, string> = {
      'Content-Type': 'application/json',
      ...(options.headers as Record<string, string>),
    };

    if (this.token) {
      headers.Authorization = `Bearer ${this.token}`;
    }

    try {
      const response = await fetch(url, {
        ...options,
        headers,
      });

      const data = await response.json();

      if (!response.ok) {
        const error: ApiError = data;
        throw new Error(error.error || error.message || `HTTP ${response.status}`);
      }

      return data as T;
    } catch (error) {
      if (error instanceof Error) {
        throw error;
      }
      throw new Error('网络请求失败');
    }
  }

  // 认证相关
  async login(credentials: LoginRequest): Promise<AuthResponse> {
    const response = await this.request<ApiResponse<AuthResponse>>('/auth/login', {
      method: 'POST',
      body: JSON.stringify(credentials),
    });
    return response.data!;
  }

  async register(userData: RegisterRequest): Promise<AuthResponse> {
    const response = await this.request<ApiResponse<AuthResponse>>('/auth/register', {
      method: 'POST',
      body: JSON.stringify(userData),
    });
    return response.data!;
  }

  async refreshToken(): Promise<{ token: string; expires_at: string }> {
    const response = await this.request<ApiResponse<{ token: string; expires_at: string }>>('/auth/refresh', {
      method: 'POST',
    });
    return response.data!;
  }

  async getProfile(): Promise<User> {
    const response = await this.request<ApiResponse<User>>('/user/profile');
    return response.data!;
  }

  async updateProfile(updates: Partial<User>): Promise<void> {
    await this.request('/user/profile', {
      method: 'PUT',
      body: JSON.stringify(updates),
    });
  }

  // 聊天相关
  async sendMessage(messageData: SendMessageRequest): Promise<SendMessageResponse> {
    const response = await this.request<ApiResponse<SendMessageResponse>>('/chat/send', {
      method: 'POST',
      body: JSON.stringify(messageData),
    });
    return response.data!;
  }

  async getChatHistory(conversationId:string, limit = 50, offset = 0): Promise<{
    messages: ChatMessage[];
    limit: number;
    offset: number;
    count: number;
  }> {
    const response = await this.request<ApiResponse<{
      messages: ChatMessage[];
      limit: number;
      offset: number;
      count: number;
    }>>(`/chat/history?limit=${limit}&offset=${offset}&conversationId=${conversationId}`);
    return response.data!;
  }

  async deleteMessage(messageId: string): Promise<void> {
    await this.request(`/chat/history/${messageId}`, {
      method: 'DELETE',
    });
  }

  async clearHistory(): Promise<void> {
    await this.request('/chat/clear', {
      method: 'POST',
    });
  }

  // 对话管理相关
  async getConversations(): Promise<any[]> {
    const response = await this.request<ApiResponse<any[]>>('/chat/conversations');
    return response.data || [];
  }

  async createConversation(title?: string): Promise<any> {
    const response = await this.request<ApiResponse<any>>('/chat/conversations', {
      method: 'POST',
      body: JSON.stringify({ title: title || '新的对话' }),
    });
    return response.data!;
  }

  async deleteConversation(conversationId: string): Promise<void> {
    await this.request(`/chat/conversations/${conversationId}`, {
      method: 'DELETE',
    });
  }

  async getConversationHistory(conversationId: string, limit = 50, offset = 0): Promise<{
    messages: ChatMessage[];
    limit: number;
    offset: number;
    count: number;
  }> {
    const response = await this.request<ApiResponse<{
      messages: ChatMessage[];
      limit: number;
      offset: number;
      count: number;
    }>>(`/chat/conversations/${conversationId}/history?limit=${limit}&offset=${offset}`);
    return response.data!;
  }
}

export const apiClient = new ApiClient();
