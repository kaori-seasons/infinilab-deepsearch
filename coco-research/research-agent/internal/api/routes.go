package api

import (
	"github.com/coco-ai/research-agent/internal/agent"
	"github.com/coco-ai/research-agent/internal/api/handlers"
	"github.com/gin-gonic/gin"
)

// SetupRoutes 设置路由
func SetupRoutes(r *gin.Engine, agentManager *agent.AgentManager) {
	// API版本
	v1 := r.Group("/api/v1")
	
	// 健康检查
	v1.GET("/health", handlers.HealthCheck)
	
	// 智能体相关路由
	agentHandler := handlers.NewAgentHandler(agentManager)
	agents := v1.Group("/agents")
	{
		agents.GET("", agentHandler.ListAgents)                    // 列出所有智能体
		agents.POST("", agentHandler.CreateAgent)                  // 创建智能体
		agents.GET("/:id", agentHandler.GetAgent)                 // 获取智能体详情
		agents.DELETE("/:id", agentHandler.DeleteAgent)           // 删除智能体
		agents.POST("/:id/execute", agentHandler.ExecuteAgent)    // 执行智能体
		agents.POST("/:id/step", agentHandler.StepAgent)          // 执行智能体单步
		agents.POST("/:id/stop", agentHandler.StopAgent)          // 停止智能体
		agents.PUT("/:id/mode", agentHandler.SwitchAgentMode)     // 切换智能体模式
		agents.GET("/:id/state", agentHandler.GetAgentState)      // 获取智能体状态
		agents.GET("/statistics", agentHandler.GetAgentStatistics) // 获取智能体统计信息
		
		// 增强功能路由
		agents.POST("/enhanced/search", agentHandler.ExecuteEnhancedSearch)           // 执行增强搜索
		agents.GET("/enhanced/user/:user_id/interest", agentHandler.GetUserInterest) // 获取用户兴趣
		agents.PUT("/enhanced/user/:user_id/interest", agentHandler.UpdateUserInterest) // 更新用户兴趣
		agents.GET("/enhanced/search/stats", agentHandler.GetSearchStats)            // 获取搜索统计信息
	}
	
	// 会话相关路由
	sessionHandler := handlers.NewSessionHandler()
	sessions := v1.Group("/sessions")
	{
		sessions.GET("", sessionHandler.ListSessions)           // 列出所有会话
		sessions.POST("", sessionHandler.CreateSession)         // 创建会话
		sessions.GET("/:id", sessionHandler.GetSession)        // 获取会话详情
		sessions.PUT("/:id", sessionHandler.UpdateSession)     // 更新会话
		sessions.DELETE("/:id", sessionHandler.DeleteSession)  // 删除会话
	}
	
	// 任务相关路由
	taskHandler := handlers.NewTaskHandler()
	tasks := v1.Group("/tasks")
	{
		tasks.GET("", taskHandler.GetTasks)                    // 列出所有任务
		tasks.GET("/:id", taskHandler.GetTask)                 // 获取任务详情
		tasks.PUT("/:id", taskHandler.UpdateTask)              // 更新任务
		tasks.DELETE("/:id", taskHandler.DeleteTask)           // 删除任务
	}
	
	// 工具相关路由
	toolHandler := handlers.NewToolHandler()
	tools := v1.Group("/tools")
	{
		tools.GET("", toolHandler.ListTools)                   // 列出所有工具
		tools.GET("/:name", toolHandler.GetTool)               // 获取工具详情
		tools.POST("/:name/execute", toolHandler.ExecuteTool)  // 执行工具
	}
} 