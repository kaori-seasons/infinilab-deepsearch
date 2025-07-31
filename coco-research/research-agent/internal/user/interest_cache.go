package user

import (
	"sync"
	"time"

	"github.com/patrickmn/go-cache"
)

// InterestCache 用户兴趣缓存
type InterestCache struct {
	cache *cache.Cache
	mu    sync.RWMutex
}

// NewInterestCache 创建用户兴趣缓存
func NewInterestCache() *InterestCache {
	return &InterestCache{
		cache: cache.New(30*time.Minute, 10*time.Minute), // 30分钟过期，10分钟清理
	}
}

// GetUserInterest 获取用户兴趣
func (ic *InterestCache) GetUserInterest(userID string) (*UserInterestModel, error) {
	ic.mu.RLock()
	defer ic.mu.RUnlock()
	
	if cached, found := ic.cache.Get(userID); found {
		if interest, ok := cached.(*UserInterestModel); ok {
			return interest, nil
		}
	}
	
	return nil, cache.ErrCacheMiss
}

// SetUserInterest 设置用户兴趣
func (ic *InterestCache) SetUserInterest(userID string, interest *UserInterestModel) {
	ic.mu.Lock()
	defer ic.mu.Unlock()
	
	// 设置缓存，30分钟过期
	ic.cache.Set(userID, interest, 30*time.Minute)
}

// ClearUserInterest 清除用户兴趣缓存
func (ic *InterestCache) ClearUserInterest(userID string) {
	ic.mu.Lock()
	defer ic.mu.Unlock()
	
	ic.cache.Delete(userID)
}

// ClearAll 清除所有缓存
func (ic *InterestCache) ClearAll() {
	ic.mu.Lock()
	defer ic.mu.Unlock()
	
	ic.cache.Flush()
}

// GetStats 获取缓存统计信息
func (ic *InterestCache) GetStats() map[string]interface{} {
	ic.mu.RLock()
	defer ic.mu.RUnlock()
	
	stats := ic.cache.Stats()
	return map[string]interface{}{
		"items_count":    stats.ItemsCount,
		"hit_count":      stats.HitCount,
		"miss_count":     stats.MissCount,
		"hit_rate":       float64(stats.HitCount) / float64(stats.HitCount+stats.MissCount),
		"eviction_count": stats.EvictionCount,
	}
} 