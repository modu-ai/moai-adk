---
title: FastAPI 0.120 Async Patterns & Best Practices
description: Master FastAPI async patterns, validation, middleware, and production deployment strategies
freedom_level: high
tier: language
updated: 2025-10-31
---

# FastAPI 0.120 Async Patterns & Best Practices

## Overview

FastAPI is built on async I/O for high-performance APIs. This skill covers async patterns, Pydantic validation, middleware design, dependency injection, and production optimization strategies for FastAPI 0.120+.

## Key Patterns

### 1. Async vs Sync Route Decision Matrix

**Pattern**: Understand when to use `async def` vs `def` for optimal performance.

```python
from fastapi import FastAPI
import asyncio
import time

app = FastAPI()

# ✅ GOOD: Async route with async I/O operations
@app.get("/async-users")
async def get_users_async():
    # Non-blocking database call
    users = await db.fetch_all("SELECT * FROM users")
    return users

# ✅ GOOD: Sync route with blocking operations
@app.get("/sync-report")
def generate_report():
    # CPU-bound operation - FastAPI runs in threadpool
    report = heavy_computation()
    return report

# ❌ BAD: Async route with blocking operations
@app.get("/bad-async")
async def bad_async_route():
    # Blocks event loop! Should be sync or use run_in_executor
    time.sleep(5)
    return {"status": "done"}

# ✅ FIXED: Offload blocking code to thread pool
@app.get("/fixed-async")
async def fixed_async_route():
    loop = asyncio.get_event_loop()
    result = await loop.run_in_executor(None, blocking_function)
    return result

def blocking_function():
    time.sleep(5)
    return {"status": "done"}
```

**Decision Rules**:
- **async def**: Database queries (with async driver), HTTP requests, file I/O (with aiofiles)
- **def**: CPU-intensive tasks (ML inference, image processing, data transformations)

### 2. Dependency Injection for Database Sessions

**Pattern**: Use FastAPI dependencies for database session management.

```python
from fastapi import FastAPI, Depends
from sqlalchemy.ext.asyncio import AsyncSession, create_async_engine
from sqlalchemy.orm import sessionmaker

DATABASE_URL = "postgresql+asyncpg://user:pass@localhost/db"

engine = create_async_engine(DATABASE_URL, echo=True)
AsyncSessionLocal = sessionmaker(
    engine, class_=AsyncSession, expire_on_commit=False
)

async def get_db() -> AsyncSession:
    async with AsyncSessionLocal() as session:
        try:
            yield session
            await session.commit()
        except Exception:
            await session.rollback()
            raise
        finally:
            await session.close()

@app.get("/users/{user_id}")
async def get_user(
    user_id: int,
    db: AsyncSession = Depends(get_db)
):
    result = await db.execute(
        select(User).where(User.id == user_id)
    )
    user = result.scalar_one_or_none()
    if not user:
        raise HTTPException(status_code=404, detail="User not found")
    return user
```

**Benefits**: Automatic session cleanup, transaction management, testability via dependency override.

### 3. Pydantic Validation with Custom Validators

**Pattern**: Use Pydantic v2 for request/response validation with custom logic.

```python
from pydantic import BaseModel, Field, field_validator, model_validator
from typing import Optional
from datetime import datetime

class UserCreate(BaseModel):
    email: str = Field(..., pattern=r"^[\w\.-]+@[\w\.-]+\.\w+$")
    password: str = Field(..., min_length=8)
    age: int = Field(..., ge=18, le=120)
    phone: Optional[str] = None
    
    @field_validator('password')
    @classmethod
    def validate_password_strength(cls, v: str) -> str:
        if not any(c.isupper() for c in v):
            raise ValueError('Password must contain uppercase letter')
        if not any(c.isdigit() for c in v):
            raise ValueError('Password must contain digit')
        return v
    
    @model_validator(mode='after')
    def validate_phone_format(self):
        if self.phone and not self.phone.startswith('+'):
            raise ValueError('Phone must start with country code (+)')
        return self

class UserResponse(BaseModel):
    id: int
    email: str
    created_at: datetime
    
    model_config = {
        "from_attributes": True  # Pydantic v2 syntax (was orm_mode)
    }

@app.post("/users", response_model=UserResponse)
async def create_user(
    user: UserCreate,
    db: AsyncSession = Depends(get_db)
):
    # Validation already done by Pydantic
    db_user = User(**user.model_dump(exclude={'password'}))
    db_user.password_hash = hash_password(user.password)
    db.add(db_user)
    await db.commit()
    await db.refresh(db_user)
    return db_user
```

### 4. Middleware for Logging & Performance Monitoring

**Pattern**: Class-based middleware for stateful operations.

