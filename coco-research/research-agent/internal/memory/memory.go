package memory

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/coco-ai/research-agent/internal/agent"
	"github.com/coco-ai/research-agent/internal/llm"
	"github.com/coco-ai/research-agent/pkg/logger"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/sirupsen/logrus"
)

// Memory 记忆系统实现
type Memory struct {
	// 工作记忆：最近上下文
	workingMemory map[uuid.UUID]*WorkingMemory
	
	// 短期记忆：16个记忆槽
	shortTermMemory map[uuid.UUID]*ShortTermMemory
	
	// 长期记忆：Elasticsearch
	esClient *elasticsearch.Client
	
	// Redis用于缓存
	redisClient *redis.Client
	
	// 嵌入模型客户端
	embeddingClient llm.EmbeddingClient
	
	mu     sync.RWMutex
	logger *logrus.Entry
	config *MemoryConfig
}

// MemoryConfig 记忆系统配置
type MemoryConfig struct {
	RedisHost     string `json:"redis_host"`
	RedisPort     int    `json:"redis_port"`
	RedisPassword string `json:"redis_password"`
	RedisDB       int    `json:"redis_db"`
	
	ESHosts    []string `json:"es_hosts"`
	ESUsername string   `json:"es_username"`
	ESPassword string   `json:"es_password"`
	
	WorkingMemorySize int `json:"working_memory_size"` // 工作记忆大小
	ShortTermSlots    int `json:"short_term_slots"`    // 短期记忆槽数量
	MaxRetrieve       int `json:"max_retrieve"`        // 最大检索数量
	
	// 嵌入模型配置
	EnableVectorSearch bool `json:"enable_vector_search"` // 是否启用向量搜索
}

// WorkingMemory 工作记忆（最近上下文）
type WorkingMemory struct {
	SessionID uuid.UUID
	Messages  []agent.Message
	MaxSize   int
	mu        sync.RWMutex
}

// ShortTermMemory 短期记忆（16个记忆槽）
type ShortTermMemory struct {
	SessionID uuid.UUID
	Slots     []MemorySlot
	MaxSlots  int
	mu        sync.RWMutex
}

// MemorySlot 记忆槽
type MemorySlot struct {
	ID        uuid.UUID `json:"id"`
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	Priority  int       `json:"priority"`  // 优先级 1-10
	AccessCount int     `json:"access_count"` // 访问次数
	LastAccess time.Time `json:"last_access"`
	CreatedAt  time.Time `json:"created_at"`
}

// NewMemory 创建记忆系统
func NewMemory(config *MemoryConfig, embeddingClient llm.EmbeddingClient) *Memory {
	if config == nil {
		config = &MemoryConfig{
			RedisHost:         "localhost",
			RedisPort:         6379,
			RedisPassword:     "",
			RedisDB:           0,
			ESHosts:           []string{"http://localhost:9200"},
			WorkingMemorySize: 20,  // 工作记忆保存最近20条消息
			ShortTermSlots:    16,  // 16个记忆槽
			MaxRetrieve:       10,
			EnableVectorSearch: true, // 默认启用向量搜索
		}
	}

	// 初始化Redis客户端
	redisClient := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", config.RedisHost, config.RedisPort),
		Password: config.RedisPassword,
		DB:       config.RedisDB,
	})

	// 初始化Elasticsearch客户端
	esConfig := elasticsearch.Config{
		Addresses: config.ESHosts,
	}
	if config.ESUsername != "" {
		esConfig.Username = config.ESUsername
		esConfig.Password = config.ESPassword
	}
	
	esClient, err := elasticsearch.NewClient(esConfig)
	if err != nil {
		logger.Error("Failed to create Elasticsearch client", "error", err)
	}

	return &Memory{
		workingMemory:   make(map[uuid.UUID]*WorkingMemory),
		shortTermMemory: make(map[uuid.UUID]*ShortTermMemory),
		esClient:        esClient,
		redisClient:     redisClient,
		embeddingClient: embeddingClient,
		logger:          logger.WithField("component", "memory"),
		config:          config,
	}
}

