package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/coco-ai/research-agent/internal/agent"
	"github.com/coco-ai/research-agent/pkg/logger"
	"github.com/sirupsen/logrus"
)

// Client LLM客户端接口
type Client interface {
	Chat(ctx context.Context, messages []agent.Message, options *agent.LLMOptions) (string, error)
	GenerateEmbedding(ctx context.Context, text string) ([]float32, error)
}

// OpenAIClient OpenAI客户端
type OpenAIClient struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
	logger     *logrus.Entry
}

// ClaudeClient Claude客户端
type ClaudeClient struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
	logger     *logrus.Entry
}

// DeepSeekClient DeepSeek客户端
type DeepSeekClient struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
	logger     *logrus.Entry
}

// OpenAIRequest OpenAI请求结构
type OpenAIRequest struct {
	Model       string        `json:"model"`
	Messages    []OpenAIMessage `json:"messages"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
	Temperature float64       `json:"temperature,omitempty"`
	Stream      bool          `json:"stream,omitempty"`
}

// OpenAIMessage OpenAI消息结构
type OpenAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// OpenAIResponse OpenAI响应结构
type OpenAIResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// ClaudeRequest Claude请求结构
type ClaudeRequest struct {
	Model       string        `json:"model"`
	Messages    []ClaudeMessage `json:"messages"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
	Temperature float64       `json:"temperature,omitempty"`
}

// ClaudeMessage Claude消息结构
type ClaudeMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ClaudeResponse Claude响应结构
type ClaudeResponse struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Role    string `json:"role"`
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Model     string `json:"model"`
	StopReason string `json:"stop_reason"`
	StopSequence string `json:"stop_sequence"`
	Usage struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

