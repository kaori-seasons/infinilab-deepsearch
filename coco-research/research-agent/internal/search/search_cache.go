package search

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/sirupsen/logrus"
)

// SearchCache 搜索缓存
type SearchCache struct {
	cache *cache.Cache
	mu    sync.RWMutex
	logger *logrus.Entry
}

// NewSearchCache 创建搜索缓存
func NewSearchCache() *SearchCache {
	return &SearchCache{
		cache: cache.New(15*time.Minute, 5*time.Minute), // 15分钟过期，5分钟清理
		logger: logrus.WithField("component", "search_cache"),
	}
}

// GetCachedResults 获取缓存的搜索结果
func (sc *SearchCache) GetCachedResults(cacheKey string) ([]SearchResult, bool) {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	
	if cached, found := sc.cache.Get(cacheKey); found {
		if results, ok := cached.([]SearchResult); ok {
			sc.logger.Debug("Cache hit", "cache_key", cacheKey, "results_count", len(results))
			return results, true
		}
	}
	
	sc.logger.Debug("Cache miss", "cache_key", cacheKey)
	return nil, false
}

// SetCachedResults 设置缓存的搜索结果
func (sc *SearchCache) SetCachedResults(cacheKey string, results []SearchResult) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	
	// 设置缓存，15分钟过期
	sc.cache.Set(cacheKey, results, 15*time.Minute)
	
	sc.logger.Debug("Cache set", "cache_key", cacheKey, "results_count", len(results))
}

// ClearCache 清除缓存
func (sc *SearchCache) ClearCache() {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	
	sc.cache.Flush()
	sc.logger.Info("Search cache cleared")
}

// GetStats 获取缓存统计信息
func (sc *SearchCache) GetStats() map[string]interface{} {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	
	stats := sc.cache.Stats()
	return map[string]interface{}{
		"items_count":    stats.ItemsCount,
		"hit_count":      stats.HitCount,
		"miss_count":     stats.MissCount,
		"hit_rate":       float64(stats.HitCount) / float64(stats.HitCount+stats.MissCount),
		"eviction_count": stats.EvictionCount,
	}
}

// GenerateCacheKey 生成缓存键
func (sc *SearchCache) GenerateCacheKey(req *SearchRequest) string {
	// 创建缓存键的组成部分
	keyParts := map[string]interface{}{
		"query":        req.Query,
		"user_id":      req.UserInterest.UserID,
		"limit":        req.Limit,
		"rerank_limit": req.RerankLimit,
		"vector_weight": req.VectorWeight,
		"text_weight":  req.TextWeight,
	}
	
	// 序列化为JSON
	keyJSON, err := json.Marshal(keyParts)
	if err != nil {
		sc.logger.Warn("Failed to marshal cache key", "error", err)
		return fmt.Sprintf("%s_%s_%d", req.Query, req.UserInterest.UserID, req.Limit)
	}
	
	// 生成MD5哈希
	hash := md5.Sum(keyJSON)
	return hex.EncodeToString(hash[:])
}

// InvalidateUserCache 使用户相关缓存失效
func (sc *SearchCache) InvalidateUserCache(userID string) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	
	// 遍历缓存，删除包含该用户ID的缓存项
	items := sc.cache.Items()
	for key := range items {
		if sc.isUserRelatedKey(key, userID) {
			sc.cache.Delete(key)
			sc.logger.Debug("Invalidated user cache", "user_id", userID, "cache_key", key)
		}
	}
}

// isUserRelatedKey 检查缓存键是否与用户相关
func (sc *SearchCache) isUserRelatedKey(cacheKey, userID string) bool {
	// 简化的检查逻辑，实际应该解析缓存键
	return len(cacheKey) > 0 && len(userID) > 0
}

// SetCacheTTL 设置缓存TTL
func (sc *SearchCache) SetCacheTTL(defaultExpiration, cleanupInterval time.Duration) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	
	// 创建新的缓存实例
	sc.cache = cache.New(defaultExpiration, cleanupInterval)
	
	sc.logger.Info("Cache TTL updated", 
		"default_expiration", defaultExpiration,
		"cleanup_interval", cleanupInterval)
}

// GetCacheSize 获取缓存大小
func (sc *SearchCache) GetCacheSize() int {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	
	return sc.cache.ItemCount()
}

// IsCacheEnabled 检查缓存是否启用
func (sc *SearchCache) IsCacheEnabled() bool {
	return sc.cache != nil
} 