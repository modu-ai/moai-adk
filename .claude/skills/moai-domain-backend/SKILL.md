---
name: moai-domain-backend
version: 4.0.0
created: 2025-11-12
updated: 2025-11-12
status: active
tier: domain
description: "Enterprise-grade backend architecture expertise with modern async patterns, microservices orchestration, API design best practices, and production-ready deployment strategies. Covers FastAPI, Django, service mesh, database optimization, and cloud-native architectures for 2025. Enhanced with Context7 MCP integration for latest documentation."
allowed-tools: "Read, Bash, WebSearch, WebFetch, mcp__context7__resolve-library-id, mcp__context7__get-library-docs"
primary-agent: "backend-expert"
secondary-agents: [qa-validator, alfred, doc-syncer]
keywords: [backend, api, microservices, database, async, fastapi, django, kubernetes, service-mesh]
tags: [domain-expert]
orchestration:
  can_resume: true
  typical_chain_position: "middle"
  depends_on: []
---

# moai-domain-backend

**Enterprise Backend Architecture Expertise**

> **Primary Agent**: backend-expert  
> **Secondary Agents**: qa-validator, alfred, doc-syncer  
> **Version**: 4.0.0  
> **Keywords**: backend, api, microservices, database, async, fastapi, django, kubernetes, service-mesh

---

## 📖 Progressive Disclosure

### Level 1: Quick Reference (Core Concepts)

**Purpose**: Enterprise-grade backend architecture expertise with modern async patterns, microservices orchestration, API design best practices, and production-ready deployment strategies for 2025.

**When to Use:**
- ✅ Building high-performance REST/GraphQL APIs
- ✅ Designing microservices architecture with service mesh
- ✅ Implementing async Python backends (FastAPI, Django Ninja)
- ✅ Optimizing database queries and connection pooling
- ✅ Deploying scalable backend systems on Kubernetes
- ✅ Implementing authentication, rate limiting, observability
- ✅ Migrating legacy monoliths to cloud-native architecture

**Quick Start Pattern:**

```python
# FastAPI 0.118+ Modern Async Backend
from fastapi import FastAPI, Depends, HTTPException, BackgroundTasks
from sqlalchemy.ext.asyncio import AsyncSession, create_async_engine
from contextlib import asynccontextmanager
from typing import List

# Lifespan context manager (FastAPI 0.100+)
@asynccontextmanager
async def lifespan(app: FastAPI):
    # Startup: Initialize database pool
    app.state.db_engine = create_async_engine(
        "postgresql+asyncpg://user:pass@localhost/db",
        pool_size=20,
        max_overflow=10
    )
    yield
    # Shutdown: Close connections
    await app.state.db_engine.dispose()

app = FastAPI(lifespan=lifespan)

# Dependency injection for database sessions
async def get_db() -> AsyncSession:
    async with AsyncSession(app.state.db_engine) as session:
        try:
            yield session
        finally:
            await session.close()

# Modern endpoint with async patterns
@app.get("/users/{user_id}", response_model=UserOut)
async def get_user(
    user_id: int,
    db: AsyncSession = Depends(get_db),
    background_tasks: BackgroundTasks = None
):
    """Retrieve user with async database query"""
    user = await db.get(User, user_id)
    if not user:
        raise HTTPException(status_code=404, detail="User not found")
    
    # Non-blocking background task
    background_tasks.add_task(log_user_access, user_id)
    return user
```

**Core Technology Stack (2025):**
- **Frameworks**: FastAPI 0.118+, Django 5.2+, Django Ninja 1.4+
- **Async**: asyncio, uvloop, asyncpg, motor (async MongoDB)
- **Databases**: PostgreSQL 17+, MongoDB 8+, Redis 8+
- **Deployment**: Kubernetes 1.30+, Docker, Istio 1.24+
- **Observability**: OpenTelemetry 1.28+, Prometheus 3.0+, Jaeger 1.55+

---

### Level 2: Practical Implementation (Common Patterns)

#### 🔍 Pattern 1: Async Database Access with Connection Pooling

**Problem**: Synchronous database calls block the event loop, limiting concurrency.

**Solution**: Use async database drivers with proper connection pooling.

