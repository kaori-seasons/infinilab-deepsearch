package handlers

import (
	"net/http"

	"github.com/coco-ai/research-agent/internal/tool"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// ToolHandler 工具处理器
type ToolHandler struct {
	toolCollection *tool.Collection
	logger         *logrus.Entry
}

// NewToolHandler 创建工具处理器
func NewToolHandler(toolCollection *tool.Collection) *ToolHandler {
	return &ToolHandler{
		toolCollection: toolCollection,
		logger:         logrus.WithField("component", "tool_handler"),
	}
}

// ListTools 列出所有工具
func (h *ToolHandler) ListTools(c *gin.Context) {
	tools := h.toolCollection.ListTools()
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    tools,
		"count":   len(tools),
	})
}

// GetTool 获取工具信息
func (h *ToolHandler) GetTool(c *gin.Context) {
	toolName := c.Param("name")
	
	tool, err := h.toolCollection.GetTool(toolName)
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
			"name":        tool.GetName(),
			"description": tool.GetDescription(),
			"parameters":  tool.GetParameters(),
		},
	})
}

// ExecuteTool 执行工具
func (h *ToolHandler) ExecuteTool(c *gin.Context) {
	toolName := c.Param("name")
	
	var request struct {
		Input map[string]interface{} `json:"input" binding:"required"`
	}
	
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request body: " + err.Error(),
		})
		return
	}
	
	// 执行工具
	result, err := h.toolCollection.Execute(toolName, request.Input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Tool execution failed: " + err.Error(),
		})
		return
	}
	
	h.logger.Info("Tool executed", "tool", toolName)
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"tool":   toolName,
			"result": result,
		},
	})
}

// AddTool 添加工具
func (h *ToolHandler) AddTool(c *gin.Context) {
	var request struct {
		Name        string                 `json:"name" binding:"required"`
		Description string                 `json:"description"`
		Parameters  map[string]interface{} `json:"parameters"`
		Type        string                 `json:"type" binding:"required"`
	}
	
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request body: " + err.Error(),
		})
		return
	}
	
	// 根据类型创建工具
	var newTool tool.Tool
	switch request.Type {
	case "web_search":
		apiKey := "your-serper-api-key" // 应该从配置中获取
		newTool = tool.NewWebSearchTool(apiKey)
	case "data_analysis":
		newTool = tool.NewDataAnalysisTool()
	case "report_generation":
		newTool = tool.NewReportGenerationTool()
	case "competitor_analysis":
		newTool = tool.NewCompetitorAnalysisTool()
	case "trend_analysis":
		newTool = tool.NewTrendAnalysisTool()
	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Unsupported tool type: " + request.Type,
		})
		return
	}
	
	// 添加工具
	err := h.toolCollection.AddTool(newTool)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to add tool: " + err.Error(),
		})
		return
	}
	
	h.logger.Info("Tool added", "name", request.Name, "type", request.Type)
	
	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data": gin.H{
			"name":        newTool.GetName(),
			"description": newTool.GetDescription(),
			"type":        request.Type,
		},
	})
}

// RemoveTool 移除工具
func (h *ToolHandler) RemoveTool(c *gin.Context) {
	toolName := c.Param("name")
	
	err := h.toolCollection.RemoveTool(toolName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to remove tool: " + err.Error(),
		})
		return
	}
	
	h.logger.Info("Tool removed", "name", toolName)
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Tool removed successfully",
	})
} 