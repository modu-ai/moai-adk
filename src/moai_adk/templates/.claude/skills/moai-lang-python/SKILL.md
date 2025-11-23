---
name: moai-lang-python
description: Enterprise-grade Python 3.13+ with FastAPI, Django, async patterns, Pydantic, and SQLAlchemy 2.0 for backend development, data science, and automation
version: 1.0.0
modularized: false
tags:
  - python
  - programming-language
  - enterprise
  - development
updated: 2025-11-24
status: active
---

## 📊 Skill Metadata

**version**: 1.0.0  
**modularized**: false  
**last_updated**: 2025-11-22  
**compliance_score**: 75%  
**auto_trigger_keywords**: lang, moai, django, fastapi, python, api, backend  


## Quick Reference

Enterprise Python 3.13+ with FastAPI, Django, async patterns, Pydantic v2, and SQLAlchemy 2.0 for production backend development, data science/ML, and system automation.

---

## What It Does

Python provides a comprehensive platform for backend development, data science, and automation. It excels in:

- **Web Development**: FastAPI 0.115+, Django 5.2 LTS, async I/O with asyncio
- **Data Science & ML**: NumPy, Pandas, scikit-learn, PyTorch for AI/ML workflows
- **System Automation**: Scripts, DevOps tools, CLI applications
- **Performance**: CPython 3.13 JIT compiler, GIL-free mode for parallelism
- **Type Safety**: Type hints with Pydantic v2 for runtime validation

Python combines readable syntax with powerful libraries, making it ideal for rapid development and production systems. The 3.13 release brings significant performance improvements through JIT compilation and experimental GIL-free threading.

---

## When to Use

**Use Python when**:
- Building data science/ML applications (NumPy, scikit-learn, PyTorch, TensorFlow)
- Need rapid development with readable, maintainable code
- Creating backend APIs with async support (FastAPI, Django async views)
- Automating DevOps/system tasks and scripting workflows
- Prototyping and iterating quickly on business logic

**Avoid Python when**:
- Real-time performance is critical (< 1ms latency requirements, use Rust/C++)
- Building embedded systems or firmware with strict memory constraints
- Need compile-time type safety guarantees (consider Rust/Go instead)

---

## Key Features

1. **Type Hints (3.5+)**: Optional static type checking with mypy, runtime validation with Pydantic v2
2. **Async/Await (3.5+)**: Native async I/O with asyncio, Trio, async frameworks (FastAPI, HTTPX)
3. **Pattern Matching (3.10)**: Match-case statements for elegant control flow and data validation
4. **Walrus Operator (3.8)**: Assignment expressions `:=` for cleaner, more concise code
5. **Decorators & Metaclasses**: Powerful metaprogramming capabilities for frameworks
6. **GIL-free Mode (3.13)**: Experimental multi-threading support for true parallelism
7. **JIT Compilation (3.13)**: Performance optimization for hot code paths (PEP 744)
8. **Dataclasses (3.7+)**: Simplified class definitions with automatic `__init__`, `__repr__`

---

## Works Well With

- `moai-domain-backend` — FastAPI/Django REST API and GraphQL backend development
  - Best for: Microservices, async APIs, ORM with SQLAlchemy 2.0
  
- `moai-domain-ml-ops` — Data science & ML workflows (NumPy, PyTorch, scikit-learn)
  - Best for: Model training, data preprocessing, inference pipelines
  
- `moai-domain-database` — PostgreSQL, MongoDB integration with SQLAlchemy and Motor
  - Best for: Data persistence, ORM patterns, async database operations
  
- `moai-domain-cloud` — AWS Lambda, Google Cloud Functions, Azure Functions
  - Best for: Serverless applications, cloud-native microservices
  
- `moai-lang-typescript` — Python backend + TypeScript frontend full-stack architecture
  - Best for: Modern web applications with FastAPI backend and React/Next.js frontend

---

## Core Concepts

### Everything is an Object
Python treats all entities (functions, classes, modules) as first-class objects. Functions can be passed as arguments, returned from other functions, and assigned to variables. This enables powerful functional programming patterns and metaprogramming.

### Duck Typing
Python uses dynamic typing with duck typing: "If it walks like a duck and quacks like a duck, it's a duck." Type checking happens at runtime, not compile time. This allows flexible and expressive code, but requires disciplined use of type hints for large codebases.

### The Zen of Python (PEP 20)
Python emphasizes readability and simplicity. Key principles:
- **Explicit is better than implicit**: Clear code over clever tricks
- **Simple is better than complex**: Favor straightforward solutions
- **Readability counts**: Code is read more often than written
- **There should be one obvious way to do it**: Consistency over flexibility

---

## Best Practices

### ✅ DO

