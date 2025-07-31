# Coco AI Research 项目结构

## 整体项目结构

```
coco-ai-research/
├── README.md                    # 项目说明文档
├── design.md                    # 详细设计文档
├── project-structure.md         # 项目结构说明
├── joyagent-jdgenie/           # 京东开源项目参考
│   ├── genie-backend/          # Java后端服务
│   ├── genie-tool/             # Python工具服务
│   ├── ui/                     # React前端
│   └── ...
└── coco-research/              # Coco AI Research项目
    ├── research-agent/          # Go后端服务
    ├── tool-service/            # Python工具服务
    ├── frontend/                # React前端
    ├── docs/                    # 文档
    ├── k8s/                     # Kubernetes配置
    └── docker/                  # Docker配置
```

## 核心模块结构

### 1. Research Agent (Go后端)

```
research-agent/
├── cmd/
│   └── main.go                 # 应用入口
├── internal/
│   ├── agent/                  # 智能体核心
│   │   ├── base.go            # 基础智能体接口
│   │   ├── research.go        # 研究智能体实现
│   │   └── manager.go         # 智能体管理器
│   ├── tool/                   # 工具系统
│   │   ├── interface.go       # 工具接口
│   │   ├── collection.go      # 工具集合
│   │   └── tools/             # 具体工具实现
│   ├── memory/                 # 记忆系统
│   │   ├── interface.go       # 记忆接口
│   │   ├── es_memory.go       # ES记忆实现
│   │   └── redis_cache.go     # Redis缓存
│   ├── llm/                    # LLM服务
│   │   ├── client.go          # LLM客户端
│   │   ├── router.go          # 模型路由
│   │   └── providers/         # 不同LLM提供商
│   ├── api/                    # API层
│   │   ├── handlers/          # 请求处理器
│   │   ├── middleware/        # 中间件
│   │   └── routes.go          # 路由配置
│   ├── database/               # 数据库层
│   │   ├── models/            # 数据模型
│   │   ├── migrations/        # 数据库迁移
│   │   └── repository/        # 数据访问层
│   └── config/                 # 配置管理
│       ├── app.go             # 应用配置
│       └── database.go        # 数据库配置
├── pkg/                        # 公共包
│   ├── logger/                # 日志工具
│   ├── utils/                 # 工具函数
│   └── constants/             # 常量定义
├── go.mod                      # Go模块文件
├── go.sum                      # 依赖锁定文件
└── Dockerfile                  # Docker镜像配置
```

### 2. Tool Service (Python工具服务)

```
tool-service/
├── genie_tool/
│   ├── __init__.py
│   ├── api/                    # API路由
│   │   ├── __init__.py
│   │   ├── routes.py          # 路由定义
│   │   └── handlers.py        # 请求处理
│   ├── tools/                  # 工具实现
│   │   ├── __init__.py
│   │   ├── web_search.py      # 网络搜索工具
│   │   ├── data_analysis.py   # 数据分析工具
│   │   ├── document_gen.py    # 文档生成工具
│   │   └── code_executor.py   # 代码执行工具
│   ├── db/                     # 数据库
│   │   ├── __init__.py
│   │   ├── models.py          # 数据模型
│   │   └── db_engine.py       # 数据库引擎
│   ├── util/                   # 工具函数
│   │   ├── __init__.py
│   │   ├── middleware_util.py # 中间件工具
│   │   └── logger.py          # 日志工具
│   └── config/                 # 配置
│       ├── __init__.py
│       └── settings.py        # 配置设置
├── server.py                   # 服务入口
├── requirements.txt            # Python依赖
├── pyproject.toml             # 项目配置
└── Dockerfile                 # Docker镜像配置
```

### 3. Frontend (React前端)

