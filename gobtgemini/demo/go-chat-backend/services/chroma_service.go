package services

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"go-chat-backend/config"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type EmbeddingRequest struct {
	Input string `json:"input"`
	Model string `json:"model"`
}

type Collection struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	// API 还返回其他字段，如 metadata, tenant 等，但我们这里不需要
}

type EmbeddingResponse struct {
	Data []struct {
		Embedding []float64 `json:"embedding"`
	} `json:"data"`
}

type QueryRequestWithEmbeddings struct {
	QueryEmbeddings [][]float64            `json:"query_embeddings"`
	NResults        int                    `json:"n_results,omitempty"`
	Where           map[string]interface{} `json:"where,omitempty"`
}

// ChromaService Chroma向量数据库服务
type ChromaService struct {
	baseURL      string
	collection   string
	httpClient   *http.Client
	collectionId string
}

// NewChromaService 创建Chroma服务
func NewChromaService() (*ChromaService, error) {
	cfg := config.Get()
	baseURL := fmt.Sprintf("http://%s:%s", cfg.ChromaHost, cfg.ChromaPort)

	service := &ChromaService{
		baseURL:    baseURL,
		collection: cfg.ChromaCollection,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}

	// 修改点2：在创建实例后，立即调用初始化函数
	if err := service.InitCollection(); err != nil {
		// 如果初始化失败（例如ChromaDB服务连不上），则返回错误
		return nil, fmt.Errorf("ChromaService 初始化失败: %w", err)
	}

	return service, nil // 修改点3：返回创建好的 service 和 nil error
}

// Document Chroma文档结构
type Document struct {
	ID       string            `json:"id"`
	Content  string            `json:"document"`
	Metadata map[string]string `json:"metadata"`
}

// QueryResult 查询结果结构
type QueryResult struct {
	IDs       [][]string            `json:"ids"`
	Documents [][]string            `json:"documents"`
	Metadatas [][]map[string]string `json:"metadatas"`
	Distances [][]float64           `json:"distances"`
}

// AddRequest 添加文档请求结构
type AddRequest struct {
	IDs       []string                 `json:"ids"`
	Documents []string                 `json:"documents"`
	Metadatas []map[string]interface{} `json:"metadatas,omitempty"`
}

// QueryRequest 查询请求结构
type QueryRequest struct {
	QueryTexts []string               `json:"query_texts"`
	NResults   int                    `json:"n_results,omitempty"`
	Where      map[string]interface{} `json:"where,omitempty"`
}

// InitCollection 初始化集合
func (s *ChromaService) InitCollection() error {
	// 检查集合是否存在
	exists, err := s.collectionExists()
	if err != nil {
		return fmt.Errorf("检查集合存在性失败: %w", err)
	}

	if !exists {
		// 创建集合
		if err := s.createCollection(); err != nil {
			return fmt.Errorf("创建集合失败: %w", err)
		}

		time.Sleep(10 * time.Second)
		logrus.Info("成功创建Chroma集合: ", s.collection)
	} else {
		logrus.Info("Chroma集合已存在: ", s.collection)
	}
	coID, err1 := s.getCollectionIDByName()
	if err1 != nil {
		return fmt.Errorf("获取集合ID失败：%w", err)
	}
	s.collectionId = coID

	return nil
}

// 获取集合的id
func (s *ChromaService) getCollectionIDByName() (string, error) {
	url := fmt.Sprintf("%s/api/v1/collections", s.baseURL)

	// 2. 发送 HTTP GET 请求
	fmt.Printf("正在向 %s 发送请求以获取所有集合...\n", url)
	resp, err := s.httpClient.Get(url)
	if err != nil {
		return "", fmt.Errorf("发送 HTTP GET 请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 3. 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("获取集合列表失败，服务器返回状态码: %d", resp.StatusCode)
	}

	// 4. 读取并解析 JSON 响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应体失败: %w", err)
	}

	var collections []Collection
	if err := json.Unmarshal(body, &collections); err != nil {
		return "", fmt.Errorf("解析 JSON 响应失败: %w", err)
	}

	fmt.Printf("成功获取到 %d 个集合的信息。\n", len(collections))

	// 5. 遍历返回的集合列表，查找名称匹配的集合
	for _, c := range collections {
		if c.Name == s.collection {
			fmt.Printf("找到了！集合 '%s' 对应的 ID 是: %s\n", s.collection, c.ID)
			s.collectionId = c.ID
			return c.ID, nil // 找到后，返回 ID 和 nil 错误
		}
	}

	// 6. 如果循环结束仍未找到，返回错误
	return "", fmt.Errorf("未找到名为 '%s' 的集合", s.collection)
}

// collectionExists 检查集合是否存在
func (s *ChromaService) collectionExists() (bool, error) {
	url := fmt.Sprintf("%s/api/v1/collections/%s", s.baseURL, s.collection)
	resp, err := s.httpClient.Get(url)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return true, nil // 200 OK: 集合已存在
	}
	// 按照 RESTful 最佳实践，也应该处理 404 Not Found 的情况
	if resp.StatusCode == http.StatusNotFound {
		return false, nil // 404 Not Found: 集合不存在，这是正常情况
	}

	// 对于 ChromaDB 返回 500 的特殊情况，我们也将其视为“不存在”，但打印一条警告日志
	if resp.StatusCode == http.StatusInternalServerError {
		logrus.Warnf("检查Chroma集合'%s'是否存在时收到500错误，暂时将其视为不存在", s.collection)
		return false, nil
	}

	// 其他所有非预期的状态码都应被视为一个真正的错误
	return false, fmt.Errorf("检查Chroma集合存在性时收到意外的状态码: %d", resp.StatusCode)

}

