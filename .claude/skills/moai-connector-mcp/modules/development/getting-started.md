# FastMCP Getting Started

Quick start guide for building your first MCP server.

---

## Installation

**Requirements**:
- Python 3.9+
- pip or uv

**Install FastMCP**:
```bash
pip install fastmcp
```

**Verify Installation**:
```bash
python -c "import fastmcp; print(fastmcp.__version__)"
```

---

## Minimal Server (2 Minutes)

**File: `hello_server.py`**:

```python
from fastmcp import FastMCP

# Create server instance
server = FastMCP("hello-server")

# Define a tool
@server.tool()
def greet(name: str) -> str:
    """Greet someone by name."""
    return f"Hello, {name}!"

# Run the server
if __name__ == "__main__":
    server.run()
```

**Start the server**:
```bash
python hello_server.py
```

**Expected output**:
```
FastMCP server "hello-server" started (stdio mode)
Waiting for MCP client connections...
```

---

## Core Building Blocks

### 1. Tools (Functions)

Functions that Claude can invoke.

```python
from fastmcp import FastMCP
from pydantic import Field

server = FastMCP("tools-demo")

# Simple tool
@server.tool()
def multiply(a: int, b: int) -> int:
    """Multiply two numbers."""
    return a * b

# Tool with validation
@server.tool()
def search_users(
    query: str,
    limit: int = Field(default=10, ge=1, le=100)
) -> list[dict]:
    """Search users with pagination."""
    return execute_search(query, limit)

# Tool with enums (constrained choices)
from typing import Literal

@server.tool()
def get_report(
    report_type: Literal["sales", "inventory", "finances"]
) -> dict:
    """Get specific report type."""
    return fetch_report(report_type)
```

**Tool Best Practices**:
- ✅ One tool = one meaningful task
- ✅ Clear docstring (first line is summary)
- ✅ Type annotations (required for schema generation)
- ✅ Constrained parameters (min/max, choices)
- ✅ Meaningful return types

---

### 2. Resources (Data Endpoints)

URI-based data endpoints.

```python
@server.resource("user://{user_id}")
def get_user_profile(user_id: str) -> dict:
    """Fetch user profile by ID."""
    return fetch_user(user_id)

@server.resource("post://{post_id}/comments")
def get_post_comments(post_id: str) -> list[dict]:
    """Get comments for a post."""
    return fetch_comments(post_id)

# Typed URI parameters
@server.resource("api://data/{version:int}/items")
def get_items(version: int) -> dict:
    """Get items for specific API version."""
    return fetch_items(version)
```

**Resource URI Patterns**:
- `user://{user_id}` - Simple parameter
- `db://{table}/{record_id}` - Multiple parameters
- `api://v{version:int}/endpoint` - Typed parameters
- `data://{date:date}/report` - Date parameters

---

### 3. Prompts (Templates)

Reusable conversation templates.

```python
@server.prompt("code-reviewer")
def code_review_prompt(language: str = "python") -> str:
    """System prompt for code review."""
    return f"""You are an expert {language} code reviewer.

    Focus on:
    - Code quality and readability
    - Performance and efficiency
    - Security vulnerabilities
    - Best practices and patterns

    Provide constructive feedback."""

@server.prompt("data-analyst")
def analyst_prompt(domain: str) -> str:
    """System prompt for data analysis."""
    return f"""You are a {domain} data analyst.
    Analyze data systematically and provide insights."""
```

---

## Project Structure

**Recommended layout**:

```
my-mcp-server/
├── server.py           # Main server file
├── requirements.txt    # Dependencies
├── .env.example       # Environment variables
├── tests/
│   ├── __init__.py
│   ├── test_tools.py
│   └── test_resources.py
├── Dockerfile         # Container setup
└── README.md         # Documentation
```

---

## Common Patterns

### Error Handling

```python
@server.tool()
def divide(a: float, b: float) -> float:
    """Divide two numbers."""
    if b == 0:
        raise ValueError("Cannot divide by zero")
    return a / b
```

### Input Validation

```python
from pydantic import BaseModel, validator, Field

class UserQuery(BaseModel):
    email: str = Field(..., regex=r"^[^@]+@[^@]+$")
    age: int = Field(ge=0, le=150)
    name: str = Field(min_length=1, max_length=100)

    @validator("name")
    def name_not_empty(cls, v):
        if not v.strip():
            raise ValueError("Name cannot be only whitespace")
        return v.strip()

@server.tool()
def create_user(query: UserQuery) -> dict:
    """Create user with validation."""
    return save_user(query)
```

### Pagination

```python
@server.tool()
def list_items(
    limit: int = Field(default=20, ge=1, le=100),
    offset: int = Field(default=0, ge=0)
) -> dict:
    """List items with pagination."""
    items = fetch_items(limit, offset)
    total = count_total_items()

    return {
        "items": items,
        "limit": limit,
        "offset": offset,
        "total": total,
        "has_more": offset + limit < total
    }
```

---

## Configuration

### Command-Line Server

**Default (stdio)**:
```python
if __name__ == "__main__":
    server.run()
```

### HTTP Server

```python
if __name__ == "__main__":
    server.run_http(host="0.0.0.0", port=8000)
```

### Configuration via Environment

```python
import os

server = FastMCP(
    name="my-server",
    version=os.getenv("SERVER_VERSION", "1.0.0")
)

# Access environment variables
api_key = os.getenv("API_KEY")
if not api_key:
    raise ValueError("API_KEY environment variable not set")
```

---

## Testing Locally

### Interactive Testing

```python
# Add to your server file for testing
if __name__ == "__main__":
    import sys

    if len(sys.argv) > 1 and sys.argv[1] == "test":
        # Test mode
        result = server.invoke_tool("greet", {"name": "World"})
        print(f"Result: {result}")
    else:
        # Normal mode
        server.run()
```

**Run tests**:
```bash
python server.py test
```

---

## Configuration in Claude Code

**`.claude/settings.json`** (Claude Code configuration):

```json
{
  "mcpServers": {
    "my-server": {
      "command": "python",
      "args": ["/path/to/server.py"],
      "env": {
        "API_KEY": "${API_KEY}",
        "LOG_LEVEL": "info"
      }
    }
  }
}
```

---

## Next Steps

1. **Learn Architecture**: [server-design.md](../development/server-design.md)
2. **Build Examples**: [implementation.md](../development/implementation.md)
3. **Test & Validate**: [testing.md](../development/testing.md)
4. **Deploy**: [deployment.md](../development/deployment.md)
5. **Advanced Patterns**: [patterns/](../development/patterns/)

---

**Last Updated**: 2025-11-27 | Ready for production use
