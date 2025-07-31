from fastapi import FastAPI, HTTPException, BackgroundTasks
from fastapi.middleware.cors import CORSMiddleware
from pydantic import BaseModel
from typing import Dict, Any, List, Optional
import uvicorn
import asyncio
import json
import logging
from datetime import datetime

from tools.web_search import WebSearchTool
from tools.data_analysis import DataAnalysisTool
from tools.report_generation import ReportGenerationTool
from tools.competitor_analysis import CompetitorAnalysisTool
from tools.trend_analysis import TrendAnalysisTool

# 配置日志
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

app = FastAPI(
    title="Coco AI Tool Service",
    description="智能研究工具服务",
    version="1.0.0"
)

# 配置CORS
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# 请求模型
class ToolRequest(BaseModel):
    tool_name: str
    parameters: Dict[str, Any]

class ToolResponse(BaseModel):
    success: bool
    data: Any
    message: str
    execution_time: float

# 工具实例
tools = {
    "web_search": WebSearchTool(),
    "data_analysis": DataAnalysisTool(),
    "report_generation": ReportGenerationTool(),
    "competitor_analysis": CompetitorAnalysisTool(),
    "trend_analysis": TrendAnalysisTool(),
}

@app.get("/")
async def root():
    """根路径"""
    return {
        "message": "Coco AI Tool Service",
        "version": "1.0.0",
        "status": "running"
    }

@app.get("/health")
async def health_check():
    """健康检查"""
    return {
        "status": "healthy",
        "timestamp": datetime.now().isoformat(),
        "available_tools": list(tools.keys())
    }

@app.get("/tools")
async def list_tools():
    """列出所有可用工具"""
    tool_list = []
    for name, tool in tools.items():
        tool_list.append({
            "name": name,
            "description": tool.description,
            "parameters": tool.get_parameters(),
            "status": "available"
        })
    return {"tools": tool_list}

@app.get("/tools/{tool_name}")
async def get_tool_info(tool_name: str):
    """获取工具信息"""
    if tool_name not in tools:
        raise HTTPException(status_code=404, detail="Tool not found")
    
    tool = tools[tool_name]
    return {
        "name": tool_name,
        "description": tool.description,
        "parameters": tool.get_parameters(),
        "status": "available"
    }

@app.post("/tools/{tool_name}/execute")
async def execute_tool(tool_name: str, request: ToolRequest):
    """执行工具"""
    if tool_name not in tools:
        raise HTTPException(status_code=404, detail="Tool not found")
    
    tool = tools[tool_name]
    start_time = datetime.now()
    
    try:
        # 执行工具
        result = await tool.execute(request.parameters)
        execution_time = (datetime.now() - start_time).total_seconds()
        
        return ToolResponse(
            success=True,
            data=result,
            message="Tool executed successfully",
            execution_time=execution_time
        )
    except Exception as e:
        execution_time = (datetime.now() - start_time).total_seconds()
        logger.error(f"Tool execution failed: {str(e)}")
        
        return ToolResponse(
            success=False,
            data=None,
            message=f"Tool execution failed: {str(e)}",
            execution_time=execution_time
        )

@app.post("/tools/batch")
async def execute_batch_tools(requests: List[ToolRequest]):
    """批量执行工具"""
    results = []
    
    for request in requests:
        if request.tool_name not in tools:
            results.append({
                "tool_name": request.tool_name,
                "success": False,
                "error": "Tool not found"
            })
            continue
        
        tool = tools[request.tool_name]
        start_time = datetime.now()
        
        try:
            result = await tool.execute(request.parameters)
            execution_time = (datetime.now() - start_time).total_seconds()
            
            results.append({
                "tool_name": request.tool_name,
                "success": True,
                "data": result,
                "execution_time": execution_time
            })
        except Exception as e:
            execution_time = (datetime.now() - start_time).total_seconds()
            results.append({
                "tool_name": request.tool_name,
                "success": False,
                "error": str(e),
                "execution_time": execution_time
            })
    
    return {"results": results}

if __name__ == "__main__":
    uvicorn.run(
        "main:app",
        host="0.0.0.0",
        port=8001,
        reload=True,
        log_level="info"
    ) 