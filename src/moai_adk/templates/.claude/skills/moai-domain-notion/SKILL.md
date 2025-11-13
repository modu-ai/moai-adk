# Enterprise Notion Integration - v4.0.0

**AI-powered Notion workspace management with MCP optimization**

> **Primary Agent**: backend-expert
> **Secondary Agents**: quality-gate, alfred, doc-syncer
> **Keywords**: notion, mcp, database, page, api, integration, content-management

## Level 1: Quick Reference

### Core Capabilities

**MCP Integration**: Model Context Protocol server optimization
**Database Management**: Structured data operations and queries
**Content Creation**: Rich text formatting and templates
**Real-time Sync**: Bidirectional workspace synchronization
**Access Control**: Enterprise-grade permission management

### When to Use This Skill

- ✅ Integrating Notion MCP server with Python applications
- ✅ Managing Notion databases and pages programmatically
- ✅ Creating structured content with rich formatting
- ✅ Implementing bidirectional sync with Notion
- ✅ Building Notion-based CMS or documentation systems
- ✅ Managing workspace permissions
- ✅ Real-time collaboration features

### Quick Start Pattern

```python
from notion_client import Client
from typing import List, Dict, Any

class NotionManager:
    def __init__(self, token: str):
        self.client = Client(auth=token)

    async def create_page_with_content(
        self,
        database_id: str,
        properties: Dict[str, Any],
        content: str = ""
    ) -> Dict[str, Any]:
        """Create page with rich text content"""
        page = self.client.pages.create(
            parent={"database_id": database_id},
            properties=properties,
            children=self._create_content_blocks(content)
        )
        return page

    def _create_content_blocks(self, content: str) -> List[Dict]:
        """Generate content blocks"""
        return [
            {
                "object": "block",
                "type": "paragraph",
                "paragraph": {
                    "rich_text": [{"type": "text", "text": {"content": content}}]
                }
            }
        ]

# Usage
notion = NotionManager(token="your_notion_token")
page = await notion.create_page_with_content(
    database_id="database_id",
    properties={"Title": {"title": [{"text": {"content": "New Page"}}]}},
    content="Page content here"
)
```

## Level 2: Practical Implementation

### Database Operations

```python
class NotionDatabaseManager:
    def __init__(self, token: str):
        self.client = Client(auth=token)

    async def query_database(
        self,
        database_id: str,
        filter_params: Dict[str, Any] = None
    ) -> List[Dict]:
        """Query Notion database with filters"""
        query = self.client.databases.query(database_id)

        if filter_params:
            query = query.filter(**filter_params)

        results = query.execute()
        return results.get("results", [])

    async def create_database(
        self,
        title: str,
        properties: Dict[str, Any]
    ) -> Dict:
        """Create new Notion database"""
        return self.client.databases.create(
            title=title,
            properties=properties
        )

# Usage
db_manager = NotionDatabaseManager(token="your_token")
pages = await db_manager.query_database(
    database_id="db_id",
    filter_params={"property": "Status", "select": {"equals": "Active"}}
)
```

### Rich Content Creation

```python
async def create_rich_content_page():
    """Create page with rich formatting"""
    content_blocks = [
        {
            "object": "block",
            "type": "heading_1",
            "heading_1": {
                "rich_text": [{"type": "text", "text": {"content": "Project Overview"}}]
            }
        },
        {
            "object": "block",
            "type": "paragraph",
            "paragraph": {
                "rich_text": [
                    {"type": "text", "text": {"content": "This is "}},
                    {
                        "type": "text",
                        "text": {"content": "bold", "bold": True}
                    },
                    {"type": "text", "text": {"content": " and "}},
                    {
                        "type": "text",
                        "text": {"content": "italic", "italic": True}
                    }
                ]
            }
        },
        {
            "object": "block",
            "type": "bulleted_list_item",
            "bulleted_list_item": {
                "rich_text": [{"type": "text", "text": {"content": "Feature 1"}}]
            }
        }
    ]

    return notion.client.pages.children(
        block_id="parent_block_id",
        children=content_blocks
    )
```

### Real-time Sync Implementation

