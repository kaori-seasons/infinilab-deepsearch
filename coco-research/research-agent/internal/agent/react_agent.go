package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// ReactAgent React模式智能体
type ReactAgent struct {
	*BaseAgent
	
	// React特定字段
	Thoughts     []Thought
	Actions      []Action
	MaxThoughts  int
	MaxActions   int
	CurrentCycle int
}

// Thought 思考过程
type Thought struct {
	ID        string    `json:"id"`
	Content   string    `json:"content"`
	Reasoning string    `json:"reasoning"`
	Timestamp time.Time `json:"timestamp"`
}

// Action 行动
type Action struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"` // tool_call, observation, conclusion
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters,omitempty"`
	Result      string                 `json:"result,omitempty"`
	Timestamp   time.Time              `json:"timestamp"`
}

// ReactResponse React模式的LLM响应
type ReactResponse struct {
	Thought    string                 `json:"thought,omitempty"`
	Action     string                 `json:"action,omitempty"`
	ToolCalls  []ToolCall            `json:"tool_calls,omitempty"`
	Observation string                `json:"observation,omitempty"`
	Conclusion  string                `json:"conclusion,omitempty"`
	ShouldStop  bool                  `json:"should_stop"`
}

// NewReactAgent 创建React模式智能体
func NewReactAgent() *ReactAgent {
	baseAgent := NewBaseAgent("ReactAgent", "基于React模式的智能体，能够思考-行动-观察循环")
	
	agent := &ReactAgent{
		BaseAgent:    baseAgent,
		Thoughts:     []Thought{},
		Actions:      []Action{},
		MaxThoughts:  10,
		MaxActions:   20,
		CurrentCycle: 0,
	}
	
	// 设置React模式的系统提示
	agent.SystemPrompt = getReactSystemPrompt()
	
	return agent
}

// Run 运行React智能体
func (agent *ReactAgent) Run(ctx context.Context, query string) (string, error) {
	agent.Logger.Info("Starting React agent", "query", query)
	agent.State = AgentStateRunning
	
	// 初始化上下文
	agent.Context.Query = query
	agent.Context.History = []Message{}
	
	// 开始React循环
	for agent.CurrentCycle < agent.MaxActions {
		agent.CurrentCycle++
		
		// 执行一个React步骤
		result, err := agent.Step(ctx)
		if err != nil {
			agent.Logger.Error("React step failed", "error", err)
			agent.State = AgentStateFailed
			return "", err
		}
		
		// 检查是否应该停止
		if agent.shouldStop(result) {
			agent.Logger.Info("React agent completed", "cycles", agent.CurrentCycle)
			agent.State = AgentStateSuccess
			return agent.generateFinalResult(), nil
		}
		
		// 检查是否达到最大循环次数
		if agent.CurrentCycle >= agent.MaxActions {
			agent.Logger.Warn("React agent reached max cycles", "max_cycles", agent.MaxActions)
			agent.State = AgentStateSuccess
			return agent.generateFinalResult(), nil
		}
	}
	
	agent.State = AgentStateSuccess
	return agent.generateFinalResult(), nil
}

// Step 执行单个React步骤
func (agent *ReactAgent) Step(ctx context.Context) (string, error) {
	agent.Logger.Debug("Executing React step", "cycle", agent.CurrentCycle)
	
	// 1. 思考阶段
	thought, shouldAct, err := agent.think(ctx)
	if err != nil {
		return "", fmt.Errorf("thinking failed: %w", err)
	}
	
	// 记录思考过程
	if thought != "" {
		agent.addThought(thought)
	}
	
	// 2. 如果不需要行动，返回观察结果
	if !shouldAct {
		return "思考完成，无需进一步行动", nil
	}
	
	// 3. 行动阶段
	action, err := agent.act(ctx)
	if err != nil {
		return "", fmt.Errorf("action failed: %w", err)
	}
	
	// 记录行动
	if action != nil {
		agent.addAction(*action)
	}
	
	return fmt.Sprintf("执行行动: %s", action.Description), nil
}

