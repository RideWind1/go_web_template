package handlers

import (
	"go-chat-backend/services"
	"go-chat-backend/utils"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// AuthHandler 认证处理器
type AuthHandler struct {
	userService *services.UserService
}

// NewAuthHandler 创建认证处理器
func NewAuthHandler(userService *services.UserService) *AuthHandler {
	return &AuthHandler{
		userService: userService,
	}
}

// RegisterRequest 注册请求结构
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6,max=128"`
}

// LoginRequest 登录请求结构
type LoginRequest struct {
	UsernameOrEmail string `json:"username_or_email" binding:"required"`
	Password        string `json:"password" binding:"required"`
}

// AuthResponse 认证响应结构
type AuthResponse struct {
	Token     string      `json:"token"`
	User      interface{} `json:"user"`
	ExpiresAt time.Time   `json:"expires_at"`
}

// UpdateProfileRequest 更新资料请求结构
type UpdateProfileRequest struct {
	Nickname string `json:"nickname,omitempty"`
	Avatar   string `json:"avatar,omitempty"`
}

// Register 用户注册
func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{
			Error:   "请求参数错误",
			Code:    "INVALID_REQUEST",
			Message: err.Error(),
		})
		return
	}

	// 清理输入数据
	req.Username = strings.TrimSpace(req.Username)
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))

	// 额外验证
	if !utils.ValidateEmail(req.Email) {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{
			Error: "邮箱格式不正确",
			Code:  "INVALID_EMAIL",
		})
		return
	}

	if !utils.ValidateUsername(req.Username) {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{
			Error: "用户名长度必须在3-50个字符之间",
			Code:  "INVALID_USERNAME",
		})
		return
	}

	// 注册用户
	user, err := h.userService.Register(req.Username, req.Email, req.Password)
	if err != nil {
		logrus.WithError(err).Error("用户注册失败")
		c.JSON(http.StatusConflict, utils.ErrorResponse{
			Error: err.Error(),
			Code:  "REGISTRATION_FAILED",
		})
		return
	}

	// 生成JWT token
	token, err := utils.GenerateJWT(user.ID, user.Username, user.Email)
	if err != nil {
		logrus.WithError(err).Error("JWT生成失败")
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Error: "登录凭证生成失败",
			Code:  "TOKEN_GENERATION_FAILED",
		})
		return
	}

	// 返回成功响应
	c.JSON(http.StatusCreated, utils.SuccessResponse{
		Data: AuthResponse{
			Token: token,
			User: gin.H{
				"id":       user.ID,
				"username": user.Username,
				"email":    user.Email,
				"nickname": user.Nickname,
				"avatar":   user.Avatar,
			},
			ExpiresAt: time.Now().Add(24 * time.Hour),
		},
		Message: "注册成功",
	})

	logrus.WithFields(logrus.Fields{
		"user_id":  user.ID,
		"username": user.Username,
		"email":    user.Email,
	}).Info("用户注册成功")
}

// Login 用户登录
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{
			Error:   "请求参数错误",
			Code:    "INVALID_REQUEST",
			Message: err.Error(),
		})
		return
	}

	// 清理输入数据
	req.UsernameOrEmail = strings.TrimSpace(req.UsernameOrEmail)

	// 用户登录
	user, err := h.userService.Login(req.UsernameOrEmail, req.Password)
	if err != nil {
		logrus.WithError(err).WithField("username_or_email", req.UsernameOrEmail).Warn("用户登录失败")
		c.JSON(http.StatusUnauthorized, utils.ErrorResponse{
			Error: "用户名/邮箱或密码错误",
			Code:  "LOGIN_FAILED",
		})
		return
	}

	// 生成JWT token
	token, err := utils.GenerateJWT(user.ID, user.Username, user.Email)
	if err != nil {
		logrus.WithError(err).Error("JWT生成失败")
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Error: "登录凭证生成失败",
			Code:  "TOKEN_GENERATION_FAILED",
		})
		return
	}

	// 返回成功响应
	c.JSON(http.StatusOK, utils.SuccessResponse{
		Data: AuthResponse{
			Token: token,
			User: gin.H{
				"id":         user.ID,
				"username":   user.Username,
				"email":      user.Email,
				"nickname":   user.Nickname,
				"avatar":     user.Avatar,
				"created_at": user.CreatedAt,
			},
			ExpiresAt: time.Now().Add(24 * time.Hour),
		},
		Message: "登录成功",
	})

	logrus.WithFields(logrus.Fields{
		"user_id":  user.ID,
		"username": user.Username,
	}).Info("用户登录成功")
}

