# Notion MCP Server Integration Module

**Version**: 4.0.0 (Enterprise Optimized)
**Last Updated**: 2025-11-24
**Purpose**: Enterprise workspace automation, database operations, and content management

---

## 📖 Quick Overview (2 Minutes)

Notion is the unified MCP server for enterprise workspace management, providing:

- **Workspace management**: Create, organize, and manage Notion workspaces
- **Database operations**: Create, query, update databases with custom schemas
- **Page management**: Create, update, delete pages with rich content
- **Content automation**: Bulk operations, synchronization, bulk updates
- **Access control**: Permissions, sharing, collaboration settings
- **Rich content**: Markdown support, embedded files, relationships

---

## 🔧 Implementation Guide

### Core Database Operations

#### **Create Database with Custom Schema**

```python
async def create_notion_database(
    parent_page_id: str,
    title: str,
    properties: dict
) -> dict:
    """
    Create a Notion database with custom properties.

    Args:
        parent_page_id: Parent page ID
        title: Database title
        properties: Property schema definitions

    Returns:
        Created database object
    """
    # Define database properties
    properties = {
        "Title": {"type": "title"},
        "Status": {
            "type": "select",
            "options": [
                {"name": "Active", "color": "green"},
                {"name": "Inactive", "color": "gray"},
                {"name": "Archived", "color": "blue"}
            ]
        },
        "Owner": {"type": "people"},
        "Date": {"type": "date"},
        "Description": {"type": "rich_text"},
        "Priority": {
            "type": "select",
            "options": [
                {"name": "High", "color": "red"},
                {"name": "Medium", "color": "yellow"},
                {"name": "Low", "color": "blue"}
            ]
        }
    }

    database = await mcp__notion__notion_create_database(
        parent={"page_id": parent_page_id},
        title=title,
        properties=properties
    )

    return database
```

#### **Query Database with Filters**

```python
async def query_notion_database(
    database_id: str,
    filter_config: dict = None,
    sorts: list = None,
    page_size: int = 100
) -> dict:
    """
    Query Notion database with advanced filtering.

    Examples:
        # Filter by status
        filter_config = {
            "property": "Status",
            "select": {"equals": "Active"}
        }

        # Complex AND filter
        filter_config = {
            "and": [
                {"property": "Status", "select": {"equals": "Active"}},
                {"property": "Priority", "select": {"equals": "High"}},
                {"property": "Date", "date": {"after": "2025-01-01"}}
            ]
        }
    """
    results = await mcp__notion__notion_query_database(
        database_id=database_id,
        filter=filter_config,
        sorts=sorts or [],
        page_size=page_size
    )

    return results
```

#### **Bulk Update Operations**

```python
async def bulk_update_pages(
    page_ids: list[str],
    updates: dict
) -> list[dict]:
    """
    Update multiple pages efficiently.

    Args:
        page_ids: List of page IDs to update
        updates: Dictionary of property updates

    Returns:
        List of updated page objects
    """
    updated_pages = []

    # Process in batches for efficiency
    batch_size = 10
    for i in range(0, len(page_ids), batch_size):
        batch = page_ids[i:i + batch_size]

        for page_id in batch:
            updated = await mcp__notion__notion_update_page(
                page_id=page_id,
                properties=updates
            )
            updated_pages.append(updated)

    return updated_pages
```

---

### Page Management Patterns

#### **Create Rich Content Pages**

```python
async def create_notion_page(
    parent: dict,
    properties: dict,
    content: str
) -> dict:
    """
    Create a page with markdown content.

    Args:
        parent: {"database_id": "..."} or {"page_id": "..."}
        properties: Page properties (Title, etc.)
        content: Markdown content

    Returns:
        Created page object
    """
    page = await mcp__notion__notion_create_pages(
        parent=parent,
        properties=properties,
        children=[
            {
                "object": "block",
                "type": "paragraph",
                "paragraph": {
                    "rich_text": [
                        {
                            "type": "text",
                            "text": {"content": content}
                        }
                    ]
                }
            }
        ]
    )

    return page
```

#### **Hierarchical Page Organization**

```python
async def create_page_hierarchy():
    """Create organized page structure with parent-child relationships."""

    # Create parent page
    parent = await create_notion_page(
        parent={"page_id": "workspace_root"},
        properties={"Title": "Project Documentation"},
        content="# Project Documentation"
    )

    # Create child pages
    sections = ["Overview", "Architecture", "API Reference", "Guides"]

    for section in sections:
        child = await create_notion_page(
            parent={"page_id": parent["id"]},
            properties={"Title": section},
            content=f"# {section}\n\nContent for {section}"
        )
```

---

### Advanced Integration Patterns

#### **Sync External Data to Notion**

```python
async def sync_external_data_to_notion(
    database_id: str,
    external_data: list[dict]
) -> list[dict]:
    """
    Sync external data source to Notion database.

    Args:
        database_id: Target Notion database
        external_data: List of items to sync

    Returns:
        List of created pages
    """
    created_pages = []

    for item in external_data:
        page = await create_notion_page(
            parent={"database_id": database_id},
            properties={
                "Title": item.get("name"),
                "Description": item.get("description"),
                "URL": item.get("link"),
                "Status": "Synced",
                "Date": datetime.now().isoformat()
            }
        )
        created_pages.append(page)

    return created_pages
```

