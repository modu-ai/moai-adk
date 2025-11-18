---
name: moai-domain-backend
version: 4.0.0
created: 2025-11-18
updated: '2025-11-18'
status: stable
stability: stable
---


# Enterprise Backend Architecture - 

**Modern async patterns, microservices, API design, and production deployment**

> **Primary Agent**: backend-expert
> **Stack**: FastAPI 0.118+, Django 5.2+, async Python, Kubernetes 1.30+, PostgreSQL 17+
> **Keywords**: backend, api, microservices, database, async, fastapi, django, kubernetes

## Level 1: Quick Reference

### Core Technology Stack (2025)

**Frameworks**: FastAPI 0.118+, Django 5.2+, Django Ninja 1.4+
**Async Runtime**: asyncio, uvloop, asyncpg, motor (async MongoDB)
**Databases**: PostgreSQL 17+, MongoDB 8+, Redis 8+
**Deployment**: Kubernetes 1.30+, Docker, Istio 1.24+
**Observability**: OpenTelemetry 1.28+, Prometheus 3.0+, Jaeger 1.55+

### When to Use This Skill

- ✅ Building high-performance REST/GraphQL APIs
- ✅ Designing microservices with service mesh
- ✅ Implementing async Python backends
- ✅ Optimizing database queries and pooling
- ✅ Deploying on Kubernetes
- ✅ Authentication, rate limiting, observability
- ✅ Cloud-native migration strategies

### Modern FastAPI Backend

```python
from fastapi import FastAPI, Depends, HTTPException, BackgroundTasks
from sqlalchemy.ext.asyncio import AsyncSession, create_async_engine
from contextlib import asynccontextmanager

@asynccontextmanager
async def lifespan(app: FastAPI):
    # Initialize database pool
    app.state.db_engine = create_async_engine(
        "postgresql+asyncpg://user:pass@localhost/db",
        pool_size=20, max_overflow=10
    )
    yield
    await app.state.db_engine.dispose()

app = FastAPI(lifespan=lifespan)

# Dependency injection
async def get_db() -> AsyncSession:
    async with AsyncSession(app.state.db_engine) as session:
        try:
            yield session
        finally:
            await session.close()

# Async endpoint with background tasks
@app.get("/users/{user_id}")
async def get_user(
    user_id: int,
    db: AsyncSession = Depends(get_db),
    background_tasks: BackgroundTasks = None
):
    user = await db.get(User, user_id)
    if not user:
        raise HTTPException(status_code=404, detail="User not found")

    background_tasks.add_task(log_user_access, user_id)
    return user
```

## Level 2: Practical Implementation

### Async Database with Connection Pooling

```python
from sqlalchemy.ext.asyncio import (
    AsyncSession, create_async_engine, async_sessionmaker
)
from sqlalchemy.pool import QueuePool

# Production async engine
engine = create_async_engine(
    "postgresql+asyncpg://user:pass@localhost/db",
    pool_size=20, max_overflow=10,
    pool_timeout=30, pool_recycle=3600,
    pool_pre_ping=True, poolclass=QueuePool
)

AsyncSessionLocal = async_sessionmaker(
    engine, class_=AsyncSession,
    expire_on_commit=False, autoflush=False
)

# Dependency injection
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

# Usage
@app.get("/users/{user_id}")
async def get_user(user_id: int, db: AsyncSession = Depends(get_db)):
    result = await db.execute(select(User).where(User.id == user_id))
    user = result.scalar_one_or_none()
    if not user:
        raise HTTPException(status_code=404)
    return user
```

**Benefits**: 3,000+ req/sec, non-blocking operations, graceful error handling

### Background Tasks for Long Operations

```python
from fastapi import BackgroundTasks
import asyncio

async def send_welcome_email(user_email: str):
    await asyncio.sleep(2)  # Simulate email sending
    print(f"Email sent to {user_email}")

@app.post("/users/", status_code=201)
async def create_user(
    user: UserCreate,
    background_tasks: BackgroundTasks,
    db: AsyncSession = Depends(get_db)
):
    # Create user (fast)
    new_user = User(**user.dict())
    db.add(new_user)
    await db.commit()

    # Schedule background tasks (non-blocking)
    background_tasks.add_task(send_welcome_email, user.email)

    return {"id": new_user.id, "message": "User created"}
```

