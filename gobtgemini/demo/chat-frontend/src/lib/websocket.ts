import { WebSocketMessage } from '@/types/chat';

type MessageHandler = (message: WebSocketMessage) => void;
type ConnectionHandler = () => void;
type ErrorHandler = (error: Event) => void;

export class WebSocketClient {
  private ws: WebSocket | null = null;
  private token: string | null = null;
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 5;
  private reconnectDelay = 1000;
  private isConnecting = false;
  private messageHandlers: MessageHandler[] = [];
  private connectionHandlers: ConnectionHandler[] = [];
  private disconnectionHandlers: ConnectionHandler[] = [];
  private errorHandlers: ErrorHandler[] = [];
  private heartbeatInterval: NodeJS.Timeout | null = null;

  constructor(private baseUrl: string) {} // 修改：这里只接收基础URL

  setToken(token: string | null) {
    this.token = token;
  }

  connect() {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      console.log('WebSocket 已经连接.');
      return;
    }

    if (this.isConnecting) {
      console.log('WebSocket 正在连接中...');
      return;
    }

    // ⭐ 修改点 1：检查 Token 是否存在
    if (!this.token) {
      console.error('无法连接 WebSocket：Token 未设置.');
      return;
    }

    this.isConnecting = true;

    try {
      // ⭐ 修改点 2：在连接时动态构建包含 Token 的 URL
      const authenticatedUrl = new URL(this.baseUrl);
      authenticatedUrl.searchParams.append('token', this.token);
      
      console.log('尝试连接 WebSocket:', authenticatedUrl.toString());
      this.ws = new WebSocket(authenticatedUrl.toString());

      this.ws.onopen = () => {
        console.log('WebSocket 连接已建立');
        this.isConnecting = false;
        this.reconnectAttempts = 0;
        this.startHeartbeat();
        this.connectionHandlers.forEach(handler => handler());
      };

      this.ws.onmessage = (event) => {
        try {
          const message: WebSocketMessage = JSON.parse(event.data);
          this.messageHandlers.forEach(handler => handler(message));
        } catch (error) {
          console.error('解析 WebSocket 消息失败:', error);
        }
      };

      this.ws.onclose = () => {
        console.log('WebSocket 连接已关闭');
        this.isConnecting = false;
        this.stopHeartbeat();
        this.disconnectionHandlers.forEach(handler => handler());
        this.handleReconnect();
      };

      this.ws.onerror = (error) => {
        console.error('WebSocket 连接错误:', error);
        this.isConnecting = false;
        this.errorHandlers.forEach(handler => handler(error));
      };
    } catch (error) {
      console.error('创建 WebSocket 连接失败:', error);
      this.isConnecting = false;
    }
  }

  disconnect() {
    // 修改：重置重连尝试次数，避免立即重连
    this.reconnectAttempts = this.maxReconnectAttempts;
    this.stopHeartbeat();
    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
    console.log('WebSocket 已手动断开');
  }

  send(message: any) {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(message));
    } else {
      console.warn('WebSocket 未连接，无法发送消息');
    }
  }

  private startHeartbeat() {
    this.heartbeatInterval = setInterval(() => {
      this.send({ type: 'ping', content: 'ping' });
    }, 30000); // 每30秒发送一次心跳
  }

  private stopHeartbeat() {
    if (this.heartbeatInterval) {
      clearInterval(this.heartbeatInterval);
      this.heartbeatInterval = null;
    }
  }

  private handleReconnect() {
    if (this.reconnectAttempts < this.maxReconnectAttempts) {
      this.reconnectAttempts++;
      console.log(`尝试重连 WebSocket (${this.reconnectAttempts}/${this.maxReconnectAttempts})`);
      setTimeout(() => {
        this.connect();
      }, this.reconnectDelay * this.reconnectAttempts);
    } else {
      console.error('WebSocket 重连失败，已达到最大重试次数');
    }
  }

  onMessage(handler: MessageHandler) {
    this.messageHandlers.push(handler);
    return () => {
      this.messageHandlers = this.messageHandlers.filter(h => h !== handler);
    };
  }

  onConnect(handler: ConnectionHandler) {
    this.connectionHandlers.push(handler);
    return () => {
      this.connectionHandlers = this.connectionHandlers.filter(h => h !== handler);
    };
  }

  onDisconnect(handler: ConnectionHandler) {
    this.disconnectionHandlers.push(handler);
    return () => {
      this.disconnectionHandlers = this.disconnectionHandlers.filter(h => h !== handler);
    };
  }

  onError(handler: ErrorHandler) {
    this.errorHandlers.push(handler);
    return () => {
      this.errorHandlers = this.errorHandlers.filter(h => h !== handler);
    };
  }

  get isConnected() {
    return this.ws?.readyState === WebSocket.OPEN;
  }
}

// ⭐ 修改点 3：仍然创建并导出一个单例，但不在此时连接
const WS_URL = import.meta.env.VITE_WS_URL || 'ws://192.168.31.205:8080/api/v1/ws/chat';
export const wsClient = new WebSocketClient(WS_URL);