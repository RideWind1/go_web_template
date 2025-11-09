package services

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"go-chat-backend/config"
	"go-chat-backend/models"
	"go-chat-backend/utils"
	"io"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

// LLMService 大语言模型服务
type LLMService struct {
	httpClient *utils.HTTPClient
}

// NewLLMService 创建大语言模型服务
func NewLLMService() *LLMService {
	return &LLMService{
		httpClient: utils.NewHTTPClient(30 * time.Second),
	}
}

// --- 新的 Gemini API 结构体 ---

// GeminiPart 对应 Gemini API 请求体中的 "parts"
type GeminiPart struct {
	Text string `json:"text"`
}

// GeminiContent 对应 Gemini API 请求体中的 "contents"
type GeminiContent struct {
	Parts []GeminiPart `json:"parts"`
	Role  string       `json:"role,omitempty"`
}

// GeminiChatRequest 对应 Gemini API 的完整请求体
type GeminiChatRequest struct {
	Contents []GeminiContent `json:"contents"`
	// Gemini 的 Temperature/MaxTokens 等参数在 GenerationConfig 中，这里简化
}

// --- 新的 Gemini API 响应结构体 ---

// GeminiResponseCandidate 对应 Gemini 响应中的 "candidates"
type GeminiResponseCandidate struct {
	Content      GeminiContent `json:"content"`
	FinishReason string        `json:"finishReason"`
}

// GeminiUsageMetadata 对应 Gemini 响应中的 "usageMetadata"
type GeminiUsageMetadata struct {
	PromptTokenCount     int `json:"promptTokenCount"`
	CandidatesTokenCount int `json:"candidatesTokenCount"`
	TotalTokenCount      int `json:"totalTokenCount"`
}

// GeminiChatResponse 对应 Gemini API 的完整响应体
type GeminiChatResponse struct {
	Candidates    []GeminiResponseCandidate `json:"candidates"`
	UsageMetadata GeminiUsageMetadata       `json:"usageMetadata"`
	Error         *struct {
		Message string `json:"message"`
		Code    int    `json:"code"`
		Status  string `json:"status"`
	} `json:"error,omitempty"`
}

// GenerateResponse 生成回复
func (s *LLMService) GenerateResponse(messages []models.ChatMessage, userPreference *models.UserPreference) (string, error) {
	if len(messages) == 0 {
		// 返回一个清晰的、源自我们自己代码的错误
		return "", errors.New("传入的消息列表为空，无法生成回复")
	}
	//fmt.Println(messages)

	cfg := config.Get()

	// 检查API配置
	if cfg.LLMAPIURL == "" {
		return "", errors.New("未配置LLM API URL")
	}

	// --- 1. 修改：构建 Gemini 格式的请求体 ---
	geminiContents := make([]GeminiContent, 0, len(messages)+1)

	// 添加系统提示 (Gemini 推荐将系统提示放在第一个 User 角色的内容里)
	systemPrompt := "你是一个智能助手，请提供准确、有用的信息和帮助。请用中文回复。"
	if userPreference.SystemPrompt != "" {
		systemPrompt = userPreference.SystemPrompt
	}

	// 将系统提示和第一条用户消息合并
	if len(messages) > 0 {
		firstUserMessage := messages[0]
		// 确保第一条消息是 user
		if firstUserMessage.Role == "user" {
			fullPrompt := systemPrompt + "\n\n" + firstUserMessage.Content
			geminiContents = append(geminiContents, GeminiContent{
				Role:  "user",
				Parts: []GeminiPart{{Text: fullPrompt}},
			})
			// 从第二条消息开始处理
			messages = messages[1:]
		}
	}

	// 添加剩余的历史消息
	for _, msg := range messages {
		geminiContents = append(geminiContents, GeminiContent{
			Role:  msg.Role,
			Parts: []GeminiPart{{Text: msg.Content}},
		})
	}

	// 构建请求
	request := GeminiChatRequest{
		Contents: geminiContents,
	}

	// 序列化请求
	requestBody, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("请求序列化失败: %w", err)
	}

	// ======================= 调试步骤 1：打印将要发送的 JSON =======================
	//fmt.Println("即将发送的 JSON Body:", string(requestBody))
	// =========================================================================

	// --- 2. 修改：URL 和认证头 ---
	// 注意：Gemini 的模型名称是 URL 的一部分，需要确保 cfg.LLMAPIURL 已经包含模型名称
	// 例如: "https://.../v1beta/models/gemini-pro:generateContent"
	apiURL := cfg.LLMAPIURL

	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(requestBody))
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if cfg.LLMAPIKey != "" {
		// 使用 X-goog-api-key 而不是 Authorization: Bearer
		req.Header.Set("X-goog-api-key", cfg.LLMAPIKey)
	}

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	// fmt.Println("这是请求：")
	// fmt.Println(req)
	// fmt.Println("这是回复：")
	// fmt.Println(resp)

	if err != nil {
		logrus.WithError(err).Error("LLM API请求失败")
		return "", fmt.Errorf("LLM API请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 增加对非 200 OK 状态码的处理
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)

		// ======================= 调试步骤 2：打印详细的错误响应体 =======================
		// 这条日志现在至关重要，它会告诉我们失败的具体原因
		errorDetail := string(bodyBytes)
		fmt.Printf("API 返回错误，状态码: %d, 错误详情: %s\n", resp.StatusCode, errorDetail)
		// =========================================================================

		logrus.Errorf("LLM API 返回错误状态码: %d, 响应: %s", resp.StatusCode, string(bodyBytes))
		return "", fmt.Errorf("LLM API 返回错误状态码: %d", resp.StatusCode)
	}

	// --- 3. 修改：解析 Gemini 格式的响应 ---
	var response GeminiChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("响应解析失败: %w", err)
	}

	// 检查是否有 API 错误
	if response.Error != nil {
		logrus.WithFields(logrus.Fields{
			"error_message": response.Error.Message,
			"error_status":  response.Error.Status,
		}).Error("LLM API返回错误")
		return "", fmt.Errorf("LLM API错误: %s", response.Error.Message)
	}

	// 检查是否有回复
	if len(response.Candidates) == 0 || len(response.Candidates[0].Content.Parts) == 0 {
		return "", errors.New("LLM API未返回任何回复")
	}

	content := response.Candidates[0].Content.Parts[0].Text
	if content == "" {
		return "", errors.New("LLM API返回的回复为空")
	}

	// 记录token使用情况
	logrus.WithFields(logrus.Fields{
		"prompt_tokens":     response.UsageMetadata.PromptTokenCount,
		"completion_tokens": response.UsageMetadata.CandidatesTokenCount,
		"total_tokens":      response.UsageMetadata.TotalTokenCount,
		"model":             "gemini", // 模型名称在URL中，这里可以写死或从URL解析
	}).Info("LLM API调用成功")

	return content, nil
}

// ValidateAPIConfig 验证API配置
func (s *LLMService) ValidateAPIConfig() error {
	cfg := config.Get()

	if cfg.LLMAPIURL == "" {
		return errors.New("LLM_API_URL 未配置")
	}

	// 可以添加更多验证逻辑，比如发送测试请求

	return nil
}

// GetSupportedModels 获取支持的模型列表
func (s *LLMService) GetSupportedModels() []string {
	// 返回常见的模型列表，可以根据实际API支持情况调整
	return []string{
		"gpt-3.5-turbo",
		"gpt-4",
		"gpt-4-turbo-preview",
		"claude-3-sonnet-20240229",
		"claude-3-opus-20240229",
		"gemini-pro",
		"text-davinci-003",
	}
}
