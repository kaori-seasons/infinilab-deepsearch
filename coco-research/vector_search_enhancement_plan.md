# Coco AI Research 向量搜索增强改造方案

## 一、项目背景与需求分析

### 1.1 项目现状
Coco AI Research项目目前已经实现了基础的向量搜索功能：
- 使用`iic/nlp_corom_sentence-embedding_chinese-base`作为嵌入模型
- 支持Elasticsearch作为向量存储
- 实现了React模式和Plan-and-Execute模式智能体
- 具备基础的语义搜索能力

### 1.2 业务需求
基于用户兴趣centroid向量的相似性搜索场景：
1. **用户兴趣建模**：计算用户的兴趣centroid，得到兴趣向量
2. **相似性搜索**：使用兴趣向量快速找到相似的内容/用户
3. **混合检索**：结合向量相似性和全文搜索进行精准匹配

### 1.3 技术挑战
- Elasticsearch目前只支持L1/L2/Cosine Distance
- 不支持BGE等model-based similarity calculation
- 需要更高级的语义相似性计算能力

## 二、技术方案设计

### 2.1 整体架构

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   用户兴趣建模    │    │   向量存储层     │    │   混合检索层     │
│                 │    │                 │    │                 │
│ • 兴趣centroid   │───►│ • Easysearch    │───►│ • 初步过滤      │
│ • 向量生成       │    │ • 阿里云DashScope│    │ • 精确重排序    │
│ • 实时更新       │    │ • 本地Ollama    │    │ • 结果融合      │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                                │
                                ▼
                       ┌─────────────────┐
                       │   模型服务层     │
                       │                 │
                       │ • BGE模型       │
                       │ • 自定义相似性   │
                       │ • 重排序模型     │
                       └─────────────────┘
```

### 2.2 核心组件设计

#### 2.2.1 用户兴趣建模模块
```go
// UserInterestModel 用户兴趣模型
type UserInterestModel struct {
    UserID       string
    InterestVector []float32
    Categories   []string
    LastUpdated  time.Time
    Confidence   float32
}

// InterestCentroidCalculator 兴趣centroid计算器
type InterestCentroidCalculator struct {
    embeddingClient llm.EmbeddingClient
    memory          memory.Memory
    categories      []string
}

func (icc *InterestCentroidCalculator) CalculateUserInterest(userID string) (*UserInterestModel, error) {
    // 1. 收集用户历史行为数据
    userHistory := icc.collectUserHistory(userID)
    
    // 2. 按类别分组
    categorizedData := icc.categorizeData(userHistory)
    
    // 3. 计算每个类别的centroid
    categoryVectors := icc.calculateCategoryCentroids(categorizedData)
    
    // 4. 生成用户兴趣向量
    interestVector := icc.generateInterestVector(categoryVectors)
    
    return &UserInterestModel{
        UserID: userID,
        InterestVector: interestVector,
        Categories: icc.getTopCategories(categoryVectors),
        LastUpdated: time.Now(),
        Confidence: icc.calculateConfidence(categoryVectors),
    }, nil
}
```

#### 2.2.2 混合检索模块
```go
// HybridSearchEngine 混合搜索引擎
type HybridSearchEngine struct {
    esClient        *elasticsearch.Client
    bgeModel        *BGESimilarityModel
    rerankModel     *RerankModel
    config          *SearchConfig
}

// SearchRequest 搜索请求
type SearchRequest struct {
    Query           string
    UserInterest    *UserInterestModel
    Filters         map[string]interface{}
    Limit           int
    RerankLimit     int
}

// SearchResult 搜索结果
type SearchResult struct {
    ID              string
    Content         string
    VectorScore     float32
    TextScore       float32
    RerankScore     float32
    FinalScore      float32
    Metadata        map[string]interface{}
}

