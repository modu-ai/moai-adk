---
name: moai-lang-python
version: 4.0.0
updated: '2025-11-19'
status: stable
stability: stable
description: Enterprise-grade Python expertise with production patterns for Python
allowed-tools:
- Read
- Bash
- WebSearch
- WebFetch
---



# Modern Python Development — Enterprise  

## Quick Summary

**Primary Focus**: Python 3.13 with FastAPI, Django, async patterns, and production deployment
**Best For**: REST APIs, microservices, async programming, ORM usage, testing
**Key Libraries**: FastAPI 0.115, Django 5.2 LTS, Pydantic v2, SQLAlchemy 2.0, pytest 8.3
**Auto-triggers**: FastAPI, Django, async, SQLAlchemy, Pydantic, pytest, asyncio

| Version | Release | Support |
|---------|---------|---------|
| Python 3.13.9 | 2025-10 | Oct 2029 |
| FastAPI 0.115 | 2025-11 | Active |
| Django 5.2 LTS | 2025-08 | Apr 2027 |
| SQLAlchemy 2.0 | 2024-01 | Active |

---

## Three-Level Learning Path

### Level 1: Fundamentals (Read examples.md)

Core Python 3.13 concepts with practical examples:
- **Python 3.13 Core**: JIT compiler, free-threaded mode, REPL improvements
- **FastAPI Basics**: API structure, Pydantic validation, dependency injection
- **Django Setup**: Project structure, async views, model inheritance
- **Async/Await**: Concurrent HTTP requests, context managers, timeouts
- **Examples**: See `examples.md` for full code samples

### Level 2: Advanced Patterns (See reference.md)

Production-ready enterprise patterns:
- **SQLAlchemy 2.0 ORM**: Async sessions, relationships, CRUD operations
- **Pydantic v2 Validation**: Field validation, custom validators, JSON schemas
- **pytest Testing**: Fixtures, parametrization, async test functions
- **Web Frameworks**: FastAPI middleware, Django authentication, error handling
- **Pattern Reference**: See `reference.md` for API details and best practices

### Level 3: Production Deployment (Consult security/performance skills)

Enterprise deployment and optimization:
- **Docker & Compose**: Container packaging with multi-stage builds
- **Performance**: Query optimization, Redis caching, N+1 prevention
- **Monitoring**: Prometheus metrics, structured logging, observability
- **Scaling**: Worker management, connection pooling, graceful shutdown
- **Details**: Skill("moai-essentials-perf"), Skill("moai-security-backend")

---

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

---

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

---

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

---

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

---

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

---

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

## Learn More

- **Examples**: See `examples.md` for FastAPI, Django, pytest, and async patterns
- **Reference**: See `reference.md` for API details, configuration, and troubleshooting
- **FastAPI Docs**: https://fastapi.tiangolo.com/
- **SQLAlchemy 2.0**: https://docs.sqlalchemy.org/en/20/orm/
- **Python 3.13**: https://docs.python.org/3/whatsnew/3.13.html
- **pytest Guide**: https://docs.pytest.org/

---

**Skills**: Skill("moai-essentials-debug"), Skill("moai-essentials-perf"), Skill("moai-security-backend")
**Auto-loads**: Python projects mentioning FastAPI, Django, async, SQLAlchemy, Pydantic, pytest

