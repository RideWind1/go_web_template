package handlers

import (
	"go-chat-backend/middleware"
	"go-chat-backend/models"
	"go-chat-backend/services"
	"go-chat-backend/utils"
	"go-chat-backend/websocket"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// ChatHandler 聊天处理器
type ChatHandler struct {
	chatService   *services.ChatService
	llmService    *services.LLMService
	chromaService *services.ChromaService
	userService   *services.UserService
	hub           *websocket.Hub
}

// NewChatHandler 创建聊天处理器
func NewChatHandler(chatService *services.ChatService, llmService *services.LLMService, chromaService *services.ChromaService) *ChatHandler {
	return &ChatHandler{
		chatService:   chatService,
		llmService:    llmService,
		chromaService: chromaService,
	}
}

// SetUserService 设置用户服务（用于解决循环依赖）
func (h *ChatHandler) SetUserService(userService *services.UserService) {
	h.userService = userService
}

// SetWebSocketHub 设置WebSocket Hub
func (h *ChatHandler) SetWebSocketHub(hub *websocket.Hub) {
	h.hub = hub
}

// SendMessageRequest 发送消息请求结构
type SendMessageRequest struct {
	ConversationID uuid.UUID `json:"conversation_id" binding:"required"`
	Content string `json:"content" binding:"required,min=1,max=4000"`
}

// SendMessageResponse 发送消息响应结构
type SendMessageResponse struct {
	UserMessage      *models.ChatMessage `json:"user_message"`
	AssistantMessage *models.ChatMessage `json:"assistant_message"`
	ProcessingTime   string              `json:"processing_time"`
}

// 定义一个用于绑定请求体的结构体
type CreateConversationRequest struct {
	Title string `json:"title"`
}

// SendMessage 发送聊天消息
func (h *ChatHandler) SendMessage(c *gin.Context) {
	startTime := time.Now()

	// 获取用户信息
	user, err := middleware.GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, utils.ErrorResponse{
			Error: "无效的认证信息",
			Code:  "INVALID_AUTH",
		})
		return
	}

	// 解析请求 (现在会包含 ConversationID)
	var req SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{
			Error:   "请求参数错误，需要 conversation_id 和 content",
			Code:    "INVALID_REQUEST",
			Message: err.Error(),
		})
		return
	}

	// 清理输入
	req.Content = strings.TrimSpace(req.Content)
	if req.Content == "" {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{
			Error: "消息内容不能为空",
			Code:  "EMPTY_MESSAGE",
		})
		return
	}

	// 保存用户消息 (现在传入 ConversationID)
	userMessage, err := h.chatService.SendMessage(user.ID, req.ConversationID, req.Content, "user")
	if err != nil {
		logrus.WithError(err).Error("保存用户消息失败")
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Error: "消息保存失败",
			Code:  "MESSAGE_SAVE_FAILED",
		})
		return
	}

	// 获取用户偏好设置
	var userPreference *models.UserPreference
	if h.userService != nil {
		userPreference, err = h.userService.GetUserPreference(user.ID)
		if err != nil {
			logrus.WithError(err).Warn("获取用户偏好失败，使用默认设置")
			// 使用默认偏好
			userPreference = &models.UserPreference{
				LLMModel:      "gpt-3.5-turbo",
				Temperature:   0.7,
				MaxTokens:     2000,
				ContextWindow: 10,
				MemoryEnabled: true,
			}
		}
	} else {
		// 默认偏好
		userPreference = &models.UserPreference{
			LLMModel:      "gpt-3.5-turbo",
			Temperature:   0.7,
			MaxTokens:     2000,
			ContextWindow: 10,
			MemoryEnabled: true,
		}
	}

	// 获取上下文消息 (现在也需要 ConversationID)
	contextMessages, err := h.chatService.GetRecentMessages(req.ConversationID, userPreference.ContextWindow)
	if err != nil {
		logrus.WithError(err).Warn("获取上下文消息失败")
		contextMessages = []models.ChatMessage{*userMessage}
	}

	// 如果启用了记忆功能，搜索相关记忆
	var memoryContext []string
	if userPreference.MemoryEnabled && h.chromaService != nil {
		memoryContext, err = h.chromaService.SearchMemory(user.ID, req.Content, 3)
		if err != nil {
			logrus.WithError(err).Warn("搜索记忆失败")
		}
	}

	// 如果有记忆上下文，将其添加到系统提示中
	if len(memoryContext) > 0 {
		memoryPrompt := "相关记忆：" + strings.Join(memoryContext, "\n")
		if userPreference.SystemPrompt != "" {
			userPreference.SystemPrompt += "\n\n" + memoryPrompt
		} else {
			userPreference.SystemPrompt = memoryPrompt
		}
	}

	// 生成AI回复
	response, err := h.llmService.GenerateResponse(contextMessages, userPreference)
	if err != nil {
		logrus.WithError(err).Error("AI回复生成失败")
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Error: "AI服务暂时不可用，请稍后再试",
			Code:  "AI_SERVICE_ERROR",
		})
		return
	}

	// 保存AI回复 (现在也传入 ConversationID)
	assistantMessage, err := h.chatService.SendMessage(user.ID, req.ConversationID, response, "assistant")
	if err != nil {
		logrus.WithError(err).Error("保存AI回复失败")
		// 不阻止请求，但记录错误
	}

	// 如果启用了记忆功能，保存对话到向量数据库
	if userPreference.MemoryEnabled && h.chromaService != nil {
		// 保存用户消息
		if err := h.chromaService.AddMemory(user.ID, req.Content, "user"); err != nil {
			logrus.WithError(err).Warn("保存用户消息到记忆失败")
		}

		// 保存AI回复
		if err := h.chromaService.AddMemory(user.ID, response, "assistant"); err != nil {
			logrus.WithError(err).Warn("保存AI回复到记忆失败")
		}
	}

	// 通过WebSocket发送实时消息
	if h.hub != nil {
		wsMessage := websocket.Message{
			Type:      "chat_response",
			Content:   response,
			UserID:    user.ID,
			Username:  user.Username,
			Timestamp: time.Now(),
			Data: gin.H{
				"user_message":      userMessage,
				"assistant_message": assistantMessage,
			},
		}
		if err := h.hub.SendToUser(user.ID, wsMessage); err != nil {
			logrus.WithError(err).Warn("发送WebSocket消息失败")
		}
	}

	processingTime := time.Since(startTime)

	// 返回响应
	c.JSON(http.StatusOK, utils.SuccessResponse{
		Data: SendMessageResponse{
			UserMessage:      userMessage,
			AssistantMessage: assistantMessage,
			ProcessingTime:   processingTime.String(),
		},
		Message: "消息发送成功",
	})

	logrus.WithFields(logrus.Fields{
		"user_id":         user.ID,
		"message_length":  len(req.Content),
		"response_length": len(response),
		"processing_time": processingTime.String(),
	}).Info("聊天消息处理完成")
}

