package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/coco-ai/research-agent/internal/agent"
	"github.com/coco-ai/research-agent/internal/api"
	"github.com/coco-ai/research-agent/internal/config"
	"github.com/coco-ai/research-agent/internal/database"
	"github.com/coco-ai/research-agent/internal/llm"
	"github.com/coco-ai/research-agent/internal/memory"
	"github.com/coco-ai/research-agent/internal/tool"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func main() {
	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 设置日志级别
	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 初始化日志
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)
	if cfg.Server.Mode == "debug" {
		logger.SetLevel(logrus.DebugLevel)
	}

	// 连接数据库
	db, err := database.Connect(&cfg.Database)
	if err != nil {
		logger.Fatalf("Failed to connect to database: %v", err)
	}
	defer func() {
		sqlDB, err := db.DB()
		if err == nil {
			sqlDB.Close()
		}
	}()

	// 初始化嵌入模型客户端
	embeddingClient, err := llm.NewEmbeddingClient(&cfg.LLM.Embedding)
	if err != nil {
		logger.Warn("Failed to initialize embedding client, vector search will be disabled", "error", err)
		embeddingClient = nil
	} else {
		logger.Info("Embedding client initialized", "model", cfg.LLM.Embedding.Model, "provider", cfg.LLM.Embedding.Provider)
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
		EnableVectorSearch: embeddingClient != nil,
	}
	memorySystem := memory.NewMemory(memoryConfig, embeddingClient)

	// 初始化LLM客户端
	var llmClient llm.Client
	switch cfg.LLM.Provider {
	case "openai":
		apiKey := cfg.LLM.APIKey
		if apiKey == "" {
			apiKey = cfg.LLM.OpenAI.APIKey
		}
		llmClient = llm.NewOpenAIClient(&cfg.LLM.OpenAI)
	case "claude":
		apiKey := cfg.LLM.APIKey
		if apiKey == "" {
			apiKey = cfg.LLM.Claude.APIKey
		}
		llmClient = llm.NewClaudeClient(&cfg.LLM.Claude)
	case "deepseek":
		apiKey := cfg.LLM.APIKey
		if apiKey == "" {
			apiKey = cfg.LLM.DeepSeek.APIKey
		}
		llmClient = llm.NewDeepSeekClient(&cfg.LLM.DeepSeek)
	default:
		logger.Fatalf("Unsupported LLM provider: %s", cfg.LLM.Provider)
	}

	// 初始化工具集合
	toolCollection := tool.NewCollection()

	// 注册默认工具
	registerDefaultTools(toolCollection, cfg)

	// 初始化智能体管理器
	agentManager := agent.NewAgentManager()

	// 创建Gin引擎
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	// 设置路由
	api.SetupRoutes(r, agentManager)

	// 创建HTTP服务器
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Server.Port),
		Handler: r,
	}

	// 启动服务器
	go func() {
		logger.Info("Starting server", "port", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("Failed to start server: %v", err)
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// 优雅关闭
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown:", err)
	}

	logger.Info("Server exited")
}

// registerDefaultTools 注册默认工具
func registerDefaultTools(collection *tool.Collection, cfg *config.Config) {
	// 根据配置的LLM提供商选择API密钥
	var apiKey string
	switch cfg.LLM.Provider {
	case "openai":
		apiKey = cfg.LLM.APIKey
		if apiKey == "" {
			apiKey = cfg.LLM.OpenAI.APIKey
		}
	case "claude":
		apiKey = cfg.LLM.APIKey
		if apiKey == "" {
			apiKey = cfg.LLM.Claude.APIKey
		}
	case "deepseek":
		apiKey = cfg.LLM.APIKey
		if apiKey == "" {
			apiKey = cfg.LLM.DeepSeek.APIKey
		}
	}

	// 注册网络搜索工具
	webSearchTool := tool.NewWebSearchTool(apiKey)
	collection.RegisterTool(webSearchTool)

	// 注册数据分析工具
	dataAnalysisTool := tool.NewDataAnalysisTool()
	collection.RegisterTool(dataAnalysisTool)

	// 注册报告生成工具
	reportGenerationTool := tool.NewReportGenerationTool()
	collection.RegisterTool(reportGenerationTool)

	// 注册竞争对手分析工具
	competitorAnalysisTool := tool.NewCompetitorAnalysisTool()
	collection.RegisterTool(competitorAnalysisTool)

	// 注册趋势分析工具
	trendAnalysisTool := tool.NewTrendAnalysisTool()
	collection.RegisterTool(trendAnalysisTool)
} 