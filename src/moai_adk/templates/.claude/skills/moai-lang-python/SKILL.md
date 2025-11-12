---
name: "moai-lang-python"
version: "4.0.0"
description: Enterprise-grade Python expertise with production patterns for Python 3.13.9, FastAPI 0.115.x, Django 5.2 LTS, Pydantic v2, SQLAlchemy 2.0; activates for API development, ORM usage, async patterns, testing frameworks, and production deployment strategies.
allowed-tools: 
  - Read
  - Bash
  - WebSearch
  - WebFetch
status: stable
---

# Modern Python Development — Enterprise v4.0

## Skill Metadata

| Field | Value |
| ----- | ----- |
| **Version** | **4.0.0 Enterprise** |
| **Created** | 2025-11-12 |
| **Updated** | 2025-11-12 |
| **Lines** | ~950 lines |
| **Size** | ~30KB |
| **Tier** | **3 (Professional)** |
| **Allowed tools** | Read, Bash, WebSearch, WebFetch |
| **Auto-load** | Python API development, async patterns, ORM usage |
| **Trigger cues** | Python, FastAPI, Django, async, API, ORM, SQLAlchemy, Pydantic, pytest |

## Technology Stack (November 2025 Stable)

### Core Language
- **Python 3.13.9** (Latest maintenance release, 2025-10-14)
  - JIT compiler support (PEP 744)
  - Free-threaded mode (PEP 703)
  - Interactive interpreter improvements
  - Maintenance updates until October 2029

### Web Frameworks
- **FastAPI 0.115.x** (Production-ready, 2025-11)
  - Async-first architecture
  - OpenAPI/Swagger automatic documentation
  - Dependency injection
  - Validation with Pydantic v2

- **Django 5.2 LTS** (Long-term support, 2025-08)
  - Streaming responses
  - MariaDB 10.9+ support
  - Improved model form inheritance
  - Support until April 2027

- **Flask 3.1.x** (Lightweight alternative)
  - Blueprint system
  - Request context management
  - Extension ecosystem

### Async Runtime
- **asyncio (stdlib)** - Native async support
- **aiohttp 3.10.x** - Async HTTP client
- **asyncpg 0.30.x** - Async PostgreSQL driver

### Data & Validation
- **Pydantic v2.9.x** (Model validation framework)
  - Type hints support
  - JSON schema generation
  - Custom validators
  - Performance improvements over v1

- **SQLAlchemy 2.0.x** (ORM excellence)
  - Declarative ORM patterns
  - Query expressions
  - Relationship management
  - Connection pooling

- **Tortoise ORM 0.21.x** (Async-first alternative)
  - AsyncIO compatible
  - Migration system
  - Relationship handling

### Testing & Quality
- **pytest 8.3.x** (Modern testing framework)
  - Fixtures system
  - Parametrization
  - Plugin ecosystem
  - Coverage integration

- **pytest-asyncio 0.24.x** (Async testing support)
  - Async test functions
  - Fixture support
  - Event loop management

### Deployment
- **Uvicorn 0.30.x** (ASGI server)
  - Production performance
  - Auto-reload for development
  - Worker management
  - SSL/TLS support

- **Gunicorn 22.x** (WSGI server)
  - Multiple worker models
  - Graceful reloading
  - Reverse proxy compatible

## Level 1: Fundamentals (High Freedom)

### 1. Python 3.13 Core Features

Python 3.13 brings significant improvements for enterprise applications:

**JIT Compiler (PEP 744)**:
```python
# Python 3.13+ automatically optimizes hot paths
def calculate(n: int) -> int:
    """JIT-compiled for performance"""
    total = 0
    for i in range(n):
        total += i * 2
    return total

# 2-3x faster than Python 3.12 in performance tests
result = calculate(1_000_000)
```

**Free-Threaded Mode (PEP 703)**:
```python
# Compile with ./configure --disable-gil
# Eliminates GIL for true parallelism

from threading import Thread

def cpu_intensive_task(n):
    return sum(i**2 for i in range(n))

# Multiple threads can run Python code simultaneously
threads = [Thread(target=cpu_intensive_task, args=(1000000,)) 
           for _ in range(4)]
for t in threads:
    t.start()
for t in threads:
    t.join()
```

**Interactive Shell Improvements**:
```python
# Enhanced REPL with better multi-line support
# Type hints in completions
# Improved error messages
```

### 2. FastAPI Modern Patterns (0.115.x)

FastAPI is production-tested framework for REST APIs:

