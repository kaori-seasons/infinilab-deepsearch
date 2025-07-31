# Coco AI Research Agent

Coco AI Research Agent 是一个基于Go语言开发的智能研究代理系统，支持深度研究、多模态分析和智能报告生成。

## 功能特性

### 核心功能
- **智能研究代理**: 基于LLM的研究规划和执行
- **多级记忆系统**: 工作记忆、短期记忆（16个记忆槽）、长期记忆（Elasticsearch）
- **向量搜索**: 支持语义相似性搜索，使用中文嵌入模型
- **工具集成**: 网络搜索、数据分析、报告生成等
- **多LLM支持**: OpenAI、Claude、DeepSeek等

### 记忆系统
- **工作记忆**: 最近上下文，保持最近20条消息
- **短期记忆**: 16个记忆槽，基于优先级和访问频率管理
- **长期记忆**: Elasticsearch存储，支持向量搜索

### 嵌入模型
- **默认模型**: `iic/nlp_corom_sentence-embedding_chinese-base`
- **支持提供商**: HuggingFace、OpenAI
- **向量维度**: 768维
- **最大文本长度**: 512字符

## 系统架构

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   React UI      │    │   API Gateway   │    │   Go Backend    │
│   (Frontend)    │◄──►│   (Nginx)       │◄──►│   (Research     │
│                 │    │                 │    │    Agent)       │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                                                       │
                       ┌─────────────────┐            │
                       │   Python Tool   │            │
                       │   Service       │◄───────────┘
                       │   (FastAPI)     │
                       └─────────────────┘
                                
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   PostgreSQL    │    │   Redis Cache   │    │   Elasticsearch │
│   (Database)    │    │   (Memory)      │    │   (Vector DB)   │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## 快速开始

### 1. 环境要求
- Go 1.21+
- PostgreSQL 13+
- Redis 6+
- Elasticsearch 8+
- Docker & Docker Compose

### 2. 配置设置

复制配置文件模板：
```bash
cp config/app.yml.example config/app.yml
```

编辑配置文件，设置必要的API密钥：
```yaml
# LLM配置
llm:
  provider: openai
  openai:
    api_key: "your-openai-api-key"
    model: "gpt-4"
  
  # 嵌入模型配置
  embedding:
    model: "iic/nlp_corom_sentence-embedding_chinese-base"
    provider: "huggingface"
    api_key: "your-huggingface-api-key"
    base_url: "https://api-inference.huggingface.co"
    max_length: 512
    dimension: 768
```

### 3. 启动服务

使用Docker Compose启动所有服务：
```bash
docker-compose up -d
```

或者单独启动Go服务：
```bash
go run cmd/main.go
```

### 4. API使用

#### 创建研究任务
```bash
curl -X POST http://localhost:8080/api/v1/agents \
  -H "Content-Type: application/json" \
  -d '{
    "name": "market_research",
    "description": "研究AI市场趋势",
    "type": "research"
  }'
```

#### 执行任务
```bash
curl -X POST http://localhost:8080/api/v1/agents/{agent_id}/execute \
  -H "Content-Type: application/json" \
  -d '{
    "query": "分析2024年AI市场发展趋势",
    "parameters": {
      "depth": "deep",
      "sources": ["web", "reports"]
    }
  }'
```

## 嵌入模型配置

### 支持的模型

#### 1. HuggingFace模型
```yaml
embedding:
  provider: "huggingface"
  model: "iic/nlp_corom_sentence-embedding_chinese-base"
  api_key: "your-huggingface-api-key"
  base_url: "https://api-inference.huggingface.co"
  max_length: 512
  dimension: 768
```

#### 2. OpenAI模型
```yaml
embedding:
  provider: "openai"
  model: "text-embedding-ada-002"
  api_key: "your-openai-api-key"
  base_url: "https://api.openai.com/v1"
  dimension: 1536
```

### 向量搜索功能

系统会自动为存储的内容生成嵌入向量，支持：

1. **语义搜索**: 基于内容相似性检索相关记忆
2. **混合搜索**: 结合关键词和向量相似性
3. **会话隔离**: 每个会话的向量独立存储和检索

## 记忆系统详解

### 工作记忆（Working Memory）
- **容量**: 最近20条消息
- **用途**: 保持对话上下文
- **存储**: 内存中，会话结束时清除

### 短期记忆（Short-Term Memory）
- **容量**: 16个记忆槽
- **管理策略**: 基于优先级和访问频率
- **替换策略**: 当槽位满时，替换优先级最低且访问次数最少的槽位

### 长期记忆（Long-Term Memory）
- **存储**: Elasticsearch
- **搜索**: 支持关键词和向量混合搜索
- **持久化**: 跨会话保持

## API文档

### 智能体管理
- `GET /api/v1/agents` - 列出所有智能体
- `POST /api/v1/agents` - 创建新智能体
- `GET /api/v1/agents/:id` - 获取智能体信息
- `POST /api/v1/agents/:id/execute` - 执行任务

### 工具管理
- `GET /api/v1/tools` - 列出所有工具
- `POST /api/v1/tools` - 添加新工具
- `GET /api/v1/tools/:name` - 获取工具信息
- `POST /api/v1/tools/:name/execute` - 执行工具

### 会话管理
- `GET /api/v1/sessions` - 列出所有会话
- `POST /api/v1/sessions` - 创建新会话
- `GET /api/v1/sessions/:id` - 获取会话信息

## 开发指南

### 项目结构
```
research-agent/
├── cmd/
│   └── main.go              # 主程序入口
├── internal/
│   ├── agent/               # 智能体实现
│   ├── api/                 # API层
│   ├── config/              # 配置管理
│   ├── database/            # 数据库操作
│   ├── llm/                 # LLM客户端
│   ├── memory/              # 记忆系统
│   └── tool/                # 工具系统
├── pkg/                     # 公共包
├── config/                  # 配置文件
└── docker-compose.yml       # Docker配置
```

### 添加新工具
1. 实现`tool.Tool`接口
2. 在`registerDefaultTools`中注册
3. 配置工具参数

### 添加新LLM提供商
1. 实现`llm.Client`接口
2. 在main.go中添加初始化逻辑
3. 更新配置文件结构

## 部署

### Docker部署
```bash
# 构建镜像
docker build -t coco-research-agent .

# 运行容器
docker run -d \
  --name research-agent \
  -p 8080:8080 \
  -v $(pwd)/config:/app/config \
  coco-research-agent
```

### Kubernetes部署
```bash
kubectl apply -f k8s/
```

## 监控和日志

### 日志级别
- `debug`: 详细调试信息
- `info`: 一般信息
- `warn`: 警告信息
- `error`: 错误信息

### 健康检查
```bash
curl http://localhost:8080/health
```

### 指标监控
- 智能体执行时间
- 记忆系统使用情况
- API调用统计

## 故障排除

### 常见问题

1. **嵌入模型连接失败**
   - 检查API密钥是否正确
   - 确认网络连接正常
   - 查看日志中的错误信息

2. **Elasticsearch连接问题**
   - 确认Elasticsearch服务运行正常
   - 检查索引是否创建
   - 验证认证信息

3. **记忆系统性能问题**
   - 调整工作记忆大小
   - 优化短期记忆槽数量
   - 检查向量搜索配置

## 贡献指南

1. Fork项目
2. 创建功能分支
3. 提交更改
4. 创建Pull Request

## 许可证

MIT License

## 联系方式

- 项目主页: https://github.com/coco-ai/research-agent
- 问题反馈: https://github.com/coco-ai/research-agent/issues 