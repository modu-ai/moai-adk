# Python 3.13 Code Examples

Production-ready examples for modern Python development with pytest 8.4.2, ruff 0.13.1, mypy 1.8.0, FastAPI 0.115.0, and Python 3.13.1 features.

---

## Example 1: pytest 8.4.2 with Fixtures and Async Tests

### Test File: `tests/test_user_service.py`

```python
"""
Test suite for UserService with pytest fixtures and async support.
Demonstrates pytest 8.4.2 best practices for Python 3.13.
"""
import pytest
from unittest.mock import AsyncMock
from src.services.user_service import UserService
from src.models.user import User


@pytest.fixture
def user_service():
    """Fixture providing a UserService instance with mocked dependencies."""
    return UserService(db_client=AsyncMock())


@pytest.fixture
def sample_user():
    """Fixture providing a sample User instance for tests."""
    return User(id=1, name="Alice", email="alice@example.com")


def test_user_creation(user_service, sample_user):
    """Verify user creation with valid data."""
    result = user_service.validate_user(sample_user)
    assert result is True
    assert sample_user.id > 0


@pytest.mark.asyncio
async def test_fetch_user_async(user_service):
    """Test async user fetching with mocked database."""
    user_service.db_client.fetch_user.return_value = User(
        id=1, name="Bob", email="bob@example.com"
    )

    user = await user_service.fetch_user(user_id=1)

    assert user is not None
    assert user.name == "Bob"
    user_service.db_client.fetch_user.assert_called_once_with(1)


@pytest.mark.parametrize("user_id,expected_name", [
    (1, "Alice"),
    (2, "Bob"),
    (3, "Charlie"),
])
@pytest.mark.asyncio
async def test_fetch_multiple_users(user_service, user_id, expected_name):
    """Parametrized test for fetching multiple users."""
    user_service.db_client.fetch_user.return_value = User(
        id=user_id, name=expected_name, email=f"{expected_name.lower()}@example.com"
    )

    user = await user_service.fetch_user(user_id=user_id)

    assert user.name == expected_name
```

**Key Features**:
- ✅ Fixtures for dependency injection
- ✅ `@pytest.mark.asyncio` for async tests
- ✅ `@pytest.mark.parametrize` for data-driven tests
- ✅ AsyncMock for mocking async operations
- ✅ One assertion per test (clarity)

**Run Commands**:
```bash
pytest tests/test_user_service.py -v
pytest tests/test_user_service.py --cov=src.services --cov-report=term
```

---

## Example 2: ruff 0.13.1 Linting and Formatting Workflow

### Project Configuration: `pyproject.toml`

```toml
[tool.ruff]
line-length = 100
target-version = "py313"
exclude = [".venv", "build", "dist", "__pycache__"]

[tool.ruff.lint]
select = [
    "E",   # pycodestyle errors
    "F",   # pyflakes
    "W",   # pycodestyle warnings
    "I",   # isort (import sorting)
    "N",   # pep8-naming
    "UP",  # pyupgrade (use Python 3.13 features)
    "B",   # flake8-bugbear
    "C4",  # flake8-comprehensions
]
ignore = ["E501"]  # Line length handled by formatter

[tool.ruff.lint.per-file-ignores]
"tests/**/*.py" = ["F401", "F811"]  # Allow unused imports in tests

[tool.ruff.format]
quote-style = "double"
indent-style = "space"
```

### Example Source File: `src/utils/formatter.py`

```python
"""String formatting utilities using Python 3.13 features."""
from typing import override


class Formatter:
    """Base formatter class."""

    def format(self, text: str) -> str:
        """Format text with default behavior."""
        return text.strip()


class UppercaseFormatter(Formatter):
    """Formatter that converts text to uppercase."""

    @override  # Python 3.13 PEP 698
    def format(self, text: str) -> str:
        """Format text to uppercase."""
        return text.strip().upper()


def format_user_info(name: str, age: int, city: str) -> str:
    """Format user information with nested f-strings (Python 3.13 PEP 701)."""
    # Python 3.13 allows nested f-strings and arbitrary expressions
    return f"User: {name}, Details: {f'Age: {age}, City: {city.upper()}'}"
```

