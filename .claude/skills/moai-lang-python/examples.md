# Python Production-Ready Code Examples

## Complete FastAPI Application

### Project Structure
```
fastapi_app/
├── app/
│   ├── __init__.py
│   ├── main.py
│   ├── config.py
│   ├── database.py
│   ├── dependencies.py
│   ├── models/
│   │   ├── __init__.py
│   │   └── user.py
│   ├── schemas/
│   │   ├── __init__.py
│   │   └── user.py
│   ├── repositories/
│   │   ├── __init__.py
│   │   └── user_repository.py
│   ├── services/
│   │   ├── __init__.py
│   │   └── user_service.py
│   └── api/
│       ├── __init__.py
│       └── v1/
│           ├── __init__.py
│           ├── router.py
│           └── endpoints/
│               └── users.py
├── tests/
│   ├── conftest.py
│   ├── test_users.py
│   └── test_services.py
├── pyproject.toml
└── Dockerfile
```

### Main Application Entry

```python
# app/main.py
from contextlib import asynccontextmanager
from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware

from app.config import get_settings
from app.database import init_db, close_db
from app.api.v1.router import api_router

settings = get_settings()

@asynccontextmanager
async def lifespan(app: FastAPI):
    # Startup
    await init_db()
    yield
    # Shutdown
    await close_db()

app = FastAPI(
    title=settings.app_name,
    version="1.0.0",
    lifespan=lifespan,
    docs_url="/api/docs",
    redoc_url="/api/redoc",
)

# CORS middleware
app.add_middleware(
    CORSMiddleware,
    allow_origins=settings.cors_origins,
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# Include API router
app.include_router(api_router, prefix="/api/v1")

@app.get("/health")
async def health_check():
    return {"status": "healthy", "version": "1.0.0"}
```

### Configuration

```python
# app/config.py
from functools import lru_cache
from pydantic_settings import BaseSettings, SettingsConfigDict

class Settings(BaseSettings):
    model_config = SettingsConfigDict(
        env_file=".env",
        env_file_encoding="utf-8",
        case_sensitive=False,
    )

    # Application
    app_name: str = "FastAPI App"
    debug: bool = False
    environment: str = "development"

    # Database
    database_url: str = "postgresql+asyncpg://user:pass@localhost/db"
    db_pool_size: int = 5
    db_max_overflow: int = 10
    db_pool_timeout: int = 30

    # Security
    secret_key: str
    access_token_expire_minutes: int = 30
    refresh_token_expire_days: int = 7
    algorithm: str = "HS256"

    # CORS
    cors_origins: list[str] = ["http://localhost:3000"]

    # Redis (optional)
    redis_url: str | None = None

@lru_cache
def get_settings() -> Settings:
    return Settings()
```

### Database Setup

```python
# app/database.py
from sqlalchemy.ext.asyncio import (
    create_async_engine,
    async_sessionmaker,
    AsyncSession,
    AsyncEngine,
)
from sqlalchemy.orm import DeclarativeBase

from app.config import get_settings

settings = get_settings()

class Base(DeclarativeBase):
    pass

engine: AsyncEngine | None = None
async_session_factory: async_sessionmaker[AsyncSession] | None = None

async def init_db():
    global engine, async_session_factory

    engine = create_async_engine(
        settings.database_url,
        pool_size=settings.db_pool_size,
        max_overflow=settings.db_max_overflow,
        pool_timeout=settings.db_pool_timeout,
        pool_pre_ping=True,
        echo=settings.debug,
    )

    async_session_factory = async_sessionmaker(
        engine,
        class_=AsyncSession,
        expire_on_commit=False,
        autoflush=False,
    )

    # Create tables (for development only)
    if settings.debug:
        async with engine.begin() as conn:
            await conn.run_sync(Base.metadata.create_all)

async def close_db():
    global engine
    if engine:
        await engine.dispose()

async def get_db() -> AsyncSession:
    if async_session_factory is None:
        raise RuntimeError("Database not initialized")

    async with async_session_factory() as session:
        try:
            yield session
            await session.commit()
        except Exception:
            await session.rollback()
            raise
```