```
frontend/
├── public/                     # 静态资源
│   ├── index.html
│   └── favicon.ico
├── src/
│   ├── components/             # 组件
│   │   ├── common/            # 通用组件
│   │   ├── research/          # 研究相关组件
│   │   └── layout/            # 布局组件
│   ├── pages/                  # 页面
│   │   ├── Home.tsx           # 首页
│   │   ├── Research.tsx       # 研究页面
│   │   ├── Sessions.tsx       # 会话管理
│   │   └── Reports.tsx        # 报告查看
│   ├── services/               # API服务
│   │   ├── api.ts             # API客户端
│   │   ├── websocket.ts       # WebSocket服务
│   │   └── types.ts           # 类型定义
│   ├── hooks/                  # 自定义Hooks
│   │   ├── useResearch.ts     # 研究相关Hook
│   │   └── useWebSocket.ts    # WebSocket Hook
│   ├── utils/                  # 工具函数
│   │   ├── constants.ts       # 常量
│   │   └── helpers.ts         # 辅助函数
│   ├── styles/                 # 样式
│   │   ├── global.css         # 全局样式
│   │   └── components.css     # 组件样式
│   ├── App.tsx                # 应用根组件
│   ├── main.tsx               # 应用入口
│   └── vite-env.d.ts          # Vite类型定义
├── package.json                # 项目配置
├── vite.config.ts             # Vite配置
├── tsconfig.json              # TypeScript配置
└── Dockerfile                 # Docker镜像配置
```

### 4. 配置文件结构

```
config/
├── app.yml                     # 应用配置
├── database.yml                # 数据库配置
├── redis.yml                   # Redis配置
├── elasticsearch.yml           # ES配置
├── llm.yml                     # LLM配置
└── tools.yml                   # 工具配置
```

### 5. 部署配置

```
k8s/
├── namespace.yaml              # 命名空间
├── configmap.yaml              # 配置映射
├── secret.yaml                 # 密钥
├── deployment.yaml             # 部署配置
├── service.yaml                # 服务配置
└── ingress.yaml                # 入口配置

docker/
├── docker-compose.yml          # Docker Compose配置
├── research-agent/
│   └── Dockerfile             # Research Agent镜像
├── tool-service/
│   └── Dockerfile             # Tool Service镜像
└── frontend/
    └── Dockerfile             # Frontend镜像
```

## 数据流架构

```
用户请求 → Nginx → API Gateway → Research Agent
                                    ↓
                              Tool Service ← LLM Service
                                    ↓
                              Memory Service → Elasticsearch
                                    ↓
                              Database → PostgreSQL
```

## 开发环境设置

### 1. 本地开发环境

```bash
# 克隆项目
git clone https://github.com/coco-ai/research-agent.git
cd research-agent

# 启动数据库服务
docker-compose up -d postgres redis elasticsearch

# 启动后端服务
cd research-agent
go mod download
go run cmd/main.go

# 启动工具服务
cd tool-service
pip install -r requirements.txt
python server.py

# 启动前端服务
cd frontend
npm install
npm start
```

### 2. 开发工具配置

```bash
# Go开发工具
go install golang.org/x/tools/gopls@latest
go install github.com/go-delve/delve/cmd/dlv@latest

# Python开发工具
pip install black isort mypy pytest

# 前端开发工具
npm install -g @typescript-eslint/eslint-plugin
npm install -g @typescript-eslint/parser
```

## 测试结构

```
tests/
├── unit/                       # 单元测试
│   ├── agent/                 # 智能体测试
│   ├── tool/                  # 工具测试
│   └── memory/                # 记忆测试
├── integration/                # 集成测试
│   ├── api/                   # API测试
│   └── database/              # 数据库测试
├── e2e/                       # 端到端测试
│   ├── research_flow/         # 研究流程测试
│   └── user_scenarios/        # 用户场景测试
└── fixtures/                   # 测试数据
    ├── research_data/         # 研究数据
    └── mock_responses/        # 模拟响应
```

## 监控和日志

```
monitoring/
├── prometheus/                 # Prometheus配置
│   ├── prometheus.yml         # 监控配置
│   └── rules/                 # 告警规则
├── grafana/                    # Grafana配置
│   ├── dashboards/            # 仪表板
│   └── datasources/           # 数据源
└── logs/                       # 日志配置
    ├── logback.xml            # 日志配置
    └── logrotate.conf         # 日志轮转
```

这个项目结构设计遵循了微服务架构的最佳实践，确保了代码的可维护性、可扩展性和可测试性。每个模块都有清晰的职责分工，便于团队协作开发。 