func (hse *HybridSearchEngine) Search(ctx context.Context, req *SearchRequest) ([]SearchResult, error) {
    // 第一阶段：初步过滤
    candidates, err := hse.initialFilter(ctx, req)
    if err != nil {
        return nil, err
    }
    
    // 第二阶段：精确重排序
    results, err := hse.preciseRerank(ctx, candidates, req.UserInterest)
    if err != nil {
        return nil, err
    }
    
    return results, nil
}

// initialFilter 初步过滤
func (hse *HybridSearchEngine) initialFilter(ctx context.Context, req *SearchRequest) ([]SearchCandidate, error) {
    // 1. 向量相似性搜索（L2/Cosine）
    vectorResults, err := hse.vectorSearch(ctx, req)
    if err != nil {
        return nil, err
    }
    
    // 2. 全文搜索（BM25）
    textResults, err := hse.textSearch(ctx, req)
    if err != nil {
        return nil, err
    }
    
    // 3. 结果融合
    candidates := hse.mergeResults(vectorResults, textResults)
    
    return candidates, nil
}

// preciseRerank 精确重排序
func (hse *HybridSearchEngine) preciseRerank(ctx context.Context, candidates []SearchCandidate, userInterest *UserInterestModel) ([]SearchResult, error) {
    var results []SearchResult
    
    for _, candidate := range candidates {
        // 使用BGE模型计算精确相似性
        similarity, err := hse.bgeModel.CalculateSimilarity(
            userInterest.InterestVector,
            candidate.Vector,
        )
        if err != nil {
            continue
        }
        
        // 重排序评分
        rerankScore, err := hse.rerankModel.Score(candidate, userInterest)
        if err != nil {
            continue
        }
        
        results = append(results, SearchResult{
            ID:          candidate.ID,
            Content:     candidate.Content,
            VectorScore: candidate.VectorScore,
            TextScore:   candidate.TextScore,
            RerankScore: rerankScore,
            FinalScore:  hse.calculateFinalScore(similarity, rerankScore),
            Metadata:    candidate.Metadata,
        })
    }
    
    // 按最终分数排序
    sort.Slice(results, func(i, j int) bool {
        return results[i].FinalScore > results[j].FinalScore
    })
    
    return results, nil
}
```

#### 2.2.3 BGE相似性计算模块
```go
// BGESimilarityModel BGE相似性计算模型
type BGESimilarityModel struct {
    modelPath       string
    device          string
    maxLength       int
    batchSize       int
}

func (bge *BGESimilarityModel) CalculateSimilarity(vec1, vec2 []float32) (float32, error) {
    // 使用BGE模型计算语义相似性
    // 这里可以集成BGE-M3或BGE-Large等模型
    return bge.computeSemanticSimilarity(vec1, vec2), nil
}

func (bge *BGESimilarityModel) BatchCalculateSimilarity(queries [][]float32, candidates [][]float32) ([][]float32, error) {
    // 批量计算相似性，提高性能
    return bge.batchComputeSimilarity(queries, candidates), nil
}
```

### 2.3 Easysearch集成方案

#### 2.3.1 配置Easysearch与阿里云DashScope
```yaml
# config/easysearch.yml
easysearch:
  embedding:
    dashscope:
      url: "https://dashscope.aliyuncs.com/compatible-mode/v1/embeddings"
      api_key: "${DASHSCOPE_API_KEY}"
      model: "text-embedding-v4"
      dims: 256
      batch_size: 5
    
    ollama:
      url: "http://localhost:11434/api/embed"
      model: "nomic-embed-text:latest"
      dims: 768
  
  search:
    pipeline: "hybrid_search_pipeline"
    rerank_limit: 100
    similarity_threshold: 0.7