// Store 存储记忆
func (m *Memory) Store(sessionID uuid.UUID, role string, content string) error {
	// 1. 存储到工作记忆（最近上下文）
	err := m.storeWorkingMemory(sessionID, role, content)
	if err != nil {
		m.logger.Error("Failed to store working memory", "error", err)
	}

	// 2. 存储到短期记忆（16个记忆槽）
	err = m.storeShortTermMemory(sessionID, role, content)
	if err != nil {
		m.logger.Error("Failed to store short-term memory", "error", err)
	}

	// 3. 存储到长期记忆（Elasticsearch）
	err = m.storeLongTermMemory(sessionID, role, content)
	if err != nil {
		m.logger.Error("Failed to store long-term memory", "error", err)
	}

	m.logger.Info("Memory stored", "session_id", sessionID, "role", role)
	return nil
}

// Retrieve 检索记忆
func (m *Memory) Retrieve(sessionID uuid.UUID, query string, limit int) ([]agent.Message, error) {
	if limit <= 0 {
		limit = m.config.MaxRetrieve
	}

	// 1. 从工作记忆检索（最近上下文）
	workingResults := m.retrieveWorkingMemory(sessionID, limit/3)

	// 2. 从短期记忆检索（16个记忆槽）
	shortTermResults := m.retrieveShortTermMemory(sessionID, query, limit/3)

	// 3. 从长期记忆检索（Elasticsearch）
	longTermResults := m.retrieveLongTermMemory(sessionID, query, limit/3)

	// 合并结果
	results := append(workingResults, shortTermResults...)
	results = append(results, longTermResults...)

	// 限制结果数量
	if len(results) > limit {
		results = results[:limit]
	}

	m.logger.Info("Memory retrieved", "session_id", sessionID, "count", len(results))
	return results, nil
}

// Clear 清除记忆
func (m *Memory) Clear(sessionID uuid.UUID) error {
	// 清除工作记忆
	m.clearWorkingMemory(sessionID)

	// 清除短期记忆
	m.clearShortTermMemory(sessionID)

	// 清除长期记忆
	err := m.clearLongTermMemory(sessionID)
	if err != nil {
		m.logger.Error("Failed to clear long-term memory", "error", err)
	}

	m.logger.Info("Memory cleared", "session_id", sessionID)
	return nil
}

// storeWorkingMemory 存储工作记忆
func (m *Memory) storeWorkingMemory(sessionID uuid.UUID, role string, content string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 获取或创建工作记忆
	wm, exists := m.workingMemory[sessionID]
	if !exists {
		wm = &WorkingMemory{
			SessionID: sessionID,
			Messages:  make([]agent.Message, 0),
			MaxSize:   m.config.WorkingMemorySize,
		}
		m.workingMemory[sessionID] = wm
	}

	// 添加新消息
	wm.mu.Lock()
	defer wm.mu.Unlock()

	message := agent.Message{
		Role:      role,
		Content:   content,
		Timestamp: time.Now(),
	}

	wm.Messages = append(wm.Messages, message)

	// 保持工作记忆大小限制
	if len(wm.Messages) > wm.MaxSize {
		wm.Messages = wm.Messages[len(wm.Messages)-wm.MaxSize:]
	}

	return nil
}

