package search

import (
	"context"
	"fmt"
	"sort"
	"sync"

	"github.com/coco-ai/research-agent/internal/user"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/sirupsen/logrus"
)

// SearchRequest 搜索请求
type SearchRequest struct {
	Query           string                 `json:"query"`
	UserInterest    *user.UserInterestModel `json:"user_interest"`
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

// HybridSearchEngine 混合搜索引擎
type HybridSearchEngine struct {
	esClient        *elasticsearch.Client
	bgeModel        *BGESimilarityModel
	rerankModel     *RerankModel
	config          *SearchConfig
	logger          *logrus.Entry
	cache           *SearchCache
}

// SearchConfig 搜索配置
type SearchConfig struct {
	IndexName           string  `json:"index_name"`
	VectorField         string  `json:"vector_field"`
	TextField           string  `json:"text_field"`
	MaxCandidates       int     `json:"max_candidates"`
	RerankLimit         int     `json:"rerank_limit"`
	VectorWeight        float32 `json:"vector_weight"`
	TextWeight          float32 `json:"text_weight"`
	SimilarityThreshold float32 `json:"similarity_threshold"`
}

// NewHybridSearchEngine 创建混合搜索引擎
func NewHybridSearchEngine(esClient *elasticsearch.Client, config *SearchConfig) *HybridSearchEngine {
	return &HybridSearchEngine{
		esClient:    esClient,
		bgeModel:    NewBGESimilarityModel(),
		rerankModel: NewRerankModel(),
		config:      config,
		logger:      logrus.WithField("component", "hybrid_search_engine"),
		cache:       NewSearchCache(),
	}
}

// Search 执行混合搜索
func (hse *HybridSearchEngine) Search(ctx context.Context, req *SearchRequest) ([]SearchResult, error) {
	// 生成缓存键
	cacheKey := hse.generateCacheKey(req)
	
	// 尝试从缓存获取
	if cached, found := hse.cache.GetCachedResults(cacheKey); found {
		hse.logger.Info("Cache hit for search request", "cache_key", cacheKey)
		return cached, nil
	}
	
	// 第一阶段：初步过滤
	candidates, err := hse.initialFilter(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("initial filter failed: %w", err)
	}
	
	// 第二阶段：精确重排序
	results, err := hse.preciseRerank(ctx, candidates, req.UserInterest)
	if err != nil {
		return nil, fmt.Errorf("precise rerank failed: %w", err)
	}
	
	// 存入缓存
	hse.cache.SetCachedResults(cacheKey, results)
	
	return results, nil
}

// initialFilter 初步过滤
func (hse *HybridSearchEngine) initialFilter(ctx context.Context, req *SearchRequest) ([]SearchCandidate, error) {
	var candidates []SearchCandidate
	
	// 1. 向量相似性搜索
	vectorResults, err := hse.vectorSearch(ctx, req)
	if err != nil {
		hse.logger.Warn("Vector search failed", "error", err)
	} else {
		candidates = append(candidates, vectorResults...)
	}
	
	// 2. 全文搜索
	textResults, err := hse.textSearch(ctx, req)
	if err != nil {
		hse.logger.Warn("Text search failed", "error", err)
	} else {
		candidates = append(candidates, textResults...)
	}
	
	// 3. 结果融合和去重
	mergedCandidates := hse.mergeAndDeduplicate(candidates)
	
	// 4. 限制候选数量
	if len(mergedCandidates) > req.Limit*2 {
		mergedCandidates = mergedCandidates[:req.Limit*2]
	}
	
	return mergedCandidates, nil
}

// vectorSearch 向量搜索
func (hse *HybridSearchEngine) vectorSearch(ctx context.Context, req *SearchRequest) ([]SearchCandidate, error) {
	if req.UserInterest == nil || len(req.UserInterest.InterestVector) == 0 {
		return nil, fmt.Errorf("no user interest vector available")
	}
	
	// 构建向量搜索查询
	query := map[string]interface{}{
		"size": req.Limit * 2,
		"query": map[string]interface{}{
			"knn": map[string]interface{}{
				hse.config.VectorField: map[string]interface{}{
					"vector": req.UserInterest.InterestVector,
					"k":      req.Limit * 2,
					"num_candidates": req.Limit * 4,
				},
			},
		},
	}
	
	// 执行搜索
	response, err := hse.esClient.Search(
		hse.esClient.Search.WithContext(ctx),
		hse.esClient.Search.WithIndex(hse.config.IndexName),
		hse.esClient.Search.WithBody(hse.buildSearchBody(query)),
	)
	if err != nil {
		return nil, fmt.Errorf("elasticsearch search failed: %w", err)
	}
	defer response.Body.Close()
	
	// 解析结果
	return hse.parseVectorSearchResults(response)
}

// textSearch 全文搜索
func (hse *HybridSearchEngine) textSearch(ctx context.Context, req *SearchRequest) ([]SearchCandidate, error) {
	// 构建全文搜索查询
	query := map[string]interface{}{
		"size": req.Limit * 2,
		"query": map[string]interface{}{
			"multi_match": map[string]interface{}{
				"query":  req.Query,
				"fields": []string{hse.config.TextField, hse.config.TextField + "^2"},
				"type":   "best_fields",
			},
		},
	}
	
	// 执行搜索
	response, err := hse.esClient.Search(
		hse.esClient.Search.WithContext(ctx),
		hse.esClient.Search.WithIndex(hse.config.IndexName),
		hse.esClient.Search.WithBody(hse.buildSearchBody(query)),
	)
	if err != nil {
		return nil, fmt.Errorf("elasticsearch search failed: %w", err)
	}
	defer response.Body.Close()
	
	// 解析结果
	return hse.parseTextSearchResults(response)
}

// preciseRerank 精确重排序
func (hse *HybridSearchEngine) preciseRerank(ctx context.Context, candidates []SearchCandidate, userInterest *user.UserInterestModel) ([]SearchResult, error) {
	var results []SearchResult
	
	// 限制重排序数量
	if len(candidates) > hse.config.RerankLimit {
		candidates = candidates[:hse.config.RerankLimit]
	}
	
	for _, candidate := range candidates {
		// 使用BGE模型计算精确相似性
		similarity, err := hse.bgeModel.CalculateSimilarity(
			userInterest.InterestVector,
			candidate.Vector,
		)
		if err != nil {
			hse.logger.Warn("Failed to calculate similarity", "candidate_id", candidate.ID, "error", err)
			continue
		}
		
		// 重排序评分
		rerankScore, err := hse.rerankModel.Score(candidate, userInterest)
		if err != nil {
			hse.logger.Warn("Failed to calculate rerank score", "candidate_id", candidate.ID, "error", err)
			continue
		}
		
		// 计算最终分数
		finalScore := hse.calculateFinalScore(similarity, rerankScore, candidate)
		
		results = append(results, SearchResult{
			ID:          candidate.ID,
			Content:     candidate.Content,
			VectorScore: candidate.VectorScore,
			TextScore:   candidate.TextScore,
			RerankScore: rerankScore,
			FinalScore:  finalScore,
			Metadata:    candidate.Metadata,
		})
	}
	
	// 按最终分数排序
	sort.Slice(results, func(i, j int) bool {
		return results[i].FinalScore > results[j].FinalScore
	})
	
	return results, nil
}

// mergeAndDeduplicate 融合和去重
func (hse *HybridSearchEngine) mergeAndDeduplicate(candidates []SearchCandidate) []SearchCandidate {
	seen := make(map[string]bool)
	var merged []SearchCandidate
	
	for _, candidate := range candidates {
		if !seen[candidate.ID] {
			seen[candidate.ID] = true
			merged = append(merged, candidate)
		}
	}
	
	return merged
}

// calculateFinalScore 计算最终分数
func (hse *HybridSearchEngine) calculateFinalScore(similarity, rerankScore float32, candidate SearchCandidate) float32 {
	// 综合多个分数
	vectorWeight := hse.config.VectorWeight
	textWeight := hse.config.TextWeight
	rerankWeight := 1.0 - vectorWeight - textWeight
	
	finalScore := similarity*vectorWeight +
		candidate.TextScore*textWeight +
		rerankScore*rerankWeight
	
	return finalScore
}

// generateCacheKey 生成缓存键
func (hse *HybridSearchEngine) generateCacheKey(req *SearchRequest) string {
	// 简化的缓存键生成，实际应该使用更复杂的哈希算法
	return fmt.Sprintf("%s_%s_%d", req.Query, req.UserInterest.UserID, req.Limit)
}

// buildSearchBody 构建搜索请求体
func (hse *HybridSearchEngine) buildSearchBody(query map[string]interface{}) interface{} {
	// 这里应该将query转换为JSON字符串
	// 简化实现，实际应该使用json.Marshal
	return query
}

// parseVectorSearchResults 解析向量搜索结果
func (hse *HybridSearchEngine) parseVectorSearchResults(response *elasticsearch.Response) ([]SearchCandidate, error) {
	// 解析Elasticsearch响应
	// 简化实现，实际应该解析JSON响应
	return []SearchCandidate{
		{
			ID:          "1",
			Content:     "AI技术发展趋势",
			Vector:      []float32{0.1, 0.2, 0.3},
			VectorScore: 0.95,
			TextScore:   0.0,
			Metadata:    map[string]interface{}{"category": "technology"},
		},
	}, nil
}

// parseTextSearchResults 解析全文搜索结果
func (hse *HybridSearchEngine) parseTextSearchResults(response *elasticsearch.Response) ([]SearchCandidate, error) {
	// 解析Elasticsearch响应
	// 简化实现，实际应该解析JSON响应
	return []SearchCandidate{
		{
			ID:          "2",
			Content:     "机器学习算法应用",
			Vector:      []float32{0.2, 0.3, 0.4},
			VectorScore: 0.0,
			TextScore:   0.88,
			Metadata:    map[string]interface{}{"category": "technology"},
		},
	}, nil
} 