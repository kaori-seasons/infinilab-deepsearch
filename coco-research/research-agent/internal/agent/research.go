package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// ResearchAgent 研究智能体
type ResearchAgent struct {
	*BaseAgent
	
	// 研究特定配置
	ResearchPlan    *ResearchPlan
	CurrentPhase    ResearchPhase
	SearchResults   []SearchResult
	AnalysisResults []AnalysisResult
	ReportData      *ReportData
}

// ResearchPlan 研究计划
type ResearchPlan struct {
	ID          uuid.UUID `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Phases      []ResearchPhase `json:"phases"`
	CreatedAt   time.Time `json:"created_at"`
}

// ResearchPhase 研究阶段
type ResearchPhase struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Status      string `json:"status"` // pending, running, completed, failed
	Tools       []string `json:"tools"`
	Output      string `json:"output"`
}

// SearchResult 搜索结果
type SearchResult struct {
	ID          uuid.UUID `json:"id"`
	Query       string    `json:"query"`
	URL         string    `json:"url"`
	Title       string    `json:"title"`
	Content     string    `json:"content"`
	Relevance   float64   `json:"relevance"`
	Timestamp   time.Time `json:"timestamp"`
}

// AnalysisResult 分析结果
type AnalysisResult struct {
	ID          uuid.UUID `json:"id"`
	Type        string    `json:"type"` // data_analysis, sentiment_analysis, trend_analysis
	Data        string    `json:"data"`
	Insights    []string  `json:"insights"`
	Visualization string  `json:"visualization"`
	Timestamp   time.Time `json:"timestamp"`
}

// ReportData 报告数据
type ReportData struct {
	Title       string    `json:"title"`
	Summary     string    `json:"summary"`
	KeyFindings []string  `json:"key_findings"`
	DataSources []string  `json:"data_sources"`
	Methodology string    `json:"methodology"`
	Conclusion  string    `json:"conclusion"`
	Recommendations []string `json:"recommendations"`
}

// NewResearchAgent 创建研究智能体
func NewResearchAgent() *ResearchAgent {
	baseAgent := NewBaseAgent("ResearchAgent", "专业的研究智能体，能够进行深度研究和分析")
	
	agent := &ResearchAgent{
		BaseAgent: baseAgent,
		ResearchPlan: &ResearchPlan{
			ID:        uuid.New(),
			CreatedAt: time.Now(),
		},
		CurrentPhase: ResearchPhase{},
		SearchResults: []SearchResult{},
		AnalysisResults: []AnalysisResult{},
		ReportData: &ReportData{},
	}
	
	// 设置研究特定的系统提示
	agent.SystemPrompt = getResearchSystemPrompt()
	
	return agent
}

// Run 运行研究智能体
func (agent *ResearchAgent) Run(ctx context.Context, query string) (string, error) {
	agent.Logger.Info("Starting research agent", "query", query)
	
	// 1. 分析查询并制定研究计划
	err := agent.createResearchPlan(query)
	if err != nil {
		return "", fmt.Errorf("failed to create research plan: %w", err)
	}
	
	// 2. 执行研究计划
	for _, phase := range agent.ResearchPlan.Phases {
		agent.CurrentPhase = phase
		agent.Logger.Info("Executing research phase", "phase", phase.Name)
		
		result, err := agent.executePhase(ctx, phase)
		if err != nil {
			agent.Logger.Error("Phase execution failed", "phase", phase.Name, "error", err)
			return "", err
		}
		
		// 更新阶段状态
		agent.updatePhaseStatus(phase.ID, "completed", result)
	}
	
	// 3. 生成研究报告
	report, err := agent.generateReport()
	if err != nil {
		return "", fmt.Errorf("failed to generate report: %w", err)
	}
	
	agent.Logger.Info("Research completed successfully")
	return report, nil
}

// createResearchPlan 创建研究计划
func (agent *ResearchAgent) createResearchPlan(query string) error {
	agent.Logger.Info("Creating research plan", "query", query)
	
	// 分析查询类型
	queryType := agent.analyzeQueryType(query)
	
	// 根据查询类型制定研究计划
	switch queryType {
	case "market_research":
		agent.ResearchPlan = agent.createMarketResearchPlan(query)
	case "competitor_analysis":
		agent.ResearchPlan = agent.createCompetitorAnalysisPlan(query)
	case "trend_analysis":
		agent.ResearchPlan = agent.createTrendAnalysisPlan(query)
	case "data_analysis":
		agent.ResearchPlan = agent.createDataAnalysisPlan(query)
	default:
		agent.ResearchPlan = agent.createGeneralResearchPlan(query)
	}
	
	agent.ResearchPlan.Title = query
	agent.ResearchPlan.Description = fmt.Sprintf("Research plan for: %s", query)
	
	agent.Logger.Info("Research plan created", "phases", len(agent.ResearchPlan.Phases))
	return nil
}

// analyzeQueryType 分析查询类型
func (agent *ResearchAgent) analyzeQueryType(query string) string {
	query = strings.ToLower(query)
	
	if strings.Contains(query, "market") || strings.Contains(query, "industry") {
		return "market_research"
	}
	if strings.Contains(query, "competitor") || strings.Contains(query, "competition") {
		return "competitor_analysis"
	}
	if strings.Contains(query, "trend") || strings.Contains(query, "growth") {
		return "trend_analysis"
	}
	if strings.Contains(query, "data") || strings.Contains(query, "analysis") {
		return "data_analysis"
	}
	
	return "general_research"
}

// createMarketResearchPlan 创建市场研究计划
func (agent *ResearchAgent) createMarketResearchPlan(query string) *ResearchPlan {
	return &ResearchPlan{
		Phases: []ResearchPhase{
			{
				ID:          "phase_1",
				Name:        "市场信息收集",
				Description: "收集目标市场的基本信息和数据",
				Status:      "pending",
				Tools:       []string{"web_search", "data_collection"},
			},
			{
				ID:          "phase_2",
				Name:        "市场规模分析",
				Description: "分析市场规模、增长趋势和潜力",
				Status:      "pending",
				Tools:       []string{"data_analysis", "trend_analysis"},
			},
			{
				ID:          "phase_3",
				Name:        "竞争格局分析",
				Description: "分析主要竞争对手和市场格局",
				Status:      "pending",
				Tools:       []string{"competitor_analysis", "swot_analysis"},
			},
			{
				ID:          "phase_4",
				Name:        "报告生成",
				Description: "生成市场研究报告",
				Status:      "pending",
				Tools:       []string{"report_generation"},
			},
		},
	}
}

// createCompetitorAnalysisPlan 创建竞争对手分析计划
func (agent *ResearchAgent) createCompetitorAnalysisPlan(query string) *ResearchPlan {
	return &ResearchPlan{
		Phases: []ResearchPhase{
			{
				ID:          "phase_1",
				Name:        "竞争对手识别",
				Description: "识别主要竞争对手",
				Status:      "pending",
				Tools:       []string{"competitor_identification"},
			},
			{
				ID:          "phase_2",
				Name:        "竞争对手分析",
				Description: "深入分析竞争对手的优劣势",
				Status:      "pending",
				Tools:       []string{"competitor_analysis", "swot_analysis"},
			},
			{
				ID:          "phase_3",
				Name:        "竞争策略分析",
				Description: "分析竞争策略和差异化",
				Status:      "pending",
				Tools:       []string{"strategy_analysis"},
			},
			{
				ID:          "phase_4",
				Name:        "报告生成",
				Description: "生成竞争对手分析报告",
				Status:      "pending",
				Tools:       []string{"report_generation"},
			},
		},
	}
}

// createTrendAnalysisPlan 创建趋势分析计划
func (agent *ResearchAgent) createTrendAnalysisPlan(query string) *ResearchPlan {
	return &ResearchPlan{
		Phases: []ResearchPhase{
			{
				ID:          "phase_1",
				Name:        "趋势数据收集",
				Description: "收集相关趋势数据",
				Status:      "pending",
				Tools:       []string{"trend_data_collection"},
			},
			{
				ID:          "phase_2",
				Name:        "趋势分析",
				Description: "分析趋势模式和变化",
				Status:      "pending",
				Tools:       []string{"trend_analysis", "pattern_recognition"},
			},
			{
				ID:          "phase_3",
				Name:        "预测分析",
				Description: "预测未来趋势发展",
				Status:      "pending",
				Tools:       []string{"forecasting", "prediction_analysis"},
			},
			{
				ID:          "phase_4",
				Name:        "报告生成",
				Description: "生成趋势分析报告",
				Status:      "pending",
				Tools:       []string{"report_generation"},
			},
		},
	}
}

// createDataAnalysisPlan 创建数据分析计划
func (agent *ResearchAgent) createDataAnalysisPlan(query string) *ResearchPlan {
	return &ResearchPlan{
		Phases: []ResearchPhase{
			{
				ID:          "phase_1",
				Name:        "数据收集",
				Description: "收集相关数据",
				Status:      "pending",
				Tools:       []string{"data_collection"},
			},
			{
				ID:          "phase_2",
				Name:        "数据清洗",
				Description: "清洗和预处理数据",
				Status:      "pending",
				Tools:       []string{"data_cleaning", "data_preprocessing"},
			},
			{
				ID:          "phase_3",
				Name:        "数据分析",
				Description: "进行深入的数据分析",
				Status:      "pending",
				Tools:       []string{"data_analysis", "statistical_analysis"},
			},
			{
				ID:          "phase_4",
				Name:        "可视化报告",
				Description: "生成数据可视化报告",
				Status:      "pending",
				Tools:       []string{"data_visualization", "report_generation"},
			},
		},
	}
}

// createGeneralResearchPlan 创建通用研究计划
func (agent *ResearchAgent) createGeneralResearchPlan(query string) *ResearchPlan {
	return &ResearchPlan{
		Phases: []ResearchPhase{
			{
				ID:          "phase_1",
				Name:        "信息收集",
				Description: "收集相关信息",
				Status:      "pending",
				Tools:       []string{"web_search", "information_gathering"},
			},
			{
				ID:          "phase_2",
				Name:        "信息分析",
				Description: "分析收集到的信息",
				Status:      "pending",
				Tools:       []string{"information_analysis"},
			},
			{
				ID:          "phase_3",
				Name:        "报告生成",
				Description: "生成研究报告",
				Status:      "pending",
				Tools:       []string{"report_generation"},
			},
		},
	}
}

// executePhase 执行研究阶段
func (agent *ResearchAgent) executePhase(ctx context.Context, phase ResearchPhase) (string, error) {
	agent.Logger.Info("Executing phase", "phase", phase.Name)
	
	// 更新阶段状态
	agent.updatePhaseStatus(phase.ID, "running", "")
	
	var result string
	var err error
	
	// 根据阶段类型执行不同的逻辑
	switch phase.Name {
	case "市场信息收集", "信息收集":
		result, err = agent.executeInformationGathering(ctx)
	case "市场规模分析", "信息分析":
		result, err = agent.executeDataAnalysis(ctx)
	case "竞争格局分析", "竞争对手分析":
		result, err = agent.executeCompetitorAnalysis(ctx)
	case "报告生成":
		result, err = agent.executeReportGeneration(ctx)
	default:
		result, err = agent.executeGeneralPhase(ctx, phase)
	}
	
	if err != nil {
		agent.updatePhaseStatus(phase.ID, "failed", err.Error())
		return "", err
	}
	
	return result, nil
}

// executeInformationGathering 执行信息收集
func (agent *ResearchAgent) executeInformationGathering(ctx context.Context) (string, error) {
	agent.Logger.Info("Executing information gathering")
	
	// 使用搜索工具收集信息
	searchQuery := agent.Context.Query
	results, err := agent.AvailableTools.Execute("web_search", map[string]interface{}{
		"query": searchQuery,
		"limit": 10,
	})
	
	if err != nil {
		return "", fmt.Errorf("web search failed: %w", err)
	}
	
	// 解析搜索结果
	searchResults := agent.parseSearchResults(results)
	agent.SearchResults = append(agent.SearchResults, searchResults...)
	
	return fmt.Sprintf("收集到 %d 条相关信息", len(searchResults)), nil
}

// executeDataAnalysis 执行数据分析
func (agent *ResearchAgent) executeDataAnalysis(ctx context.Context) (string, error) {
	agent.Logger.Info("Executing data analysis")
	
	// 分析收集到的数据
	analysisResult, err := agent.AvailableTools.Execute("data_analysis", map[string]interface{}{
		"data": agent.SearchResults,
		"type": "comprehensive",
	})
	
	if err != nil {
		return "", fmt.Errorf("data analysis failed: %w", err)
	}
	
	// 保存分析结果
	agent.AnalysisResults = append(agent.AnalysisResults, AnalysisResult{
		ID:        uuid.New(),
		Type:      "comprehensive_analysis",
		Data:      analysisResult.(string),
		Timestamp: time.Now(),
	})
	
	return "数据分析完成", nil
}

// executeCompetitorAnalysis 执行竞争对手分析
func (agent *ResearchAgent) executeCompetitorAnalysis(ctx context.Context) (string, error) {
	agent.Logger.Info("Executing competitor analysis")
	
	// 执行竞争对手分析
	result, err := agent.AvailableTools.Execute("competitor_analysis", map[string]interface{}{
		"query": agent.Context.Query,
		"data":  agent.SearchResults,
	})
	
	if err != nil {
		return "", fmt.Errorf("competitor analysis failed: %w", err)
	}
	
	return result.(string), nil
}

// executeReportGeneration 执行报告生成
func (agent *ResearchAgent) executeReportGeneration(ctx context.Context) (string, error) {
	agent.Logger.Info("Executing report generation")
	
	// 生成报告
	report, err := agent.AvailableTools.Execute("report_generation", map[string]interface{}{
		"title":           agent.ResearchPlan.Title,
		"search_results":  agent.SearchResults,
		"analysis_results": agent.AnalysisResults,
		"format":          "html",
	})
	
	if err != nil {
		return "", fmt.Errorf("report generation failed: %w", err)
	}
	
	return report.(string), nil
}

// executeGeneralPhase 执行通用阶段
func (agent *ResearchAgent) executeGeneralPhase(ctx context.Context, phase ResearchPhase) (string, error) {
	agent.Logger.Info("Executing general phase", "phase", phase.Name)
	
	// 使用LLM生成执行计划
	prompt := fmt.Sprintf("请为研究阶段 '%s' 制定执行计划，描述：%s", phase.Name, phase.Description)
	
	response, err := agent.LLMClient.Chat(ctx, []Message{
		{
			Role:    "user",
			Content: prompt,
		},
	}, &LLMOptions{
		Model:       "gpt-4",
		MaxTokens:   1024,
		Temperature: 0.7,
	})
	
	if err != nil {
		return "", fmt.Errorf("LLM call failed: %w", err)
	}
	
	return response, nil
}

// updatePhaseStatus 更新阶段状态
func (agent *ResearchAgent) updatePhaseStatus(phaseID, status, output string) {
	for i, phase := range agent.ResearchPlan.Phases {
		if phase.ID == phaseID {
			agent.ResearchPlan.Phases[i].Status = status
			agent.ResearchPlan.Phases[i].Output = output
			break
		}
	}
}

// parseSearchResults 解析搜索结果
func (agent *ResearchAgent) parseSearchResults(data interface{}) []SearchResult {
	// 这里实现搜索结果解析逻辑
	// 根据实际的数据格式进行解析
	return []SearchResult{}
}

// generateReport 生成研究报告
func (agent *ResearchAgent) generateReport() (string, error) {
	agent.Logger.Info("Generating final report")
	
	// 汇总所有结果
	agent.ReportData = &ReportData{
		Title:       agent.ResearchPlan.Title,
		Summary:     agent.generateSummary(),
		KeyFindings: agent.extractKeyFindings(),
		DataSources: agent.extractDataSources(),
		Methodology: agent.generateMethodology(),
		Conclusion:  agent.generateConclusion(),
		Recommendations: agent.generateRecommendations(),
	}
	
	// 生成最终报告
	report, err := json.MarshalIndent(agent.ReportData, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal report: %w", err)
	}
	
	return string(report), nil
}

// generateSummary 生成摘要
func (agent *ResearchAgent) generateSummary() string {
	return fmt.Sprintf("本研究针对 '%s' 进行了深入分析，共收集了 %d 条相关信息，完成了 %d 个分析阶段。",
		agent.ResearchPlan.Title,
		len(agent.SearchResults),
		len(agent.ResearchPlan.Phases))
}

// extractKeyFindings 提取关键发现
func (agent *ResearchAgent) extractKeyFindings() []string {
	// 从分析结果中提取关键发现
	return []string{
		"发现1: 市场趋势分析",
		"发现2: 竞争格局分析",
		"发现3: 机会与挑战分析",
	}
}

// extractDataSources 提取数据源
func (agent *ResearchAgent) extractDataSources() []string {
	sources := make([]string, 0)
	for _, result := range agent.SearchResults {
		sources = append(sources, result.URL)
	}
	return sources
}

// generateMethodology 生成方法论
func (agent *ResearchAgent) generateMethodology() string {
	return "本研究采用多阶段研究方法，包括信息收集、数据分析、竞争分析和报告生成等阶段。"
}

// generateConclusion 生成结论
func (agent *ResearchAgent) generateConclusion() string {
	return "基于全面的研究和分析，我们得出了关于该领域的深入见解。"
}

// generateRecommendations 生成建议
func (agent *ResearchAgent) generateRecommendations() []string {
	return []string{
		"建议1: 持续关注市场动态",
		"建议2: 加强竞争分析",
		"建议3: 优化战略定位",
	}
}

// getResearchSystemPrompt 获取研究系统提示
func getResearchSystemPrompt() string {
	return `你是一个专业的研究智能体，专门负责深度研究和分析工作。

你的核心能力包括：
1. 制定详细的研究计划
2. 收集和分析相关信息
3. 进行竞争对手分析
4. 生成专业的研究报告

研究流程：
1. 分析用户需求，制定研究计划
2. 分阶段执行研究任务
3. 收集和分析数据
4. 生成专业报告

请根据用户的研究需求，制定详细的研究计划并逐步执行。`
} 