// storeShortTermMemory 存储短期记忆
func (m *Memory) storeShortTermMemory(sessionID uuid.UUID, role string, content string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 获取或创建短期记忆
	stm, exists := m.shortTermMemory[sessionID]
	if !exists {
		stm = &ShortTermMemory{
			SessionID: sessionID,
			Slots:     make([]MemorySlot, 0),
			MaxSlots:  m.config.ShortTermSlots,
		}
		m.shortTermMemory[sessionID] = stm
	}

	stm.mu.Lock()
	defer stm.mu.Unlock()

	// 创建新的记忆槽
	slot := MemorySlot{
		ID:         uuid.New(),
		Role:       role,
		Content:    content,
		Priority:   m.calculatePriority(content),
		AccessCount: 0,
		LastAccess:  time.Now(),
		CreatedAt:   time.Now(),
	}

	// 如果记忆槽已满，替换优先级最低的槽
	if len(stm.Slots) >= stm.MaxSlots {
		lowestPriorityIndex := m.findLowestPrioritySlot(stm.Slots)
		stm.Slots[lowestPriorityIndex] = slot
	} else {
		stm.Slots = append(stm.Slots, slot)
	}

	return nil
}

// storeLongTermMemory 存储长期记忆
func (m *Memory) storeLongTermMemory(sessionID uuid.UUID, role string, content string) error {
	ctx := context.Background()

	// 准备索引文档
	doc := map[string]interface{}{
		"session_id": sessionID.String(),
		"role":       role,
		"content":    content,
		"created_at": time.Now(),
	}

	// 如果启用向量搜索，生成嵌入向量
	if m.config.EnableVectorSearch && m.embeddingClient != nil {
		embedding, err := m.embeddingClient.GenerateEmbedding(content)
		if err != nil {
			m.logger.Warn("Failed to generate embedding, storing without vector", "error", err)
		} else {
			doc["embedding"] = embedding
		}
	}

	// 存储到Elasticsearch
	indexName := "memory_items"
	docBytes, err := json.Marshal(doc)
	if err != nil {
		return fmt.Errorf("failed to marshal document: %w", err)
	}
	
	_, err = m.esClient.Index(
		indexName,
		bytes.NewReader(docBytes),
		m.esClient.Index.WithContext(ctx),
		m.esClient.Index.WithDocumentID(uuid.New().String()),
	)
	if err != nil {
		return fmt.Errorf("failed to store in Elasticsearch: %w", err)
	}

	return nil
}

// retrieveWorkingMemory 检索工作记忆
func (m *Memory) retrieveWorkingMemory(sessionID uuid.UUID, limit int) []agent.Message {
	m.mu.RLock()
	defer m.mu.RUnlock()

	wm, exists := m.workingMemory[sessionID]
	if !exists {
		return []agent.Message{}
	}

	wm.mu.RLock()
	defer wm.mu.RUnlock()

	// 返回最近的消息
	if len(wm.Messages) <= limit {
		return wm.Messages
	}
	return wm.Messages[len(wm.Messages)-limit:]
}

// retrieveShortTermMemory 检索短期记忆
func (m *Memory) retrieveShortTermMemory(sessionID uuid.UUID, query string, limit int) []agent.Message {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stm, exists := m.shortTermMemory[sessionID]
	if !exists {
		return []agent.Message{}
	}

	stm.mu.Lock()
	defer stm.mu.Unlock()

	// 按优先级和访问次数排序
	slots := m.sortSlotsByRelevance(stm.Slots, query)

	// 转换为消息格式
	messages := make([]agent.Message, 0, len(slots))
	for _, slot := range slots {
		messages = append(messages, agent.Message{
			Role:      slot.Role,
			Content:   slot.Content,
			Timestamp: slot.CreatedAt,
		})
		
		// 更新访问次数和时间
		slot.AccessCount++
		slot.LastAccess = time.Now()
	}

	// 限制结果数量
	if len(messages) > limit {
		messages = messages[:limit]
	}

	return messages
}

