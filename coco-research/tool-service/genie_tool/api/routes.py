#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
API路由模块
"""

from fastapi import APIRouter

from .handlers import health, tools, search, document

# 创建主路由器
router = APIRouter()

# 注册子路由器
router.include_router(health.router, tags=["health"])
router.include_router(tools.router, prefix="/tools", tags=["tools"])
router.include_router(search.router, prefix="/search", tags=["search"])
router.include_router(document.router, prefix="/document", tags=["document"]) 