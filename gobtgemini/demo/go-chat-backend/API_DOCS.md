# Go Chat Backend API 文档

本文档详细描述了 Go Chat Backend 的所有 API 接口。

## 基本信息

- **Base URL**: `http://localhost:8080`
- **API Version**: `v1`
- **API Prefix**: `/api/v1`
- **Content-Type**: `application/json`

## 认证

本API使用JWT Bearer Token认证。

### 请求头格式
```
Authorization: Bearer <your-jwt-token>
```

### Token获取
通过登录或注册接口获取JWT Token。

## 错误响应格式

所有错误响应都遵循以下格式：

```json
{
  "error": "错误描述",
  "code": "ERROR_CODE",
  "message": "详细错误信息（可选）"
}
```

### 常见错误码

| 状态码 | 错误码 | 描述 |
|--------|--------|------|
| 400 | INVALID_REQUEST | 请求参数错误 |
| 401 | INVALID_AUTH | 认证失败 |
| 401 | MISSING_TOKEN | 缺少认证Token |
| 401 | INVALID_TOKEN | Token无效或过期 |
| 403 | FORBIDDEN | 无权限访问 |
| 404 | NOT_FOUND | 资源不存在 |
| 409 | CONFLICT | 资源冲突（如用户已存在） |
| 500 | INTERNAL_ERROR | 服务器内部错误 |

## 接口列表

### 1. 健康检查

#### GET /health

检查服务器状态。

**请求示例**
```http
GET /health
```

**响应示例**
```json
{
  "status": "healthy",
  "message": "Go Chat Backend is running"
}
```

---

## 认证相关接口

### 2. 用户注册

#### POST /api/v1/auth/register

注册新用户。

**请求参数**

| 参数 | 类型 | 必填 | 描述 | 验证规则 |
|------|------|------|------|----------|
| username | string | 是 | 用户名 | 3-50个字符 |
| email | string | 是 | 邮箱地址 | 有效的邮箱格式 |
| password | string | 是 | 密码 | 6-128个字符 |

**请求示例**
```http
POST /api/v1/auth/register
Content-Type: application/json

{
  "username": "testuser",
  "email": "test@example.com",
  "password": "password123"
}
```

**响应示例**
```json
{
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIs...",
    "user": {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "username": "testuser",
      "email": "test@example.com",
      "nickname": "testuser",
      "avatar": ""
    },
    "expires_at": "2024-01-16T10:30:00Z"
  },
  "message": "注册成功"
}
```

### 3. 用户登录

#### POST /api/v1/auth/login

用户登录获取JWT Token。

**请求参数**

| 参数 | 类型 | 必填 | 描述 |
|------|------|------|------|
| username_or_email | string | 是 | 用户名或邮箱 |
| password | string | 是 | 密码 |

**请求示例**
```http
POST /api/v1/auth/login
Content-Type: application/json

{
  "username_or_email": "testuser",
  "password": "password123"
}
```

**响应示例**
```json
{
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIs...",
    "user": {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "username": "testuser",
      "email": "test@example.com",
      "nickname": "测试用户",
      "avatar": "https://example.com/avatar.jpg",
      "created_at": "2024-01-15T10:30:00Z"
    },
    "expires_at": "2024-01-16T10:30:00Z"
  },
  "message": "登录成功"
}
```

### 4. 刷新Token

#### POST /api/v1/auth/refresh

刷新JWT Token。

**请求头**
```
Authorization: Bearer <current-token>
```

**请求示例**
```http
POST /api/v1/auth/refresh
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
```

**响应示例**
```json
{
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIs...",
    "expires_at": "2024-01-16T10:30:00Z"
  },
  "message": "Token刷新成功"
}
```

---

## 用户相关接口

### 5. 获取用户资料

#### GET /api/v1/user/profile

获取当前用户的详细资料。

**请求头**
```
Authorization: Bearer <your-token>
```

**请求示例**
```http
GET /api/v1/user/profile
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
```

**响应示例**
```json
{
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "username": "testuser",
    "email": "test@example.com",
    "nickname": "测试用户",
    "avatar": "https://example.com/avatar.jpg",
    "is_active": true,
    "created_at": "2024-01-15T10:30:00Z",
    "updated_at": "2024-01-15T10:30:00Z",
    "preferences": {
      "llm_model": "gpt-3.5-turbo",
      "temperature": 0.7,
      "max_tokens": 2000,
      "system_prompt": "你是一个智能助手...",
      "context_window": 10,
      "memory_enabled": true
    }
  }
}
```