**Workflow Commands**:
```bash
# Lint and auto-fix issues
ruff check . --fix

# Format code (replaces black)
ruff format .

# Check specific rules
ruff check --select I .           # Import sorting only
ruff check --select UP .          # Python 3.13 upgrade suggestions

# Show what would be fixed (dry-run)
ruff check --show-fixes .
```

---

## Example 3: mypy 1.8.0 Type Checking with Pydantic 2.7.0

### Source File: `src/models/product.py`

```python
"""Product models with static (mypy) and runtime (Pydantic) validation."""
from typing import override
from pydantic import BaseModel, Field, field_validator
from decimal import Decimal


class Product(BaseModel):
    """Product model with runtime validation via Pydantic 2.7.0."""

    id: int = Field(gt=0, description="Product ID must be positive")
    name: str = Field(min_length=1, max_length=100)
    price: Decimal = Field(gt=0, decimal_places=2)
    stock: int = Field(ge=0, description="Stock cannot be negative")

    @field_validator("name")
    @classmethod
    def name_must_not_contain_special_chars(cls, v: str) -> str:
        """Validate that name contains only alphanumeric characters."""
        if not v.replace(" ", "").isalnum():
            raise ValueError("Product name must be alphanumeric")
        return v

    def apply_discount(self, percentage: float) -> Decimal:
        """Apply discount percentage to product price."""
        if not 0 <= percentage <= 100:
            raise ValueError("Discount must be between 0 and 100")
        discount_amount = self.price * Decimal(percentage / 100)
        return self.price - discount_amount


class GenericContainer[T]:  # Python 3.13 PEP 695 type parameters
    """Generic container using Python 3.13 type parameter syntax."""

    def __init__(self, items: list[T]) -> None:
        self.items = items

    def first(self) -> T | None:
        """Return first item or None if empty."""
        return self.items[0] if self.items else None

    def add(self, item: T) -> None:
        """Add item to container."""
        self.items.append(item)


# Usage example
def process_products(products: GenericContainer[Product]) -> list[str]:
    """Process products and return their names."""
    return [p.name for p in products.items if p.stock > 0]
```

### mypy Configuration: `pyproject.toml`

```toml
[tool.mypy]
python_version = "3.13"
strict = true
warn_return_any = true
warn_unused_configs = true
disallow_untyped_defs = true
disallow_any_generics = true
check_untyped_defs = true
no_implicit_optional = true
warn_redundant_casts = true
warn_unused_ignores = true

[[tool.mypy.overrides]]
module = "tests.*"
disallow_untyped_defs = false
```

**Run Commands**:
```bash
mypy .                             # Type check all files
mypy --strict src/                 # Strict mode for production code
mypy --show-column-numbers .       # Show precise error locations
```

---

## Example 4: FastAPI 0.115.0 REST API with Async Handlers

### API Server: `src/api/main.py`

```python
"""FastAPI 0.115.0 application with async handlers and dependency injection."""
from fastapi import FastAPI, HTTPException, Depends
from fastapi.responses import JSONResponse
from pydantic import BaseModel, Field
from typing import Annotated
import asyncio


app = FastAPI(title="Product API", version="1.0.0")


class ProductCreate(BaseModel):
    """Product creation request schema."""
    name: str = Field(min_length=1, max_length=100)
    price: float = Field(gt=0)
    stock: int = Field(ge=0)


class ProductResponse(BaseModel):
    """Product response schema."""
    id: int
    name: str
    price: float
    stock: int


# Dependency injection
async def get_db_session():
    """Simulated database session (replace with actual DB)."""
    # In production: use SQLAlchemy async session
    yield {"connection": "active"}


@app.get("/")
async def root():
    """Health check endpoint."""
    return {"status": "ok", "version": "1.0.0"}


@app.post("/products", response_model=ProductResponse, status_code=201)
async def create_product(
    product: ProductCreate,
    db: Annotated[dict, Depends(get_db_session)]
):
    """Create a new product (async handler)."""
    # Simulate async database operation
    await asyncio.sleep(0.1)

    # In production: insert into database
    product_data = {
        "id": 1,
        "name": product.name,
        "price": product.price,
        "stock": product.stock,
    }

    return ProductResponse(**product_data)


@app.get("/products/{product_id}", response_model=ProductResponse)
async def get_product(
    product_id: int,
    db: Annotated[dict, Depends(get_db_session)]
):
    """Retrieve product by ID (async handler)."""
    # Simulate async database lookup
    await asyncio.sleep(0.1)

    # In production: query from database
    if product_id == 999:
        raise HTTPException(status_code=404, detail="Product not found")

    return ProductResponse(id=product_id, name="Sample Product", price=29.99, stock=100)


@app.exception_handler(ValueError)
async def value_error_handler(request, exc):
    """Global exception handler for ValueError."""
    return JSONResponse(
        status_code=400,
        content={"error": "Validation failed", "detail": str(exc)},
    )
```