### SQLAlchemy Models

```python
# app/models/user.py
from datetime import datetime
from sqlalchemy import String, Boolean, DateTime, func
from sqlalchemy.orm import Mapped, mapped_column

from app.database import Base

class User(Base):
    __tablename__ = "users"

    id: Mapped[int] = mapped_column(primary_key=True, index=True)
    email: Mapped[str] = mapped_column(
        String(255), unique=True, index=True, nullable=False
    )
    hashed_password: Mapped[str] = mapped_column(String(255), nullable=False)
    name: Mapped[str] = mapped_column(String(100), nullable=False)
    is_active: Mapped[bool] = mapped_column(Boolean, default=True)
    is_superuser: Mapped[bool] = mapped_column(Boolean, default=False)
    created_at: Mapped[datetime] = mapped_column(
        DateTime(timezone=True), server_default=func.now()
    )
    updated_at: Mapped[datetime] = mapped_column(
        DateTime(timezone=True), server_default=func.now(), onupdate=func.now()
    )

    def __repr__(self) -> str:
        return f"<User(id={self.id}, email={self.email})>"
```

### Pydantic Schemas

```python
# app/schemas/user.py
from datetime import datetime
from pydantic import BaseModel, ConfigDict, EmailStr, Field

class UserBase(BaseModel):
    email: EmailStr
    name: str = Field(min_length=1, max_length=100)

class UserCreate(UserBase):
    password: str = Field(min_length=8, max_length=100)

class UserUpdate(BaseModel):
    name: str | None = Field(None, min_length=1, max_length=100)
    password: str | None = Field(None, min_length=8, max_length=100)

class UserResponse(UserBase):
    model_config = ConfigDict(from_attributes=True)

    id: int
    is_active: bool
    created_at: datetime
    updated_at: datetime

class UserListResponse(BaseModel):
    users: list[UserResponse]
    total: int
    page: int
    size: int

class Token(BaseModel):
    access_token: str
    refresh_token: str
    token_type: str = "bearer"

class TokenPayload(BaseModel):
    sub: int
    exp: datetime
    type: str  # "access" or "refresh"
```

### Repository Pattern

```python
# app/repositories/user_repository.py
from sqlalchemy import select, func
from sqlalchemy.ext.asyncio import AsyncSession

from app.models.user import User
from app.schemas.user import UserCreate, UserUpdate

class UserRepository:
    def __init__(self, session: AsyncSession):
        self.session = session

    async def get_by_id(self, user_id: int) -> User | None:
        result = await self.session.execute(
            select(User).where(User.id == user_id)
        )
        return result.scalar_one_or_none()

    async def get_by_email(self, email: str) -> User | None:
        result = await self.session.execute(
            select(User).where(User.email == email)
        )
        return result.scalar_one_or_none()

    async def get_multi(
        self,
        skip: int = 0,
        limit: int = 100,
        is_active: bool | None = None,
    ) -> tuple[list[User], int]:
        query = select(User)
        count_query = select(func.count(User.id))

        if is_active is not None:
            query = query.where(User.is_active == is_active)
            count_query = count_query.where(User.is_active == is_active)

        # Get total count
        total_result = await self.session.execute(count_query)
        total = total_result.scalar_one()

        # Get users
        query = query.offset(skip).limit(limit).order_by(User.created_at.desc())
        result = await self.session.execute(query)
        users = result.scalars().all()

        return list(users), total

    async def create(self, user_create: UserCreate, hashed_password: str) -> User:
        user = User(
            email=user_create.email,
            name=user_create.name,
            hashed_password=hashed_password,
        )
        self.session.add(user)
        await self.session.flush()
        await self.session.refresh(user)
        return user

    async def update(self, user: User, user_update: UserUpdate) -> User:
        update_data = user_update.model_dump(exclude_unset=True)
        for field, value in update_data.items():
            setattr(user, field, value)
        await self.session.flush()
        await self.session.refresh(user)
        return user

    async def delete(self, user: User) -> None:
        await self.session.delete(user)
        await self.session.flush()

    async def deactivate(self, user: User) -> User:
        user.is_active = False
        await self.session.flush()
        await self.session.refresh(user)
        return user
```