```

#### 2.3.2 Ingest Pipeline配置
```json
PUT _ingest/pipeline/user_interest_embedding
{
  "description": "用户兴趣向量生成管道",
  "processors": [
    {
      "text_embedding": {
        "url": "https://dashscope.aliyuncs.com/compatible-mode/v1/embeddings",
        "vendor": "openai",
        "api_key": "${DASHSCOPE_API_KEY}",
        "text_field": "interest_text",
        "vector_field": "interest_vector",
        "model_id": "text-embedding-v4",
        "dims": 256,
        "batch_size": 5
      }
    },
    {
      "set": {
        "field": "user_metadata",
        "value": "{{user_id}}"
      }
    }
  ]
}
```

#### 2.3.3 Search Pipeline配置
```json
PUT /_search/pipeline/hybrid_search_pipeline
{
  "request_processors": [
    {
      "semantic_query_enricher": {
        "tag": "semantic_search",
        "description": "混合搜索语义查询增强",
        "url": "https://dashscope.aliyuncs.com/compatible-mode/v1/embeddings",
        "vendor": "openai",
        "api_key": "${DASHSCOPE_API_KEY}",
        "default_model_id": "text-embedding-v4",
        "vector_field_model_id": {
          "interest_vector": "text-embedding-v4"
        }
      }
    }
  ],
  "response_processors": [
    {
      "rerank": {
        "model": "bge-reranker-v2-m3",
        "field": "content",
        "query_field": "query",
        "window_size": 100
      }
    }
  ]
}
```

## 三、实现计划

### 3.1 第一阶段：基础架构搭建（2周）

#### 3.1.1 用户兴趣建模
- [ ] 实现`UserInterestModel`结构
- [ ] 开发`InterestCentroidCalculator`
- [ ] 集成阿里云DashScope API
- [ ] 实现用户行为数据收集

#### 3.1.2 向量存储优化
- [ ] 配置Easysearch与阿里云集成
- [ ] 实现Ingest Pipeline
- [ ] 优化向量索引结构
- [ ] 添加批量向量处理能力

### 3.2 第二阶段：混合检索实现（3周）

#### 3.2.1 搜索引擎核心
- [ ] 实现`HybridSearchEngine`
- [ ] 开发初步过滤逻辑
- [ ] 实现结果融合算法
- [ ] 添加缓存机制

#### 3.2.2 BGE模型集成
- [ ] 集成BGE-M3模型
- [ ] 实现精确相似性计算
- [ ] 开发批量处理能力
- [ ] 优化推理性能

### 3.3 第三阶段：智能体增强（2周）

#### 3.3.1 智能体改造
- [ ] 在React模式中集成用户兴趣
- [ ] 在Plan-and-Execute模式中添加个性化搜索
- [ ] 实现动态兴趣更新
- [ ] 添加个性化推荐能力

#### 3.3.2 API接口扩展
- [ ] 扩展智能体API支持用户兴趣
- [ ] 添加相似性搜索接口
- [ ] 实现批量搜索能力
- [ ] 添加搜索历史记录

### 3.4 第四阶段：性能优化与测试（1周）

#### 3.4.1 性能优化
- [ ] 向量计算性能优化
- [ ] 缓存策略优化
- [ ] 并发处理优化
- [ ] 内存使用优化

#### 3.4.2 测试与验证
- [ ] 单元测试覆盖
- [ ] 集成测试
- [ ] 性能基准测试
- [ ] 用户验收测试

## 四、技术细节

### 4.1 向量相似性计算策略

#### 4.1.1 初步过滤阶段
```go
// 使用L2距离进行快速过滤
func (hse *HybridSearchEngine) vectorSearch(ctx context.Context, req *SearchRequest) ([]SearchCandidate, error) {
    query := map[string]interface{}{
        "size": req.Limit * 2, // 获取更多候选，为后续重排序准备
        "query": map[string]interface{}{
            "knn": map[string]interface{}{
                "interest_vector": map[string]interface{}{
                    "vector": req.UserInterest.InterestVector,
                    "k": req.Limit * 2,
                    "num_candidates": req.Limit * 4,
                },
            },
        },
    }
    
    // 执行向量搜索
    return hse.executeVectorSearch(ctx, query)
}
```

#### 4.1.2 精确重排序阶段
```go
// 使用BGE模型进行精确相似性计算
func (bge *BGESimilarityModel) computeSemanticSimilarity(vec1, vec2 []float32) float32 {
    // 1. 向量归一化
    normalizedVec1 := bge.normalizeVector(vec1)
    normalizedVec2 := bge.normalizeVector(vec2)
    
    // 2. 计算余弦相似性
    cosineSim := bge.cosineSimilarity(normalizedVec1, normalizedVec2)
    
    // 3. 应用BGE特定的相似性变换
    return bge.applyBGESimilarityTransform(cosineSim)
}
```

### 4.2 缓存策略

#### 4.2.1 用户兴趣缓存
```go
type InterestCache struct {
    cache    *cache.Cache
    ttl      time.Duration
}