1. **Use Type Hints**: Add type annotations for better IDE support, catch errors early, enable runtime validation with Pydantic
   - Reason: Catches type errors before runtime, improves code documentation
   
2. **Leverage Async/Await**: Use async libraries (asyncio, HTTPX, FastAPI) for I/O-bound operations
   - Reason: Higher concurrency, better resource utilization without threads
   
3. **Virtual Environments**: Always use venv/pyenv/poetry for dependency isolation
   - Reason: Avoids dependency conflicts, ensures reproducible builds
   
4. **List Comprehensions**: Use for concise iterations `[x*2 for x in items]`
   - Reason: More readable and 2-3x faster than equivalent loops
   
5. **Context Managers**: Use `with` statements for resource cleanup (files, DB connections)
   - Reason: Guarantees cleanup even if exceptions occur, prevents resource leaks

### ❌ DON'T

1. **Global Variables**: Avoid modifying global state across modules
   - Reason: Makes code hard to reason about, causes hidden bugs, breaks testability
   
2. **Bare Except Clauses**: Never catch all exceptions silently with `except:`
   - Reason: Hides errors, makes debugging impossible, catches system exits
   
3. **Mutable Default Arguments**: Don't use lists/dicts as function defaults
   - Reason: Shared state across function calls causes unexpected behavior
   
4. **GIL-blocking Operations**: Avoid CPU-bound work in async functions
   - Reason: Blocks entire event loop, kills concurrency benefits
   
5. **Star Imports**: Never use `from module import *` in production code
   - Reason: Pollutes namespace, makes code unclear, breaks static analysis

---

## Implementation Guide

## Technology Stack (November 2025 Stable)

### Runtime & Core
- **Python 3.13.9** (Latest, Oct 2025)
  - JIT compiler (PEP 744) for hot paths
  - Free-threaded mode (PEP 703) for parallelism
  - Enhanced REPL and error messages
- **asyncio** (stdlib) - Native async/await support
- **typing** (stdlib) - Type hints and runtime validation

### Web & API
- **FastAPI 0.115** (production-ready async framework)
  - OpenAPI/Swagger documentation
  - Dependency injection system
  - Async-first request handling
- **Django 5.2 LTS** (2025-08, support until Apr 2027)
  - Streaming responses and async views
  - MariaDB 10.9+ support
  - Improved model forms
- **Flask 3.1** (lightweight alternative)

### Data & Validation
- **Pydantic v2.9** - Type validation with JSON schema generation
- **SQLAlchemy 2.0** - Modern ORM with async support
- **Tortoise ORM 0.21** - Async-first alternative

### Testing & Quality
- **pytest 8.3** - Fixtures, parametrization, plugins
- **pytest-asyncio 0.24** - Async test support
- **mypy 1.8** - Static type checking
- **ruff 0.13** - Fast Python linter

### Deployment
- **Uvicorn 0.30** - ASGI server for async apps
- **Gunicorn 22** - WSGI server with workers
- **asyncpg 0.30** - Async PostgreSQL client
- **aiohttp 3.10** - Async HTTP client

## FastAPI Essential Patterns

### Project Structure
```
myapp/
├── main.py           # FastAPI app instance
├── models.py         # Pydantic models
├── database.py       # SQLAlchemy setup
├── schemas.py        # Request/response DTOs
├── api/
│   ├── __init__.py
│   └── routes.py     # Route handlers
├── services/
│   └── user_service.py   # Business logic
└── tests/
    ├── test_api.py
    └── test_services.py
```

### Minimal API
```python
from fastapi import FastAPI
from pydantic import BaseModel

app = FastAPI(title="API", version="1.0.0")

class Item(BaseModel):
    name: str
    price: float

@app.get("/items/{item_id}")
async def get_item(item_id: int) -> Item:
    return Item(name="Sample", price=9.99)

@app.post("/items/", status_code=201)
async def create_item(item: Item) -> Item:
    return item
```

### Dependency Injection
```python
from fastapi import Depends
from sqlalchemy.ext.asyncio import AsyncSession

async def get_db() -> AsyncSession:
    async with async_session() as session:
        yield session

async def get_current_user(token: str = Depends(oauth2_scheme)):
    return decode_token(token)

@app.get("/profile")
async def profile(
    user = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    return user
```

## SQLAlchemy 2.0 Async ORM

### Setup
```python
from sqlalchemy.ext.asyncio import (
    create_async_engine,
    AsyncSession,
    async_sessionmaker
)
from sqlalchemy.orm import declarative_base

DATABASE_URL = "postgresql+asyncpg://user:pass@localhost/db"
engine = create_async_engine(DATABASE_URL, echo=False)
async_session = async_sessionmaker(engine, class_=AsyncSession)
Base = declarative_base()
```

