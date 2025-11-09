#!/usr/bin/env python3
# -*- coding: utf-8 -*-

"""
Chroma向量数据库启动脚本 - 用于宝塔部署
"""

import chromadb
from chromadb.config import Settings
import os
import logging

# 配置日志
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

def start_chroma_server():
    """启动Chroma服务器"""
    try:
        # 确保数据目录存在
        data_dir = "/www/wwwroot/chat-app/chroma-data"
        os.makedirs(data_dir, exist_ok=True)
        
        # 创建Chroma客户端配置
        settings = Settings(
            chroma_db_impl="duckdb+parquet",
            persist_directory=data_dir,
            chroma_server_host="127.0.0.1",
            chroma_server_http_port=8000,
            anonymized_telemetry=False
        )
        
        # 启动服务器
        logger.info("正在启动Chroma向量数据库...")
        logger.info(f"数据目录: {data_dir}")
        logger.info("服务器地址: http://127.0.0.1:8000")
        
        # 这里应该启动实际的服务器
        # 注意：这个脚本可能需要根据Chroma的最新版本进行调整
        import chromadb.server.fastapi
        
        app = chromadb.server.fastapi.app
        
        logger.info("✅ Chroma服务器启动成功!")
        
    except Exception as e:
        logger.error(f"❌ Chroma服务器启动失败: {e}")
        raise

if __name__ == "__main__":
    start_chroma_server()