```python
# Production-Ready Async Database Setup
from sqlalchemy.ext.asyncio import (
    AsyncSession,
    create_async_engine,
    async_sessionmaker
)
from sqlalchemy.pool import NullPool, QueuePool

# PostgreSQL async engine with connection pooling
engine = create_async_engine(
    "postgresql+asyncpg://user:pass@localhost/db",
    echo=False,  # Set True for SQL logging in dev
    pool_size=20,  # Concurrent connections
    max_overflow=10,  # Additional connections on demand
    pool_timeout=30,  # Wait time for connection
    pool_recycle=3600,  # Recycle connections every hour
    pool_pre_ping=True,  # Verify connections before use
    poolclass=QueuePool  # Use NullPool for serverless
)

# Session factory
AsyncSessionLocal = async_sessionmaker(
    engine,
    class_=AsyncSession,
    expire_on_commit=False,  # Don't expire objects after commit
    autoflush=False  # Explicit flushing for better control
)

# Dependency injection for FastAPI
async def get_db() -> AsyncSession:
    async with AsyncSessionLocal() as session:
        try:
            yield session
            await session.commit()  # Auto-commit on success
        except Exception:
            await session.rollback()  # Rollback on error
            raise
        finally:
            await session.close()

# Usage in endpoint
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
        raise HTTPException(status_code=404)
    return user
```

**Key Benefits:**
- **3,000+ req/sec** throughput with async I/O
- Non-blocking database operations
- Automatic connection management
- Graceful error handling with rollback

**When to Use:**
- High-concurrency APIs (1000+ concurrent requests)
- I/O-bound workloads (database, network calls)
- Microservices with database-per-service pattern

---

#### 🔍 Pattern 2: FastAPI Background Tasks for Long-Running Operations

**Problem**: Long-running operations (email sending, file processing) block request handling.

**Solution**: Use FastAPI's `BackgroundTasks` for async background execution.

```python
from fastapi import BackgroundTasks
from typing import List
import asyncio

# Background task function
async def send_welcome_email(user_email: str):
    """Send welcome email asynchronously"""
    await asyncio.sleep(2)  # Simulate email sending
    print(f"Welcome email sent to {user_email}")

async def update_analytics(user_id: int):
    """Update user analytics in background"""
    # Non-blocking analytics update
    pass

# Endpoint with background tasks
@app.post("/users/", status_code=201)
async def create_user(
    user: UserCreate,
    background_tasks: BackgroundTasks,
    db: AsyncSession = Depends(get_db)
):
    """Create user with background email notification"""
    # Create user (fast)
    new_user = User(**user.dict())
    db.add(new_user)
    await db.commit()
    await db.refresh(new_user)
    
    # Schedule background tasks (non-blocking)
    background_tasks.add_task(send_welcome_email, user.email)
    background_tasks.add_task(update_analytics, new_user.id)
    
    return {"id": new_user.id, "message": "User created"}
```

**Use Cases:**
- Email/SMS notifications
- File uploads to S3/cloud storage
- Data export generation
- Webhook delivery
- Log aggregation

**Alternative: Celery for Heavy Workloads**

```python
# For CPU-intensive or distributed tasks
from celery import Celery

celery = Celery('tasks', broker='redis://localhost:6379/0')

@celery.task
def process_video(video_path: str):
    """Heavy video processing in background worker"""
    pass

@app.post("/videos/")
async def upload_video(video: UploadFile):
    # Save file quickly
    file_path = save_file(video)
    
    # Delegate to Celery worker (distributed)
    process_video.delay(file_path)
    
    return {"status": "processing"}
```

---

#### 🔍 Pattern 3: Dependency Injection with FastAPI for Clean Architecture

**Problem**: Tightly coupled code with repeated authentication, database, and service logic.

**Solution**: Use FastAPI's dependency injection system for modular, testable code.

```python
from fastapi import Depends, HTTPException, Header
from typing import Annotated

# Reusable dependency: JWT authentication
async def get_current_user(
    authorization: Annotated[str, Header()] = None,
    db: AsyncSession = Depends(get_db)
) -> User:
    """Extract and validate JWT token"""
    if not authorization or not authorization.startswith("Bearer "):
        raise HTTPException(status_code=401, detail="Invalid token")
    
    token = authorization.split(" ")[1]
    payload = decode_jwt(token)  # Decode JWT
    
    user = await db.get(User, payload["user_id"])
    if not user:
        raise HTTPException(status_code=401, detail="User not found")
    return user

# Reusable dependency: admin-only access
async def require_admin(
    current_user: Annotated[User, Depends(get_current_user)]
) -> User:
    """Verify user is admin"""
    if not current_user.is_admin:
        raise HTTPException(status_code=403, detail="Admin access required")
    return current_user

# Using dependencies in endpoints
@app.get("/users/")
async def list_users(
    current_user: Annotated[User, Depends(get_current_user)],
    db: AsyncSession = Depends(get_db)
):
    """List users (authenticated)"""
    users = await db.execute(select(User))
    return users.scalars().all()

@app.delete("/users/{user_id}")
async def delete_user(
    user_id: int,
    admin: Annotated[User, Depends(require_admin)],
    db: AsyncSession = Depends(get_db)
):
    """Delete user (admin only)"""
    await db.execute(delete(User).where(User.id == user_id))
    await db.commit()
    return {"status": "deleted"}
```

