import { Link } from 'react-router-dom';
import { Home, ArrowLeft } from 'lucide-react';
import { Button } from '@/components/ui/button';

export function NotFoundPage() {
  return (
    <div className="min-h-screen flex items-center justify-center p-4">
      <div className="max-w-md w-full text-center space-y-6">
        {/* 404 大标题 */}
        <div className="space-y-2">
          <h1 className="text-6xl font-bold text-muted-foreground">404</h1>
          <h2 className="text-2xl font-semibold">页面未找到</h2>
          <p className="text-muted-foreground">
            抱歉，您访问的页面不存在或已被移除。
          </p>
        </div>

        {/* 操作按钮 */}
        <div className="flex flex-col sm:flex-row gap-4 justify-center">
          <Button onClick={() => window.history.back()} variant="outline">
            <ArrowLeft className="mr-2 h-4 w-4" />
            返回上一页
          </Button>
          <Button asChild>
            <Link to="/">
              <Home className="mr-2 h-4 w-4" />
              回到首页
            </Link>
          </Button>
        </div>

        {/* 帮助信息 */}
        <div className="text-sm text-muted-foreground space-y-2">
          <p>如果您认为这是一个错误，请检查：</p>
          <ul className="list-disc list-inside text-left space-y-1">
            <li>URL 是否输入正确</li>
            <li>您是否有权限访问该页面</li>
            <li>页面是否已被移动到其他位置</li>
          </ul>
        </div>
      </div>
    </div>
  );
}