**Use Cases**: Email notifications, file processing, data exports, webhooks

### Dependency Injection for Clean Architecture

```python
from fastapi import Depends, HTTPException, Header
from typing import Annotated

# JWT authentication dependency
async def get_current_user(
    authorization: Annotated[str, Header()] = None,
    db: AsyncSession = Depends(get_db)
) -> User:
    if not authorization or not authorization.startswith("Bearer "):
        raise HTTPException(status_code=401, detail="Invalid token")

    token = authorization.split(" ")[1]
    payload = decode_jwt(token)

    user = await db.get(User, payload["user_id"])
    if not user:
        raise HTTPException(status_code=401, detail="User not found")
    return user

# Admin-only dependency
async def require_admin(
    current_user: Annotated[User, Depends(get_current_user)]
) -> User:
    if not current_user.is_admin:
        raise HTTPException(status_code=403, detail="Admin access required")
    return current_user

# Usage
@app.get("/admin/users")
async def list_admin_users(
    admin_user: Annotated[User, Depends(require_admin)],
    db: AsyncSession = Depends(get_db)
):
    users = await db.execute(select(User))
    return users.scalars().all()
```

## Level 3: Advanced Integration

### Rate Limiting & Caching

```python
from slowapi import Limiter, _rate_limit_exceeded_handler
from slowapi.util import get_remote_address
from fastapi import Request

# Rate limiting
limiter = Limiter(key_func=get_remote_address)
app.state.limiter = limiter

@app.get("/api/users")
@limiter.limit("100/minute")
async def list_users(request: Request, db: AsyncSession = Depends(get_db)):
    # API with rate limiting
    users = await db.execute(select(User))
    return users.scalars().all()

# Redis caching
import redis.asyncio as redis
from fastapi_cache import FastAPICache, Coder
from fastapi_cache.backends.redis import RedisBackend

@app.post("/compute-heavy")
@cache(expire=60)  # Cache for 60 seconds
async def compute_heavy_operation(data: InputData):
    # Expensive computation cached in Redis
    result = await expensive_calculation(data)
    return result
```

### Microservices with Service Discovery

```python
# Service registration with Consul
import aiohttp
import asyncio

class ServiceRegistry:
    def __init__(self, consul_url: str):
        self.consul_url = consul_url
        self.service_id = f"user-service-{uuid4()}"

    async def register(self):
        async with aiohttp.ClientSession() as session:
            await session.put(
                f"{self.consul_url}/v1/agent/service/register",
                json={
                    "ID": self.service_id,
                    "Name": "user-service",
                    "Address": "user-service",
                    "Port": 8000,
                    "Check": {
                        "HTTP": "http://user-service:8000/health",
                        "Interval": "10s"
                    }
                }
            )

# Service discovery and load balancing
async def call_service(service_name: str, endpoint: str):
    services = await discover_services(service_name)
    selected = random.choice(services)
    url = f"http://{selected['Address']}:{selected['Port']}{endpoint}"

    async with aiohttp.ClientSession() as session:
        async with session.get(url) as response:
            return await response.json()
```

### API Gateway & Service Mesh

```yaml
# Istio Virtual Service
apiVersion: networking.istio.io/v1beta1
kind: VirtualService
metadata:
  name: user-service
spec:
  http:
  - match:
    - uri:
        prefix: "/api/v1/users"
    route:
    - destination:
        host: user-service
        port:
          number: 8000
      weight: 100
    fault:
      delay:
        percentage:
          value: 0.1
        fixedDelay: 5s
```

### OpenTelemetry Observability

