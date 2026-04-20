# Python 3.13+ Complete Reference

## Language Features Reference

### Python 3.13 Feature Matrix

| Feature | Status | PEP | Production Ready |
|---------|--------|-----|------------------|
| JIT Compiler | Experimental | PEP 744 | No |
| Free Threading (GIL-free) | Experimental | PEP 703 | No |
| Pattern Matching | Stable | PEP 634-636 | Yes |
| Type Parameter Syntax | Stable | PEP 695 | Yes |
| Exception Groups | Stable | PEP 654 | Yes |

### JIT Compiler Details (PEP 744)

Build Configuration:
```bash
# Build Python with JIT support
./configure --enable-experimental-jit
make

# Or with "disabled by default" mode
./configure --enable-experimental-jit=yes-off
make
```

Runtime Activation:
```bash
# Enable JIT at runtime
PYTHON_JIT=1 python my_script.py

# With debugging info
PYTHON_JIT=1 PYTHON_JIT_DEBUG=1 python my_script.py
```

Expected Benefits:
- 5-10% performance improvement for CPU-bound code
- Better optimization for hot loops
- Future foundation for more aggressive optimizations

### Free Threading (PEP 703)

Installation:
```bash
# macOS/Windows: Use official installers with free-threaded option
# Linux: Build from source
./configure --disable-gil
make

# Verify installation
python3.13t -c "import sys; print(sys._is_gil_enabled())"
```

Thread-Safe Patterns:
```python
import threading
from queue import Queue

def parallel_processing(items: list[str], workers: int = 4) -> list[str]:
    results = Queue()
    threads = []

    def worker(chunk: list[str]):
        for item in chunk:
            processed = heavy_computation(item)
            results.put(processed)

    chunk_size = len(items) // workers
    for i in range(workers):
        start = i * chunk_size
        end = start + chunk_size if i < workers - 1 else len(items)
        t = threading.Thread(target=worker, args=(items[start:end],))
        threads.append(t)
        t.start()

    for t in threads:
        t.join()

    return [results.get() for _ in range(results.qsize())]
```

### Pattern Matching Complete Guide

Literal Patterns:
```python
def http_status(status: int) -> str:
    match status:
        case 200:
            return "OK"
        case 201:
            return "Created"
        case 400:
            return "Bad Request"
        case 404:
            return "Not Found"
        case 500:
            return "Internal Server Error"
        case _:
            return f"Unknown status: {status}"
```

Structural Patterns:
```python
def process_event(event: dict) -> None:
    match event:
        case {"type": "click", "x": x, "y": y}:
            handle_click(x, y)
        case {"type": "keypress", "key": key, "modifiers": [*mods]}:
            handle_keypress(key, mods)
        case {"type": "scroll", "delta": delta, **rest}:
            handle_scroll(delta, rest)
```

Class Patterns:
```python
from dataclasses import dataclass

@dataclass
class Point:
    x: float
    y: float

@dataclass
class Circle:
    center: Point
    radius: float

@dataclass
class Rectangle:
    top_left: Point
    width: float
    height: float

def area(shape) -> float:
    match shape:
        case Circle(center=_, radius=r):
            return 3.14159 * r ** 2
        case Rectangle(width=w, height=h):
            return w * h
        case Point():
            return 0.0
```

Guard Clauses:
```python
def validate_user(user: dict) -> str:
    match user:
        case {"age": age} if age < 0:
            return "Invalid age"
        case {"age": age} if age < 18:
            return "Minor"
        case {"age": age, "verified": True} if age >= 18:
            return "Verified adult"
        case {"age": age} if age >= 18:
            return "Unverified adult"
        case _:
            return "Invalid user data"
```

---

## Web Framework Reference

### FastAPI 0.115+ Complete Reference

