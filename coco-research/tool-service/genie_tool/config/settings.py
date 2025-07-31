#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
配置管理模块
"""

import os
from typing import List, Optional
from pydantic import BaseSettings, Field


class Settings(BaseSettings):
    """应用配置"""
    
    # 服务器配置
    host: str = Field(default="0.0.0.0", env="HOST")
    port: int = Field(default=1601, env="PORT")
    debug: bool = Field(default=False, env="DEBUG")
    
    # 日志配置
    log_level: str = Field(default="INFO", env="LOG_LEVEL")
    log_output: str = Field(default="stdout", env="LOG_OUTPUT")
    
    # 数据库配置
    database_url: str = Field(
        default="postgresql://coco:password@localhost:5432/coco_research",
        env="DATABASE_URL"
    )
    
    # Redis配置
    redis_host: str = Field(default="localhost", env="REDIS_HOST")
    redis_port: int = Field(default=6379, env="REDIS_PORT")
    redis_password: Optional[str] = Field(default=None, env="REDIS_PASSWORD")
    redis_db: int = Field(default=0, env="REDIS_DB")
    
    # Elasticsearch配置
    es_hosts: List[str] = Field(
        default=["http://localhost:9200"],
        env="ES_HOSTS"
    )
    es_username: Optional[str] = Field(default=None, env="ES_USERNAME")
    es_password: Optional[str] = Field(default=None, env="ES_PASSWORD")
    
    # 工具配置
    serper_api_key: Optional[str] = Field(default=None, env="SERPER_API_KEY")
    openai_api_key: Optional[str] = Field(default=None, env="OPENAI_API_KEY")
    openai_base_url: str = Field(
        default="https://api.openai.com/v1",
        env="OPENAI_BASE_URL"
    )
    
    # 文件存储配置
    upload_dir: str = Field(default="./uploads", env="UPLOAD_DIR")
    max_file_size: int = Field(default=10 * 1024 * 1024, env="MAX_FILE_SIZE")  # 10MB
    
    class Config:
        env_file = ".env"
        case_sensitive = False 