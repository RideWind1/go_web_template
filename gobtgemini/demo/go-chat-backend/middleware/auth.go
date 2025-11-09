package middleware

import (
	"go-chat-backend/models"
	"go-chat-backend/utils"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// JWTClaims JWT声明结构
type JWTClaims struct {
	UserID   uuid.UUID `json:"user_id"`
	Username string    `json:"username"`
	Email    string    `json:"email"`
	jwt.RegisteredClaims
}

// JWTAuthMiddleware JWT认证中间件
func JWTAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		var tokenString string

		// 1. 优先从请求头中获取token (适用于普通HTTP API)
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
			tokenString = strings.TrimPrefix(authHeader, "Bearer ")
		}

		// 2. 如果请求头中没有token，则尝试从URL查询参数中获取 (适用于WebSocket)
		if tokenString == "" {
			tokenString = c.Query("token")
		}

		// 3. 如果两种方式都找不到token，则返回错误
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "请求未携带有效token",
				"code":  "MISSING_TOKEN",
			})
			c.Abort()
			return
		}

		// 解析token
		claims, err := utils.ParseJWT(tokenString)
		if err != nil {
			logrus.WithError(err).Error("JWT解析失败")
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Token无效或已过期",
				"code":  "INVALID_TOKEN",
			})
			c.Abort()
			return
		}

		// 将用户信息存储在上下文中
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("email", claims.Email)

		// 继续处理
		c.Next()
	}
}

// GetUserFromContext 从上下文中获取用户信息
func GetUserFromContext(c *gin.Context) (*models.User, error) {
	userID, exists := c.Get("user_id")
	if !exists {
		return nil, jwt.ErrTokenNotValidYet
	}

	username, _ := c.Get("username")
	email, _ := c.Get("email")

	user := &models.User{
		ID:       userID.(uuid.UUID),
		Username: username.(string),
		Email:    email.(string),
	}

	return user, nil
}

// LoggingMiddleware 日志中间件
func LoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 记录请求信息
		logrus.WithFields(logrus.Fields{
			"method":     c.Request.Method,
			"path":       c.Request.URL.Path,
			"ip":         c.ClientIP(),
			"user_agent": c.Request.UserAgent(),
		}).Info("请求开始")

		c.Next()

		// 记录响应信息
		logrus.WithFields(logrus.Fields{
			"method": c.Request.Method,
			"path":   c.Request.URL.Path,
			"status": c.Writer.Status(),
		}).Info("请求完成")
	}
}

// ErrorHandlerMiddleware 错误处理中间件
func ErrorHandlerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// 处理错误
		if len(c.Errors) > 0 {
			err := c.Errors.Last()
			logrus.WithError(err).Error("请求处理出错")

			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "内部服务器错误",
				"code":  "INTERNAL_ERROR",
			})
		}
	}
}