Application Structure:
```
project/
├── app/
│   ├── __init__.py
│   ├── main.py              # Application entry point
│   ├── config.py            # Settings and configuration
│   ├── dependencies.py      # Shared dependencies
│   ├── api/
│   │   ├── __init__.py
│   │   ├── v1/
│   │   │   ├── __init__.py
│   │   │   ├── router.py    # API router
│   │   │   └── endpoints/
│   │   │       ├── users.py
│   │   │       └── items.py
│   ├── core/
│   │   ├── security.py      # Auth and security
│   │   └── exceptions.py    # Custom exceptions
│   ├── models/
│   │   ├── __init__.py
│   │   ├── user.py          # SQLAlchemy models
│   │   └── item.py
│   ├── schemas/
│   │   ├── __init__.py
│   │   ├── user.py          # Pydantic schemas
│   │   └── item.py
│   ├── services/
│   │   ├── __init__.py
│   │   └── user_service.py  # Business logic
│   └── repositories/
│       ├── __init__.py
│       └── user_repo.py     # Data access
├── tests/
├── pyproject.toml
└── Dockerfile
```

Configuration with Pydantic Settings:
```python
# app/config.py
from pydantic_settings import BaseSettings, SettingsConfigDict
from functools import lru_cache

class Settings(BaseSettings):
    model_config = SettingsConfigDict(
        env_file=".env",
        env_file_encoding="utf-8",
        case_sensitive=False,
    )

    # Application
    app_name: str = "My API"
    debug: bool = False
    api_v1_prefix: str = "/api/v1"

    # Database
    database_url: str
    db_pool_size: int = 5
    db_max_overflow: int = 10

    # Security
    secret_key: str
    access_token_expire_minutes: int = 30
    algorithm: str = "HS256"

    # External Services
    redis_url: str | None = None

@lru_cache
def get_settings() -> Settings:
    return Settings()
```

Advanced Dependency Injection:
```python
# app/dependencies.py
from fastapi import Depends, HTTPException, status
from fastapi.security import OAuth2PasswordBearer
from sqlalchemy.ext.asyncio import AsyncSession
from jose import jwt, JWTError

oauth2_scheme = OAuth2PasswordBearer(tokenUrl="auth/token")

async def get_db() -> AsyncGenerator[AsyncSession, None]:
    async with async_session() as session:
        try:
            yield session
            await session.commit()
        except Exception:
            await session.rollback()
            raise

async def get_current_user(
    token: str = Depends(oauth2_scheme),
    db: AsyncSession = Depends(get_db),
) -> User:
    credentials_exception = HTTPException(
        status_code=status.HTTP_401_UNAUTHORIZED,
        detail="Could not validate credentials",
        headers={"WWW-Authenticate": "Bearer"},
    )
    try:
        payload = jwt.decode(token, SECRET_KEY, algorithms=[ALGORITHM])
        user_id: int = payload.get("sub")
        if user_id is None:
            raise credentials_exception
    except JWTError:
        raise credentials_exception

    user = await UserRepository(db).get_by_id(user_id)
    if user is None:
        raise credentials_exception
    return user

async def get_current_active_user(
    current_user: User = Depends(get_current_user),
) -> User:
    if not current_user.is_active:
        raise HTTPException(status_code=400, detail="Inactive user")
    return current_user

def require_role(required_role: str):
    async def role_checker(
        current_user: User = Depends(get_current_active_user),
    ) -> User:
        if current_user.role != required_role:
            raise HTTPException(
                status_code=status.HTTP_403_FORBIDDEN,
                detail="Insufficient permissions",
            )
        return current_user
    return role_checker
```

Background Tasks:
```python
from fastapi import BackgroundTasks

async def send_notification(email: str, message: str):
    # Simulate email sending
    await asyncio.sleep(1)
    print(f"Sent to {email}: {message}")

@app.post("/users/")
async def create_user(
    user: UserCreate,
    background_tasks: BackgroundTasks,
    db: AsyncSession = Depends(get_db),
) -> User:
    db_user = await UserRepository(db).create(user)
    background_tasks.add_task(
        send_notification,
        db_user.email,
        "Welcome to our platform!",
    )
    return db_user
```

### Django 5.2 LTS Reference

