package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/coco-ai/research-agent/internal/agent"
	"github.com/coco-ai/research-agent/internal/user"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// AgentHandler 智能体处理器
type AgentHandler struct {
	agentManager *agent.AgentManager
	logger       *logrus.Entry
}

// EnhancedSearchRequest 增强搜索请求
type EnhancedSearchRequest struct {
	Query    string `json:"query"`
	UserID   string `json:"user_id"`
	Limit    int    `json:"limit"`
	RerankLimit int `json:"rerank_limit"`
}

// UserInterestResponse 用户兴趣响应
type UserInterestResponse struct {
	UserID         string    `json:"user_id"`
	Categories     []string  `json:"categories"`
	Confidence     float32   `json:"confidence"`
	LastUpdated    time.Time `json:"last_updated"`
	InterestVector []float32 `json:"interest_vector"`
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

// ExecuteEnhancedSearch 执行增强搜索
func (h *AgentHandler) ExecuteEnhancedSearch(c *gin.Context) {
	var req EnhancedSearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "请求参数无效",
			"error":   err.Error(),
		})
		return
	}
	
	// 验证请求参数
	if req.Query == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "查询内容不能为空",
		})
		return
	}
	
	if req.UserID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "用户ID不能为空",
		})
		return
	}
	
	// 设置默认值
	if req.Limit == 0 {
		req.Limit = 20
	}
	if req.RerankLimit == 0 {
		req.RerankLimit = 10
	}
	
	// 查找增强智能体
	var enhancedAgent *agent.EnhancedAgent
	agents := h.agentManager.ListAgents()
	for _, agentInfo := range agents {
		if agentInfo.Mode == agent.AgentModeEnhanced {
			agent, err := h.agentManager.GetAgent(agentInfo.ID)
			if err == nil {
				if ea, ok := agent.(*agent.EnhancedAgent); ok {
					enhancedAgent = ea
					break
				}
			}
		}
	}
	
	// 如果没有找到增强智能体，创建一个
	if enhancedAgent == nil {
		h.logger.Info("Creating enhanced agent for enhanced search")
		// 这里应该创建增强智能体，简化实现
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "增强智能体不可用",
		})
		return
	}
	
	// 执行增强搜索
	ctx := c.Request.Context()
	result, err := enhancedAgent.ExecuteWithUserInterest(ctx, req.Query)
	if err != nil {
		h.logger.Error("Enhanced search failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "增强搜索执行失败",
			"error":   err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"query":         result.Query,
			"response":      result.Response,
			"search_results": result.SearchResults,
			"user_interest": result.UserInterest,
			"timestamp":     result.Timestamp,
		},
		"message": "增强搜索执行成功",
	})
}

// GetUserInterest 获取用户兴趣
func (h *AgentHandler) GetUserInterest(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "用户ID不能为空",
		})
		return
	}
	
	// 查找增强智能体
	var enhancedAgent *agent.EnhancedAgent
	agents := h.agentManager.ListAgents()
	for _, agentInfo := range agents {
		if agentInfo.Mode == agent.AgentModeEnhanced {
			agent, err := h.agentManager.GetAgent(agentInfo.ID)
			if err == nil {
				if ea, ok := agent.(*agent.EnhancedAgent); ok {
					enhancedAgent = ea
					break
				}
			}
		}
	}
	
	if enhancedAgent == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "增强智能体不可用",
		})
		return
	}
	
	// 获取用户兴趣
	ctx := c.Request.Context()
	userInterest, err := enhancedAgent.GetUserInterest(ctx)
	if err != nil {
		h.logger.Error("Failed to get user interest", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "获取用户兴趣失败",
			"error":   err.Error(),
		})
		return
	}
	
	response := UserInterestResponse{
		UserID:         userInterest.UserID,
		Categories:     userInterest.Categories,
		Confidence:     userInterest.Confidence,
		LastUpdated:    userInterest.LastUpdated,
		InterestVector: userInterest.InterestVector,
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
		"message": "用户兴趣获取成功",
	})
}

// GetSearchStats 获取搜索统计信息
func (h *AgentHandler) GetSearchStats(c *gin.Context) {
	// 查找增强智能体
	var enhancedAgent *agent.EnhancedAgent
	agents := h.agentManager.ListAgents()
	for _, agentInfo := range agents {
		if agentInfo.Mode == agent.AgentModeEnhanced {
			agent, err := h.agentManager.GetAgent(agentInfo.ID)
			if err == nil {
				if ea, ok := agent.(*agent.EnhancedAgent); ok {
					enhancedAgent = ea
					break
				}
			}
		}
	}
	
	if enhancedAgent == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "增强智能体不可用",
		})
		return
	}
	
	// 获取搜索统计信息
	stats := enhancedAgent.GetSearchStats()
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
		"message": "搜索统计信息获取成功",
	})
}

// UpdateUserInterest 更新用户兴趣
func (h *AgentHandler) UpdateUserInterest(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "用户ID不能为空",
		})
		return
	}
	
	var behavior user.UserBehavior
	if err := c.ShouldBindJSON(&behavior); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "请求参数无效",
			"error":   err.Error(),
		})
		return
	}
	
	// 设置用户ID
	behavior.UserID = userID
	
	// 查找增强智能体
	var enhancedAgent *agent.EnhancedAgent
	agents := h.agentManager.ListAgents()
	for _, agentInfo := range agents {
		if agentInfo.Mode == agent.AgentModeEnhanced {
			agent, err := h.agentManager.GetAgent(agentInfo.ID)
			if err == nil {
				if ea, ok := agent.(*agent.EnhancedAgent); ok {
					enhancedAgent = ea
					break
				}
			}
		}
	}
	
	if enhancedAgent == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "增强智能体不可用",
		})
		return
	}
	
	// 更新用户兴趣
	ctx := c.Request.Context()
	err := enhancedAgent.interestCalculator.UpdateUserInterest(ctx, userID, behavior)
	if err != nil {
		h.logger.Error("Failed to update user interest", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "更新用户兴趣失败",
			"error":   err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "用户兴趣更新成功",
	})
} 