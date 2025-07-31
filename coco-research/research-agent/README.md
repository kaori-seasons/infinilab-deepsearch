# Coco AI Research Agent - 深度研究智能体系统

## 项目简介

Coco AI Research Agent是一个基于京东开源JoyAgent-JDGenie架构的深度研究智能体系统，旨在为企业提供自动化、智能化的研究解决方案。

### 核心特性

- **多模式智能体**: 支持React模式、Plan-and-Execute模式、研究模式、增强模式四种智能体模式
- **智能体管理器**: 统一的智能体创建、管理和模式切换功能
- **用户兴趣建模**: 基于用户行为动态计算兴趣centroid
- **混合搜索**: 结合向量搜索和全文搜索的混合检索策略
- **BGE相似性计算**: 支持高级语义相似性计算
- **智能重排序**: 多维度重排序算法
- **丰富的研究工具**: 集成网络搜索、数据分析、文档生成等多种工具
- **多种输出格式**: 支持HTML报告、PPT演示、Markdown文档等多种输出
- **实时进度跟踪**: WebSocket实时通信，实时展示研究进度
- **可扩展架构**: 支持自定义工具和智能体扩展
- **向量搜索**: 支持语义搜索和知识图谱构建

### 技术架构

```
前端层: React + TypeScript + WebSocket
网关层: Nginx + API Gateway  
服务层: Go + Python + FastAPI
数据层: PostgreSQL + Redis + Elasticsearch
```

## 智能体模式

### 1. React模式 (ReAct Pattern)
React模式基于"思考-行动-观察"的循环机制：

- **思考(Think)**: 分析当前情况，决定下一步行动
- **行动(Act)**: 执行具体的行动，可能是调用工具或观察
- **观察(Observe)**: 观察行动的结果
- **循环**: 继续思考-行动-观察的循环

**适用场景**: 需要动态决策和灵活响应的研究任务

### 2. Plan-and-Execute模式
Plan-and-Execute模式将任务分解为两个阶段：

- **计划阶段(Plan)**: 分析任务，制定详细的执行计划
- **执行阶段(Execute)**: 按照计划逐步执行任务

**适用场景**: 复杂、多步骤的研究任务，需要系统性的执行计划

### 3. 研究模式 (Research Mode)
传统的研究智能体模式，专注于深度研究和分析：

- **需求分析**: 自动分析用户需求
- **研究计划**: 制定研究计划
- **多步骤执行**: 执行多步骤研究任务
- **报告生成**: 生成研究报告

**适用场景**: 标准的研究和分析任务

### 4. 增强模式 (Enhanced Mode)
增强模式集成了用户兴趣建模和混合搜索功能：

- **用户兴趣建模**: 基于用户历史行为计算兴趣centroid
- **混合搜索**: 结合向量搜索和全文搜索的两阶段检索
- **BGE相似性计算**: 使用BGE模型进行精确语义相似性计算
- **智能重排序**: 多维度重排序算法提升搜索质量
- **个性化推荐**: 基于用户兴趣的个性化内容推荐

**适用场景**: 需要个性化服务和高质量搜索的研究任务

## 快速开始

### 环境要求

- Go 1.21+
- Python 3.11+
- Node.js 18+
- PostgreSQL 15+
- Redis 7+
- Elasticsearch 8.11+

### 安装步骤

1. **克隆项目**
```bash
git clone https://github.com/coco-ai/research-agent.git
cd research-agent
```

2. **配置环境**
```bash
# 复制配置文件
cp config/app.yml.example config/app.yml

# 编辑配置文件
vim config/app.yml
```

3. **启动后端服务**
```bash
# 启动Research Agent服务
cd research-agent
go mod download
go run main.go

# 启动Tool Service
cd tool-service
pip install -r requirements.txt
python server.py
```

4. **启动前端服务**
```bash
cd frontend
npm install
npm start
```

5. **访问系统**
打开浏览器访问 http://localhost:3000

## API接口

### 智能体管理

#### 创建智能体
```http
POST /api/v1/agents
Content-Type: application/json

{
  "name": "研究助手",
  "type": "research",
  "mode": "react",
  "description": "基于React模式的研究智能体"
}
```

#### 列出智能体
```http
GET /api/v1/agents
```