Composite Primary Keys:
```python
# models.py
from django.db import models

class Enrollment(models.Model):
    student = models.ForeignKey("Student", on_delete=models.CASCADE)
    course = models.ForeignKey("Course", on_delete=models.CASCADE)
    enrolled_at = models.DateTimeField(auto_now_add=True)
    grade = models.CharField(max_length=2, blank=True)

    class Meta:
        pk = models.CompositePrimaryKey("student", "course")
        verbose_name = "Enrollment"
        verbose_name_plural = "Enrollments"

# Usage
enrollment = Enrollment.objects.get(pk=(student_id, course_id))
```

Async Views and ORM:
```python
# views.py
from django.http import JsonResponse
from asgiref.sync import sync_to_async

async def async_user_list(request):
    users = await sync_to_async(list)(User.objects.all()[:100])
    return JsonResponse({"users": [u.to_dict() for u in users]})

# With Django 5.2 async ORM support
async def async_user_detail(request, user_id):
    user = await User.objects.aget(pk=user_id)
    return JsonResponse(user.to_dict())
```

Custom Form Rendering:
```python
# forms.py
from django import forms

class CustomBoundField(forms.BoundField):
    def label_tag(self, contents=None, attrs=None, label_suffix=None):
        attrs = attrs or {}
        attrs["class"] = attrs.get("class", "") + " custom-label"
        return super().label_tag(contents, attrs, label_suffix)

class CustomFormMixin:
    def get_bound_field(self, field, field_name):
        return CustomBoundField(self, field, field_name)

class UserForm(CustomFormMixin, forms.ModelForm):
    class Meta:
        model = User
        fields = ["name", "email"]
```

---

## Data Validation Reference

### Pydantic v2.9 Complete Patterns

Discriminated Unions:
```python
from typing import Literal, Union
from pydantic import BaseModel, Field

class EmailNotification(BaseModel):
    type: Literal["email"] = "email"
    recipient: str
    subject: str
    body: str

class SMSNotification(BaseModel):
    type: Literal["sms"] = "sms"
    phone_number: str
    message: str

class PushNotification(BaseModel):
    type: Literal["push"] = "push"
    device_token: str
    title: str
    body: str

Notification = Union[EmailNotification, SMSNotification, PushNotification]

class NotificationRequest(BaseModel):
    notification: Notification = Field(discriminator="type")
```

Computed Fields:
```python
from pydantic import BaseModel, computed_field

class Product(BaseModel):
    name: str
    price: float
    quantity: int
    tax_rate: float = 0.1

    @computed_field
    @property
    def subtotal(self) -> float:
        return self.price * self.quantity

    @computed_field
    @property
    def tax(self) -> float:
        return self.subtotal * self.tax_rate

    @computed_field
    @property
    def total(self) -> float:
        return self.subtotal + self.tax
```

Custom JSON Serialization:
```python
from pydantic import BaseModel, field_serializer
from datetime import datetime
from decimal import Decimal

class Transaction(BaseModel):
    id: int
    amount: Decimal
    created_at: datetime

    @field_serializer("amount")
    def serialize_amount(self, amount: Decimal) -> str:
        return f"${amount:.2f}"

    @field_serializer("created_at")
    def serialize_datetime(self, dt: datetime) -> str:
        return dt.isoformat()
```

TypeAdapter for Dynamic Validation:
```python
from pydantic import TypeAdapter
from typing import Any

# Validate arbitrary data without a model
int_adapter = TypeAdapter(int)
validated_int = int_adapter.validate_python("42")  # Returns 42

# Validate complex types
list_adapter = TypeAdapter(list[dict[str, int]])
validated_list = list_adapter.validate_json('[{"a": 1}, {"b": 2}]')

# Validate with custom types
UserListAdapter = TypeAdapter(list[User])
users = UserListAdapter.validate_python(raw_data)
```

---

## ORM Reference

### SQLAlchemy 2.0 Complete Patterns

