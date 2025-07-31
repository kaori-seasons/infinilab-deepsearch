package agent

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/coco-ai/research-agent/internal/llm"
	"github.com/coco-ai/research-agent/internal/search"
	"github.com/coco-ai/research-agent/internal/user"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// EnhancedAgent 增强的智能体，支持用户兴趣和混合搜索
type EnhancedAgent struct {
	*BaseAgent
	interestCalculator *user.InterestCentroidCalculator
	searchEngine       *search.HybridSearchEngine
	userID             string
	logger             *logrus.Entry
}

// NewEnhancedAgent 创建增强的智能体
func NewEnhancedAgent(
	baseAgent *BaseAgent,
	interestCalculator *user.InterestCentroidCalculator,
	searchEngine *search.HybridSearchEngine,
	userID string,
) *EnhancedAgent {
	return &EnhancedAgent{
		BaseAgent:          baseAgent,
		interestCalculator: interestCalculator,
		searchEngine:       searchEngine,
		userID:            userID,
		logger:            logrus.WithField("component", "enhanced_agent").WithField("user_id", userID),
	}
}

// ExecuteWithUserInterest 基于用户兴趣执行任务
func (ea *EnhancedAgent) ExecuteWithUserInterest(ctx context.Context, query string) (*EnhancedExecutionResult, error) {
	ea.logger.Info("Starting enhanced execution with user interest", "query", query)
	
	// 1. 计算用户兴趣
	userInterest, err := ea.interestCalculator.CalculateUserInterest(ctx, ea.userID)
	if err != nil {
		ea.logger.Warn("Failed to calculate user interest, using default", "error", err)
		userInterest = &user.UserInterestModel{
			UserID:         ea.userID,
			InterestVector: make([]float32, 256), // 默认向量
			Categories:     []string{"general"},
			Confidence:     0.5,
		}
	}
	
	// 2. 执行混合搜索
	searchReq := &search.SearchRequest{
		Query:        query,
		UserInterest: userInterest,
		Limit:        20,
		RerankLimit:  10,
		VectorWeight: 0.4,
		TextWeight:   0.3,
	}
	
	searchResults, err := ea.searchEngine.Search(ctx, searchReq)
	if err != nil {
		ea.logger.Error("Failed to execute hybrid search", "error", err)
		return nil, fmt.Errorf("hybrid search failed: %w", err)
	}
	
	// 3. 基于搜索结果生成响应
	response, err := ea.generateResponseFromSearchResults(ctx, query, searchResults, userInterest)
	if err != nil {
		ea.logger.Error("Failed to generate response", "error", err)
		return nil, fmt.Errorf("response generation failed: %w", err)
	}
	
	// 4. 更新用户兴趣
	err = ea.updateUserInterest(ctx, query, searchResults)
	if err != nil {
		ea.logger.Warn("Failed to update user interest", "error", err)
	}
	
	result := &EnhancedExecutionResult{
		Query:         query,
		UserInterest:  userInterest,
		SearchResults: searchResults,
		Response:      response,
		Timestamp:     time.Now(),
	}
	
	ea.logger.Info("Enhanced execution completed", 
		"query", query,
		"results_count", len(searchResults),
		"user_confidence", userInterest.Confidence)
	
	return result, nil
}

// generateResponseFromSearchResults 基于搜索结果生成响应
func (ea *EnhancedAgent) generateResponseFromSearchResults(
	ctx context.Context,
	query string,
	searchResults []search.SearchResult,
	userInterest *user.UserInterestModel,
) (string, error) {
	// 构建提示词
	prompt := ea.buildEnhancedPrompt(query, searchResults, userInterest)
	
	// 调用LLM生成响应
	response, err := ea.llmClient.Chat(ctx, []llm.Message{
		{
			Role:    "system",
			Content: ea.getEnhancedSystemPrompt(),
		},
		{
			Role:    "user",
			Content: prompt,
		},
	})
	if err != nil {
		return "", fmt.Errorf("LLM chat failed: %w", err)
	}
	
	return response, nil
}

