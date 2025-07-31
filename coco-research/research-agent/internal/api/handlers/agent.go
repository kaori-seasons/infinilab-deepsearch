package handlers

import (
	"net/http"

	"github.com/coco-ai/research-agent/internal/agent"
	"github.com/coco-ai/research-agent/internal/llm"
	"github.com/coco-ai/research-agent/internal/memory"
	"github.com/coco-ai/research-agent/internal/tool"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// AgentHandler 智能体处理器
type AgentHandler struct {
	agentManager *agent.Manager
	llmClient    llm.Client
	memory       *memory.Memory
	toolCollection *tool.Collection
	logger       *logrus.Entry
}

// NewAgentHandler 创建智能体处理器
func NewAgentHandler(
	agentManager *agent.Manager,
	llmClient llm.Client,
	memory *memory.Memory,
	toolCollection *tool.Collection,
) *AgentHandler {
	return &AgentHandler{
		agentManager:   agentManager,
		llmClient:      llmClient,
		memory:         memory,
		toolCollection: toolCollection,
		logger:         logrus.WithField("component", "agent_handler"),
	}
}

// ListAgents 列出所有智能体
func (h *AgentHandler) ListAgents(c *gin.Context) {
	agents := h.agentManager.ListAgents()
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    agents,
		"count":   len(agents),
	})
}

// GetAgent 获取智能体信息
func (h *AgentHandler) GetAgent(c *gin.Context) {
	agentID := c.Param("id")
	
	agent, err := h.agentManager.GetAgent(agentID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"id":          agent.GetName(),
			"name":        agent.GetName(),
			"description": agent.GetDescription(),
			"state":       agent.GetState(),
		},
	})
}

// CreateAgent 创建智能体
func (h *AgentHandler) CreateAgent(c *gin.Context) {
	var request struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
		Type        string `json:"type" binding:"required"`
	}
	
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request body: " + err.Error(),
		})
		return
	}
	
	// 根据类型创建智能体
	var newAgent agent.Agent
	switch request.Type {
	case "research":
		researchAgent := agent.NewResearchAgent()
		researchAgent.LLMClient = h.llmClient
		researchAgent.AvailableTools = h.toolCollection
		researchAgent.Memory = h.memory
		newAgent = researchAgent
	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Unsupported agent type: " + request.Type,
		})
		return
	}
	
	// 注册智能体
	err := h.agentManager.RegisterAgent(newAgent)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to register agent: " + err.Error(),
		})
		return
	}
	
	h.logger.Info("Agent created", "name", request.Name, "type", request.Type)
	
	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data": gin.H{
			"id":          newAgent.GetName(),
			"name":        newAgent.GetName(),
			"description": newAgent.GetDescription(),
			"type":        request.Type,
		},
	})
}

// ExecuteTask 执行智能体任务
func (h *AgentHandler) ExecuteTask(c *gin.Context) {
	agentID := c.Param("id")
	
	var request struct {
		Query string `json:"query" binding:"required"`
	}
	
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request body: " + err.Error(),
		})
		return
	}
	
	// 执行任务
	task, err := h.agentManager.ExecuteTask(c.Request.Context(), agentID, request.Query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to execute task: " + err.Error(),
		})
		return
	}
	
	h.logger.Info("Task executed", "task_id", task.ID, "agent_id", agentID)
	
	c.JSON(http.StatusAccepted, gin.H{
		"success": true,
		"data": gin.H{
			"task_id":  task.ID,
			"agent_id": task.AgentID,
			"status":   task.Status,
			"query":    task.Query,
		},
	})
}

// GetTaskStatus 获取任务状态
func (h *AgentHandler) GetTaskStatus(c *gin.Context) {
	taskIDStr := c.Param("task_id")
	
	taskID, err := uuid.Parse(taskIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid task ID: " + err.Error(),
		})
		return
	}
	
	task, err := h.agentManager.GetTaskStatus(taskID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Task not found: " + err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    task,
	})
}

// StopAgent 停止智能体
func (h *AgentHandler) StopAgent(c *gin.Context) {
	agentID := c.Param("id")
	
	err := h.agentManager.StopAgent(agentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to stop agent: " + err.Error(),
		})
		return
	}
	
	h.logger.Info("Agent stopped", "agent_id", agentID)
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Agent stopped successfully",
	})
}

// GetAgentMetrics 获取智能体指标
func (h *AgentHandler) GetAgentMetrics(c *gin.Context) {
	metrics := h.agentManager.GetAgentMetrics()
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    metrics,
	})
}

// HealthCheck 健康检查
func (h *AgentHandler) HealthCheck(c *gin.Context) {
	health := h.agentManager.HealthCheck()
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    health,
	})
} 