// retrieveLongTermMemory 检索长期记忆
func (m *Memory) retrieveLongTermMemory(sessionID uuid.UUID, query string, limit int) []agent.Message {
	ctx := context.Background()

	var searchQuery map[string]interface{}

	// 如果启用向量搜索且有嵌入客户端，使用向量搜索
	if m.config.EnableVectorSearch && m.embeddingClient != nil && query != "" {
		// 生成查询的嵌入向量
		queryEmbedding, err := m.embeddingClient.GenerateEmbedding(query)
		if err != nil {
			m.logger.Warn("Failed to generate query embedding, falling back to text search", "error", err)
			searchQuery = m.buildTextSearchQuery(sessionID, query, limit)
		} else {
			searchQuery = m.buildVectorSearchQuery(sessionID, queryEmbedding, limit)
		}
	} else {
		// 使用文本搜索
		searchQuery = m.buildTextSearchQuery(sessionID, query, limit)
	}

	// 执行搜索
	searchQueryBytes, err := json.Marshal(searchQuery)
	if err != nil {
		m.logger.Error("Failed to marshal search query", "error", err)
		return []agent.Message{}
	}
	
	res, err := m.esClient.Search(
		m.esClient.Search.WithContext(ctx),
		m.esClient.Search.WithIndex("memory_items"),
		m.esClient.Search.WithBody(bytes.NewReader(searchQueryBytes)),
	)
	if err != nil {
		m.logger.Error("Failed to search long-term memory", "error", err)
		return []agent.Message{}
	}
	defer res.Body.Close()

	// 解析搜索结果
	var searchResult map[string]interface{}
	err = json.NewDecoder(res.Body).Decode(&searchResult)
	if err != nil {
		m.logger.Error("Failed to decode search result", "error", err)
		return []agent.Message{}
	}

	// 提取消息
	messages := make([]agent.Message, 0)
	if hits, ok := searchResult["hits"].(map[string]interface{}); ok {
		if hitsList, ok := hits["hits"].([]interface{}); ok {
			for _, hit := range hitsList {
				if hitMap, ok := hit.(map[string]interface{}); ok {
					if source, ok := hitMap["_source"].(map[string]interface{}); ok {
						message := m.parseMessageFromSource(source)
						if message != nil {
							messages = append(messages, *message)
						}
					}
				}
			}
		}
	}

	return messages
}

// clearWorkingMemory 清除工作记忆
func (m *Memory) clearWorkingMemory(sessionID uuid.UUID) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.workingMemory, sessionID)
}

// clearShortTermMemory 清除短期记忆
func (m *Memory) clearShortTermMemory(sessionID uuid.UUID) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.shortTermMemory, sessionID)
}

// clearLongTermMemory 清除长期记忆
func (m *Memory) clearLongTermMemory(sessionID uuid.UUID) error {
	ctx := context.Background()

	// 构建删除查询
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"term": map[string]interface{}{
				"session_id": sessionID.String(),
			},
		},
	}

	// 执行删除
	queryBytes, err := json.Marshal(query)
	if err != nil {
		return fmt.Errorf("failed to marshal query: %w", err)
	}
	
	// 使用正确的DeleteByQuery API
	res, err := m.esClient.DeleteByQuery(
		[]string{"memory_items"},
		bytes.NewReader(queryBytes),
		m.esClient.DeleteByQuery.WithContext(ctx),
	)
	if err != nil {
		return fmt.Errorf("failed to delete from Elasticsearch: %w", err)
	}
	defer res.Body.Close()

	return nil
}

// calculatePriority 计算记忆优先级
func (m *Memory) calculatePriority(content string) int {
	// 简单的优先级计算逻辑
	// 可以根据内容长度、关键词、重要性等来计算
	priority := 5 // 默认优先级

	if len(content) > 100 {
		priority += 2
	}
	if len(content) > 500 {
		priority += 3
	}

	// 检查是否包含重要关键词
	importantKeywords := []string{"重要", "关键", "核心", "总结", "结论"}
	for _, keyword := range importantKeywords {
		if len(content) > 0 && len(keyword) > 0 {
			// 这里应该实现实际的字符串匹配
			priority += 1
		}
	}

	if priority > 10 {
		priority = 10
	}
	if priority < 1 {
		priority = 1
	}

	return priority
}