// think 思考过程
func (agent *ReactAgent) think(ctx context.Context) (string, bool, error) {
	// 构建思考提示
	prompt := agent.buildThinkPrompt()
	
	// 调用LLM进行思考
	response, err := agent.LLMClient.Chat(ctx, []Message{
		{Role: "system", Content: agent.SystemPrompt},
		{Role: "user", Content: prompt},
	}, &LLMOptions{
		Model:       "gpt-4",
		MaxTokens:   500,
		Temperature: 0.7,
	})
	
	if err != nil {
		return "", false, err
	}
	
	// 解析响应
	reactResponse, err := agent.parseReactResponse(response)
	if err != nil {
		return "", false, err
	}
	
	return reactResponse.Thought, reactResponse.Action != "", nil
}

// act 执行行动
func (agent *ReactAgent) act(ctx context.Context) (*Action, error) {
	// 构建行动提示
	prompt := agent.buildActPrompt()
	
	// 调用LLM决定行动
	response, err := agent.LLMClient.Chat(ctx, []Message{
		{Role: "system", Content: agent.SystemPrompt},
		{Role: "user", Content: prompt},
	}, &LLMOptions{
		Model:       "gpt-4",
		MaxTokens:   1000,
		Temperature: 0.3,
	})
	
	if err != nil {
		return nil, err
	}
	
	// 解析响应
	reactResponse, err := agent.parseReactResponse(response)
	if err != nil {
		return nil, err
	}
	
	// 执行工具调用
	if len(reactResponse.ToolCalls) > 0 {
		return agent.executeToolCalls(ctx, reactResponse.ToolCalls)
	}
	
	// 创建观察行动
	if reactResponse.Observation != "" {
		action := &Action{
			ID:          uuid.New().String(),
			Type:        "observation",
			Description: reactResponse.Observation,
			Timestamp:   time.Now(),
		}
		return action, nil
	}
	
	// 创建结论行动
	if reactResponse.Conclusion != "" {
		action := &Action{
			ID:          uuid.New().String(),
			Type:        "conclusion",
			Description: reactResponse.Conclusion,
			Timestamp:   time.Now(),
		}
		return action, nil
	}
	
	return nil, fmt.Errorf("no valid action found in response")
}

// executeToolCalls 执行工具调用
func (agent *ReactAgent) executeToolCalls(ctx context.Context, toolCalls []ToolCall) (*Action, error) {
	if len(toolCalls) == 0 {
		return nil, fmt.Errorf("no tool calls to execute")
	}
	
	// 执行第一个工具调用
	toolCall := toolCalls[0]
	
	// 查找工具
	tool, err := agent.AvailableTools.GetTool(toolCall.Name)
	if err != nil {
		return nil, fmt.Errorf("tool not found: %s", toolCall.Name)
	}
	
	// 执行工具
	result, err := tool.Execute(ctx, toolCall.Parameters)
	if err != nil {
		return nil, fmt.Errorf("tool execution failed: %w", err)
	}
	
	// 创建工具调用行动
	action := &Action{
		ID:          uuid.New().String(),
		Type:        "tool_call",
		Description: fmt.Sprintf("调用工具: %s", toolCall.Name),
		Parameters:  toolCall.Parameters,
		Result:      fmt.Sprintf("%v", result),
		Timestamp:   time.Now(),
	}
	
	return action, nil
}

// buildThinkPrompt 构建思考提示
func (agent *ReactAgent) buildThinkPrompt() string {
	prompt := fmt.Sprintf(`
当前任务: %s

历史思考过程:
%s

历史行动:
%s

请思考下一步应该做什么。如果需要执行工具或观察，请说明原因。
如果任务已经完成，请说明结论。

请以JSON格式回复:
{
  "thought": "你的思考过程",
  "action": "下一步行动描述",
  "should_stop": false
}
`, agent.Context.Query, agent.formatThoughts(), agent.formatActions())
	
	return prompt
}

// buildActPrompt 构建行动提示
func (agent *ReactAgent) buildActPrompt() string {
	prompt := fmt.Sprintf(`
当前任务: %s

可用工具:
%s

历史行动:
%s

请决定具体的行动。如果需要调用工具，请提供工具名称和参数。
如果只是观察或得出结论，请直接说明。

请以JSON格式回复:
{
  "action": "行动描述",
  "tool_calls": [{"name": "工具名", "parameters": {}}],
  "observation": "观察结果",
  "conclusion": "结论"
}
`, agent.Context.Query, agent.formatAvailableTools(), agent.formatActions())
	
	return prompt
}