// GetChatHistory 获取聊天历史
func (h *ChatHandler) GetChatHistory(c *gin.Context) {
	// 获取用户信息
	user, err := middleware.GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, utils.ErrorResponse{
			Error: "无效的认证信息",
			Code:  "INVALID_AUTH",
		})
		return
	}

	// 获取查询参数
	limitStr := c.DefaultQuery("limit", "50")
	offsetStr := c.DefaultQuery("offset", "0")
	

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 100 {
		limit = 50
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}


	// 获取聊天历史
	messages, err := h.chatService.GetChatHistory(user.ID,limit, offset)
	if err != nil {
		logrus.WithError(err).Error("获取聊天历史失败")
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Error: "获取聊天历史失败",
			Code:  "HISTORY_FETCH_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, utils.SuccessResponse{
		Data: gin.H{
			"messages": messages,
			"limit":    limit,
			"offset":   offset,
			"count":    len(messages),
		},
	})
}

// DeleteMessage 删除消息
func (h *ChatHandler) DeleteMessage(c *gin.Context) {
	// 获取用户信息
	user, err := middleware.GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, utils.ErrorResponse{
			Error: "无效的认证信息",
			Code:  "INVALID_AUTH",
		})
		return
	}

	// 获取消息ID
	messageIDStr := c.Param("id")
	messageID, err := uuid.Parse(messageIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{
			Error: "无效的消息ID",
			Code:  "INVALID_MESSAGE_ID",
		})
		return
	}

	// 删除消息
	err = h.chatService.DeleteMessage(user.ID, messageID)
	if err != nil {
		logrus.WithError(err).Error("删除消息失败")
		c.JSON(http.StatusNotFound, utils.ErrorResponse{
			Error: err.Error(),
			Code:  "DELETE_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, utils.SuccessResponse{
		Message: "消息删除成功",
	})

	logrus.WithFields(logrus.Fields{
		"user_id":    user.ID,
		"message_id": messageID,
	}).Info("消息删除成功")
}