### Service Layer

```python
# app/services/user_service.py
from datetime import datetime, timedelta, timezone
from jose import jwt
from passlib.context import CryptContext

from app.config import get_settings
from app.models.user import User
from app.schemas.user import UserCreate, UserUpdate, Token, TokenPayload
from app.repositories.user_repository import UserRepository

settings = get_settings()
pwd_context = CryptContext(schemes=["bcrypt"], deprecated="auto")

class UserService:
    def __init__(self, repository: UserRepository):
        self.repository = repository

    @staticmethod
    def hash_password(password: str) -> str:
        return pwd_context.hash(password)

    @staticmethod
    def verify_password(plain_password: str, hashed_password: str) -> bool:
        return pwd_context.verify(plain_password, hashed_password)

    @staticmethod
    def create_token(user_id: int, token_type: str, expires_delta: timedelta) -> str:
        expire = datetime.now(timezone.utc) + expires_delta
        payload = TokenPayload(
            sub=user_id,
            exp=expire,
            type=token_type,
        )
        return jwt.encode(
            payload.model_dump(),
            settings.secret_key,
            algorithm=settings.algorithm,
        )

    def create_tokens(self, user: User) -> Token:
        access_token = self.create_token(
            user.id,
            "access",
            timedelta(minutes=settings.access_token_expire_minutes),
        )
        refresh_token = self.create_token(
            user.id,
            "refresh",
            timedelta(days=settings.refresh_token_expire_days),
        )
        return Token(access_token=access_token, refresh_token=refresh_token)

    async def authenticate(self, email: str, password: str) -> User | None:
        user = await self.repository.get_by_email(email)
        if not user:
            return None
        if not self.verify_password(password, user.hashed_password):
            return None
        if not user.is_active:
            return None
        return user

    async def register(self, user_create: UserCreate) -> User:
        hashed_password = self.hash_password(user_create.password)
        return await self.repository.create(user_create, hashed_password)

    async def update(self, user: User, user_update: UserUpdate) -> User:
        if user_update.password:
            user_update.password = self.hash_password(user_update.password)
        return await self.repository.update(user, user_update)
```

### API Endpoints

```python
# app/api/v1/endpoints/users.py
from fastapi import APIRouter, Depends, HTTPException, status, Query
from sqlalchemy.ext.asyncio import AsyncSession

from app.database import get_db
from app.dependencies import get_current_user, get_current_active_superuser
from app.models.user import User
from app.schemas.user import (
    UserCreate,
    UserUpdate,
    UserResponse,
    UserListResponse,
    Token,
)
from app.repositories.user_repository import UserRepository
from app.services.user_service import UserService

router = APIRouter(prefix="/users", tags=["users"])

def get_user_service(db: AsyncSession = Depends(get_db)) -> UserService:
    repository = UserRepository(db)
    return UserService(repository)

@router.post("/register", response_model=UserResponse, status_code=status.HTTP_201_CREATED)
async def register(
    user_create: UserCreate,
    service: UserService = Depends(get_user_service),
):
    """Register a new user."""
    existing = await service.repository.get_by_email(user_create.email)
    if existing:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail="Email already registered",
        )
    user = await service.register(user_create)
    return user

@router.post("/login", response_model=Token)
async def login(
    email: str,
    password: str,
    service: UserService = Depends(get_user_service),
):
    """Authenticate and get tokens."""
    user = await service.authenticate(email, password)
    if not user:
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail="Invalid credentials",
        )
    return service.create_tokens(user)

@router.get("/me", response_model=UserResponse)
async def get_current_user_info(
    current_user: User = Depends(get_current_user),
):
    """Get current user information."""
    return current_user

@router.patch("/me", response_model=UserResponse)
async def update_current_user(
    user_update: UserUpdate,
    current_user: User = Depends(get_current_user),
    service: UserService = Depends(get_user_service),
):
    """Update current user."""
    return await service.update(current_user, user_update)

@router.get("", response_model=UserListResponse)
async def list_users(
    page: int = Query(1, ge=1),
    size: int = Query(20, ge=1, le=100),
    is_active: bool | None = None,
    current_user: User = Depends(get_current_active_superuser),
    service: UserService = Depends(get_user_service),
):
    """List all users (admin only)."""
    skip = (page - 1) * size
    users, total = await service.repository.get_multi(
        skip=skip, limit=size, is_active=is_active
    )
    return UserListResponse(users=users, total=total, page=page, size=size)

@router.get("/{user_id}", response_model=UserResponse)
async def get_user(
    user_id: int,
    current_user: User = Depends(get_current_active_superuser),
    service: UserService = Depends(get_user_service),
):
    """Get user by ID (admin only)."""
    user = await service.repository.get_by_id(user_id)
    if not user:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail="User not found",
        )
    return user

@router.delete("/{user_id}", status_code=status.HTTP_204_NO_CONTENT)
async def deactivate_user(
    user_id: int,
    current_user: User = Depends(get_current_active_superuser),
    service: UserService = Depends(get_user_service),
):
    """Deactivate user (admin only)."""
    user = await service.repository.get_by_id(user_id)
    if not user:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail="User not found",
        )
    await service.repository.deactivate(user)
```

