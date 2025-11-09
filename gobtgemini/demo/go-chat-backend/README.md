# Go Chat Backend - 智能聊天后端服务

一个功能完整的Go语言聊天应用后端服务，支持上下文记忆的智能聊天体验，集成了JWT认证、WebSocket实时通信、Chroma向量数据库和外部LLM API。

## ✨ 功能特性

### 核心功能
- 🔐 **JWT用户认证系统** - 安全的用户注册、登录和token管理
- 💬 **智能聊天API** - 支持与大语言模型的对话交互
- 🧠 **上下文记忆** - 基于Chroma向量数据库的语义搜索和记忆存储
- ⚡ **WebSocket实时通信** - 实时消息推送和双向通信
- 🗄️ **PostgreSQL数据存储** - 可靠的用户信息和聊天记录存储
- 🎯 **可配置LLM API** - 支持OpenAI、Claude等多种大模型接入

### 技术特性
- 🏗️ **RESTful API设计** - 标准化的API接口
- 🔄 **CORS跨域支持** - 前端应用友好
- 📝 **结构化日志** - 完整的操作日志记录
- 🛡️ **安全防护** - 输入验证、SQL注入防护
- ⚙️ **环境配置管理** - 灵活的配置系统
- 🐳 **容器化支持** - Docker部署就绪

## 🏛️ 项目架构

```
go-chat-backend/
├── config/          # 配置管理
├── database/        # 数据库连接和迁移
├── handlers/        # HTTP请求处理器
├── middleware/      # 中间件（认证、CORS等）
├── models/          # 数据模型定义
├── services/        # 业务逻辑服务层
├── utils/           # 工具函数
├── websocket/       # WebSocket实时通信
├── main.go          # 应用入口点
├── go.mod           # Go模块定义
└── .env.example     # 环境变量示例
```

## 🚀 快速开始

### 环境要求
- Go 1.21+
- PostgreSQL 12+
- Chroma向量数据库
- 外部LLM API密钥（OpenAI/Claude等）

### 1. 克隆项目
```bash
git clone <your-repo-url>
cd go-chat-backend
```

### 2. 配置环境变量
```bash
cp .env.example .env
# 编辑 .env 文件，填入你的配置信息
```

### 3. 安装依赖
```bash
go mod tidy
```

### 4. 启动依赖服务

#### PostgreSQL
```bash
# 使用Docker启动PostgreSQL
docker run --name postgres-chat \
  -e POSTGRES_PASSWORD=your_password \
  -e POSTGRES_DB=go_chat_db \
  -p 5432:5432 \
  -d postgres:15
```

#### Chroma向量数据库
```bash
# 使用Docker启动Chroma
docker run --name chroma-chat \
  -p 8000:8000 \
  -d chromadb/chroma:latest
```

### 5. 运行应用
```bash
go run main.go
```

服务将在 `http://localhost:8080` 启动

## 📡 API接口文档

### 认证相关

#### 用户注册
```http
POST /api/v1/auth/register
Content-Type: application/json

{
  "username": "testuser",
  "email": "test@example.com",
  "password": "password123"
}
```

#### 用户登录
```http
POST /api/v1/auth/login
Content-Type: application/json

{
  "username_or_email": "testuser",
  "password": "password123"
}
```

#### 获取用户资料
```http
GET /api/v1/user/profile
Authorization: Bearer <your-jwt-token>
```

### 聊天相关

#### 发送消息
```http
POST /api/v1/chat/send
Authorization: Bearer <your-jwt-token>
Content-Type: application/json

{
  "content": "你好，请介绍一下自己"
}
```

#### 获取聊天历史
```http
GET /api/v1/chat/history?limit=20&offset=0
Authorization: Bearer <your-jwt-token>
```

#### 清空聊天历史
```http
POST /api/v1/chat/clear
Authorization: Bearer <your-jwt-token>
```

### WebSocket连接
```javascript
// 连接WebSocket
const ws = new WebSocket('ws://localhost:8080/api/v1/ws?token=<your-jwt-token>');

// 监听消息
ws.onmessage = function(event) {
  const message = JSON.parse(event.data);
  console.log('收到消息:', message);
};

// 发送心跳
ws.send(JSON.stringify({
  type: 'ping',
  content: 'ping'
}));
```

## ⚙️ 配置说明

### 环境变量配置

