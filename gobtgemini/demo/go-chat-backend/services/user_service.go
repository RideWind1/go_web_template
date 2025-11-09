package services

import (
	"errors"
	"go-chat-backend/models"
	"go-chat-backend/utils"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// UserService 用户服务
type UserService struct {
	db *gorm.DB
}

// NewUserService 创建用户服务
func NewUserService(db *gorm.DB) *UserService {
	return &UserService{db: db}
}

// Register 用户注册
func (s *UserService) Register(username, email, password string) (*models.User, error) {
	// 验证输入
	if !utils.ValidateUsername(username) {
		return nil, errors.New("用户名格式不正确")
	}
	if !utils.ValidateEmail(email) {
		return nil, errors.New("邮箱格式不正确")
	}
	if !utils.ValidatePassword(password) {
		return nil, errors.New("密码长度必须在6-128之间")
	}

	// 检查用户名和邮箱是否已存在
	var existingUser models.User
	if err := s.db.Where("username = ? OR email = ?", username, email).First(&existingUser).Error; err == nil {
		return nil, errors.New("用户名或邮箱已存在")
	}

	// 加密密码
	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		return nil, errors.New("密码加密失败")
	}

	// 创建用户
	user := &models.User{
		Username: username,
		Email:    email,
		Password: hashedPassword,
		Nickname: username, // 默认使用用户名作为昵称
		IsActive: true,
	}

	if err := s.db.Create(user).Error; err != nil {
		logrus.WithError(err).Error("用户注册失败")
		return nil, errors.New("用户注册失败")
	}

	// 创建用户偏好设置
	preference := &models.UserPreference{
		UserID:        user.ID,
		LLMModel:      "gpt-3.5-turbo",
		Temperature:   0.7,
		MaxTokens:     2000,
		ContextWindow: 10,
		MemoryEnabled: true,
	}

	if err := s.db.Create(preference).Error; err != nil {
		logrus.WithError(err).Warn("创建用户偏好设置失败")
	}

	return user, nil
}

// Login 用户登录
func (s *UserService) Login(usernameOrEmail, password string) (*models.User, error) {
	var user models.User

	// 查找用户（支持用户名或邮箱登录）
	err := s.db.Where("username = ? OR email = ?", usernameOrEmail, usernameOrEmail).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("用户不存在")
		}
		return nil, err
	}

	// 检查用户是否激活
	if !user.IsActive {
		return nil, errors.New("用户已被禁用")
	}

	// 验证密码
	if !utils.CheckPassword(password, user.Password) {
		return nil, errors.New("密码错误")
	}

	return &user, nil
}

// GetUserByID 根据ID获取用户
func (s *UserService) GetUserByID(userID uuid.UUID) (*models.User, error) {
	var user models.User
	err := s.db.Where("id = ? AND is_active = true", userID).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("用户不存在")
		}
		return nil, err
	}
	return &user, nil
}

// UpdateProfile 更新用户资料
func (s *UserService) UpdateProfile(userID uuid.UUID, updates map[string]interface{}) error {
	// 允许更新的字段
	allowedFields := map[string]bool{
		"nickname": true,
		"avatar":   true,
	}

	// 过滤不允许的字段
	filteredUpdates := make(map[string]interface{})
	for key, value := range updates {
		if allowedFields[key] {
			filteredUpdates[key] = value
		}
	}

	if len(filteredUpdates) == 0 {
		return errors.New("没有有效的更新字段")
	}

	err := s.db.Model(&models.User{}).Where("id = ?", userID).Updates(filteredUpdates).Error
	if err != nil {
		logrus.WithError(err).Error("用户资料更新失败")
		return errors.New("用户资料更新失败")
	}

	return nil
}

// GetUserPreference 获取用户偏好设置
func (s *UserService) GetUserPreference(userID uuid.UUID) (*models.UserPreference, error) {
	var preference models.UserPreference
	err := s.db.Where("user_id = ?", userID).First(&preference).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 如果没有找到，创建默认设置
			defaultPreference := &models.UserPreference{
				UserID:        userID,
				LLMModel:      "gpt-3.5-turbo",
				Temperature:   0.7,
				MaxTokens:     2000,
				ContextWindow: 10,
				MemoryEnabled: true,
			}
			if createErr := s.db.Create(defaultPreference).Error; createErr != nil {
				return nil, createErr
			}
			return defaultPreference, nil
		}
		return nil, err
	}
	return &preference, nil
}

// UpdateUserPreference 更新用户偏好设置
func (s *UserService) UpdateUserPreference(userID uuid.UUID, updates map[string]interface{}) error {
	// 允许更新的字段
	allowedFields := map[string]bool{
		"llm_model":      true,
		"temperature":    true,
		"max_tokens":     true,
		"system_prompt":  true,
		"context_window": true,
		"memory_enabled": true,
	}

	// 过滤不允许的字段
	filteredUpdates := make(map[string]interface{})
	for key, value := range updates {
		if allowedFields[key] {
			filteredUpdates[key] = value
		}
	}

	if len(filteredUpdates) == 0 {
		return errors.New("没有有效的更新字段")
	}

	err := s.db.Model(&models.UserPreference{}).Where("user_id = ?", userID).Updates(filteredUpdates).Error
	if err != nil {
		logrus.WithError(err).Error("用户偏好更新失败")
		return errors.New("用户偏好更新失败")
	}

	return nil
}