// RefreshToken 刷新token
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	// 从中间件获取用户信息
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, utils.ErrorResponse{
			Error: "无效的认证信息",
			Code:  "INVALID_AUTH",
		})
		return
	}

	username, _ := c.Get("username")
	email, _ := c.Get("email")

	// 生成新的token
	token, err := utils.GenerateJWT(
		userID.(uuid.UUID),
		username.(string),
		email.(string),
	)
	if err != nil {
		logrus.WithError(err).Error("JWT刷新失败")
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Error: "凭证刷新失败",
			Code:  "TOKEN_REFRESH_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, utils.SuccessResponse{
		Data: gin.H{
			"token":      token,
			"expires_at": time.Now().Add(24 * time.Hour),
		},
		Message: "Token刷新成功",
	})
}

// GetProfile 获取用户资料
func (h *AuthHandler) GetProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, utils.ErrorResponse{
			Error: "无效的认证信息",
			Code:  "INVALID_AUTH",
		})
		return
	}

	// 获取用户详细信息
	user, err := h.userService.GetUserByID(userID.(uuid.UUID))
	if err != nil {
		logrus.WithError(err).Error("获取用户信息失败")
		c.JSON(http.StatusNotFound, utils.ErrorResponse{
			Error: "用户不存在",
			Code:  "USER_NOT_FOUND",
		})
		return
	}

	// 获取用户偏好设置
	preference, err := h.userService.GetUserPreference(user.ID)
	if err != nil {
		logrus.WithError(err).Warn("获取用户偏好失败")
		// 不阻止请求，只是警告
	}

	responseData := gin.H{
		"id":         user.ID,
		"username":   user.Username,
		"email":      user.Email,
		"nickname":   user.Nickname,
		"avatar":     user.Avatar,
		"is_active":  user.IsActive,
		"created_at": user.CreatedAt,
		"updated_at": user.UpdatedAt,
	}

	if preference != nil {
		responseData["preferences"] = gin.H{
			"llm_model":      preference.LLMModel,
			"temperature":    preference.Temperature,
			"max_tokens":     preference.MaxTokens,
			"system_prompt":  preference.SystemPrompt,
			"context_window": preference.ContextWindow,
			"memory_enabled": preference.MemoryEnabled,
		}
	}

	c.JSON(http.StatusOK, utils.SuccessResponse{
		Data: responseData,
	})
}

// UpdateProfile 更新用户资料
func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, utils.ErrorResponse{
			Error: "无效的认证信息",
			Code:  "INVALID_AUTH",
		})
		return
	}

	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{
			Error:   "请求参数错误",
			Code:    "INVALID_REQUEST",
			Message: err.Error(),
		})
		return
	}

	// 清理输入数据
	updates := make(map[string]interface{})
	if req.Nickname != "" {
		updates["nickname"] = utils.SanitizeString(strings.TrimSpace(req.Nickname))
	}
	if req.Avatar != "" {
		updates["avatar"] = strings.TrimSpace(req.Avatar)
	}

	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{
			Error: "没有提供更新内容",
			Code:  "NO_UPDATE_DATA",
		})
		return
	}

	// 更新用户资料
	err := h.userService.UpdateProfile(userID.(uuid.UUID), updates)
	if err != nil {
		logrus.WithError(err).Error("更新用户资料失败")
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Error: "更新失败",
			Code:  "UPDATE_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, utils.SuccessResponse{
		Message: "资料更新成功",
		Data:    updates,
	})

	logrus.WithFields(logrus.Fields{
		"user_id": userID,
		"updates": updates,
	}).Info("用户资料更新成功")
}