```python
from fastapi import FastAPI, Request
from starlette.middleware.base import BaseHTTPMiddleware
import time
import logging

logger = logging.getLogger(__name__)

class RequestLoggingMiddleware(BaseHTTPMiddleware):
    async def dispatch(self, request: Request, call_next):
        request_id = request.headers.get("X-Request-ID", "N/A")
        start_time = time.time()
        
        # Log request
        logger.info(f"Request {request_id}: {request.method} {request.url.path}")
        
        try:
            response = await call_next(request)
            process_time = time.time() - start_time
            
            # Add performance headers
            response.headers["X-Process-Time"] = str(process_time)
            response.headers["X-Request-ID"] = request_id
            
            # Log response
            logger.info(
                f"Response {request_id}: status={response.status_code} "
                f"time={process_time:.3f}s"
            )
            
            return response
        except Exception as e:
            logger.error(f"Error {request_id}: {str(e)}")
            raise

app = FastAPI()
app.add_middleware(RequestLoggingMiddleware)

# Selective middleware for specific routes
from starlette.middleware.base import BaseHTTPMiddleware

class RateLimitMiddleware(BaseHTTPMiddleware):
    async def dispatch(self, request: Request, call_next):
        if request.url.path.startswith("/api/public"):
            # Apply rate limiting
            if not await check_rate_limit(request.client.host):
                return JSONResponse(
                    status_code=429,
                    content={"error": "Too many requests"}
                )
        return await call_next(request)
```

### 5. Background Tasks for Async Operations

**Pattern**: Use BackgroundTasks for non-blocking post-processing.

```python
from fastapi import BackgroundTasks
import asyncio

async def send_welcome_email(email: str):
    await asyncio.sleep(2)  # Simulate email sending
    print(f"Email sent to {email}")

async def update_analytics(user_id: int):
    # Non-blocking analytics update
    await db.execute(
        "INSERT INTO analytics (user_id, event) VALUES ($1, $2)",
        user_id, "signup"
    )

@app.post("/register")
async def register_user(
    user: UserCreate,
    background_tasks: BackgroundTasks,
    db: AsyncSession = Depends(get_db)
):
    # Create user synchronously
    db_user = User(**user.model_dump())
    db.add(db_user)
    await db.commit()
    
    # Queue background tasks (don't wait)
    background_tasks.add_task(send_welcome_email, user.email)
    background_tasks.add_task(update_analytics, db_user.id)
    
    return {"id": db_user.id, "status": "registered"}
```

**Use Cases**: Email notifications, cache warming, analytics events, file cleanup.

### 6. Exception Handlers for Consistent Error Responses

**Pattern**: Global exception handlers for standardized error format.

```python
from fastapi import FastAPI, HTTPException, Request
from fastapi.responses import JSONResponse
from sqlalchemy.exc import IntegrityError
from pydantic import ValidationError

app = FastAPI()

@app.exception_handler(HTTPException)
async def http_exception_handler(request: Request, exc: HTTPException):
    return JSONResponse(
        status_code=exc.status_code,
        content={
            "error": exc.detail,
            "status_code": exc.status_code,
            "path": str(request.url)
        }
    )

@app.exception_handler(ValidationError)
async def validation_exception_handler(request: Request, exc: ValidationError):
    return JSONResponse(
        status_code=422,
        content={
            "error": "Validation failed",
            "details": exc.errors()
        }
    )

@app.exception_handler(IntegrityError)
async def database_exception_handler(request: Request, exc: IntegrityError):
    return JSONResponse(
        status_code=409,
        content={"error": "Database conflict (duplicate key or constraint)"}
    )
```

### 7. Production Optimization: Connection Pooling

**Pattern**: Configure optimal database connection pooling.

```python
from sqlalchemy.ext.asyncio import create_async_engine

engine = create_async_engine(
    DATABASE_URL,
    echo=False,  # Disable SQL logging in production
    pool_size=20,  # Max 20 connections
    max_overflow=10,  # Allow 10 overflow connections
    pool_pre_ping=True,  # Verify connection before use
    pool_recycle=3600,  # Recycle connections every hour
)

# Best practices for production
@app.on_event("startup")
async def startup():
    # Warm up connection pool
    async with engine.begin() as conn:
        await conn.execute("SELECT 1")

@app.on_event("shutdown")
async def shutdown():
    await engine.dispose()
```

## Checklist

- [ ] Audit routes: use `async def` only for async I/O, `def` for CPU-bound tasks
- [ ] Convert `time.sleep()` to `await asyncio.sleep()` in async functions
- [ ] Implement dependency injection for database sessions
- [ ] Add Pydantic validators for business logic validation
- [ ] Configure connection pooling for database (pool_size, max_overflow)
- [ ] Use BackgroundTasks for non-blocking operations (emails, analytics)
- [ ] Add global exception handlers for consistent error responses
- [ ] Enable CORS middleware for frontend integration
- [ ] Add request logging middleware with performance metrics

## Resources

- **FastAPI Official Docs**: https://fastapi.tiangolo.com/
- **Async Best Practices**: https://github.com/zhanymkanov/fastapi-best-practices
- **Advanced Middleware Guide**: https://medium.com/@rameshkannanyt0078/advanced-fastapi-middleware-beyond-the-basics-e34245a2254b
- **Security Best Practices (2025)**: https://toxigon.com/python-fastapi-security-best-practices-2025
- **Performance Tuning**: https://loadforge.com/guides/fastapi-performance-tuning-tricks-to-enhance-speed-and-scalability

---

**Version**: 1.0.0  
**Last Updated**: 2025-10-31  
**Model Recommendation**: Sonnet (deep reasoning for async patterns and architecture)
