package main

import (
	"fmt"
	"log"

	"github.com/coco-ai/research-agent/internal/config"
	"github.com/coco-ai/research-agent/internal/llm"
)

func main() {
	// 创建嵌入模型配置
	embeddingConfig := &config.EmbeddingConfig{
		Model:     "iic/nlp_corom_sentence-embedding_chinese-base",
		Provider:  "huggingface",
		APIKey:    "", // 需要设置你的HuggingFace API Key
		BaseURL:   "https://api-inference.huggingface.co",
		MaxLength: 512,
		Dimension: 768,
	}

	// 创建嵌入客户端
	embeddingClient, err := llm.NewEmbeddingClient(embeddingConfig)
	if err != nil {
		log.Fatalf("Failed to create embedding client: %v", err)
	}

	// 测试文本
	testTexts := []string{
		"人工智能是计算机科学的一个分支",
		"机器学习是AI的一个重要组成部分",
		"深度学习使用神经网络进行模式识别",
		"自然语言处理让计算机理解人类语言",
		"计算机视觉使机器能够识别图像",
	}

	fmt.Println("测试嵌入模型功能...")
	fmt.Printf("模型: %s\n", embeddingConfig.Model)
	fmt.Printf("提供商: %s\n", embeddingConfig.Provider)
	fmt.Printf("向量维度: %d\n", embeddingConfig.Dimension)
	fmt.Println()

	// 生成单个文本的嵌入向量
	fmt.Println("1. 测试单个文本嵌入:")
	embedding, err := embeddingClient.GenerateEmbedding(testTexts[0])
	if err != nil {
		fmt.Printf("错误: %v\n", err)
	} else {
		fmt.Printf("文本: %s\n", testTexts[0])
		fmt.Printf("向量维度: %d\n", len(embedding))
		fmt.Printf("向量前5个值: %v\n", embedding[:5])
	}
	fmt.Println()

	// 批量生成嵌入向量
	fmt.Println("2. 测试批量文本嵌入:")
	embeddings, err := embeddingClient.GenerateEmbeddings(testTexts)
	if err != nil {
		fmt.Printf("错误: %v\n", err)
	} else {
		fmt.Printf("成功生成 %d 个文本的嵌入向量\n", len(embeddings))
		for i, emb := range embeddings {
			fmt.Printf("文本 %d: 维度=%d, 前3个值=%v\n", i+1, len(emb), emb[:3])
		}
	}

	fmt.Println("\n嵌入模型测试完成！")
	fmt.Println("注意: 如果看到错误，请确保设置了正确的API密钥")
} 