func (ic *InterestCache) GetUserInterest(userID string) (*UserInterestModel, error) {
    // 从缓存获取用户兴趣
    if cached, found := ic.cache.Get(userID); found {
        return cached.(*UserInterestModel), nil
    }
    
    // 缓存未命中，重新计算
    interest, err := ic.calculateUserInterest(userID)
    if err != nil {
        return nil, err
    }
    
    // 存入缓存
    ic.cache.Set(userID, interest, ic.ttl)
    return interest, nil
}
```

#### 4.2.2 搜索结果缓存
```go
type SearchCache struct {
    cache    *cache.Cache
    ttl      time.Duration
}

func (sc *SearchCache) GetCachedResults(cacheKey string) ([]SearchResult, bool) {
    if cached, found := sc.cache.Get(cacheKey); found {
        return cached.([]SearchResult), true
    }
    return nil, false
}
```

### 4.3 监控与指标

#### 4.3.1 性能指标
```go
type SearchMetrics struct {
    VectorSearchLatency    prometheus.Histogram
    RerankLatency         prometheus.Histogram
    CacheHitRate          prometheus.Counter
    UserInterestAccuracy  prometheus.Gauge
    SearchThroughput      prometheus.Counter
}
```

#### 4.3.2 业务指标
```go
type BusinessMetrics struct {
    UserEngagement        prometheus.Counter
    SearchSatisfaction    prometheus.Gauge
    InterestUpdateRate    prometheus.Counter
    SimilarityAccuracy    prometheus.Gauge
}
```

## 五、部署方案

### 5.1 环境配置

#### 5.1.1 开发环境
```yaml
# docker-compose.dev.yml
version: '3.8'
services:
  easysearch:
    image: infinilabs/easysearch:latest
    ports:
      - "9200:9200"
    environment:
      - ES_JAVA_OPTS=-Xms1g -Xmx1g
    volumes:
      - ./config/easysearch.yml:/usr/share/easysearch/config/easysearch.yml
  
  ollama:
    image: ollama/ollama:latest
    ports:
      - "11434:11434"
    volumes:
      - ollama_data:/root/.ollama
  
  research-agent:
    build: .
    ports:
      - "8080:8080"
    environment:
      - EASYSEARCH_URL=http://easysearch:9200
      - OLLAMA_URL=http://ollama:11434
      - DASHSCOPE_API_KEY=${DASHSCOPE_API_KEY}
    depends_on:
      - easysearch
      - ollama

volumes:
  ollama_data:
```

#### 5.1.2 生产环境
```yaml
# kubernetes/production.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: research-agent
spec:
  replicas: 3
  selector:
    matchLabels:
      app: research-agent
  template:
    metadata:
      labels:
        app: research-agent
    spec:
      containers:
      - name: research-agent
        image: coco-ai/research-agent:latest
        ports:
        - containerPort: 8080
        env:
        - name: EASYSEARCH_URL
          value: "http://easysearch-cluster:9200"
        - name: DASHSCOPE_API_KEY
          valueFrom:
            secretKeyRef:
              name: dashscope-secret
              key: api-key
        resources:
          requests:
            memory: "2Gi"
            cpu: "500m"
          limits:
            memory: "4Gi"
            cpu: "1000m"
