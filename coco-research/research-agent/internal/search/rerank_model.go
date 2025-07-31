package search

import (
	"context"
	"fmt"
	"math"
	"strings"

	"github.com/coco-ai/research-agent/internal/user"
	"github.com/sirupsen/logrus"
)

// RerankModel 重排序模型
type RerankModel struct {
	modelPath string
	device    string
	logger    *logrus.Entry
	weights   *RerankWeights
}

// RerankWeights 重排序权重配置
type RerankWeights struct {
	ContentRelevance float32 `json:"content_relevance"`
	UserPreference   float32 `json:"user_preference"`
	Freshness        float32 `json:"freshness"`
	Quality          float32 `json:"quality"`
	Popularity       float32 `json:"popularity"`
}

// NewRerankModel 创建重排序模型
func NewRerankModel() *RerankModel {
	return &RerankModel{
		modelPath: "BAAI/bge-reranker-v2-m3",
		device:    "cpu",
		logger:    logrus.WithField("component", "rerank_model"),
		weights: &RerankWeights{
			ContentRelevance: 0.4,
			UserPreference:   0.3,
			Freshness:        0.1,
			Quality:          0.1,
			Popularity:       0.1,
		},
	}
}

// Score 计算重排序分数
func (rm *RerankModel) Score(candidate SearchCandidate, userInterest *user.UserInterestModel) (float32, error) {
	var totalScore float32
	
	// 1. 内容相关性评分
	contentScore, err := rm.calculateContentRelevance(candidate, userInterest)
	if err != nil {
		rm.logger.Warn("Failed to calculate content relevance", "error", err)
		contentScore = 0
	}
	
	// 2. 用户偏好评分
	preferenceScore, err := rm.calculateUserPreference(candidate, userInterest)
	if err != nil {
		rm.logger.Warn("Failed to calculate user preference", "error", err)
		preferenceScore = 0
	}
	
	// 3. 新鲜度评分
	freshnessScore := rm.calculateFreshness(candidate)
	
	// 4. 质量评分
	qualityScore := rm.calculateQuality(candidate)
	
	// 5. 流行度评分
	popularityScore := rm.calculatePopularity(candidate)
	
	// 加权计算最终分数
	totalScore = contentScore*rm.weights.ContentRelevance +
		preferenceScore*rm.weights.UserPreference +
		freshnessScore*rm.weights.Freshness +
		qualityScore*rm.weights.Quality +
		popularityScore*rm.weights.Popularity
	
	rm.logger.Debug("Calculated rerank score", 
		"candidate_id", candidate.ID,
		"content_score", contentScore,
		"preference_score", preferenceScore,
		"freshness_score", freshnessScore,
		"quality_score", qualityScore,
		"popularity_score", popularityScore,
		"total_score", totalScore)
	
	return totalScore, nil
}

// calculateContentRelevance 计算内容相关性
func (rm *RerankModel) calculateContentRelevance(candidate SearchCandidate, userInterest *user.UserInterestModel) (float32, error) {
	// 基于用户兴趣类别和内容类别的匹配度
	contentCategory := rm.extractCategory(candidate.Content)
	userCategories := userInterest.Categories
	
	// 计算类别匹配度
	var maxMatchScore float32
	for _, userCategory := range userCategories {
		matchScore := rm.calculateCategoryMatch(contentCategory, userCategory)
		if matchScore > maxMatchScore {
			maxMatchScore = matchScore
		}
	}
	
	// 结合向量相似性
	vectorScore := candidate.VectorScore
	textScore := candidate.TextScore
	
	// 综合评分
	relevanceScore := (maxMatchScore + vectorScore + textScore) / 3
	
	return relevanceScore, nil
}

// calculateUserPreference 计算用户偏好
func (rm *RerankModel) calculateUserPreference(candidate SearchCandidate, userInterest *user.UserInterestModel) (float32, error) {
	// 基于用户历史行为和当前兴趣的偏好计算
	
	// 1. 类别偏好
	categoryPreference := rm.calculateCategoryPreference(candidate, userInterest)
	
	// 2. 内容类型偏好
	contentTypePreference := rm.calculateContentTypePreference(candidate, userInterest)
	
	// 3. 时间偏好
	timePreference := rm.calculateTimePreference(candidate, userInterest)
	
	// 综合偏好分数
	preferenceScore := (categoryPreference + contentTypePreference + timePreference) / 3
	
	return preferenceScore, nil
}