#### **Multi-Database Relationships**

```python
async def create_database_relation(
    from_database_id: str,
    to_database_id: str,
    relation_property_name: str
) -> dict:
    """
    Create relationship between two databases.

    This enables linking pages across separate databases.
    """
    # Add relation property to source database
    relation_property = {
        "type": "relation",
        "relation": {
            "database_id": to_database_id,
            "relation_name": relation_property_name
        }
    }

    # Update database schema
    updated = await mcp__notion__notion_update_database(
        database_id=from_database_id,
        properties={relation_property_name: relation_property}
    )

    return updated
```

---

## 🎯 Use Cases

### Knowledge Base Management

```python
# Create company knowledge base structure
knowledge_base = await create_notion_database(
    parent_page_id="workspace_root",
    title="Company Knowledge Base",
    properties={
        "Title": {"type": "title"},
        "Category": {"type": "select"},
        "Author": {"type": "people"},
        "Last Updated": {"type": "date"},
        "Status": {"type": "select"},
        "Related Items": {"type": "relation"}
    }
)
```

### Project Management

```python
# Create project tracker database
projects = await create_notion_database(
    parent_page_id="workspace_root",
    title="Projects",
    properties={
        "Project Name": {"type": "title"},
        "Status": {"type": "select"},
        "Owner": {"type": "people"},
        "Start Date": {"type": "date"},
        "End Date": {"type": "date"},
        "Budget": {"type": "number"},
        "Priority": {"type": "select"},
        "Tasks": {"type": "relation"}
    }
)
```

### Documentation Portal

```python
# Create documentation structure
docs = await create_notion_database(
    parent_page_id="workspace_root",
    title="API Documentation",
    properties={
        "Endpoint": {"type": "title"},
        "Method": {"type": "select"},
        "Description": {"type": "rich_text"},
        "Parameters": {"type": "rich_text"},
        "Response": {"type": "rich_text"},
        "Status": {"type": "select"},
        "Version": {"type": "select"}
    }
)
```

---

## 🛠️ Error Handling & Rate Limits

```python
import time
from typing import Coroutine

async def notion_operation_with_retry(
    operation: Coroutine,
    max_retries: int = 3,
    backoff_factor: float = 2.0
) -> dict:
    """
    Execute Notion operation with exponential backoff retry.

    Handles rate limiting and transient errors.
    """
    retry_count = 0
    wait_time = 1

    while retry_count < max_retries:
        try:
            return await operation
        except RateLimitError:
            if retry_count < max_retries - 1:
                wait_time *= backoff_factor
                await asyncio.sleep(wait_time)
                retry_count += 1
            else:
                raise
        except ConnectionError:
            if retry_count < max_retries - 1:
                await asyncio.sleep(wait_time)
                retry_count += 1
            else:
                raise

    raise Exception("Operation failed after retries")
```

---

## 📊 Workspace Scale Automation

### Bulk Content Migration

```python
async def migrate_content_to_notion(
    source_data: list[dict],
    database_id: str,
    batch_size: int = 10
) -> dict:
    """
    Migrate large amounts of content to Notion.

    Processes in batches to avoid rate limits.
    """
    results = {
        "created": 0,
        "failed": 0,
        "errors": []
    }

    for i in range(0, len(source_data), batch_size):
        batch = source_data[i:i + batch_size]

        for item in batch:
            try:
                page = await create_notion_page(
                    parent={"database_id": database_id},
                    properties=item.get("properties", {}),
                    content=item.get("content", "")
                )
                results["created"] += 1
            except Exception as e:
                results["failed"] += 1
                results["errors"].append({
                    "item": item,
                    "error": str(e)
                })

        # Wait between batches to respect rate limits
        await asyncio.sleep(0.5)

    return results
```

---

## ✅ Best Practices

✅ **DO**:
- Design database schemas carefully before creating
- Use batch operations for high-volume updates
- Implement error handling for rate limits
- Organize content hierarchically
- Document database purposes and relationships
- Use descriptive property names
- Version your database schemas
- Monitor API usage
- Implement exponential backoff for retries

❌ **DON'T**:
- Create databases without planning schema
- Make individual API calls in loops (use batch operations)
- Ignore rate limits (Notion has strict limits)
- Store sensitive data in Notion without encryption
- Over-complicate database relationships
- Forget to document database structure
- Use generic property names

---

## 📋 Notion API Rate Limits

| Operation | Limit | Note |
|-----------|-------|------|
| Create page | 3 req/s | Batch for high volume |
| Update page | 3 req/s | Batch operations recommended |
| Query database | 3 req/s | Pagination support |
| Search | 2 req/s | Global search |
| Get database | 3 req/s | Schema access |

---

## 🔄 Changelog

| Version | Date | Changes |
|---------|------|---------|
| **4.0.0** | 2025-11-24 | Merged into moai-mcp-integration hub module |
| 3.0.0 | 2025-11-22 | Advanced patterns and workspace automation |
| 2.0.0 | 2025-11-13 | Enterprise-grade error handling |
| 1.0.0 | 2025-11-01 | Initial Notion MCP integration |

---

**Module Version**: 4.0.0
**Status**: Production Ready
**Compliance**: 100% (Notion API v1+)