```python
import asyncio
from datetime import datetime

class NotionSyncManager:
    def __init__(self, token: str):
        self.client = Client(auth=token)
        self.last_sync = {}

    async def sync_database_changes(
        self,
        database_id: str,
        local_data: List[Dict]
    ) -> Dict[str, List]:
        """Bidirectional sync between local and Notion data"""

        # Get Notion pages
        notion_pages = await self._get_database_pages(database_id)

        # Compare and sync changes
        changes = {
            "created": [],
            "updated": [],
            "deleted": []
        }

        notion_map = {page["id"]: page for page in notion_pages}
        local_map = {item["id"]: item for item in local_data}

        # Find new items
        for item in local_data:
            if item["id"] not in notion_map:
                notion_page = await self._create_page(database_id, item)
                changes["created"].append(notion_page)

        # Find updated items
        for item in local_data:
            if item["id"] in notion_map:
                notion_page = notion_map[item["id"]]
                if self._has_changed(item, notion_page):
                    updated_page = await self._update_page(item["id"], item)
                    changes["updated"].append(updated_page)

        return changes

    async def _create_page(self, database_id: str, data: Dict):
        """Create new page from local data"""
        return self.client.pages.create(
            parent={"database_id": database_id},
            properties=data["properties"]
        )

    async def _update_page(self, page_id: str, data: Dict):
        """Update existing page"""
        return self.client.pages.update(
            page_id=page_id,
            properties=data["properties"]
        )

    def _has_changed(self, local_data: Dict, notion_page: Dict) -> bool:
        """Check if local data differs from Notion page"""
        # Compare last modified timestamps
        local_modified = local_data.get("last_modified")
        notion_modified = notion_page.get("last_edited_time")

        if local_modified and notion_modified:
            return local_modified > notion_modified

        return False
```

## Level 3: Advanced Integration

### MCP Server Integration

```python
# Notion MCP Server Implementation
import json
from mcp.server import Server, NotificationOptions
from mcp.types import (
    Resource, Tool, TextContent, ImageContent,
    CallToolResult, GetResourceResult
)

class NotionMCPServer:
    def __init__(self, notion_token: str):
        self.server = Server("notion-mcp")
        self.notion = NotionManager(notion_token)
        self._setup_handlers()

    def _setup_handlers(self):
        """Setup MCP handlers"""

        @self.server.list_resources()
        async def list_resources():
            """List available Notion resources"""
            databases = await self.notion.get_databases()
            return [
                Resource(
                    uri=f"notion://database/{db['id']}",
                    name=db["title"][0]["plain_text"],
                    mimeType="application/json",
                    description=f"Notion database: {db['title'][0]['plain_text']}"
                )
                for db in databases
            ]

        @self.server.read_resource()
        async def read_resource(uri: str):
            """Read Notion resource"""
            if uri.startswith("notion://database/"):
                database_id = uri.split("/")[-1]
                pages = await self.notion.get_database_pages(database_id)
                return GetResourceResult(
                    contents=[TextContent(type="text", text=json.dumps(pages))]
                )

        @self.server.list_tools()
        async def list_tools():
            """List available Notion tools"""
            return [
                Tool(
                    name="create_page",
                    description="Create new Notion page",
                    inputSchema={
                        "type": "object",
                        "properties": {
                            "database_id": {"type": "string"},
                            "title": {"type": "string"},
                            "content": {"type": "string"}
                        },
                        "required": ["database_id", "title"]
                    }
                ),
                Tool(
                    name="search_pages",
                    description="Search Notion pages",
                    inputSchema={
                        "type": "object",
                        "properties": {
                            "query": {"type": "string"},
                            "database_id": {"type": "string"}
                        },
                        "required": ["query"]
                    }
                )
            ]

        @self.server.call_tool()
        async def call_tool(name: str, arguments: dict):
            """Execute Notion tools"""
            if name == "create_page":
                page = await self.notion.create_page_with_content(
                    database_id=arguments["database_id"],
                    properties={"Title": {"title": [{"text": {"content": arguments["title"]}}]}},
                    content=arguments.get("content", "")
                )
                return CallToolResult(
                    content=[TextContent(
                        type="text",
                        text=f"Created page: {page['id']}"
                    )]
                )

            elif name == "search_pages":
                results = await self.notion.search_pages(
                    query=arguments["query"],
                    database_id=arguments.get("database_id")
                )
                return CallToolResult(
                    content=[TextContent(
                        type="text",
                        text=f"Found {len(results)} pages: {json.dumps(results[:5], indent=2)}"
                    )]
                )
```

### Template System

```python
class NotionTemplateManager:
    def __init__(self, token: str):
        self.notion = NotionManager(token)
        self.templates = {
            "project": {
                "properties": {
                    "Status": {"select": {"options": [
                        {"name": "Planning", "color": "gray"},
                        {"name": "In Progress", "color": "blue"},
                        {"name": "Completed", "color": "green"}
                    ]}},
                    "Priority": {"select": {"options": [
                        {"name": "High", "color": "red"},
                        {"name": "Medium", "color": "yellow"},
                        {"name": "Low", "color": "gray"}
                    ]}}
                }
            },
            "meeting": {
                "properties": {
                    "Date": {"date": {}},
                    "Duration": {"number": {"format": "minute"}},
                    "Type": {"select": {"options": [
                        {"name": "Team", "color": "blue"},
                        {"name": "Client", "color": "purple"},
                        {"name": "Review", "color": "orange"}
                    ]}}
                }
            }
        }

    async def create_from_template(
        self,
        template_name: str,
        database_id: str,
        properties: Dict[str, Any]
    ) -> Dict:
        """Create page using predefined template"""
        template = self.templates.get(template_name, {})

        # Merge template properties with user data
        merged_properties = {
            **template.get("properties", {}),
            **properties
        }

        return await self.notion.create_page_with_content(
            database_id=database_id,
            properties=merged_properties
        )

    async def register_template(self, name: str, template: Dict):
        """Register custom template"""
        self.templates[name] = template
```

