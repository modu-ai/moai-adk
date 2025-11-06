---
name: moai-lang-python
version: 2.0.0
created: 2025-11-06
updated: 2025-11-06
status: active
description: "Python best practices with modern frameworks, AI/ML integration, and performance optimization for 2025"
keywords: [python, programming, backend, ai, ml, data-science, fastapi, django, testing]
allowed-tools:
  - Read
  - Write
  - Edit
  - Bash
  - WebFetch
  - WebSearch
---

# Python Development Mastery

**Modern Python Development with 2025 Best Practices**

> Comprehensive Python development guidance covering backend APIs, AI/ML integration, data science, and production-ready applications using the latest tools and frameworks.

## What It Does

- **Backend API Development**: FastAPI, Django, Flask applications with modern async patterns
- **AI/ML Integration**: TensorFlow, PyTorch, scikit-learn with deployment strategies
- **Data Science Pipelines**: Pandas, NumPy, Polars for high-performance data processing
- **Testing & Quality**: pytest, coverage, type checking with mypy and ruff
- **Performance Optimization**: Async programming, profiling, memory optimization
- **Production Deployment**: Docker, Kubernetes, CI/CD pipelines, monitoring
- **Security Best Practices**: Input validation, authentication, dependency scanning
- **Code Quality**: Type hints, formatting, linting, pre-commit hooks

## When to Use

### Perfect Scenarios
- **Building REST APIs and microservices**
- **Developing AI/ML models and data pipelines**
- **Creating web applications with Django/Flask**
- **Data analysis and scientific computing**
- **Automating workflows and DevOps tasks**
- **Prototyping and rapid development**
- **Integrating with cloud services (AWS, GCP, Azure)**

### Common Triggers
- "Create a Python API"
- "Set up Django project"
- "Optimize Python performance"
- "Test Python code"
- "Deploy Python application"
- "Python best practices"
- "AI/ML with Python"

## Tool Version Matrix (2025-11-06)

### Core Python
- **Python**: 3.13.0 (latest) / 3.12.x (LTS)
- **Package Managers**: uv 0.5.x (primary), pip 24.x, poetry 1.8.x
- **Virtual Environment**: venv (built-in), conda 24.x, uv venv

### Web Frameworks
- **FastAPI**: 0.115.x - High-performance async APIs
- **Django**: 5.1.x - Full-stack web framework
- **Flask**: 3.1.x - Lightweight web framework
- **Starlette**: 0.41.x - ASGI toolkit

### Testing Tools
- **pytest**: 8.3.x - Primary testing framework
- **pytest-asyncio**: 0.24.x - Async testing support
- **coverage**: 7.6.x - Code coverage analysis
- **tox**: 4.23.x - Multi-environment testing

### Code Quality
- **ruff**: 0.7.x - Fast linter and formatter
- **black**: 24.x - Code formatting
- **mypy**: 1.13.x - Static type checking
- **pre-commit**: 4.0.x - Git hooks

### AI/ML Stack
- **TensorFlow**: 2.18.x - Deep learning
- **PyTorch**: 2.5.x - Deep learning framework
- **scikit-learn**: 1.6.x - Machine learning
- **pandas**: 2.2.x - Data manipulation
- **polars**: 1.9.x - High-performance dataframes

### Performance Tools
- **py-spy**: 0.3.x - Sampling profiler
- **memory-profiler**: 0.64.x - Memory profiling
- **line-profiler**: 4.2.x - Line-by-line profiling

## Ecosystem Overview

### Package Management

```bash
# Modern Python with uv (recommended)
uv init project-name
uv add fastapi uvicorn pytest ruff mypy
uv run python main.py

# Traditional approach
python -m venv venv
source venv/bin/activate  # Windows: venv\Scripts\activate
pip install -r requirements.txt
```

### Project Structure (2025 Best Practice)

```
my-python-project/
├── pyproject.toml          # Modern project configuration
├── src/
│   └── my_project/
│       ├── __init__.py
│       ├── main.py         # Application entry point
│       ├── api/            # API endpoints
│       ├── models/         # Data models
│       ├── services/       # Business logic
│       └── utils/          # Utilities
├── tests/
│   ├── unit/               # Unit tests
│   ├── integration/        # Integration tests
│   └── conftest.py         # pytest configuration
├── .github/
│   └── workflows/          # CI/CD pipelines
├── Dockerfile
├── README.md
└── .pre-commit-config.yaml
```

