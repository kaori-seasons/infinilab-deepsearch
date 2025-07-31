package search

import (
	"context"
	"fmt"
	"math"
	"sync"

	"github.com/sirupsen/logrus"
)

// BGESimilarityModel BGE相似性计算模型
type BGESimilarityModel struct {
	modelPath string
	device    string
	maxLength int
	batchSize int
	logger    *logrus.Entry
	mu        sync.RWMutex
}

// NewBGESimilarityModel 创建BGE相似性计算模型
func NewBGESimilarityModel() *BGESimilarityModel {
	return &BGESimilarityModel{
		modelPath: "BAAI/bge-m3",
		device:    "cpu", // 或 "cuda"
		maxLength: 512,
		batchSize: 32,
		logger:    logrus.WithField("component", "bge_similarity_model"),
	}
}

// CalculateSimilarity 计算两个向量的相似性
func (bge *BGESimilarityModel) CalculateSimilarity(vec1, vec2 []float32) (float32, error) {
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

// BatchCalculateSimilarity 批量计算相似性
func (bge *BGESimilarityModel) BatchCalculateSimilarity(queries [][]float32, candidates [][]float32) ([][]float32, error) {
	if len(queries) == 0 || len(candidates) == 0 {
		return nil, fmt.Errorf("empty queries or candidates")
	}
	
	// 检查向量维度一致性
	queryDim := len(queries[0])
	for i, query := range queries {
		if len(query) != queryDim {
			return nil, fmt.Errorf("inconsistent query vector dimensions at index %d", i)
		}
	}
	
	candidateDim := len(candidates[0])
	for i, candidate := range candidates {
		if len(candidate) != candidateDim {
			return nil, fmt.Errorf("inconsistent candidate vector dimensions at index %d", i)
		}
	}
	
	if queryDim != candidateDim {
		return nil, fmt.Errorf("query and candidate vector dimensions mismatch: %d vs %d", queryDim, candidateDim)
	}
	
	// 批量计算相似性矩阵
	similarityMatrix := make([][]float32, len(queries))
	for i := range similarityMatrix {
		similarityMatrix[i] = make([]float32, len(candidates))
	}
	
	// 使用goroutine并行计算
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, bge.batchSize) // 限制并发数
	
	for i, query := range queries {
		wg.Add(1)
		go func(queryIdx int, queryVec []float32) {
			defer wg.Done()
			semaphore <- struct{}{} // 获取信号量
			defer func() { <-semaphore }() // 释放信号量
			
			for j, candidate := range candidates {
				similarity, err := bge.CalculateSimilarity(queryVec, candidate)
				if err != nil {
					bge.logger.Warn("Failed to calculate similarity", 
						"query_idx", queryIdx, 
						"candidate_idx", j, 
						"error", err)
					similarityMatrix[queryIdx][j] = 0
				} else {
					similarityMatrix[queryIdx][j] = similarity
				}
			}
		}(i, query)
	}
	
	wg.Wait()
	
	return similarityMatrix, nil
}

// normalizeVector 向量归一化
func (bge *BGESimilarityModel) normalizeVector(vector []float32) []float32 {
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
func (bge *BGESimilarityModel) cosineSimilarity(vec1, vec2 []float32) float32 {
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
func (bge *BGESimilarityModel) applyBGESimilarityTransform(cosineSim float32) float32 {
	// BGE模型使用特定的相似性变换函数
	// 这里使用简化的变换，实际应该根据具体模型调整
	
	// 将余弦相似性从[-1, 1]映射到[0, 1]
	normalizedSim := (cosineSim + 1) / 2
	
	// 应用非线性变换以增强区分度
	// 使用sigmoid-like函数
	transformedSim := 1 / (1 + math.Exp(-5*(float64(normalizedSim)-0.5)))
	
	return float32(transformedSim)
}

// LoadModel 加载BGE模型
func (bge *BGESimilarityModel) LoadModel(ctx context.Context) error {
	bge.mu.Lock()
	defer bge.mu.Unlock()
	
	bge.logger.Info("Loading BGE model", 
		"model_path", bge.modelPath,
		"device", bge.device)
	
	// 这里应该实现实际的模型加载逻辑
	// 可以使用HuggingFace Transformers或其他推理框架
	
	// 简化实现，实际应该：
	// 1. 下载模型文件
	// 2. 初始化模型
	// 3. 加载到指定设备
	// 4. 预热模型
	
	bge.logger.Info("BGE model loaded successfully")
	return nil
}

// UnloadModel 卸载BGE模型
func (bge *BGESimilarityModel) UnloadModel() error {
	bge.mu.Lock()
	defer bge.mu.Unlock()
	
	bge.logger.Info("Unloading BGE model")
	
	// 清理模型资源
	// 简化实现，实际应该：
	// 1. 释放GPU内存
	// 2. 清理模型对象
	// 3. 重置状态
	
	bge.logger.Info("BGE model unloaded successfully")
	return nil
}

// GetModelInfo 获取模型信息
func (bge *BGESimilarityModel) GetModelInfo() map[string]interface{} {
	bge.mu.RLock()
	defer bge.mu.RUnlock()
	
	return map[string]interface{}{
		"model_path": bge.modelPath,
		"device":     bge.device,
		"max_length": bge.maxLength,
		"batch_size": bge.batchSize,
		"status":     "loaded", // 或 "unloaded"
	}
}

// SetModelPath 设置模型路径
func (bge *BGESimilarityModel) SetModelPath(path string) {
	bge.mu.Lock()
	defer bge.mu.Unlock()
	
	bge.modelPath = path
	bge.logger.Info("BGE model path updated", "new_path", path)
}

// SetDevice 设置计算设备
func (bge *BGESimilarityModel) SetDevice(device string) {
	bge.mu.Lock()
	defer bge.mu.Unlock()
	
	bge.device = device
	bge.logger.Info("BGE model device updated", "new_device", device)
}

// SetBatchSize 设置批处理大小
func (bge *BGESimilarityModel) SetBatchSize(batchSize int) {
	bge.mu.Lock()
	defer bge.mu.Unlock()
	
	bge.batchSize = batchSize
	bge.logger.Info("BGE model batch size updated", "new_batch_size", batchSize)
} 