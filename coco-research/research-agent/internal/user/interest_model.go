package user

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/coco-ai/research-agent/internal/llm"
	"github.com/coco-ai/research-agent/internal/memory"
	"github.com/google/uuid"
)

// UserInterestModel 用户兴趣模型
type UserInterestModel struct {
	UserID         string    `json:"user_id"`
	InterestVector []float32 `json:"interest_vector"`
	Categories     []string  `json:"categories"`
	LastUpdated    time.Time `json:"last_updated"`
	Confidence     float32   `json:"confidence"`
	Version        int       `json:"version"`
}

// UserBehavior 用户行为数据
type UserBehavior struct {
	ID          uuid.UUID `json:"id"`
	UserID      string    `json:"user_id"`
	Action      string    `json:"action"`      // search, view, like, share, etc.
	Content     string    `json:"content"`     // 行为相关的内容
	Category    string    `json:"category"`    // 内容类别
	Timestamp   time.Time `json:"timestamp"`
	Weight      float32   `json:"weight"`      // 行为权重
	Metadata    map[string]interface{} `json:"metadata"`
}

// InterestCentroidCalculator 兴趣centroid计算器
type InterestCentroidCalculator struct {
	embeddingClient llm.EmbeddingClient
	memory          memory.Memory
	categories      []string
	cache           *InterestCache
}

// NewInterestCentroidCalculator 创建兴趣centroid计算器
func NewInterestCentroidCalculator(embeddingClient llm.EmbeddingClient, memory memory.Memory) *InterestCentroidCalculator {
	return &InterestCentroidCalculator{
		embeddingClient: embeddingClient,
		memory:          memory,
		categories:      []string{"technology", "business", "science", "culture", "politics"},
		cache:           NewInterestCache(),
	}
}

// CalculateUserInterest 计算用户兴趣
func (icc *InterestCentroidCalculator) CalculateUserInterest(ctx context.Context, userID string) (*UserInterestModel, error) {
	// 1. 尝试从缓存获取
	if cached, err := icc.cache.GetUserInterest(userID); err == nil {
		return cached, nil
	}

	// 2. 收集用户历史行为数据
	userHistory, err := icc.collectUserHistory(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to collect user history: %w", err)
	}

	// 3. 按类别分组
	categorizedData := icc.categorizeData(userHistory)

	// 4. 计算每个类别的centroid
	categoryVectors, err := icc.calculateCategoryCentroids(ctx, categorizedData)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate category centroids: %w", err)
	}

	// 5. 生成用户兴趣向量
	interestVector, err := icc.generateInterestVector(ctx, categoryVectors)
	if err != nil {
		return nil, fmt.Errorf("failed to generate interest vector: %w", err)
	}

	// 6. 创建用户兴趣模型
	interestModel := &UserInterestModel{
		UserID:         userID,
		InterestVector: interestVector,
		Categories:     icc.getTopCategories(categoryVectors),
		LastUpdated:    time.Now(),
		Confidence:     icc.calculateConfidence(categoryVectors),
		Version:        1,
	}

	// 7. 存入缓存
	icc.cache.SetUserInterest(userID, interestModel)

	return interestModel, nil
}

// collectUserHistory 收集用户历史行为
func (icc *InterestCentroidCalculator) collectUserHistory(ctx context.Context, userID string) ([]UserBehavior, error) {
	// 从记忆系统获取用户历史行为
	// 这里简化实现，实际应该从数据库或缓存中获取
	behaviors := []UserBehavior{
		{
			ID:        uuid.New(),
			UserID:    userID,
			Action:    "search",
			Content:   "AI技术发展趋势",
			Category:  "technology",
			Timestamp: time.Now().Add(-24 * time.Hour),
			Weight:    1.0,
		},
		{
			ID:        uuid.New(),
			UserID:    userID,
			Action:    "view",
			Content:   "机器学习算法",
			Category:  "technology",
			Timestamp: time.Now().Add(-12 * time.Hour),
			Weight:    0.8,
		},
		{
			ID:        uuid.New(),
			UserID:    userID,
			Action:    "like",
			Content:   "企业数字化转型",
			Category:  "business",
			Timestamp: time.Now().Add(-6 * time.Hour),
			Weight:    0.9,
		},
	}

	return behaviors, nil
}

// categorizeData 按类别分组数据
func (icc *InterestCentroidCalculator) categorizeData(behaviors []UserBehavior) map[string][]UserBehavior {
	categorized := make(map[string][]UserBehavior)
	
	for _, behavior := range behaviors {
		category := behavior.Category
		if category == "" {
			category = "general"
		}
		categorized[category] = append(categorized[category], behavior)
	}
	
	return categorized
}