**Basic API Structure**:
```python
from fastapi import FastAPI, HTTPException
from pydantic import BaseModel, Field
from typing import List

app = FastAPI(
    title="Enterprise API",
    version="1.0.0",
    description="Production-grade API with Pydantic v2"
)

class Item(BaseModel):
    """Pydantic v2 model with validation"""
    id: int = Field(..., gt=0, description="Positive integer ID")
    name: str = Field(..., min_length=1, max_length=100)
    price: float = Field(..., gt=0, description="Item price")
    tags: List[str] = Field(default_factory=list)

@app.get("/items/{item_id}", response_model=Item)
async def get_item(item_id: int) -> Item:
    """Retrieve item by ID with automatic OpenAPI documentation"""
    if item_id < 1:
        raise HTTPException(status_code=400, detail="Invalid ID")
    return Item(id=item_id, name="Sample", price=9.99, tags=["featured"])

@app.post("/items/", response_model=Item, status_code=201)
async def create_item(item: Item) -> Item:
    """Create new item with validation"""
    # Pydantic validates input automatically
    return item
```

**Dependency Injection**:
```python
from fastapi import Depends
from sqlalchemy.ext.asyncio import AsyncSession

async def get_db() -> AsyncSession:
    """Database dependency"""
    async with async_session() as session:
        yield session

async def get_current_user(token: str = Depends(oauth2_scheme)):
    """Authentication dependency"""
    return decode_token(token)

@app.get("/profile")
async def get_profile(
    user = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """Endpoint with multiple dependencies"""
    return await db.execute(...)
```

### 3. Django 5.2 LTS (2025 Enterprise)

Django 5.2 is long-term support version for production:

**Project Setup**:
```bash
pip install django==5.2.8 djangorestframework==3.15.x
django-admin startproject myproject
django-admin startapp myapp
```

**Models with SQLAlchemy-style Features**:
```python
from django.db import models
from django.contrib.auth.models import AbstractUser

class CustomUser(AbstractUser):
    """Extended user model"""
    bio = models.TextField(blank=True)
    avatar = models.ImageField(upload_to='avatars/', null=True)
    created_at = models.DateTimeField(auto_now_add=True)

class Article(models.Model):
    """Article model with relationships"""
    title = models.CharField(max_length=200)
    author = models.ForeignKey(CustomUser, on_delete=models.CASCADE)
    content = models.TextField()
    published = models.DateTimeField(auto_now_add=True)
    updated = models.DateTimeField(auto_now=True)
    
    class Meta:
        ordering = ['-published']
        indexes = [
            models.Index(fields=['author', '-published']),
        ]
```

**Async Views (Django 5.2)**:
```python
from django.http import JsonResponse
from asgiref.sync import sync_to_async

async def async_article_list(request):
    """Async view with database queries"""
    articles = await sync_to_async(Article.objects.all)()
    return JsonResponse({'articles': list(articles.values())})
```

### 4. Async/Await Patterns

Enterprise async programming in Python 3.13:

**Basic Async Function**:
```python
import asyncio
import aiohttp
from typing import List

async def fetch_url(session: aiohttp.ClientSession, url: str) -> str:
    """Fetch single URL asynchronously"""
    async with session.get(url) as response:
        return await response.text()

async def fetch_multiple(urls: List[str]) -> List[str]:
    """Fetch multiple URLs concurrently"""
    async with aiohttp.ClientSession() as session:
        tasks = [fetch_url(session, url) for url in urls]
        return await asyncio.gather(*tasks)

# Usage
urls = ["https://api.example.com/1", "https://api.example.com/2"]
results = asyncio.run(fetch_multiple(urls))
```

**Async Context Managers**:
```python
from contextlib import asynccontextmanager

@asynccontextmanager
async def managed_resource():
    """Async context manager for resource management"""
    resource = await allocate_resource()
    try:
        yield resource
    finally:
        await resource.cleanup()

async def use_resource():
    """Use async context manager"""
    async with managed_resource() as res:
        await res.do_something()
```

**Task Cancellation & Timeouts**:
```python
async def long_running_task():
    """Task that can be cancelled"""
    try:
        for i in range(100):
            await asyncio.sleep(1)
            print(f"Progress: {i}%")
    except asyncio.CancelledError:
        print("Task was cancelled")
        raise

async def with_timeout():
    """Set timeout for async operation"""
    try:
        result = await asyncio.wait_for(
            long_running_task(),
            timeout=10.0
        )
    except asyncio.TimeoutError:
        print("Operation timed out")
```

## Level 2: Advanced Patterns (Medium Freedom)

### 1. ORM with SQLAlchemy 2.0

SQLAlchemy 2.0 provides modern async ORM patterns:

