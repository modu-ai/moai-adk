---
name: moai-domain-backend
type: domain
description: Backend architecture patterns and best practices
tier: domain
---

# Backend Architecture

## Quick Start (30 seconds)
Backend systems must handle concurrent requests, manage data consistently, ensure security, and scale horizontally. Success requires understanding request/response cycles, proper error handling, caching strategies, and monitoring.

## Core Patterns

### Pattern 1: Request/Response Lifecycle
Structure your application to handle requests consistently.

```python
from fastapi import FastAPI, Request, Response, HTTPException
from starlette.middleware.cors import CORSMiddleware
import logging

app = FastAPI()
logger = logging.getLogger(__name__)

# CORS configuration for cross-origin requests
app.add_middleware(
    CORSMiddleware,
    allow_origins=["https://example.com"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# Request logging middleware
@app.middleware("http")
async def log_requests(request: Request, call_next):
    """Log all incoming requests"""
    logger.info(f"→ {request.method} {request.url.path}")
    response = await call_next(request)
    logger.info(f"← {response.status_code}")
    return response

# Structured request handling
@app.post("/api/users/")
async def create_user(request: Request):
    """Create user with comprehensive error handling"""
    try:
        data = await request.json()

        # Validate input
        if not data.get("email"):
            raise HTTPException(status_code=400, detail="Email required")

        # Process request
        user = await create_user_in_db(data)

        # Return structured response
        return {
            "success": True,
            "data": user,
            "status_code": 201
        }
    except ValueError as e:
        logger.error(f"Validation error: {e}")
        raise HTTPException(status_code=400, detail=str(e))
    except Exception as e:
        logger.exception(f"Unexpected error: {e}")
        raise HTTPException(status_code=500, detail="Internal server error")
```

### Pattern 2: Error Handling and Resilience
Design systems that gracefully handle failures.

```python
from typing import Optional
import asyncio

class RetryPolicy:
    """Retry failed operations with exponential backoff"""
    def __init__(self, max_retries: int = 3, base_delay: float = 1.0):
        self.max_retries = max_retries
        self.base_delay = base_delay

    async def execute(self, func, *args, **kwargs):
        """Execute function with retry logic"""
        for attempt in range(self.max_retries):
            try:
                return await func(*args, **kwargs)
            except Exception as e:
                if attempt == self.max_retries - 1:
                    raise
                delay = self.base_delay * (2 ** attempt)
                await asyncio.sleep(delay)

class CircuitBreaker:
    """Prevent cascading failures by stopping requests"""
    def __init__(self, failure_threshold: int = 5, timeout: int = 60):
        self.failure_count = 0
        self.failure_threshold = failure_threshold
        self.timeout = timeout
        self.last_failure_time = None
        self.is_open = False

    async def call(self, func, *args, **kwargs):
        """Call function with circuit breaker protection"""
        if self.is_open:
            raise Exception("Circuit breaker is open")

        try:
            result = await func(*args, **kwargs)
            self.failure_count = 0
            return result
        except Exception as e:
            self.failure_count += 1
            if self.failure_count >= self.failure_threshold:
                self.is_open = True
            raise
```

### Pattern 3: Caching Strategies
Reduce load and improve response times with caching.

```python
from functools import lru_cache
import asyncio
from datetime import datetime, timedelta
import json

# Memory caching
@lru_cache(maxsize=128)
def get_user_by_id(user_id: int):
    """Simple in-memory cache (good for small datasets)"""
    return query_database(f"SELECT * FROM users WHERE id = {user_id}")

# Time-based caching
class CacheWithTTL:
    def __init__(self, ttl_seconds: int = 300):
        self.cache = {}
        self.ttl = ttl_seconds

    def get(self, key: str):
        """Get value if not expired"""
        if key not in self.cache:
            return None

        value, timestamp = self.cache[key]
        if datetime.now() - timestamp > timedelta(seconds=self.ttl):
            del self.cache[key]
            return None
        return value

    def set(self, key: str, value):
        """Set value with current timestamp"""
        self.cache[key] = (value, datetime.now())

# Redis caching (distributed)
import redis
redis_client = redis.Redis(host='localhost', port=6379, db=0)

async def get_user_cached(user_id: int):
    """Multi-level cache: Redis → Database"""
    # Check Redis first
    cached = redis_client.get(f"user:{user_id}")
    if cached:
        return json.loads(cached)

    # Fetch from database
    user = await query_database(user_id)

    # Store in Redis for 1 hour
    redis_client.setex(f"user:{user_id}", 3600, json.dumps(user))
    return user
```

