#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
日志工具模块
"""

import sys
from pathlib import Path
from loguru import logger


def setup_logger(level: str = "INFO", output: str = "stdout"):
    """设置日志配置"""
    
    # 移除默认的日志处理器
    logger.remove()
    
    # 设置日志级别
    log_level = level.upper()
    
    # 设置输出格式
    log_format = (
        "<green>{time:YYYY-MM-DD HH:mm:ss.SSS}</green> | "
        "<level>{level: <8}</level> | "
        "<cyan>{name}</cyan>:<cyan>{function}</cyan>:<cyan>{line}</cyan> | "
        "<level>{message}</level>"
    )
    
    # 添加控制台输出
    if output == "stdout":
        logger.add(
            sys.stdout,
            format=log_format,
            level=log_level,
            colorize=True
        )
    else:
        # 添加文件输出
        log_file = Path(output)
        log_file.parent.mkdir(parents=True, exist_ok=True)
        
        logger.add(
            str(log_file),
            format=log_format,
            level=log_level,
            rotation="100 MB",
            retention="7 days",
            compression="zip"
        )
    
    return logger


def get_logger(name: str = None):
    """获取日志记录器"""
    if name:
        return logger.bind(name=name)
    return logger 