```bash
# 服务器配置
PORT=8080
GIN_MODE=debug

# 数据库配置
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=go_chat_db
DB_SSLMODE=disable

# JWT配置
JWT_SECRET=your_super_secret_jwt_key_here
JWT_EXPIRES_HOURS=24

# Chroma配置
CHROMA_HOST=localhost
CHROMA_PORT=8000
CHROMA_COLLECTION_NAME=chat_memory

# 外部LLM API配置
LLM_API_URL=https://api.openai.com/v1/chat/completions
LLM_API_KEY=your_llm_api_key
LLM_MODEL=gpt-3.5-turbo

# 日志配置
LOG_LEVEL=info
LOG_FILE=logs/app.log
```

### 支持的LLM API

- **OpenAI**: `https://api.openai.com/v1/chat/completions`
- **Azure OpenAI**: `https://your-resource.openai.azure.com/openai/deployments/your-deployment/chat/completions?api-version=2023-05-15`
- **Anthropic Claude**: 需要适配器或代理服务
- **其他兼容OpenAI格式的API**

## 🗄️ 数据库模型

### 用户表 (users)
- `id` - UUID主键
- `username` - 用户名（唯一）
- `email` - 邮箱（唯一）
- `password` - 加密密码
- `nickname` - 昵称
- `avatar` - 头像URL
- `is_active` - 是否激活
- `created_at/updated_at` - 时间戳

### 聊天消息表 (chat_messages)
- `id` - UUID主键
- `user_id` - 用户ID（外键）
- `content` - 消息内容
- `role` - 角色（user/assistant）
- `message_id` - 消息关联ID
- `metadata` - 元数据（JSON）
- `created_at/updated_at` - 时间戳

### 用户偏好表 (user_preferences)
- `id` - UUID主键
- `user_id` - 用户ID（外键）
- `llm_model` - 首选模型
- `temperature` - 温度参数
- `max_tokens` - 最大token数
- `system_prompt` - 系统提示词
- `context_window` - 上下文窗口大小
- `memory_enabled` - 是否启用记忆功能

## 🐳 Docker部署

### 构建镜像
```bash
docker build -t go-chat-backend .
```

### 使用Docker Compose
```yaml
version: '3.8'
services:
  postgres:
    image: postgres:15
    environment:
      POSTGRES_PASSWORD: your_password
      POSTGRES_DB: go_chat_db
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data

  chroma:
    image: chromadb/chroma:latest
    ports:
      - "8000:8000"
    volumes:
      - chroma_data:/chroma/chroma

  app:
    build: .
    ports:
      - "8080:8080"
    environment:
      - DB_HOST=postgres
      - CHROMA_HOST=chroma
      - LLM_API_KEY=your_api_key
    depends_on:
      - postgres
      - chroma

volumes:
  postgres_data:
  chroma_data:
```

## 🔧 开发指南

### 添加新的API端点
1. 在 `handlers/` 目录下创建处理器函数
2. 在 `main.go` 中注册路由
3. 如需数据库操作，在 `services/` 中添加业务逻辑
4. 更新API文档

### 自定义中间件
```go
// middleware/custom.go
func CustomMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // 中间件逻辑
        c.Next()
    }
}
```

### 扩展WebSocket功能
在 `websocket/hub.go` 中添加新的消息类型处理：
```go
case "new_message_type":
    // 处理新消息类型
```

## 🧪 测试

### 运行单元测试
```bash
go test ./...
```

### API测试示例
```bash
# 健康检查
curl http://localhost:8080/health

# 用户注册
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"test","email":"test@example.com","password":"password123"}'
```

## 📈 性能优化

- **数据库连接池**: 已配置连接池管理
- **JWT缓存**: 考虑添加Redis缓存JWT状态
- **消息队列**: 大并发时可集成消息队列
- **负载均衡**: 支持水平扩展

## 🔒 安全考虑

- ✅ JWT token认证
- ✅ 密码bcrypt加密
- ✅ SQL注入防护（GORM ORM）
- ✅ 输入验证和清理
- ✅ CORS配置
- ⚠️ 考虑添加速率限制
- ⚠️ 考虑添加HTTPS支持

## 📝 日志格式

应用使用结构化日志（JSON格式）：
```json
{
  "level": "info",
  "time": "2024-01-15T10:30:00Z",
  "msg": "用户登录成功",
  "user_id": "uuid-here",
  "username": "testuser"
}
```

## 🤝 贡献指南

1. Fork项目
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 创建Pull Request

## 📄 许可证

本项目采用 MIT 许可证 - 详见 [LICENSE](LICENSE) 文件

## 💬 支持

如果你有任何问题或建议，请：
- 创建 [Issue](https://github.com/your-repo/issues)
- 发送邮件到 your-email@example.com

---

**⭐ 如果这个项目对你有帮助，请给它一个星星！**