## Progressive Disclosure

### Scalability Patterns

```python
# Load balancing: Run multiple processes
# Use gunicorn: gunicorn -w 4 -k uvicorn.workers.UvicornWorker app:app

# Database connection pooling
from sqlalchemy import create_engine
from sqlalchemy.pool import QueuePool

engine = create_engine(
    "postgresql://user:password@localhost/db",
    poolclass=QueuePool,
    pool_size=20,  # Number of connections
    max_overflow=0,  # Max extra connections
)

# Asynchronous database queries
from sqlalchemy.ext.asyncio import create_async_engine, AsyncSession

async_engine = create_async_engine(
    "postgresql+asyncpg://user:password@localhost/db",
    echo=True,
)

async def fetch_users_efficiently(skip: int = 0, limit: int = 10):
    """Fetch paginated results efficiently"""
    async with AsyncSession(async_engine) as session:
        result = await session.execute(
            select(User).offset(skip).limit(limit)
        )
        return result.scalars().all()
```

### Monitoring and Logging

```python
import logging
import structlog
from prometheus_client import Counter, Histogram

# Structured logging
structlog.configure(
    processors=[structlog.processors.JSONRenderer()]
)
logger = structlog.get_logger()

# Metrics
request_count = Counter('http_requests_total', 'Total HTTP requests')
request_duration = Histogram('http_request_duration_seconds', 'HTTP request duration')

@app.get("/health")
async def health_check():
    """Health check endpoint for load balancers"""
    return {
        "status": "healthy",
        "timestamp": datetime.now().isoformat(),
        "version": "1.0.0"
    }

@app.middleware("http")
async def metrics_middleware(request: Request, call_next):
    """Collect request metrics"""
    request_count.inc()
    start_time = time.time()
    response = await call_next(request)
    request_duration.observe(time.time() - start_time)
    return response
```

### Security Best Practices

```python
from fastapi import Security, HTTPException, status
from fastapi.security import HTTPBearer, HTTPAuthorizationCredentials
import jwt

security = HTTPBearer()

async def verify_token(credentials: HTTPAuthorizationCredentials = Security(security)):
    """Verify JWT token"""
    try:
        token = credentials.credentials
        payload = jwt.decode(token, SECRET_KEY, algorithms=["HS256"])
        return payload
    except jwt.ExpiredSignatureError:
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail="Token expired"
        )
    except jwt.InvalidTokenError:
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail="Invalid token"
        )

@app.get("/protected")
async def protected_route(user = Depends(verify_token)):
    """Protected endpoint requiring authentication"""
    return {"message": f"Hello {user['username']}"}
```

## Works Well With
- **moai-lang-fastapi-patterns** - Web framework implementation
- **moai-lang-python** - Core language features
- **moai-domain-database** - Data storage strategies

## Common Pitfalls

1. **No input validation**
   - ❌ Wrong: Trusting user input directly
   - ✅ Right: Always validate and sanitize user input with Pydantic

2. **Hardcoded configuration**
   - ❌ Wrong: Database credentials in source code
   - ✅ Right: Use environment variables and .env files

3. **Poor error messages**
   - ❌ Wrong: Exposing stack traces and internal details to users
   - ✅ Right: Log details internally, return generic messages to users

4. **No monitoring**
   - ❌ Wrong: Waiting for users to report issues
   - ✅ Right: Add logging, metrics, and health checks from the start

## References
- [12 Factor App](https://12factor.net/)
- [Building Scalable Web Applications](https://developer.mozilla.org/en-US/docs/Learn/Server-side)
- [RESTful API Best Practices](https://restfulapi.net/)
- [API Security Checklist](https://github.com/shieldfy/API-Security-Checklist)