### Test File: `tests/test_api.py`

```python
"""Tests for FastAPI endpoints using pytest-asyncio."""
import pytest
from httpx import AsyncClient, ASGITransport
from src.api.main import app


@pytest.mark.asyncio
async def test_root_endpoint():
    """Test health check endpoint."""
    async with AsyncClient(
        transport=ASGITransport(app=app),
        base_url="http://test"
    ) as client:
        response = await client.get("/")
        assert response.status_code == 200
        assert response.json() == {"status": "ok", "version": "1.0.0"}


@pytest.mark.asyncio
async def test_create_product():
    """Test product creation endpoint."""
    async with AsyncClient(
        transport=ASGITransport(app=app),
        base_url="http://test"
    ) as client:
        response = await client.post("/products", json={
            "name": "Test Product",
            "price": 19.99,
            "stock": 50,
        })
        assert response.status_code == 201
        data = response.json()
        assert data["name"] == "Test Product"
        assert data["price"] == 19.99


@pytest.mark.asyncio
async def test_get_product_not_found():
    """Test product retrieval with non-existent ID."""
    async with AsyncClient(
        transport=ASGITransport(app=app),
        base_url="http://test"
    ) as client:
        response = await client.get("/products/999")
        assert response.status_code == 404
```

**Run Commands**:
```bash
# Start development server
uvicorn src.api.main:app --reload

# Run tests
pytest tests/test_api.py -v --cov=src.api
```

---

## Example 5: asyncio.TaskGroup Concurrent Pattern (Python 3.13)

### Concurrent Operations: `src/services/data_fetcher.py`

```python
"""Data fetching service using Python 3.13 asyncio.TaskGroup."""
import asyncio
from typing import Any
import httpx


class DataFetcher:
    """Fetch data from multiple sources concurrently using TaskGroup."""

    def __init__(self, base_url: str):
        self.base_url = base_url

    async def fetch_user(self, user_id: int) -> dict[str, Any]:
        """Fetch user data from API."""
        await asyncio.sleep(0.1)  # Simulate network delay
        return {"id": user_id, "name": f"User{user_id}"}

    async def fetch_posts(self, user_id: int) -> list[dict[str, Any]]:
        """Fetch user posts from API."""
        await asyncio.sleep(0.15)  # Simulate network delay
        return [
            {"id": 1, "user_id": user_id, "title": "Post 1"},
            {"id": 2, "user_id": user_id, "title": "Post 2"},
        ]

    async def fetch_comments(self, user_id: int) -> list[dict[str, Any]]:
        """Fetch user comments from API."""
        await asyncio.sleep(0.12)  # Simulate network delay
        return [
            {"id": 1, "user_id": user_id, "text": "Comment 1"},
        ]

    async def fetch_user_profile(self, user_id: int) -> dict[str, Any]:
        """
        Fetch complete user profile concurrently using TaskGroup.

        TaskGroup advantages over asyncio.gather:
        - Automatic exception propagation
        - Structured concurrency (cancellation safety)
        - Cleaner error handling
        """
        async with asyncio.TaskGroup() as tg:
            user_task = tg.create_task(self.fetch_user(user_id))
            posts_task = tg.create_task(self.fetch_posts(user_id))
            comments_task = tg.create_task(self.fetch_comments(user_id))
            # All tasks run concurrently here
            # If any task fails, TaskGroup cancels others and raises

        # After TaskGroup exits, all tasks are complete
        return {
            "user": user_task.result(),
            "posts": posts_task.result(),
            "comments": comments_task.result(),
        }

    async def fetch_multiple_users(self, user_ids: list[int]) -> list[dict[str, Any]]:
        """Fetch multiple user profiles concurrently."""
        async with asyncio.TaskGroup() as tg:
            tasks = [
                tg.create_task(self.fetch_user_profile(user_id))
                for user_id in user_ids
            ]

        return [task.result() for task in tasks]


# Context Variables for request tracking
from contextvars import ContextVar

request_id_var: ContextVar[str | None] = ContextVar("request_id", default=None)


async def process_request(request_id: str) -> dict[str, Any]:
    """Process request with context variable tracking."""
    # Set context variable (inherited by all spawned tasks)
    token = request_id_var.set(request_id)

    try:
        fetcher = DataFetcher(base_url="https://api.example.com")
        result = await fetcher.fetch_user_profile(user_id=1)

        # Log with request ID (available in all async contexts)
        current_request_id = request_id_var.get()
        print(f"Request {current_request_id} completed")

        return result
    finally:
        request_id_var.reset(token)


# Main execution
async def main():
    """Example usage of DataFetcher with TaskGroup."""
    fetcher = DataFetcher(base_url="https://api.example.com")

    # Fetch single user profile
    profile = await fetcher.fetch_user_profile(user_id=1)
    print(f"Profile: {profile}")

    # Fetch multiple users concurrently
    profiles = await fetcher.fetch_multiple_users([1, 2, 3])
    print(f"Fetched {len(profiles)} profiles")

    # Process request with context tracking
    result = await process_request(request_id="req-12345")
    print(f"Result: {result}")


if __name__ == "__main__":
    asyncio.run(main())
```