**Benefits:**
- **DRY Principle**: Reusable authentication logic
- **Testability**: Mock dependencies easily
- **Separation of Concerns**: Business logic separated from infrastructure
- **Type Safety**: Full editor support with type hints

---

#### 🔍 Pattern 4: Microservices Service Mesh with Istio

**Problem**: Managing inter-service communication, retries, circuit breaking, and observability.

**Solution**: Deploy Istio service mesh for declarative traffic management.

```yaml
# Service Mesh Architecture with Istio 1.24+
apiVersion: v1
kind: Service
metadata:
  name: backend-api
  labels:
    app: backend
    version: v1
spec:
  ports:
  - port: 8000
    name: http
  selector:
    app: backend
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: backend-api
spec:
  replicas: 3
  selector:
    matchLabels:
      app: backend
      version: v1
  template:
    metadata:
      labels:
        app: backend
        version: v1
      annotations:
        # Istio sidecar injection
        sidecar.istio.io/inject: "true"
    spec:
      containers:
      - name: backend
        image: backend:v1
        ports:
        - containerPort: 8000
        resources:
          requests:
            cpu: 100m
            memory: 256Mi
          limits:
            cpu: 1000m
            memory: 1Gi
---
# VirtualService: Traffic Management
apiVersion: networking.istio.io/v1beta1
kind: VirtualService
metadata:
  name: backend-routes
spec:
  hosts:
  - backend-api
  http:
  - match:
    - headers:
        x-version:
          exact: "v2"
    route:
    - destination:
        host: backend-api
        subset: v2
      weight: 100
  - route:
    - destination:
        host: backend-api
        subset: v1
      weight: 90
    - destination:
        host: backend-api
        subset: v2
      weight: 10  # Canary 10%
  timeout: 30s
  retries:
    attempts: 3
    perTryTimeout: 10s
---
# DestinationRule: Circuit Breaker
apiVersion: networking.istio.io/v1beta1
kind: DestinationRule
metadata:
  name: backend-circuit-breaker
spec:
  host: backend-api
  trafficPolicy:
    connectionPool:
      tcp:
        maxConnections: 100
      http:
        http1MaxPendingRequests: 50
        http2MaxRequests: 100
    outlierDetection:
      consecutiveErrors: 5
      interval: 30s
      baseEjectionTime: 30s
      maxEjectionPercent: 50
  subsets:
  - name: v1
    labels:
      version: v1
  - name: v2
    labels:
      version: v2
```

**Service Mesh Benefits:**
- **Traffic Splitting**: Canary deployments without code changes
- **Resilience**: Automatic retries, circuit breakers, timeouts
- **Security**: Mutual TLS (mTLS) between services
- **Observability**: Distributed tracing, metrics, logs
- **Zero Code Changes**: All traffic management in configuration

---

#### 🔍 Pattern 5: API Rate Limiting with Redis

**Problem**: Protect APIs from abuse and ensure fair usage.

**Solution**: Implement sliding window rate limiting with Redis.

```python
from fastapi import Request, HTTPException
from redis.asyncio import Redis
import time

# Redis connection pool
redis = Redis(
    host="localhost",
    port=6379,
    decode_responses=True,
    max_connections=50
)

# Rate limiter dependency
async def rate_limit(
    request: Request,
    max_requests: int = 100,
    window_seconds: int = 60
):
    """Sliding window rate limiter using Redis"""
    client_ip = request.client.host
    key = f"rate_limit:{client_ip}"
    
    current_time = int(time.time())
    window_start = current_time - window_seconds
    
    # Use Redis sorted set for sliding window
    pipeline = redis.pipeline()
    
    # Remove old entries
    pipeline.zremrangebyscore(key, 0, window_start)
    
    # Count requests in current window
    pipeline.zcard(key)
    
    # Add current request
    pipeline.zadd(key, {str(current_time): current_time})
    
    # Set expiration
    pipeline.expire(key, window_seconds)
    
    results = await pipeline.execute()
    request_count = results[1]
    
    if request_count >= max_requests:
        raise HTTPException(
            status_code=429,
            detail=f"Rate limit exceeded: {max_requests} requests per {window_seconds}s"
        )
    
    # Add rate limit headers
    remaining = max_requests - request_count - 1
    request.state.rate_limit_remaining = remaining

# Apply to endpoints
@app.get("/api/data", dependencies=[Depends(rate_limit)])
async def get_data():
    return {"data": "protected content"}

# Custom rate limits per endpoint
@app.post("/api/expensive-operation")
async def expensive_op(
    _: None = Depends(lambda req: rate_limit(req, max_requests=10, window_seconds=60))
):
    return {"status": "processing"}
```

