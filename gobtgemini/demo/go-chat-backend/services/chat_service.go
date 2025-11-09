package services

import (
	"errors"
	"go-chat-backend/models"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// ChatService 聊天服务
type ChatService struct {
	db *gorm.DB
}

// NewChatService 创建聊天服务
func NewChatService(db *gorm.DB) *ChatService {
	return &ChatService{db: db}
}

// SendMessage 发送消息
func (s *ChatService) SendMessage(userID uuid.UUID, conversationID uuid.UUID, content string, role string) (*models.ChatMessage, error) {
	if content == "" {
		return nil, errors.New("消息内容不能为空")
	}

	// 创建消息
	message := &models.ChatMessage{
		// ⭐ 新增：为每条消息生成一个新的、唯一的ID ⭐
		MessageID: uuid.New(),

		// ⭐ 新增：关联消息到指定的对话 ⭐
		ConversationID: conversationID,

		UserID:  userID,
		Content: content,
		Role:    role,
	}

	if err := s.db.Create(message).Error; err != nil {
		logrus.WithError(err).Error("消息保存失败")
		return nil, errors.New("消息保存失败")
	}

	return message, nil
}

// GetChatHistory 获取聊天历史
func (s *ChatService) GetChatHistory(userID uuid.UUID, limit int, offset int) ([]models.ChatMessage, error) {
	var messages []models.ChatMessage

	if limit <= 0 || limit > 100 {
		limit = 50 // 默认限制
	}

	err := s.db.Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&messages).Error

	if err != nil {
		logrus.WithError(err).Error("获取聊天历史失败")
		return nil, errors.New("获取聊天历史失败")
	}

	// 反转数组，让最早的消息在前
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, nil
}

// GetRecentMessages 获取最近的消息（用于上下文）
func (s *ChatService) GetRecentMessages(conversationID uuid.UUID, limit int) ([]models.ChatMessage, error) {
	var messages []models.ChatMessage

	if limit <= 0 || limit > 50 {
		limit = 10 // 默认上下文窗口大小
	}

	err := s.db.Where("conversation_id = ?", conversationID).
		Order("created_at DESC").
		Limit(limit).
		Find(&messages).Error

	if err != nil {
		logrus.WithError(err).Error("获取最近消息失败")
		return nil, errors.New("获取最近消息失败")
	}

	// 反转数组，让最早的消息在前
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, nil
}

// DeleteMessage 删除消息
func (s *ChatService) DeleteMessage(userID, messageID uuid.UUID) error {
	result := s.db.Where("id = ? AND user_id = ?", messageID, userID).Delete(&models.ChatMessage{})
	if result.Error != nil {
		logrus.WithError(result.Error).Error("删除消息失败")
		return errors.New("删除消息失败")
	}

	if result.RowsAffected == 0 {
		return errors.New("消息不存在或无权删除")
	}

	return nil
}

// ClearHistory 清空聊天历史
func (s *ChatService) ClearHistory(userID uuid.UUID) error {
	err := s.db.Where("user_id = ?", userID).Delete(&models.ChatMessage{}).Error
	if err != nil {
		logrus.WithError(err).Error("清空聊天历史失败")
		return errors.New("清空聊天历史失败")
	}

	return nil
}

// CreateChatSession 创建聊天会话
func (s *ChatService) CreateChatSession(userID uuid.UUID, title string) (*models.ChatSession, error) {
	if title == "" {
		title = "新的聊天" + time.Now().Format("01-02 15:04")
	}

	session := &models.ChatSession{
		ID:       uuid.New(), // 明确地在代码中生成ID
		UserID:   userID,
		Title:    title,
		IsActive: true,
	}

	if err := s.db.Create(session).Error; err != nil {
		logrus.WithError(err).Error("创建聊天会话失败")
		return nil, errors.New("创建聊天会话失败")
	}
	// 因为ID是在代码中生成的，所以返回的session对象一定包含正确的ID
	logrus.WithFields(logrus.Fields{
		"user_id":    userID,
		"session_id": session.ID,
	}).Info("新的聊天会话已成功创建")

	return session, nil
}

// GetChatSessions 获取用户的聊天会话列表
func (s *ChatService) GetChatSessions(userID uuid.UUID, limit int, offset int) ([]models.ChatSession, error) {
	var sessions []models.ChatSession

	if limit <= 0 || limit > 100 {
		limit = 20
	}

	err := s.db.Where("user_id = ? AND is_active = true", userID).
		Order("updated_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&sessions).Error

	if err != nil {
		logrus.WithError(err).Error("获取聊天会话列表失败")
		return nil, errors.New("获取聊天会话列表失败")
	}

	return sessions, nil
}

// UpdateChatSession 更新聊天会话
func (s *ChatService) UpdateChatSession(userID, sessionID uuid.UUID, updates map[string]interface{}) error {
	allowedFields := map[string]bool{
		"title":       true,
		"description": true,
	}

	filteredUpdates := make(map[string]interface{})
	for key, value := range updates {
		if allowedFields[key] {
			filteredUpdates[key] = value
		}
	}

	if len(filteredUpdates) == 0 {
		return errors.New("没有有效的更新字段")
	}

	result := s.db.Model(&models.ChatSession{}).
		Where("id = ? AND user_id = ?", sessionID, userID).
		Updates(filteredUpdates)

	if result.Error != nil {
		logrus.WithError(result.Error).Error("更新聊天会话失败")
		return errors.New("更新聊天会话失败")
	}

	if result.RowsAffected == 0 {
		return errors.New("会话不存在或无权修改")
	}

	return nil
}

// DeleteChatSession 删除聊天会话
func (s *ChatService) DeleteChatSession(userID, sessionID uuid.UUID) error {
	// 软删除，只是标记为非激活状态
	result := s.db.Model(&models.ChatSession{}).
		Where("id = ? AND user_id = ?", sessionID, userID).
		Update("is_active", false)

	if result.Error != nil {
		logrus.WithError(result.Error).Error("删除聊天会话失败")
		return errors.New("删除聊天会话失败")
	}

	if result.RowsAffected == 0 {
		return errors.New("会话不存在或无权删除")
	}

	return nil
}

func (s *ChatService) GetOneConversationHistory(conversationID uuid.UUID, limit, offset int) ([]models.ChatMessage, error) {
	// 1. 声明一个空的 ChatMessage 切片，用于存放查询结果
	var messages []models.ChatMessage

	// 2. (可选但推荐) 对分页参数进行安全检查和设置默认值
	if limit <= 0 || limit > 200 { // 设置一个最大值，防止一次请求过多数据
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	// 3. 使用 GORM 构建数据库查询
	err := s.db.
		// 关键查询条件：只查找属于特定 conversation_id 的消息
		Where("conversation_id = ?", conversationID).
		// 按创建时间升序排列，确保聊天记录从旧到新，顺序正确
		Order("created_at ASC").
		// 应用分页：限制返回的记录数量
		Limit(limit).
		// 应用分页：跳过指定的记录数量
		Offset(offset).
		// 执行查询，并将结果填充到 messages 切片中
		Find(&messages).Error

	// 4. 检查查询过程中是否发生错误
	if err != nil {
		logrus.WithError(err).
			WithField("conversation_id", conversationID). // 在日志中记录上下文信息，便于排查
			Error("从数据库获取对话历史失败")
		return nil, errors.New("获取对话历史失败")
	}

	// 5. 如果没有错误，返回查询到的消息列表
	return messages, nil
}
