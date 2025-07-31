from typing import Dict, Any, List
from .base_tool import BaseTool
import logging
import json
from datetime import datetime

logger = logging.getLogger(__name__)

class ReportGenerationTool(BaseTool):
    """报告生成工具"""
    
    def __init__(self):
        super().__init__()
        self.name = "report_generation"
        self.description = "生成研究报告，支持多种格式"
        self.parameters = {
            "title": {
                "type": "string",
                "description": "报告标题",
                "required": True
            },
            "content": {
                "type": "object",
                "description": "报告内容",
                "required": True
            },
            "format": {
                "type": "string",
                "description": "输出格式",
                "required": False,
                "default": "markdown",
                "enum": ["markdown", "html", "json"]
            },
            "template": {
                "type": "string",
                "description": "报告模板",
                "required": False,
                "default": "standard"
            }
        }
    
    async def execute(self, parameters: Dict[str, Any]) -> Dict[str, Any]:
        """生成报告"""
        title = parameters.get("title")
        content = parameters.get("content")
        format_type = parameters.get("format", "markdown")
        template = parameters.get("template", "standard")
        
        try:
            if format_type == "markdown":
                report = await self._generate_markdown_report(title, content, template)
            elif format_type == "html":
                report = await self._generate_html_report(title, content, template)
            elif format_type == "json":
                report = await self._generate_json_report(title, content, template)
            else:
                raise ValueError(f"Unsupported format: {format_type}")
            
            return {
                "title": title,
                "format": format_type,
                "template": template,
                "content": report,
                "generated_at": datetime.now().isoformat()
            }
        except Exception as e:
            logger.error(f"Report generation failed: {str(e)}")
            raise
    
    async def _generate_markdown_report(self, title: str, content: Dict[str, Any], template: str) -> str:
        """生成Markdown报告"""
        markdown = f"# {title}\n\n"
        markdown += f"**生成时间**: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}\n\n"
        
        # 添加摘要
        if "summary" in content:
            markdown += "## 摘要\n\n"
            markdown += f"{content['summary']}\n\n"
        
        # 添加主要发现
        if "findings" in content:
            markdown += "## 主要发现\n\n"
            for i, finding in enumerate(content["findings"], 1):
                markdown += f"{i}. {finding}\n"
            markdown += "\n"
        
        # 添加数据
        if "data" in content:
            markdown += "## 数据\n\n"
            markdown += f"```json\n{json.dumps(content['data'], indent=2, ensure_ascii=False)}\n```\n\n"
        
        # 添加结论
        if "conclusion" in content:
            markdown += "## 结论\n\n"
            markdown += f"{content['conclusion']}\n\n"
        
        # 添加建议
        if "recommendations" in content:
            markdown += "## 建议\n\n"
            for i, rec in enumerate(content["recommendations"], 1):
                markdown += f"{i}. {rec}\n"
        
        return markdown
    
    async def _generate_html_report(self, title: str, content: Dict[str, Any], template: str) -> str:
        """生成HTML报告"""
        html = f"""
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{title}</title>
    <style>
        body {{ font-family: Arial, sans-serif; margin: 40px; }}
        h1 {{ color: #333; border-bottom: 2px solid #eee; }}
        h2 {{ color: #666; margin-top: 30px; }}
        .summary {{ background: #f9f9f9; padding: 15px; border-radius: 5px; }}
        .finding {{ margin: 10px 0; }}
        .conclusion {{ background: #e8f5e8; padding: 15px; border-radius: 5px; }}
        .recommendation {{ margin: 10px 0; }}
    </style>
</head>
<body>
    <h1>{title}</h1>
    <p><strong>生成时间</strong>: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}</p>
"""
        
        # 添加摘要
        if "summary" in content:
            html += f'<h2>摘要</h2><div class="summary">{content["summary"]}</div>'
        
        # 添加主要发现
        if "findings" in content:
            html += "<h2>主要发现</h2><ul>"
            for finding in content["findings"]:
                html += f'<li class="finding">{finding}</li>'
            html += "</ul>"
        
        # 添加数据
        if "data" in content:
            html += f'<h2>数据</h2><pre>{json.dumps(content["data"], indent=2, ensure_ascii=False)}</pre>'
        
        # 添加结论
        if "conclusion" in content:
            html += f'<h2>结论</h2><div class="conclusion">{content["conclusion"]}</div>'
        
        # 添加建议
        if "recommendations" in content:
            html += "<h2>建议</h2><ul>"
            for rec in content["recommendations"]:
                html += f'<li class="recommendation">{rec}</li>'
            html += "</ul>"
        
        html += "</body></html>"
        return html
    
    async def _generate_json_report(self, title: str, content: Dict[str, Any], template: str) -> Dict[str, Any]:
        """生成JSON报告"""
        return {
            "title": title,
            "generated_at": datetime.now().isoformat(),
            "template": template,
            "content": content
        } 