### 6. 更新用户资料

#### PUT /api/v1/user/profile

更新用户资料。

**请求头**
```
Authorization: Bearer <your-token>
```

**请求参数**

| 参数 | 类型 | 必填 | 描述 |
|------|------|------|------|
| nickname | string | 否 | 用户昵称 |
| avatar | string | 否 | 头像URL |

**请求示例**
```http
PUT /api/v1/user/profile
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
Content-Type: application/json

{
  "nickname": "新昵称",
  "avatar": "https://example.com/new-avatar.jpg"
}
```

**响应示例**
```json
{
  "data": {
    "nickname": "新昵称",
    "avatar": "https://example.com/new-avatar.jpg"
  },
  "message": "资料更新成功"
}
```

---

## 聊天相关接口

### 7. 发送消息

#### POST /api/v1/chat/send

发送聊天消息并获取AI回复。

**请求头**
```
Authorization: Bearer <your-token>
```

**请求参数**

| 参数 | 类型 | 必填 | 描述 | 验证规则 |
|------|------|------|------|----------|
| content | string | 是 | 消息内容 | 1-4000个字符 |

**请求示例**
```http
POST /api/v1/chat/send
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
Content-Type: application/json

{
  "content": "你好，请介绍一下你自己"
}
```

**响应示例**
```json
{
  "data": {
    "user_message": {
      "id": "550e8400-e29b-41d4-a716-446655440001",
      "user_id": "550e8400-e29b-41d4-a716-446655440000",
      "content": "你好，请介绍一下你自己",
      "role": "user",
      "created_at": "2024-01-15T10:30:00Z"
    },
    "assistant_message": {
      "id": "550e8400-e29b-41d4-a716-446655440002",
      "user_id": "550e8400-e29b-41d4-a716-446655440000",
      "content": "你好！我是一个智能聊天助手，可以回答各种问题，提供帮助和建议。我具备上下文记忆功能，能够理解和记住我们之间的对话内容。有什么我可以帮助你的吗？",
      "role": "assistant",
      "created_at": "2024-01-15T10:30:01Z"
    },
    "processing_time": "1.234s"
  },
  "message": "消息发送成功"
}
```

### 8. 获取聊天历史

#### GET /api/v1/chat/history

获取用户的聊天历史记录。

**请求头**
```
Authorization: Bearer <your-token>
```

**查询参数**

| 参数 | 类型 | 必填 | 默认值 | 描述 |
|------|------|------|--------|------|
| limit | int | 否 | 50 | 每页数量，最大100 |
| offset | int | 否 | 0 | 偏移量 |

**请求示例**
```http
GET /api/v1/chat/history?limit=20&offset=0
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
```

**响应示例**
```json
{
  "data": {
    "messages": [
      {
        "id": "550e8400-e29b-41d4-a716-446655440001",
        "user_id": "550e8400-e29b-41d4-a716-446655440000",
        "content": "你好，请介绍一下你自己",
        "role": "user",
        "created_at": "2024-01-15T10:30:00Z"
      },
      {
        "id": "550e8400-e29b-41d4-a716-446655440002",
        "user_id": "550e8400-e29b-41d4-a716-446655440000",
        "content": "你好！我是一个智能聊天助手...",
        "role": "assistant",
        "created_at": "2024-01-15T10:30:01Z"
      }
    ],
    "limit": 20,
    "offset": 0,
    "count": 2
  }
}
```

### 9. 删除消息

#### DELETE /api/v1/chat/history/:id

删除指定的聊天消息。

**请求头**
```
Authorization: Bearer <your-token>
```

**路径参数**

| 参数 | 类型 | 必填 | 描述 |
|------|------|------|------|
| id | string | 是 | 消息ID (UUID) |

**请求示例**
```http
DELETE /api/v1/chat/history/550e8400-e29b-41d4-a716-446655440001
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
```

**响应示例**
```json
{
  "message": "消息删除成功"
}
```

### 10. 清空聊天历史

#### POST /api/v1/chat/clear

清空用户的所有聊天历史记录。

**请求头**
```
Authorization: Bearer <your-token>
```

**请求示例**
```http
POST /api/v1/chat/clear
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
```

**响应示例**
```json
{
  "message": "聊天历史清空成功"
}
```

---

## WebSocket 接口

### 11. WebSocket连接

#### GET /api/v1/ws

建立WebSocket连接以实现实时通信。

**请求头**
```
Authorization: Bearer <your-token>
```

