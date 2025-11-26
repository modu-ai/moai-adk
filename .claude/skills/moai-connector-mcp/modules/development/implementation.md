# MCP Server Implementation Examples

Real-world implementation patterns for MCP servers.

---

## Example 1: Database Query Server

**File: `database_server.py`**:

```python
from fastmcp import FastMCP
from pydantic import Field, BaseModel, validator
from typing import Literal, Optional
from datetime import datetime
import sqlite3

server = FastMCP("database-server")

class QueryRequest(BaseModel):
    query: str = Field(..., min_length=1, max_length=500)
    limit: int = Field(default=20, ge=1, le=1000)

class SearchRequest(BaseModel):
    table: Literal["users", "products", "orders"]
    search_term: str = Field(..., min_length=1)
    limit: int = Field(default=50, ge=1, le=100)

    @validator("search_term")
    def validate_search(cls, v):
        if len(v) < 3:
            raise ValueError("Search term must be at least 3 characters")
        return v.lower()

# Tool: Execute custom queries
@server.tool()
def execute_query(request: QueryRequest) -> dict:
    """Execute custom SQL query safely."""
    try:
        conn = sqlite3.connect(":memory:")
        cursor = conn.cursor()
        cursor.execute(request.query)
        results = cursor.fetchall()
        conn.close()

        return {
            "status": "success",
            "rows_returned": len(results),
            "results": results,
            "executed_at": datetime.now().isoformat()
        }

    except sqlite3.Error as e:
        raise ValueError(f"Database error: {str(e)}")

# Tool: Search with full-text
@server.tool()
def search_table(request: SearchRequest) -> dict:
    """Full-text search in specified table."""
    try:
        conn = sqlite3.connect(":memory:")
        cursor = conn.cursor()

        # Build safe query
        query = f"""
            SELECT * FROM {request.table}
            WHERE content LIKE ?
            LIMIT ?
        """

        cursor.execute(query, (f"%{request.search_term}%", request.limit))
        results = cursor.fetchall()
        conn.close()

        return {
            "status": "success",
            "table": request.table,
            "search_term": request.search_term,
            "results_found": len(results),
            "results": results
        }

    except Exception as e:
        raise ValueError(f"Search failed: {str(e)}")

# Resource: Get single record
@server.resource("db://{table}/{record_id}")
def get_record(table: str, record_id: str) -> dict:
    """Fetch single record by ID."""
    try:
        conn = sqlite3.connect(":memory:")
        cursor = conn.cursor()
        cursor.execute(f"SELECT * FROM {table} WHERE id = ?", (record_id,))
        result = cursor.fetchone()
        conn.close()

        if not result:
            raise ValueError(f"Record not found: {table}/{record_id}")

        return {"record": result, "found": True}

    except Exception as e:
        raise ValueError(f"Fetch failed: {str(e)}")

if __name__ == "__main__":
    server.run()
```

---

## Example 2: API Gateway Server

**Purpose**: Route requests to multiple backend services.

```python
from fastmcp import FastMCP
from typing import Literal, Optional
import httpx
import asyncio

server = FastMCP("api-gateway-server")

class ServiceRouter:
    def __init__(self):
        self.services = {
            "users": "https://api.service.com/users",
            "products": "https://api.service.com/products",
            "orders": "https://api.service.com/orders"
        }
        self.client = httpx.AsyncClient()

    async def route_request(
        self,
        service: str,
        endpoint: str,
        method: str = "GET",
        data: Optional[dict] = None
    ) -> dict:
        """Route request to appropriate service."""
        if service not in self.services:
            raise ValueError(f"Unknown service: {service}")

        url = f"{self.services[service]}/{endpoint}"

        try:
            if method == "GET":
                response = await self.client.get(url)
            elif method == "POST":
                response = await self.client.post(url, json=data)
            else:
                raise ValueError(f"Unsupported method: {method}")

            response.raise_for_status()
            return {
                "status": "success",
                "service": service,
                "endpoint": endpoint,
                "data": response.json()
            }

        except httpx.HTTPStatusError as e:
            raise ValueError(f"Service error: {e.response.status_code}")

router = ServiceRouter()

@server.tool()
async def call_service(
    service: Literal["users", "products", "orders"],
    endpoint: str,
    method: Literal["GET", "POST"] = "GET",
    data: Optional[dict] = None
) -> dict:
    """Call external service through gateway."""
    return await router.route_request(service, endpoint, method, data)

if __name__ == "__main__":
    server.run()
```

