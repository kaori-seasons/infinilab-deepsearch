package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// PlanExecuteAgent Plan-and-Execute模式智能体
type PlanExecuteAgent struct {
	*BaseAgent
	
	// Plan-and-Execute特定字段
	Plan       *ExecutionPlan
	Executor   *PlanExecutor
	MaxSteps   int
	CurrentStep int
}

// ExecutionPlan 执行计划
type ExecutionPlan struct {
	ID          uuid.UUID    `json:"id"`
	Title       string       `json:"title"`
	Description string       `json:"description"`
	Steps       []PlanStep   `json:"steps"`
	Status      string       `json:"status"` // planning, executing, completed, failed
	Progress    int          `json:"progress"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

// PlanStep 计划步骤
type PlanStep struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	Description  string                 `json:"description"`
	Type         string                 `json:"type"` // tool_call, data_analysis, report_generation
	Tools        []string               `json:"tools"`
	Dependencies []string               `json:"dependencies"`
	Parameters   map[string]interface{} `json:"parameters"`
	Status       string                 `json:"status"` // pending, running, completed, failed
	Result       string                 `json:"result"`
	StartedAt    *time.Time             `json:"started_at,omitempty"`
	CompletedAt  *time.Time             `json:"completed_at,omitempty"`
}

// PlanExecutor 计划执行器
type PlanExecutor struct {
	Plan       *ExecutionPlan
	Agent      *PlanExecuteAgent
	CurrentStep int
}

// PlanResponse 计划响应
type PlanResponse struct {
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Steps       []PlanStep `json:"steps"`
	EstimatedTime string   `json:"estimated_time"`
}

// NewPlanExecuteAgent 创建Plan-and-Execute模式智能体
func NewPlanExecuteAgent() *PlanExecuteAgent {
	baseAgent := NewBaseAgent("PlanExecuteAgent", "基于Plan-and-Execute模式的智能体，先制定计划再执行")
	
	agent := &PlanExecuteAgent{
		BaseAgent:   baseAgent,
		MaxSteps:    50,
		CurrentStep: 0,
	}
	
	// 设置Plan-and-Execute模式的系统提示
	agent.SystemPrompt = getPlanExecuteSystemPrompt()
	
	return agent
}

// Run 运行Plan-and-Execute智能体
func (agent *PlanExecuteAgent) Run(ctx context.Context, query string) (string, error) {
	agent.Logger.Info("Starting Plan-and-Execute agent", "query", query)
	agent.State = AgentStateRunning
	
	// 初始化上下文
	agent.Context.Query = query
	agent.Context.History = []Message{}
	
	// 1. 制定计划阶段
	agent.Logger.Info("Phase 1: Creating execution plan")
	plan, err := agent.createPlan(ctx, query)
	if err != nil {
		agent.Logger.Error("Failed to create plan", "error", err)
		agent.State = AgentStateFailed
		return "", fmt.Errorf("plan creation failed: %w", err)
	}
	
	agent.Plan = plan
	agent.Logger.Info("Plan created successfully", "steps", len(plan.Steps))
	
	// 2. 执行计划阶段
	agent.Logger.Info("Phase 2: Executing plan")
	executor := NewPlanExecutor(plan, agent)
	agent.Executor = executor
	
	result, err := executor.Execute(ctx)
	if err != nil {
		agent.Logger.Error("Plan execution failed", "error", err)
		agent.State = AgentStateFailed
		return "", fmt.Errorf("plan execution failed: %w", err)
	}
	
	agent.Logger.Info("Plan execution completed successfully")
	agent.State = AgentStateSuccess
	
	return result, nil
}

// Step 执行单个步骤
func (agent *PlanExecuteAgent) Step(ctx context.Context) (string, error) {
	if agent.Plan == nil {
		return "", fmt.Errorf("no plan available")
	}
	
	if agent.Executor == nil {
		agent.Executor = NewPlanExecutor(agent.Plan, agent)
	}
	
	return agent.Executor.ExecuteStep(ctx)
}

// createPlan 制定执行计划
func (agent *PlanExecuteAgent) createPlan(ctx context.Context, query string) (*ExecutionPlan, error) {
	// 构建计划制定提示
	prompt := agent.buildPlanPrompt(query)
	
	// 调用LLM制定计划
	response, err := agent.LLMClient.Chat(ctx, []Message{
		{Role: "system", Content: agent.SystemPrompt},
		{Role: "user", Content: prompt},
	}, &LLMOptions{
		Model:       "gpt-4",
		MaxTokens:   2000,
		Temperature: 0.3,
	})
	
	if err != nil {
		return nil, err
	}
	
	// 解析计划响应
	planResponse, err := agent.parsePlanResponse(response)
	if err != nil {
		return nil, err
	}
	
	// 创建执行计划
	plan := &ExecutionPlan{
		ID:          uuid.New(),
		Title:       planResponse.Title,
		Description: planResponse.Description,
		Steps:       planResponse.Steps,
		Status:      "planning",
		Progress:    0,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	
	// 为每个步骤分配ID
	for i := range plan.Steps {
		plan.Steps[i].ID = uuid.New().String()
		plan.Steps[i].Status = "pending"
	}
	
	return plan, nil
}

// buildPlanPrompt 构建计划制定提示
func (agent *PlanExecuteAgent) buildPlanPrompt(query string) string {
	prompt := fmt.Sprintf(`
任务: %s

可用工具:
%s

请为这个任务制定一个详细的执行计划。计划应该包含以下内容：

1. 计划标题和描述
2. 具体的执行步骤，每个步骤应该：
   - 有明确的名称和描述
   - 指定使用的工具
   - 说明步骤之间的依赖关系
   - 提供必要的参数

请以JSON格式回复:
{
  "title": "计划标题",
  "description": "计划描述",
  "steps": [
    {
      "name": "步骤名称",
      "description": "步骤描述",
      "type": "步骤类型",
      "tools": ["工具名称"],
      "dependencies": ["依赖步骤ID"],
      "parameters": {}
    }
  ],
  "estimated_time": "预计执行时间"
}
`, query, agent.formatAvailableTools())
	
	return prompt
}

// parsePlanResponse 解析计划响应
func (agent *PlanExecuteAgent) parsePlanResponse(response string) (*PlanResponse, error) {
	// 尝试解析JSON
	var planResponse PlanResponse
	err := json.Unmarshal([]byte(response), &planResponse)
	if err != nil {
		// 如果不是JSON格式，尝试提取JSON部分
		jsonStart := strings.Index(response, "{")
		jsonEnd := strings.LastIndex(response, "}")
		
		if jsonStart >= 0 && jsonEnd > jsonStart {
			jsonStr := response[jsonStart:jsonEnd+1]
			err = json.Unmarshal([]byte(jsonStr), &planResponse)
			if err != nil {
				return nil, fmt.Errorf("failed to parse plan response: %w", err)
			}
		} else {
			return nil, fmt.Errorf("invalid plan response format")
		}
	}
	
	return &planResponse, nil
}

// formatAvailableTools 格式化可用工具
func (agent *PlanExecuteAgent) formatAvailableTools() string {
	if agent.AvailableTools == nil {
		return "无可用工具"
	}
	
	var tools []string
	for _, tool := range agent.AvailableTools.GetAllTools() {
		tools = append(tools, fmt.Sprintf("- %s: %s", tool.GetName(), tool.GetDescription()))
	}
	
	return strings.Join(tools, "\n")
}

// getPlanExecuteSystemPrompt 获取Plan-and-Execute模式系统提示
func getPlanExecuteSystemPrompt() string {
	return `你是一个基于Plan-and-Execute模式的智能体。这个模式包含两个主要阶段：

1. 计划阶段(Plan)：分析任务，制定详细的执行计划
2. 执行阶段(Execute)：按照计划逐步执行任务

在制定计划时，你应该：
- 仔细分析任务需求
- 将复杂任务分解为可执行的步骤
- 考虑步骤之间的依赖关系
- 为每个步骤分配合适的工具
- 估计执行时间

计划应该：
- 逻辑清晰，步骤明确
- 充分利用可用工具
- 考虑错误处理和回退方案
- 包含必要的参数和配置

请始终以JSON格式回复，包含完整的计划结构。`
}

// NewPlanExecutor 创建计划执行器
func NewPlanExecutor(plan *ExecutionPlan, agent *PlanExecuteAgent) *PlanExecutor {
	return &PlanExecutor{
		Plan:        plan,
		Agent:       agent,
		CurrentStep: 0,
	}
}

// Execute 执行整个计划
func (executor *PlanExecutor) Execute(ctx context.Context) (string, error) {
	executor.Plan.Status = "executing"
	executor.Plan.UpdatedAt = time.Now()
	
	var results []string
	
	// 按顺序执行每个步骤
	for i, step := range executor.Plan.Steps {
		executor.CurrentStep = i
		
		// 检查依赖关系
		if !executor.checkDependencies(step) {
			executor.Agent.Logger.Warn("Step dependencies not met", "step", step.Name)
			continue
		}
		
		// 执行步骤
		result, err := executor.executeStep(ctx, &executor.Plan.Steps[i])
		if err != nil {
			executor.Agent.Logger.Error("Step execution failed", "step", step.Name, "error", err)
			executor.Plan.Steps[i].Status = "failed"
			executor.Plan.Status = "failed"
			return "", fmt.Errorf("step '%s' failed: %w", step.Name, err)
		}
		
		results = append(results, result)
		executor.Plan.Progress = (i + 1) * 100 / len(executor.Plan.Steps)
	}
	
	executor.Plan.Status = "completed"
	executor.Plan.UpdatedAt = time.Now()
	
	return strings.Join(results, "\n\n"), nil
}

// ExecuteStep 执行单个步骤
func (executor *PlanExecutor) ExecuteStep(ctx context.Context) (string, error) {
	if executor.CurrentStep >= len(executor.Plan.Steps) {
		return "所有步骤已完成", nil
	}
	
	step := &executor.Plan.Steps[executor.CurrentStep]
	
	// 检查依赖关系
	if !executor.checkDependencies(*step) {
		return fmt.Sprintf("步骤 '%s' 的依赖未满足", step.Name), nil
	}
	
	// 执行步骤
	result, err := executor.executeStep(ctx, step)
	if err != nil {
		return "", err
	}
	
	executor.CurrentStep++
	return result, nil
}

// executeStep 执行单个步骤
func (executor *PlanExecutor) executeStep(ctx context.Context, step *PlanStep) (string, error) {
	executor.Agent.Logger.Info("Executing step", "step", step.Name, "type", step.Type)
	
	// 更新步骤状态
	step.Status = "running"
	now := time.Now()
	step.StartedAt = &now
	
	// 根据步骤类型执行不同的逻辑
	var result string
	var err error
	
	switch step.Type {
	case "tool_call":
		result, err = executor.executeToolStep(ctx, step)
	case "data_analysis":
		result, err = executor.executeDataAnalysisStep(ctx, step)
	case "report_generation":
		result, err = executor.executeReportGenerationStep(ctx, step)
	default:
		result, err = executor.executeGeneralStep(ctx, step)
	}
	
	if err != nil {
		step.Status = "failed"
		return "", err
	}
	
	// 更新步骤状态
	step.Status = "completed"
	step.Result = result
	completedAt := time.Now()
	step.CompletedAt = &completedAt
	
	executor.Agent.Logger.Info("Step completed", "step", step.Name, "result_length", len(result))
	
	return result, nil
}

// executeToolStep 执行工具调用步骤
func (executor *PlanExecutor) executeToolStep(ctx context.Context, step *PlanStep) (string, error) {
	if len(step.Tools) == 0 {
		return "", fmt.Errorf("no tools specified for step")
	}
	
	toolName := step.Tools[0]
	tool, err := executor.Agent.AvailableTools.GetTool(toolName)
	if err != nil {
		return "", fmt.Errorf("tool not found: %s", toolName)
	}
	
	// 执行工具
	result, err := tool.Execute(ctx, step.Parameters)
	if err != nil {
		return "", fmt.Errorf("tool execution failed: %w", err)
	}
	
	return fmt.Sprintf("工具 '%s' 执行成功: %v", toolName, result), nil
}

// executeDataAnalysisStep 执行数据分析步骤
func (executor *PlanExecutor) executeDataAnalysisStep(ctx context.Context, step *PlanStep) (string, error) {
	// 这里可以调用专门的数据分析工具
	// 目前返回模拟结果
	return "数据分析完成，发现关键趋势和模式", nil
}

// executeReportGenerationStep 执行报告生成步骤
func (executor *PlanExecutor) executeReportGenerationStep(ctx context.Context, step *PlanStep) (string, error) {
	// 这里可以调用报告生成工具
	// 目前返回模拟结果
	return "报告生成完成，包含详细的分析结果和建议", nil
}

// executeGeneralStep 执行通用步骤
func (executor *PlanExecutor) executeGeneralStep(ctx context.Context, step *PlanStep) (string, error) {
	// 对于通用步骤，可以调用LLM进行推理
	prompt := fmt.Sprintf("执行步骤: %s\n描述: %s\n参数: %v", 
		step.Name, step.Description, step.Parameters)
	
	response, err := executor.Agent.LLMClient.Chat(ctx, []Message{
		{Role: "system", Content: "你是一个任务执行助手，请根据描述执行任务。"},
		{Role: "user", Content: prompt},
	}, &LLMOptions{
		Model:       "gpt-4",
		MaxTokens:   500,
		Temperature: 0.3,
	})
	
	if err != nil {
		return "", err
	}
	
	return response, nil
}

// checkDependencies 检查依赖关系
func (executor *PlanExecutor) checkDependencies(step PlanStep) bool {
	if len(step.Dependencies) == 0 {
		return true
	}
	
	for _, depID := range step.Dependencies {
		// 查找依赖步骤
		var depStep *PlanStep
		for i := range executor.Plan.Steps {
			if executor.Plan.Steps[i].ID == depID {
				depStep = &executor.Plan.Steps[i]
				break
			}
		}
		
		// 检查依赖步骤是否完成
		if depStep == nil || depStep.Status != "completed" {
			return false
		}
	}
	
	return true
} 