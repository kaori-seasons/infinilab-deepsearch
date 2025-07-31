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

- Docker 20.10+
- Docker Compose 2.0+
- 8GB+ 内存
- 10GB+ 磁盘空间

### 一键启动

```bash
# 克隆项目
git clone <repository-url>
cd coco-research

# 启动所有服务
./start.sh
```

### 手动启动

```bash
# 启动数据库和缓存服务
docker-compose -f docker/docker-compose.yml up -d postgres redis elasticsearch

# 启动后端服务
docker-compose -f docker/docker-compose.yml up -d research-agent tool-service

# 启动前端服务
docker-compose -f docker/docker-compose.yml up -d frontend
```

### 访问地址

- **前端界面**: http://localhost:3000
- **Research Agent API**: http://localhost:8080
- **Tool Service API**: http://localhost:1601
- **Elasticsearch**: http://localhost:9200

## 核心功能

### 研究会话管理

- 创建和管理研究会话
- 支持多用户并发研究
- 会话历史记录和恢复

### 智能执行引擎

- 自动任务分解和规划
- 多工具并行执行
- 实时进度监控

### 工具生态系统

- **网络搜索**: 实时网络信息检索
- **数据分析**: 数据清洗、分析和可视化
- **文档生成**: 自动生成研究报告
- **代码执行**: 安全的代码运行环境

### 记忆系统

- 短期记忆（Redis缓存）
- 长期记忆（Elasticsearch）
- 跨会话知识共享

## API文档

### Research Agent API

#### 研究会话管理

```http
POST /api/v1/research/sessions
GET /api/v1/research/sessions
GET /api/v1/research/sessions/{id}
PUT /api/v1/research/sessions/{id}
DELETE /api/v1/research/sessions/{id}
```

#### 研究任务管理

```http
POST /api/v1/research/tasks
GET /api/v1/research/tasks/{id}
PUT /api/v1/research/tasks/{id}
POST /api/v1/research/tasks/{id}/execute
GET /api/v1/research/tasks/{id}/status
```

### Tool Service API

#### 工具执行

```http
GET /api/v1/tools
POST /api/v1/tools/{name}/execute
```

#### 搜索服务

```http
POST /api/v1/search/web
POST /api/v1/search/semantic
```

## 配置说明

### 环境变量

```bash
# 数据库配置
DATABASE_URL=postgresql://coco:password@localhost:5432/coco_research

# Redis配置
REDIS_HOST=localhost
REDIS_PORT=6379

# Elasticsearch配置
ES_HOSTS=http://localhost:9200

# LLM配置
OPENAI_API_KEY=your-openai-api-key
CLAUDE_API_KEY=your-claude-api-key
DEEPSEEK_API_KEY=your-deepseek-api-key

# 搜索配置
SERPER_API_KEY=your-serper-api-key
```

### 配置文件

主要配置文件位于 `research-agent/config/app.yml` 和 `tool-service/.env`

## 开发指南

### 本地开发环境

```bash
# 启动依赖服务
docker-compose -f docker/docker-compose.yml up -d postgres redis elasticsearch

# 启动Go服务
cd research-agent
go run cmd/main.go

# 启动Python服务
cd tool-service
python server.py

# 启动前端服务
cd frontend
npm start
```

### 代码结构

```
coco-research/
├── research-agent/          # Go后端服务
│   ├── cmd/                # 应用入口
│   ├── internal/           # 内部包
│   │   ├── agent/         # 智能体核心
│   │   ├── tool/          # 工具系统
│   │   ├── memory/        # 记忆系统
│   │   ├── llm/           # LLM集成
│   │   ├── api/           # API接口
│   │   └── database/      # 数据库
│   └── config/            # 配置文件
├── tool-service/           # Python工具服务
│   ├── genie_tool/        # 工具包
│   │   ├── api/           # API接口
│   │   ├── tools/         # 工具实现
│   │   └── config/        # 配置管理
│   └── server.py          # 服务入口
├── frontend/              # React前端
├── docker/                # Docker配置
└── docs/                  # 项目文档
```

## 部署指南

### Docker部署

```bash
# 构建镜像
docker-compose -f docker/docker-compose.yml build

# 启动服务
docker-compose -f docker/docker-compose.yml up -d
```

### Kubernetes部署

```bash
# 应用Kubernetes配置
kubectl apply -f k8s/

# 查看服务状态
kubectl get pods -n coco-research
```

## 监控和日志

### 日志查看

```bash
# 查看所有服务日志
docker-compose -f docker/docker-compose.yml logs -f

# 查看特定服务日志
docker-compose -f docker/docker-compose.yml logs -f research-agent
```

### 性能监控

- **Prometheus**: 指标收集
- **Grafana**: 可视化监控
- **Jaeger**: 分布式追踪

## 故障排除

### 常见问题

1. **数据库连接失败**
   - 检查PostgreSQL服务状态
   - 验证数据库连接配置

2. **Redis连接失败**
   - 检查Redis服务状态
   - 验证Redis连接配置

3. **Elasticsearch启动失败**
   - 检查内存使用情况
   - 验证ES配置参数

### 日志分析

```bash
# 查看错误日志
docker-compose -f docker/docker-compose.yml logs --tail=100 | grep ERROR

# 查看特定时间段的日志
docker-compose -f docker/docker-compose.yml logs --since="2024-01-01T00:00:00"
```

## 贡献指南

### 开发流程

1. Fork项目仓库
2. 创建功能分支
3. 提交代码变更
4. 创建Pull Request

### 代码规范

- Go代码遵循Go官方规范
- Python代码遵循PEP 8规范
- 前端代码遵循ESLint配置

### 测试要求

- 单元测试覆盖率 > 80%
- 集成测试覆盖核心功能
- 性能测试满足要求

## 许可证

本项目基于MIT许可证开源。

## 致谢

本项目基于京东开源的JoyAgent-JDGenie项目架构，感谢京东CHO企业信息化团队的贡献。

---

**注意**: 本项目仍在开发中，部分功能可能不稳定。建议在生产环境使用前进行充分测试。 