### Models
```python
from sqlalchemy import Column, Integer, String, ForeignKey
from sqlalchemy.orm import relationship

class User(Base):
    __tablename__ = "users"
    id = Column(Integer, primary_key=True)
    username = Column(String(50), unique=True, nullable=False)
    posts = relationship("Post", back_populates="author")

class Post(Base):
    __tablename__ = "posts"
    id = Column(Integer, primary_key=True)
    title = Column(String(200), nullable=False)
    author_id = Column(Integer, ForeignKey("users.id"))
    author = relationship("User", back_populates="posts")
```

### Async CRUD
```python
from sqlalchemy import select

async def create_user(username: str) -> User:
    async with async_session() as session:
        user = User(username=username)
        session.add(user)
        await session.commit()
        return user

async def get_user(user_id: int) -> User:
    async with async_session() as session:
        result = await session.execute(select(User).where(User.id == user_id))
        return result.scalars().first()
```

## Async/Await Fundamentals

### Concurrent Operations
```python
import asyncio
import aiohttp

async def fetch_url(url: str) -> str:
    async with aiohttp.ClientSession() as session:
        async with session.get(url) as response:
            return await response.text()

async def fetch_multiple(urls):
    tasks = [fetch_url(url) for url in urls]
    return await asyncio.gather(*tasks)

# Run
results = asyncio.run(fetch_multiple(urls))
```

### Timeouts & Cancellation
```python
async def with_timeout():
    try:
        result = await asyncio.wait_for(long_task(), timeout=10.0)
    except asyncio.TimeoutError:
        print("Timed out")

async def cancellable_task():
    try:
        await asyncio.sleep(100)
    except asyncio.CancelledError:
        print("Cancelled")
        raise
```

## Pydantic v2 Validation

### Models
```python
from pydantic import BaseModel, Field, field_validator, EmailStr

class UserSchema(BaseModel):
    id: int = Field(..., gt=0)
    username: str = Field(..., min_length=3, max_length=50)
    email: EmailStr
    bio: str = Field(None, max_length=500)

    @field_validator('username')
    @classmethod
    def validate_username(cls, v):
        if not v.isalnum():
            raise ValueError('Alphanumeric only')
        return v.lower()

# Usage
user = UserSchema(id=1, username="john", email="john@example.com")
schema = UserSchema.model_json_schema()  # JSON Schema
```

## Testing with pytest

### Basic Tests
```python
import pytest
from fastapi.testclient import TestClient

@pytest.fixture
def client():
    return TestClient(app)

def test_create_item(client):
    response = client.post("/items/", json={"name": "Item", "price": 10})
    assert response.status_code == 201
```

### Async Tests
```python
@pytest.mark.asyncio
async def test_async_fetch():
    result = await fetch_url("http://example.com")
    assert result is not None
```

---

## Advanced Patterns

## Production Best Practices

1. **Always use type hints** for IDE support and validation
2. **Prefer async/await** for I/O-bound operations (not threads)
3. **Use Pydantic v2** for data validation (not manual checks)
4. **Test async code** with pytest-asyncio markers
5. **Implement error handling** with custom exceptions
6. **Use SQLAlchemy 2.0** for ORM (not raw SQL)
7. **Cache frequently accessed data** with Redis
8. **Monitor metrics** in production (Prometheus, Sentry)
9. **Use Uvicorn** for FastAPI (not development server)
10. **Enable JIT compiler** for performance-critical code in Python 3.13

---

## Context7 Integration

### Related Libraries & Tools
- [FastAPI](/tiangolo/fastapi): Modern async web framework for building APIs with Python
- [SQLAlchemy](/sqlalchemy/sqlalchemy): The Python SQL toolkit and Object Relational Mapper
- [Pydantic](/pydantic/pydantic): Data validation using Python type annotations
- [pytest](/pytest-dev/pytest): The pytest framework for testing Python applications
- [Django](/django/django): High-level Python web framework for rapid development

### Official Documentation
- [Python 3.13](https://docs.python.org/3.13/)
- [FastAPI Documentation](https://fastapi.tiangolo.com/)
- [SQLAlchemy 2.0](https://docs.sqlalchemy.org/en/20/)
- [Pydantic v2](https://docs.pydantic.dev/)

### Version-Specific Guides
Latest stable version: Python 3.13.9, FastAPI 0.115.x, Django 5.2 LTS
- [Python 3.13 What's New](https://docs.python.org/3/whatsnew/3.13.html)
- [FastAPI 0.115 Release Notes](https://github.com/tiangolo/fastapi/releases)
- [Django 5.2 Release Notes](https://docs.djangoproject.com/en/5.2/releases/5.2/)
- [SQLAlchemy 2.0 Migration Guide](https://docs.sqlalchemy.org/en/20/changelog/migration_20.html)

---

**Last Updated**: 2025-11-22  
**Status**: Production Ready  
**Version**: 4.0.0