// calculateFreshness 计算新鲜度
func (rm *RerankModel) calculateFreshness(candidate SearchCandidate) float32 {
	// 基于内容的发布时间计算新鲜度
	// 这里简化实现，实际应该从metadata中获取发布时间
	
	// 假设内容越新分数越高
	freshnessScore := float32(0.8) // 默认新鲜度
	
	// 如果有发布时间信息，可以基于时间差计算
	if publishTime, exists := candidate.Metadata["publish_time"]; exists {
		// 计算时间差并转换为新鲜度分数
		// 这里简化实现
		freshnessScore = 0.9
	}
	
	return freshnessScore
}

// calculateQuality 计算质量分数
func (rm *RerankModel) calculateQuality(candidate SearchCandidate) float32 {
	// 基于多个质量指标计算
	
	// 1. 内容长度
	contentLength := len(candidate.Content)
	lengthScore := rm.normalizeLength(contentLength)
	
	// 2. 内容完整性
	completenessScore := rm.calculateCompleteness(candidate.Content)
	
	// 3. 来源权威性
	authorityScore := rm.calculateAuthority(candidate)
	
	// 4. 引用数量
	citationScore := rm.calculateCitations(candidate)
	
	// 综合质量分数
	qualityScore := (lengthScore + completenessScore + authorityScore + citationScore) / 4
	
	return qualityScore
}

// calculatePopularity 计算流行度
func (rm *RerankModel) calculatePopularity(candidate SearchCandidate) float32 {
	// 基于多个流行度指标计算
	
	// 1. 点击量
	clickScore := rm.getClickScore(candidate)
	
	// 2. 分享数
	shareScore := rm.getShareScore(candidate)
	
	// 3. 评论数
	commentScore := rm.getCommentScore(candidate)
	
	// 4. 收藏数
	bookmarkScore := rm.getBookmarkScore(candidate)
	
	// 综合流行度分数
	popularityScore := (clickScore + shareScore + commentScore + bookmarkScore) / 4
	
	return popularityScore
}

// extractCategory 提取内容类别
func (rm *RerankModel) extractCategory(content string) string {
	// 基于关键词提取类别
	content = strings.ToLower(content)
	
	categories := map[string][]string{
		"technology": {"ai", "artificial intelligence", "machine learning", "deep learning", "algorithm", "software", "programming"},
		"business":   {"business", "enterprise", "startup", "investment", "market", "strategy", "management"},
		"science":    {"research", "study", "experiment", "discovery", "theory", "hypothesis"},
		"culture":    {"art", "music", "literature", "history", "philosophy", "culture"},
		"politics":   {"politics", "government", "policy", "election", "democracy"},
	}
	
	for category, keywords := range categories {
		for _, keyword := range keywords {
			if strings.Contains(content, keyword) {
				return category
			}
		}
	}
	
	return "general"
}

// calculateCategoryMatch 计算类别匹配度
func (rm *RerankModel) calculateCategoryMatch(contentCategory, userCategory string) float32 {
	if contentCategory == userCategory {
		return 1.0
	}
	
	// 类别相似性矩阵
	similarityMatrix := map[string]map[string]float32{
		"technology": {
			"science": 0.8,
			"business": 0.6,
			"culture": 0.2,
			"politics": 0.1,
		},
		"business": {
			"technology": 0.6,
			"politics": 0.5,
			"science": 0.3,
			"culture": 0.2,
		},
		"science": {
			"technology": 0.8,
			"culture": 0.4,
			"business": 0.3,
			"politics": 0.2,
		},
		"culture": {
			"science": 0.4,
			"politics": 0.3,
			"technology": 0.2,
			"business": 0.2,
		},
		"politics": {
			"business": 0.5,
			"culture": 0.3,
			"science": 0.2,
			"technology": 0.1,
		},
	}
	
	if similarities, exists := similarityMatrix[contentCategory]; exists {
		if similarity, exists := similarities[userCategory]; exists {
			return similarity
		}
	}
	
	return 0.1 // 默认低相似性
}