**Async Session Setup**:
```python
from sqlalchemy.ext.asyncio import (
    create_async_engine, 
    AsyncSession,
    async_sessionmaker
)
from sqlalchemy.orm import declarative_base

DATABASE_URL = "postgresql+asyncpg://user:password@localhost/dbname"

engine = create_async_engine(
    DATABASE_URL,
    echo=False,
    pool_size=20,
    max_overflow=0
)

async_session = async_sessionmaker(
    engine,
    class_=AsyncSession,
    expire_on_commit=False
)

Base = declarative_base()
```

**Model Definition with Relationships**:
```python
from sqlalchemy import Column, Integer, String, ForeignKey, DateTime, Text
from sqlalchemy.orm import relationship
from datetime import datetime

class User(Base):
    """User model with relationships"""
    __tablename__ = "users"
    
    id = Column(Integer, primary_key=True)
    username = Column(String(50), unique=True, nullable=False)
    email = Column(String(120), unique=True, nullable=False)
    created_at = Column(DateTime, default=datetime.utcnow)
    
    posts = relationship("Post", back_populates="author", cascade="all, delete-orphan")

class Post(Base):
    """Post model with foreign key"""
    __tablename__ = "posts"
    
    id = Column(Integer, primary_key=True)
    title = Column(String(200), nullable=False)
    content = Column(Text)
    author_id = Column(Integer, ForeignKey("users.id"), nullable=False)
    created_at = Column(DateTime, default=datetime.utcnow)
    
    author = relationship("User", back_populates="posts")
```

**Async CRUD Operations**:
```python
from sqlalchemy import select

async def create_user(username: str, email: str) -> User:
    """Create user asynchronously"""
    async with async_session() as session:
        user = User(username=username, email=email)
        session.add(user)
        await session.commit()
        return user

async def get_user_by_id(user_id: int) -> User:
    """Get user by ID"""
    async with async_session() as session:
        stmt = select(User).where(User.id == user_id)
        result = await session.execute(stmt)
        return result.scalars().first()

async def get_user_posts(user_id: int) -> list:
    """Get all posts for user"""
    async with async_session() as session:
        stmt = select(Post).where(Post.author_id == user_id)
        result = await session.execute(stmt)
        return result.scalars().all()

async def update_user(user_id: int, **kwargs) -> User:
    """Update user fields"""
    async with async_session() as session:
        stmt = select(User).where(User.id == user_id)
        result = await session.execute(stmt)
        user = result.scalars().first()
        
        for key, value in kwargs.items():
            if hasattr(user, key):
                setattr(user, key, value)
        
        await session.commit()
        return user

async def delete_user(user_id: int):
    """Delete user"""
    async with async_session() as session:
        stmt = select(User).where(User.id == user_id)
        result = await session.execute(stmt)
        user = result.scalars().first()
        await session.delete(user)
        await session.commit()
```

### 2. Pydantic v2 Validation

Pydantic v2 provides type-safe data validation:

**Model Validation**:
```python
from pydantic import BaseModel, Field, field_validator, EmailStr
from typing import Optional
from datetime import datetime

class UserSchema(BaseModel):
    """Pydantic v2 model with validation"""
    id: int = Field(..., gt=0)
    username: str = Field(..., min_length=3, max_length=50)
    email: EmailStr  # Built-in email validation
    age: int = Field(..., ge=0, le=150)
    bio: Optional[str] = Field(None, max_length=500)
    created_at: datetime = Field(default_factory=datetime.utcnow)
    
    @field_validator('username')
    @classmethod
    def validate_username(cls, v):
        """Custom validation for username"""
        if not v.isalnum():
            raise ValueError('Username must be alphanumeric')
        return v.lower()

# Usage
user_data = {
    "id": 1,
    "username": "john_doe",
    "email": "john@example.com",
    "age": 30,
    "bio": "Software developer"
}

user = UserSchema(**user_data)  # Validates automatically
print(user.model_dump())  # Pydantic v2 method
```

**JSON Schema Generation**:
```python
schema = UserSchema.model_json_schema()
# Automatically generates JSON Schema for API documentation
```

### 3. Testing with pytest

Comprehensive testing framework for Python:

**Unit Tests**:
```python
import pytest
from fastapi.testclient import TestClient

@pytest.fixture
def client():
    """Fixture for test client"""
    return TestClient(app)

def test_create_item(client):
    """Test item creation"""
    response = client.post("/items/", json={
        "id": 1,
        "name": "Test Item",
        "price": 10.0
    })
    assert response.status_code == 201
    assert response.json()["name"] == "Test Item"

def test_get_item(client):
    """Test item retrieval"""
    response = client.get("/items/1")
    assert response.status_code == 200
```

