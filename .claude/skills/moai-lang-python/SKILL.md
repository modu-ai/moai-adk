---
name: moai-lang-python
description: Python 3.13+ development specialist covering FastAPI, Django, async patterns, data science, testing with pytest, and modern Python features. Use when developing Python APIs, web applications, data pipelines, or writing tests.
version: 1.0.0
updated: 2025-12-07
status: active
allowed-tools: Read, Grep, Glob, Bash, mcp__context7__resolve-library-id, mcp__context7__get-library-docs
---

## Quick Reference (30 seconds)

Python 3.13+ Development Specialist - FastAPI, Django, async patterns, pytest, and modern Python features.

Auto-Triggers: `.py` files, `pyproject.toml`, `requirements.txt`, `pytest.ini`, FastAPI/Django discussions

Core Capabilities:
- Python 3.13 Features: JIT compiler (PEP 744), GIL-free mode (PEP 703), pattern matching
- Web Frameworks: FastAPI 0.115+, Django 5.2 LTS
- Data Validation: Pydantic v2.9 with model_validate patterns
- ORM: SQLAlchemy 2.0 async patterns
- Testing: pytest with fixtures, async testing, parametrize
- Package Management: poetry, uv, pip with pyproject.toml
- Type Hints: Protocol, TypeVar, ParamSpec, modern typing patterns
- Async: asyncio, async generators, task groups
- Data Science: numpy, pandas, polars basics

### Quick Patterns

FastAPI Endpoint:
```python
from fastapi import FastAPI, Depends
from pydantic import BaseModel

app = FastAPI()

class UserCreate(BaseModel):
    name: str
    email: str

@app.post("/users/")
async def create_user(user: UserCreate) -> User:
    return await UserService.create(user)
```

Pydantic v2.9 Validation:
```python
from pydantic import BaseModel, ConfigDict

class User(BaseModel):
    model_config = ConfigDict(from_attributes=True, str_strip_whitespace=True)

    id: int
    name: str
    email: str

user = User.model_validate(orm_obj)  # from ORM object
user = User.model_validate_json(json_data)  # from JSON
```

pytest Async Test:
```python
import pytest

@pytest.mark.asyncio
async def test_create_user(async_client):
    response = await async_client.post("/users/", json={"name": "Test"})
    assert response.status_code == 201
```

---

## Implementation Guide (5 minutes)

### Python 3.13 New Features

JIT Compiler (PEP 744):
- Experimental feature, disabled by default
- Enable: `PYTHON_JIT=1` environment variable
- Build option: `--enable-experimental-jit`
- Provides performance improvements for CPU-bound code
- Copy-and-patch JIT that translates specialized bytecode to machine code

GIL-Free Mode (PEP 703):
- Experimental free-threaded build (python3.13t)
- Allows true parallel thread execution
- Available in official Windows/macOS installers
- Best for: CPU-intensive multi-threaded applications
- Not recommended for production yet

Pattern Matching (match/case):
```python
def process_response(response: dict) -> str:
    match response:
        case {"status": "ok", "data": data}:
            return f"Success: {data}"
        case {"status": "error", "message": msg}:
            return f"Error: {msg}"
        case {"status": status} if status in ("pending", "processing"):
            return "In progress..."
        case _:
            return "Unknown response"
```

### FastAPI 0.115+ Patterns

Async Dependency Injection:
```python
from fastapi import FastAPI, Depends
from sqlalchemy.ext.asyncio import AsyncSession
from contextlib import asynccontextmanager

@asynccontextmanager
async def lifespan(app: FastAPI):
    # Startup
    await init_db()
    yield
    # Shutdown
    await cleanup()

app = FastAPI(lifespan=lifespan)

async def get_db() -> AsyncGenerator[AsyncSession, None]:
    async with async_session() as session:
        yield session

@app.get("/users/{user_id}")
async def get_user(
    user_id: int,
    db: AsyncSession = Depends(get_db)
) -> UserResponse:
    user = await get_user_by_id(db, user_id)
    return UserResponse.model_validate(user)
```

