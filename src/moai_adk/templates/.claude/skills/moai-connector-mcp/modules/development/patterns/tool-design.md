# Tool Design Best Practices

Guidelines for designing effective MCP tools.

---

## Workflow-Optimized Tool Design

### Principle: One Tool = One Workflow Step

**Bad**: Tool does too much
```python
@server.tool()
def process_all_data(files: list[str], format: str, filters: dict,
                     sort_by: str, limit: int) -> dict:
    """Process, filter, and transform data (too complex)."""
    # Complex implementation with many responsibilities
```

**Good**: Focused, single-responsibility tools
```python
@server.tool()
def upload_files(files: list[str]) -> dict:
    """Upload files to system."""
    return upload(files)

@server.tool()
def parse_files(file_ids: list[str], format: str) -> dict:
    """Parse uploaded files in specified format."""
    return parse(file_ids, format)

@server.tool()
def filter_data(data_id: str, filters: dict) -> dict:
    """Apply filters to parsed data."""
    return apply_filters(data_id, filters)

@server.tool()
def transform_data(data_id: str, transformations: list[dict]) -> dict:
    """Apply transformations to data."""
    return apply_transforms(data_id, transformations)
```

**Workflow**: Claude chains multiple tools for complex operations:
1. User: "Upload these files and parse them"
2. Claude calls: `upload_files()` → `parse_files()`
3. User: "Filter by date and sort by value"
4. Claude calls: `filter_data()` → `transform_data()`

---

## Parameter Design

### Use Constrained Types

```python
from pydantic import Field
from typing import Literal

# Bad: String with no constraints
@server.tool()
def get_report(report_type: str) -> dict:
    """Get report."""
    # Claude might send invalid types
    pass

# Good: Literal enum (Claude knows available options)
@server.tool()
def get_report(
    report_type: Literal["sales", "inventory", "finances"]
) -> dict:
    """Get report of specified type."""
    pass

# Good: Constrained integers
@server.tool()
def list_items(
    limit: int = Field(default=20, ge=1, le=100),
    page: int = Field(default=1, ge=1)
) -> dict:
    """List items with pagination."""
    pass

# Good: Pattern validation
from pydantic import EmailStr, Field

@server.tool()
def send_email(
    to: EmailStr,
    subject: str = Field(..., min_length=1, max_length=200),
    body: str = Field(..., min_length=1)
) -> dict:
    """Send email with validation."""
    pass
```

### Pagination Parameters

```python
@server.tool()
def list_users(
    limit: int = Field(
        default=20,
        ge=1,
        le=100,
        description="Max results (1-100)"
    ),
    offset: int = Field(
        default=0,
        ge=0,
        description="Number of results to skip"
    )
) -> dict:
    """
    List users with pagination.

    Args:
        limit: Maximum results per page
        offset: Results offset for pagination
    """
    users = fetch_users(limit, offset)
    total = count_users()

    return {
        "users": users,
        "pagination": {
            "limit": limit,
            "offset": offset,
            "total": total,
            "has_more": offset + limit < total
        }
    }
```

---

## Input Validation

### Validation with Pydantic

```python
from pydantic import BaseModel, Field, validator

class CreateUserRequest(BaseModel):
    """User creation request with validation."""

    email: str = Field(
        ...,
        regex=r"^[^\s@]+@[^\s@]+\.[^\s@]+$",
        description="Valid email address"
    )
    name: str = Field(
        ...,
        min_length=1,
        max_length=100,
        description="User full name"
    )
    age: int = Field(
        ge=18,
        le=120,
        description="User age (18-120)"
    )
    role: Literal["user", "admin", "moderator"]

    @validator("email")
    def email_valid(cls, v):
        """Validate email format."""
        if not "@" in v:
            raise ValueError("Invalid email format")
        return v.lower()

    @validator("name")
    def name_normalized(cls, v):
        """Normalize name."""
        return v.strip().title()

@server.tool()
def create_user(request: CreateUserRequest) -> dict:
    """Create user with validated input."""
    user = save_user(request)
    return {
        "status": "success",
        "user_id": user.id,
        "email": user.email,
        "name": user.name
    }
```

### Pre-Validation Error Handling

```python
@server.tool()
def search(
    query: str = Field(..., min_length=1, max_length=200)
) -> dict:
    """
    Search with input validation.

    Args:
        query: Search query (1-200 characters)

    Raises:
        ValueError: If query is empty or too long
    """
    if not query.strip():
        raise ValueError("Query cannot be empty or whitespace only")

    if len(query) > 200:
        raise ValueError("Query cannot exceed 200 characters")

    results = execute_search(query)
    return {
        "query": query,
        "results": results,
        "count": len(results)
    }
```

---

## Return Value Design

### Consistent Response Structure

```python
@server.tool()
def operation() -> dict:
    """Operation with consistent response."""
    try:
        result = execute_operation()

        # Consistent success response
        return {
            "status": "success",
            "data": result,
            "timestamp": datetime.now().isoformat(),
            "request_id": generate_request_id()
        }

    except ValueError as e:
        # Consistent error response
        return {
            "status": "error",
            "error_code": "VALIDATION_ERROR",
            "message": str(e),
            "timestamp": datetime.now().isoformat()
        }

    except Exception as e:
        # Generic error response
        return {
            "status": "error",
            "error_code": "INTERNAL_ERROR",
            "message": "Operation failed",
            "timestamp": datetime.now().isoformat()
        }
```

### Metadata in Responses

```python
@server.tool()
def list_items(limit: int = 20, offset: int = 0) -> dict:
    """List with metadata."""
    items = fetch_items(limit, offset)
    total = count_items()

    return {
        "status": "success",
        "data": items,
        "metadata": {
            "limit": limit,
            "offset": offset,
            "total": total,
            "has_more": offset + limit < total,
            "returned": len(items)
        },
        "timestamp": datetime.now().isoformat()
    }
```

