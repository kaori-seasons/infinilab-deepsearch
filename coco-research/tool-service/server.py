#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
Coco AI Research Tool Service
工具服务入口文件
"""

import os
import sys
from pathlib import Path

# 添加项目根目录到Python路径
project_root = Path(__file__).parent
sys.path.insert(0, str(project_root))

from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware
from loguru import logger

from genie_tool.api.routes import router
from genie_tool.config.settings import Settings
from genie_tool.util.logger import setup_logger

def create_app() -> FastAPI:
    """创建FastAPI应用"""
    
    # 加载配置
    settings = Settings()
    
    # 设置日志
    setup_logger(settings.log_level, settings.log_output)
    
    # 创建应用
    app = FastAPI(
        title="Coco AI Research Tool Service",
        description="深度研究智能体工具服务",
        version="1.0.0",
        docs_url="/docs",
        redoc_url="/redoc"
    )
    
    # 添加CORS中间件
    app.add_middleware(
        CORSMiddleware,
        allow_origins=["*"],
        allow_credentials=True,
        allow_methods=["*"],
        allow_headers=["*"],
    )
    
    # 注册路由
    app.include_router(router, prefix="/api/v1")
    
    # 健康检查
    @app.get("/health")
    async def health_check():
        return {
            "status": "ok",
            "message": "Tool Service is running",
            "version": "1.0.0"
        }
    
    return app

def main():
    """主函数"""
    import uvicorn
    
    # 获取配置
    settings = Settings()
    
    # 创建应用
    app = create_app()
    
    # 启动服务器
    uvicorn.run(
        app,
        host=settings.host,
        port=settings.port,
        reload=settings.debug,
        log_level=settings.log_level.lower()
    )

if __name__ == "__main__":
    main() 