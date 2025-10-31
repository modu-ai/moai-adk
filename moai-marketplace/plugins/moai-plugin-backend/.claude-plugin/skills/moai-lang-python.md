---
name: moai-lang-python
type: language
description: Python best practices, type hints, and modern patterns
tier: language
---

# Python Best Practices

## Quick Start (30 seconds)
Python 3.13+ provides powerful tools for writing safe, maintainable code: type hints for static checking, async/await for concurrent I/O, context managers for resource management, and rich standard library features.

## Core Patterns

### Pattern 1: Type Hints and Static Typing
Use type hints to catch errors early and improve code documentation.

```python
from typing import List, Optional, Dict, Tuple, Callable, TypeVar
from dataclasses import dataclass

# Function type hints
def process_items(items: List[str], threshold: int = 10) -> Dict[str, int]:
    """Process items and return count by type"""
    return {item: len(item) for item in items if len(item) >= threshold}

# Class with type hints
@dataclass
class User:
    id: int
    name: str
    email: str
    is_active: bool = True
    tags: List[str] = None

    def __post_init__(self):
        if self.tags is None:
            self.tags = []

# Generic types
T = TypeVar('T')
def first_or_none(items: List[T]) -> Optional[T]:
    """Get first item or None"""
    return items[0] if items else None

# Callable types
def apply_operation(x: int, y: int, op: Callable[[int, int], int]) -> int:
    return op(x, y)

result = apply_operation(5, 3, lambda a, b: a + b)
```

**Type checking**:
- Run `mypy your_file.py` to catch type errors
- Use `# type: ignore` to suppress specific warnings
- Enable strict mode: `mypy --strict`

### Pattern 2: Async/Await for Concurrent I/O
Write non-blocking code for I/O-bound operations.

```python
import asyncio
from typing import Coroutine, List
import aiohttp

async def fetch_url(session: aiohttp.ClientSession, url: str) -> str:
    """Fetch URL content asynchronously"""
    async with session.get(url) as response:
        return await response.text()

async def fetch_multiple_urls(urls: List[str]) -> List[str]:
    """Fetch multiple URLs concurrently"""
    async with aiohttp.ClientSession() as session:
        tasks = [fetch_url(session, url) for url in urls]
        return await asyncio.gather(*tasks)

# Run async code
urls = ["https://example.com", "https://example.org"]
results = asyncio.run(fetch_multiple_urls(urls))

# Async context manager
class AsyncDatabase:
    async def __aenter__(self):
        await self.connect()
        return self

    async def __aexit__(self, exc_type, exc_val, exc_tb):
        await self.disconnect()

    async def connect(self):
        print("Connecting...")

    async def disconnect(self):
        print("Disconnecting...")

async def main():
    async with AsyncDatabase() as db:
        pass  # db is connected here
```

### Pattern 3: Context Managers for Resource Management
Ensure resources are properly cleaned up using context managers.

```python
from contextlib import contextmanager
from typing import Generator

# Class-based context manager
class FileHandler:
    def __init__(self, filename: str):
        self.filename = filename
        self.file = None

    def __enter__(self):
        self.file = open(self.filename, 'r')
        return self.file

    def __exit__(self, exc_type, exc_val, exc_tb):
        if self.file:
            self.file.close()
        if exc_type is not None:
            print(f"Exception: {exc_val}")
        return False  # Propagate exception

# Function-based context manager
@contextmanager
def database_connection(db_url: str) -> Generator:
    """Context manager for database connections"""
    conn = create_connection(db_url)
    try:
        yield conn
    finally:
        conn.close()

# Usage
with FileHandler("data.txt") as f:
    content = f.read()

with database_connection("postgresql://localhost/mydb") as conn:
    result = conn.execute("SELECT * FROM users")
```

## Progressive Disclosure

### Decorators for Code Reuse

```python
from functools import wraps
from typing import Callable, Any

def timing_decorator(func: Callable) -> Callable:
    """Measure function execution time"""
    @wraps(func)
    def wrapper(*args, **kwargs) -> Any:
        import time
        start = time.time()
        result = func(*args, **kwargs)
        elapsed = time.time() - start
        print(f"{func.__name__} took {elapsed:.3f}s")
        return result
    return wrapper

def retry(max_attempts: int = 3):
    """Retry decorator with exponential backoff"""
    def decorator(func: Callable) -> Callable:
        @wraps(func)
        def wrapper(*args, **kwargs):
            import time
            for attempt in range(max_attempts):
                try:
                    return func(*args, **kwargs)
                except Exception as e:
                    if attempt == max_attempts - 1:
                        raise
                    wait_time = 2 ** attempt
                    print(f"Attempt {attempt + 1} failed, retrying in {wait_time}s...")
                    time.sleep(wait_time)
        return wrapper
    return decorator

@timing_decorator
@retry(max_attempts=3)
def unstable_operation():
    """Operation that might fail"""
    import random
    if random.random() < 0.7:
        raise ValueError("Random failure")
    return "Success"
```

### Generators and Lazy Evaluation

```python
from typing import Generator, Iterator

def read_large_file(filename: str, chunk_size: int = 8192) -> Generator[str, None, None]:
    """Read file line by line without loading entire file"""
    with open(filename) as f:
        while True:
            chunk = f.read(chunk_size)
            if not chunk:
                break
            yield chunk

def range_generator(start: int, end: int) -> Iterator[int]:
    """Lazy range generator"""
    current = start
    while current < end:
        yield current
        current += 1

# Usage
for chunk in read_large_file("huge_file.txt"):
    process_chunk(chunk)  # Process one chunk at a time
```

### Property Decorators

```python
class Temperature:
    def __init__(self, celsius: float = 0):
        self._celsius = celsius

    @property
    def celsius(self) -> float:
        """Get temperature in Celsius"""
        return self._celsius

    @celsius.setter
    def celsius(self, value: float):
        """Set temperature in Celsius"""
        if value < -273.15:
            raise ValueError("Temperature below absolute zero")
        self._celsius = value

    @property
    def fahrenheit(self) -> float:
        """Get temperature in Fahrenheit"""
        return self._celsius * 9/5 + 32

    @fahrenheit.setter
    def fahrenheit(self, value: float):
        """Set temperature in Fahrenheit"""
        self.celsius = (value - 32) * 5/9
```

## Works Well With
- **moai-lang-fastapi-patterns** - Using Python in web framework context
- **moai-domain-backend** - Backend system design principles
- **moai-domain-database** - Database integration patterns

## Common Pitfalls

1. **Mutable default arguments**
   - ❌ Wrong: `def func(items: List = [])` mutates shared list
   - ✅ Right: `def func(items: List = None)` then `items = items or []`

2. **String formatting**
   - ❌ Wrong: `"Hello %s" % name` or `"Hello {}".format(name)`
   - ✅ Right: `f"Hello {name}"` (f-strings)

3. **Exception handling**
   - ❌ Wrong: Bare `except:` catches all exceptions including KeyboardInterrupt
   - ✅ Right: `except ValueError:` catch specific exceptions

4. **Global variables**
   - ❌ Wrong: Using global state across modules
   - ✅ Right: Use dependency injection or class attributes

## References
- [Python 3.13 Documentation](https://docs.python.org/3.13/)
- [Type Hints (PEP 484)](https://www.python.org/dev/peps/pep-0484/)
- [asyncio Documentation](https://docs.python.org/3/library/asyncio.html)
- [Dataclasses Documentation](https://docs.python.org/3/library/dataclasses.html)
- [Python Patterns](https://python-patterns.guide/)