## Modern Development Patterns

### Type-Driven Development with PEP 695

```python
from typing import TypeVar, Iterable, Protocol
from collections.abc import Sequence

# New type parameter syntax (Python 3.12+)
def max[T](args: Iterable[T]) -> T: ...
type Point = tuple[float, float]

# Runtime-checkable protocols
class Drawable(Protocol):
    def draw(self) -> None: ...

class Shape:
    def draw(self) -> None:
        print("Drawing shape")

def render(obj: Drawable) -> None:
    obj.draw()
```

### Async Programming Patterns

```python
import asyncio
from contextlib import asynccontextmanager
from fastapi import FastAPI
from typing import AsyncGenerator

@asynccontextmanager
async def lifespan(app: FastAPI) -> AsyncGenerator[None, None]:
    # Startup
    print("Application starting...")
    yield
    # Shutdown
    print("Application shutting down...")

app = FastAPI(lifespan=lifespan)

# Efficient async patterns
async def process_items(items: Sequence[str]) -> list[str]:
    # Process items concurrently
    semaphore = asyncio.Semaphore(10)  # Limit concurrency
    async def process_item(item: str) -> str:
        async with semaphore:
            return await expensive_operation(item)
    
    tasks = [process_item(item) for item in items]
    return await asyncio.gather(*tasks)
```

### Data Class Patterns with Pydantic v2

```python
from pydantic import BaseModel, Field, field_validator, computed_field
from datetime import datetime
from typing import Optional

class User(BaseModel):
    id: int
    username: str = Field(min_length=3, max_length=50)
    email: str = Field(pattern=r"[^@]+@[^@]+\.[^@]+")
    created_at: datetime
    updated_at: Optional[datetime] = None
    
    @field_validator('username')
    @classmethod
    def username_must_be_alpha(cls, v: str) -> str:
        if not v.isalnum():
            raise ValueError('Username must be alphanumeric')
        return v.lower()
    
    @computed_field
    @property
    def display_name(self) -> str:
        return self.username.title()
```

## Performance Considerations

### Algorithmic Optimization

```python
# Use built-in functions and comprehensions
def process_data_fast(data: list[int]) -> list[int]:
    # List comprehensions are optimized in Python 3.12+
    return [x * 2 for x in data if x % 2 == 0]

# Use efficient data structures
from collections import deque
from typing import Deque

def sliding_window(items: list[int], window_size: int) -> list[int]:
    window: Deque[int] = deque(maxlen=window_size)
    results = []
    for item in items:
        window.append(item)
        if len(window) == window_size:
            results.append(sum(window))
    return results
```

### Memory Optimization

```python
import gc
import weakref
from dataclasses import dataclass
from typing import Optional

@dataclass(slots=True)  # Memory-efficient dataclass
class Point:
    x: float
    y: float

# Use generators for large datasets
def process_large_file(filename: str) -> None:
    with open(filename) as file:
        for line in file:  # Generator pattern
            yield process_line(line)

# Weak references to avoid memory leaks
class Cache:
    def __init__(self):
        self._cache = weakref.WeakValueDictionary()
    
    def get(self, key: str) -> Optional[object]:
        return self._cache.get(key)
```

### Profiling and Monitoring

```bash
# Performance profiling
py-spy top -- python my_app.py
py-spy record -o profile.svg -- python my_app.py

# Memory profiling
python -m memory_profiler my_app.py

# Line-by-line profiling
kernprof -l -v my_script.py
```

## Testing Strategy

### pytest Configuration (pyproject.toml)

```toml
[tool.pytest.ini_options]
minversion = "6.0"
addopts = [
    "-ra",
    "--strict-markers",
    "--strict-config",
    "--cov=src",
    "--cov-report=term-missing",
    "--cov-report=html",
    "--cov-report=xml",
]
testpaths = ["tests"]
python_files = ["test_*.py", "*_test.py"]
python_classes = ["Test*"]
python_functions = ["test_*"]
markers = [
    "slow: marks tests as slow",
    "integration: marks tests as integration tests",
    "unit: marks tests as unit tests",
]
```

