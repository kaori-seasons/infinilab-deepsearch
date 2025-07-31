package agent

import (
	"context"
	"fmt"
	"time"

	"github.com/coco-ai/research-agent/internal/tool"
	"github.com/coco-ai/research-agent/pkg/logger"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// AgentState 智能体状态
type AgentState string

const (
	AgentStateIdle     AgentState = "idle"
	AgentStateRunning  AgentState = "running"
	AgentStateSuccess  AgentState = "success"
	AgentStateFailed   AgentState = "failed"
	AgentStateStopped  AgentState = "stopped"
)

// Agent 智能体接口
type Agent interface {
	// Run 运行智能体，执行完整的研究任务
	Run(ctx context.Context, query string) (string, error)
	
	// Step 执行单个步骤
	Step(ctx context.Context) (string, error)
	
	// GetState 获取智能体状态
	GetState() AgentState
	
	// GetName 获取智能体名称
	GetName() string
	
	// GetDescription 获取智能体描述
	GetDescription() string
}

// BaseAgent 基础智能体实现
type BaseAgent struct {
	ID          uuid.UUID
	Name        string
	Description string
	State       AgentState
	
	// 配置
	SystemPrompt    string
	NextStepPrompt  string
	MaxSteps        int
	CurrentStep     int
	DuplicateThreshold int
	
	// 工具和记忆
	AvailableTools *tool.Collection
	Memory         Memory
	
	// 上下文
	Context    *AgentContext
	LLMClient  LLMClient
	
	// 输出
	Printer    Printer
	Logger     *logrus.Entry
}

// AgentContext 智能体上下文
type AgentContext struct {
	SessionID uuid.UUID
	TaskID    uuid.UUID
	UserID    string
	Query     string
	History   []Message
	Variables map[string]interface{}
}

// Message 消息结构
type Message struct {
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

// LLMClient LLM客户端接口
type LLMClient interface {
	Chat(ctx context.Context, messages []Message, options *LLMOptions) (string, error)
}

// LLMOptions LLM选项
type LLMOptions struct {
	Model       string  `json:"model"`
	MaxTokens   int     `json:"max_tokens"`
	Temperature float64 `json:"temperature"`
}

// Printer 输出接口
type Printer interface {
	Print(message string)
	PrintError(message string)
	PrintProgress(step int, total int, message string)
}

// Memory 记忆接口
type Memory interface {
	Store(sessionID uuid.UUID, role string, content string) error
	Retrieve(sessionID uuid.UUID, query string, limit int) ([]Message, error)
	Clear(sessionID uuid.UUID) error
}

// NewBaseAgent 创建基础智能体
func NewBaseAgent(name, description string) *BaseAgent {
	return &BaseAgent{
		ID:                 uuid.New(),
		Name:               name,
		Description:        description,
		State:              AgentStateIdle,
		SystemPrompt:       getDefaultSystemPrompt(),
		NextStepPrompt:     getDefaultNextStepPrompt(),
		MaxSteps:           10,
		CurrentStep:        0,
		DuplicateThreshold: 2,
		AvailableTools:     tool.NewCollection(),
		Memory:             nil, // 需要外部设置
		Context:            &AgentContext{},
		Logger:             logger.WithField("agent", name),
	}
}

// Run 运行智能体
func (agent *BaseAgent) Run(ctx context.Context, query string) (string, error) {
	agent.Logger.Info("Starting agent execution", "query", query)
	
	// 设置初始状态
	agent.State = AgentStateRunning
	agent.Context.Query = query
	agent.CurrentStep = 0
	
	// 执行主循环
	for agent.CurrentStep < agent.MaxSteps {
		select {
		case <-ctx.Done():
			agent.State = AgentStateStopped
			return "", ctx.Err()
		default:
			// 执行单个步骤
			result, err := agent.Step(ctx)
			if err != nil {
				agent.State = AgentStateFailed
				agent.Logger.Error("Step execution failed", "error", err)
				return "", err
			}
			
			// 检查是否完成
			if agent.isTaskComplete(result) {
				agent.State = AgentStateSuccess
				agent.Logger.Info("Agent execution completed successfully")
				return result, nil
			}
			
			agent.CurrentStep++
		}
	}
	
	agent.State = AgentStateFailed
	return "", fmt.Errorf("agent exceeded maximum steps (%d)", agent.MaxSteps)
}

// Step 执行单个步骤
func (agent *BaseAgent) Step(ctx context.Context) (string, error) {
	agent.Logger.Debug("Executing step", "step", agent.CurrentStep)
	
	// 构建当前上下文
	messages := agent.buildMessages()
	
	// 调用LLM获取下一步行动
	response, err := agent.LLMClient.Chat(ctx, messages, &LLMOptions{
		Model:       "gpt-4",
		MaxTokens:   2048,
		Temperature: 0.7,
	})
	if err != nil {
		return "", fmt.Errorf("LLM call failed: %w", err)
	}
	
	// 解析响应并执行工具调用
	result, err := agent.executeToolCalls(ctx, response)
	if err != nil {
		return "", fmt.Errorf("tool execution failed: %w", err)
	}
	
	// 更新记忆
	if agent.Memory != nil {
		agent.Memory.Store(agent.Context.SessionID, "assistant", response)
		agent.Memory.Store(agent.Context.SessionID, "system", result)
	}
	
	return result, nil
}

// GetState 获取智能体状态
func (agent *BaseAgent) GetState() AgentState {
	return agent.State
}

// GetName 获取智能体名称
func (agent *BaseAgent) GetName() string {
	return agent.Name
}

// GetDescription 获取智能体描述
func (agent *BaseAgent) GetDescription() string {
	return agent.Description
}

// buildMessages 构建消息列表
func (agent *BaseAgent) buildMessages() []Message {
	messages := []Message{
		{
			Role:      "system",
			Content:   agent.SystemPrompt,
			Timestamp: time.Now(),
		},
	}
	
	// 添加历史消息
	if agent.Context.History != nil {
		messages = append(messages, agent.Context.History...)
	}
	
	// 添加当前查询
	messages = append(messages, Message{
		Role:      "user",
		Content:   agent.Context.Query,
		Timestamp: time.Now(),
	})
	
	return messages
}

// executeToolCalls 执行工具调用
func (agent *BaseAgent) executeToolCalls(ctx context.Context, response string) (string, error) {
	// 解析响应中的工具调用
	toolCalls := agent.parseToolCalls(response)
	
	if len(toolCalls) == 0 {
		return response, nil
	}
	
	// 执行工具调用
	results := make(map[string]interface{})
	for _, toolCall := range toolCalls {
		result, err := agent.AvailableTools.Execute(toolCall.Name, toolCall.Parameters)
		if err != nil {
			agent.Logger.Error("Tool execution failed", "tool", toolCall.Name, "error", err)
			results[toolCall.Name] = fmt.Sprintf("Error: %v", err)
		} else {
			results[toolCall.Name] = result
		}
	}
	
	// 格式化结果
	return agent.formatToolResults(results), nil
}

// parseToolCalls 解析工具调用
func (agent *BaseAgent) parseToolCalls(response string) []ToolCall {
	// 这里实现工具调用解析逻辑
	// 可以根据响应格式进行解析
	return []ToolCall{}
}

// ToolCall 工具调用结构
type ToolCall struct {
	Name       string                 `json:"name"`
	Parameters map[string]interface{} `json:"parameters"`
}

// formatToolResults 格式化工具执行结果
func (agent *BaseAgent) formatToolResults(results map[string]interface{}) string {
	// 格式化工具执行结果
	return fmt.Sprintf("Tool execution results: %v", results)
}

// isTaskComplete 检查任务是否完成
func (agent *BaseAgent) isTaskComplete(result string) bool {
	// 实现任务完成检查逻辑
	return false
}

// getDefaultSystemPrompt 获取默认系统提示
func getDefaultSystemPrompt() string {
	return `你是一个专业的研究智能体，能够帮助用户进行深度研究。

你的能力包括：
1. 网络搜索和信息收集
2. 数据分析和处理
3. 文档生成和报告编写
4. 代码执行和验证

请根据用户的需求，制定研究计划并逐步执行。`
}

// getDefaultNextStepPrompt 获取默认下一步提示
func getDefaultNextStepPrompt() string {
	return `请分析当前情况，确定下一步行动：

1. 如果需要更多信息，请使用搜索工具
2. 如果需要分析数据，请使用数据分析工具
3. 如果需要生成报告，请使用文档生成工具
4. 如果任务已完成，请总结结果

请以JSON格式返回你的决策：
{
  "action": "tool_call",
  "tool": "tool_name",
  "parameters": {...}
}`
} 