#### 执行智能体
```http
POST /api/v1/agents/{id}/execute
Content-Type: application/json

{
  "query": "分析当前AI市场的发展趋势"
}
```

#### 切换智能体模式
```http
PUT /api/v1/agents/{id}/mode
Content-Type: application/json

{
  "mode": "plan_execute"
}
```

#### 获取智能体状态
```http
GET /api/v1/agents/{id}/state
```

#### 获取智能体统计
```http
GET /api/v1/agents/statistics
```

### 会话管理

#### 创建会话
```http
POST /api/v1/sessions
Content-Type: application/json

{
  "name": "AI市场研究",
  "description": "研究AI市场发展趋势",
  "agent_id": "uuid"
}
```

#### 列出会话
```http
GET /api/v1/sessions
```

### 工具管理

#### 列出工具
```http
GET /api/v1/tools
```

#### 执行工具
```http
POST /api/v1/tools/{name}/execute
Content-Type: application/json

{
  "parameters": {
    "query": "AI市场趋势"
  }
}
```

## 配置说明

### 智能体配置
```yaml
agent:
  max_concurrent: 10
  timeout: 30m
  modes:
    react:
      max_thoughts: 10
      max_actions: 20
    plan_execute:
      max_steps: 50
```

### LLM配置
```yaml
llm:
  provider: "openai"  # openai, claude, deepseek
  api_key: "your-api-key"
  model: "gpt-4"
  max_tokens: 2000
  temperature: 0.7
  embedding:
    model: "iic/nlp_corom_sentence-embedding_chinese-base"
    provider: "huggingface"
    max_length: 512
    dimension: 768
```

### 记忆系统配置
```yaml
memory:
  redis:
    host: "localhost"
    port: 6379
    password: ""
    db: 0
  elasticsearch:
    hosts: ["localhost:9200"]
    username: ""
    password: ""
    index: "memory_items"
  enable_vector_search: true
```

## 开发指南

### 添加新的智能体模式

1. 创建新的智能体类型
```go
type CustomAgent struct {
    *BaseAgent
    // 自定义字段
}

func NewCustomAgent() *CustomAgent {
    baseAgent := NewBaseAgent("CustomAgent", "自定义智能体描述")
    return &CustomAgent{BaseAgent: baseAgent}
}

func (agent *CustomAgent) Run(ctx context.Context, query string) (string, error) {
    // 实现自定义逻辑
}
```

2. 在智能体管理器中注册
```go
case AgentModeCustom:
    customAgent := NewCustomAgent()
    customAgent.Name = req.Name
    customAgent.Description = req.Description
    agent = customAgent
```

### 添加新的工具

1. 实现工具接口
```go
type CustomTool struct {
    name        string
    description string
}

func (t *CustomTool) GetName() string {
    return t.name
}

func (t *CustomTool) GetDescription() string {
    return t.description
}

func (t *CustomTool) Execute(input interface{}) (interface{}, error) {
    // 实现工具逻辑
}
```

2. 注册工具
```go
customTool := NewCustomTool()
toolCollection.RegisterTool(customTool)
```

## 测试

### 运行单元测试
```bash
go test ./...
```

### 运行集成测试
```bash
go test -tags=integration ./...
```

### 测试智能体模式
```bash
go run test_agent_modes.go
```

## 部署

### Docker部署
```bash
# 构建镜像
docker build -t coco-research-agent .

# 运行容器
docker run -p 8080:8080 coco-research-agent
```

### Kubernetes部署
```bash
kubectl apply -f k8s/
```

## 监控和日志

### 健康检查
```http
GET /health
```

### 日志级别
- DEBUG: 详细调试信息
- INFO: 一般信息
- WARN: 警告信息
- ERROR: 错误信息

### 指标监控
- 智能体执行时间
- 工具调用次数
- 记忆存储使用量
- API响应时间

## 贡献指南

1. Fork项目
2. 创建功能分支
3. 提交更改
4. 推送到分支
5. 创建Pull Request

## 许可证

MIT License

## 联系方式

- 项目主页: https://github.com/coco-ai/research-agent
- 问题反馈: https://github.com/coco-ai/research-agent/issues
- 邮箱: support@coco-ai.com 