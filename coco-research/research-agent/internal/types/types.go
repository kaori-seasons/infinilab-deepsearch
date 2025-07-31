package types

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Message 消息结构
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// LLMOptions LLM选项
type LLMOptions struct {
	MaxTokens   int     `json:"max_tokens,omitempty"`
	Temperature float64 `json:"temperature,omitempty"`
	Stream      bool    `json:"stream,omitempty"`
}

// Memory 记忆接口
type Memory interface {
	Store(sessionID uuid.UUID, role string, content string) error
	Retrieve(sessionID uuid.UUID, query string, limit int) ([]Message, error)
	Clear(sessionID uuid.UUID) error
}

// Agent 智能体接口
type Agent interface {
	GetID() uuid.UUID
	GetName() string
	GetDescription() string
	GetState() string
	GetType() string
	Execute(ctx context.Context, query string) (string, error)
	Stop() error
}

// Tool 工具接口
type Tool interface {
	GetName() string
	GetDescription() string
	Execute(ctx context.Context, params map[string]interface{}) (string, error)
}

// UserInterest 用户兴趣模型
type UserInterest struct {
	UserID         string    `json:"user_id"`
	InterestVector []float32 `json:"interest_vector"`
	Categories     []string  `json:"categories"`
	LastUpdated    time.Time `json:"last_updated"`
	Confidence     float32   `json:"confidence"`
	Version        int       `json:"version"`
}

// UserBehavior 用户行为数据
type UserBehavior struct {
	ID          uuid.UUID              `json:"id"`
	UserID      string                 `json:"user_id"`
	Action      string                 `json:"action"`
	Content     string                 `json:"content"`
	Category    string                 `json:"category"`
	Timestamp   time.Time              `json:"timestamp"`
	Weight      float32                `json:"weight"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// SearchRequest 搜索请求
type SearchRequest struct {
	Query           string                 `json:"query"`
	UserInterest    *UserInterest          `json:"user_interest"`
	Filters         map[string]interface{} `json:"filters"`
	Limit           int                    `json:"limit"`
	RerankLimit     int                    `json:"rerank_limit"`
	VectorWeight    float32                `json:"vector_weight"`
	TextWeight      float32                `json:"text_weight"`
}

// SearchCandidate 搜索候选
type SearchCandidate struct {
	ID          string                 `json:"id"`
	Content     string                 `json:"content"`
	Vector      []float32              `json:"vector"`
	VectorScore float32                `json:"vector_score"`
	TextScore   float32                `json:"text_score"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// SearchResult 搜索结果
type SearchResult struct {
	ID          string                 `json:"id"`
	Content     string                 `json:"content"`
	VectorScore float32                `json:"vector_score"`
	TextScore   float32                `json:"text_score"`
	RerankScore float32                `json:"rerank_score"`
	FinalScore  float32                `json:"final_score"`
	Metadata    map[string]interface{} `json:"metadata"`
} 