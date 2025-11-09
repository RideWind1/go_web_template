package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// HTTPClient HTTP客户端包装
type HTTPClient struct {
	client *http.Client
}

// NewHTTPClient 创建HTTP客户端
func NewHTTPClient(timeout time.Duration) *HTTPClient {
	return &HTTPClient{
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

// Get 发送GET请求
func (h *HTTPClient) Get(url string, headers map[string]string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	// 添加请求头
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	return h.client.Do(req)
}

// Post 发送POST请求
func (h *HTTPClient) Post(url string, body interface{}, headers map[string]string) (*http.Response, error) {
	var bodyReader *strings.Reader

	if body != nil {
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = strings.NewReader(string(bodyBytes))
	} else {
		bodyReader = strings.NewReader("")
	}

	req, err := http.NewRequest("POST", url, bodyReader)
	if err != nil {
		return nil, err
	}

	// 设置内容类型
	req.Header.Set("Content-Type", "application/json")

	// 添加请求头
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	return h.client.Do(req)
}

// ErrorResponse 错误响应结构
type ErrorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

// SuccessResponse 成功响应结构
type SuccessResponse struct {
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
	Code    string      `json:"code,omitempty"`
}

// LogError 日志记录错误
func LogError(err error, context string) {
	logrus.WithError(err).Error(context)
}

// LogInfo 日志记录信息
func LogInfo(message string, fields map[string]interface{}) {
	entry := logrus.WithFields(logrus.Fields(fields))
	entry.Info(message)
}

// SanitizeString 清理字符串，防止XSS
func SanitizeString(input string) string {
	// 简单的字符串清理
	replacer := strings.NewReplacer(
		"<", "&lt;",
		">", "&gt;",
		"&", "&amp;",
		"\"", "&quot;",
		"'", "&#39;",
	)
	return replacer.Replace(input)
}

// TruncateString 截取字符串
func TruncateString(input string, maxLength int) string {
	if len(input) <= maxLength {
		return input
	}
	return input[:maxLength] + "..."
}

// FormatError 格式化错误信息
func FormatError(err error, context string) string {
	if err == nil {
		return ""
	}
	return fmt.Sprintf("%s: %v", context, err)
}

// IsValidUUID 检查是否为有效的UUID
func IsValidUUID(id string) bool {
	// 简单的UUID格式检查
	return len(id) == 36 && strings.Count(id, "-") == 4
}