Class-Based Dependencies:
```python
from fastapi import Depends

class Paginator:
    def __init__(self, page: int = 1, size: int = 20):
        self.page = max(1, page)
        self.size = min(100, max(1, size))
        self.offset = (self.page - 1) * self.size

@app.get("/items/")
async def list_items(pagination: Paginator = Depends()) -> list[Item]:
    return await Item.get_page(pagination.offset, pagination.size)
```

### Django 5.2 LTS Features

Composite Primary Keys:
```python
from django.db import models

class OrderItem(models.Model):
    order = models.ForeignKey(Order, on_delete=models.CASCADE)
    product = models.ForeignKey(Product, on_delete=models.CASCADE)
    quantity = models.IntegerField()

    class Meta:
        pk = models.CompositePrimaryKey("order", "product")
```

URL Reverse with Query Parameters:
```python
from django.urls import reverse

url = reverse("search", query={"q": "django", "page": 1}, fragment="results")
# /search/?q=django&page=1#results
```

Automatic Model Imports in Shell:
```bash
python manage.py shell
# Models from all installed apps are automatically imported
```

### Pydantic v2.9 Deep Patterns

Reusable Validators with Annotated:
```python
from typing import Annotated
from pydantic import AfterValidator, BaseModel

def validate_positive(v: int) -> int:
    if v <= 0:
        raise ValueError("Must be positive")
    return v

PositiveInt = Annotated[int, AfterValidator(validate_positive)]

class Product(BaseModel):
    price: PositiveInt
    quantity: PositiveInt
```

Model Validator for Cross-Field Validation:
```python
from pydantic import BaseModel, model_validator
from typing import Self

class DateRange(BaseModel):
    start_date: date
    end_date: date

    @model_validator(mode="after")
    def validate_dates(self) -> Self:
        if self.end_date < self.start_date:
            raise ValueError("end_date must be after start_date")
        return self
```

ConfigDict Best Practices:
```python
from pydantic import BaseModel, ConfigDict

class BaseSchema(BaseModel):
    model_config = ConfigDict(
        from_attributes=True,      # ORM object support
        populate_by_name=True,     # Allow aliases
        extra="forbid",            # Fail on unknown fields
        str_strip_whitespace=True, # Clean strings
    )
```

### SQLAlchemy 2.0 Async Patterns

Engine and Session Setup:
```python
from sqlalchemy.ext.asyncio import (
    create_async_engine,
    async_sessionmaker,
    AsyncSession,
)

engine = create_async_engine(
    "postgresql+asyncpg://user:pass@localhost/db",
    pool_pre_ping=True,
    echo=True,
)

async_session = async_sessionmaker(
    engine,
    class_=AsyncSession,
    expire_on_commit=False,  # Prevent detached instance errors
)
```

Repository Pattern:
```python
from sqlalchemy import select
from sqlalchemy.ext.asyncio import AsyncSession

class UserRepository:
    def __init__(self, session: AsyncSession):
        self.session = session

    async def get_by_id(self, user_id: int) -> User | None:
        result = await self.session.execute(
            select(User).where(User.id == user_id)
        )
        return result.scalar_one_or_none()

    async def create(self, user: UserCreate) -> User:
        db_user = User(**user.model_dump())
        self.session.add(db_user)
        await self.session.commit()
        await self.session.refresh(db_user)
        return db_user
```

Streaming Large Results:
```python
async def stream_users(db: AsyncSession):
    result = await db.stream(select(User))
    async for user in result.scalars():
        yield user
```

### pytest Advanced Patterns

Async Fixtures with pytest-asyncio:
```python
import pytest
import pytest_asyncio
from httpx import AsyncClient

@pytest_asyncio.fixture
async def async_client(app) -> AsyncGenerator[AsyncClient, None]:
    async with AsyncClient(app=app, base_url="http://test") as client:
        yield client

@pytest_asyncio.fixture
async def db_session() -> AsyncGenerator[AsyncSession, None]:
    async with async_session() as session:
        async with session.begin():
            yield session
            await session.rollback()
```

Parametrized Tests:
```python
@pytest.mark.parametrize(
    "input_data,expected_status",
    [
        ({"name": "Valid"}, 201),
        ({"name": ""}, 422),
        ({}, 422),
    ],
    ids=["valid", "empty_name", "missing_name"],
)
async def test_create_user(async_client, input_data, expected_status):
    response = await async_client.post("/users/", json=input_data)
    assert response.status_code == expected_status
```