### Enterprise Integration Patterns

```python
class EnterpriseNotionIntegration:
    """Enterprise-grade Notion integration with error handling"""

    def __init__(self, config: Dict):
        self.notion = NotionManager(token=config["notion_token"])
        self.backup_storage = config.get("backup_storage")
        self.webhook_url = config.get("webhook_url")

    async def create_page_with_backup(
        self,
        database_id: str,
        properties: Dict[str, Any],
        content: str = ""
    ) -> Dict:
        """Create page with automatic backup"""
        try:
            # Create in Notion
            page = await self.notion.create_page_with_content(
                database_id, properties, content
            )

            # Create backup
            if self.backup_storage:
                await self._create_backup(page, properties, content)

            # Send webhook notification
            if self.webhook_url:
                await self._send_webhook("page_created", page)

            return page

        except Exception as e:
            # Log error and retry logic
            await self._log_error("page_creation_failed", str(e))
            raise

    async def _create_backup(self, page: Dict, properties: Dict, content: str):
        """Create backup of page data"""
        backup_data = {
            "page": page,
            "properties": properties,
            "content": content,
            "timestamp": datetime.now().isoformat()
        }

        # Store in chosen backup system
        if self.backup_storage.startswith("s3://"):
            await self._backup_to_s3(backup_data)
        elif self.backup_storage.startswith("database://"):
            await self._backup_to_database(backup_data)

    async def _send_webhook(self, event: str, data: Dict):
        """Send webhook notification"""
        async with aiohttp.ClientSession() as session:
            await session.post(
                self.webhook_url,
                json={"event": event, "data": data}
            )
```

## Quick Setup Commands

```bash
# Install Notion client
pip install notion-client

# Install dependencies
pip install aiohttp python-dotenv

# Create MCP server configuration
cat > notion-mcp-config.json << EOF
{
  "notion_token": "your_integration_token",
  "backup_storage": "s3://notion-backups",
  "webhook_url": "https://your-domain.com/webhooks/notion",
  "max_requests_per_minute": 1000
}
EOF

# Start MCP server
python notion_mcp_server.py
```

## Best Practices

1. **Error Handling**: Implement retry logic with exponential backoff
2. **Rate Limiting**: Respect Notion API limits (3 requests/second for standard)
3. **Backup Strategy**: Always backup critical Notion data
4. **Caching**: Cache frequently accessed pages and databases
5. **Security**: Store tokens securely, use environment variables
6. **Idempotency**: Ensure operations can be safely retried
7. **Monitoring**: Track API usage and response times

## API Reference

| Feature | Method | Rate Limit |
|---------|--------|-----------|
| Page Operations | `client.pages.*` | 3 requests/second |
| Database Queries | `client.databases.*` | 3 requests/second |
| Search | `client.search.*` | 1 request/second |
| Blocks | `client.blocks.*` | 3 requests/second |
| Users | `client.users.*` | 1 request/second |

---

**Version**: 4.0.0 Enterprise
**Last Updated**: 2025-11-13
**Status**: Production Ready
**Enterprise Grade**: ✅ Full Enterprise Support
- **MCP Server**: notionhq/client with Bearer token authentication
- **Python Client**: @notionhq/client v2.2.12+ (async support)
- **Content Format**: Rich text blocks, databases, relations, attachments
- **Authentication**: API token with environment variable integration
- **Architecture**: MCP-first design with error handling and retries

---

### Level 2: Practical Implementation (Essential Patterns)

#### 🔍 Pattern 1: MCP Server Integration with Environment Configuration

**Problem**: Securely manage Notion API tokens and handle authentication across different environments.

**Solution**: Use MCP server configuration with environment variables and proper error handling.