### Modern Testing Patterns

```python
import pytest
from unittest.mock import AsyncMock, patch
from httpx import AsyncClient
from typing import Generator

# Fixtures with dependency injection
@pytest.fixture
async def async_client() -> AsyncGenerator[AsyncClient, None]:
    from my_app.main import app
    async with AsyncClient(app=app, base_url="http://test") as ac:
        yield ac

# Async testing
@pytest.mark.asyncio
async def test_create_user(async_client: AsyncClient) -> None:
    response = await async_client.post(
        "/users",
        json={"username": "testuser", "email": "test@example.com"}
    )
    assert response.status_code == 201
    assert response.json()["username"] == "testuser"

# Property-based testing with hypothesis
from hypothesis import given, strategies as st

@given(st.lists(st.integers(), min_size=1))
def test_sort_preserves_length(numbers: list[int]) -> None:
    sorted_numbers = sorted(numbers)
    assert len(sorted_numbers) == len(numbers)
```

### Integration Testing

```python
import pytest
from testcontainers.compose import DockerCompose
from testcontainers.postgres import PostgresContainer
import asyncio

@pytest.fixture(scope="session")
async def postgres_container() -> AsyncGenerator[str, None]:
    with PostgresContainer("postgres:16") as postgres:
        yield postgres.get_connection_url()

@pytest.mark.integration
async def test_database_operations(postgres_container: str) -> None:
    # Test with real database
    await setup_database(postgres_container)
    result = await perform_database_operation()
    assert result.success
```

## Security Best Practices

### Input Validation and Sanitization

```python
from pydantic import BaseModel, Field, validator
import bleach
import re

class UserInput(BaseModel):
    content: str = Field(max_length=10000)
    
    @validator('content')
    def sanitize_content(cls, v: str) -> str:
        # Remove HTML tags and scripts
        return bleach.clean(v, tags=[], attributes={})

class SecureForm(BaseModel):
    email: str = Field(regex=r"^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$")
    phone: Optional[str] = Field(regex=r"^\+?1?\d{9,15}$")
    
    @validator('phone')
    def validate_phone(cls, v: Optional[str]) -> Optional[str]:
        if v and not re.match(r"^\+?1?\d{9,15}$", v):
            raise ValueError("Invalid phone number format")
        return v
```

### Authentication and Authorization

```python
from fastapi import Depends, HTTPException, status
from fastapi.security import OAuth2PasswordBearer
from jose import JWTError, jwt
from passlib.context import CryptContext
from datetime import datetime, timedelta

SECRET_KEY = "your-secret-key-here"
ALGORITHM = "HS256"
ACCESS_TOKEN_EXPIRE_MINUTES = 30

pwd_context = CryptContext(schemes=["bcrypt"], deprecated="auto")
oauth2_scheme = OAuth2PasswordBearer(tokenUrl="token")

def verify_password(plain_password: str, hashed_password: str) -> bool:
    return pwd_context.verify(plain_password, hashed_password)

def get_password_hash(password: str) -> str:
    return pwd_context.hash(password)

async def get_current_user(token: str = Depends(oauth2_scheme)):
    credentials_exception = HTTPException(
        status_code=status.HTTP_401_UNAUTHORIZED,
        detail="Could not validate credentials",
        headers={"WWW-Authenticate": "Bearer"},
    )
    try:
        payload = jwt.decode(token, SECRET_KEY, algorithms=[ALGORITHM])
        username: str = payload.get("sub")
        if username is None:
            raise credentials_exception
    except JWTError:
        raise credentials_exception
    
    user = get_user(username)
    if user is None:
        raise credentials_exception
    return user
```

### Dependency Security Scanning

```bash
# Add to pyproject.toml
[tool.pip-audit]
requirements = ["pyproject.toml"]
ignore-vulns = []

# Regular security scanning
pip-audit --requirement pyproject.toml
safety check --json
bandit -r src/ -f json
```

