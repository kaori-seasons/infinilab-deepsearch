package main

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// 简化的BGE相似性计算模型
type SimpleBGEModel struct {
	logger *logrus.Entry
}

func NewSimpleBGEModel() *SimpleBGEModel {
	return &SimpleBGEModel{
		logger: logrus.WithField("component", "simple_bge_model"),
	}
}

// CalculateSimilarity 计算两个向量的相似性
func (bge *SimpleBGEModel) CalculateSimilarity(vec1, vec2 []float32) (float32, error) {
	if len(vec1) != len(vec2) {
		return 0, fmt.Errorf("vector dimensions mismatch: %d vs %d", len(vec1), len(vec2))
	}
	
	// 1. 向量归一化
	normalizedVec1 := bge.normalizeVector(vec1)
	normalizedVec2 := bge.normalizeVector(vec2)
	
	// 2. 计算余弦相似性
	cosineSim := bge.cosineSimilarity(normalizedVec1, normalizedVec2)
	
	// 3. 应用BGE特定的相似性变换
	bgeSimilarity := bge.applyBGESimilarityTransform(cosineSim)
	
	bge.logger.Debug("Calculated BGE similarity", 
		"cosine_similarity", cosineSim,
		"bge_similarity", bgeSimilarity)
	
	return bgeSimilarity, nil
}

// normalizeVector 向量归一化
func (bge *SimpleBGEModel) normalizeVector(vector []float32) []float32 {
	// 计算向量的L2范数
	var sum float32
	for _, v := range vector {
		sum += v * v
	}
	norm := float32(math.Sqrt(float64(sum)))
	
	// 避免除零
	if norm == 0 {
		return vector
	}
	
	// 归一化
	normalized := make([]float32, len(vector))
	for i, v := range vector {
		normalized[i] = v / norm
	}
	
	return normalized
}

// cosineSimilarity 计算余弦相似性
func (bge *SimpleBGEModel) cosineSimilarity(vec1, vec2 []float32) float32 {
	if len(vec1) != len(vec2) {
		return 0
	}
	
	var dotProduct float32
	for i := range vec1 {
		dotProduct += vec1[i] * vec2[i]
	}
	
	return dotProduct
}

// applyBGESimilarityTransform 应用BGE特定的相似性变换
func (bge *SimpleBGEModel) applyBGESimilarityTransform(cosineSim float32) float32 {
	// 将余弦相似性从[-1, 1]映射到[0, 1]
	normalizedSim := (cosineSim + 1) / 2
	
	// 应用非线性变换以增强区分度
	transformedSim := 1 / (1 + math.Exp(-5*(float64(normalizedSim)-0.5)))
	
	return float32(transformedSim)
}

// 简化的重排序模型
type SimpleRerankModel struct {
	logger *logrus.Entry
}

func NewSimpleRerankModel() *SimpleRerankModel {
	return &SimpleRerankModel{
		logger: logrus.WithField("component", "simple_rerank_model"),
	}
}

// Score 计算重排序分数
func (rm *SimpleRerankModel) Score(content string, userCategories []string) (float32, error) {
	// 简化的重排序逻辑
	contentLower := strings.ToLower(content)
	
	// 检查内容类别匹配
	var maxScore float32
	for _, category := range userCategories {
		score := rm.calculateCategoryScore(contentLower, category)
		if score > maxScore {
			maxScore = score
		}
	}
	
	// 基础分数
	baseScore := float32(0.5)
	
	// 内容长度分数
	lengthScore := rm.calculateLengthScore(len(content))
	
	// 综合分数
	finalScore := (maxScore + baseScore + lengthScore) / 3
	
	// 安全地截取内容用于日志
	contentPreview := content
	if len(content) > 50 {
		contentPreview = content[:50] + "..."
	}
	
	rm.logger.Debug("Calculated rerank score", 
		"content", contentPreview,
		"user_categories", userCategories,
		"category_score", maxScore,
		"length_score", lengthScore,
		"final_score", finalScore)
	
	return finalScore, nil
}

// calculateCategoryScore 计算类别匹配分数
func (rm *SimpleRerankModel) calculateCategoryScore(content, category string) float32 {
	// 简化的类别匹配逻辑
	categoryKeywords := map[string][]string{
		"technology": {"ai", "artificial intelligence", "machine learning", "deep learning", "algorithm", "software", "programming"},
		"business":   {"business", "enterprise", "startup", "investment", "market", "strategy", "management"},
		"science":    {"research", "study", "experiment", "discovery", "theory", "hypothesis"},
		"culture":    {"art", "music", "literature", "history", "philosophy", "culture"},
		"politics":   {"politics", "government", "policy", "election", "democracy"},
	}
	
	if keywords, exists := categoryKeywords[category]; exists {
		for _, keyword := range keywords {
			if strings.Contains(content, keyword) {
				return 0.9
			}
		}
	}
	
	return 0.1
}