// buildEnhancedPrompt 构建增强的提示词
func (ea *EnhancedAgent) buildEnhancedPrompt(
	query string,
	searchResults []search.SearchResult,
	userInterest *user.UserInterestModel,
) string {
	prompt := fmt.Sprintf(`
用户查询: %s

用户兴趣信息:
- 主要类别: %v
- 置信度: %.2f
- 最后更新: %s

搜索结果 (按相关性排序):
`, query, userInterest.Categories, userInterest.Confidence, userInterest.LastUpdated.Format("2006-01-02 15:04:05"))
	
	// 添加前5个搜索结果
	for i, result := range searchResults {
		if i >= 5 {
			break
		}
		prompt += fmt.Sprintf(`
%d. %s
   相关性分数: %.3f (向量: %.3f, 文本: %.3f, 重排序: %.3f)
   类别: %s
`, i+1, result.Content, result.FinalScore, result.VectorScore, result.TextScore, result.RerankScore, result.Metadata["category"])
	}
	
	prompt += `

请基于以上搜索结果和用户兴趣，生成一个全面、准确、个性化的回答。回答应该：
1. 直接回应用户的查询
2. 考虑用户的兴趣偏好
3. 综合多个来源的信息
4. 提供有价值的见解和建议
5. 使用用户偏好的内容类型和风格

请用中文回答，并确保回答的准确性和相关性。
`
	
	return prompt
}

// getEnhancedSystemPrompt 获取增强的系统提示词
func (ea *EnhancedAgent) getEnhancedSystemPrompt() string {
	return `你是一个智能研究助手，专门为用户提供个性化的研究支持。

你的能力包括：
1. 基于用户兴趣进行个性化搜索
2. 综合分析多个信息源
3. 提供准确、有价值的见解
4. 适应用户的偏好和需求

回答要求：
- 准确、全面、有条理
- 考虑用户兴趣和偏好
- 提供实用的建议和见解
- 使用清晰、易懂的语言
- 适当引用信息来源

请始终以用户为中心，提供最有价值的帮助。`
}

// updateUserInterest 更新用户兴趣
func (ea *EnhancedAgent) updateUserInterest(ctx context.Context, query string, searchResults []search.SearchResult) error {
	// 创建用户行为记录
	behavior := user.UserBehavior{
		ID:        uuid.New(),
		UserID:    ea.userID,
		Action:    "search",
		Content:   query,
		Category:  ea.extractQueryCategory(query),
		Timestamp: time.Now(),
		Weight:    1.0,
		Metadata: map[string]interface{}{
			"results_count": len(searchResults),
			"agent_type":    ea.GetType(),
		},
	}
	
	// 更新用户兴趣
	err := ea.interestCalculator.UpdateUserInterest(ctx, ea.userID, behavior)
	if err != nil {
		return fmt.Errorf("failed to update user interest: %w", err)
	}
	
	ea.logger.Debug("Updated user interest", 
		"query", query,
		"category", behavior.Category,
		"results_count", len(searchResults))
	
	return nil
}

// extractQueryCategory 提取查询类别
func (ea *EnhancedAgent) extractQueryCategory(query string) string {
	// 简化的类别提取逻辑
	query = strings.ToLower(query)
	
	categories := map[string][]string{
		"technology": {"ai", "artificial intelligence", "machine learning", "deep learning", "algorithm", "software", "programming"},
		"business":   {"business", "enterprise", "startup", "investment", "market", "strategy", "management"},
		"science":    {"research", "study", "experiment", "discovery", "theory", "hypothesis"},
		"culture":    {"art", "music", "literature", "history", "philosophy", "culture"},
		"politics":   {"politics", "government", "policy", "election", "democracy"},
	}
	
	for category, keywords := range categories {
		for _, keyword := range keywords {
			if strings.Contains(query, keyword) {
				return category
			}
		}
	}
	
	return "general"
}

// GetUserInterest 获取用户兴趣
func (ea *EnhancedAgent) GetUserInterest(ctx context.Context) (*user.UserInterestModel, error) {
	return ea.interestCalculator.CalculateUserInterest(ctx, ea.userID)
}

// GetSearchStats 获取搜索统计信息
func (ea *EnhancedAgent) GetSearchStats() map[string]interface{} {
	stats := ea.searchEngine.cache.GetStats()
	stats["user_id"] = ea.userID
	stats["agent_type"] = ea.GetType()
	return stats
}

// EnhancedExecutionResult 增强的执行结果
type EnhancedExecutionResult struct {
	Query         string                    `json:"query"`
	UserInterest  *user.UserInterestModel   `json:"user_interest"`
	SearchResults []search.SearchResult     `json:"search_results"`
	Response      string                    `json:"response"`
	Timestamp     time.Time                 `json:"timestamp"`
	ExecutionTime time.Duration             `json:"execution_time"`
	Metadata      map[string]interface{}    `json:"metadata"`
} 