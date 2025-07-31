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
	"github.com/coco-ai/research-agent/pkg/logger"
	"github.com/gin-gonic/gin"
)

func main() {
	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 初始化日志
	logConfig := logger.LogConfig{
		Level:  cfg.Log.Level,
		Format: cfg.Log.Format,
		Output: cfg.Log.Output,
	}
	logger.Init(logConfig)

	logger.Info("Starting Coco AI Research Agent", "version", "1.0.0")

	// 初始化数据库连接
	db, err := database.Connect(&cfg.Database)
	if err != nil {
		logger.Fatal("Failed to connect to database", "error", err)
	}
	defer func() {
		sqlDB, err := db.DB()
		if err == nil {
			sqlDB.Close()
		}
	}()

	// 初始化LLM客户端
	var llmClient llm.Client
	switch cfg.LLM.Provider {
	case "openai":
		apiKey := cfg.LLM.APIKey
		if apiKey == "" {
			apiKey = cfg.LLM.OpenAI.APIKey
		}
		llmClient = llm.NewOpenAIClient(apiKey)
	case "claude":
		apiKey := cfg.LLM.APIKey
		if apiKey == "" {
			apiKey = cfg.LLM.Claude.APIKey
		}
		llmClient = llm.NewClaudeClient(apiKey)
	case "deepseek":
		apiKey := cfg.LLM.APIKey
		if apiKey == "" {
			apiKey = cfg.LLM.DeepSeek.APIKey
		}
		llmClient = llm.NewDeepSeekClient(apiKey)
	default:
		logger.Fatal("Unsupported LLM provider", "provider", cfg.LLM.Provider)
	}

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
		EnableVectorSearch: embeddingClient != nil, // 根据嵌入客户端是否可用决定是否启用向量搜索
	}
	memorySystem := memory.NewMemory(memoryConfig, embeddingClient)

	// 初始化工具集合
	toolCollection := tool.NewCollection()

	// 注册默认工具
	registerDefaultTools(toolCollection, cfg)

	// 初始化智能体管理器
	agentConfig := &agent.ManagerConfig{
		MaxConcurrentAgents: 10,
		AgentTimeout:        30 * time.Minute,
		EnableMetrics:       true,
	}
	agentManager := agent.NewManager(agentConfig)

	// 注册默认智能体
	registerDefaultAgents(agentManager, llmClient, memorySystem, toolCollection)

	// 设置Gin模式
	gin.SetMode(gin.ReleaseMode)

	// 创建Gin路由器
	router := gin.New()

	// 设置路由
	api.SetupRoutes(router, agentManager, llmClient, memorySystem, toolCollection)

	// 创建HTTP服务器
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Server.Port),
		Handler: router,
	}

	// 启动服务器
	go func() {
		logger.Info("Starting HTTP server", "port", cfg.Server.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", "error", err)
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

	if err := server.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", "error", err)
	}

	logger.Info("Server exited")
}

// registerDefaultTools 注册默认工具
func registerDefaultTools(toolCollection *tool.Collection, cfg *config.Config) {
	// 注册网络搜索工具
	apiKey := cfg.LLM.APIKey
	if apiKey == "" {
		// 根据提供商选择对应的API Key
		switch cfg.LLM.Provider {
		case "openai":
			apiKey = cfg.LLM.OpenAI.APIKey
		case "claude":
			apiKey = cfg.LLM.Claude.APIKey
		case "deepseek":
			apiKey = cfg.LLM.DeepSeek.APIKey
		}
	}
	webSearchTool := tool.NewWebSearchTool(apiKey) // 使用LLM API Key作为搜索API Key
	toolCollection.AddTool(webSearchTool)

	// 注册数据分析工具
	dataAnalysisTool := tool.NewDataAnalysisTool()
	toolCollection.AddTool(dataAnalysisTool)

	// 注册报告生成工具
	reportGenerationTool := tool.NewReportGenerationTool()
	toolCollection.AddTool(reportGenerationTool)

	// 注册竞争对手分析工具
	competitorAnalysisTool := tool.NewCompetitorAnalysisTool()
	toolCollection.AddTool(competitorAnalysisTool)

	// 注册趋势分析工具
	trendAnalysisTool := tool.NewTrendAnalysisTool()
	toolCollection.AddTool(trendAnalysisTool)

	logger.Info("Default tools registered", "count", 5)
}

// registerDefaultAgents 注册默认智能体
func registerDefaultAgents(
	agentManager *agent.Manager,
	llmClient llm.Client,
	memory *memory.Memory,
	toolCollection *tool.Collection,
) {
	// 创建研究智能体
	researchAgent := agent.NewResearchAgent()
	researchAgent.LLMClient = llmClient
	researchAgent.AvailableTools = toolCollection
	researchAgent.Memory = memory

	// 注册智能体
	err := agentManager.RegisterAgent(researchAgent)
	if err != nil {
		logger.Error("Failed to register research agent", "error", err)
	}

	logger.Info("Default agents registered", "count", 1)
} 