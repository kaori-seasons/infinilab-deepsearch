package tool

import (
	"context"
	"fmt"
	"sync"

	"github.com/coco-ai/research-agent/pkg/logger"
	"github.com/sirupsen/logrus"
)

// Tool 工具接口
type Tool interface {
	GetName() string
	GetDescription() string
	GetParameters() map[string]interface{}
	Execute(ctx context.Context, input map[string]interface{}) (interface{}, error)
}

// Collection 工具集合
type Collection struct {
	tools map[string]Tool
	mu    sync.RWMutex
	logger *logrus.Entry
}

// NewCollection 创建工具集合
func NewCollection() *Collection {
	return &Collection{
		tools: make(map[string]Tool),
		logger: logger.WithField("component", "tool_collection"),
	}
}

// AddTool 添加工具
func (c *Collection) AddTool(tool Tool) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	name := tool.GetName()
	if _, exists := c.tools[name]; exists {
		return fmt.Errorf("tool %s already exists", name)
	}

	c.tools[name] = tool
	c.logger.Info("Tool added", "name", name, "description", tool.GetDescription())
	return nil
}

// GetTool 获取工具
func (c *Collection) GetTool(name string) (Tool, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	tool, exists := c.tools[name]
	if !exists {
		return nil, fmt.Errorf("tool %s not found", name)
	}

	return tool, nil
}

// ListTools 列出所有工具
func (c *Collection) ListTools() []ToolInfo {
	c.mu.RLock()
	defer c.mu.RUnlock()

	tools := make([]ToolInfo, 0, len(c.tools))
	for name, tool := range c.tools {
		tools = append(tools, ToolInfo{
			Name:        name,
			Description: tool.GetDescription(),
			Parameters:  tool.GetParameters(),
		})
	}

	return tools
}

// Execute 执行工具
func (c *Collection) Execute(name string, input map[string]interface{}) (interface{}, error) {
	tool, err := c.GetTool(name)
	if err != nil {
		return nil, err
	}

	c.logger.Info("Executing tool", "name", name, "input", input)
	
	ctx := context.Background()
	result, err := tool.Execute(ctx, input)
	if err != nil {
		c.logger.Error("Tool execution failed", "name", name, "error", err)
		return nil, fmt.Errorf("tool %s execution failed: %w", name, err)
	}

	c.logger.Info("Tool execution completed", "name", name)
	return result, nil
}

// RemoveTool 移除工具
func (c *Collection) RemoveTool(name string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.tools[name]; !exists {
		return fmt.Errorf("tool %s not found", name)
	}

	delete(c.tools, name)
	c.logger.Info("Tool removed", "name", name)
	return nil
}

// ToolInfo 工具信息
type ToolInfo struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

// BaseTool 基础工具实现
type BaseTool struct {
	name        string
	description string
	parameters  map[string]interface{}
	logger      *logrus.Entry
}

// NewBaseTool 创建基础工具
func NewBaseTool(name, description string) *BaseTool {
	return &BaseTool{
		name:        name,
		description: description,
		parameters:  make(map[string]interface{}),
		logger:      logger.WithField("tool", name),
	}
}

// GetName 获取工具名称
func (t *BaseTool) GetName() string {
	return t.name
}

// GetDescription 获取工具描述
func (t *BaseTool) GetDescription() string {
	return t.description
}

// GetParameters 获取工具参数
func (t *BaseTool) GetParameters() map[string]interface{} {
	return t.parameters
}

// SetParameter 设置参数
func (t *BaseTool) SetParameter(key string, value interface{}) {
	t.parameters[key] = value
}

// Execute 执行工具（需要子类实现）
func (t *BaseTool) Execute(ctx context.Context, input map[string]interface{}) (interface{}, error) {
	return nil, fmt.Errorf("execute method not implemented for tool %s", t.name)
}