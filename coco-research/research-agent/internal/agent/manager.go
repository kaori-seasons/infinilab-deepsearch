package agent

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/coco-ai/research-agent/pkg/logger"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// Manager 智能体管理器
type Manager struct {
	agents    map[string]Agent
	mu        sync.RWMutex
	logger    *logrus.Entry
	config    *ManagerConfig
}

// ManagerConfig 管理器配置
type ManagerConfig struct {
	MaxConcurrentAgents int           `json:"max_concurrent_agents"`
	AgentTimeout        time.Duration `json:"agent_timeout"`
	EnableMetrics       bool          `json:"enable_metrics"`
}

// AgentInfo 智能体信息
type AgentInfo struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	State       AgentState `json:"state"`
	CreatedAt   time.Time `json:"created_at"`
	LastActive  time.Time `json:"last_active"`
}

// AgentTask 智能体任务
type AgentTask struct {
	ID        uuid.UUID `json:"id"`
	AgentID   string    `json:"agent_id"`
	Query     string    `json:"query"`
	Status    string    `json:"status"` // pending, running, completed, failed
	Result    string    `json:"result"`
	Error     string    `json:"error"`
	CreatedAt time.Time `json:"created_at"`
	StartedAt *time.Time `json:"started_at"`
	CompletedAt *time.Time `json:"completed_at"`
}

// NewManager 创建智能体管理器
func NewManager(config *ManagerConfig) *Manager {
	if config == nil {
		config = &ManagerConfig{
			MaxConcurrentAgents: 10,
			AgentTimeout:        30 * time.Minute,
			EnableMetrics:       true,
		}
	}

	return &Manager{
		agents: make(map[string]Agent),
		logger: logger.WithField("component", "agent_manager"),
		config: config,
	}
}

// RegisterAgent 注册智能体
func (m *Manager) RegisterAgent(agent Agent) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	agentID := agent.GetName()
	if _, exists := m.agents[agentID]; exists {
		return fmt.Errorf("agent %s already registered", agentID)
	}

	m.agents[agentID] = agent
	m.logger.Info("Agent registered", "agent_id", agentID, "name", agent.GetName())
	return nil
}

// GetAgent 获取智能体
func (m *Manager) GetAgent(agentID string) (Agent, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	agent, exists := m.agents[agentID]
	if !exists {
		return nil, fmt.Errorf("agent %s not found", agentID)
	}

	return agent, nil
}

// ListAgents 列出所有智能体
func (m *Manager) ListAgents() []AgentInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	agents := make([]AgentInfo, 0, len(m.agents))
	for agentID, agent := range m.agents {
		agents = append(agents, AgentInfo{
			ID:          agentID,
			Name:        agent.GetName(),
			Description: agent.GetDescription(),
			State:       agent.GetState(),
			CreatedAt:   time.Now(), // 这里应该从agent中获取创建时间
			LastActive:  time.Now(), // 这里应该从agent中获取最后活跃时间
		})
	}

	return agents
}

// ExecuteTask 执行智能体任务
func (m *Manager) ExecuteTask(ctx context.Context, agentID, query string) (*AgentTask, error) {
	// 获取智能体
	agent, err := m.GetAgent(agentID)
	if err != nil {
		return nil, err
	}

	// 创建任务
	task := &AgentTask{
		ID:        uuid.New(),
		AgentID:   agentID,
		Query:     query,
		Status:    "pending",
		CreatedAt: time.Now(),
	}

	// 检查并发限制
	if m.getActiveTaskCount() >= m.config.MaxConcurrentAgents {
		return nil, fmt.Errorf("maximum concurrent agents reached (%d)", m.config.MaxConcurrentAgents)
	}

	// 异步执行任务
	go m.executeTaskAsync(ctx, task, agent)

	return task, nil
}

// executeTaskAsync 异步执行任务
func (m *Manager) executeTaskAsync(ctx context.Context, task *AgentTask, agent Agent) {
	// 设置任务状态
	task.Status = "running"
	now := time.Now()
	task.StartedAt = &now

	m.logger.Info("Starting agent task", "task_id", task.ID, "agent_id", task.AgentID)

	// 创建带超时的上下文
	timeoutCtx, cancel := context.WithTimeout(ctx, m.config.AgentTimeout)
	defer cancel()

	// 执行智能体
	result, err := agent.Run(timeoutCtx, task.Query)

	// 更新任务状态
	if err != nil {
		task.Status = "failed"
		task.Error = err.Error()
		m.logger.Error("Agent task failed", "task_id", task.ID, "error", err)
	} else {
		task.Status = "completed"
		task.Result = result
		now := time.Now()
		task.CompletedAt = &now
		m.logger.Info("Agent task completed", "task_id", task.ID)
	}
}

// getActiveTaskCount 获取活跃任务数量
func (m *Manager) getActiveTaskCount() int {
	// 这里应该实现实际的活跃任务计数
	// 目前返回一个固定值
	return 0
}

// GetTaskStatus 获取任务状态
func (m *Manager) GetTaskStatus(taskID uuid.UUID) (*AgentTask, error) {
	// 这里应该从存储中获取任务状态
	// 目前返回一个模拟的任务
	return &AgentTask{
		ID:        taskID,
		Status:    "completed",
		CreatedAt: time.Now(),
	}, nil
}

// StopAgent 停止智能体
func (m *Manager) StopAgent(agentID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	_, exists := m.agents[agentID]
	if !exists {
		return fmt.Errorf("agent %s not found", agentID)
	}

	// 这里应该实现智能体的停止逻辑
	m.logger.Info("Agent stopped", "agent_id", agentID)
	return nil
}

// UnregisterAgent 注销智能体
func (m *Manager) UnregisterAgent(agentID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.agents[agentID]; !exists {
		return fmt.Errorf("agent %s not found", agentID)
	}

	delete(m.agents, agentID)
	m.logger.Info("Agent unregistered", "agent_id", agentID)
	return nil
}

// GetAgentMetrics 获取智能体指标
func (m *Manager) GetAgentMetrics() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	metrics := map[string]interface{}{
		"total_agents":     len(m.agents),
		"active_agents":    m.getActiveTaskCount(),
		"max_concurrent":   m.config.MaxConcurrentAgents,
		"agent_timeout":    m.config.AgentTimeout.String(),
	}

	return metrics
}

// HealthCheck 健康检查
func (m *Manager) HealthCheck() map[string]interface{} {
	metrics := m.GetAgentMetrics()
	metrics["status"] = "healthy"
	metrics["timestamp"] = time.Now()
	return metrics
} 