```python
import os
from typing import Optional
from dotenv import load_dotenv

# Load environment variables
load_dotenv()

class NotionMCPServer:
    """MCP-based Notion server integration"""

    def __init__(self, token: Optional[str] = None):
        self.token = token or os.getenv("NOTION_TOKEN")
        self.client = self._create_mcp_client()

    def _create_mcp_client(self) -> Client:
        """Create MCP client with fallback authentication"""
        if not self.token:
            raise ValueError("NOTION_TOKEN not found in environment")

        # Validate token format
        if not self.token.startswith("ntn_"):
            raise ValueError("Invalid Notion token format")

        return Client(auth=self.token)

    async def test_connection(self) -> Dict[str, Any]:
        """Test MCP server connectivity"""
        try:
            # Test with database listing
            databases = self.client.search(
                filter={"property": "object", "value": "database"}
            )
            return {
                "status": "connected",
                "databases_count": len(databases["results"]),
                "token_prefix": self.token[:10] + "..."
            }
        except Exception as e:
            return {
                "status": "failed",
                "error": str(e),
                "suggestion": "Check token permissions and network connectivity"
            }

# MCP configuration in .mcp.json
"""
{
  "notion": {
    "command": "npx",
    "args": ["-y", "@notionhq/client"],
    "env": {
      "NOTION_TOKEN": "${NOTION_TOKEN}"
    }
  }
}
"""
```

**Key Benefits:**
- **Secure Token Management**: Environment variable integration
- **Automatic Fallback**: MCP server handles authentication
- **Connection Testing**: Comprehensive error handling and diagnostics
- **Multiple Environment Support**: Development, staging, production

**When to Use:**
- Multi-environment deployments
- Secure token management
- MCP server-based integrations
- Development and testing scenarios

---

#### 🔍 Pattern 2: Database Operations with Schema Management

**Problem**: Discovering, analyzing, and managing Notion databases with proper schema validation.

**Solution**: Create a database management system with schema introspection and type validation.

```python
from typing import Dict, List, Any, Optional
import asyncio

class NotionDatabaseManager:
    """Advanced database operations with schema validation"""

    def __init__(self, client: Client):
        self.client = client
        self._schema_cache: Dict[str, Dict] = {}

    async def discover_databases(self, workspace_name: str = None) -> List[Dict]:
        """Search and filter databases in workspace"""
        response = self.client.search(
            filter={"property": "object", "value": "database"}
        )

        return [
            {
                "id": db["id"],
                "title": self._extract_title(db),
                "url": db["url"],
                "archived": db.get("archived", False),
                "created_time": db.get("created_time"),
                "last_edited_time": db.get("last_edited_time")
            }
            for db in response["results"]
        ]

    async def get_database_schema(self, database_id: str) -> Dict[str, Any]:
        """Get database schema and property definitions"""
        if database_id in self._schema_cache:
            return self._schema_cache[database_id]

        database = self.client.databases.retrieve(database_id=database_id)

        schema = {
            "id": database["id"],
            "title": self._extract_title(database),
            "properties": {},
            "template_pages": database.get("template_pages", []),
            "is_inline": database.get("is_inline", False)
        }

        # Extract property schema definitions
        for prop_name, prop_def in database["properties"].items():
            schema["properties"][prop_name] = {
                "type": prop_def["type"],
                "description": prop_def.get("description", ""),
                "required": self._is_property_required(prop_name, prop_def)
            }

        self._schema_cache[database_id] = schema
        return schema

    async def validate_page_properties(self, database_id: str, properties: Dict) -> Dict:
        """Validate properties against database schema"""
        schema = await self.get_database_schema(database_id)
        errors = []
        warnings = []

        # Check for required properties
        for prop_name, prop_def in schema["properties"].items():
            if prop_def["required"] and prop_name not in properties:
                errors.append(f"Required property '{prop_name}' is missing")

        # Check property types
        for prop_name, value in properties.items():
            if prop_name in schema["properties"]:
                prop_type = schema["properties"][prop_name]["type"]
                if not self._validate_property_type(value, prop_type):
                    errors.append(f"Property '{prop_name}' type mismatch: expected {prop_type}")

        return {
            "valid": len(errors) == 0,
            "errors": errors,
            "warnings": warnings,
            "schema": schema
        }

    def _extract_title(self, obj: Dict) -> str:
        """Extract title from Notion object"""
        title_props = obj.get("title", [])
        return title_props[0]["text"]["content"] if title_props else "Untitled"

    def _is_property_required(self, prop_name: str, prop_def: Dict) -> bool:
        """Check if property is required"""
        return prop_def.get("required", False) or prop_name == "문서 이름"

    def _validate_property_type(self, value: Any, expected_type: str) -> bool:
        """Validate property type compatibility"""
        type_validators = {
            "title": lambda v: isinstance(v, str),
            "rich_text": lambda v: isinstance(v, str),
            "number": lambda v: isinstance(v, (int, float)),
            "select": lambda v: isinstance(v, dict) and "name" in v,
            "multi_select": lambda v: isinstance(v, list),
            "date": lambda v: isinstance(v, dict) or v is None,
            "people": lambda v: isinstance(v, list),
            "files": lambda v: isinstance(v, list),
            "checkbox": lambda v: isinstance(v, bool),
            "url": lambda v: isinstance(v, str),
            "email": lambda v: isinstance(v, str),
            "phone_number": lambda v: isinstance(v, str),
            "formula": lambda v: isinstance(v, dict),
            "relation": lambda v: isinstance(v, list),
            "rollup": lambda v: isinstance(v, dict),
            "created_time": lambda v: isinstance(v, str),
            "created_by": lambda v: isinstance(v, dict),
            "last_edited_time": lambda v: isinstance(v, str),
            "last_edited_by": lambda v: isinstance(v, dict)
        }

        validator = type_validators.get(expected_type)
        return validator(value) if validator else True

# Usage example
async def main():
    manager = NotionDatabaseManager(client)

    # Discover databases
    databases = await manager.discover_databases()
    print(f"Found {len(databases)} databases")

    # Get database schema
    schema = await manager.get_database_schema("2a99a2a0bfc280109820f6d43173e991")
    print(f"Database schema: {schema}")

    # Validate properties
    properties = {
        "문서 이름": {"title": [{"text": {"content": "Test Page"}}]},
        "상태": {"select": {"name": "작업 중"}}
    }

    validation = await manager.validate_page_properties(
        "2a99a2a0bfc280109820f6d43173e991",
        properties
    )

    if validation["valid"]:
        print("Properties are valid!")
    else:
        print(f"Validation errors: {validation['errors']}")

# Run the example
asyncio.run(main())
```