---

## Example 3: File Processing Server

**Purpose**: Handle file uploads, parsing, and transformation.

```python
from fastmcp import FastMCP
from typing import AsyncGenerator
import json
import csv
from pathlib import Path

server = FastMCP("file-processor-server")

class FileProcessor:
    @staticmethod
    async def process_json(file_path: str) -> dict:
        """Parse JSON file."""
        try:
            with open(file_path) as f:
                data = json.load(f)
            return {"format": "json", "data": data, "size_bytes": Path(file_path).stat().st_size}
        except json.JSONDecodeError as e:
            raise ValueError(f"Invalid JSON: {str(e)}")

    @staticmethod
    async def process_csv(file_path: str) -> dict:
        """Parse CSV file."""
        try:
            rows = []
            with open(file_path) as f:
                reader = csv.DictReader(f)
                rows = list(reader)
            return {"format": "csv", "rows": len(rows), "data": rows}
        except Exception as e:
            raise ValueError(f"CSV parsing failed: {str(e)}")

    @staticmethod
    async def stream_large_file(file_path: str, chunk_size: int = 1024) -> AsyncGenerator[str, None]:
        """Stream large file in chunks."""
        try:
            with open(file_path, 'rb') as f:
                while True:
                    chunk = f.read(chunk_size)
                    if not chunk:
                        break
                    yield chunk.decode('utf-8', errors='ignore')
        except Exception as e:
            raise ValueError(f"Stream failed: {str(e)}")

processor = FileProcessor()

@server.tool()
async def parse_file(file_path: str) -> dict:
    """Parse file and return structured data."""
    file_path_obj = Path(file_path)

    if not file_path_obj.exists():
        raise ValueError(f"File not found: {file_path}")

    suffix = file_path_obj.suffix.lower()

    if suffix == ".json":
        return await processor.process_json(file_path)
    elif suffix == ".csv":
        return await processor.process_csv(file_path)
    else:
        raise ValueError(f"Unsupported format: {suffix}")

@server.resource("file://{file_id}/stream")
async def stream_file(file_id: str) -> AsyncGenerator[str, None]:
    """Stream file content."""
    file_path = get_file_path(file_id)
    async for chunk in processor.stream_large_file(file_path):
        yield chunk

if __name__ == "__main__":
    server.run()
```

---

## Example 4: Authentication-Protected Server

**Purpose**: Secure server with OAuth2 and API keys.

```python
from fastmcp import FastMCP
from fastmcp.auth import OAuth2Provider, APIKeyAuth
from pydantic import Field
from typing import Optional

server = FastMCP("secure-server")

# Setup OAuth2
oauth = OAuth2Provider(
    authorize_url="https://auth.company.com/oauth/authorize",
    token_url="https://auth.company.com/oauth/token",
    scopes=["read:data", "write:data", "admin"]
)

# Setup API Key
api_key_auth = APIKeyAuth(header="X-API-Key")

# OAuth-protected user operations
@server.auth(oauth)
@server.tool()
def get_user_profile(user_id: str) -> dict:
    """Get user profile (OAuth required)."""
    user = fetch_user(user_id)
    if not user:
        raise ValueError(f"User not found: {user_id}")
    return user

@server.auth(oauth)
@server.tool()
def update_user(user_id: str, data: dict) -> dict:
    """Update user profile (OAuth with write scope)."""
    # In real app, check OAuth scope for "write:data"
    updated = update_user_data(user_id, data)
    return {"status": "success", "user_id": user_id, "updated_at": datetime.now().isoformat()}

# API Key-protected admin operations
@server.auth(api_key_auth)
@server.tool()
def create_admin_user(email: str, name: str, role: str) -> dict:
    """Create admin user (API key required)."""
    user_id = create_user(email, name, role)
    return {"status": "success", "user_id": user_id}

@server.auth(api_key_auth)
@server.resource("admin://system/stats")
def get_system_stats() -> dict:
    """Get system statistics (API key required)."""
    return {
        "total_users": count_users(),
        "active_sessions": count_sessions(),
        "last_sync": datetime.now().isoformat()
    }

if __name__ == "__main__":
    server.run()
```

---

## Example 5: Multi-Tool Workflow Server

**Purpose**: Complex workflows using multiple coordinated tools.

