# Coco AI Research - 深度研究智能体系统

## 项目简介

Coco AI Research是一个基于京东开源JoyAgent-JDGenie架构的深度研究智能体系统，旨在为企业提供自动化、智能化的研究解决方案。

### 核心特性

- **多智能体协作**: 支持多种智能体模式（React模式、Plan-and-Execute模式）
- **丰富的研究工具**: 集成网络搜索、数据分析、文档生成等多种工具
- **多种输出格式**: 支持HTML报告、PPT演示、Markdown文档等多种输出
- **实时进度跟踪**: WebSocket实时通信，实时展示研究进度
- **可扩展架构**: 支持自定义工具和智能体扩展

### 技术架构

```
前端层: React + TypeScript + WebSocket
网关层: Nginx + API Gateway  
服务层: Go + Python + FastAPI
数据层: PostgreSQL + Redis + Elasticsearch
```

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

2. **启动后端服务**
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

3. **启动前端服务**
```bash
cd frontend
npm install
npm start
```

4. **访问系统**
打开浏览器访问 http://localhost:3000

## 核心功能

### 1. 研究会话管理

- 创建新的研究项目
- 管理研究任务和进度
- 查看历史研究记录

### 2. 智能研究执行

- 自动分析用户需求
- 制定研究计划
- 执行多步骤研究任务
- 生成研究报告

### 3. 工具生态系统

- **网络搜索工具**: 智能搜索相关信息
- **数据分析工具**: 处理和分析数据
- **文档生成工具**: 生成各种格式的报告
- **代码执行工具**: 执行Python代码分析

### 4. 记忆系统

- 语义搜索历史对话
- 上下文感知的智能体
- 知识图谱构建

## API文档

### 研究会话API

```bash
# 创建研究会话
POST /api/v1/research/sessions
{
    "title": "市场分析研究",
    "description": "分析某行业市场趋势",
    "user_id": "user123"
}

# 获取研究会话列表
GET /api/v1/research/sessions?user_id=user123&page=1&size=10

# 获取研究会话详情
GET /api/v1/research/sessions/{session_id}
```

### 研究任务API

```bash
# 创建研究任务
POST /api/v1/research/sessions/{session_id}/tasks
{
    "task_type": "market_analysis",
    "title": "竞争对手分析",
    "description": "分析主要竞争对手的优势劣势"
}

# 执行研究任务
POST /api/v1/research/tasks/{task_id}/execute

# 获取任务状态
GET /api/v1/research/tasks/{task_id}/status
```

### WebSocket API

```bash
# 连接WebSocket
WS /api/v1/ws/research/{session_id}

# 消息格式
{
    "type": "task_update",
    "data": {
        "task_id": "task123",
        "status": "in_progress",
        "progress": 50,
        "message": "正在分析数据..."
    }
}
```

## 配置说明

### 环境变量配置

```bash
# 数据库配置
DB_HOST=localhost
DB_PORT=5432
DB_NAME=coco_research
DB_USER=coco
DB_PASSWORD=password

# Redis配置
REDIS_HOST=localhost
REDIS_PORT=6379

# Elasticsearch配置
ES_HOST=localhost
ES_PORT=9200

# LLM配置
OPENAI_API_KEY=your_openai_key
OPENAI_BASE_URL=https://api.openai.com/v1
DEFAULT_MODEL=gpt-4
```

### 工具配置

在 `config/tools.yml` 中配置可用的工具：

```yaml
tools:
  web_search:
    name: "web_search"
    description: "网络搜索工具"
    enabled: true
    config:
      api_key: "your_search_api_key"
      
  data_analysis:
    name: "data_analysis"
    description: "数据分析工具"
    enabled: true
    config:
      python_path: "/usr/bin/python3"
```

## 扩展开发

### 添加自定义工具

1. 实现Tool接口：

```go
type CustomTool struct {
    name        string
    description string
}

func (ct *CustomTool) GetName() string {
    return ct.name
}

func (ct *CustomTool) GetDescription() string {
    return ct.description
}

func (ct *CustomTool) Execute(input interface{}) (interface{}, error) {
    // 实现工具逻辑
    return result, nil
}
```

2. 注册工具：

```go
toolCollection.AddTool(&CustomTool{
    name:        "custom_tool",
    description: "自定义工具描述",
})
```

### 添加自定义智能体

1. 实现Agent接口：

```go
type CustomAgent struct {
    BaseAgent
}

func (ca *CustomAgent) Step() (string, error) {
    // 实现智能体逻辑
    return result, nil
}
```

2. 注册智能体：

```go
agentManager.RegisterAgent("custom_agent", &CustomAgent{})
```

## 部署指南

### Docker部署

```bash
# 构建镜像
docker-compose build

# 启动服务
docker-compose up -d

# 查看日志
docker-compose logs -f
```

### Kubernetes部署

```bash
# 应用配置
kubectl apply -f k8s/namespace.yaml
kubectl apply -f k8s/configmap.yaml
kubectl apply -f k8s/deployment.yaml
kubectl apply -f k8s/service.yaml

# 查看状态
kubectl get pods -n coco-research
```

## 监控和日志

### 监控指标

- API响应时间
- 并发用户数
- 系统资源使用率
- 错误率统计

### 日志配置

```yaml
logging:
  level: INFO
  format: json
  output: 
    - stdout
    - file
  file:
    path: /var/log/coco-research
    max_size: 100MB
    max_age: 7d
```

## 故障排除

### 常见问题

1. **服务启动失败**
   - 检查数据库连接
   - 验证环境变量配置
   - 查看日志文件

2. **工具执行失败**
   - 检查工具配置
   - 验证API密钥
   - 查看工具日志

3. **前端无法连接**
   - 检查后端服务状态
   - 验证端口配置
   - 检查网络连接

### 性能优化

1. **数据库优化**
   - 添加索引
   - 优化查询语句
   - 配置连接池

2. **缓存优化**
   - 使用Redis缓存
   - 配置缓存策略
   - 监控缓存命中率

3. **并发优化**
   - 调整goroutine数量
   - 优化资源使用
   - 监控系统负载

## 贡献指南

### 开发环境设置

1. Fork项目
2. 创建功能分支
3. 提交代码
4. 创建Pull Request

### 代码规范

- 遵循Go和Python代码规范
- 添加单元测试
- 更新文档
- 通过CI/CD检查

## 许可证

本项目采用MIT许可证，详见LICENSE文件。

## 联系我们

- 项目主页: https://github.com/coco-ai/research-agent
- 问题反馈: https://github.com/coco-ai/research-agent/issues
- 邮箱: research@coco-ai.com

## 更新日志

### v1.0.0 (2025-01-22)
- 初始版本发布
- 基础智能体功能
- 核心工具集成
- Web界面实现