### Dependencies

```python
# app/dependencies.py
from fastapi import Depends, HTTPException, status
from fastapi.security import OAuth2PasswordBearer
from jose import jwt, JWTError
from sqlalchemy.ext.asyncio import AsyncSession

from app.config import get_settings
from app.database import get_db
from app.models.user import User
from app.repositories.user_repository import UserRepository
from app.schemas.user import TokenPayload

settings = get_settings()
oauth2_scheme = OAuth2PasswordBearer(tokenUrl="/api/v1/users/login")

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
        payload = jwt.decode(
            token, settings.secret_key, algorithms=[settings.algorithm]
        )
        token_data = TokenPayload(**payload)

        if token_data.type != "access":
            raise credentials_exception

    except JWTError:
        raise credentials_exception

    repository = UserRepository(db)
    user = await repository.get_by_id(token_data.sub)

    if user is None:
        raise credentials_exception

    return user

async def get_current_active_user(
    current_user: User = Depends(get_current_user),
) -> User:
    if not current_user.is_active:
        raise HTTPException(
            status_code=status.HTTP_403_FORBIDDEN,
            detail="Inactive user",
        )
    return current_user

async def get_current_active_superuser(
    current_user: User = Depends(get_current_active_user),
) -> User:
    if not current_user.is_superuser:
        raise HTTPException(
            status_code=status.HTTP_403_FORBIDDEN,
            detail="Insufficient permissions",
        )
    return current_user
```

---

## Complete pytest Test Suite

```python
# tests/conftest.py
import pytest
import pytest_asyncio
from httpx import AsyncClient, ASGITransport
from sqlalchemy.ext.asyncio import create_async_engine, AsyncSession
from sqlalchemy.orm import sessionmaker
from sqlalchemy.pool import StaticPool

from app.database import Base, get_db
from app.main import app
from app.config import get_settings

settings = get_settings()

@pytest.fixture(scope="session")
def event_loop():
    import asyncio
    loop = asyncio.new_event_loop()
    yield loop
    loop.close()

@pytest_asyncio.fixture(scope="session")
async def engine():
    engine = create_async_engine(
        "sqlite+aiosqlite:///:memory:",
        connect_args={"check_same_thread": False},
        poolclass=StaticPool,
    )
    async with engine.begin() as conn:
        await conn.run_sync(Base.metadata.create_all)
    yield engine
    await engine.dispose()

@pytest_asyncio.fixture
async def db_session(engine):
    async_session = sessionmaker(
        engine, class_=AsyncSession, expire_on_commit=False
    )
    async with async_session() as session:
        yield session
        await session.rollback()

@pytest_asyncio.fixture
async def async_client(db_session):
    async def override_get_db():
        yield db_session

    app.dependency_overrides[get_db] = override_get_db

    async with AsyncClient(
        transport=ASGITransport(app=app),
        base_url="http://test",
    ) as client:
        yield client

    app.dependency_overrides.clear()

@pytest.fixture
def user_data():
    return {
        "email": "test@example.com",
        "name": "Test User",
        "password": "password123",
    }
```

