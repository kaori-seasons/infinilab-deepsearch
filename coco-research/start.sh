#!/bin/bash

# Coco AI Research 项目启动脚本

set -e

echo "🚀 启动 Coco AI Research 项目..."

# 检查Docker是否安装
if ! command -v docker &> /dev/null; then
    echo "❌ Docker未安装，请先安装Docker"
    exit 1
fi

# 检查Docker Compose是否安装
if ! command -v docker-compose &> /dev/null; then
    echo "❌ Docker Compose未安装，请先安装Docker Compose"
    exit 1
fi

# 检查端口是否被占用
check_port() {
    local port=$1
    local service=$2
    
    if lsof -Pi :$port -sTCP:LISTEN -t >/dev/null ; then
        echo "⚠️  端口 $port 已被占用，$service 可能无法启动"
        read -p "是否继续？(y/N): " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            exit 1
        fi
    fi
}

# 检查必要端口
check_port 5432 "PostgreSQL"
check_port 6379 "Redis"
check_port 9200 "Elasticsearch"
check_port 8080 "Research Agent"
check_port 1601 "Tool Service"
check_port 3000 "Frontend"

# 进入项目目录
cd "$(dirname "$0")"

# 创建必要的目录
echo "📁 创建必要的目录..."
mkdir -p uploads
mkdir -p logs

# 设置环境变量
export COMPOSE_PROJECT_NAME=coco-research

# 启动服务
echo "🔧 启动服务..."
docker-compose -f docker/docker-compose.yml up -d

# 等待服务启动
echo "⏳ 等待服务启动..."
sleep 30

# 检查服务状态
echo "🔍 检查服务状态..."
docker-compose -f docker/docker-compose.yml ps

# 显示访问信息
echo ""
echo "🎉 Coco AI Research 项目启动成功！"
echo ""
echo "📊 服务访问地址："
echo "   - 前端界面: http://localhost:3000"
echo "   - Research Agent API: http://localhost:8080"
echo "   - Tool Service API: http://localhost:1601"
echo "   - Elasticsearch: http://localhost:9200"
echo ""
echo "📚 API文档："
echo "   - Research Agent: http://localhost:8080/docs"
echo "   - Tool Service: http://localhost:1601/docs"
echo ""
echo "🔧 管理命令："
echo "   - 查看日志: docker-compose -f docker/docker-compose.yml logs -f"
echo "   - 停止服务: docker-compose -f docker/docker-compose.yml down"
echo "   - 重启服务: docker-compose -f docker/docker-compose.yml restart"
echo ""
echo "💡 提示：首次启动可能需要几分钟时间下载镜像和初始化数据库" 