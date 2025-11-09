package utils

import (
	"go-chat-backend/config"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// JWTClaims JWT声明结构
type JWTClaims struct {
	UserID   uuid.UUID `json:"user_id"`
	Username string    `json:"username"`
	Email    string    `json:"email"`
	jwt.RegisteredClaims
}

// GenerateJWT 生成JWT token
func GenerateJWT(userID uuid.UUID, username, email string) (string, error) {
	cfg := config.Get()
	now := time.Now()

	claims := &JWTClaims{
		UserID:   userID,
		Username: username,
		Email:    email,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(cfg.JWTExpires) * time.Hour)),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "go-chat-backend",
			Subject:   userID.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.JWTSecret))
}

// ParseJWT 解析JWT token
func ParseJWT(tokenString string) (*JWTClaims, error) {
	cfg := config.Get()

	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(cfg.JWTSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, jwt.ErrTokenNotValidYet
}

// HashPassword 加密密码
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPassword 验证密码
func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// GenerateRefreshToken 生成刷新token
func GenerateRefreshToken() string {
	return uuid.New().String()
}

// ValidateEmail 验证邮箱格式
func ValidateEmail(email string) bool {
	// 简单的邮箱验证
	return len(email) > 5 &&
		len(email) <= 254 &&
		contains(email, "@") &&
		contains(email, ".")
}

// ValidateUsername 验证用户名格式
func ValidateUsername(username string) bool {
	return len(username) >= 3 && len(username) <= 50
}

// ValidatePassword 验证密码强度
func ValidatePassword(password string) bool {
	return len(password) >= 6 && len(password) <= 128
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
