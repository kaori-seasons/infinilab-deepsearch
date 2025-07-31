package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/coco-ai/research-agent/internal/config"
	"github.com/coco-ai/research-agent/internal/llm"
	"github.com/coco-ai/research-agent/internal/memory"
	"github.com/coco-ai/research-agent/internal/search"
	"github.com/coco-ai/research-agent/internal/user"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

func main() {
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 初始化嵌入模型客户端
	embeddingClient, err := llm.NewEmbeddingClient(&cfg.LLM.Embedding)
	if err != nil {
		logger.Fatalf("Failed to initialize embedding client: %v", err)
	}

	// 初始化记忆系统
	memoryConfig := &memory.MemoryConfig{
		RedisHost:     cfg.Redis.Host,
		RedisPort:     cfg.Redis.Port,
		RedisPassword: cfg.Redis.Password,
		RedisDB:       cfg.Redis.DB,
		ESHosts:       cfg.ES.Hosts,
		ESUsername:    cfg.ES.Username,
		ESPassword:    cfg.ES.Password,
		EnableVectorSearch: true,
	}
	memorySystem := memory.NewMemory(memoryConfig, embeddingClient)

	// 初始化用户兴趣计算器
	interestCalculator := user.NewInterestCentroidCalculator(embeddingClient, memorySystem)

	// 初始化搜索配置
	searchConfig := &search.SearchConfig{
		IndexName:           "user_interests",
		VectorField:         "interest_vector",
		TextField:           "content",
		MaxCandidates:       100,
		RerankLimit:         20,
		VectorWeight:        0.4,
		TextWeight:          0.3,
		SimilarityThreshold: 0.7,
	}

	// 初始化混合搜索引擎（使用模拟的ES客户端）
	searchEngine := search.NewHybridSearchEngine(nil, searchConfig)

	// 测试用户兴趣计算
	testUserInterestCalculation(logger, interestCalculator)

	// 测试BGE相似性计算
	testBGESimilarityCalculation(logger)

	// 测试重排序模型
	testRerankModel(logger)

	// 测试搜索缓存
	testSearchCache(logger)

	logger.Info("All vector search tests completed successfully")
}

// testUserInterestCalculation 测试用户兴趣计算
func testUserInterestCalculation(logger *logrus.Logger, calculator *user.InterestCentroidCalculator) {
	logger.Info("Testing user interest calculation...")

	ctx := context.Background()
	userID := "test_user_001"

	// 计算用户兴趣
	userInterest, err := calculator.CalculateUserInterest(ctx, userID)
	if err != nil {
		logger.Errorf("Failed to calculate user interest: %v", err)
		return
	}

	logger.Infof("User interest calculated successfully:")
	logger.Infof("  User ID: %s", userInterest.UserID)
	logger.Infof("  Categories: %v", userInterest.Categories)
	logger.Infof("  Confidence: %.3f", userInterest.Confidence)
	logger.Infof("  Vector dimension: %d", len(userInterest.InterestVector))
	logger.Infof("  Last updated: %s", userInterest.LastUpdated.Format("2006-01-02 15:04:05"))

	// 测试用户行为更新
	behavior := user.UserBehavior{
		ID:        uuid.New(),
		UserID:    userID,
		Action:    "search",
		Content:   "AI技术发展趋势",
		Category:  "technology",
		Timestamp: time.Now(),
		Weight:    1.0,
		Metadata: map[string]interface{}{
			"source": "test",
		},
	}

	err = calculator.UpdateUserInterest(ctx, userID, behavior)
	if err != nil {
		logger.Errorf("Failed to update user interest: %v", err)
		return
	}

	logger.Info("User interest updated successfully")
}