## Integration Patterns

### Database Integration

```python
# SQLAlchemy 2.0 async patterns
from sqlalchemy.ext.asyncio import create_async_engine, AsyncSession
from sqlalchemy.orm import declarative_base, sessionmaker
from sqlalchemy import Column, Integer, String, DateTime

Base = declarative_base()

class User(Base):
    __tablename__ = "users"
    
    id = Column(Integer, primary_key=True)
    username = Column(String, unique=True)
    created_at = Column(DateTime, default=datetime.utcnow)

# Async database session
async def get_async_session() -> AsyncGenerator[AsyncSession, None]:
    engine = create_async_engine("postgresql+asyncpg://user:pass@localhost/db")
    async_session = sessionmaker(
        engine, class_=AsyncSession, expire_on_commit=False
    )
    async with async_session() as session:
        yield session

# Usage in FastAPI
@app.get("/users/{user_id}")
async def get_user(user_id: int, db: AsyncSession = Depends(get_async_session)):
    result = await db.execute(select(User).where(User.id == user_id))
    user = result.scalar_one_or_none()
    if user is None:
        raise HTTPException(status_code=404, detail="User not found")
    return user
```

### Cloud Service Integration

```python
# AWS S3 with boto3
import boto3
from botocore.exceptions import ClientError
import io

class S3Service:
    def __init__(self, bucket_name: str):
        self.s3_client = boto3.client("s3")
        self.bucket_name = bucket_name
    
    async def upload_file(self, key: str, file_data: bytes) -> bool:
        try:
            self.s3_client.put_object(
                Bucket=self.bucket_name,
                Key=key,
                Body=file_data
            )
            return True
        except ClientError:
            return False

# Redis with aioredis
import aioredis
from typing import Optional

class CacheService:
    def __init__(self, redis_url: str):
        self.redis_url = redis_url
        self._redis: Optional[aioredis.Redis] = None
    
    async def connect(self) -> None:
        self._redis = await aioredis.from_url(self.redis_url)
    
    async def get(self, key: str) -> Optional[str]:
        if self._redis:
            return await self._redis.get(key)
        return None
    
    async def set(self, key: str, value: str, expire: int = 3600) -> None:
        if self._redis:
            await self._redis.setex(key, expire, value)
```

### Message Queue Integration

```python
# Celery with Redis
from celery import Celery
from celery.result import AsyncResult

celery_app = Celery(
    "tasks",
    broker="redis://localhost:6379/0",
    backend="redis://localhost:6379/1"
)

@celery_app.task
def process_data_task(data: dict) -> dict:
    # Heavy processing task
    result = expensive_computation(data)
    return {"result": result, "status": "completed"}

# In FastAPI
@app.post("/process")
async def start_processing(data: dict):
    task = process_data_task.delay(data)
    return {"task_id": task.id, "status": "processing"}

@app.get("/task/{task_id}")
async def get_task_status(task_id: str):
    result = AsyncResult(task_id)
    return {
        "task_id": task_id,
        "status": result.status,
        "result": result.result if result.successful() else None
    }
```

## Modern Development Workflow

### pyproject.toml Configuration

```toml
[build-system]
requires = ["hatchling"]
build-backend = "hatchling.build"

[project]
name = "my-python-project"
version = "0.1.0"
description = "Modern Python project"
dependencies = [
    "fastapi>=0.115.0",
    "uvicorn[standard]>=0.32.0",
    "pydantic>=2.9.0",
    "sqlalchemy>=2.0.0",
    "alembic>=1.13.0",
]

[project.optional-dependencies]
dev = [
    "pytest>=8.3.0",
    "pytest-asyncio>=0.24.0",
    "pytest-cov>=5.0.0",
    "ruff>=0.7.0",
    "mypy>=1.13.0",
    "black>=24.0.0",
    "pre-commit>=4.0.0",
]
test = [
    "pytest>=8.3.0",
    "pytest-asyncio>=0.24.0",
    "pytest-cov>=5.0.0",
    "httpx>=0.28.0",
    "testcontainers>=3.7.0",
]

[tool.ruff]
line-length = 88
target-version = "py312"
select = ["E", "F", "W", "I", "N", "B", "C90", "UP"]

[tool.mypy]
python_version = "3.12"
warn_return_any = true
warn_unused_configs = true
disallow_untyped_defs = true
```