// createCollection 创建集合
func (s *ChromaService) createCollection() error {
	url := fmt.Sprintf("%s/api/v1/collections", s.baseURL)
	requestBody := map[string]string{
		"name": s.collection,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return err
	}

	resp, err := s.httpClient.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("创建集合失败，状态码: %d", resp.StatusCode)
	}

	return nil
}

// AddMemory 添加记忆
func (s *ChromaService) AddMemory(userID uuid.UUID, content string, messageType string) error {
	// 生成文档ID
	docID := fmt.Sprintf("%s_%s_%d", userID.String(), messageType, time.Now().Unix())

	// 构建元数据
	metadata := map[string]interface{}{
		"user_id":      userID.String(), // 保持为字符串
		"message_type": messageType,
		"timestamp":    time.Now().Unix(), // 可以直接使用数字类型
	}

	// 添加文档
	return s.addDocument(docID, content, metadata)
}

// addDocument 添加文档
func (s *ChromaService) addDocument(id, content string, metadata map[string]interface{}) error {
	url := fmt.Sprintf("%s/api/v1/collections/%s/add", s.baseURL, s.collectionId)

	requestBody := AddRequest{
		IDs:       []string{id},
		Documents: []string{content},
		Metadatas: []map[string]interface{}{metadata},
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("序列化请求失败: %w", err)
	}

	resp, err := s.httpClient.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("添加文档失败，状态码: %d", resp.StatusCode)
	}

	logrus.WithFields(logrus.Fields{
		"doc_id":   id,
		"content":  content[:min(50, len(content))] + "...",
		"metadata": metadata,
	}).Debug("成功添加文档到Chroma")

	return nil
}

// CreateEmbedding 为给定的文本创建向量
func (s *ChromaService) CreateEmbedding(text string) ([]float64, error) {
	embeddingServiceURL := "http://embedding-service/embed"

	requestBody := map[string]interface{}{
		"inputs":    text,
		"normalize": true, // 推荐开启，使向量长度归一化
		"truncate":  true, // 自动截断超长文本
	}
	jsonData, _ := json.Marshal(requestBody)

	req, err := http.NewRequest("POST", embeddingServiceURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("创建 embedding 请求失败: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求 embedding 服务失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// 可以添加更多日志来读取 resp.Body 的错误信息
		return nil, fmt.Errorf("embedding 服务返回错误状态码: %d", resp.StatusCode)
	}

	// ⭐ 修改点 3：响应体是一个直接的二维数组
	var embeddings [][]float64
	if err := json.NewDecoder(resp.Body).Decode(&embeddings); err != nil {
		return nil, fmt.Errorf("解析 embedding 响应失败: %w", err)
	}

	if len(embeddings) == 0 || len(embeddings[0]) == 0 {
		return nil, errors.New("从 embedding 服务收到的向量为空")
	}

	return embeddings[0], nil
}

// SearchMemory 搜索相关记忆
func (s *ChromaService) SearchMemory(userID uuid.UUID, query string, limit int) ([]string, error) {
	if limit <= 0 || limit > 20 {
		limit = 5 // 默认返回5个结果
	}
	queryEmbedding, err := s.CreateEmbedding(query)
	if err != nil {
		// 如果获取 embedding 失败，就无法继续查询
		return nil, fmt.Errorf("创建查询向量失败: %w", err)
	}

	url := fmt.Sprintf("%s/api/v1/collections/%s/query", s.baseURL, s.collectionId)

	// 构建查询请求
	requestBody := QueryRequestWithEmbeddings{
		QueryEmbeddings: [][]float64{queryEmbedding},
		NResults:        limit,
		Where: map[string]interface{}{
			"user_id": map[string]string{
				"$eq": userID.String(),
			},
		},
	}
	//fmt.Println("开始搜索记忆")

	jsonData, err := json.Marshal(requestBody)

	if err != nil {
		return nil, fmt.Errorf("序列化查询请求失败: %w", err)
	}
	logrus.Infof("发送给 Chroma 的查询请求: %s", string(jsonData))
	resp, err := s.httpClient.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("发送查询请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("查询失败，状态码: %d", resp.StatusCode)
	}

	var result QueryResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("解析查询结果失败: %w", err)
	}

	// 提取相关文档
	var documents []string
	if len(result.Documents) > 0 {
		documents = result.Documents[0] // 取第一个查询的结果
	}
	//fmt.Println("搜索记忆结束")
	logrus.WithFields(logrus.Fields{
		"user_id": userID.String(),
		"query":   query,
		"results": len(documents),
	}).Debug("成功搜索Chroma记忆")

	return documents, nil
}

// ClearUserMemory 清空用户记忆
func (s *ChromaService) ClearUserMemory(userID uuid.UUID) error {
	// 注意：Chroma不支持按元数据删除，这里只是示例实现
	// 实际使用中可能需要重新创建集合或使用其他策略
	logrus.WithField("user_id", userID.String()).Info("请求清空用户记忆（暂不支持）")
	return nil
}

// GetMemoryStats 获取记忆统计信息
func (s *ChromaService) GetMemoryStats(userID uuid.UUID) (map[string]interface{}, error) {
	// 这里返回一些模拟数据，实际使用中可以通过查询获取真实统计
	stats := map[string]interface{}{
		"user_id":         userID.String(),
		"total_memories":  0,
		"last_updated":    time.Now().Format(time.RFC3339),
		"collection_name": s.collection,
	}

	return stats, nil
}

// HealthCheck 健康检查
func (s *ChromaService) HealthCheck() error {
	url := fmt.Sprintf("%s/api/v1/heartbeat", s.baseURL)
	resp, err := s.httpClient.Get(url)
	if err != nil {
		return fmt.Errorf("Chroma连接失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Chroma服务不可用，状态码: %d", resp.StatusCode)
	}

	return nil
}

// min 辅助函数
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
