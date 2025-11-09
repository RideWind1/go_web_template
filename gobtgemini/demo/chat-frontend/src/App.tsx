import { ThemeProvider } from '@/contexts/ThemeContext';
import { AppRouter } from '@/components/AppRouter';
import { Toaster } from '@/components/ui/toaster';
import { useWebSocket } from '@/hooks/useWebSocket';
import './App.css';

function AppContent() {
  // 初始化 WebSocket 连接
  useWebSocket();
  
  return (
    <>
      <AppRouter />
      <Toaster />
    </>
  );
}

function App() {
  return (
    <ThemeProvider defaultTheme="system" storageKey="chat-app-theme">
      <AppContent />
    </ThemeProvider>
  );
}

export default App;
