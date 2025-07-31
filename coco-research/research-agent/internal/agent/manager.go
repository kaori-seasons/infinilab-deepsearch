package agent

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// AgentType 智能体类型
type AgentType string

const (
	AgentTypeResearch     AgentType = "research"      // 研究智能体
	AgentTypeReact        AgentType = "react"         // React模式智能体
	AgentTypePlanExecute  AgentType = "plan_execute"  // Plan-and-Execute模式智能体
)

// AgentMode 智能体模式
type AgentMode string

const (
	AgentModeResearch     AgentMode = "research"      // 研究模式
	AgentModeReact        AgentMode = "react"         // React模式
	AgentModePlanExecute  AgentMode = "plan_execute"  // Plan-and-Execute模式
)

// AgentManager 智能体管理器
type AgentManager struct {
	agents map[uuid.UUID]Agent
	mutex  sync.RWMutex
	logger *logrus.Entry
}

// AgentInfo 智能体信息
type AgentInfo struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Type        AgentType `json:"type"`
	Mode        AgentMode `json:"mode"`
	Description string    `json:"description"`
	State       AgentState `json:"state"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CreateAgentRequest 创建智能体请求
type CreateAgentRequest struct {
	Name        string    `json:"name"`
	Type        AgentType `json:"type"`
	Mode        AgentMode `json:"mode"`
	Description string    `json:"description"`
}

// NewAgentManager 创建智能体管理器
func NewAgentManager() *AgentManager {
	return &AgentManager{
		agents: make(map[uuid.UUID]Agent),
		logger: logrus.WithField("component", "agent_manager"),
	}
}

// CreateAgent 创建智能体
func (manager *AgentManager) CreateAgent(req CreateAgentRequest) (*AgentInfo, error) {
	manager.mutex.Lock()
	defer manager.mutex.Unlock()
	
	// 验证请求
	if err := manager.validateCreateRequest(req); err != nil {
		return nil, err
	}
	
	// 创建智能体
	var agent Agent
	
	switch req.Mode {
	case AgentModeReact:
		reactAgent := NewReactAgent()
		reactAgent.Name = req.Name
		reactAgent.Description = req.Description
		agent = reactAgent
		
	case AgentModePlanExecute:
		planAgent := NewPlanExecuteAgent()
		planAgent.Name = req.Name
		planAgent.Description = req.Description
		agent = planAgent
		
	case AgentModeResearch:
		researchAgent := NewResearchAgent()
		researchAgent.Name = req.Name
		researchAgent.Description = req.Description
		agent = researchAgent
		
	default:
		return nil, fmt.Errorf("unsupported agent mode: %s", req.Mode)
	}
	
	// 生成ID
	agentID := uuid.New()
	
	// 存储智能体
	manager.agents[agentID] = agent
	
	// 创建智能体信息
	agentInfo := &AgentInfo{
		ID:          agentID,
		Name:        req.Name,
		Type:        req.Type,
		Mode:        req.Mode,
		Description: req.Description,
		State:       agent.GetState(),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	
	manager.logger.Info("Agent created", "id", agentID, "name", req.Name, "mode", req.Mode)
	
	return agentInfo, nil
}

// GetAgent 获取智能体
func (manager *AgentManager) GetAgent(agentID uuid.UUID) (Agent, error) {
	manager.mutex.RLock()
	defer manager.mutex.RUnlock()
	
	agent, exists := manager.agents[agentID]
	if !exists {
		return nil, fmt.Errorf("agent not found: %s", agentID)
	}
	
	return agent, nil
}

// ListAgents 列出所有智能体
func (manager *AgentManager) ListAgents() []AgentInfo {
	manager.mutex.RLock()
	defer manager.mutex.RUnlock()
	
	var agentInfos []AgentInfo
	
	for id, agent := range manager.agents {
		agentInfo := AgentInfo{
			ID:          id,
			Name:        agent.GetName(),
			Description: agent.GetDescription(),
			State:       agent.GetState(),
			CreatedAt:   time.Now(), // 这里应该从智能体获取创建时间
			UpdatedAt:   time.Now(), // 这里应该从智能体获取更新时间
		}
		
		// 根据智能体类型设置Type和Mode
		switch agent.(type) {
		case *ReactAgent:
			agentInfo.Type = AgentTypeReact
			agentInfo.Mode = AgentModeReact
		case *PlanExecuteAgent:
			agentInfo.Type = AgentTypePlanExecute
			agentInfo.Mode = AgentModePlanExecute
		case *ResearchAgent:
			agentInfo.Type = AgentTypeResearch
			agentInfo.Mode = AgentModeResearch
		default:
			agentInfo.Type = AgentTypeResearch
			agentInfo.Mode = AgentModeResearch
		}
		
		agentInfos = append(agentInfos, agentInfo)
	}
	
	return agentInfos
}

// DeleteAgent 删除智能体
func (manager *AgentManager) DeleteAgent(agentID uuid.UUID) error {
	manager.mutex.Lock()
	defer manager.mutex.Unlock()
	
	if _, exists := manager.agents[agentID]; !exists {
		return fmt.Errorf("agent not found: %s", agentID)
	}
	
	delete(manager.agents, agentID)
	manager.logger.Info("Agent deleted", "id", agentID)
	
	return nil
}

// ExecuteAgent 执行智能体
func (manager *AgentManager) ExecuteAgent(ctx context.Context, agentID uuid.UUID, query string) (string, error) {
	agent, err := manager.GetAgent(agentID)
	if err != nil {
		return "", err
	}
	
	manager.logger.Info("Executing agent", "id", agentID, "query", query)
	
	// 执行智能体
	result, err := agent.Run(ctx, query)
	if err != nil {
		manager.logger.Error("Agent execution failed", "id", agentID, "error", err)
		return "", err
	}
	
	manager.logger.Info("Agent execution completed", "id", agentID)
	
	return result, nil
}

// StepAgent 执行智能体单步
func (manager *AgentManager) StepAgent(ctx context.Context, agentID uuid.UUID) (string, error) {
	agent, err := manager.GetAgent(agentID)
	if err != nil {
		return "", err
	}
	
	manager.logger.Info("Stepping agent", "id", agentID)
	
	// 执行单步
	result, err := agent.Step(ctx)
	if err != nil {
		manager.logger.Error("Agent step failed", "id", agentID, "error", err)
		return "", err
	}
	
	manager.logger.Info("Agent step completed", "id", agentID)
	
	return result, nil
}

// StopAgent 停止智能体
func (manager *AgentManager) StopAgent(agentID uuid.UUID) error {
	_, err := manager.GetAgent(agentID)
	if err != nil {
		return err
	}
	
	// 这里需要实现停止逻辑
	// 目前智能体接口没有Stop方法，需要扩展
	manager.logger.Info("Stopping agent", "id", agentID)
	
	return nil
}

// GetAgentState 获取智能体状态
func (manager *AgentManager) GetAgentState(agentID uuid.UUID) (AgentState, error) {
	agent, err := manager.GetAgent(agentID)
	if err != nil {
		return "", err
	}
	
	return agent.GetState(), nil
}

// SwitchAgentMode 切换智能体模式
func (manager *AgentManager) SwitchAgentMode(agentID uuid.UUID, newMode AgentMode) error {
	manager.mutex.Lock()
	defer manager.mutex.Unlock()
	
	// 检查智能体是否存在
	oldAgent, exists := manager.agents[agentID]
	if !exists {
		return fmt.Errorf("agent not found: %s", agentID)
	}
	
	// 检查当前状态
	if oldAgent.GetState() == AgentStateRunning {
		return fmt.Errorf("cannot switch mode while agent is running")
	}
	
	// 创建新的智能体（保持相同的ID和基本信息）
	var newAgent Agent
	
	switch newMode {
	case AgentModeReact:
		reactAgent := NewReactAgent()
		reactAgent.ID = agentID
		reactAgent.Name = oldAgent.GetName()
		reactAgent.Description = oldAgent.GetDescription()
		newAgent = reactAgent
		
	case AgentModePlanExecute:
		planAgent := NewPlanExecuteAgent()
		planAgent.ID = agentID
		planAgent.Name = oldAgent.GetName()
		planAgent.Description = oldAgent.GetDescription()
		newAgent = planAgent
		
	case AgentModeResearch:
		researchAgent := NewResearchAgent()
		researchAgent.ID = agentID
		researchAgent.Name = oldAgent.GetName()
		researchAgent.Description = oldAgent.GetDescription()
		newAgent = researchAgent
		
	default:
		return fmt.Errorf("unsupported agent mode: %s", newMode)
	}
	
	// 替换智能体
	manager.agents[agentID] = newAgent
	
	manager.logger.Info("Agent mode switched", "id", agentID, "old_mode", getAgentMode(oldAgent), "new_mode", newMode)
	
	return nil
}

// GetAgentMode 获取智能体模式
func getAgentMode(agent Agent) AgentMode {
	switch agent.(type) {
	case *ReactAgent:
		return AgentModeReact
	case *PlanExecuteAgent:
		return AgentModePlanExecute
	case *ResearchAgent:
		return AgentModeResearch
	default:
		return AgentModeResearch
	}
}

// validateCreateRequest 验证创建请求
func (manager *AgentManager) validateCreateRequest(req CreateAgentRequest) error {
	if req.Name == "" {
		return fmt.Errorf("agent name is required")
	}
	
	if req.Description == "" {
		return fmt.Errorf("agent description is required")
	}
	
	// 验证模式
	switch req.Mode {
	case AgentModeReact, AgentModePlanExecute, AgentModeResearch:
		// 有效模式
	default:
		return fmt.Errorf("unsupported agent mode: %s", req.Mode)
	}
	
	// 验证类型
	switch req.Type {
	case AgentTypeResearch, AgentTypeReact, AgentTypePlanExecute:
		// 有效类型
	default:
		return fmt.Errorf("unsupported agent type: %s", req.Type)
	}
	
	return nil
}

// GetAgentStatistics 获取智能体统计信息
func (manager *AgentManager) GetAgentStatistics() map[string]interface{} {
	manager.mutex.RLock()
	defer manager.mutex.RUnlock()
	
	stats := map[string]interface{}{
		"total_agents": len(manager.agents),
		"modes":        make(map[string]int),
		"states":       make(map[string]int),
	}
	
	for _, agent := range manager.agents {
		// 统计模式
		mode := getAgentMode(agent)
		stats["modes"].(map[string]int)[string(mode)]++
		
		// 统计状态
		state := agent.GetState()
		stats["states"].(map[string]int)[string(state)]++
	}
	
	return stats
} 