// findLowestPrioritySlot 找到优先级最低的记忆槽
func (m *Memory) findLowestPrioritySlot(slots []MemorySlot) int {
	if len(slots) == 0 {
		return 0
	}

	lowestIndex := 0
	lowestPriority := slots[0].Priority
	lowestAccessCount := slots[0].AccessCount

	for i, slot := range slots {
		if slot.Priority < lowestPriority {
			lowestIndex = i
			lowestPriority = slot.Priority
			lowestAccessCount = slot.AccessCount
		} else if slot.Priority == lowestPriority && slot.AccessCount < lowestAccessCount {
			lowestIndex = i
			lowestAccessCount = slot.AccessCount
		}
	}

	return lowestIndex
}

// sortSlotsByRelevance 按相关性排序记忆槽
func (m *Memory) sortSlotsByRelevance(slots []MemorySlot, query string) []MemorySlot {
	// 简单的排序逻辑：按优先级和访问次数排序
	sortedSlots := make([]MemorySlot, len(slots))
	copy(sortedSlots, slots)

	// 这里应该实现更复杂的相关性排序算法
	// 目前按优先级和访问次数排序
	for i := 0; i < len(sortedSlots)-1; i++ {
		for j := i + 1; j < len(sortedSlots); j++ {
			if sortedSlots[i].Priority < sortedSlots[j].Priority ||
				(sortedSlots[i].Priority == sortedSlots[j].Priority &&
					sortedSlots[i].AccessCount < sortedSlots[j].AccessCount) {
				sortedSlots[i], sortedSlots[j] = sortedSlots[j], sortedSlots[i]
			}
		}
	}

	return sortedSlots
}

// parseMessageFromSource 从Elasticsearch源解析消息
func (m *Memory) parseMessageFromSource(source map[string]interface{}) *agent.Message {
	message := &agent.Message{}

	if role, ok := source["role"].(string); ok {
		message.Role = role
	}

	if content, ok := source["content"].(string); ok {
		message.Content = content
	}

	if createdAtStr, ok := source["created_at"].(string); ok {
		if createdAt, err := time.Parse(time.RFC3339, createdAtStr); err == nil {
			message.Timestamp = createdAt
		} else {
			message.Timestamp = time.Now()
		}
	} else {
		message.Timestamp = time.Now()
	}

	return message
}

// buildTextSearchQuery 构建文本搜索查询
func (m *Memory) buildTextSearchQuery(sessionID uuid.UUID, query string, limit int) map[string]interface{} {
	return map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": []map[string]interface{}{
					{
						"term": map[string]interface{}{
							"session_id": sessionID.String(),
						},
					},
				},
				"should": []map[string]interface{}{
					{
						"match": map[string]interface{}{
							"content": query,
						},
					},
				},
			},
		},
		"size": limit,
		"sort": []map[string]interface{}{
			{
				"created_at": map[string]interface{}{
					"order": "desc",
				},
			},
		},
	}
}

// buildVectorSearchQuery 构建向量搜索查询
func (m *Memory) buildVectorSearchQuery(sessionID uuid.UUID, queryEmbedding []float32, limit int) map[string]interface{} {
	return map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": []map[string]interface{}{
					{
						"term": map[string]interface{}{
							"session_id": sessionID.String(),
						},
					},
				},
				"should": []map[string]interface{}{
					{
						"script_score": map[string]interface{}{
							"query": map[string]interface{}{
								"match_all": map[string]interface{}{},
							},
							"script": map[string]interface{}{
								"source": "cosineSimilarity(params.query_vector, 'embedding') + 1.0",
								"params": map[string]interface{}{
									"query_vector": queryEmbedding,
								},
							},
						},
					},
				},
			},
		},
		"size": limit,
		"_source": []string{"session_id", "role", "content", "created_at"},
	}
} 