Declarative Models with Type Hints:
```python
from sqlalchemy import String, ForeignKey
from sqlalchemy.orm import (
    DeclarativeBase,
    Mapped,
    mapped_column,
    relationship,
)
from datetime import datetime

class Base(DeclarativeBase):
    pass

class User(Base):
    __tablename__ = "users"

    id: Mapped[int] = mapped_column(primary_key=True)
    email: Mapped[str] = mapped_column(String(255), unique=True, index=True)
    name: Mapped[str] = mapped_column(String(100))
    created_at: Mapped[datetime] = mapped_column(default=datetime.utcnow)

    # Relationships
    posts: Mapped[list["Post"]] = relationship(back_populates="author")

class Post(Base):
    __tablename__ = "posts"

    id: Mapped[int] = mapped_column(primary_key=True)
    title: Mapped[str] = mapped_column(String(200))
    content: Mapped[str]
    author_id: Mapped[int] = mapped_column(ForeignKey("users.id"))

    author: Mapped["User"] = relationship(back_populates="posts")
```

Advanced Queries:
```python
from sqlalchemy import select, func, and_, or_
from sqlalchemy.orm import selectinload, joinedload

# Eager loading
async def get_user_with_posts(db: AsyncSession, user_id: int) -> User | None:
    result = await db.execute(
        select(User)
        .options(selectinload(User.posts))
        .where(User.id == user_id)
    )
    return result.scalar_one_or_none()

# Aggregations
async def get_post_counts_by_user(db: AsyncSession) -> list[tuple[str, int]]:
    result = await db.execute(
        select(User.name, func.count(Post.id).label("post_count"))
        .join(Post, isouter=True)
        .group_by(User.id)
        .order_by(func.count(Post.id).desc())
    )
    return result.all()

# Complex filtering
async def search_posts(
    db: AsyncSession,
    search: str | None = None,
    author_id: int | None = None,
    limit: int = 20,
) -> list[Post]:
    query = select(Post).options(joinedload(Post.author))

    conditions = []
    if search:
        conditions.append(
            or_(
                Post.title.ilike(f"%{search}%"),
                Post.content.ilike(f"%{search}%"),
            )
        )
    if author_id:
        conditions.append(Post.author_id == author_id)

    if conditions:
        query = query.where(and_(*conditions))

    result = await db.execute(query.limit(limit))
    return result.scalars().unique().all()
```

Upsert (Insert or Update):
```python
from sqlalchemy.dialects.postgresql import insert

async def upsert_user(db: AsyncSession, user_data: dict) -> User:
    stmt = insert(User).values(**user_data)
    stmt = stmt.on_conflict_do_update(
        index_elements=[User.email],
        set_={
            "name": stmt.excluded.name,
            "updated_at": datetime.utcnow(),
        },
    )
    await db.execute(stmt)
    await db.commit()

    return await db.execute(
        select(User).where(User.email == user_data["email"])
    ).scalar_one()
```

---

## Testing Reference

### pytest Complete Patterns

Conftest Configuration:
```python
# conftest.py
import pytest
import pytest_asyncio
from httpx import AsyncClient, ASGITransport
from sqlalchemy.ext.asyncio import create_async_engine, AsyncSession
from sqlalchemy.orm import sessionmaker

@pytest.fixture(scope="session")
def event_loop():
    """Create event loop for async tests."""
    import asyncio
    loop = asyncio.new_event_loop()
    yield loop
    loop.close()

@pytest_asyncio.fixture(scope="session")
async def engine():
    """Create test database engine."""
    engine = create_async_engine(
        "sqlite+aiosqlite:///:memory:",
        echo=True,
    )
    async with engine.begin() as conn:
        await conn.run_sync(Base.metadata.create_all)
    yield engine
    await engine.dispose()

@pytest_asyncio.fixture
async def db_session(engine) -> AsyncGenerator[AsyncSession, None]:
    """Create isolated database session for each test."""
    async_session = sessionmaker(
        engine, class_=AsyncSession, expire_on_commit=False
    )
    async with async_session() as session:
        async with session.begin():
            yield session
            await session.rollback()

@pytest_asyncio.fixture
async def async_client(db_session) -> AsyncGenerator[AsyncClient, None]:
    """Create async HTTP client for API testing."""
    def get_db_override():
        return db_session

    app.dependency_overrides[get_db] = get_db_override

    async with AsyncClient(
        transport=ASGITransport(app=app),
        base_url="http://test",
    ) as client:
        yield client

    app.dependency_overrides.clear()
```