**Advanced: Token Bucket with Burst Capacity**

```python
async def token_bucket_limiter(
    request: Request,
    rate: int = 10,  # Tokens per second
    burst: int = 20  # Max burst capacity
):
    """Token bucket algorithm for burst handling"""
    client_ip = request.client.host
    key = f"token_bucket:{client_ip}"
    
    current_time = time.time()
    
    # Get current tokens and last refill time
    data = await redis.hgetall(key)
    
    if data:
        tokens = float(data.get("tokens", burst))
        last_refill = float(data.get("last_refill", current_time))
    else:
        tokens = burst
        last_refill = current_time
    
    # Refill tokens based on elapsed time
    elapsed = current_time - last_refill
    tokens = min(burst, tokens + (elapsed * rate))
    
    if tokens < 1:
        raise HTTPException(status_code=429, detail="Rate limit exceeded")
    
    # Consume one token
    tokens -= 1
    
    # Update Redis
    await redis.hset(key, mapping={
        "tokens": str(tokens),
        "last_refill": str(current_time)
    })
    await redis.expire(key, 3600)
```

---

#### 🔍 Pattern 6: WebSocket Real-Time Communication

**Problem**: Implementing real-time bidirectional communication for chat, notifications, live updates.

**Solution**: Use FastAPI WebSocket support with connection management.

```python
from fastapi import WebSocket, WebSocketDisconnect
from typing import List

class ConnectionManager:
    """Manage WebSocket connections"""
    def __init__(self):
        self.active_connections: List[WebSocket] = []
    
    async def connect(self, websocket: WebSocket):
        await websocket.accept()
        self.active_connections.append(websocket)
    
    def disconnect(self, websocket: WebSocket):
        self.active_connections.remove(websocket)
    
    async def broadcast(self, message: str):
        """Broadcast message to all clients"""
        for connection in self.active_connections:
            await connection.send_text(message)
    
    async def send_personal(self, message: str, websocket: WebSocket):
        """Send message to specific client"""
        await websocket.send_text(message)

manager = ConnectionManager()

# WebSocket endpoint
@app.websocket("/ws/chat/{client_id}")
async def chat_endpoint(websocket: WebSocket, client_id: str):
    """Real-time chat WebSocket"""
    await manager.connect(websocket)
    
    try:
        # Send welcome message
        await manager.send_personal(
            f"Welcome {client_id}! You are connected.",
            websocket
        )
        
        # Notify others
        await manager.broadcast(f"{client_id} joined the chat")
        
        # Message loop
        while True:
            # Receive message from client
            data = await websocket.receive_text()
            
            # Broadcast to all clients
            await manager.broadcast(f"{client_id}: {data}")
            
    except WebSocketDisconnect:
        manager.disconnect(websocket)
        await manager.broadcast(f"{client_id} left the chat")

# JavaScript client example
"""
const ws = new WebSocket("ws://localhost:8000/ws/chat/user123");

ws.onopen = () => console.log("Connected");
ws.onmessage = (event) => console.log(event.data);
ws.onerror = (error) => console.error(error);
ws.onclose = () => console.log("Disconnected");

// Send message
ws.send("Hello everyone!");
"""
```

**Production Considerations:**
- **Redis Pub/Sub**: For multi-instance deployments
- **Heartbeat**: Detect and close dead connections
- **Authentication**: Validate JWT tokens in WebSocket handshake
- **Backpressure**: Handle slow clients without blocking fast ones

---

#### 🔍 Pattern 7: Django Ninja Async ORM Access

**Problem**: Django's ORM is synchronous by default, blocking async views.

**Solution**: Use Django 5.2+ async ORM or `sync_to_async` adapter.