```python
# tests/test_users.py
import pytest
from httpx import AsyncClient

@pytest.mark.asyncio
async def test_register_user(async_client: AsyncClient, user_data: dict):
    response = await async_client.post("/api/v1/users/register", json=user_data)

    assert response.status_code == 201
    data = response.json()
    assert data["email"] == user_data["email"]
    assert data["name"] == user_data["name"]
    assert "id" in data
    assert "password" not in data

@pytest.mark.asyncio
async def test_register_duplicate_email(async_client: AsyncClient, user_data: dict):
    # First registration
    await async_client.post("/api/v1/users/register", json=user_data)

    # Second registration with same email
    response = await async_client.post("/api/v1/users/register", json=user_data)

    assert response.status_code == 400
    assert "already registered" in response.json()["detail"]

@pytest.mark.asyncio
@pytest.mark.parametrize(
    "invalid_data,expected_detail",
    [
        ({"email": "invalid", "name": "Test", "password": "pass123"}, "email"),
        ({"email": "test@example.com", "name": "", "password": "pass123"}, "name"),
        ({"email": "test@example.com", "name": "Test", "password": "short"}, "password"),
    ],
    ids=["invalid_email", "empty_name", "short_password"],
)
async def test_register_validation(
    async_client: AsyncClient,
    invalid_data: dict,
    expected_detail: str,
):
    response = await async_client.post("/api/v1/users/register", json=invalid_data)

    assert response.status_code == 422

@pytest.mark.asyncio
async def test_login_success(async_client: AsyncClient, user_data: dict):
    # Register user first
    await async_client.post("/api/v1/users/register", json=user_data)

    # Login
    response = await async_client.post(
        "/api/v1/users/login",
        params={"email": user_data["email"], "password": user_data["password"]},
    )

    assert response.status_code == 200
    data = response.json()
    assert "access_token" in data
    assert "refresh_token" in data
    assert data["token_type"] == "bearer"

@pytest.mark.asyncio
async def test_get_current_user(async_client: AsyncClient, user_data: dict):
    # Register and login
    await async_client.post("/api/v1/users/register", json=user_data)
    login_response = await async_client.post(
        "/api/v1/users/login",
        params={"email": user_data["email"], "password": user_data["password"]},
    )
    token = login_response.json()["access_token"]

    # Get current user
    response = await async_client.get(
        "/api/v1/users/me",
        headers={"Authorization": f"Bearer {token}"},
    )

    assert response.status_code == 200
    data = response.json()
    assert data["email"] == user_data["email"]
```

---

## Async Patterns Examples

### Task Groups (Python 3.11+)

```python
import asyncio
from typing import Any

async def fetch_user(user_id: int) -> dict:
    await asyncio.sleep(0.1)  # Simulate API call
    return {"id": user_id, "name": f"User {user_id}"}

async def fetch_all_users(user_ids: list[int]) -> list[dict]:
    async with asyncio.TaskGroup() as tg:
        tasks = [tg.create_task(fetch_user(uid)) for uid in user_ids]

    return [task.result() for task in tasks]

# Exception handling with TaskGroup
async def fetch_with_error_handling(user_ids: list[int]) -> tuple[list[dict], list[Exception]]:
    results = []
    errors = []

    async def safe_fetch(user_id: int):
        try:
            result = await fetch_user(user_id)
            results.append(result)
        except Exception as e:
            errors.append(e)

    async with asyncio.TaskGroup() as tg:
        for uid in user_ids:
            tg.create_task(safe_fetch(uid))

    return results, errors
```

### Semaphore for Rate Limiting