### Pre-commit Configuration

```yaml
# .pre-commit-config.yaml
repos:
  - repo: https://github.com/pre-commit/pre-commit-hooks
    rev: v5.0.0
    hooks:
      - id: trailing-whitespace
      - id: end-of-file-fixer
      - id: check-yaml
      - id: check-added-large-files
      - id: check-merge-conflict

  - repo: https://github.com/astral-sh/ruff-pre-commit
    rev: v0.7.0
    hooks:
      - id: ruff
        args: [--fix]
      - id: ruff-format

  - repo: https://github.com/pre-commit/mirrors-mypy
    rev: v1.13.0
    hooks:
      - id: mypy
        additional_dependencies: [types-all]

  - repo: https://github.com/pycqa/bandit
    rev: 1.7.6
    hooks:
      - id: bandit
        args: [-r, src/]
```

### Docker Best Practices

```dockerfile
# Multi-stage build
FROM python:3.13-slim as builder

WORKDIR /app
COPY pyproject.toml .
RUN pip install --no-cache-dir uv && \
    uv pip install --no-deps -r <(uv pip compile pyproject.toml)

FROM python:3.13-slim as runtime

# Security best practices
RUN adduser --disabled-password --gecos '' appuser && \
    mkdir -p /app && chown -R appuser:appuser /app

WORKDIR /app
COPY --from=builder /usr/local/lib/python3.13/site-packages /usr/local/lib/python3.13/site-packages
COPY --from=builder /usr/local/bin /usr/local/bin
COPY --chown=appuser:appuser . .

USER appuser
EXPOSE 8000

CMD ["uvicorn", "my_project.main:app", "--host", "0.0.0.0", "--port", "8000"]
```

## AI/ML Integration

### Model Deployment with FastAPI

```python
import torch
import numpy as np
from fastapi import FastAPI, File, UploadFile
from PIL import Image
import io

app = FastAPI()

# Load model at startup
model = None

@app.on_event("startup")
async def load_model():
    global model
    model = torch.load("model.pth")
    model.eval()

@app.post("/predict")
async def predict(file: UploadFile = File(...)):
    # Preprocess image
    image = Image.open(io.BytesIO(await file.read()))
    image = preprocess_image(image)
    
    # Make prediction
    with torch.no_grad():
        tensor = torch.tensor(image).unsqueeze(0)
        prediction = model(tensor)
    
    return {
        "prediction": prediction.tolist(),
        "confidence": torch.softmax(prediction, dim=1).max().item()
    }

def preprocess_image(image: Image.Image) -> np.ndarray:
    # Image preprocessing logic
    image = image.resize((224, 224))
    image_array = np.array(image) / 255.0
    return image_array.transpose(2, 0, 1)
```

### Data Pipeline with Polars

```python
import polars as pl
from typing import Optional

class DataProcessor:
    def __init__(self, data_path: str):
        self.data_path = data_path
        self.df: Optional[pl.DataFrame] = None
    
    def load_data(self) -> None:
        """Load data using high-performance polars"""
        self.df = pl.read_parquet(self.data_path)
    
    def process_data(self) -> pl.DataFrame:
        """Process data with polars lazy evaluation"""
        return (
            self.df.lazy()
            .filter(pl.col("value") > 0)
            .group_by("category")
            .agg([
                pl.mean("value").alias("avg_value"),
                pl.std("value").alias("std_value"),
                pl.len().alias("count")
            ])
            .sort("avg_value", descending=True)
            .collect()
        )
    
    def export_results(self, output_path: str) -> None:
        """Export processed data"""
        processed = self.process_data()
        processed.write_csv(output_path)
```

---

**Created by**: MoAI Language Skill Factory  
**Last Updated**: 2025-11-06  
**Version**: 2.0.0  
**Python Target**: 3.12+ with modern async and type features  

This skill provides comprehensive Python development guidance with 2025 best practices, covering everything from basic project setup to advanced AI/ML integration and production deployment patterns.