// parseReactResponse 解析React响应
func (agent *ReactAgent) parseReactResponse(response string) (*ReactResponse, error) {
	// 尝试解析JSON
	var reactResponse ReactResponse
	err := json.Unmarshal([]byte(response), &reactResponse)
	if err != nil {
		// 如果不是JSON格式，尝试提取JSON部分
		jsonStart := strings.Index(response, "{")
		jsonEnd := strings.LastIndex(response, "}")
		
		if jsonStart >= 0 && jsonEnd > jsonStart {
			jsonStr := response[jsonStart:jsonEnd+1]
			err = json.Unmarshal([]byte(jsonStr), &reactResponse)
			if err != nil {
				return nil, fmt.Errorf("failed to parse response: %w", err)
			}
		} else {
			// 如果不是JSON格式，创建简单的响应
			reactResponse = ReactResponse{
				Thought: response,
				Action:  "",
			}
		}
	}
	
	return &reactResponse, nil
}

// shouldStop 检查是否应该停止
func (agent *ReactAgent) shouldStop(result string) bool {
	// 检查结果中是否包含停止信号
	stopKeywords := []string{"任务完成", "已完成", "完成", "结论", "总结"}
	for _, keyword := range stopKeywords {
		if strings.Contains(result, keyword) {
			return true
		}
	}
	
	return false
}

// addThought 添加思考
func (agent *ReactAgent) addThought(content string) {
	thought := Thought{
		ID:        uuid.New().String(),
		Content:   content,
		Timestamp: time.Now(),
	}
	agent.Thoughts = append(agent.Thoughts, thought)
}

// addAction 添加行动
func (agent *ReactAgent) addAction(action Action) {
	agent.Actions = append(agent.Actions, action)
}

// formatThoughts 格式化思考历史
func (agent *ReactAgent) formatThoughts() string {
	if len(agent.Thoughts) == 0 {
		return "无"
	}
	
	var thoughts []string
	for _, thought := range agent.Thoughts {
		thoughts = append(thoughts, fmt.Sprintf("- %s", thought.Content))
	}
	
	return strings.Join(thoughts, "\n")
}

// formatActions 格式化行动历史
func (agent *ReactAgent) formatActions() string {
	if len(agent.Actions) == 0 {
		return "无"
	}
	
	var actions []string
	for _, action := range agent.Actions {
		actions = append(actions, fmt.Sprintf("- %s", action.Description))
	}
	
	return strings.Join(actions, "\n")
}

// formatAvailableTools 格式化可用工具
func (agent *ReactAgent) formatAvailableTools() string {
	if agent.AvailableTools == nil {
		return "无可用工具"
	}
	
	var tools []string
	for _, tool := range agent.AvailableTools.GetAllTools() {
		tools = append(tools, fmt.Sprintf("- %s: %s", tool.GetName(), tool.GetDescription()))
	}
	
	return strings.Join(tools, "\n")
}

// generateFinalResult 生成最终结果
func (agent *ReactAgent) generateFinalResult() string {
	var results []string
	
	// 添加结论
	for _, action := range agent.Actions {
		if action.Type == "conclusion" {
			results = append(results, action.Description)
		}
	}
	
	// 如果没有结论，使用最后一个行动
	if len(results) == 0 && len(agent.Actions) > 0 {
		lastAction := agent.Actions[len(agent.Actions)-1]
		results = append(results, lastAction.Description)
	}
	
	if len(results) == 0 {
		return "任务执行完成"
	}
	
	return strings.Join(results, "\n")
}

// getReactSystemPrompt 获取React模式系统提示
func getReactSystemPrompt() string {
	return `你是一个基于React模式的智能体。React模式包含以下步骤：

1. 思考(Think)：分析当前情况，决定下一步行动
2. 行动(Act)：执行具体的行动，可能是调用工具或观察
3. 观察(Observe)：观察行动的结果
4. 重复：继续思考-行动-观察的循环

你的目标是完成用户的任务。在每次思考时，你应该：
- 分析当前的任务状态
- 考虑可用的工具
- 决定是否需要调用工具或得出结论
- 提供清晰的推理过程

在行动时，你应该：
- 明确指定要调用的工具和参数
- 或者提供观察结果
- 或者给出最终结论

请始终以JSON格式回复，包含必要的字段。`
} 