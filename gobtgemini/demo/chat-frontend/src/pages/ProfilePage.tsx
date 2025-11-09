import { useState } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { Camera, Save, Loader2 } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from '@/components/ui/card';
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { useAuthStore } from '@/store/authStore';

const profileSchema = z.object({
  username: z.string().min(3, '用户名至少3个字符').max(50, '用户名不能超过50个字符'),
  email: z.string().email('请输入有效的邮箱地址'),
  currentPassword: z.string().optional(),
  newPassword: z.string().optional(),
  confirmPassword: z.string().optional(),
}).refine((data) => {
  if (data.newPassword || data.confirmPassword) {
    return data.currentPassword && data.newPassword === data.confirmPassword;
  }
  return true;
}, {
  message: '密码信息不完整或不匹配',
  path: ['confirmPassword'],
});

type ProfileForm = z.infer<typeof profileSchema>;

export function ProfilePage() {
  const [isLoading, setIsLoading] = useState(false);
  const [message, setMessage] = useState<{ type: 'success' | 'error'; text: string } | null>(null);
  const { user, updateUser } = useAuthStore();
  
  const {
    register,
    handleSubmit,
    formState: { errors },
    reset,
  } = useForm<ProfileForm>({
    resolver: zodResolver(profileSchema),
    defaultValues: {
      username: user?.username || '',
      email: user?.email || '',
    },
  });

  const onSubmit = async (data: ProfileForm) => {
    try {
      setIsLoading(true);
      setMessage(null);
      
      // TODO: 实现实际的API调用
      // 模拟更新操作
      await new Promise(resolve => setTimeout(resolve, 1000));
      
      // 更新用户信息
      updateUser({
        username: data.username,
        email: data.email,
      });
      
      setMessage({ type: 'success', text: '个人资料更新成功' });
      
      // 清空密码字段
      reset({
        username: data.username,
        email: data.email,
        currentPassword: '',
        newPassword: '',
        confirmPassword: '',
      });
    } catch (error) {
      setMessage({ type: 'error', text: '更新失败，请稍后再试' });
    } finally {
      setIsLoading(false);
    }
  };

  const handleAvatarChange = async (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0];
    if (!file) return;

    // TODO: 实现头像上传逻辑
    console.log('Avatar upload:', file);
  };

  if (!user) {
    return (
      <div className="flex items-center justify-center h-full">
        <p className="text-muted-foreground">用户信息加载中...</p>
      </div>
    );
  }

  return (
    <div className="container max-w-2xl mx-auto p-4">
      <div className="space-y-6">
        <div>
          <h1 className="text-3xl font-bold">个人资料</h1>
          <p className="text-muted-foreground">
            管理您的个人信息和账户设置
          </p>
        </div>

        {message && (
          <Alert variant={message.type === 'error' ? 'destructive' : 'default'}>
            <AlertDescription>{message.text}</AlertDescription>
          </Alert>
        )}

        {/* 头像设置 */}
        <Card>
          <CardHeader>
            <CardTitle>头像</CardTitle>
            <CardDescription>更新您的个人头像</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="flex items-center space-x-4">
              <Avatar className="h-20 w-20">
                <AvatarImage src={user.avatar} alt={user.username} />
                <AvatarFallback className="text-lg">
                  {user.username.charAt(0).toUpperCase()}
                </AvatarFallback>
              </Avatar>
              <div className="space-y-2">
                <Label htmlFor="avatar-upload" className="cursor-pointer">
                  <div className="flex items-center space-x-2">
                    <Button type="button" variant="outline" size="sm">
                      <Camera className="mr-2 h-4 w-4" />
                      选择头像
                    </Button>
                  </div>
                  <Input
                    id="avatar-upload"
                    type="file"
                    accept="image/*"
                    className="hidden"
                    onChange={handleAvatarChange}
                  />
                </Label>
                <p className="text-xs text-muted-foreground">
                  支持 JPG、PNG 格式，大小不超过 2MB
                </p>
              </div>
            </div>
          </CardContent>
        </Card>

        {/* 基本信息设置 */}
        <Card>
          <CardHeader>
            <CardTitle>基本信息</CardTitle>
            <CardDescription>更新您的基本信息</CardDescription>
          </CardHeader>
          
          <form onSubmit={handleSubmit(onSubmit)}>
            <CardContent className="space-y-4">
              <div className="space-y-2">
                <Label htmlFor="username">用户名</Label>
                <Input
                  id="username"
                  {...register('username')}
                  disabled={isLoading}
                />
                {errors.username && (
                  <p className="text-sm text-destructive">{errors.username.message}</p>
                )}
              </div>
              
              <div className="space-y-2">
                <Label htmlFor="email">邮箱地址</Label>
                <Input
                  id="email"
                  type="email"
                  {...register('email')}
                  disabled={isLoading}
                />
                {errors.email && (
                  <p className="text-sm text-destructive">{errors.email.message}</p>
                )}
              </div>
              
              {/* 密码修改 */}
              <div className="space-y-4 pt-4 border-t">
                <div>
                  <h4 className="text-sm font-medium mb-2">修改密码</h4>
                  <p className="text-xs text-muted-foreground mb-4">
                    如果不需要修改密码，请保持下方字段为空
                  </p>
                </div>
                
                <div className="space-y-2">
                  <Label htmlFor="currentPassword">当前密码</Label>
                  <Input
                    id="currentPassword"
                    type="password"
                    {...register('currentPassword')}
                    disabled={isLoading}
                  />
                </div>
                
                <div className="space-y-2">
                  <Label htmlFor="newPassword">新密码</Label>
                  <Input
                    id="newPassword"
                    type="password"
                    {...register('newPassword')}
                    disabled={isLoading}
                  />
                </div>
                
                <div className="space-y-2">
                  <Label htmlFor="confirmPassword">确认新密码</Label>
                  <Input
                    id="confirmPassword"
                    type="password"
                    {...register('confirmPassword')}
                    disabled={isLoading}
                  />
                  {errors.confirmPassword && (
                    <p className="text-sm text-destructive">{errors.confirmPassword.message}</p>
                  )}
                </div>
              </div>
            </CardContent>
            
            <CardFooter>
              <Button type="submit" disabled={isLoading}>
                {isLoading ? (
                  <>
                    <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                    保存中...
                  </>
                ) : (
                  <>
                    <Save className="mr-2 h-4 w-4" />
                    保存更改
                  </>
                )}
              </Button>
            </CardFooter>
          </form>
        </Card>
        
        {/* 账户统计 */}
        <Card>
          <CardHeader>
            <CardTitle>账户统计</CardTitle>
          </CardHeader>
          <CardContent className="space-y-2">
            <div className="flex justify-between">
              <span className="text-muted-foreground">注册时间</span>
              <span>{new Date(user.created_at || Date.now()).toLocaleDateString('zh-CN')}</span>
            </div>
            <div className="flex justify-between">
              <span className="text-muted-foreground">最后更新</span>
              <span>{new Date(user.updated_at || Date.now()).toLocaleDateString('zh-CN')}</span>
            </div>
            <div className="flex justify-between">
              <span className="text-muted-foreground">用户ID</span>
              <span className="font-mono text-sm">{user.id}</span>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
