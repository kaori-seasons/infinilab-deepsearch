package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/coco-ai/research-agent/internal/agent"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// AgentHandler 智能体处理器
type AgentHandler struct {
	agentManager *agent.AgentManager
	logger       *logrus.Entry
}

// NewAgentHandler 创建智能体处理器
func NewAgentHandler(agentManager *agent.AgentManager) *AgentHandler {
	return &AgentHandler{
		agentManager: agentManager,
		logger:       logrus.WithField("component", "agent_handler"),
	}
}

// ListAgents 列出所有智能体
func (h *AgentHandler) ListAgents(c *gin.Context) {
	agents := h.agentManager.ListAgents()
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    agents,
		"message": "智能体列表获取成功",
	})
}

// GetAgent 获取智能体详情
func (h *AgentHandler) GetAgent(c *gin.Context) {
	agentIDStr := c.Param("id")
	agentID, err := uuid.Parse(agentIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "无效的智能体ID",
		})
		return
	}
	
	agent, err := h.agentManager.GetAgent(agentID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "智能体不存在",
		})
		return
	}
	
	// 构建智能体信息
	agentInfo := agent.AgentInfo{
		ID:          agentID,
		Name:        agent.GetName(),
		Description: agent.GetDescription(),
		State:       agent.GetState(),
		CreatedAt:   time.Now(), // 这里应该从智能体获取
		UpdatedAt:   time.Now(), // 这里应该从智能体获取
	}
	
	// 根据智能体类型设置Type和Mode
	switch agent.(type) {
	case *agent.ReactAgent:
		agentInfo.Type = agent.AgentTypeReact
		agentInfo.Mode = agent.AgentModeReact
	case *agent.PlanExecuteAgent:
		agentInfo.Type = agent.AgentTypePlanExecute
		agentInfo.Mode = agent.AgentModePlanExecute
	case *agent.ResearchAgent:
		agentInfo.Type = agent.AgentTypeResearch
		agentInfo.Mode = agent.AgentModeResearch
	default:
		agentInfo.Type = agent.AgentTypeResearch
		agentInfo.Mode = agent.AgentModeResearch
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    agentInfo,
		"message": "智能体详情获取成功",
	})
}

// CreateAgent 创建智能体
func (h *AgentHandler) CreateAgent(c *gin.Context) {
	var req agent.CreateAgentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "请求参数无效",
			"error":   err.Error(),
		})
		return
	}
	
	agentInfo, err := h.agentManager.CreateAgent(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "创建智能体失败",
			"error":   err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    agentInfo,
		"message": "智能体创建成功",
	})
}

// ExecuteAgent 执行智能体
func (h *AgentHandler) ExecuteAgent(c *gin.Context) {
	agentIDStr := c.Param("id")
	agentID, err := uuid.Parse(agentIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "无效的智能体ID",
		})
		return
	}
	
	var req struct {
		Query string `json:"query" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "请求参数无效",
			"error":   err.Error(),
		})
		return
	}
	
	result, err := h.agentManager.ExecuteAgent(c.Request.Context(), agentID, req.Query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "智能体执行失败",
			"error":   err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"result": result,
		},
		"message": "智能体执行成功",
	})
}

// StepAgent 执行智能体单步
func (h *AgentHandler) StepAgent(c *gin.Context) {
	agentIDStr := c.Param("id")
	agentID, err := uuid.Parse(agentIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "无效的智能体ID",
		})
		return
	}
	
	result, err := h.agentManager.StepAgent(c.Request.Context(), agentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "智能体单步执行失败",
			"error":   err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"result": result,
		},
		"message": "智能体单步执行成功",
	})
}

// StopAgent 停止智能体
func (h *AgentHandler) StopAgent(c *gin.Context) {
	agentIDStr := c.Param("id")
	agentID, err := uuid.Parse(agentIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "无效的智能体ID",
		})
		return
	}
	
	err = h.agentManager.StopAgent(agentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "停止智能体失败",
			"error":   err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "智能体已停止",
	})
}

// DeleteAgent 删除智能体
func (h *AgentHandler) DeleteAgent(c *gin.Context) {
	agentIDStr := c.Param("id")
	agentID, err := uuid.Parse(agentIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "无效的智能体ID",
		})
		return
	}
	
	err = h.agentManager.DeleteAgent(agentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "删除智能体失败",
			"error":   err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "智能体删除成功",
	})
}

// SwitchAgentMode 切换智能体模式
func (h *AgentHandler) SwitchAgentMode(c *gin.Context) {
	agentIDStr := c.Param("id")
	agentID, err := uuid.Parse(agentIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "无效的智能体ID",
		})
		return
	}
	
	var req struct {
		Mode agent.AgentMode `json:"mode" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "请求参数无效",
			"error":   err.Error(),
		})
		return
	}
	
	err = h.agentManager.SwitchAgentMode(agentID, req.Mode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "切换智能体模式失败",
			"error":   err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "智能体模式切换成功",
	})
}

// GetAgentState 获取智能体状态
func (h *AgentHandler) GetAgentState(c *gin.Context) {
	agentIDStr := c.Param("id")
	agentID, err := uuid.Parse(agentIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "无效的智能体ID",
		})
		return
	}
	
	state, err := h.agentManager.GetAgentState(agentID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "获取智能体状态失败",
			"error":   err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"state": state,
		},
		"message": "智能体状态获取成功",
	})
}

// GetAgentStatistics 获取智能体统计信息
func (h *AgentHandler) GetAgentStatistics(c *gin.Context) {
	stats := h.agentManager.GetAgentStatistics()
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
		"message": "智能体统计信息获取成功",
	})
} 