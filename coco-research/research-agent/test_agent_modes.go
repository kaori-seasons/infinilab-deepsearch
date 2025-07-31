package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/google/uuid"
	"github.com/coco-ai/research-agent/internal/agent"
	"github.com/coco-ai/research-agent/internal/config"
	"github.com/coco-ai/research-agent/internal/llm"
	"github.com/coco-ai/research-agent/internal/tool"
	"github.com/sirupsen/logrus"
)

func main() {
	// 加载配置
	_, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 初始化日志
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	// 初始化LLM客户端（模拟）
	llmClient := &MockLLMClient{}

	// 初始化工具集合
	toolCollection := tool.NewCollection()
	registerTestTools(toolCollection)

	// 初始化记忆系统（模拟）
	memorySystem := &MockMemory{}

	// 测试React模式智能体
	fmt.Println("=== 测试React模式智能体 ===")
	testReactAgent(llmClient, toolCollection, memorySystem)

	// 测试Plan-and-Execute模式智能体
	fmt.Println("\n=== 测试Plan-and-Execute模式智能体 ===")
	testPlanExecuteAgent(llmClient, toolCollection, memorySystem)

	// 测试智能体管理器
	fmt.Println("\n=== 测试智能体管理器 ===")
	testAgentManager(llmClient, toolCollection, memorySystem)
}

// testReactAgent 测试React模式智能体
func testReactAgent(llmClient llm.Client, toolCollection *tool.Collection, memory agent.Memory) {
	// 创建React智能体
	reactAgent := agent.NewReactAgent()
	reactAgent.LLMClient = llmClient
	reactAgent.AvailableTools = toolCollection
	reactAgent.Memory = memory

	// 设置测试查询
	query := "分析当前AI市场的发展趋势"

	fmt.Printf("测试查询: %s\n", query)

	// 执行智能体
	ctx := context.Background()
	result, err := reactAgent.Run(ctx, query)
	if err != nil {
		fmt.Printf("React智能体执行失败: %v\n", err)
		return
	}

	fmt.Printf("React智能体执行结果: %s\n", result)
	fmt.Printf("思考次数: %d\n", len(reactAgent.Thoughts))
	fmt.Printf("行动次数: %d\n", len(reactAgent.Actions))
}

// testPlanExecuteAgent 测试Plan-and-Execute模式智能体
func testPlanExecuteAgent(llmClient llm.Client, toolCollection *tool.Collection, memory agent.Memory) {
	// 创建Plan-and-Execute智能体
	planAgent := agent.NewPlanExecuteAgent()
	planAgent.LLMClient = llmClient
	planAgent.AvailableTools = toolCollection
	planAgent.Memory = memory

	// 设置测试查询
	query := "研究竞争对手的产品策略"

	fmt.Printf("测试查询: %s\n", query)

	// 执行智能体
	ctx := context.Background()
	result, err := planAgent.Run(ctx, query)
	if err != nil {
		fmt.Printf("Plan-and-Execute智能体执行失败: %v\n", err)
		return
	}

	fmt.Printf("Plan-and-Execute智能体执行结果: %s\n", result)
	if planAgent.Plan != nil {
		fmt.Printf("计划步骤数: %d\n", len(planAgent.Plan.Steps))
		fmt.Printf("计划状态: %s\n", planAgent.Plan.Status)
		fmt.Printf("计划进度: %d%%\n", planAgent.Plan.Progress)
	}
}

// testAgentManager 测试智能体管理器
func testAgentManager(llmClient llm.Client, toolCollection *tool.Collection, memory agent.Memory) {
	// 创建智能体管理器
	manager := agent.NewAgentManager()

	// 创建React模式智能体
	reactReq := agent.CreateAgentRequest{
		Name:        "React研究助手",
		Type:        agent.AgentTypeReact,
		Mode:        agent.AgentModeReact,
		Description: "基于React模式的研究智能体",
	}

	reactInfo, err := manager.CreateAgent(reactReq)
	if err != nil {
		fmt.Printf("创建React智能体失败: %v\n", err)
		return
	}

	fmt.Printf("创建React智能体成功: %s\n", reactInfo.ID)

	// 创建Plan-and-Execute模式智能体
	planReq := agent.CreateAgentRequest{
		Name:        "计划执行助手",
		Type:        agent.AgentTypePlanExecute,
		Mode:        agent.AgentModePlanExecute,
		Description: "基于Plan-and-Execute模式的研究智能体",
	}

	planInfo, err := manager.CreateAgent(planReq)
	if err != nil {
		fmt.Printf("创建Plan-and-Execute智能体失败: %v\n", err)
		return
	}

	fmt.Printf("创建Plan-and-Execute智能体成功: %s\n", planInfo.ID)

	// 列出所有智能体
	agents := manager.ListAgents()
	fmt.Printf("智能体总数: %d\n", len(agents))

	for _, agent := range agents {
		fmt.Printf("- %s (%s模式): %s\n", agent.Name, agent.Mode, agent.Description)
	}

	// 获取统计信息
	stats := manager.GetAgentStatistics()
	fmt.Printf("智能体统计: %+v\n", stats)
}

