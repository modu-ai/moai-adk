# Resource Design Patterns

Guidelines for designing MCP resources.

---

## URI Pattern Design

### Hierarchical Resources

```python
# Single resource by ID
@server.resource("user://{user_id}")
def get_user(user_id: str) -> dict:
    """Get user profile."""
    return fetch_user(user_id)

# Hierarchical resources
@server.resource("org://{org_id}/team/{team_id}")
def get_team(org_id: str, team_id: str) -> dict:
    """Get team in organization."""
    return fetch_team(org_id, team_id)

# Sub-resources
@server.resource("user://{user_id}/profile")
def get_user_profile(user_id: str) -> dict:
    """Get user profile details."""
    return fetch_profile(user_id)

@server.resource("user://{user_id}/settings")
def get_user_settings(user_id: str) -> dict:
    """Get user settings."""
    return fetch_settings(user_id)

# Collections
@server.resource("org://{org_id}/users")
def get_org_users(org_id: str) -> list[dict]:
    """Get all users in organization."""
    return fetch_org_users(org_id)
```

### Typed Parameters

```python
# Integer parameters
@server.resource("post://{post_id:int}/comments")
def get_post_comments(post_id: int) -> list[dict]:
    """Get comments for post."""
    return fetch_comments(post_id)

# Date parameters
@server.resource("report://{date:date}/data")
def get_daily_report(date: str) -> dict:
    """Get report for specific date."""
    return fetch_report(date)

# Multiple typed parameters
@server.resource("api://v{version:int}/items/{item_id:int}")
def get_api_item(version: int, item_id: int) -> dict:
    """Get item from specific API version."""
    return fetch_item(version, item_id)
```

---

## Permission-Based Access

### Access Control

```python
@server.resource("private://{owner_id}/{resource_id}")
def get_private_resource(owner_id: str, resource_id: str) -> dict:
    """Get resource with ownership check."""
    # Check ownership
    if not is_owner(owner_id, resource_id):
        raise PermissionError(
            f"Access denied: {resource_id} "
            f"is not owned by {owner_id}"
        )

    return fetch_resource(resource_id)

@server.resource("admin://system/{setting_id}")
def get_admin_setting(setting_id: str) -> dict:
    """Get admin setting (requires admin role)."""
    # Check admin role
    if not is_admin():
        raise PermissionError("Admin access required")

    return fetch_setting(setting_id)
```

---

## Streaming Large Data

### Async Streaming

```python
from typing import AsyncGenerator
import json

@server.resource("export://{dataset_id}/stream")
async def stream_dataset(dataset_id: str) -> AsyncGenerator[str, None]:
    """Stream large dataset line by line."""
    try:
        async for row in fetch_dataset_stream(dataset_id):
            # Yield JSON lines format
            yield json.dumps(row) + "\n"

    except Exception as e:
        raise ValueError(f"Stream failed: {str(e)}")

@server.resource("log://{log_id}/tail")
async def stream_log_tail(log_id: str) -> AsyncGenerator[str, None]:
    """Stream log file tail."""
    async for line in fetch_log_lines(log_id):
        yield line + "\n"
```

### Chunked Responses

```python
from typing import AsyncGenerator

@server.resource("file://{file_id}/download")
async def download_file(file_id: str) -> AsyncGenerator[bytes, None]:
    """Download file in chunks."""
    chunk_size = 1024 * 1024  # 1MB chunks

    async for chunk in fetch_file_chunks(file_id, chunk_size):
        yield chunk
```

---

## Resource Metadata

### Response Structure

```python
@server.resource("data://{dataset_id}")
def get_dataset(dataset_id: str) -> dict:
    """Get dataset with metadata."""
    data = fetch_dataset(dataset_id)

    return {
        "id": dataset_id,
        "data": data,
        "metadata": {
            "size_bytes": len(str(data)),
            "row_count": len(data),
            "last_updated": get_last_update_time(dataset_id),
            "schema": get_schema(dataset_id)
        }
    }
```

### Pagination for Collections

```python
@server.resource("collection://{collection_id}/items")
def get_collection_items(
    collection_id: str,
    limit: int = 50,
    offset: int = 0
) -> dict:
    """Get collection items with pagination."""
    items = fetch_items(collection_id, limit, offset)
    total = count_items(collection_id)

    return {
        "items": items,
        "pagination": {
            "limit": limit,
            "offset": offset,
            "total": total,
            "has_more": offset + limit < total
        }
    }
```

---

## Error Handling in Resources

```python
@server.resource("secure://{resource_id}")
def get_secure_resource(resource_id: str) -> dict:
    """Get resource with comprehensive error handling."""

    # Check if resource exists
    if not resource_exists(resource_id):
        raise ValueError(
            f"Resource not found: {resource_id}. "
            f"Check the ID and try again."
        )

    # Check permissions
    if not has_access(resource_id):
        raise PermissionError(
            f"Access denied to resource: {resource_id}. "
            f"Contact administrator for access."
        )

    try:
        return fetch_resource(resource_id)

    except TimeoutError:
        raise ValueError(
            f"Failed to retrieve resource: {resource_id}. "
            f"Request timeout. Try again later."
        )

    except Exception as e:
        raise RuntimeError(
            f"Error accessing resource: {resource_id}. "
            f"Reason: {str(e)}"
        )
```

---

## Caching Considerations

### Cache Headers

```python
@server.resource("cached://{cache_key}")
def get_cached_resource(cache_key: str) -> dict:
    """Get resource with caching support."""

    # Check cache first
    cached = get_cache(cache_key)
    if cached:
        return {
            "data": cached["data"],
            "from_cache": True,
            "cached_at": cached["timestamp"]
        }

    # Fetch from source
    data = fetch_data(cache_key)

    # Cache result
    set_cache(cache_key, data, ttl=3600)  # 1 hour TTL

    return {
        "data": data,
        "from_cache": False,
        "cached_at": datetime.now().isoformat()
    }
```

---

## RESTful Design Patterns

### Resource Hierarchy

```
GET user://123              → Get user #123
GET user://123/profile      → Get user #123's profile
GET user://123/posts        → Get user #123's posts
GET post://456              → Get post #456
GET post://456/comments     → Get post #456's comments
```

### Collection Resources

```python
@server.resource("users://list")
def list_users() -> list[dict]:
    """List all users."""
    return fetch_all_users()

@server.resource("users://search/{query}")
def search_users(query: str) -> list[dict]:
    """Search users."""
    return search(query)

@server.resource("users://by-role/{role}")
def users_by_role(role: str) -> list[dict]:
    """Get users by role."""
    return fetch_by_role(role)
```

---

## Best Practices

✅ **DO**:
- Use hierarchical URIs matching domain structure
- Include meaningful type information in parameters
- Provide comprehensive error messages
- Support streaming for large data
- Include pagination for collections
- Document URI patterns clearly

❌ **DON'T**:
- Use overly complex URI patterns
- Expose sensitive data
- Return unlimited result sets
- Skip error handling
- Use non-RESTful patterns inconsistently
- Ignore pagination needs

---

**Last Updated**: 2025-11-27
