#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
健康检查处理器
"""

from fastapi import APIRouter

router = APIRouter()


@router.get("/health")
async def health_check():
    """健康检查"""
    return {
        "status": "ok",
        "message": "Tool Service is running",
        "version": "1.0.0"
    } 