// calculateCategoryCentroids 计算每个类别的centroid
func (icc *InterestCentroidCalculator) calculateCategoryCentroids(ctx context.Context, categorizedData map[string][]UserBehavior) (map[string][]float32, error) {
	categoryVectors := make(map[string][]float32)
	
	for category, behaviors := range categorizedData {
		if len(behaviors) == 0 {
			continue
		}
		
		// 提取类别内容
		var contents []string
		var weights []float32
		for _, behavior := range behaviors {
			contents = append(contents, behavior.Content)
			weights = append(weights, behavior.Weight)
		}
		
		// 生成嵌入向量
		embeddings, err := icc.embeddingClient.GenerateEmbeddings(ctx, contents)
		if err != nil {
			return nil, fmt.Errorf("failed to generate embeddings for category %s: %w", category, err)
		}
		
		// 计算加权平均centroid
		centroid := icc.calculateWeightedCentroid(embeddings, weights)
		categoryVectors[category] = centroid
	}
	
	return categoryVectors, nil
}

// calculateWeightedCentroid 计算加权平均centroid
func (icc *InterestCentroidCalculator) calculateWeightedCentroid(embeddings [][]float32, weights []float32) []float32 {
	if len(embeddings) == 0 {
		return nil
	}
	
	dim := len(embeddings[0])
	centroid := make([]float32, dim)
	totalWeight := float32(0)
	
	for i, embedding := range embeddings {
		weight := weights[i]
		totalWeight += weight
		
		for j, value := range embedding {
			centroid[j] += value * weight
		}
	}
	
	// 归一化
	if totalWeight > 0 {
		for i := range centroid {
			centroid[i] /= totalWeight
		}
	}
	
	return centroid
}

// generateInterestVector 生成用户兴趣向量
func (icc *InterestCentroidCalculator) generateInterestVector(ctx context.Context, categoryVectors map[string][]float32) ([]float32, error) {
	if len(categoryVectors) == 0 {
		// 返回默认向量
		return make([]float32, 256), nil
	}
	
	// 将所有类别向量合并
	var allVectors [][]float32
	var weights []float32
	
	for category, vector := range categoryVectors {
		allVectors = append(allVectors, vector)
		// 根据类别重要性分配权重
		weight := icc.getCategoryWeight(category)
		weights = append(weights, weight)
	}
	
	// 计算总体兴趣向量
	interestVector := icc.calculateWeightedCentroid(allVectors, weights)
	
	return interestVector, nil
}

// getCategoryWeight 获取类别权重
func (icc *InterestCentroidCalculator) getCategoryWeight(category string) float32 {
	// 根据类别重要性分配权重
	weights := map[string]float32{
		"technology": 1.0,
		"business":   0.8,
		"science":    0.7,
		"culture":    0.6,
		"politics":   0.5,
	}
	
	if weight, exists := weights[category]; exists {
		return weight
	}
	return 0.5 // 默认权重
}

// getTopCategories 获取主要类别
func (icc *InterestCentroidCalculator) getTopCategories(categoryVectors map[string][]float32) []string {
	var categories []string
	for category := range categoryVectors {
		categories = append(categories, category)
	}
	
	// 按权重排序
	sort.Slice(categories, func(i, j int) bool {
		weightI := icc.getCategoryWeight(categories[i])
		weightJ := icc.getCategoryWeight(categories[j])
		return weightI > weightJ
	})
	
	// 返回前3个类别
	if len(categories) > 3 {
		return categories[:3]
	}
	return categories
}

// calculateConfidence 计算置信度
func (icc *InterestCentroidCalculator) calculateConfidence(categoryVectors map[string][]float32) float32 {
	// 基于类别数量和向量质量计算置信度
	numCategories := len(categoryVectors)
	if numCategories == 0 {
		return 0.0
	}
	
	// 基础置信度
	confidence := float32(numCategories) / 5.0 // 假设最多5个类别
	
	// 根据向量质量调整
	for _, vector := range categoryVectors {
		if len(vector) > 0 {
			// 检查向量是否为零向量
			hasNonZero := false
			for _, v := range vector {
				if v != 0 {
					hasNonZero = true
					break
				}
			}
			if hasNonZero {
				confidence += 0.1
			}
		}
	}
	
	// 限制在0-1之间
	if confidence > 1.0 {
		confidence = 1.0
	}
	
	return confidence
}

// UpdateUserInterest 更新用户兴趣
func (icc *InterestCentroidCalculator) UpdateUserInterest(ctx context.Context, userID string, behavior UserBehavior) error {
	// 清除缓存
	icc.cache.ClearUserInterest(userID)
	
	// 重新计算用户兴趣
	_, err := icc.CalculateUserInterest(ctx, userID)
	return err
} 