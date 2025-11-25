# Notion Integration Reference

## API Client Setup

```python
from notion_client import AsyncClient
import os

# Initialize client
NOTION_API_KEY = os.getenv("NOTION_API_KEY")
client = AsyncClient(auth=NOTION_API_KEY)
```

---

## Core Objects

### Page Object
```python
{
    "id": "page_id",
    "created_time": "2025-11-22T10:00:00Z",
    "last_edited_time": "2025-11-22T11:00:00Z",
    "created_by": {"id": "user_id"},
    "properties": {
        "Title": {"title": [{"text": {"content": "Page Name"}}]},
        "Status": {"select": {"name": "Active"}},
        "Date": {"date": {"start": "2025-11-22"}}
    }
}
```

### Database Object
```python
{
    "id": "database_id",
    "title": [{"text": {"content": "My Database"}}],
    "properties": {
        "Name": {"title": {}},
        "Status": {"select": {"options": [...]}},
        "Owner": {"people": {}}
    }
}
```

### Block Object
```python
{
    "id": "block_id",
    "type": "paragraph",
    "paragraph": {
        "rich_text": [{"text": {"content": "Text content"}}],
        "color": "default"
    }
}
```

---

## Common Operations

### Create Page
```python
client.pages.create(
    parent={"database_id": "db_id"},
    properties={...}
)
```

### Update Page
```python
client.pages.update(
    page_id="page_id",
    properties={...}
)
```

### Query Database
```python
client.databases.query(
    database_id="db_id",
    filter={...},
    sorts=[...],
    page_size=100
)
```

### Add Blocks
```python
client.blocks.children.append(
    block_id="page_id",
    children=[...]
)
```

### Search
```python
client.search(query="search term")
```

---

## Property Types

| Type | Example |
|------|---------|
| title | `{"title": [{"text": {"content": "Title"}}]}` |
| rich_text | `{"rich_text": [{"text": {"content": "Text"}}]}` |
| select | `{"select": {"name": "Option"}}` |
| multi_select | `{"multi_select": [{"name": "Option1"}, {"name": "Option2"}]}` |
| checkbox | `{"checkbox": true}` |
| number | `{"number": 42}` |
| date | `{"date": {"start": "2025-11-22"}}` |
| people | `{"people": [{"id": "user_id"}]}` |
| url | `{"url": "https://example.com"}` |
| email | `{"email": "user@example.com"}` |
| relation | `{"relation": [{"id": "page_id"}]}` |

---

## Error Handling

### Common Errors
- `APIResponseError`: API call failed
- `RateLimitError`: Too many requests (429)
- `ObjectNotFound`: Resource not found
- `ValidationError`: Invalid input

### Rate Limits
- 3 requests per second per integration
- Batch operations recommended

---

**Version**: 4.0.0 | **Last Updated**: 2025-11-22