```python
import asyncio
from contextlib import asynccontextmanager

class RateLimiter:
    def __init__(self, max_concurrent: int = 10):
        self._semaphore = asyncio.Semaphore(max_concurrent)

    @asynccontextmanager
    async def acquire(self):
        async with self._semaphore:
            yield

rate_limiter = RateLimiter(max_concurrent=5)

async def rate_limited_fetch(url: str) -> dict:
    async with rate_limiter.acquire():
        async with httpx.AsyncClient() as client:
            response = await client.get(url)
            return response.json()
```

### Async Generator Streaming

```python
from typing import AsyncGenerator

async def stream_large_data(
    db: AsyncSession,
    batch_size: int = 1000,
) -> AsyncGenerator[list[User], None]:
    offset = 0
    while True:
        result = await db.execute(
            select(User)
            .offset(offset)
            .limit(batch_size)
            .order_by(User.id)
        )
        users = result.scalars().all()

        if not users:
            break

        yield users
        offset += batch_size

# Usage
async def process_all_users(db: AsyncSession):
    async for batch in stream_large_data(db):
        for user in batch:
            await process_user(user)
```

---

## Docker Production Dockerfile

```dockerfile
# Dockerfile
FROM python:3.13-slim AS builder

WORKDIR /app

# Install build dependencies
RUN apt-get update && apt-get install -y --no-install-recommends \
    build-essential \
    && rm -rf /var/lib/apt/lists/*

# Install Python dependencies
COPY pyproject.toml poetry.lock ./
RUN pip install poetry && \
    poetry config virtualenvs.create false && \
    poetry install --no-dev --no-interaction --no-ansi

FROM python:3.13-slim AS runtime

WORKDIR /app

# Create non-root user
RUN addgroup --system --gid 1001 appgroup && \
    adduser --system --uid 1001 --gid 1001 appuser

# Copy dependencies from builder
COPY --from=builder /usr/local/lib/python3.13/site-packages /usr/local/lib/python3.13/site-packages
COPY --from=builder /usr/local/bin /usr/local/bin

# Copy application code
COPY --chown=appuser:appgroup . .

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 8000

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD python -c "import httpx; httpx.get('http://localhost:8000/health')"

# Run application
CMD ["uvicorn", "app.main:app", "--host", "0.0.0.0", "--port", "8000", "--workers", "4"]
```

---

## pyproject.toml Complete Configuration

```toml
[tool.poetry]
name = "fastapi-app"
version = "1.0.0"
description = "Production FastAPI Application"
authors = ["Developer <dev@example.com>"]
python = "^3.13"

[tool.poetry.dependencies]
python = "^3.13"
fastapi = "^0.115.0"
uvicorn = {extras = ["standard"], version = "^0.32.0"}
pydantic = "^2.9.0"
pydantic-settings = "^2.6.0"
sqlalchemy = {extras = ["asyncio"], version = "^2.0.0"}
asyncpg = "^0.30.0"
python-jose = {extras = ["cryptography"], version = "^3.3.0"}
passlib = {extras = ["bcrypt"], version = "^1.7.4"}
httpx = "^0.28.0"

[tool.poetry.group.dev.dependencies]
pytest = "^8.3.0"
pytest-asyncio = "^0.24.0"
pytest-cov = "^6.0.0"
aiosqlite = "^0.20.0"
ruff = "^0.8.0"
mypy = "^1.13.0"

[tool.ruff]
line-length = 100
target-version = "py313"

[tool.ruff.lint]
select = ["E", "F", "I", "N", "W", "UP", "B", "C4", "SIM"]
ignore = ["E501"]

[tool.ruff.lint.isort]
known-first-party = ["app"]

[tool.pytest.ini_options]
asyncio_mode = "auto"
asyncio_default_fixture_loop_scope = "function"
testpaths = ["tests"]
addopts = "-v --tb=short --cov=app --cov-report=term-missing"

[tool.mypy]
python_version = "3.13"
strict = true
plugins = ["pydantic.mypy"]

[tool.coverage.run]
source = ["app"]
omit = ["tests/*"]

[build-system]
requires = ["poetry-core>=1.0.0"]
build-backend = "poetry.core.masonry.api"
```

---

Last Updated: 2025-12-07
Version: 1.0.0