**Async Test Functions**:
```python
import pytest
import asyncio

@pytest.mark.asyncio
async def test_async_fetch():
    """Test async function"""
    results = await fetch_multiple(["http://example.com"])
    assert len(results) == 1

@pytest.fixture
async def async_client():
    """Async fixture"""
    async with aiohttp.ClientSession() as session:
        yield session
```

## Level 3: Production Deployment (Low Freedom, Expert Only)

### 1. Production Deployment Architecture

**Docker Deployment**:
```dockerfile
FROM python:3.13-slim

WORKDIR /app

# Copy and install dependencies
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Copy application
COPY app/ .

# Run with Uvicorn
CMD ["uvicorn", "main:app", "--host", "0.0.0.0", "--port", "8000"]
```

**Docker Compose for Full Stack**:
```yaml
version: '3.8'
services:
  app:
    build: .
    ports:
      - "8000:8000"
    environment:
      - DATABASE_URL=postgresql+asyncpg://user:password@db:5432/mydb
    depends_on:
      - db
  
  db:
    image: postgres:16
    environment:
      - POSTGRES_USER=user
      - POSTGRES_PASSWORD=password
      - POSTGRES_DB=mydb
    volumes:
      - postgres_data:/var/lib/postgresql/data
  
  redis:
    image: redis:7
    ports:
      - "6379:6379"

volumes:
  postgres_data:
```

### 2. Performance Optimization

**Database Query Optimization**:
```python
# Use SQLAlchemy select expressions for efficiency
from sqlalchemy import select
from sqlalchemy.orm import joinedload

async def get_users_with_posts():
    """Eager loading to prevent N+1 queries"""
    async with async_session() as session:
        stmt = select(User).options(joinedload(User.posts))
        result = await session.execute(stmt)
        return result.unique().scalars().all()
```

**Caching with Redis**:
```python
import redis.asyncio as redis

redis_client = redis.from_url("redis://localhost:6379")

async def get_cached_user(user_id: int):
    """Get user with Redis caching"""
    # Check cache first
    cached = await redis_client.get(f"user:{user_id}")
    if cached:
        return json.loads(cached)
    
    # Fetch from DB if not cached
    user = await get_user_by_id(user_id)
    
    # Cache for 1 hour
    await redis_client.setex(
        f"user:{user_id}",
        3600,
        json.dumps(user.model_dump())
    )
    return user
```

### 3. Monitoring & Observability

**Application Metrics**:
```python
from prometheus_client import Counter, Histogram, start_http_server

# Metrics
request_count = Counter(
    'http_requests_total',
    'Total HTTP requests',
    ['method', 'endpoint', 'status']
)

request_duration = Histogram(
    'http_request_duration_seconds',
    'HTTP request duration in seconds',
    ['method', 'endpoint']
)

# Start metrics server
start_http_server(8001)

@app.middleware("http")
async def add_metrics(request, call_next):
    """Middleware to record metrics"""
    start_time = time.time()
    response = await call_next(request)
    
    duration = time.time() - start_time
    request_count.labels(
        method=request.method,
        endpoint=request.url.path,
        status=response.status_code
    ).inc()
    
    request_duration.labels(
        method=request.method,
        endpoint=request.url.path
    ).observe(duration)
    
    return response
```

## Auto-Load Triggers

This Skill automatically activates when you:
- Work with Python projects and mention FastAPI, Django, or async patterns
- Need to understand ORM patterns with SQLAlchemy 2.0
- Implement REST APIs or microservices
- Work with Pydantic data validation
- Debug async/await issues in Python 3.13+
- Set up production deployments with pytest

## Best Practices Summary

1. **Always use Pydantic v2** for data validation (not manual checks)
2. **Prefer async/await** over threading for I/O-bound tasks
3. **Use type hints** extensively for IDE support and runtime validation
4. **Test async code** with pytest-asyncio markers
5. **Implement proper error handling** with custom exceptions
6. **Use SQLAlchemy 2.0** for ORM patterns, not raw SQL
7. **Cache frequently accessed data** with Redis
8. **Monitor metrics** in production environments
9. **Use Uvicorn** for FastAPI (not Flask development server)
10. **Enable JIT compiler** for performance-critical Python 3.13 code

## See Also

- **Pydantic v2 Documentation**: https://docs.pydantic.dev/latest/
- **FastAPI Official Docs**: https://fastapi.tiangolo.com/
- **SQLAlchemy 2.0 ORM Guide**: https://docs.sqlalchemy.org/en/20/orm/
- **Python 3.13 What's New**: https://docs.python.org/3/whatsnew/3.13.html
- **pytest Documentation**: https://docs.pytest.org/
- **Python AsyncIO Guide**: https://docs.python.org/3/library/asyncio.html

