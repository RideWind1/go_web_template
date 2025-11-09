import { Outlet, useLocation } from 'react-router-dom';
import { Header } from './Header';
import { Sidebar } from './Sidebar';
import { useAuthStore } from '@/store/authStore';
import { cn } from '@/lib/utils';

export function AppLayout() {
  const location = useLocation();
  const { isAuthenticated } = useAuthStore();
  
  // 在登录和注册页面不显示侧边栏
  const isAuthPage = location.pathname === '/login' || location.pathname === '/register';
  const showSidebar = isAuthenticated && !isAuthPage;

  return (
    <div className="min-h-screen bg-background">
      {/* 顶部导航栏 */}
      <Header />
      
      <div className="flex h-[calc(100vh-4rem)]">
        {/* 侧边栏 */}
        {showSidebar && (
          <aside className="w-64 border-r border-border bg-card">
            <Sidebar />
          </aside>
        )}
        
        {/* 主内容区域 */}
        <main 
          className={cn(
            "flex-1 overflow-hidden",
            !showSidebar && "w-full"
          )}
        >
          <Outlet />
        </main>
      </div>
    </div>
  );
}
