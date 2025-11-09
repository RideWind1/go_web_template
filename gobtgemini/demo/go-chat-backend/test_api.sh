# Go Chat Backend API 测试集合
# 使用 HTTPie 进行 API 测试
# 安装: pip install httpie

# ===========================================
# 基本配置
# ===========================================
BASE_URL="http://localhost:8080"
API_URL="$BASE_URL/api/v1"

# ===========================================
# 1. 健康检查
# ===========================================
echo "✅ 测试健康检查..."
http GET $BASE_URL/health

# ===========================================
# 2. 用户注册
# ===========================================
echo "✅ 测试用户注册..."
http POST $API_URL/auth/register \
    username="testuser" \
    email="test@example.com" \
    password="password123"

# ===========================================
# 3. 用户登录
# ===========================================
echo "✅ 测试用户登录..."
response=$(http POST $API_URL/auth/login \
    username_or_email="testuser" \
    password="password123")

# 提取 JWT Token
token=$(echo $response | jq -r '.data.token')
echo "🔑 Token: $token"

# ===========================================
# 4. 获取用户资料
# ===========================================
echo "✅ 测试获取用户资料..."
http GET $API_URL/user/profile \
    "Authorization:Bearer $token"

# ===========================================
# 5. 更新用户资料
# ===========================================
echo "✅ 测试更新用户资料..."
http PUT $API_URL/user/profile \
    "Authorization:Bearer $token" \
    nickname="测试用户" \
    avatar="https://example.com/avatar.jpg"

# ===========================================
# 6. 发送聊天消息
# ===========================================
echo "✅ 测试发送聊天消息..."
http POST $API_URL/chat/send \
    "Authorization:Bearer $token" \
    content="你好，请介绍一下你自己"

# ===========================================
# 7. 获取聊天历史
# ===========================================
echo "✅ 测试获取聊天历史..."
http GET $API_URL/chat/history \
    "Authorization:Bearer $token" \
    limit==10 \
    offset==0

# ===========================================
# 8. 发送更多消息测试上下文
# ===========================================
echo "✅ 测试上下文记忆..."
http POST $API_URL/chat/send \
    "Authorization:Bearer $token" \
    content="我喜欢编程"

http POST $API_URL/chat/send \
    "Authorization:Bearer $token" \
    content="你记得我刚刚说什么吗？"

# ===========================================
# 9. 测试 Token 刷新
# ===========================================
echo "✅ 测试 Token 刷新..."
http POST $API_URL/auth/refresh \
    "Authorization:Bearer $token"

# ===========================================
# 10. 清空聊天历史
# ===========================================
echo "✅ 测试清空聊天历史..."
http POST $API_URL/chat/clear \
    "Authorization:Bearer $token"

# ===========================================
# 11. 错误测试 - 无效 Token
# ===========================================
echo "✅ 测试无效 Token..."
http GET $API_URL/user/profile \
    "Authorization:Bearer invalid_token"

# ===========================================
# 12. 错误测试 - 缺少参数
# ===========================================
echo "✅ 测试缺少参数..."
http POST $API_URL/auth/register \
    username="testuser2"
    # 缺少 email 和 password

echo "✅ API 测试完成！"