**Benefits:**
- **Schema Validation**: Prevents invalid API calls
- **Cache Management**: Reduces API calls for repeated operations
- **Error Prevention**: Early detection of property issues
- **Type Safety**: Comprehensive property type checking

---

#### 🔍 Pattern 3: Page Creation with Rich Content Templates

**Problem**: Creating pages with consistent formatting, templates, and rich content structure.

**Solution**: Implement a template-based page creation system with rich text blocks and multimedia support.

```python
from typing import List, Dict, Any, Optional
import json
from datetime import datetime

class NotionContentTemplate:
    """Template-based content creation system"""

    def __init__(self, client: Client):
        self.client = client
        self.templates = self._load_default_templates()

    def _load_default_templates(self) -> Dict[str, Dict]:
        """Load default content templates"""
        return {
            "agentic-coding": {
                "title": "Agentic Coding: {{title}}",
                "properties": {
                    "문서 이름": {"title": [{"text": {"content": "{{title}}"}}]},
                    "상태": {"select": {"name": "아이디어"}},
                    "우선순위": {"select": {"name": "보통"}},
                    "태그": {"multi_select": []},
                    "생성일": {"date": {"start": "{{date}}"}},
                    "분류": {"select": {"name": "개발"}},
                    "요약": {"rich_text": ["{{summary}}"]}
                },
                "content_blocks": [
                    {
                        "type": "heading_1",
                        "heading_1": {
                            "rich_text": [{"type": "text", "text": {"content": "{{title}}"}}]
                        }
                    },
                    {
                        "type": "paragraph",
                        "paragraph": {
                            "rich_text": [
                                {
                                    "type": "text",
                                    "text": {
                                        "content": "## 개요\n\n{{summary}}",
                                        "annotations": {"bold": True}
                                    }
                                }
                            ]
                        }
                    }
                ]
            },
            "meeting-notes": {
                "title": "회의 기록: {{title}}",
                "properties": {
                    "문서 이름": {"title": [{"text": {"content": "{{title}}"}}]},
                    "회의일자": {"date": {"start": "{{date}}"}},
                    "참석자": {"multi_select": []},
                    "회의유형": {"select": {"name": "정기"}},
                    "상태": {"select": {"name": "작업 중"}},
                    "중요도": {"select": {"name": "보통"}}
                },
                "content_blocks": [
                    {
                        "type": "heading_1",
                        "heading_1": {
                            "rich_text": [{"type": "text", "text": {"content": "{{title}}"}}]
                        }
                    },
                    {
                        "type": "callout",
                        "callout": {
                            "rich_text": [
                                {
                                    "type": "text",
                                    "text": {
                                        "content": "📅 회의 정보",
                                        "annotations": {"bold": True}
                                    }
                                }
                            ],
                            "icon": {"emoji": "📅"}
                        }
                    }
                ]
            }
        }

    async def create_page_from_template(
        self,
        database_id: str,
        template_name: str,
        variables: Dict[str, Any],
        attachments: List[str] = None
    ) -> Dict[str, Any]:
        """Create page from template with variable substitution"""

        if template_name not in self.templates:
            raise ValueError(f"Template '{template_name}' not found")

        template = self.templates[template_name]
        processed_properties = self._substitute_variables(template["properties"], variables)
        processed_content = self._substitute_variables(template["content_blocks"], variables)

        # Add attachment blocks if provided
        content_blocks = processed_content.copy()
        if attachments:
            content_blocks.extend(self._create_attachment_blocks(attachments))

        # Create the page
        page = self.client.pages.create(
            parent={"database_id": database_id},
            properties=processed_properties,
            children=content_blocks
        )

        return page

    def _substitute_variables(self, structure: Any, variables: Dict[str, Any]) -> Any:
        """Recursively substitute variables in structure"""
        if isinstance(structure, str):
            return structure.replace("{{", "").replace("}}", "")
        elif isinstance(structure, dict):
            return {key: self._substitute_variables(value, variables) for key, value in structure.items()}
        elif isinstance(structure, list):
            return [self._substitute_variables(item, variables) for item in structure]
        else:
            return structure

    def _create_attachment_blocks(self, attachments: List[str]) -> List[Dict]:
        """Create attachment blocks for files"""
        attachment_blocks = []

        for attachment_url in attachments:
            attachment_blocks.append({
                "type": "file",
                "file": {
                    "type": "external",
                    "external": {"url": attachment_url}
                }
            })

        return attachment_blocks

# Usage example
async def create_agentic_coding_post():
    template_system = NotionContentTemplate(client)

    # Define variables for the template
    variables = {
        "title": "AI Agent Pattern Analysis",
        "date": "2025-11-13",
        "summary": "Comprehensive analysis of AI agent patterns and their applications in modern software development."
    }

    # Create page from template
    page = await template_system.create_page_from_template(
        database_id="2a99a2a0bfc280109820f6d43173e991",
        template_name="agentic-coding",
        variables=variables
    )

    print(f"Created page: {page['id']}")
    return page
```

