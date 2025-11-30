# Universal MCP Server Architecture

## Overview
Comprehensive MCP server architecture supporting multiple connectors including Figma, Notion, Nano-Banana, and custom integrations with modular tool registration and unified API management.

## Quick Implementation

### Base MCP Server with Multiple Connectors

```python
from fastmcp import FastMCP
from typing import Dict, List, Optional
import asyncio

class UniversalMCPServer:
 def __init__(self, server_name: str):
 self.server = FastMCP(server_name)
 self.connectors = {}
 self.setup_connectors()
 self.register_tools()

 def setup_connectors(self):
 """Initialize all available connectors."""
 self.connectors = {
 'figma': FigmaConnector(),
 'notion': NotionConnector(),
 'nano_banana': NanoBananaConnector(),
 'custom': CustomConnector()
 }

 def register_tools(self):
 """Register all tools from connectors."""
 for connector_name, connector in self.connectors.items():
 connector.register_tools(self.server)

# Figma Connector
class FigmaConnector:
 def __init__(self, api_key: str = None):
 self.api_key = api_key
 self.figma_client = None

 async def initialize(self):
 if self.api_key:
 self.figma_client = FigmaClient(self.api_key)

 def register_tools(self, server):
 @server.tool()
 async def extract_figma_components(
 file_id: str,
 component_names: Optional[List[str]] = None
 ) -> dict:
 """Extract React components from Figma design."""
 if not self.figma_client:
 raise ValueError("Figma API key not configured")

 file_data = await self.figma_client.get_file(file_id)
 components = self.parse_components(file_data, component_names)

 return {
 "components": components,
 "design_tokens": self.extract_design_tokens(file_data),
 "metadata": self.extract_metadata(file_data)
 }

 @server.tool()
 async def sync_figma_tokens(
 file_id: str,
 output_path: str = "src/styles/tokens.json"
 ) -> dict:
 """Sync design tokens from Figma to code."""
 file_data = await self.figma_client.get_file(file_id)
 tokens = self.extract_design_tokens(file_data)

 # Convert to W3C DTCG 2.0 format
 formatted_tokens = self.format_design_tokens(tokens)

 return {
 "tokens": formatted_tokens,
 "file_path": output_path,
 "sync_status": "success"
 }

# Notion Connector
class NotionConnector:
 def __init__(self, api_key: str = None):
 self.api_key = api_key
 self.notion_client = None

 async def initialize(self):
 if self.api_key:
 self.notion_client = NotionClient(auth=self.api_key)

 def register_tools(self, server):
 @server.tool()
 async def query_notion_database(
 database_id: str,
 query: Optional[Dict] = None
 ) -> dict:
 """Query Notion database with filtering."""
 if not self.notion_client:
 raise ValueError("Notion API key not configured")

 database = self.notion_client.databases.retrieve(database_id)
 results = self.notion_client.databases.query(
 database_id=database_id,
 query=query or {}
 )

 return {
 "results": results.get("results", []),
 "database_info": database,
 "total_count": len(results.get("results", []))
 }

 @server.tool()
 async def create_notion_page(
 database_id: str,
 properties: Dict,
 content: Optional[List[Dict]] = None
 ) -> dict:
 """Create new page in Notion database."""
 page_data = {
 "parent": {"database_id": database_id},
 "properties": properties
 }

 if content:
 page_data["children"] = content

 new_page = self.notion_client.pages.create(page_data)
 return {"page": new_page, "status": "created"}

# Nano-Banana Connector
class NanoBananaConnector:
 def __init__(self, api_key: str = None):
 self.api_key = api_key
 self.nb_client = None

 async def initialize(self):
 if self.api_key:
 self.nb_client = NanoBananaClient(api_key=self.api_key)

 def register_tools(self, server):
 @server.tool()
 async def generate_ai_content(
 prompt: str,
 model: str = "claude-3-5-sonnet-20241022",
 max_tokens: int = 4000
 ) -> dict:
 """Generate content using Nano-Banana AI models."""
 if not self.nb_client:
 raise ValueError("Nano-Banana API key not configured")

 response = await self.nb_client.chat.completions.create(
 model=model,
 messages=[{"role": "user", "content": prompt}],
 max_tokens=max_tokens
 )

 return {
 "content": response.choices[0].message.content,
 "model": model,
 "usage": response.usage
 }

 @server.tool()
 async def analyze_with_ai(
 content: str,
 analysis_type: str = "summary"
 ) -> dict:
 """Analyze content using AI models."""
 analysis_prompts = {
 "summary": "Please provide a comprehensive summary of the following content:",
 "sentiment": "Analyze the sentiment and emotional tone of this content:",
 "key_points": "Extract the key points and main ideas from this content:",
 "action_items": "Identify action items and next steps from this content:"
 }

 prompt = f"{analysis_prompts.get(analysis_type, analysis_prompts['summary'])}\n\n{content}"

 response = await self.nb_client.chat.completions.create(
 model="claude-3-5-sonnet-20241022",
 messages=[{"role": "user", "content": prompt}],
 max_tokens=2000
 )

 return {
 "analysis": response.choices[0].message.content,
 "type": analysis_type,
 "content_length": len(content)
 }
```

## Key Features

### 1. Multi-Connector Architecture
- Unified server interface
- Modular connector registration
- Dynamic tool discovery
- Configuration management

### 2. Service Integration
- Figma design system integration
- Notion knowledge base management
- Nano-Banana AI content generation
- Custom connector framework

### 3. Tool Registration
- Automatic tool discovery
- Parameter validation
- Error handling
- Response standardization

### 4. Configuration Management
- Secure API key storage
- Environment variable support
- Runtime configuration
- Service health checks

## Connector Patterns

### Standard Interface
```python
class BaseConnector:
 def __init__(self, api_key: str = None):
 self.api_key = api_key
 self.client = None

 async def initialize(self):
 """Initialize connector with API client."""
 pass

 def register_tools(self, server):
 """Register connector tools with MCP server."""
 pass

 async def health_check(self):
 """Check connector health and connectivity."""
 pass
```

### Tool Registration
- Decorator-based registration
- Automatic parameter validation
- Standardized response format
- Error handling and retry logic

## Integration Points
- FastMCP server framework
- External API clients
- Authentication systems
- Configuration management
- Monitoring and logging