```python
from django.http import JsonResponse
from ninja import NinjaAPI, Schema
from asgiref.sync import sync_to_async
from typing import List

api = NinjaAPI()

# Native async ORM (Django 5.2+)
@api.get("/users/{user_id}")
async def get_user(request, user_id: int):
    """Async user retrieval using Django async ORM"""
    from myapp.models import User
    
    # Native async query
    user = await User.objects.aget(id=user_id)
    
    return {
        "id": user.id,
        "username": user.username,
        "email": user.email
    }

# Async queryset iteration
@api.get("/users/")
async def list_users(request):
    """List all users asynchronously"""
    from myapp.models import User
    
    # Async for loop
    users = [user async for user in User.objects.all()]
    
    return {"users": [{"id": u.id, "username": u.username} for u in users]}

# Fallback: sync_to_async for older Django or complex queries
@sync_to_async
def get_user_with_relations(user_id: int):
    """Synchronous query wrapped for async context"""
    from myapp.models import User
    return User.objects.select_related('profile').get(id=user_id)

@api.get("/users/{user_id}/profile")
async def get_user_profile(request, user_id: int):
    """Async endpoint using sync_to_async"""
    user = await get_user_with_relations(user_id)
    
    return {
        "username": user.username,
        "profile": {
            "bio": user.profile.bio,
            "avatar": user.profile.avatar_url
        }
    }
```

**Performance Tips:**
- Use `select_related()` for foreign keys to avoid N+1 queries
- Use `prefetch_related()` for many-to-many and reverse foreign keys
- Use `only()` and `defer()` to load specific fields
- Use `acount()`, `aexists()`, `afirst()` for async aggregations

---

#### 🔍 Pattern 8: API Versioning Strategies

**Problem**: Evolving APIs without breaking existing clients.

**Solution**: Implement versioning through URL path, headers, or content negotiation.

```python
from fastapi import APIRouter, Header
from typing import Annotated

# Strategy 1: URL Path Versioning (Most Common)
app = FastAPI()

v1_router = APIRouter(prefix="/api/v1")
v2_router = APIRouter(prefix="/api/v2")

# Version 1: Legacy endpoint
@v1_router.get("/users/")
def list_users_v1():
    """Legacy user list format"""
    return {"users": [{"id": 1, "name": "John"}]}

# Version 2: Enhanced endpoint
@v2_router.get("/users/")
def list_users_v2():
    """Enhanced user list with pagination"""
    return {
        "users": [{"id": 1, "username": "john", "email": "john@example.com"}],
        "pagination": {"page": 1, "total": 100}
    }

app.include_router(v1_router)
app.include_router(v2_router)

# Strategy 2: Header-Based Versioning
@app.get("/users/")
def list_users(
    accept_version: Annotated[str, Header(alias="Accept-Version")] = "v1"
):
    """Version selected via Accept-Version header"""
    if accept_version == "v2":
        return list_users_v2()
    return list_users_v1()

# Strategy 3: Content Negotiation (Accept Header)
@app.get("/users/")
def list_users_negotiated(
    accept: Annotated[str, Header()] = "application/vnd.api.v1+json"
):
    """Version via Accept header"""
    if "v2" in accept:
        return list_users_v2()
    return list_users_v1()
```

**Versioning Best Practices:**
- **Default to Latest**: Omitted version should use latest stable
- **Deprecation Warnings**: Add `Deprecated: true` header for sunset versions
- **Semantic Versioning**: Major.Minor.Patch (v1.2.3)
- **Documentation**: Clear migration guides between versions
- **Sunset Policy**: Announce deprecation 6-12 months in advance

---

#### 🔍 Pattern 9: Database Query Optimization

**Problem**: Slow queries causing high latency and poor scalability.

**Solution**: Use proper indexing, query optimization, and caching.

```python
from sqlalchemy import select, func
from sqlalchemy.orm import selectinload, joinedload

# ❌ BAD: N+1 Query Problem
@app.get("/posts/")
async def get_posts_bad(db: AsyncSession = Depends(get_db)):
    """Inefficient: N+1 queries"""
    posts = await db.execute(select(Post))
    posts = posts.scalars().all()
    
    # This triggers N separate queries for authors!
    return [{"title": p.title, "author": p.author.name} for p in posts]

# ✅ GOOD: Eager Loading with joinedload
@app.get("/posts/")
async def get_posts_good(db: AsyncSession = Depends(get_db)):
    """Efficient: Single JOIN query"""
    posts = await db.execute(
        select(Post).options(joinedload(Post.author))
    )
    posts = posts.unique().scalars().all()
    
    return [{"title": p.title, "author": p.author.name} for p in posts]

# ✅ BETTER: selectinload for One-to-Many relationships
@app.get("/authors/{author_id}/posts")
async def get_author_posts(
    author_id: int,
    db: AsyncSession = Depends(get_db)
):
    """Efficient: Separate optimized queries"""
    author = await db.execute(
        select(Author)
        .where(Author.id == author_id)
        .options(selectinload(Author.posts))
    )
    author = author.scalar_one()
    
    return {
        "author": author.name,
        "posts": [{"title": p.title} for p in author.posts]
    }

# ✅ BEST: Pagination with limit/offset
@app.get("/posts/")
async def get_posts_paginated(
    page: int = 1,
    per_page: int = 20,
    db: AsyncSession = Depends(get_db)
):
    """Paginated results for large datasets"""
    offset = (page - 1) * per_page
    
    query = select(Post).options(joinedload(Post.author))
    
    # Get total count
    count_query = select(func.count()).select_from(Post)
    total = await db.execute(count_query)
    total = total.scalar()
    
    # Get page results
    posts = await db.execute(
        query.limit(per_page).offset(offset)
    )
    posts = posts.unique().scalars().all()
    
    return {
        "posts": [{"title": p.title, "author": p.author.name} for p in posts],
        "pagination": {
            "page": page,
            "per_page": per_page,
            "total": total,
            "pages": (total + per_page - 1) // per_page
        }
    }

# Index creation for performance
from sqlalchemy import Index

class Post(Base):
    __tablename__ = "posts"
    
    id = Column(Integer, primary_key=True)
    title = Column(String, index=True)  # Single column index
    author_id = Column(Integer, ForeignKey("authors.id"))
    created_at = Column(DateTime, index=True)
    status = Column(String)
    
    # Composite index for common queries
    __table_args__ = (
        Index('idx_author_status', 'author_id', 'status'),
        Index('idx_created_status', 'created_at', 'status'),
    )
```