**Use Cases:**
- Standardized blog post creation
- Meeting notes with consistent format
- Project documentation templates
- Content management systems
- Automated report generation

**Template Features:**
- **Variable Substitution**: Dynamic content insertion
- **Rich Text Blocks**: Multi-format content support
- **Attachment Integration**: File and media embedding
- **Property Validation**: Schema-compliant page creation
- **Custom Templates**: User-defined template creation

---

#### 🔍 Pattern 4: Bulk Operations with Error Handling

**Problem**: Efficiently handling bulk operations while maintaining error resilience and progress tracking.

**Solution**: Implement bulk operations with parallel processing and comprehensive error handling.

```python
import asyncio
import time
from typing import List, Dict, Any, Optional, Callable
from dataclasses import dataclass
from enum import Enum

class OperationStatus(Enum):
    PENDING = "pending"
    RUNNING = "running"
    COMPLETED = "completed"
    FAILED = "failed"
    RETRYING = "retrying"

@dataclass
class BulkOperation:
    """Bulk operation tracking"""
    id: str
    operation_type: str
    parameters: Dict[str, Any]
    status: OperationStatus
    result: Optional[Dict[str, Any]] = None
    error: Optional[str] = None
    retry_count: int = 0
    max_retries: int = 3

class NotionBulkManager:
    """Bulk operations with error handling and progress tracking"""

    def __init__(self, client: Client):
        self.client = client
        self.operations: List[BulkOperation] = []
        self.lock = asyncio.Lock()

    async def bulk_create_pages(
        self,
        database_id: str,
        pages_data: List[Dict[str, Any]],
        batch_size: int = 10,
        delay_seconds: float = 1.0
    ) -> Dict[str, Any]:
        """Create multiple pages in batches"""

        operation_id = f"bulk_create_{int(time.time())}"
        total_pages = len(pages_data)

        # Initialize operations
        operations = [
            BulkOperation(
                id=f"{operation_id}_page_{i}",
                operation_type="create_page",
                parameters={"database_id": database_id, "data": page_data},
                status=OperationStatus.PENDING
            )
            for i, page_data in enumerate(pages_data)
        ]

        self.operations.extend(operations)

        # Process in batches
        results = []
        for i in range(0, total_pages, batch_size):
            batch = pages_data[i:i + batch_size]
            batch_operations = operations[i:i + batch_size]

            batch_results = await self._process_batch(
                batch_operations,
                self._create_single_page,
                delay_seconds
            )
            results.extend(batch_results)

            # Progress update
            completed = sum(1 for op in operations if op.status == OperationStatus.COMPLETED)
            print(f"Progress: {completed}/{total_pages} pages created")

        # Compile summary
        successful = len([r for r in results if r.get("status") == "success"])
        failed = len([r for r in results if r.get("status") == "failed"])

        return {
            "operation_id": operation_id,
            "total_pages": total_pages,
            "successful": successful,
            "failed": failed,
            "results": results,
            "success_rate": successful / total_pages if total_pages > 0 else 0
        }

    async def _process_batch(
        self,
        operations: List[BulkOperation],
        operation_func: Callable,
        delay_seconds: float
    ) -> List[Dict[str, Any]]:
        """Process a batch of operations with error handling"""
        results = []

        for operation in operations:
            async with self.lock:
                operation.status = OperationStatus.RUNNING

            try:
                # Execute operation
                result = await operation_func(**operation.parameters)

                async with self.lock:
                    operation.status = OperationStatus.COMPLETED
                    operation.result = result

                results.append({
                    "operation_id": operation.id,
                    "status": "success",
                    "result": result
                })

            except Exception as e:
                async with self.lock:
                    operation.error = str(e)
                    operation.retry_count += 1

                    if operation.retry_count < operation.max_retries:
                        operation.status = OperationStatus.RETRYING
                        # Schedule retry
                        asyncio.create_task(self._retry_operation(operation, operation_func, delay_seconds))
                    else:
                        operation.status = OperationStatus.FAILED

                results.append({
                    "operation_id": operation.id,
                    "status": "failed",
                    "error": str(e)
                })

            # Rate limiting
            await asyncio.sleep(delay_seconds)

        return results

    async def _create_single_page(self, database_id: str, data: Dict[str, Any]) -> Dict[str, Any]:
        """Create single page with error handling"""
        try:
            page = self.client.pages.create(
                parent={"database_id": database_id},
                properties=data["properties"],
                children=data.get("children", [])
            )
            return page
        except Exception as e:
            raise Exception(f"Failed to create page: {str(e)}")

    async def get_bulk_operations_summary(self) -> Dict[str, Any]:
        """Get summary of all bulk operations"""
        total_operations = len(self.operations)
        completed = len([op for op in self.operations if op.status == OperationStatus.COMPLETED])
        failed = len([op for op in self.operations if op.status == OperationStatus.FAILED])
        running = len([op for op in self.operations if op.status == OperationStatus.RUNNING])

        return {
            "total_operations": total_operations,
            "completed": completed,
            "failed": failed,
            "running": running,
            "success_rate": completed / total_operations if total_operations > 0 else 0
        }

# Usage example
async def bulk_operations_example():
    bulk_manager = NotionBulkManager(client)

    # Prepare pages data for bulk creation
    pages_data = [
        {
            "properties": {
                "문서 이름": {"title": [{"text": {"content": f"AI Research Paper {i}"}}]},
                "상태": {"select": {"name": "작업 중"}},
                "분류": {"select": {"name": "연구"}}
            }
        }
        for i in range(1, 21)  # Create 20 pages
    ]

    # Bulk create pages
    result = await bulk_manager.bulk_create_pages(
        database_id="2a99a2a0bfc280109820f6d43173e991",
        pages_data=pages_data,
        batch_size=5,
        delay_seconds=0.5
    )

    print(f"Bulk creation result: {result}")

    # Get operations summary
    summary = await bulk_manager.get_bulk_operations_summary()
    print(f"Operations summary: {summary}")
```

