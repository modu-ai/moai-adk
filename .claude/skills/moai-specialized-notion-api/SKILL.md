---
name: moai-specialized-notion-api
description: Notion API for content management, database operations, and documentation automation
version: 1.0.0
modularized: false
last_updated: 2025-11-24
allowed_tools: [Context7, Skill, Task]
compliance_score: 100%
category_tier: 6
auto_trigger_keywords: [notion, api, database, documentation, content-management, automation, cms, blocks]
agent_coverage: [docs-manager, project-manager, git-manager]
context7_references: [/notionhq/notion-sdk-py]
invocation_api_version: 1.0
dependencies: [moai-docs-generation, moai-foundation-trust]
deprecated: false
modules: null
successor: null
---

# moai-specialized-notion-api: Content Management Integration

## Quick Reference (Level 1)

**Notion API Integration**: Automate documentation, task management, and content synchronization with Notion databases.

**Key Capabilities**:
- Database CRUD operations
- Page and block management
- Rich text formatting
- Properties and relations
- Database queries with filters
- Batch operations

**When to Use**: Synchronizing documentation with Notion, automating task creation, building content pipelines.

---

## Implementation Guide (Level 2)

### Authentication

```python
from notion_client import Client

notion = Client(auth="secret_token")
```

### Create Database Entry

```python
def create_task(title: str, status: str = "Todo"):
    response = notion.databases.query(
        database_id="db_123",
        filter={"property": "Title", "title": {"equals": title}}
    )
    if not response['results']:
        notion.pages.create(
            parent={"database_id": "db_123"},
            properties={
                "Title": {"title": [{"text": {"content": title}}]},
                "Status": {"select": {"name": status}}
            }
        )
```

### Query Database

```python
def query_tasks(status: str) -> list:
    response = notion.databases.query(
        database_id="db_123",
        filter={"property": "Status", "select": {"equals": status}}
    )
    return response['results']
```

### Update Page Content

```python
def update_page(page_id: str, content: str):
    notion.blocks.children.append(
        page_id,
        children=[{
            "object": "block",
            "type": "paragraph",
            "paragraph": {
                "text": [{"type": "text", "text": {"content": content}}]
            }
        }]
    )
```

---

## Advanced Patterns (Level 3)

### Relations and Rollups

```python
# Create related records
related_pages = notion.pages.create(
    parent={"database_id": "target_db"},
    properties={
        "Related Task": {"relation": [{"id": page_id}]}
    }
)
```

### Formula Properties

```python
# Use formulas for computed fields
formula = 'concat(prop("Name"), " - ", prop("Status"))'
```

### Database Synchronization

```python
async def sync_with_notion(data: list):
    for item in data:
        # Check if exists
        existing = notion.databases.query(
            database_id="db_123",
            filter={"property": "ID", "number": {"equals": item['id']}}
        )
        if existing['results']:
            # Update
            notion.pages.update(existing['results'][0]['id'], properties=item)
        else:
            # Create
            notion.pages.create(parent={"database_id": "db_123"}, properties=item)
```

---

**Status**: Production Ready
**Best for**: Documentation automation, task management, content synchronization