---

#### 🔍 Pattern 10: Observability with OpenTelemetry

**Problem**: Debugging distributed systems without visibility into request flows.

**Solution**: Implement distributed tracing with OpenTelemetry.

```python
from opentelemetry import trace
from opentelemetry.instrumentation.fastapi import FastAPIInstrumentor
from opentelemetry.instrumentation.sqlalchemy import SQLAlchemyInstrumentor
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.trace.export import BatchSpanProcessor
from opentelemetry.exporter.jaeger.thrift import JaegerExporter

# Initialize tracing
trace.set_tracer_provider(TracerProvider())
tracer = trace.get_tracer(__name__)

# Export to Jaeger
jaeger_exporter = JaegerExporter(
    agent_host_name="localhost",
    agent_port=6831,
)

trace.get_tracer_provider().add_span_processor(
    BatchSpanProcessor(jaeger_exporter)
)

# Auto-instrument FastAPI
FastAPIInstrumentor.instrument_app(app)

# Auto-instrument SQLAlchemy
SQLAlchemyInstrumentor().instrument(engine=engine.sync_engine)

# Manual span creation for custom operations
@app.get("/process/{item_id}")
async def process_item(item_id: int):
    """Endpoint with custom spans"""
    with tracer.start_as_current_span("process-item") as span:
        # Add attributes to span
        span.set_attribute("item.id", item_id)
        span.set_attribute("user.role", "admin")
        
        # Child span for database operation
        with tracer.start_as_current_span("fetch-item-details"):
            item = await fetch_item_from_db(item_id)
        
        # Child span for external API call
        with tracer.start_as_current_span("enrich-item-data"):
            enriched = await call_external_api(item)
        
        # Record events
        span.add_event("Item processed successfully")
        
        return enriched
```

**Metrics with Prometheus:**

```python
from prometheus_client import Counter, Histogram, make_asgi_app
import time

# Define metrics
REQUEST_COUNT = Counter(
    'api_requests_total',
    'Total API requests',
    ['method', 'endpoint', 'status']
)

REQUEST_LATENCY = Histogram(
    'api_request_duration_seconds',
    'API request latency',
    ['method', 'endpoint']
)

# Middleware to track metrics
@app.middleware("http")
async def metrics_middleware(request: Request, call_next):
    start_time = time.time()
    
    response = await call_next(request)
    
    # Record metrics
    REQUEST_COUNT.labels(
        method=request.method,
        endpoint=request.url.path,
        status=response.status_code
    ).inc()
    
    REQUEST_LATENCY.labels(
        method=request.method,
        endpoint=request.url.path
    ).observe(time.time() - start_time)
    
    return response

# Expose metrics endpoint
metrics_app = make_asgi_app()
app.mount("/metrics", metrics_app)
```

---

### Level 3: Advanced Patterns (Expert Reference)

#### 🏗️ Kubernetes Production Deployment

**Complete production-ready Kubernetes manifest:**