```

### 5.2 监控配置

#### 5.2.1 Prometheus配置
```yaml
# monitoring/prometheus.yml
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: 'research-agent'
    static_configs:
      - targets: ['research-agent:8080']
    metrics_path: '/metrics'
    scrape_interval: 5s
```

#### 5.2.2 Grafana仪表板
```json
{
  "dashboard": {
    "title": "Coco AI Research Metrics",
    "panels": [
      {
        "title": "Search Latency",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(vector_search_duration_seconds[5m])",
            "legendFormat": "Vector Search"
          },
          {
            "expr": "rate(rerank_duration_seconds[5m])",
            "legendFormat": "Rerank"
          }
        ]
      },
      {
        "title": "Cache Hit Rate",
        "type": "stat",
        "targets": [
          {
            "expr": "rate(cache_hits_total[5m]) / rate(cache_requests_total[5m])",
            "legendFormat": "Hit Rate"
          }
        ]
      }
    ]
  }
}
```

## 六、风险评估与应对

### 6.1 技术风险

#### 6.1.1 向量计算性能风险
- **风险**：BGE模型推理速度慢，影响用户体验
- **应对**：
  - 实现批量推理，提高GPU利用率
  - 使用模型量化技术，减少内存占用
  - 实现智能缓存，避免重复计算

#### 6.1.2 数据一致性风险
- **风险**：用户兴趣更新不及时，影响搜索准确性
- **应对**：
  - 实现增量更新机制
  - 添加数据版本控制
  - 实现回滚机制

### 6.2 业务风险

#### 6.2.1 成本控制风险
- **风险**：阿里云API调用成本过高
- **应对**：
  - 实现智能缓存策略
  - 使用本地Ollama作为备选
  - 监控API使用量，设置告警

#### 6.2.2 用户体验风险
- **风险**：搜索延迟增加，影响用户满意度
- **应对**：
  - 实现异步搜索机制
  - 添加搜索进度指示
  - 提供搜索建议功能

## 七、总结

本改造方案通过集成Easysearch与阿里云DashScope，结合BGE模型的高级相似性计算能力，为Coco AI Research项目构建了一个强大的混合检索系统。

### 7.1 核心优势
1. **高性能**：通过初步过滤+精确重排序的两阶段策略，平衡了性能和准确性
2. **高准确性**：集成BGE模型，提供更精确的语义相似性计算
3. **可扩展性**：支持多种嵌入模型，可根据需求灵活切换
4. **成本效益**：通过缓存和批量处理，有效控制API调用成本

### 7.2 技术亮点
1. **用户兴趣建模**：基于用户行为数据，动态计算兴趣centroid
2. **混合检索策略**：结合向量搜索和全文搜索，提供更全面的检索能力
3. **智能缓存机制**：多层次缓存策略，显著提升响应速度
4. **监控体系**：完善的指标监控，支持性能优化和问题诊断

### 7.3 后续规划
1. **模型优化**：持续优化BGE模型性能，探索更高效的推理方案
2. **功能扩展**：支持更多类型的用户兴趣建模和个性化推荐
3. **生态集成**：与更多AI服务集成，构建更完整的智能搜索生态

通过本方案的实施，Coco AI Research项目将具备业界领先的向量搜索能力，为用户提供更精准、更个性化的智能研究体验。

---

**参考资源：**
- [Easysearch集成阿里云与Ollama Embedding API](https://www.infinilabs.cn/blog/2025/Easysearch-Integration-with-Alibaba-CloudOllama-Embedding-API/)
- [BGE模型官方文档](https://github.com/FlagOpen/FlagEmbedding)
- [阿里云DashScope文档](https://help.aliyun.com/zh/dashscope/)
- [Ollama官方文档](https://ollama.ai/) 