---

## Naming Conventions

### Verb-Noun Pattern

```python
# Search/Query operations
@server.tool()
def search_users(query: str) -> dict:
    """Search users by query."""
    pass

@server.tool()
def list_products(category: str) -> dict:
    """List products in category."""
    pass

@server.tool()
def get_user_details(user_id: str) -> dict:
    """Get detailed user information."""
    pass

# Create/Insert operations
@server.tool()
def create_user(name: str, email: str) -> dict:
    """Create new user."""
    pass

@server.tool()
def add_to_cart(product_id: str, quantity: int) -> dict:
    """Add product to cart."""
    pass

# Update/Modify operations
@server.tool()
def update_user(user_id: str, data: dict) -> dict:
    """Update user information."""
    pass

@server.tool()
def change_password(user_id: str, new_password: str) -> dict:
    """Change user password."""
    pass

# Delete/Remove operations
@server.tool()
def delete_user(user_id: str, reason: str = "") -> dict:
    """Delete user account."""
    pass

@server.tool()
def remove_from_cart(item_id: str) -> dict:
    """Remove item from cart."""
    pass
```

---

## Documentation

### Clear Docstrings

```python
@server.tool()
def search_documents(
    query: str = Field(..., min_length=1, max_length=200),
    category: Literal["blog", "docs", "help"] = "docs",
    limit: int = Field(default=10, ge=1, le=100)
) -> dict:
    """
    Search documents by query and category.

    This tool performs full-text search across the document database
    with optional category filtering and result pagination.

    Args:
        query: Search query string (1-200 chars).
               Can include logical operators: AND, OR, NOT
        category: Document category to filter by.
                  Options: 'blog', 'docs', 'help'
        limit: Maximum results to return (1-100).
               Higher limits may affect performance

    Returns:
        dict: Search results with structure:
            {
                'status': 'success',
                'query': str (the search query),
                'category': str (category filtered by),
                'results': list[dict] (matching documents),
                'total_found': int (total matches),
                'limit': int (results limit applied),
                'has_more': bool (more results available)
            }

    Raises:
        ValueError: If query is empty or invalid
        TimeoutError: If search takes too long

    Example:
        search_documents(
            query="Python async programming",
            category="docs",
            limit=20
        )
    """
    if not query.strip():
        raise ValueError("Query cannot be empty")

    results = execute_search(query, category, limit)

    return {
        "status": "success",
        "query": query,
        "category": category,
        "results": results,
        "total_found": count_matches(query, category),
        "limit": limit,
        "has_more": len(results) == limit
    }
```

---

## Error Handling

### Informative Error Messages

```python
@server.tool()
def transfer_funds(
    from_account: str,
    to_account: str,
    amount: float
) -> dict:
    """Transfer funds between accounts."""

    try:
        # Validate accounts exist
        from_acc = fetch_account(from_account)
        if not from_acc:
            raise ValueError(
                f"Source account not found: {from_account}. "
                f"Please check the account number and try again."
            )

        to_acc = fetch_account(to_account)
        if not to_acc:
            raise ValueError(
                f"Destination account not found: {to_account}. "
                f"Please check the account number and try again."
            )

        # Validate amount
        if amount <= 0:
            raise ValueError(
                f"Transfer amount must be positive. "
                f"Requested: {amount}"
            )

        # Check balance
        if from_acc.balance < amount:
            raise ValueError(
                f"Insufficient funds. "
                f"Available: ${from_acc.balance:.2f}, "
                f"Requested: ${amount:.2f}"
            )

        # Execute transfer
        transfer_id = execute_transfer(from_account, to_account, amount)

        return {
            "status": "success",
            "transfer_id": transfer_id,
            "from_account": from_account,
            "to_account": to_account,
            "amount": amount,
            "timestamp": datetime.now().isoformat()
        }

    except ValueError as e:
        raise ValueError(f"Transfer failed: {str(e)}")

    except Exception as e:
        raise RuntimeError(
            f"Unexpected error during transfer: {str(e)}. "
            f"Please contact support."
        )
```

---

## Performance Optimization

### Limit Result Sets

```python
@server.tool()
def search_logs(
    query: str,
    days: int = Field(default=7, ge=1, le=90),
    limit: int = Field(default=100, ge=1, le=1000)
) -> dict:
    """Search logs with reasonable limits."""
    # Enforce time window
    if days > 90:
        days = 90

    # Enforce result limit
    if limit > 1000:
        limit = 1000

    logs = search(query, days, limit)

    return {
        "logs": logs,
        "count": len(logs),
        "time_window_days": days,
        "limit_applied": limit,
        "has_more": len(logs) == limit
    }
```

### Async Support

```python
@server.tool()
async def process_file(file_id: str) -> dict:
    """Process file asynchronously."""
    # Don't block on long operations
    task_id = await start_processing(file_id)

    return {
        "status": "processing",
        "task_id": task_id,
        "file_id": file_id,
        "estimated_duration_seconds": 30
    }
```

---

## Tool Composition

### Tools that call other tools

```python
# Low-level tool
@server.tool()
def get_user(user_id: str) -> dict:
    """Get user by ID."""
    return fetch_user(user_id)

# Mid-level tool building on low-level
@server.tool()
def get_user_with_posts(user_id: str) -> dict:
    """Get user and their recent posts."""
    user = fetch_user(user_id)
    posts = fetch_user_posts(user_id, limit=10)

    return {
        "user": user,
        "posts": posts,
        "post_count": len(posts)
    }

# High-level orchestration via Claude
# Claude calls: get_user_with_posts() → process result → call next tool
```

---

**Last Updated**: 2025-11-27 | Production-grade tool design