```python
from opentelemetry import trace, baggage
from opentelemetry.exporter.jaeger.thrift import JaegerExporter
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.trace.export import BatchSpanProcessor

# Initialize tracing
tracer_provider = TracerProvider()
trace.set_tracer_provider(tracer_provider)

jaeger_exporter = JaegerExporter(
    agent_host_name="jaeger",
    agent_port=6831,
)

span_processor = BatchSpanProcessor(jaeger_exporter)
tracer_provider.add_span_processor(span_processor)

# Custom tracing in endpoints
@app.get("/users/{user_id}")
async def get_user(user_id: int, db: AsyncSession = Depends(get_db)):
    tracer = trace.get_tracer(__name__)

    with tracer.start_as_current_span("get_user") as span:
        span.set_attribute("user.id", user_id)

        with tracer.start_as_current_span("database_query"):
            result = await db.execute(select(User).where(User.id == user_id))
            user = result.scalar_one_or_none()

        if user:
            span.set_attribute("user.found", True)
            return user
        else:
            span.set_attribute("user.found", False)
            raise HTTPException(status_code=404)
```

## Performance Optimization

### Async Patterns Checklist
- ✅ Use async database drivers (asyncpg, motor)
- ✅ Implement connection pooling
- ✅ Use background tasks for I/O operations
- ✅ Cache frequently accessed data
- ✅ Implement rate limiting
- ✅ Use circuit breakers for external services

### Database Optimization
```python
# Query optimization with SQLAlchemy 2.0
from sqlalchemy import select, func, and_

# Efficient pagination
async def get_users_paginated(
    page: int = 1, size: int = 20,
    db: AsyncSession = Depends(get_db)
):
    offset = (page - 1) * size

    result = await db.execute(
        select(User).offset(offset).limit(size)
    )
    users = result.scalars().all()

    # Get total count efficiently
    total_result = await db.execute(select(func.count(User.id)))
    total = total_result.scalar()

    return {
        "users": users,
        "total": total,
        "page": page,
        "size": size
    }

# Batch operations
async def create_users_batch(users_data: List[UserCreate], db: AsyncSession):
    users = [User(**user.dict()) for user in users_data]
    db.add_all(users)
    await db.commit()
    return users
```

## Deployment Patterns

### Kubernetes Deployment
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: user-service
spec:
  replicas: 3
  selector:
    matchLabels:
      app: user-service
  template:
    metadata:
      labels:
        app: user-service
    spec:
      containers:
      - name: api
        image: user-service:latest
        ports:
        - containerPort: 8000
        env:
        - name: DATABASE_URL
          valueFrom:
            secretKeyRef:
              name: db-secret
              key: url
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
```

## Quick Architecture Decision Matrix

| Requirement | Solution | When to Use |
|-------------|----------|------------|
| High concurrency (1000+ req/s) | FastAPI + asyncpg | I/O-heavy workloads |
| Complex business logic | Django 5.2+ | Traditional CRUD apps |
| Microservices | FastAPI + Kubernetes | Distributed systems |
| Simple APIs | FastAPI Minimal | Small services |
| Real-time features | WebSockets + FastAPI | Chat, notifications |
| File uploads | FastAPI + Background Tasks | Media processing |

## Installation Commands

```bash
# Core backend stack
pip install fastapi==0.118.0
pip install uvicorn[standard]
pip install sqlalchemy[asyncio]==2.0.0
pip install asyncpg
pip install pydantic==2.8.0

# Production addons
pip install redis
pip install slowapi  # Rate limiting
pip install fastapi-cache[redis]
pip install python-multipart  # File uploads

# Observability
pip install opentelemetry-api
pip install opentelemetry-sdk
pip install opentelemetry-exporter-jaeger
pip install prometheus-client

# Development
pip install pytest-asyncio
pip install httpx  # Async HTTP client for testing
```

## Best Practices

1. **Async by Default**: Use async/await for I/O operations
2. **Connection Pooling**: Configure appropriate pool sizes
3. **Error Handling**: Implement graceful degradation
4. **Security**: Input validation, rate limiting, authentication
5. **Testing**: Unit and integration tests with pytest-asyncio
6. **Documentation**: OpenAPI/Swagger auto-generation
7. **Monitoring**: Structured logging and metrics
8. **Versioning**: API versioning strategy

---

**Version**: 4.0.0 Enterprise
**Last Updated**: 2025-11-18
**Status**: Production Ready
**Enterprise Grade**: ✅ Full Enterprise Support