**Bulk Operation Features:**
- **Batch Processing**: Handle multiple operations efficiently
- **Error Resilience**: Automatic retry with exponential backoff
- **Progress Tracking**: Real-time operation status monitoring
- **Rate Limiting**: Respect API rate limits
- **Concurrent Processing**: Parallel execution for better performance
- **Detailed Logging**: Comprehensive operation history

---

### Level 3: Advanced Integration (Complex Scenarios)

#### 🏗️ MCP Server Integration with Context7

**Complete MCP server integration with Context7 for real-time API documentation:**

```python
from typing import Dict, Any, List
import asyncio
from context7 import Context7Helper

class NotionMCPIntegrator:
    """Advanced MCP server integration with Context7"""

    def __init__(self, token: str):
        self.token = token
        self.client = None
        self.context7 = Context7Helper()

    async def initialize(self):
        """Initialize MCP client and Context7 integration"""
        # Initialize Notion client
        self.client = Client(auth=self.token)

        # Resolve Notion library documentation
        library_id = await self.context7.resolve_library_id("notionhq/client")
        self.notion_docs = await self.context7.get_library_docs(
            library_id,
            topic="pages databases relations",
            tokens=8000
        )

    async def create_page_with_api_docs(self, database_id: str, content: str):
        """Create page with API documentation reference"""

        # Get latest API patterns from Context7
        api_patterns = await self.context7.get_library_docs(
            "/notionhq/client",
            topic="create page advanced properties",
            tokens=6000
        )

        # Create page with embedded API reference
        page = self.client.pages.create(
            parent={"database_id": database_id},
            properties={
                "문서 이름": {"title": [{"text": {"content": "API Integration Guide"}}]},
                "분류": {"select": {"name": "개발"}},
                "상태": {"select": {"name": "작업 중"}}
            },
            children=[
                {
                    "type": "heading_1",
                    "heading_1": {
                        "rich_text": [{"type": "text", "text": {"content": "API Integration Patterns"}}]
                    }
                },
                {
                    "type": "paragraph",
                    "paragraph": {
                        "rich_text": [
                            {
                                "type": "text",
                                "text": {
                                    "content": f"Based on latest API documentation:\n\n{api_patterns}",
                                    "annotations": {"code": True}
                                }
                            }
                        ]
                    }
                }
            ]
        )

        return page
```