Advanced Fixtures:
```python
@pytest.fixture
def user_factory(db_session):
    """Factory for creating test users."""
    created_users = []

    async def _create(**kwargs) -> User:
        defaults = {
            "name": f"User {len(created_users)}",
            "email": f"user{len(created_users)}@test.com",
        }
        user = User(**(defaults | kwargs))
        db_session.add(user)
        await db_session.flush()
        created_users.append(user)
        return user

    return _create

@pytest.fixture
def mock_external_api(mocker):
    """Mock external API calls."""
    return mocker.patch(
        "app.services.external_api.fetch_data",
        return_value={"status": "ok", "data": []},
    )
```

Hypothesis Property-Based Testing:
```python
from hypothesis import given, strategies as st
from hypothesis.extra.pydantic import from_model

@given(from_model(UserCreate))
def test_user_create_validation(user_data: UserCreate):
    """Test that any valid UserCreate can be processed."""
    assert user_data.name
    assert "@" in user_data.email

@given(st.lists(st.integers(min_value=0, max_value=100), min_size=1))
def test_calculate_average(numbers: list[int]):
    """Property: average is always between min and max."""
    avg = calculate_average(numbers)
    assert min(numbers) <= avg <= max(numbers)
```

---

## Type Hints Reference

### Modern Type Patterns

Generic Classes:
```python
from typing import Generic, TypeVar

T = TypeVar("T")
K = TypeVar("K")

class Cache(Generic[K, T]):
    def __init__(self, max_size: int = 100):
        self._cache: dict[K, T] = {}
        self._max_size = max_size

    def get(self, key: K) -> T | None:
        return self._cache.get(key)

    def set(self, key: K, value: T) -> None:
        if len(self._cache) >= self._max_size:
            oldest_key = next(iter(self._cache))
            del self._cache[oldest_key]
        self._cache[key] = value

# Usage
user_cache: Cache[int, User] = Cache(max_size=1000)
```

TypeVar with Bounds:
```python
from typing import TypeVar
from pydantic import BaseModel

ModelT = TypeVar("ModelT", bound=BaseModel)

def validate_and_create(model_class: type[ModelT], data: dict) -> ModelT:
    return model_class.model_validate(data)
```

Self Type:
```python
from typing import Self

class Builder:
    def __init__(self):
        self._config: dict = {}

    def with_option(self, key: str, value: str) -> Self:
        self._config[key] = value
        return self

    def build(self) -> dict:
        return self._config.copy()

# Subclassing works correctly
class AdvancedBuilder(Builder):
    def with_advanced_option(self, value: int) -> Self:
        self._config["advanced"] = value
        return self
```

---

## Context7 Integration

Library ID Resolution:
```python
# Step 1: Resolve library ID
library_id = await mcp__context7__resolve_library_id("fastapi")
# Returns: /tiangolo/fastapi

# Step 2: Get documentation
docs = await mcp__context7__get_library_docs(
    context7CompatibleLibraryID="/tiangolo/fastapi",
    topic="dependency injection async",
    tokens=5000,
)
```

Available Libraries:
| Library | Context7 ID | Topics |
|---------|-------------|--------|
| FastAPI | /tiangolo/fastapi | async, dependencies, security, websockets |
| Django | /django/django | views, models, forms, admin |
| Pydantic | /pydantic/pydantic | validation, serialization, settings |
| SQLAlchemy | /sqlalchemy/sqlalchemy | orm, async, queries, migrations |
| pytest | /pytest-dev/pytest | fixtures, markers, plugins |
| numpy | /numpy/numpy | arrays, broadcasting, ufuncs |
| pandas | /pandas-dev/pandas | dataframe, series, io |
| polars | /pola-rs/polars | lazy, expressions, streaming |

---

Last Updated: 2025-12-07
Version: 1.0.0
