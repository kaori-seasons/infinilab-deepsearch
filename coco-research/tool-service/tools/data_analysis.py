from typing import Dict, Any, List
from .base_tool import BaseTool
import logging
import json

logger = logging.getLogger(__name__)

class DataAnalysisTool(BaseTool):
    """数据分析工具"""
    
    def __init__(self):
        super().__init__()
        self.name = "data_analysis"
        self.description = "执行数据分析，包括统计分析、趋势分析等"
        self.parameters = {
            "data": {
                "type": "array",
                "description": "要分析的数据",
                "required": True
            },
            "analysis_type": {
                "type": "string",
                "description": "分析类型",
                "required": True,
                "enum": ["descriptive", "trend", "correlation", "regression"]
            },
            "columns": {
                "type": "array",
                "description": "要分析的列",
                "required": False
            }
        }
    
    async def execute(self, parameters: Dict[str, Any]) -> Dict[str, Any]:
        """执行数据分析"""
        data = parameters.get("data")
        analysis_type = parameters.get("analysis_type")
        columns = parameters.get("columns", [])
        
        try:
            # 将数据转换为DataFrame
            df = pd.DataFrame(data)
            
            if analysis_type == "descriptive":
                result = await self._descriptive_analysis(df, columns)
            elif analysis_type == "trend":
                result = await self._trend_analysis(df, columns)
            elif analysis_type == "correlation":
                result = await self._correlation_analysis(df, columns)
            elif analysis_type == "regression":
                result = await self._regression_analysis(df, columns)
            else:
                raise ValueError(f"Unsupported analysis type: {analysis_type}")
            
            return {
                "analysis_type": analysis_type,
                "data_shape": df.shape,
                "result": result,
                "timestamp": "2024-01-01T00:00:00Z"
            }
        except Exception as e:
            logger.error(f"Data analysis failed: {str(e)}")
            raise
    
    async def _descriptive_analysis(self, df: pd.DataFrame, columns: List[str]) -> Dict[str, Any]:
        """描述性统计分析"""
        if not columns:
            columns = df.select_dtypes(include=[np.number]).columns.tolist()
        
        result = {
            "summary": df[columns].describe().to_dict(),
            "missing_values": df[columns].isnull().sum().to_dict(),
            "data_types": df[columns].dtypes.to_dict()
        }
        
        return result
    
    async def _trend_analysis(self, df: pd.DataFrame, columns: List[str]) -> Dict[str, Any]:
        """趋势分析"""
        if not columns:
            columns = df.select_dtypes(include=[np.number]).columns.tolist()
        
        trends = {}
        for col in columns:
            if df[col].dtype in ['int64', 'float64']:
                # 计算趋势
                trend = np.polyfit(range(len(df)), df[col], 1)
                trends[col] = {
                    "slope": float(trend[0]),
                    "intercept": float(trend[1]),
                    "trend_direction": "increasing" if trend[0] > 0 else "decreasing"
                }
        
        return {"trends": trends}
    
    async def _correlation_analysis(self, df: pd.DataFrame, columns: List[str]) -> Dict[str, Any]:
        """相关性分析"""
        if not columns:
            columns = df.select_dtypes(include=[np.number]).columns.tolist()
        
        correlation_matrix = df[columns].corr()
        
        return {
            "correlation_matrix": correlation_matrix.to_dict(),
            "high_correlations": self._find_high_correlations(correlation_matrix)
        }
    
    async def _regression_analysis(self, df: pd.DataFrame, columns: List[str]) -> Dict[str, Any]:
        """回归分析"""
        if len(columns) < 2:
            raise ValueError("Regression analysis requires at least 2 columns")
        
        # 简单的线性回归
        X = df[columns[:-1]]
        y = df[columns[-1]]
        
        # 使用numpy进行简单回归
        X_with_intercept = np.column_stack([np.ones(len(X)), X])
        coefficients = np.linalg.lstsq(X_with_intercept, y, rcond=None)[0]
        
        return {
            "coefficients": coefficients.tolist(),
            "independent_variables": columns[:-1],
            "dependent_variable": columns[-1]
        }
    
    def _find_high_correlations(self, corr_matrix: pd.DataFrame, threshold: float = 0.7) -> List[Dict[str, Any]]:
        """找出高相关性"""
        high_correlations = []
        
        for i in range(len(corr_matrix.columns)):
            for j in range(i+1, len(corr_matrix.columns)):
                corr_value = corr_matrix.iloc[i, j]
                if abs(corr_value) >= threshold:
                    high_correlations.append({
                        "variable1": corr_matrix.columns[i],
                        "variable2": corr_matrix.columns[j],
                        "correlation": float(corr_value)
                    })
        
        return high_correlations