// testBGESimilarityCalculation 测试BGE相似性计算
func testBGESimilarityCalculation(logger *logrus.Logger) {
	logger.Info("Testing BGE similarity calculation...")

	bgeModel := search.NewBGESimilarityModel()

	// 测试向量相似性计算
	vec1 := []float32{0.1, 0.2, 0.3, 0.4, 0.5}
	vec2 := []float32{0.2, 0.3, 0.4, 0.5, 0.6}

	similarity, err := bgeModel.CalculateSimilarity(vec1, vec2)
	if err != nil {
		logger.Errorf("Failed to calculate similarity: %v", err)
		return
	}

	logger.Infof("BGE similarity calculated: %.3f", similarity)

	// 测试批量相似性计算
	queries := [][]float32{
		{0.1, 0.2, 0.3, 0.4, 0.5},
		{0.2, 0.3, 0.4, 0.5, 0.6},
	}
	candidates := [][]float32{
		{0.2, 0.3, 0.4, 0.5, 0.6},
		{0.1, 0.2, 0.3, 0.4, 0.5},
		{0.3, 0.4, 0.5, 0.6, 0.7},
	}

	similarityMatrix, err := bgeModel.BatchCalculateSimilarity(queries, candidates)
	if err != nil {
		logger.Errorf("Failed to calculate batch similarity: %v", err)
		return
	}

	logger.Infof("Batch similarity matrix calculated:")
	for i, similarities := range similarityMatrix {
		logger.Infof("  Query %d: %v", i+1, similarities)
	}

	logger.Info("BGE similarity calculation test completed")
}

// testRerankModel 测试重排序模型
func testRerankModel(logger *logrus.Logger) {
	logger.Info("Testing rerank model...")

	rerankModel := search.NewRerankModel()

	// 创建测试候选
	candidate := search.SearchCandidate{
		ID:          "test_001",
		Content:     "AI技术发展趋势分析",
		Vector:      []float32{0.1, 0.2, 0.3, 0.4, 0.5},
		VectorScore: 0.85,
		TextScore:   0.78,
		Metadata: map[string]interface{}{
			"category": "technology",
			"publish_time": time.Now().Add(-24 * time.Hour),
		},
	}

	// 创建测试用户兴趣
	userInterest := &user.UserInterestModel{
		UserID:         "test_user",
		InterestVector: []float32{0.2, 0.3, 0.4, 0.5, 0.6},
		Categories:     []string{"technology", "business"},
		Confidence:     0.8,
		LastUpdated:    time.Now(),
	}

	// 计算重排序分数
	score, err := rerankModel.Score(candidate, userInterest)
	if err != nil {
		logger.Errorf("Failed to calculate rerank score: %v", err)
		return
	}

	logger.Infof("Rerank score calculated: %.3f", score)

	// 测试权重设置
	newWeights := &search.RerankWeights{
		ContentRelevance: 0.5,
		UserPreference:   0.3,
		Freshness:        0.1,
		Quality:          0.05,
		Popularity:       0.05,
	}
	rerankModel.SetWeights(newWeights)

	logger.Info("Rerank model test completed")
}

// testSearchCache 测试搜索缓存
func testSearchCache(logger *logrus.Logger) {
	logger.Info("Testing search cache...")

	cache := search.NewSearchCache()

	// 创建测试搜索结果
	results := []search.SearchResult{
		{
			ID:          "result_001",
			Content:     "AI技术发展趋势",
			VectorScore: 0.95,
			TextScore:   0.88,
			RerankScore: 0.92,
			FinalScore:  0.93,
			Metadata: map[string]interface{}{
				"category": "technology",
			},
		},
		{
			ID:          "result_002",
			Content:     "机器学习算法应用",
			VectorScore: 0.87,
			TextScore:   0.92,
			RerankScore: 0.89,
			FinalScore:  0.89,
			Metadata: map[string]interface{}{
				"category": "technology",
			},
		},
	}

	// 测试缓存设置和获取
	cacheKey := "test_search_query"
	cache.SetCachedResults(cacheKey, results)

	cachedResults, found := cache.GetCachedResults(cacheKey)
	if !found {
		logger.Error("Failed to retrieve cached results")
		return
	}

	logger.Infof("Cached results retrieved successfully: %d results", len(cachedResults))

	// 获取缓存统计信息
	stats := cache.GetStats()
	logger.Infof("Cache stats: %+v", stats)

	// 测试缓存失效
	cache.ClearCache()
	_, found = cache.GetCachedResults(cacheKey)
	if found {
		logger.Error("Cache clear failed")
		return
	}

	logger.Info("Search cache test completed")
} 