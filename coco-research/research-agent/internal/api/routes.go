package api

import (
	"github.com/coco-ai/research-agent/internal/agent"
	"github.com/coco-ai/research-agent/internal/api/handlers"
	"github.com/coco-ai/research-agent/internal/api/middleware"
	"github.com/coco-ai/research-agent/internal/llm"
	"github.com/coco-ai/research-agent/internal/memory"
	"github.com/coco-ai/research-agent/internal/tool"
	"github.com/gin-gonic/gin"
)

// SetupRoutes 设置路由
func SetupRoutes(
	r *gin.Engine,
	agentManager *agent.Manager,
	llmClient llm.Client,
	memory *memory.Memory,
	toolCollection *tool.Collection,
) {
	// 应用中间件
	r.Use(middleware.CORS())
	r.Use(middleware.RequestID())
	r.Use(middleware.Logger())

	// 健康检查
	r.GET("/health", handlers.HealthCheck)

	// API v1 路由组
	v1 := r.Group("/api/v1")
	{
		// 智能体相关路由
		agents := v1.Group("/agents")
		{
			agentHandler := handlers.NewAgentHandler(agentManager, llmClient, memory, toolCollection)
			
			agents.GET("", agentHandler.ListAgents)
			agents.POST("", agentHandler.CreateAgent)
			agents.GET("/:id", agentHandler.GetAgent)
			agents.POST("/:id/execute", agentHandler.ExecuteTask)
			agents.GET("/tasks/:task_id", agentHandler.GetTaskStatus)
			agents.POST("/:id/stop", agentHandler.StopAgent)
			agents.GET("/metrics", agentHandler.GetAgentMetrics)
			agents.GET("/health", agentHandler.HealthCheck)
		}

		// 工具相关路由
		tools := v1.Group("/tools")
		{
			toolHandler := handlers.NewToolHandler(toolCollection)
			
			tools.GET("", toolHandler.ListTools)
			tools.POST("", toolHandler.AddTool)
			tools.GET("/:name", toolHandler.GetTool)
			tools.POST("/:name/execute", toolHandler.ExecuteTool)
			tools.DELETE("/:name", toolHandler.RemoveTool)
		}

		// 会话相关路由
		sessions := v1.Group("/sessions")
		{
			sessions.GET("", handlers.GetSessions)
			sessions.POST("", handlers.CreateSession)
			sessions.GET("/:id", handlers.GetSession)
			sessions.PUT("/:id", handlers.UpdateSession)
			sessions.DELETE("/:id", handlers.DeleteSession)
		}

		// 任务相关路由
		tasks := v1.Group("/tasks")
		{
			tasks.GET("", handlers.GetTasks)
			tasks.POST("", handlers.CreateTask)
			tasks.GET("/:id", handlers.GetTask)
			tasks.PUT("/:id", handlers.UpdateTask)
			tasks.DELETE("/:id", handlers.DeleteTask)
			tasks.POST("/:id/execute", handlers.ExecuteTask)
		}

		// WebSocket 路由
		v1.GET("/ws", handlers.WebSocketHandler)
	}
} 