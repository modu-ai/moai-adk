# MCP Server Design Patterns

Architecture and design patterns for MCP servers.

---

## Three-Component Architecture

### 1. Tools: Function Interface

**Purpose**: Expose callable functions to Claude for task execution.

**Design Principles**:
- One tool = one workflow step (not granular APIs)
- Type-safe with Pydantic validation
- Meaningful naming (verb-noun pattern)
- Clear docstrings
- Comprehensive error handling

**Tool Design Pattern**:

```python
@server.tool()
def search_documents(
    query: str,
    category: Literal["blog", "docs", "help"] = "docs",
    limit: int = Field(default=10, ge=1, le=100)
) -> dict:
    """
    Search documents by query and category.

    Args:
        query: Search query (1-200 chars)
        category: Document category filter
        limit: Max results (1-100)

    Returns:
        Dict with results, total count, and metadata
    """
    if not query or not query.strip():
        raise ValueError("Query cannot be empty")

    results = execute_search(query, category, limit)

    return {
        "status": "success",
        "query": query,
        "category": category,
        "results": results,
        "total_found": count_matches(query, category),
        "limit_applied": limit,
        "has_more": len(results) == limit
    }
```

**Tool Naming Convention**:
- `search_*` - Search/query operations
- `create_*` - Create/insert operations
- `update_*` - Modify operations
- `delete_*` - Remove operations
- `list_*` - List/enumerate operations
- `get_*` - Fetch single item

---

### 2. Resources: Data Endpoints

**Purpose**: Expose data via URI patterns for direct access.

**Design Principles**:
- URI-based (like REST resources)
- Hierarchical structure
- Type-safe parameters
- Streaming support for large data
- Permission/access control

**Resource Design Pattern**:

```python
# Simple resource
@server.resource("user://{user_id}")
def get_user(user_id: str) -> dict:
    """Get user profile."""
    user = fetch_user(user_id)
    if not user:
        raise ValueError(f"User not found: {user_id}")
    return user

# Hierarchical resource
@server.resource("org://{org_id}/team/{team_id}/members")
def get_team_members(org_id: str, team_id: str) -> list[dict]:
    """Get team members."""
    return fetch_members(org_id, team_id)

# Typed parameters
@server.resource("api://v{version:int}/stats")
def get_stats(version: int) -> dict:
    """Get API statistics for version."""
    if version < 1 or version > 3:
        raise ValueError(f"Unsupported API version: {version}")
    return fetch_stats(version)

# Streaming large data
@server.resource("export://{dataset_id}/data")
async def stream_dataset(dataset_id: str):
    """Stream large dataset."""
    async for chunk in fetch_dataset_stream(dataset_id):
        yield chunk
```

**Resource URI Patterns**:
- `{param}` - String parameter
- `{param:int}` - Integer parameter
- `{param:date}` - Date parameter
- Static prefixes: `api://`, `user://`, `org://`
- Hierarchical paths: `org/{org_id}/team/{team_id}`

---

### 3. Prompts: Conversation Templates

**Purpose**: Provide reusable system prompts and conversation templates.

**Design Principles**:
- Parameterized templates
- Context-specific instructions
- Multi-turn conversation support
- Specialized role definition

**Prompt Design Pattern**:

```python
@server.prompt("code-reviewer")
def code_review_prompt(
    language: str = "python",
    style_guide: str = "pep8"
) -> str:
    """System prompt for code review task."""
    return f"""You are an expert {language} code reviewer following {style_guide}.

Review the provided code and give constructive feedback on:
1. Code quality and readability
2. Performance and efficiency
3. Security vulnerabilities
4. Best practices and patterns
5. Documentation completeness

Format your response as:
- Summary: Brief overview
- Issues Found: List of issues with severity
- Recommendations: Suggested improvements
- Overall Quality: Score and assessment"""

@server.prompt("data-analyst")
def analyst_prompt(domain: str, expertise: str = "general") -> str:
    """System prompt for data analysis."""
    return f"""You are a {expertise} data analyst specializing in {domain}.

When analyzing data:
1. Identify patterns and trends
2. Calculate relevant metrics
3. Highlight anomalies
4. Provide business insights
5. Suggest actionable recommendations

Use clear visualizations and precise language."""
```

---

## Complete Server Architecture Example

```python
from fastmcp import FastMCP
from pydantic import Field, BaseModel
from typing import Literal, Optional, AsyncGenerator
import json

# Initialize server
server = FastMCP("enterprise-data-server")

# ============ TOOLS ============

@server.tool()
def search_records(
    query: str,
    table: Literal["users", "products", "orders"],
    limit: int = Field(default=20, ge=1, le=100),
    filters: Optional[dict] = None
) -> dict:
    """Search database records with full-text search."""
    if not query.strip():
        raise ValueError("Query cannot be empty")

    results = execute_search(query, table, limit, filters or {})

    return {
        "status": "success",
        "query": query,
        "table": table,
        "count": len(results),
        "results": results,
        "total_available": get_total_count(query, table),
        "limit": limit,
        "has_more": len(results) == limit
    }

@server.tool()
def create_record(
    table: Literal["users", "products"],
    data: dict
) -> dict:
    """Create a new record in the specified table."""
    # Validate data
    if not data:
        raise ValueError("Data cannot be empty")

    # Create record
    record_id = insert_record(table, data)

    return {
        "status": "success",
        "record_id": record_id,
        "created_at": datetime.now().isoformat(),
        "table": table
    }

# ============ RESOURCES ============

@server.resource("record://{table}/{record_id}")
def get_record(table: str, record_id: str) -> dict:
    """Fetch a single record by ID."""
    record = fetch_by_id(table, record_id)
    if not record:
        raise ValueError(f"Record not found: {table}/{record_id}")
    return record

@server.resource("report://analytics/{report_type:int}/data")
async def stream_analytics_report(report_type: int) -> AsyncGenerator[str, None]:
    """Stream large analytics report in JSON lines format."""
    async for chunk in fetch_report_stream(report_type):
        yield json.dumps(chunk) + "\n"

# ============ PROMPTS ============

@server.prompt("sql-expert")
def sql_expert_prompt(dialect: str = "postgres") -> str:
    """System prompt for SQL query writing."""
    return f"""You are an expert {dialect} SQL developer.

When writing queries:
1. Optimize for performance
2. Use proper indexing strategies
3. Handle NULL values correctly
4. Include error handling
5. Add explanatory comments

Format responses as complete, production-ready queries."""

# ============ SERVER EXECUTION ============

if __name__ == "__main__":
    server.run()
```