// registerTestTools 注册测试工具
func registerTestTools(collection *tool.Collection) {
	// 注册模拟工具
	webSearchTool := tool.NewWebSearchTool("test-api-key")
	collection.AddTool(webSearchTool)

	dataAnalysisTool := tool.NewDataAnalysisTool()
	collection.AddTool(dataAnalysisTool)

	reportGenerationTool := tool.NewReportGenerationTool()
	collection.AddTool(reportGenerationTool)
}

// MockLLMClient 模拟LLM客户端
type MockLLMClient struct{}

func (m *MockLLMClient) Chat(ctx context.Context, messages []llm.Message, options *llm.LLMOptions) (string, error) {
	lastMessage := messages[len(messages)-1].Content

	// 针对Plan-and-Execute的计划制定
	if strings.Contains(lastMessage, "制定一个详细的执行计划") || strings.Contains(lastMessage, "请为这个任务制定一个详细的执行计划") {
		return `{
			"title": "竞争对手产品策略研究计划",
			"description": "系统性研究主要竞争对手的产品策略，分阶段执行。",
			"steps": [
				{
					"name": "信息收集",
					"description": "收集竞争对手产品信息",
					"type": "tool_call",
					"tools": ["web_search"],
					"dependencies": [],
					"parameters": {"query": "竞争对手产品信息"}
				},
				{
					"name": "数据分析",
					"description": "分析收集到的数据",
					"type": "data_analysis",
					"tools": ["data_analysis"],
					"dependencies": [],
					"parameters": {"data": "collected_data"}
				},
				{
					"name": "报告生成",
					"description": "生成研究报告",
					"type": "report_generation",
					"tools": ["report_generation"],
					"dependencies": [],
					"parameters": {"analysis": "analysis_result"}
				}
			],
			"estimated_time": "2小时"
		}`, nil
	}

	if len(messages) == 1 {
		// 系统消息，返回思考
		return `{
			"thought": "我需要分析AI市场的发展趋势，首先应该搜索相关信息",
			"action": "搜索AI市场信息",
			"should_stop": false
		}`, nil
	}

	if strings.Contains(lastMessage, "思考") {
		return `{
			"thought": "基于搜索结果，我发现AI市场正在快速增长",
			"action": "分析数据",
			"should_stop": false
		}`, nil
	}

	if strings.Contains(lastMessage, "行动") {
		return `{
			"action": "调用数据分析工具",
			"tool_calls": [{"name": "data_analysis", "parameters": {"data": "market_data"}}],
			"observation": "",
			"conclusion": ""
		}`, nil
	}

	// 默认响应
	return "任务完成，已生成分析报告", nil
}

func (m *MockLLMClient) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	// 模拟嵌入生成
	return []float32{0.1, 0.2, 0.3}, nil
}

func (m *MockLLMClient) GenerateEmbeddings(ctx context.Context, texts []string) ([][]float32, error) {
	// 模拟批量嵌入生成
	embeddings := make([][]float32, len(texts))
	for i := range texts {
		embeddings[i] = []float32{0.1, 0.2, 0.3}
	}
	return embeddings, nil
}

// MockMemory 模拟记忆系统
type MockMemory struct{}

func (m *MockMemory) Store(sessionID uuid.UUID, role string, content string) error {
	return nil
}

func (m *MockMemory) Retrieve(sessionID uuid.UUID, query string, limit int) ([]agent.Message, error) {
	return []agent.Message{}, nil
}

func (m *MockMemory) Clear(sessionID uuid.UUID) error {
	return nil
} 