```yaml
# Namespace
apiVersion: v1
kind: Namespace
metadata:
  name: backend-prod
  labels:
    istio-injection: enabled
---
# ConfigMap for application config
apiVersion: v1
kind: ConfigMap
metadata:
  name: backend-config
  namespace: backend-prod
data:
  DATABASE_POOL_SIZE: "20"
  LOG_LEVEL: "INFO"
  REDIS_HOST: "redis-service.backend-prod.svc.cluster.local"
---
# Secret for sensitive data
apiVersion: v1
kind: Secret
metadata:
  name: backend-secrets
  namespace: backend-prod
type: Opaque
data:
  DATABASE_URL: <base64-encoded-url>
  JWT_SECRET: <base64-encoded-secret>
---
# Deployment with auto-scaling
apiVersion: apps/v1
kind: Deployment
metadata:
  name: backend-api
  namespace: backend-prod
spec:
  replicas: 3
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxSurge: 1
      maxUnavailable: 0
  selector:
    matchLabels:
      app: backend
  template:
    metadata:
      labels:
        app: backend
        version: v1
      annotations:
        prometheus.io/scrape: "true"
        prometheus.io/port: "8000"
        prometheus.io/path: "/metrics"
    spec:
      affinity:
        podAntiAffinity:
          preferredDuringSchedulingIgnoredDuringExecution:
          - weight: 100
            podAffinityTerm:
              labelSelector:
                matchExpressions:
                - key: app
                  operator: In
                  values:
                  - backend
              topologyKey: kubernetes.io/hostname
      containers:
      - name: backend
        image: backend:v1.0.0
        ports:
        - containerPort: 8000
          name: http
        env:
        - name: DATABASE_URL
          valueFrom:
            secretKeyRef:
              name: backend-secrets
              key: DATABASE_URL
        - name: DATABASE_POOL_SIZE
          valueFrom:
            configMapKeyRef:
              name: backend-config
              key: DATABASE_POOL_SIZE
        resources:
          requests:
            cpu: 250m
            memory: 512Mi
          limits:
            cpu: 1000m
            memory: 2Gi
        livenessProbe:
          httpGet:
            path: /health
            port: 8000
          initialDelaySeconds: 30
          periodSeconds: 10
          timeoutSeconds: 5
          failureThreshold: 3
        readinessProbe:
          httpGet:
            path: /ready
            port: 8000
          initialDelaySeconds: 10
          periodSeconds: 5
          timeoutSeconds: 3
          failureThreshold: 2
---
# Horizontal Pod Autoscaler
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: backend-hpa
  namespace: backend-prod
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: backend-api
  minReplicas: 3
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 80
---
# Service
apiVersion: v1
kind: Service
metadata:
  name: backend-service
  namespace: backend-prod
spec:
  selector:
    app: backend
  ports:
  - port: 80
    targetPort: 8000
    name: http
  type: ClusterIP
---
# Ingress
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: backend-ingress
  namespace: backend-prod
  annotations:
    cert-manager.io/cluster-issuer: "letsencrypt-prod"
    nginx.ingress.kubernetes.io/rate-limit: "100"
spec:
  ingressClassName: nginx
  tls:
  - hosts:
    - api.example.com
    secretName: backend-tls
  rules:
  - host: api.example.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: backend-service
            port:
              number: 80
```

---

## 🎯 Best Practices Checklist

**Must-Have:**
- ✅ Use async frameworks (FastAPI, Django Ninja) for I/O-bound workloads
- ✅ Implement connection pooling for databases (20-50 connections)
- ✅ Use dependency injection for authentication, database sessions
- ✅ Add health check endpoints (`/health`, `/ready`)
- ✅ Implement structured logging (JSON format)
- ✅ Use environment variables for configuration (12-factor app)
- ✅ Version APIs explicitly (`/api/v1`, `/api/v2`)
- ✅ Add rate limiting to public endpoints

**Recommended:**
- ✅ Implement distributed tracing (OpenTelemetry + Jaeger)
- ✅ Use Redis for caching and session storage
- ✅ Add background task processing (FastAPI BackgroundTasks or Celery)
- ✅ Implement circuit breakers for external API calls
- ✅ Use Kubernetes HPA for auto-scaling
- ✅ Add comprehensive monitoring (Prometheus + Grafana)
- ✅ Implement graceful shutdown for zero-downtime deployments

**Security:**
- 🔒 Use HTTPS/TLS for all external communication
- 🔒 Implement JWT authentication with short expiration (15 min)
- 🔒 Enable mTLS between microservices
- 🔒 Validate all inputs with Pydantic schemas
- 🔒 Use parameterized queries to prevent SQL injection
- 🔒 Implement CORS policies restrictively
- 🔒 Rate limit API endpoints (100 req/min default)
- 🔒 Use secrets management (Kubernetes Secrets, Vault)

---

## 🔗 Context7 MCP Integration

**When to Use Context7 for This Skill:**

This skill benefits from Context7 when:
- Working with FastAPI, Django, SQLAlchemy latest versions
- Need latest documentation for async patterns
- Verifying Kubernetes, Istio configuration details
- Checking database driver compatibility

