---
name: moai-lang-fastapi-patterns
type: language
description: FastAPI async patterns, dependency injection, and OpenAPI documentation
tier: language
---

# FastAPI Patterns

## Quick Start (30 seconds)
FastAPI is a modern, fast web framework for building APIs with Python 3.7+ features. It automatically generates OpenAPI documentation, validates requests with Pydantic, and supports async/await for high performance.

## Core Patterns

### Pattern 1: Async Route Handlers
Define non-blocking endpoints using async functions for handling concurrent requests.

```python
from fastapi import FastAPI
from pydantic import BaseModel
import asyncio

app = FastAPI()

class Item(BaseModel):
    name: str
    price: float
    description: str = None

@app.get("/items/{item_id}")
async def get_item(item_id: int):
    """Get item by ID - supports concurrent requests"""
    # Simulate async I/O operation
    await asyncio.sleep(0.1)
    return {"item_id": item_id, "name": "Sample Item"}

@app.post("/items/")
async def create_item(item: Item):
    """Create new item with automatic validation"""
    return {"created": True, "item": item}

@app.put("/items/{item_id}")
async def update_item(item_id: int, item: Item):
    """Update existing item"""
    return {"updated": True, "item_id": item_id, "item": item}
```

**When to use**: Building REST APIs that need to handle multiple concurrent requests efficiently. Async handlers allow thousands of concurrent connections without blocking.

**Best practices**:
- Use `async def` for route handlers
- Use `await` when calling async operations (database, HTTP, file I/O)
- Keep handlers focused and delegate business logic to services
- Use dependency injection for shared logic

### Pattern 2: Dependency Injection
Reuse common logic (authentication, database sessions, validation) across multiple endpoints.

```python
from fastapi import FastAPI, Depends, HTTPException, Header
from sqlalchemy.orm import Session
from sqlalchemy import create_engine
from sqlalchemy.orm import sessionmaker

DATABASE_URL = "sqlite:///./test.db"
engine = create_engine(DATABASE_URL, connect_args={"check_same_thread": False})
SessionLocal = sessionmaker(autocommit=False, autoflush=False, bind=engine)

app = FastAPI()

# Dependency: Database Session
def get_db():
    db = SessionLocal()
    try:
        yield db
    finally:
        db.close()

# Dependency: Current User Authentication
async def get_current_user(token: str = Header(...)):
    if not token.startswith("Bearer "):
        raise HTTPException(status_code=401, detail="Invalid token")
    return {"user_id": 1, "username": "john"}

# Using dependencies
@app.get("/items/")
async def list_items(db: Session = Depends(get_db), current_user: dict = Depends(get_current_user)):
    """List items for authenticated user"""
    # Access database and user info
    return {"items": [], "user": current_user["username"]}

@app.post("/items/")
async def create_item(item: Item, db: Session = Depends(get_db)):
    """Create item with automatic database session"""
    # db is automatically provided
    return {"created": True, "item": item}
```

**When to use**: When you need to:
- Share database connections across endpoints
- Enforce authentication/authorization
- Validate common input parameters
- Log request information

**Benefits**:
- Reduces code duplication
- Centralizes validation logic
- Improves testability (easy to mock dependencies)
- Enforces consistent authentication/authorization

### Pattern 3: Pydantic Models
Define request/response schemas with automatic validation and documentation.

```python
from pydantic import BaseModel, Field, EmailStr, field_validator
from typing import Optional

# Request model with validation
class UserCreate(BaseModel):
    username: str = Field(..., min_length=3, max_length=50)
    email: EmailStr
    age: int = Field(..., ge=18, le=120)
    bio: Optional[str] = Field(None, max_length=500)

    @field_validator('username')
    @classmethod
    def username_alphanumeric(cls, v):
        if not v.isalnum():
            raise ValueError('Username must be alphanumeric')
        return v

# Response model with selection
class UserResponse(BaseModel):
    id: int
    username: str
    email: str

    class Config:
        from_attributes = True  # For SQLAlchemy integration

# Usage in endpoint
@app.post("/users/", response_model=UserResponse)
async def create_user(user: UserCreate, db: Session = Depends(get_db)):
    """Create user with automatic validation and response formatting"""
    db_user = User(**user.dict())
    db.add(db_user)
    db.commit()
    return db_user
```

**Validation features**:
- Type checking (str, int, float, bool, etc.)
- String constraints (min_length, max_length, regex, EmailStr)
- Numeric constraints (ge, le, gt, lt)
- Custom validators with `@field_validator`
- Optional fields with default values
- Automatic documentation in OpenAPI

## Progressive Disclosure

### Middleware for Cross-Cutting Concerns

```python
from fastapi import FastAPI
from starlette.middleware.base import BaseHTTPMiddleware
import time
import logging

app = FastAPI()
logger = logging.getLogger(__name__)

class LoggingMiddleware(BaseHTTPMiddleware):
    async def dispatch(self, request, call_next):
        start_time = time.time()
        response = await call_next(request)
        process_time = time.time() - start_time
        logger.info(f"{request.method} {request.url.path} - {process_time:.3f}s")
        response.headers["X-Process-Time"] = str(process_time)
        return response

app.add_middleware(LoggingMiddleware)
```

### Background Tasks

```python
from fastapi import BackgroundTasks

@app.post("/send-notification/")
async def send_notification(email: str, background_tasks: BackgroundTasks):
    """Endpoint that returns immediately, processes in background"""
    background_tasks.add_task(send_email, email, "Hello!")
    return {"message": "Notification sent"}

def send_email(email: str, message: str):
    # This runs in background after response is sent
    import smtplib
    # Email sending logic
    pass
```

### WebSocket Support

```python
from fastapi import WebSocket

@app.websocket("/ws")
async def websocket_endpoint(websocket: WebSocket):
    await websocket.accept()
    try:
        while True:
            data = await websocket.receive_text()
            await websocket.send_text(f"Echo: {data}")
    except Exception as e:
        await websocket.close(code=1000)
```

## Works Well With
- **moai-lang-python** - Core Python language features and async/await patterns
- **moai-domain-backend** - Backend architecture and system design
- **moai-domain-database** - Database integration with SQLAlchemy

## Common Pitfalls

1. **Blocking Operations in Async Handlers**
   - ❌ Wrong: Using `time.sleep()` in async function (blocks entire server)
   - ✅ Right: Using `await asyncio.sleep()` for non-blocking delays

2. **Missing `await` Keywords**
   - ❌ Wrong: `result = database.query(Item)` (doesn't execute)
   - ✅ Right: `result = await database.query(Item)` (executes asynchronously)

3. **Not Using Dependencies**
   - ❌ Wrong: Repeating database session setup in every endpoint
   - ✅ Right: Use `Depends(get_db)` to inject shared dependencies

4. **Overly Complex Pydantic Models**
   - ❌ Wrong: Single model for multiple endpoints with many optional fields
   - ✅ Right: Separate models for different use cases (Create, Update, Response)

## References
- [FastAPI Official Documentation](https://fastapi.tiangolo.com/)
- [Pydantic Documentation](https://docs.pydantic.dev/)
- [Python asyncio Documentation](https://docs.python.org/3/library/asyncio.html)
- [SQLAlchemy + FastAPI](https://fastapi.tiangolo.com/advanced/sql-databases/)
- [OpenAPI Specification](https://swagger.io/specification/)