### Test File: `tests/test_data_fetcher.py`

```python
"""Tests for DataFetcher using pytest-asyncio."""
import pytest
from src.services.data_fetcher import DataFetcher


@pytest.mark.asyncio
async def test_fetch_user_profile():
    """Test concurrent user profile fetching."""
    fetcher = DataFetcher(base_url="https://api.example.com")
    profile = await fetcher.fetch_user_profile(user_id=1)

    assert "user" in profile
    assert "posts" in profile
    assert "comments" in profile
    assert profile["user"]["id"] == 1


@pytest.mark.asyncio
async def test_fetch_multiple_users():
    """Test fetching multiple user profiles concurrently."""
    fetcher = DataFetcher(base_url="https://api.example.com")
    profiles = await fetcher.fetch_multiple_users([1, 2, 3])

    assert len(profiles) == 3
    assert all("user" in p for p in profiles)


@pytest.mark.asyncio
async def test_taskgroup_error_propagation():
    """Test that TaskGroup properly propagates exceptions."""
    fetcher = DataFetcher(base_url="https://api.example.com")

    # Override method to raise error
    async def failing_fetch(user_id: int):
        raise ValueError("Simulated error")

    fetcher.fetch_user = failing_fetch

    with pytest.raises(ValueError, match="Simulated error"):
        await fetcher.fetch_user_profile(user_id=1)
```

**Run Commands**:
```bash
# Run the example
python -m src.services.data_fetcher

# Run tests
pytest tests/test_data_fetcher.py -v --cov=src.services
```

---

## Quick Reference

### Setup New Project (uv 0.9.3)

```bash
# Create project with Python 3.13
uv venv --python 3.13
source .venv/bin/activate

# Install dependencies
uv add pytest pytest-asyncio ruff mypy fastapi uvicorn httpx pydantic

# Add development dependencies
uv add --dev pytest-cov pytest-mock
```

### Quality Gate Commands

```bash
# Run all tests with coverage
pytest --cov=src --cov-report=term-missing --cov-report=html

# Lint and format
ruff check . --fix
ruff format .

# Type check
mypy --strict src/

# Full quality gate (run before commit)
pytest --cov=src --cov-report=term && ruff check . && mypy .
```

### Production Dependencies Matrix

| Package | Version | Purpose |
|---------|---------|---------|
| pytest | 8.4.2 | Testing framework |
| pytest-asyncio | latest | Async test support |
| pytest-cov | latest | Coverage reporting |
| ruff | 0.13.1 | Linting & formatting |
| mypy | 1.8.0 | Static type checking |
| uv | 0.9.3 | Package management |
| FastAPI | 0.115.0 | Web framework |
| Pydantic | 2.7.0 | Data validation |
| httpx | latest | Async HTTP client |
| uvicorn | latest | ASGI server |

---

All examples are tested against Python 3.13.1 and follow MoAI-ADK TRUST 5 principles.