---

## Authentication & Authorization

### OAuth2 with Tool Protection

```python
from fastmcp.auth import OAuth2Provider

oauth = OAuth2Provider(
    authorize_url="https://auth.company.com/oauth/authorize",
    token_url="https://auth.company.com/oauth/token",
    scopes=["read:data", "write:data"]
)

@server.auth(oauth)
@server.tool()
def delete_user(user_id: str, reason: str) -> dict:
    """Delete user (requires OAuth)."""
    # Only authenticated users with proper scope can call this
    return perform_deletion(user_id, reason)
```

### API Key Authentication

```python
from fastmcp.auth import APIKeyAuth

api_auth = APIKeyAuth(header="X-API-Key")

@server.auth(api_auth)
@server.resource("admin://users/logs")
def get_audit_logs() -> list[dict]:
    """Audit logs (API key required)."""
    return fetch_audit_logs()
```

---

## Error Handling Strategy

**Consistent Error Responses**:

```python
@server.tool()
def process_data(input_file: str) -> dict:
    """Process data file."""
    try:
        # Validate input
        if not input_file:
            raise ValueError("Input file path required")

        # Check file exists
        if not os.path.exists(input_file):
            raise FileNotFoundError(f"File not found: {input_file}")

        # Process
        result = execute_processing(input_file)
        return {"status": "success", "result": result}

    except ValueError as e:
        # Input validation error
        raise ValueError(f"Invalid input: {str(e)}")

    except FileNotFoundError as e:
        # File not found
        raise ValueError(f"File error: {str(e)}")

    except Exception as e:
        # Unexpected error
        raise RuntimeError(f"Processing failed: {str(e)}")
```

---

## Pagination for Large Results

**Cursor-Based Pagination** (recommended):

```python
@server.tool()
def list_items(
    limit: int = Field(default=20, ge=1, le=100),
    cursor: Optional[str] = None
) -> dict:
    """List items with cursor-based pagination."""
    items, next_cursor = fetch_items_paginated(limit, cursor)

    return {
        "items": items,
        "limit": limit,
        "cursor": cursor,
        "next_cursor": next_cursor,
        "has_more": next_cursor is not None
    }
```

**Offset-Based Pagination** (simpler):

```python
@server.tool()
def list_items(
    limit: int = Field(default=20, ge=1, le=100),
    offset: int = Field(default=0, ge=0)
) -> dict:
    """List items with offset pagination."""
    items = fetch_items(limit, offset)
    total = count_total()

    return {
        "items": items,
        "offset": offset,
        "limit": limit,
        "total": total,
        "has_more": offset + limit < total
    }
```

---

## Type Safety with Pydantic

**Model-Based Input Validation**:

```python
from pydantic import BaseModel, Field, validator, EmailStr

class CreateUserRequest(BaseModel):
    email: EmailStr
    name: str = Field(..., min_length=1, max_length=100)
    age: int = Field(ge=18, le=120)
    role: Literal["user", "admin", "moderator"]

    @validator("name")
    def name_normalized(cls, v):
        return v.strip().title()

@server.tool()
def create_user(request: CreateUserRequest) -> dict:
    """Create user with validated model."""
    user = save_user(request)
    return {"status": "success", "user_id": user.id}
```

---

## Performance Considerations

### Connection Pooling

```python
from typing import Optional
import sqlite3

class DatabaseManager:
    def __init__(self, db_path: str, pool_size: int = 10):
        self.db_path = db_path
        self.connections = []
        for _ in range(pool_size):
            conn = sqlite3.connect(db_path)
            self.connections.append(conn)

    def get_connection(self) -> sqlite3.Connection:
        if self.connections:
            return self.connections.pop()
        return sqlite3.connect(self.db_path)

    def return_connection(self, conn: sqlite3.Connection):
        self.connections.append(conn)

db = DatabaseManager(":memory:")

@server.tool()
def query_db(sql: str) -> list[dict]:
    """Execute query with connection pooling."""
    conn = db.get_connection()
    try:
        cursor = conn.cursor()
        cursor.execute(sql)
        results = cursor.fetchall()
        return results
    finally:
        db.return_connection(conn)
```

---

## Best Practices Summary

✅ **Architecture**:
- Clear separation: Tools (functions) vs Resources (data) vs Prompts (templates)
- Consistent naming patterns
- Type-safe with Pydantic
- Comprehensive error handling

✅ **Tools Design**:
- One meaningful task per tool
- Pagination for large results
- Input validation with constraints
- Meaningful return structures

✅ **Error Handling**:
- Specific exception types
- Clear error messages
- Proper HTTP status codes
- Graceful degradation

✅ **Performance**:
- Connection pooling for databases
- Caching for frequent queries
- Pagination to limit memory
- Async/streaming for large data

---

**Last Updated**: 2025-11-27 | Production-grade patterns