// calculateCategoryPreference 计算类别偏好
func (rm *RerankModel) calculateCategoryPreference(candidate SearchCandidate, userInterest *user.UserInterestModel) float32 {
	contentCategory := rm.extractCategory(candidate.Content)
	
	// 基于用户兴趣置信度计算偏好
	confidence := userInterest.Confidence
	
	// 检查内容类别是否在用户主要兴趣中
	for _, userCategory := range userInterest.Categories {
		if contentCategory == userCategory {
			return confidence
		}
	}
	
	// 计算与用户兴趣的相似性
	var maxSimilarity float32
	for _, userCategory := range userInterest.Categories {
		similarity := rm.calculateCategoryMatch(contentCategory, userCategory)
		if similarity > maxSimilarity {
			maxSimilarity = similarity
		}
	}
	
	return maxSimilarity * confidence
}

// calculateContentTypePreference 计算内容类型偏好
func (rm *RerankModel) calculateContentTypePreference(candidate SearchCandidate, userInterest *user.UserInterestModel) float32 {
	// 基于内容类型（文章、视频、图片等）计算偏好
	contentType := rm.extractContentType(candidate.Content)
	
	// 简化实现，假设用户偏好技术类文章
	if strings.Contains(strings.ToLower(candidate.Content), "technology") {
		return 0.9
	}
	
	return 0.5 // 默认偏好
}

// calculateTimePreference 计算时间偏好
func (rm *RerankModel) calculateTimePreference(candidate SearchCandidate, userInterest *user.UserInterestModel) float32 {
	// 基于用户活动时间模式计算偏好
	// 简化实现
	return 0.7
}

// normalizeLength 归一化内容长度
func (rm *RerankModel) normalizeLength(length int) float32 {
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

// calculateCompleteness 计算内容完整性
func (rm *RerankModel) calculateCompleteness(content string) float32 {
	// 基于内容结构完整性计算
	// 检查是否包含标题、段落、结论等
	
	completeness := float32(0.5) // 基础分数
	
	// 检查是否有标题
	if strings.Contains(content, "#") || strings.Contains(content, "标题") {
		completeness += 0.1
	}
	
	// 检查是否有段落分隔
	if strings.Count(content, "\n") > 2 {
		completeness += 0.1
	}
	
	// 检查是否有结论性词汇
	conclusionWords := []string{"总结", "结论", "因此", "总之", "综上所述"}
	for _, word := range conclusionWords {
		if strings.Contains(content, word) {
			completeness += 0.1
			break
		}
	}
	
	// 检查是否有引用或链接
	if strings.Contains(content, "http") || strings.Contains(content, "引用") {
		completeness += 0.1
	}
	
	return float32(math.Min(float64(completeness), 1.0))
}

// calculateAuthority 计算权威性
func (rm *RerankModel) calculateAuthority(candidate SearchCandidate) float32 {
	// 基于来源权威性计算
	// 简化实现
	return 0.7
}

// calculateCitations 计算引用数
func (rm *RerankModel) calculateCitations(candidate SearchCandidate) float32 {
	// 基于引用数量计算
	// 简化实现
	return 0.6
}

// getClickScore 获取点击分数
func (rm *RerankModel) getClickScore(candidate SearchCandidate) float32 {
	// 基于点击量计算
	// 简化实现
	return 0.5
}

// getShareScore 获取分享分数
func (rm *RerankModel) getShareScore(candidate SearchCandidate) float32 {
	// 基于分享数计算
	// 简化实现
	return 0.4
}

// getCommentScore 获取评论分数
func (rm *RerankModel) getCommentScore(candidate SearchCandidate) float32 {
	// 基于评论数计算
	// 简化实现
	return 0.3
}

// getBookmarkScore 获取收藏分数
func (rm *RerankModel) getBookmarkScore(candidate SearchCandidate) float32 {
	// 基于收藏数计算
	// 简化实现
	return 0.6
}

// extractContentType 提取内容类型
func (rm *RerankModel) extractContentType(content string) string {
	// 基于内容特征提取类型
	if strings.Contains(content, "视频") || strings.Contains(content, "video") {
		return "video"
	} else if strings.Contains(content, "图片") || strings.Contains(content, "image") {
		return "image"
	} else {
		return "article"
	}
}

// SetWeights 设置重排序权重
func (rm *RerankModel) SetWeights(weights *RerankWeights) {
	rm.weights = weights
	rm.logger.Info("Rerank weights updated", "weights", weights)
}

// GetWeights 获取重排序权重
func (rm *RerankModel) GetWeights() *RerankWeights {
	return rm.weights
} 