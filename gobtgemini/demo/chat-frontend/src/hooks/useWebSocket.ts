import { useEffect, useRef } from 'react';
import { wsClient } from '@/lib/websocket';
import { useAuthStore } from '@/store/authStore';
import { useChatStore } from '@/store/chatStore';
import { WebSocketMessage } from '@/types/chat';

export function useWebSocket() {
  const { token, isAuthenticated } = useAuthStore();
  const { addMessage, setConnected, setError } = useChatStore();
  const reconnectTimeoutRef = useRef<NodeJS.Timeout>();

  useEffect(() => {
    if (!isAuthenticated || !token) {
      wsClient.disconnect();
      setConnected(false);
      return;
    }

    // 设置 token
    wsClient.setToken(token);

    // 消息处理器
    const handleMessage = (message: WebSocketMessage) => {
      console.log('收到 WebSocket 消息:', message);
      
      switch (message.type) {
        case 'chat_response':
          if (message.data && message.data.message) {
            addMessage(message.data.message);
          }
          break;
        case 'system':
          console.log('系统消息:', message.content);
          break;
        case 'pong':
          // 心跳响应，不需要特殊处理
          break;
        default:
          console.log('未处理的消息类型:', message.type);
      }
    };

    // 连接处理器
    const handleConnect = () => {
      console.log('WebSocket 连接成功');
      setConnected(true);
      setError(null);
      
      // 清除重连定时器
      if (reconnectTimeoutRef.current) {
        clearTimeout(reconnectTimeoutRef.current);
        reconnectTimeoutRef.current = undefined;
      }
    };

    // 断开处理器
    const handleDisconnect = () => {
      console.log('WebSocket 连接断开');
      setConnected(false);
      
      // 尝试重连
      if (isAuthenticated && token) {
        reconnectTimeoutRef.current = setTimeout(() => {
          console.log('尝试重新连接 WebSocket...');
          wsClient.connect();
        }, 3000);
      }
    };

    // 错误处理器
    const handleError = (error: Event) => {
      console.error('WebSocket 错误:', error);
      setError('WebSocket 连接失败');
      setConnected(false);
    };

    // 注册事件处理器
    const unsubscribeMessage = wsClient.onMessage(handleMessage);
    const unsubscribeConnect = wsClient.onConnect(handleConnect);
    const unsubscribeDisconnect = wsClient.onDisconnect(handleDisconnect);
    const unsubscribeError = wsClient.onError(handleError);

    // 建立连接
    wsClient.connect();

    // 清理函数
    return () => {
      unsubscribeMessage();
      unsubscribeConnect();
      unsubscribeDisconnect();
      unsubscribeError();
      
      if (reconnectTimeoutRef.current) {
        clearTimeout(reconnectTimeoutRef.current);
      }
      
      wsClient.disconnect();
    };
  }, [isAuthenticated, token, addMessage, setConnected, setError]);

  return {
    isConnected: wsClient.isConnected,
    sendMessage: (message: any) => wsClient.send(message),
  };
}