Fixture Factories:
```python
@pytest.fixture
def user_factory():
    async def _create_user(db: AsyncSession, **kwargs) -> User:
        defaults = {"name": "Test User", "email": "test@example.com"}
        user = User(**(defaults | kwargs))
        db.add(user)
        await db.commit()
        return user
    return _create_user
```

### Type Hints Modern Patterns

Protocol for Structural Typing:
```python
from typing import Protocol, runtime_checkable

@runtime_checkable
class Repository(Protocol[T]):
    async def get(self, id: int) -> T | None: ...
    async def create(self, data: dict) -> T: ...
    async def delete(self, id: int) -> bool: ...
```

ParamSpec for Decorators:
```python
from typing import ParamSpec, TypeVar, Callable
from functools import wraps

P = ParamSpec("P")
R = TypeVar("R")

def retry(times: int = 3) -> Callable[[Callable[P, R]], Callable[P, R]]:
    def decorator(func: Callable[P, R]) -> Callable[P, R]:
        @wraps(func)
        async def wrapper(*args: P.args, **kwargs: P.kwargs) -> R:
            for attempt in range(times):
                try:
                    return await func(*args, **kwargs)
                except Exception:
                    if attempt == times - 1:
                        raise
        return wrapper
    return decorator
```

### Package Management

pyproject.toml (Poetry):
```toml
[tool.poetry]
name = "my-project"
version = "1.0.0"
python = "^3.13"

[tool.poetry.dependencies]
fastapi = "^0.115.0"
pydantic = "^2.9.0"
sqlalchemy = {extras = ["asyncio"], version = "^2.0.0"}

[tool.poetry.group.dev.dependencies]
pytest = "^8.0"
pytest-asyncio = "^0.24"
ruff = "^0.8"

[tool.ruff]
line-length = 100
target-version = "py313"

[tool.pytest.ini_options]
asyncio_mode = "auto"
```

uv (Fast Package Manager):
```bash
# Install uv
curl -LsSf https://astral.sh/uv/install.sh | sh

# Create virtual environment
uv venv

# Install dependencies
uv pip install -r requirements.txt

# Add dependency
uv add fastapi
```

---

## Advanced Implementation (10+ minutes)

For comprehensive coverage including:
- Production deployment patterns (Docker, Kubernetes)
- Advanced async patterns (task groups, semaphores)
- Data science integration (numpy, pandas, polars)
- Performance optimization techniques
- Security best practices (OWASP patterns)
- CI/CD integration patterns

See:
- [reference.md](reference.md) - Complete reference documentation
- [examples.md](examples.md) - Production-ready code examples

---

## Context7 Library Mappings

```
/tiangolo/fastapi - FastAPI async web framework
/django/django - Django web framework
/pydantic/pydantic - Data validation with type annotations
/sqlalchemy/sqlalchemy - SQL toolkit and ORM
/pytest-dev/pytest - Testing framework
/numpy/numpy - Numerical computing
/pandas-dev/pandas - Data analysis library
/pola-rs/polars - Fast DataFrame library
```

---

## Works Well With

- `moai-domain-backend` - REST API and microservices architecture
- `moai-domain-database` - SQL patterns and ORM optimization
- `moai-quality-testing` - TDD and testing strategies
- `moai-essentials-debug` - AI-powered debugging
- `moai-foundation-trust` - TRUST 5 quality principles

---

## Troubleshooting

Common Issues:

Python Version Check:
```bash
python --version  # Should be 3.13+
python -c "import sys; print(sys.version_info)"
```

Async Session Detached Error:
- Solution: Set `expire_on_commit=False` in session config
- Or: Use `await session.refresh(obj)` after commit

pytest asyncio Mode Warning:
```toml
# pyproject.toml
[tool.pytest.ini_options]
asyncio_mode = "auto"
asyncio_default_fixture_loop_scope = "function"
```

Pydantic v2 Migration:
- `parse_obj()` is now `model_validate()`
- `parse_raw()` is now `model_validate_json()`
- `from_orm()` requires `from_attributes=True` in ConfigDict

---

Last Updated: 2025-12-07
Status: Active (v1.0.0)
