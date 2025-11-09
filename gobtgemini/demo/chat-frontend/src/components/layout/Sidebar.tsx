import { useState, useEffect } from 'react';
import { Link, useLocation, useNavigate } from 'react-router-dom';
import { MessageCircle, Plus, MoreHorizontal, Trash2, Loader2 } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { ScrollArea } from '@/components/ui/scroll-area';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { cn } from '@/lib/utils';
import { useChatStore } from '@/store/chatStore';
import { useConversationStore, Conversation } from '@/store/conversationStore';
import { apiClient } from '@/lib/api';

export function Sidebar() {
  const location = useLocation();
  const navigate = useNavigate();
  const [isCreating, setIsCreating] = useState(false);
  
  const { clearMessages } = useChatStore();
  const {
    conversations,
    currentConversationId,
    isLoading,
    error,
    setConversations,
    addConversation,
    removeConversation,
    setCurrentConversation,
    setLoading,
    setError,
    clearError,
  } = useConversationStore();

  // 加载对话列表
  useEffect(() => {
    loadConversations();
  }, []);

  const loadConversations = async () => {
    try {
      setLoading(true);
      clearError();
      const conversationList = await apiClient.getConversations();
      setConversations(conversationList);
    } catch (error) {
      console.error('加载对话列表失败:', error);
      setError('加载对话列表失败');
    } finally {
      setLoading(false);
    }
  };

  const handleNewConversation = async () => {
    if (isCreating) return;
    
    try {
      setIsCreating(true);
      const newConversation = await apiClient.createConversation();
      addConversation(newConversation);
      setCurrentConversation(newConversation.id);
      clearMessages();
      navigate(`/chat?id=${newConversation.id}`);
    } catch (error) {
      console.error('创建对话失败:', error);
      setError('创建对话失败');
    } finally {
      setIsCreating(false);
    }
  };

  const handleDeleteConversation = async (conversationId: string) => {
    try {
      await apiClient.deleteConversation(conversationId);
      removeConversation(conversationId);
      
      // 如果删除的是当前对话，跳转到主页
      if (conversationId === currentConversationId) {
        navigate('/chat');
        clearMessages();
      }
    } catch (error) {
      console.error('删除对话失败:', error);
      setError('删除对话失败');
    }
  };

  const formatTime = (timestamp: string) => {
    const date = new Date(timestamp);
    const now = new Date();
    const diff = now.getTime() - date.getTime();
    const days = Math.floor(diff / (1000 * 60 * 60 * 24));
    
    if (days === 0) {
      return '今天';
    } else if (days === 1) {
      return '昨天';
    } else if (days < 7) {
      return `${days}天前`;
    } else {
      return date.toLocaleDateString('zh-CN', {
        month: 'short',
        day: 'numeric',
      });
    }
  };

  if (isLoading && conversations.length === 0) {
    return (
      <div className="flex h-full flex-col">
        <div className="p-4">
          <Button
            onClick={handleNewConversation}
            className="w-full justify-start"
            variant="outline"
            disabled={isCreating}
          >
            {isCreating ? (
              <Loader2 className="mr-2 h-4 w-4 animate-spin" />
            ) : (
              <Plus className="mr-2 h-4 w-4" />
            )}
            新对话
          </Button>
        </div>
        <div className="flex-1 flex items-center justify-center">
          <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
        </div>
      </div>
    );
  }

  return (
    <div className="flex h-full flex-col">
      {/* 新对话按钮 */}
      <div className="p-4">
        <Button
          onClick={handleNewConversation}
          className="w-full justify-start"
          variant="outline"
          disabled={isCreating}
        >
          {isCreating ? (
            <Loader2 className="mr-2 h-4 w-4 animate-spin" />
          ) : (
            <Plus className="mr-2 h-4 w-4" />
          )}
          新对话
        </Button>
      </div>

      {/* 对话列表 */}
      <ScrollArea className="flex-1 px-4">
        <div className="space-y-2">
          {conversations.map((conversation) => {
            const isActive = currentConversationId === conversation.id;
            
            return (
              <div key={conversation.id} className="group relative">
                <Link
                  to={`/chat?id=${conversation.id}`}
                  className={cn(
                    "flex items-center justify-between rounded-lg p-3 text-left transition-colors hover:bg-accent",
                    isActive && "bg-accent"
                  )}
                  onClick={() => setCurrentConversation(conversation.id)}
                >
                  <div className="flex-1 space-y-1 overflow-hidden">
                    <div className="flex items-center space-x-2">
                      <MessageCircle className="h-4 w-4 flex-shrink-0 text-muted-foreground" />
                      <p className="truncate text-sm font-medium">
                        {conversation.title}
                      </p>
                    </div>
                    {conversation.last_message && (
                      <p className="truncate text-xs text-muted-foreground">
                        {conversation.last_message}
                      </p>
                    )}
                    <p className="text-xs text-muted-foreground">
                      {formatTime(conversation.updated_at)}
                    </p>
                  </div>
                  
                  {/* 操作按钮 */}
                  <div className="opacity-0 group-hover:opacity-100 transition-opacity">
                    <DropdownMenu>
                      <DropdownMenuTrigger asChild>
                        <Button
                          variant="ghost"
                          size="sm"
                          className="h-8 w-8 p-0"
                          onClick={(e) => e.preventDefault()}
                        >
                          <MoreHorizontal className="h-4 w-4" />
                        </Button>
                      </DropdownMenuTrigger>
                      <DropdownMenuContent align="end">
                        <DropdownMenuItem
                          onClick={() => handleDeleteConversation(conversation.id)}
                          className="text-destructive"
                        >
                          <Trash2 className="mr-2 h-4 w-4" />
                          删除对话
                        </DropdownMenuItem>
                      </DropdownMenuContent>
                    </DropdownMenu>
                  </div>
                </Link>
              </div>
            );
          })}
        </div>
      </ScrollArea>
    </div>
  );
}
