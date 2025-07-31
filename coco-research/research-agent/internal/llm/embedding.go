package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/coco-ai/research-agent/internal/config"
	"github.com/sirupsen/logrus"
)

// EmbeddingClient 嵌入模型客户端接口
type EmbeddingClient interface {
	GenerateEmbedding(text string) ([]float32, error)
	GenerateEmbeddings(texts []string) ([][]float32, error)
}

// HuggingFaceEmbeddingClient HuggingFace嵌入模型客户端
type HuggingFaceEmbeddingClient struct {
	model     string
	apiKey    string
	baseURL   string
	maxLength int
	dimension int
	client    *http.Client
	logger    *logrus.Entry
}

// HuggingFaceEmbeddingRequest HuggingFace嵌入请求
type HuggingFaceEmbeddingRequest struct {
	Inputs string `json:"inputs"`
}

// HuggingFaceEmbeddingResponse HuggingFace嵌入响应
type HuggingFaceEmbeddingResponse [][]float32

// NewHuggingFaceEmbeddingClient 创建HuggingFace嵌入客户端
func NewHuggingFaceEmbeddingClient(cfg *config.EmbeddingConfig) *HuggingFaceEmbeddingClient {
	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://api-inference.huggingface.co"
	}

	return &HuggingFaceEmbeddingClient{
		model:     cfg.Model,
		apiKey:    cfg.APIKey,
		baseURL:   cfg.BaseURL,
		maxLength: cfg.MaxLength,
		dimension: cfg.Dimension,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger: logrus.WithField("component", "huggingface_embedding"),
	}
}

// GenerateEmbedding 生成单个文本的嵌入向量
func (h *HuggingFaceEmbeddingClient) GenerateEmbedding(text string) ([]float32, error) {
	// 截断文本到最大长度
	if len(text) > h.maxLength {
		text = text[:h.maxLength]
	}

	// 准备请求
	request := HuggingFaceEmbeddingRequest{
		Inputs: text,
	}

	requestBytes, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// 发送请求
	url := fmt.Sprintf("%s/models/%s", h.baseURL, h.model)
	req, err := http.NewRequest("POST", url, bytes.NewReader(requestBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if h.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+h.apiKey)
	}

	resp, err := h.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// 解析响应
	var response HuggingFaceEmbeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(response) == 0 || len(response[0]) == 0 {
		return nil, fmt.Errorf("empty embedding response")
	}

	h.logger.Debug("Generated embedding", "text_length", len(text), "embedding_dimension", len(response[0]))
	return response[0], nil
}

// GenerateEmbeddings 批量生成文本的嵌入向量
func (h *HuggingFaceEmbeddingClient) GenerateEmbeddings(texts []string) ([][]float32, error) {
	embeddings := make([][]float32, len(texts))
	
	for i, text := range texts {
		embedding, err := h.GenerateEmbedding(text)
		if err != nil {
			return nil, fmt.Errorf("failed to generate embedding for text %d: %w", i, err)
		}
		embeddings[i] = embedding
	}

	return embeddings, nil
}

// OpenAIEmbeddingClient OpenAI嵌入模型客户端
type OpenAIEmbeddingClient struct {
	model     string
	apiKey    string
	baseURL   string
	dimension int
	client    *http.Client
	logger    *logrus.Entry
}

// OpenAIEmbeddingRequest OpenAI嵌入请求
type OpenAIEmbeddingRequest struct {
	Input string `json:"input"`
	Model string `json:"model"`
}

// OpenAIEmbeddingResponse OpenAI嵌入响应
type OpenAIEmbeddingResponse struct {
	Data []struct {
		Embedding []float32 `json:"embedding"`
	} `json:"data"`
}

// NewOpenAIEmbeddingClient 创建OpenAI嵌入客户端
func NewOpenAIEmbeddingClient(cfg *config.EmbeddingConfig) *OpenAIEmbeddingClient {
	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://api.openai.com/v1"
	}
	if cfg.Model == "" {
		cfg.Model = "text-embedding-ada-002"
	}

	return &OpenAIEmbeddingClient{
		model:     cfg.Model,
		apiKey:    cfg.APIKey,
		baseURL:   cfg.BaseURL,
		dimension: cfg.Dimension,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger: logrus.WithField("component", "openai_embedding"),
	}
}

// GenerateEmbedding 生成单个文本的嵌入向量
func (o *OpenAIEmbeddingClient) GenerateEmbedding(text string) ([]float32, error) {
	request := OpenAIEmbeddingRequest{
		Input: text,
		Model: o.model,
	}

	requestBytes, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/embeddings", o.baseURL)
	req, err := http.NewRequest("POST", url, bytes.NewReader(requestBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+o.apiKey)

	resp, err := o.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var response OpenAIEmbeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(response.Data) == 0 || len(response.Data[0].Embedding) == 0 {
		return nil, fmt.Errorf("empty embedding response")
	}

	o.logger.Debug("Generated embedding", "text_length", len(text), "embedding_dimension", len(response.Data[0].Embedding))
	return response.Data[0].Embedding, nil
}

// GenerateEmbeddings 批量生成文本的嵌入向量
func (o *OpenAIEmbeddingClient) GenerateEmbeddings(texts []string) ([][]float32, error) {
	embeddings := make([][]float32, len(texts))
	
	for i, text := range texts {
		embedding, err := o.GenerateEmbedding(text)
		if err != nil {
			return nil, fmt.Errorf("failed to generate embedding for text %d: %w", i, err)
		}
		embeddings[i] = embedding
	}

	return embeddings, nil
}

// NewEmbeddingClient 根据配置创建嵌入客户端
func NewEmbeddingClient(cfg *config.EmbeddingConfig) (EmbeddingClient, error) {
	switch strings.ToLower(cfg.Provider) {
	case "huggingface":
		return NewHuggingFaceEmbeddingClient(cfg), nil
	case "openai":
		return NewOpenAIEmbeddingClient(cfg), nil
	default:
		return nil, fmt.Errorf("unsupported embedding provider: %s", cfg.Provider)
	}
}