// DeepSeekRequest DeepSeek请求结构
type DeepSeekRequest struct {
	Model       string        `json:"model"`
	Messages    []DeepSeekMessage `json:"messages"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
	Temperature float64       `json:"temperature,omitempty"`
	Stream      bool          `json:"stream,omitempty"`
}

// DeepSeekMessage DeepSeek消息结构
type DeepSeekMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// DeepSeekResponse DeepSeek响应结构
type DeepSeekResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// NewOpenAIClient 创建OpenAI客户端
func NewOpenAIClient(apiKey string) *OpenAIClient {
	return &OpenAIClient{
		apiKey: apiKey,
		baseURL: "https://api.openai.com/v1",
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
		logger: logger.WithField("component", "openai_client"),
	}
}

// NewClaudeClient 创建Claude客户端
func NewClaudeClient(apiKey string) *ClaudeClient {
	return &ClaudeClient{
		apiKey: apiKey,
		baseURL: "https://api.anthropic.com/v1",
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
		logger: logger.WithField("component", "claude_client"),
	}
}

// NewDeepSeekClient 创建DeepSeek客户端
func NewDeepSeekClient(apiKey string) *DeepSeekClient {
	return &DeepSeekClient{
		apiKey: apiKey,
		baseURL: "https://api.deepseek.com/v1",
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
		logger: logger.WithField("component", "deepseek_client"),
	}
}

// Chat OpenAI聊天
func (c *OpenAIClient) Chat(ctx context.Context, messages []agent.Message, options *agent.LLMOptions) (string, error) {
	// 转换消息格式
	openAIMessages := make([]OpenAIMessage, len(messages))
	for i, msg := range messages {
		openAIMessages[i] = OpenAIMessage{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}

	// 构建请求
	request := OpenAIRequest{
		Model:       options.Model,
		Messages:    openAIMessages,
		MaxTokens:   options.MaxTokens,
		Temperature: options.Temperature,
	}

	// 发送请求
	response, err := c.sendRequest(ctx, "/chat/completions", request)
	if err != nil {
		return "", fmt.Errorf("OpenAI API request failed: %w", err)
	}

	// 解析响应
	var openAIResponse OpenAIResponse
	err = json.Unmarshal(response, &openAIResponse)
	if err != nil {
		return "", fmt.Errorf("failed to parse OpenAI response: %w", err)
	}

	if len(openAIResponse.Choices) == 0 {
		return "", fmt.Errorf("no choices in OpenAI response")
	}

	c.logger.Info("OpenAI chat completed", 
		"model", options.Model,
		"tokens", openAIResponse.Usage.TotalTokens)

	return openAIResponse.Choices[0].Message.Content, nil
}

// GenerateEmbedding OpenAI生成嵌入
func (c *OpenAIClient) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	request := map[string]interface{}{
		"model": "text-embedding-ada-002",
		"input": text,
	}

	response, err := c.sendRequest(ctx, "/embeddings", request)
	if err != nil {
		return nil, fmt.Errorf("OpenAI embedding request failed: %w", err)
	}

	var embeddingResponse struct {
		Data []struct {
			Embedding []float64 `json:"embedding"`
		} `json:"data"`
	}

	err = json.Unmarshal(response, &embeddingResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to parse embedding response: %w", err)
	}

	if len(embeddingResponse.Data) == 0 {
		return nil, fmt.Errorf("no embedding data in response")
	}

	// 转换为float32
	embedding := make([]float32, len(embeddingResponse.Data[0].Embedding))
	for i, v := range embeddingResponse.Data[0].Embedding {
		embedding[i] = float32(v)
	}

	return embedding, nil
}

// sendRequest 发送HTTP请求
func (c *OpenAIClient) sendRequest(ctx context.Context, endpoint string, request interface{}) ([]byte, error) {
	// 序列化请求
	requestBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// 创建HTTP请求
	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+endpoint, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	// 发送请求
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// 检查HTTP状态码
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(responseBody))
	}

	return responseBody, nil
}

// Chat Claude聊天
func (c *ClaudeClient) Chat(ctx context.Context, messages []agent.Message, options *agent.LLMOptions) (string, error) {
	// 转换消息格式
	claudeMessages := make([]ClaudeMessage, len(messages))
	for i, msg := range messages {
		claudeMessages[i] = ClaudeMessage{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}

	// 构建请求
	request := ClaudeRequest{
		Model:       options.Model,
		Messages:    claudeMessages,
		MaxTokens:   options.MaxTokens,
		Temperature: options.Temperature,
	}

	// 发送请求
	response, err := c.sendRequest(ctx, "/messages", request)
	if err != nil {
		return "", fmt.Errorf("Claude API request failed: %w", err)
	}

	// 解析响应
	var claudeResponse ClaudeResponse
	err = json.Unmarshal(response, &claudeResponse)
	if err != nil {
		return "", fmt.Errorf("failed to parse Claude response: %w", err)
	}

	if len(claudeResponse.Content) == 0 {
		return "", fmt.Errorf("no content in Claude response")
	}

	c.logger.Info("Claude chat completed", 
		"model", options.Model,
		"tokens", claudeResponse.Usage.OutputTokens)

	return claudeResponse.Content[0].Text, nil
}

// GenerateEmbedding Claude生成嵌入
func (c *ClaudeClient) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	request := map[string]interface{}{
		"model": "claude-3-sonnet-20240229",
		"input": text,
	}

	response, err := c.sendRequest(ctx, "/embeddings", request)
	if err != nil {
		return nil, fmt.Errorf("Claude embedding request failed: %w", err)
	}

	var embeddingResponse struct {
		Data []struct {
			Embedding []float64 `json:"embedding"`
		} `json:"data"`
	}

	err = json.Unmarshal(response, &embeddingResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to parse embedding response: %w", err)
	}

	if len(embeddingResponse.Data) == 0 {
		return nil, fmt.Errorf("no embedding data in response")
	}

	// 转换为float32
	embedding := make([]float32, len(embeddingResponse.Data[0].Embedding))
	for i, v := range embeddingResponse.Data[0].Embedding {
		embedding[i] = float32(v)
	}

	return embedding, nil
}

// sendRequest Claude发送HTTP请求
func (c *ClaudeClient) sendRequest(ctx context.Context, endpoint string, request interface{}) ([]byte, error) {
	// 序列化请求
	requestBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// 创建HTTP请求
	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+endpoint, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	// 发送请求
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// 检查HTTP状态码
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(responseBody))
	}

	return responseBody, nil
}

// Chat DeepSeek聊天
func (c *DeepSeekClient) Chat(ctx context.Context, messages []agent.Message, options *agent.LLMOptions) (string, error) {
	// 转换消息格式
	deepSeekMessages := make([]DeepSeekMessage, len(messages))
	for i, msg := range messages {
		deepSeekMessages[i] = DeepSeekMessage{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}

	// 构建请求
	request := DeepSeekRequest{
		Model:       options.Model,
		Messages:    deepSeekMessages,
		MaxTokens:   options.MaxTokens,
		Temperature: options.Temperature,
	}

	// 发送请求
	response, err := c.sendRequest(ctx, "/chat/completions", request)
	if err != nil {
		return "", fmt.Errorf("DeepSeek API request failed: %w", err)
	}

	// 解析响应
	var deepSeekResponse DeepSeekResponse
	err = json.Unmarshal(response, &deepSeekResponse)
	if err != nil {
		return "", fmt.Errorf("failed to parse DeepSeek response: %w", err)
	}

	if len(deepSeekResponse.Choices) == 0 {
		return "", fmt.Errorf("no choices in DeepSeek response")
	}

	c.logger.Info("DeepSeek chat completed", 
		"model", options.Model,
		"tokens", deepSeekResponse.Usage.TotalTokens)

	return deepSeekResponse.Choices[0].Message.Content, nil
}

// GenerateEmbedding DeepSeek生成嵌入
func (c *DeepSeekClient) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	request := map[string]interface{}{
		"model": "deepseek-embedding",
		"input": text,
	}

	response, err := c.sendRequest(ctx, "/embeddings", request)
	if err != nil {
		return nil, fmt.Errorf("DeepSeek embedding request failed: %w", err)
	}

	var embeddingResponse struct {
		Data []struct {
			Embedding []float64 `json:"embedding"`
		} `json:"data"`
	}

	err = json.Unmarshal(response, &embeddingResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to parse embedding response: %w", err)
	}

	if len(embeddingResponse.Data) == 0 {
		return nil, fmt.Errorf("no embedding data in response")
	}

	// 转换为float32
	embedding := make([]float32, len(embeddingResponse.Data[0].Embedding))
	for i, v := range embeddingResponse.Data[0].Embedding {
		embedding[i] = float32(v)
	}

	return embedding, nil
}

// sendRequest DeepSeek发送HTTP请求
func (c *DeepSeekClient) sendRequest(ctx context.Context, endpoint string, request interface{}) ([]byte, error) {
	// 序列化请求
	requestBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// 创建HTTP请求
	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+endpoint, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	// 发送请求
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// 检查HTTP状态码
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(responseBody))
	}

	return responseBody, nil
} 