// ClearHistory 清空聊天历史
func (h *ChatHandler) ClearHistory(c *gin.Context) {
	// 获取用户信息
	user, err := middleware.GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, utils.ErrorResponse{
			Error: "无效的认证信息",
			Code:  "INVALID_AUTH",
		})
		return
	}

	// 清空聊天历史
	err = h.chatService.ClearHistory(user.ID)
	if err != nil {
		logrus.WithError(err).Error("清空聊天历史失败")
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Error: "清空失败",
			Code:  "CLEAR_FAILED",
		})
		return
	}

	// 如果启用了Chroma，也清空记忆
	if h.chromaService != nil {
		if err := h.chromaService.ClearUserMemory(user.ID); err != nil {
			logrus.WithError(err).Warn("清空用户记忆失败")
		}
	}

	c.JSON(http.StatusOK, utils.SuccessResponse{
		Message: "聊天历史清空成功",
	})

	logrus.WithField("user_id", user.ID).Info("聊天历史清空成功")
}

//获取上下会话列表
func (h *ChatHandler) GetConversations(c *gin.Context) {
	// 1. 从JWT中间件设置的上下文中获取用户ID
	userIDClaim, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未认证"})
		return
	}
	userID, ok := userIDClaim.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "无效的用户ID格式"})
		return
	}

	// 2. 从URL查询参数中获取分页信息 (可选，但推荐)
	// 如果前端没有提供，则使用默认值
	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")
	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)

	// 3. 调用您已经存在的 Service 函数来获取数据
	// 假设您的 ChatHandler 结构体中有一个名为 chatService 的字段
	sessions, err := h.chatService.GetChatSessions(userID, limit, offset)
	if err != nil {
		// Service 层已经记录了详细日志，这里直接返回通用错误
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取对话列表失败"})
		return
	}

	// 4. 成功后，返回对话列表
	c.JSON(http.StatusOK, gin.H{
		"data": sessions,
	})
}

func (h *ChatHandler) CreateConversation(c *gin.Context) {
	// 1. 从JWT中间件获取用户ID
	userIDClaim, _ := c.Get("user_id")
	userID := userIDClaim.(uuid.UUID)

	// 2. 绑定请求体中的JSON数据
	var req CreateConversationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求数据"})
		return
	}
	
	// 3. 调用 Service 函数来创建对话
	newSession, err := h.chatService.CreateChatSession(userID, req.Title)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建对话失败"})
		return
	}

	// 4. 成功后，返回新创建的对话 (HTTP状态码 201 Created 更合适)
	c.JSON(http.StatusCreated, gin.H{
		"data": newSession,
	})
}


// GetConversationHistory 获取指定对话的聊天历史
func (h *ChatHandler) GetOneConversationHistory(c *gin.Context) {
    // --- 第1步：从 JWT 上下文中获取当前用户信息 ---
	// 这是为了安全，确保用户只能访问自己的聊天记录，尽管当前逻辑暂未校验。
	_, err := middleware.GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的认证信息"})
		return
	}

	// --- 第2步：从 URL 路径参数中解析对话ID ---
	// ":id" 对应 c.Param("id")
	conversationIDStr := c.Param("id")
	conversationID, err := uuid.Parse(conversationIDStr)
	if err != nil {
		// 如果解析失败，说明URL中的ID不是一个有效的UUID格式
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的对话ID格式"})
		return
	}

	// --- 第3步：从 URL 查询参数中解析分页信息 ---
	limitStr := c.DefaultQuery("limit", "50")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 200 { // 增加一个最大值限制，防止滥用
		limit = 50
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	// --- 第4步：调用 Service 层函数，执行业务逻辑 ---
	// h.chatService 是您在 ChatHandler 结构体中定义的 ChatService 实例
	messages, err := h.chatService.GetOneConversationHistory(conversationID, limit, offset)
	if err != nil {
		// 如果 Service 层返回错误（例如数据库查询失败），则向前端返回服务器内部错误
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取对话历史失败"})
		return
	}

	// --- 第5步：成功后，将查询结果以 JSON 格式返回给前端 ---
	c.JSON(http.StatusOK, gin.H{
		"message": "获取成功",
		"data": gin.H{
			"messages": messages,
			"limit":    limit,
			"offset":   offset,
		},
	})
}