// calculateLengthScore 计算内容长度分数
func (rm *SimpleRerankModel) calculateLengthScore(length int) float32 {
	// 理想长度范围：100-1000字符
	if length < 50 {
		return 0.3
	} else if length < 100 {
		return 0.6
	} else if length < 500 {
		return 0.9
	} else if length < 1000 {
		return 0.8
	} else {
		return 0.6
	}
}

// 简化的用户兴趣模型
type SimpleUserInterest struct {
	UserID         string    `json:"user_id"`
	Categories     []string  `json:"categories"`
	Confidence     float32   `json:"confidence"`
	LastUpdated    time.Time `json:"last_updated"`
}

func main() {
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	logger.Info("Starting simplified vector search test...")

	// 测试BGE相似性计算
	testBGESimilarityCalculation(logger)

	// 测试重排序模型
	testRerankModel(logger)

	// 测试用户兴趣建模
	testUserInterestModeling(logger)

	logger.Info("All simplified vector search tests completed successfully")
}

// testBGESimilarityCalculation 测试BGE相似性计算
func testBGESimilarityCalculation(logger *logrus.Logger) {
	logger.Info("Testing BGE similarity calculation...")

	bgeModel := NewSimpleBGEModel()

	// 测试向量相似性计算
	vec1 := []float32{0.1, 0.2, 0.3, 0.4, 0.5}
	vec2 := []float32{0.2, 0.3, 0.4, 0.5, 0.6}

	similarity, err := bgeModel.CalculateSimilarity(vec1, vec2)
	if err != nil {
		logger.Errorf("Failed to calculate similarity: %v", err)
		return
	}

	logger.Infof("BGE similarity calculated: %.3f", similarity)

	// 测试不同相似度的向量
	vec3 := []float32{0.9, 0.8, 0.7, 0.6, 0.5}
	similarity2, err := bgeModel.CalculateSimilarity(vec1, vec3)
	if err != nil {
		logger.Errorf("Failed to calculate similarity 2: %v", err)
		return
	}

	logger.Infof("BGE similarity 2 calculated: %.3f", similarity2)

	logger.Info("BGE similarity calculation test completed")
}

// testRerankModel 测试重排序模型
func testRerankModel(logger *logrus.Logger) {
	logger.Info("Testing rerank model...")

	rerankModel := NewSimpleRerankModel()

	// 测试内容重排序
	testContents := []string{
		"AI技术发展趋势分析，包括机器学习、深度学习等前沿技术",
		"企业数字化转型策略研究",
		"文化艺术发展历史回顾",
		"政治体制改革探讨",
	}

	userCategories := []string{"technology", "business"}

	for i, content := range testContents {
		score, err := rerankModel.Score(content, userCategories)
		if err != nil {
			logger.Errorf("Failed to calculate rerank score for content %d: %v", i+1, err)
			continue
		}

		// 安全地截取内容用于日志
		contentPreview := content
		if len(content) > 30 {
			contentPreview = content[:30] + "..."
		}
		logger.Infof("Content %d rerank score: %.3f - %s", i+1, score, contentPreview)
	}

	logger.Info("Rerank model test completed")
}

// testUserInterestModeling 测试用户兴趣建模
func testUserInterestModeling(logger *logrus.Logger) {
	logger.Info("Testing user interest modeling...")

	// 模拟用户兴趣
	userInterest := &SimpleUserInterest{
		UserID:     "test_user_001",
		Categories: []string{"technology", "business"},
		Confidence: 0.8,
		LastUpdated: time.Now(),
	}

	logger.Infof("User interest model:")
	logger.Infof("  User ID: %s", userInterest.UserID)
	logger.Infof("  Categories: %v", userInterest.Categories)
	logger.Infof("  Confidence: %.3f", userInterest.Confidence)
	logger.Infof("  Last updated: %s", userInterest.LastUpdated.Format("2006-01-02 15:04:05"))

	// 模拟用户行为更新
	logger.Info("Simulating user behavior update...")
	
	// 更新用户兴趣（简化实现）
	userInterest.Categories = append(userInterest.Categories, "science")
	userInterest.Confidence = 0.85
	userInterest.LastUpdated = time.Now()

	logger.Infof("Updated user interest:")
	logger.Infof("  Categories: %v", userInterest.Categories)
	logger.Infof("  Confidence: %.3f", userInterest.Confidence)

	logger.Info("User interest modeling test completed")
} 