**Example Usage:**

```python
# Fetch latest FastAPI documentation
from moai_adk.integrations import Context7Helper

helper = Context7Helper()
docs = await helper.get_docs(
    library_id="/fastapi/fastapi/0.118.2",
    topic="async dependencies middleware websockets",
    tokens=8000
)
```

**Relevant Libraries:**

| Library | Context7 ID | Use Case |
|---------|-------------|----------|
| FastAPI | `/fastapi/fastapi/0.118.2` | Async API development |
| Django Ninja | `/vitalik/django-ninja/v1_4_3` | Django async APIs |
| SQLAlchemy | `/sqlalchemy/sqlalchemy` | ORM and database access |
| Pydantic | `/pydantic/pydantic` | Data validation |
| Redis | `/redis/redis-py` | Caching and rate limiting |

---

## 📊 Decision Tree

**When to use moai-domain-backend:**

```
Start
  ├─ Building REST/GraphQL API?
  │   ├─ YES → Use FastAPI patterns (async, dependencies)
  │   └─ NO → Continue
  ├─ Microservices architecture?
  │   ├─ YES → Use service mesh patterns (Istio, circuit breakers)
  │   └─ NO → Continue
  ├─ High-concurrency requirements (1000+ req/sec)?
  │   ├─ YES → Use async patterns, connection pooling
  │   └─ NO → Use synchronous patterns
  ├─ Need real-time communication?
  │   ├─ YES → Use WebSocket patterns
  │   └─ NO → Continue
  └─ Database optimization needed?
      ├─ YES → Use eager loading, indexing, caching
      └─ NO → Start with Level 1 quick reference
```

---

## 🔄 Integration with Other Skills

**Prerequisite Skills:**
- Skill("moai-domain-database") – Database design and optimization
- Skill("moai-domain-security") – Authentication and authorization patterns

**Complementary Skills:**
- Skill("moai-domain-web-api") – REST/GraphQL API design
- Skill("moai-domain-devops") – CI/CD and deployment automation
- Skill("moai-domain-frontend") – Full-stack integration patterns

**Next Steps:**
- Skill("moai-domain-ml") – Deploy ML models in production APIs
- Skill("moai-domain-testing") – Comprehensive testing strategies

---

## 📚 Official References

### FastAPI Resources
- **Official Documentation**: https://fastapi.tiangolo.com
- **Async Guide**: https://fastapi.tiangolo.com/async/
- **Dependency Injection**: https://fastapi.tiangolo.com/tutorial/dependencies/
- **WebSocket Support**: https://fastapi.tiangolo.com/advanced/websockets/
- **Background Tasks**: https://fastapi.tiangolo.com/tutorial/background-tasks/

### Django Ninja Resources
- **Official Documentation**: https://django-ninja.dev
- **Async Support**: https://django-ninja.dev/guides/async-support/
- **Schema Validation**: https://django-ninja.dev/guides/input/
- **Pagination**: https://django-ninja.dev/guides/response/pagination/

### Service Mesh Resources
- **Istio Documentation**: https://istio.io/latest/docs/
- **Kubernetes Patterns**: https://kubernetes.io/docs/concepts/
- **OpenTelemetry**: https://opentelemetry.io/docs/

### Performance Optimization
- **FastAPI Production Best Practices 2025**: Industry standard for 3,000+ req/sec
- **Async Database Drivers**: asyncpg (PostgreSQL), motor (MongoDB)
- **Connection Pooling**: SQLAlchemy async engine configuration
- **Rate Limiting**: Redis sliding window and token bucket algorithms

---

## 📈 Version History

**v4.0.0** (2025-11-12)
- ✨ Enterprise v4.0 upgrade: 800+ lines, 10+ comprehensive examples
- ✨ Context7 MCP integration with FastAPI 0.118.2, Django Ninja 1.4.3
- ✨ Progressive Disclosure structure (Quick Reference → Practical → Advanced)
- ✨ Modern async patterns (FastAPI lifespan, dependency injection, WebSocket)
- ✨ Microservices service mesh patterns (Istio 1.24+, circuit breakers)
- ✨ Production Kubernetes deployment manifests (HPA, health checks)
- ✨ Database optimization (connection pooling, query optimization)
- ✨ Rate limiting strategies (Redis sliding window, token bucket)
- ✨ Observability (OpenTelemetry, Prometheus, distributed tracing)
- ✨ Real-world production examples from 2025 best practices

---

**Generated with**: MoAI-ADK Skill Factory v4.0  
**Last Updated**: 2025-11-12  
**Maintained by**: Primary Agent (backend-expert)
