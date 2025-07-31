#!/bin/bash

# Coco AI Research Agent 测试脚本

echo "🧪 开始测试 Coco AI Research Agent..."

# 检查可执行文件是否存在
if [ ! -f "./main" ]; then
    echo "❌ 可执行文件不存在，请先构建项目"
    exit 1
fi

# 启动服务
echo "🚀 启动服务..."
./main &
SERVER_PID=$!

# 等待服务启动
echo "⏳ 等待服务启动..."
sleep 3

# 测试健康检查
echo "🔍 测试健康检查..."
HEALTH_RESPONSE=$(curl -s http://localhost:8080/api/v1/health)
if [[ $HEALTH_RESPONSE == *"ok"* ]]; then
    echo "✅ 健康检查通过"
else
    echo "❌ 健康检查失败: $HEALTH_RESPONSE"
    kill $SERVER_PID
    exit 1
fi

# 测试API端点
echo "🔍 测试API端点..."
ENDPOINTS=(
    "/api/v1/research/sessions"
    "/api/v1/research/tasks"
    "/api/v1/tools"
)

for endpoint in "${ENDPOINTS[@]}"; do
    echo "测试端点: $endpoint"
    RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:8080$endpoint)
    if [ "$RESPONSE" = "200" ]; then
        echo "✅ $endpoint 响应正常"
    else
        echo "❌ $endpoint 响应异常: $RESPONSE"
    fi
done

# 停止服务
echo "🛑 停止服务..."
kill $SERVER_PID

echo "🎉 测试完成！" 