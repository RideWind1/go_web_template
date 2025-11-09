import { useState, useEffect, useRef } from 'react';
import { useSearchParams } from 'react-router-dom';
import { Send, Loader2, Bot, User } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { ScrollArea } from '@/components/ui/scroll-area';
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';
import { Card } from '@/components/ui/card';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { useChatStore } from '@/store/chatStore';
import { useAuthStore } from '@/store/authStore';
import { useConversationStore } from '@/store/conversationStore';
import { useWebSocket } from '@/hooks/useWebSocket';
import { ChatMessage, MessageType } from '@/types/chat';
import { cn } from '@/lib/utils';
import { formatTime } from '@/lib/utils';
import { apiClient } from '@/lib/api';

export function ChatPage() {
  const [searchParams] = useSearchParams();
  const [message, setMessage] = useState('');
  const [isSending, setIsSending] = useState(false);
  const messagesEndRef = useRef<HTMLDivElement>(null);
  const inputRef = useRef<HTMLInputElement>(null);
  
  const { user } = useAuthStore();
  const { messages, addMessage, setMessages, clearMessages, isTyping, error, setError } = useChatStore();
  const { currentConversationId, setCurrentConversation } = useConversationStore();
  const { isConnected } = useWebSocket();

  // 从 URL获取对话 ID
  const conversationIdFromUrl = searchParams.get('id');

  // 初始化对话
  useEffect(() => {
    if (conversationIdFromUrl && conversationIdFromUrl !== currentConversationId) {
      console.log("1111111");
      setCurrentConversation(conversationIdFromUrl);
      loadConversationHistory(conversationIdFromUrl);
    } else if (!conversationIdFromUrl && currentConversationId) {
      console.log("222222");
      // 如果没有URL参数但有当前对话，加载当前对话的历史
      loadConversationHistory(currentConversationId);
    } else if (!conversationIdFromUrl && !currentConversationId) {
      console.log("33333333");
      // 没有对话时，显示欢迎消息
      clearMessages();
      const welcomeMessage: ChatMessage = {
        id: 'welcome-' + Date.now(),
        content: '你好！欢迎使用智能聊天助手。我可以帮助您解答问题、提供建议或进行对话。请问有什么可以帮助您的吗？',
        type: 'assistant',
        timestamp: new Date().toISOString(),
        user_id: 'assistant',
        conversation_id: 'welcome',
      };
      setMessages([welcomeMessage]);
    }
  }, [conversationIdFromUrl, currentConversationId, setCurrentConversation, setMessages, clearMessages]);

  // 加载对话历史
  const loadConversationHistory = async (conversationId: string) => {
    try {
      console.log("444444");
      const history = await apiClient.getConversationHistory(conversationId);
      setMessages(history.messages);
    } catch (error) {
      console.error('加载对话历史失败:', error);
      setError('加载对话历史失败');
    }
  };

  // 自动滚动到底部
  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  };

  useEffect(() => {
    scrollToBottom();
  }, [messages]);

  // 发送消息
  const handleSendMessage = async () => {
    if (!message.trim() || isSending || !user) return;

    const messageContent = message.trim();
    let activeConversationId = currentConversationId; // 使用 let 以便可以修改

    setMessage('');
    setIsSending(true);
    setError(null);

    try {
      // 如果没有当前对话，先创建一个
      if (!activeConversationId) {
        console.log("当前没有对话ID，正在创建新的对话...");
        const newConversation = await apiClient.createConversation();
        activeConversationId = newConversation.id;
        setCurrentConversation(activeConversationId);
        // 更新 URL
        window.history.pushState({}, '', `/chat?id=${activeConversationId}`);
        // 清空当前消息列表，准备显示新对话的消息
        clearMessages();
      }

      // 创建用户消息对象
      const userMessage: ChatMessage = {
        id: Date.now().toString(),
        content: messageContent,
        type: 'user',
        timestamp: new Date().toISOString(),
        user_id: user.id,
        conversation_id: activeConversationId!, // 使用 ! 断言，因为我们已确保它有值
      };

      // 立即在界面上显示用户发送的消息
      addMessage(userMessage);

      console.log("即将发送消息到对话ID:", activeConversationId);
      
      // 发送到后端
      const result1 = await apiClient.sendMessage({
        conversation_id: activeConversationId!,
        content: messageContent,
      });

      if (result1 && result1.assistant_message) {
        addMessage(result1.assistant_message);
      }

    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : '发送消息失败';
      console.error('发送消息失败:', error);
      setError(errorMessage);

      // ⭐ 关键修复：在这里添加错误处理和自我修复逻辑 ⭐
      // 如果后端返回 500 错误，我们有理由怀疑是 conversation_id 过期了
      // 这里的判断可以更精确，例如检查 errorMessage.includes("外键") 等
      if (activeConversationId) {
          console.log(`发送失败，可能对话ID ${activeConversationId} 已过期，正在重置状态...`);
          // 清空无效的对话ID
          setCurrentConversation(null);
          // 清理URL，避免刷新后再次加载无效ID
          window.history.pushState({}, '', '/chat');
          // 提示用户重试
          setError('发送失败，对话可能已失效，请重试一次。');
      }

    } finally {
      setIsSending(false);
      // 将焦点重新设置到输入框
      inputRef.current?.focus();
    }
  };

  const handleKeyPress = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      handleSendMessage();
    }
  };

  const renderMessage = (msg: ChatMessage) => {
    const isUser = msg.type === 'user';
    const avatar = isUser ? user?.avatar : undefined;
    const username = isUser ? user?.username : 'AI 助手';
    const initials = isUser ? user?.username.charAt(0).toUpperCase() : 'AI';

    return (
      <div key={msg.id} className={cn("flex gap-3 mb-4", isUser && "flex-row-reverse")}>
        <Avatar className="h-8 w-8 flex-shrink-0">
          <AvatarImage src={avatar} alt={username} />
          <AvatarFallback>
            {isUser ? initials : <Bot className="h-4 w-4" />}
          </AvatarFallback>
        </Avatar>
        
        <div className={cn("flex-1 space-y-1", isUser && "text-right")}>
          <div className="flex items-center gap-2">
            {!isUser && <span className="text-sm font-medium">{username}</span>}
            <span className="text-xs text-muted-foreground">
              {formatTime(msg.timestamp)}
            </span>
            {isUser && <span className="text-sm font-medium">{username}</span>}
          </div>
          
          <Card className={cn(
            "p-3 max-w-[80%] inline-block",
            isUser 
              ? "bg-primary text-primary-foreground ml-auto" 
              : "bg-muted"
          )}>
            <p className="text-sm whitespace-pre-wrap break-words">
              {msg.content}
            </p>
          </Card>
        </div>
      </div>
    );
  };

  return (
    <div className="flex flex-col h-full">
      {/* 聊天消息区域 */}
      <ScrollArea className="flex-1 p-4">
        <div className="max-w-4xl mx-auto">
          {error && (
            <Alert variant="destructive" className="mb-4">
              <AlertDescription>{error}</AlertDescription>
            </Alert>
          )}
          
          <div className="space-y-4">
            {messages.map(renderMessage)}
            
            {/* AI 打字指示器 */}
            {isTyping && (
              <div className="flex gap-3">
                <Avatar className="h-8 w-8 flex-shrink-0">
                  <AvatarFallback>
                    <Bot className="h-4 w-4" />
                  </AvatarFallback>
                </Avatar>
                <div className="flex-1">
                  <div className="flex items-center gap-2 mb-1">
                    <span className="text-sm font-medium">AI 助手</span>
                    <span className="text-xs text-muted-foreground">正在输入...</span>
                  </div>
                  <Card className="p-3 bg-muted inline-block">
                    <div className="flex items-center space-x-1">
                      <div className="flex space-x-1">
                        <div className="h-2 w-2 bg-current rounded-full animate-bounce [animation-delay:-0.3s]"></div>
                        <div className="h-2 w-2 bg-current rounded-full animate-bounce [animation-delay:-0.15s]"></div>
                        <div className="h-2 w-2 bg-current rounded-full animate-bounce"></div>
                      </div>
                    </div>
                  </Card>
                </div>
              </div>
            )}
          </div>
          <div ref={messagesEndRef} />
        </div>
      </ScrollArea>

      {/* 输入区域 */}
      <div className="border-t bg-background p-4">
        <div className="max-w-4xl mx-auto">
          <div className="flex items-end space-x-2">
            <div className="flex-1">
              <Input
                ref={inputRef}
                value={message}
                onChange={(e) => setMessage(e.target.value)}
                onKeyPress={handleKeyPress}
                placeholder="请输入您的消息..."
                disabled={isSending}
                className="min-h-[44px] resize-none"
              />
            </div>
            <Button 
              onClick={handleSendMessage} 
              disabled={!message.trim() || isSending}
              size="icon"
              className="h-11 w-11"
            >
              {isSending ? (
                <Loader2 className="h-4 w-4 animate-spin" />
              ) : (
                <Send className="h-4 w-4" />
              )}
            </Button>
          </div>
          <div className="flex items-center justify-between text-xs text-muted-foreground mt-2">
            <span>按 Enter 键发送消息，Shift + Enter 换行</span>
            {!isConnected && (
              <span className="text-yellow-500">· WebSocket 未连接</span>
            )}
            {isConnected && (
              <span className="text-green-500">· 实时连接</span>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