**连接示例**
```javascript
const ws = new WebSocket('ws://localhost:8080/api/v1/ws', [], {
  headers: {
    'Authorization': 'Bearer ' + token
  }
});

// 监听连接打开
ws.onopen = function(event) {
  console.log('WebSocket 连接已建立');
};

// 监听消息
ws.onmessage = function(event) {
  const message = JSON.parse(event.data);
  console.log('收到消息:', message);
};

// 监听连接关闭
ws.onclose = function(event) {
  console.log('WebSocket 连接已关闭');
};

// 监听错误
ws.onerror = function(error) {
  console.error('WebSocket 错误:', error);
};
```

### WebSocket 消息格式

#### 发送消息格式
```json
{
  "type": "ping|chat|其他类型",
  "content": "消息内容"
}
```

#### 接收消息格式
```json
{
  "type": "system|pong|chat_response|其他类型",
  "content": "消息内容",
  "user_id": "用户ID",
  "username": "用户名",
  "timestamp": "2024-01-15T10:30:00Z",
  "data": {
    // 额外数据
  }
}
```

#### 支持的消息类型

1. **ping**: 心跳检测
   ```json
   {
     "type": "ping",
     "content": "ping"
   }
   ```
   响应：
   ```json
   {
     "type": "pong",
     "content": "pong",
     "timestamp": "2024-01-15T10:30:00Z"
   }
   ```

2. **system**: 系统消息
   ```json
   {
     "type": "system",
     "content": "欢迎使用智能聊天助手！",
     "timestamp": "2024-01-15T10:30:00Z"
   }
   ```

3. **chat_response**: 聊天回复通知
   ```json
   {
     "type": "chat_response",
     "content": "AI回复内容",
     "user_id": "550e8400-e29b-41d4-a716-446655440000",
     "username": "testuser",
     "timestamp": "2024-01-15T10:30:00Z",
     "data": {
       "user_message": { /* 用户消息对象 */ },
       "assistant_message": { /* AI回复消息对象 */ }
     }
   }
   ```

---

## 数据模型

### User (用户)
```json
{
  "id": "UUID",
  "username": "string",
  "email": "string",
  "nickname": "string",
  "avatar": "string",
  "is_active": "boolean",
  "created_at": "datetime",
  "updated_at": "datetime"
}
```

### ChatMessage (聊天消息)
```json
{
  "id": "UUID",
  "user_id": "UUID",
  "content": "string",
  "role": "user|assistant",
  "message_id": "string",
  "metadata": "json",
  "created_at": "datetime",
  "updated_at": "datetime"
}
```

### UserPreference (用户偏好)
```json
{
  "id": "UUID",
  "user_id": "UUID",
  "llm_model": "string",
  "temperature": "float",
  "max_tokens": "int",
  "system_prompt": "string",
  "context_window": "int",
  "memory_enabled": "boolean",
  "created_at": "datetime",
  "updated_at": "datetime"
}
```

---

## 状态码说明

| 状态码 | 含义 |
|--------|------|
| 200 | 请求成功 |
| 201 | 创建成功 |
| 400 | 请求参数错误 |
| 401 | 未授权 |
| 403 | 禁止访问 |
| 404 | 资源不存在 |
| 409 | 资源冲突 |
| 500 | 服务器内部错误 |

---

## 使用示例

### 完整的聊天流程

1. **用户注册**
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "email": "test@example.com",
    "password": "password123"
  }'
```

2. **用户登录**
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username_or_email": "testuser",
    "password": "password123"
  }'
```

3. **发送聊天消息**
```bash
curl -X POST http://localhost:8080/api/v1/chat/send \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "content": "你好，请介绍一下你自己"
  }'
```

4. **获取聊天历史**
```bash
curl -X GET "http://localhost:8080/api/v1/chat/history?limit=10&offset=0" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

---

## 注意事项

1. **Token过期**: JWT Token默认有效期为24小时，过期后需要重新登录或使用刷新接口
2. **请求限制**: 为防止滥用，某些接口可能有请求频率限制
3. **消息长度**: 单条聊天消息最大长度为4000字符
4. **历史记录**: 聊天历史记录会永久保存，除非用户主动删除
5. **WebSocket**: WebSocket连接会在Token过期时自动断开
6. **CORS**: API支持跨域请求，但建议在生产环境中配置具体的允许域名

---

## 更新日志

### v1.0.0 (2024-01-15)
- 初始版本发布
- 基本的用户认证功能
- 聊天消息发送和接收
- WebSocket实时通信
- Chroma向量数据库集成
- 上下文记忆功能
