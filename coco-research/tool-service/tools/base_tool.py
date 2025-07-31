 from abc import ABC, abstractmethod
from typing import Dict, Any, Optional
import asyncio
import logging

logger = logging.getLogger(__name__)

class BaseTool(ABC):
    """基础工具类"""
    
    def __init__(self):
        self.name = self.__class__.__name__
        self.description = "Base tool description"
        self.parameters = {}
    
    @abstractmethod
    async def execute(self, parameters: Dict[str, Any]) -> Any:
        """执行工具"""
        pass
    
    def get_parameters(self) -> Dict[str, Any]:
        """获取工具参数"""
        return self.parameters
    
    def validate_parameters(self, parameters: Dict[str, Any]) -> bool:
        """验证参数"""
        required_params = [k for k, v in self.parameters.items() if v.get('required', False)]
        
        for param in required_params:
            if param not in parameters:
                raise ValueError(f"Missing required parameter: {param}")
        
        return True
    
    async def execute_with_validation(self, parameters: Dict[str, Any]) -> Any:
        """带验证的执行"""
        try:
            self.validate_parameters(parameters)
            return await self.execute(parameters)
        except Exception as e:
            logger.error(f"Tool execution failed: {str(e)}")
            raise