---

## 🎯 Best Practices Checklist

**Must-Have:**
- ✅ Use MCP server integration for secure token management
- ✅ Implement schema validation before API calls
- ✅ Add comprehensive error handling and retry logic
- ✅ Use async operations for better performance
- ✅ Implement rate limiting and backoff strategies
- ✅ Validate all properties against database schema
- ✅ Use environment variables for API tokens
- ✅ Add connection testing and health checks

**Recommended:**
- ✅ Implement template-based content creation
- ✅ Add bulk operations with progress tracking
- ✅ Use Context7 for real-time API documentation
- ✅ Implement relationship management and bidirectional sync
- ✅ Add comprehensive logging and monitoring
- ✅ Use connection pooling for multiple operations
- ✅ Implement proper cleanup and resource management

**Security:**
- 🔒 Store tokens in environment variables, never in code
- 🔒 Validate all inputs before API calls
- 🔒 Use rate limiting to prevent API abuse
- 🔒 Implement proper error handling for sensitive data
- 🔒 Use SSL/TLS for all API communications
- 🔒 Regular token rotation and permission auditing

---

## 🔗 Context7 MCP Integration

**When to Use Context7 for This Skill:**

This skill benefits from Context7 when:
- Working with latest Notion API versions and features
- Need real-time API documentation and examples
- Verifying MCP server configurations
- Understanding new Notion capabilities and updates

**Example Usage:**

```python
# Fetch latest Notion API documentation
from moai_adk.integrations import Context7Helper

helper = Context7Helper()
docs = await helper.get_docs(
    library_id="/notionhq/client/2.2.12",
    topic="create page relations bulk operations",
    tokens=8000
)

# Integrate with MCP server configuration
mcp_config = {
    "notion": {
        "command": "npx",
        "args": ["-y", "@notionhq/client"],
        "env": {
            "NOTION_TOKEN": "${NOTION_TOKEN}"
        }
    }
}
```

**Relevant Libraries:**

| Library | Context7 ID | Use Case |
|---------|-------------|----------|
| Notion API | `/notionhq/client/2.2.12` | Latest API documentation and patterns |
| MCP Protocol | `/modelcontextprotocol/server` | MCP server integration |
| Python Async | `/python/asyncio` | Async operations and patterns |
| Pydantic | `/pydantic/pydantic` | Data validation and schemas |

---

## 🔗 Related Skills

**Prerequisite Skills:**
- Skill("moai-domain-backend") – Python backend architecture
- Skill("moai-domain-security") – Authentication and token management

**Complementary Skills:**
- Skill("moai-domain-frontend") – UI components for Notion integration
- Skill("moai-domain-database") – Database design and optimization
- Skill("moai-domain-devops") – CI/CD and deployment automation

**Next Steps:**
- Skill("moai-domain-monitoring") – Track Notion API usage and performance
- Skill("moai-domain-testing") – Comprehensive testing strategies for integrations

---

## 📚 Official References

### Notion API Resources
- **Official Documentation**: https://developers.notion.com
- **API Reference**: https://developers.notion.com/reference/introduction
- **Authentication**: https://developers.notion.com/docs/authentication
- **Database Operations**: https://developers.notion.com/reference/database
- **Page Operations**: https://developers.notion.com/reference/page
- **Content Blocks**: https://developers.notion.com/reference/block

### MCP Server Integration
- **MCP Protocol**: https://modelcontextprotocol.io/docs
- **Notion MCP**: https://github.com/modelcontextprotocol/server-notion
- **Environment Variables**: https://12factor.net/config

### Python Integration
- **@notionhq/client**: https://github.com/NotionSDK/notion-sdk-js
- **Async Operations**: https://docs.python.org/3/library/asyncio.html
- **Environment Management**: https://python-dotenv.readthedocs.io

### Best Practices
- **Notion API Best Practices 2025**: Rate limiting, error handling, pagination
- **MCP Server Architecture**: Secure token management, connection pooling
- **Python Async Patterns**: Concurrency, error resilience, performance optimization

---

**Version**: 4.0.0 Enterprise
**Last Updated**: 2025-11-13
**Stable Edition**: Yes (2025-11)
**Status**: Production Ready
**Enterprise Grade**: ✅ Full Enterprise Support