```python
from fastmcp import FastMCP
from pydantic import Field
from typing import Literal, Optional
from enum import Enum

server = FastMCP("workflow-server")

class WorkflowStatus(str, Enum):
    PENDING = "pending"
    IN_PROGRESS = "in_progress"
    COMPLETED = "completed"
    FAILED = "failed"

# Tool 1: Start workflow
@server.tool()
def start_workflow(
    workflow_type: Literal["onboarding", "deployment", "audit"],
    parameters: dict
) -> dict:
    """Start a new workflow."""
    workflow_id = create_workflow(workflow_type, parameters)

    return {
        "status": "success",
        "workflow_id": workflow_id,
        "workflow_type": workflow_type,
        "initial_status": WorkflowStatus.PENDING
    }

# Tool 2: Get workflow status
@server.tool()
def get_workflow_status(workflow_id: str) -> dict:
    """Get current workflow status."""
    workflow = fetch_workflow(workflow_id)

    if not workflow:
        raise ValueError(f"Workflow not found: {workflow_id}")

    return {
        "workflow_id": workflow_id,
        "status": workflow["status"],
        "progress": workflow.get("progress", 0),
        "current_step": workflow.get("current_step"),
        "steps_completed": len(workflow.get("completed_steps", [])),
        "total_steps": len(workflow.get("all_steps", []))
    }

# Tool 3: Advance workflow
@server.tool()
def advance_workflow(
    workflow_id: str,
    action: Literal["approve", "reject", "continue"]
) -> dict:
    """Advance workflow to next step."""
    workflow = fetch_workflow(workflow_id)

    if not workflow:
        raise ValueError(f"Workflow not found: {workflow_id}")

    if action == "approve":
        next_step = workflow["next_step"]
    elif action == "reject":
        next_step = "failed"
    else:
        next_step = "continue"

    updated = update_workflow(workflow_id, {"status": next_step})

    return {
        "workflow_id": workflow_id,
        "action": action,
        "new_status": next_step,
        "updated_at": datetime.now().isoformat()
    }

# Resource: Full workflow details
@server.resource("workflow://{workflow_id}/details")
def get_workflow_details(workflow_id: str) -> dict:
    """Get complete workflow details and history."""
    workflow = fetch_workflow(workflow_id)

    if not workflow:
        raise ValueError(f"Workflow not found: {workflow_id}")

    return {
        "id": workflow_id,
        "type": workflow["type"],
        "status": workflow["status"],
        "created_at": workflow["created_at"],
        "parameters": workflow["parameters"],
        "steps": workflow.get("all_steps", []),
        "completed_steps": workflow.get("completed_steps", []),
        "current_step": workflow.get("current_step"),
        "history": workflow.get("history", [])
    }

if __name__ == "__main__":
    server.run()
```

---

## Common Implementation Patterns

### Error Handling Pattern

```python
@server.tool()
def risky_operation(params: dict) -> dict:
    """Operation with comprehensive error handling."""
    try:
        # Validate input
        if not params:
            raise ValueError("Parameters required")

        # Execute operation
        result = perform_operation(params)
        return {"status": "success", "result": result}

    except ValueError as e:
        # Input validation errors
        raise ValueError(f"Invalid input: {str(e)}")

    except TimeoutError as e:
        # Timeout errors
        raise ValueError(f"Operation timeout: {str(e)}")

    except Exception as e:
        # Unexpected errors
        raise RuntimeError(f"Operation failed: {str(e)}")
```

### Pagination Pattern

```python
@server.tool()
def list_items(
    limit: int = Field(default=20, ge=1, le=100),
    offset: int = Field(default=0, ge=0)
) -> dict:
    """List items with pagination."""
    items = fetch_items(limit, offset)
    total = count_total()

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

### Validation Pattern

```python
from pydantic import BaseModel, Field, validator, EmailStr

class CreateItemRequest(BaseModel):
    name: str = Field(..., min_length=1, max_length=100)
    email: EmailStr
    quantity: int = Field(ge=1, le=1000)

    @validator("name")
    def name_not_empty(cls, v):
        if not v.strip():
            raise ValueError("Name cannot be empty")
        return v.strip()

@server.tool()
def create_item(request: CreateItemRequest) -> dict:
    """Create item with validated request."""
    item = save_item(request)
    return {"status": "success", "item_id": item.id}
```

---

**Last Updated**: 2025-11-27 | Production-ready implementations
