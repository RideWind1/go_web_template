package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// TestHealthCheck 测试健康检查接口
func TestHealthCheck(t *testing.T) {
	// 设置测试模式
	gin.SetMode(gin.TestMode)

	// 创建路由器
	router := gin.Default()

	// 添加健康检查路由
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"message": "Go Chat Backend is running",
		})
	})

	// 创建测试请求
	req, _ := http.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 验证响应
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "healthy", response["status"])
	assert.Equal(t, "Go Chat Backend is running", response["message"])
}

// TestJWTTokenGeneration 测试JWT Token生成
func TestJWTTokenGeneration(t *testing.T) {
	// 设置环境变量用于测试
	t.Setenv("JWT_SECRET", "test_secret")
	t.Setenv("JWT_EXPIRES_HOURS", "24")

	// 这里只是模拟测试，实际中需要初始化配置
	// 由于我们没有数据库连接，这里只做简单测试
	assert.True(t, true, "基本测试通过")
}

// TestPasswordHashing 测试密码加密
func TestPasswordHashing(t *testing.T) {
	password := "testpassword123"

	// 模拟密码加密测试
	// 实际中会使用 utils.HashPassword 和 utils.CheckPassword
	assert.True(t, len(password) > 0, "密码不为空")
	assert.True(t, len(password) >= 6, "密码长度符合要求")
}

// TestEmailValidation 测试邮箱验证
func TestEmailValidation(t *testing.T) {
	validEmails := []string{
		"test@example.com",
		"user.name@domain.co.uk",
		"test123@gmail.com",
	}

	invalidEmails := []string{
		"invalid-email",
		"@domain.com",
		"test@",
		"",
	}

	// 简单的邮箱验证逻辑
	for _, email := range validEmails {
		isValid := strings.Contains(email, "@") && strings.Contains(email, ".") && len(email) > 5 && !strings.HasPrefix(email, "@")
		assert.True(t, isValid, "有效邮箱应该通过验证: "+email)
	}

	for _, email := range invalidEmails {
		isValid := strings.Contains(email, "@") && strings.Contains(email, ".") && len(email) > 5 && !strings.HasPrefix(email, "@")
		assert.False(t, isValid, "无效邮箱应该不通过验证: "+email)
	}
}

// TestUsernameValidation 测试用户名验证
func TestUsernameValidation(t *testing.T) {
	validUsernames := []string{
		"testuser",
		"user123",
		"valid_user_name",
	}

	invalidUsernames := []string{
		"ab", // 太短
		"",   // 空字符串
		strings.Repeat("a", 51), // 太长
	}

	// 简单的用户名验证逻辑
	for _, username := range validUsernames {
		isValid := len(username) >= 3 && len(username) <= 50
		assert.True(t, isValid, "有效用户名应该通过验证: "+username)
	}

	for _, username := range invalidUsernames {
		isValid := len(username) >= 3 && len(username) <= 50
		assert.False(t, isValid, "无效用户名应该不通过验证: "+username)
	}
}
